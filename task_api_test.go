package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"dashboard/internal/tasks"
	"dashboard/internal/testutil"
)

func TestTaskAPI_WithoutDatabase(t *testing.T) {
	api := &TaskAPI{app: &App{ctx: context.Background()}}

	if _, err := api.Agenda(); !errors.Is(err, errDBUnavailable) {
		t.Errorf("Agenda: err = %v, want errDBUnavailable", err)
	}
	if _, err := api.Create(TaskInputDTO{}); !errors.Is(err, errDBUnavailable) {
		t.Errorf("Create: err = %v, want errDBUnavailable", err)
	}
}

func TestParseInput(t *testing.T) {
	in, err := parseInput(TaskInputDTO{
		Title:    "Estudar",
		Urgency:  "high",
		StartsAt: "2026-10-13T14:00:00-03:00",
		EndsAt:   "2026-10-13T15:00:00-03:00",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.EndsAt.Sub(in.StartsAt) != time.Hour || in.Urgency != tasks.UrgencyHigh {
		t.Errorf("input = %+v", in)
	}

	if _, err := parseInput(TaskInputDTO{StartsAt: "13/10/2026 14:00"}); err == nil {
		t.Error("expected an error for a non RFC 3339 date")
	}
}

func TestTaskAPI_CreateThenAgenda(t *testing.T) {
	app := &App{ctx: context.Background(), tasks: tasks.NewService(testutil.Queries(t))}
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
	if created.Urgency != "medium" {
		t.Errorf("Urgency = %q, want the default medium", created.Urgency)
	}

	agenda, err := api.Agenda()
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
