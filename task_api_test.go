package main

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"dashboard/internal/tasks"
	"dashboard/internal/testutil"
)

func TestTaskAPI_WithoutDatabase(t *testing.T) {
	api := &TaskAPI{app: &App{ctx: context.Background()}}

	if _, err := api.Agenda(FilterDTO{}); !errors.Is(err, errDBUnavailable) {
		t.Errorf("Agenda: err = %v, want errDBUnavailable", err)
	}
	if _, err := api.Create(TaskInputDTO{}); !errors.Is(err, errDBUnavailable) {
		t.Errorf("Create: err = %v, want errDBUnavailable", err)
	}
	if err := api.SetItemDone(1, "", true); !errors.Is(err, errDBUnavailable) {
		t.Errorf("SetItemDone: err = %v, want errDBUnavailable", err)
	}
}

func TestParseInput(t *testing.T) {
	in, err := parseInput(TaskInputDTO{
		Title:    "Academia",
		Category: "Saúde",
		Urgency:  "high",
		StartsAt: "2026-10-13T10:00:00Z",
		EndsAt:   "2026-10-13T11:00:00Z",
		Repeat:   RepeatDTO{Kind: "weekly", Weekdays: []int{1, 3}, Until: "2026-12-31"},
		Items:    []ItemInputDTO{{ID: 4, Title: "alongar"}, {Title: "correr"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.StartsAt.Location() != time.Local {
		t.Errorf("StartsAt location = %v, want local (calendar rules are local)", in.StartsAt.Location())
	}
	if in.EndsAt.Sub(in.StartsAt) != time.Hour || in.Urgency != tasks.UrgencyHigh || in.Category != "Saúde" {
		t.Errorf("input = %+v", in)
	}
	if in.Repeat.Kind != tasks.RepeatWeekly || !slices.Equal(in.Repeat.Weekdays, []time.Weekday{time.Monday, time.Wednesday}) {
		t.Errorf("Repeat = %+v", in.Repeat)
	}
	if in.Repeat.Until == nil || tasks.DayKey(*in.Repeat.Until) != "2026-12-31" {
		t.Errorf("Until = %v", in.Repeat.Until)
	}
	if len(in.Items) != 2 || in.Items[0].ID != 4 || in.Items[1].Title != "correr" {
		t.Errorf("Items = %+v", in.Items)
	}
}

func TestParseInput_InvalidDates(t *testing.T) {
	if _, err := parseInput(TaskInputDTO{StartsAt: "13/10/2026 14:00"}); err == nil {
		t.Error("expected an error for a non RFC 3339 instant")
	}
	if _, err := parseInput(TaskInputDTO{Repeat: RepeatDTO{Until: "31/12/2026"}}); err == nil {
		t.Error("expected an error for a non YYYY-MM-DD day")
	}
}

func TestToDTO_Occurrence(t *testing.T) {
	first := time.Date(2026, 10, 7, 7, 0, 0, 0, time.Local)
	series := tasks.Task{
		ID:       1,
		StartsAt: first,
		EndsAt:   first.Add(time.Hour),
		Repeat:   tasks.Repeat{Kind: tasks.RepeatWeekly, Weekdays: []time.Weekday{time.Wednesday}},
		Items:    []tasks.Item{{ID: 2, Title: "a"}},
	}
	occ := series.At(time.Date(2026, 10, 14, 0, 0, 0, 0, time.Local))
	occ.Items[0].Done = true
	dto := toDTO(occ)

	if dto.SeriesStartsAt != first.Format(time.RFC3339) || dto.StartsAt == dto.SeriesStartsAt {
		t.Errorf("startsAt = %s, seriesStartsAt = %s", dto.StartsAt, dto.SeriesStartsAt)
	}

	if dto.OccursOn != "2026-10-14" || dto.Repeat.Kind != "weekly" || !slices.Equal(dto.Repeat.Weekdays, []int{3}) {
		t.Errorf("dto = %+v", dto)
	}
	if len(dto.Items) != 1 || !dto.Items[0].Done {
		t.Errorf("Items = %+v", dto.Items)
	}
}

func TestTaskAPI_CreateThenAgenda(t *testing.T) {
	app := &App{ctx: context.Background(), tasks: tasks.NewService(testutil.Tx(t))}
	api := &TaskAPI{app: app}
	start := time.Now().Add(time.Hour)

	created, err := api.Create(TaskInputDTO{
		Title:    "Estudar",
		StartsAt: start.Format(time.RFC3339),
		EndsAt:   start.Add(time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Urgency != "medium" || created.Repeat.Kind != "none" {
		t.Errorf("created = %+v, want the defaults", created)
	}

	agenda, err := api.Agenda(FilterDTO{})
	if err != nil {
		t.Fatalf("Agenda: %v", err)
	}
	for _, d := range agenda.Days {
		for _, tk := range d.Tasks {
			if tk.ID == created.ID {
				return
			}
		}
	}
	t.Errorf("created task missing from agenda: %+v", agenda)
}

func TestTaskAPI_AgendaRejectsInvalidUrgencyFilter(t *testing.T) {
	app := &App{ctx: context.Background(), tasks: tasks.NewService(testutil.Tx(t))}

	if _, err := (&TaskAPI{app: app}).Agenda(FilterDTO{Urgency: "urgent"}); !errors.Is(err, tasks.ErrInvalidUrgency) {
		t.Errorf("err = %v, want ErrInvalidUrgency", err)
	}
}
