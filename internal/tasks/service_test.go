package tasks_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"dashboard/internal/db/sqlc"
	"dashboard/internal/tasks"
	"dashboard/internal/testutil"
)

func newService(t *testing.T) (*tasks.Service, *sqlc.Queries) {
	tx := testutil.Tx(t)
	return tasks.NewService(tx), sqlc.New(tx)
}

// input truncates to microseconds, the precision Postgres stores.
func input(start time.Time) tasks.Input {
	start = start.Truncate(time.Microsecond)
	return tasks.Input{Title: "Estudar", Urgency: tasks.UrgencyHigh, StartsAt: start, EndsAt: start.Add(time.Hour)}
}

// today returns a one-hour slot today at the given hour, local time.
func today(hour int) time.Time {
	return tasks.StartOfDay(time.Now()).Add(time.Duration(hour) * time.Hour)
}

func items(titles ...string) []tasks.ItemInput {
	out := []tasks.ItemInput{}
	for _, title := range titles {
		out = append(out, tasks.ItemInput{Title: title})
	}
	return out
}

func titles(list []tasks.Item) []string {
	out := []string{}
	for _, it := range list {
		out = append(out, it.Title)
	}
	return out
}

// findInAgenda returns the agenda entry for the task on day (or overdue).
func findInAgenda(t *testing.T, s *tasks.Service, id int64, day time.Time) (tasks.Task, bool) {
	t.Helper()
	agenda, err := s.Agenda(context.Background(), time.Now(), tasks.Filter{})
	if err != nil {
		t.Fatalf("Agenda: %v", err)
	}
	for _, d := range agenda.Days {
		if !d.Date.Equal(tasks.StartOfDay(day)) {
			continue
		}
		for _, tk := range d.Tasks {
			if tk.ID == id {
				return tk, true
			}
		}
	}
	return tasks.Task{}, false
}

func mustCreate(t *testing.T, s *tasks.Service, in tasks.Input) tasks.Task {
	t.Helper()
	created, err := s.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	return created
}

func TestService_CreateAndAgenda(t *testing.T) {
	s, _ := newService(t)
	now := time.Now()

	created := mustCreate(t, s, input(now.Add(time.Hour)))
	if created.ID == 0 || created.Urgency != tasks.UrgencyHigh || created.Done() {
		t.Errorf("created = %+v", created)
	}
	if _, ok := findInAgenda(t, s, created.ID, created.StartsAt); !ok {
		t.Error("created task not in agenda")
	}
}

func TestService_CreateRejectsInvalidInput(t *testing.T) {
	s, _ := newService(t)
	in := input(time.Now())
	in.EndsAt = in.StartsAt

	if _, err := s.Create(context.Background(), in); !errors.Is(err, tasks.ErrEndBeforeStart) {
		t.Fatalf("err = %v, want ErrEndBeforeStart", err)
	}
}

func TestService_Update(t *testing.T) {
	ctx := context.Background()
	s, _ := newService(t)
	created := mustCreate(t, s, input(time.Now()))

	in := input(time.Now().Add(24 * time.Hour))
	in.Title = "Revisar"
	in.Urgency = tasks.UrgencyLow
	updated, err := s.Update(ctx, created.ID, in)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != "Revisar" || updated.Urgency != tasks.UrgencyLow || !updated.StartsAt.Equal(in.StartsAt) {
		t.Errorf("updated = %+v", updated)
	}

	if _, err := s.Update(ctx, -1, in); !errors.Is(err, tasks.ErrNotFound) {
		t.Errorf("Update missing: err = %v, want ErrNotFound", err)
	}
}

func TestService_SetDoneToggles(t *testing.T) {
	ctx := context.Background()
	s, q := newService(t)
	created := mustCreate(t, s, input(time.Now()))

	for _, done := range []bool{true, false} {
		if err := s.SetDone(ctx, created.ID, nil, done); err != nil {
			t.Fatalf("SetDone(%v): %v", done, err)
		}
		row, _ := q.GetTask(ctx, created.ID)
		if (row.DoneAt != nil) != done {
			t.Errorf("after SetDone(%v), done_at = %v", done, row.DoneAt)
		}
	}

	if err := s.SetDone(ctx, -1, nil, true); !errors.Is(err, tasks.ErrNotFound) {
		t.Errorf("SetDone missing: err = %v, want ErrNotFound", err)
	}
}

func TestService_Delete(t *testing.T) {
	ctx := context.Background()
	s, _ := newService(t)
	in := input(time.Now())
	in.Items = items("a", "b")
	created := mustCreate(t, s, in)

	if err := s.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := s.Delete(ctx, created.ID); !errors.Is(err, tasks.ErrNotFound) {
		t.Errorf("second Delete: err = %v, want ErrNotFound", err)
	}
	if err := s.SetItemDone(ctx, created.Items[0].ID, nil, true); !errors.Is(err, tasks.ErrItemNotFound) {
		t.Errorf("items should be deleted with the task: err = %v", err)
	}
}

func TestDatabase_RejectsEndBeforeStart(t *testing.T) {
	_, q := newService(t)
	start := time.Now()

	_, err := q.CreateTask(context.Background(), sqlc.CreateTaskParams{
		Title:          "x",
		Urgency:        sqlc.TaskUrgencyLow,
		StartsAt:       start,
		EndsAt:         start.Add(-time.Hour),
		Repeat:         sqlc.TaskRepeatNone,
		RepeatWeekdays: []int16{},
	})
	if err == nil {
		t.Fatal("expected the database check constraint to reject the row")
	}
}

func TestService_CategoryReusesExistingSpelling(t *testing.T) {
	ctx := context.Background()
	s, _ := newService(t)

	first := input(time.Now())
	first.Category = "Faculdade"
	mustCreate(t, s, first)

	second := input(time.Now())
	second.Category = "  faculdade "
	created := mustCreate(t, s, second)
	if created.Category != "Faculdade" {
		t.Errorf("Category = %q, want the existing spelling", created.Category)
	}

	third := input(time.Now())
	third.Category = "Trabalho"
	mustCreate(t, s, third)

	got, err := s.Categories(ctx)
	if err != nil {
		t.Fatalf("Categories: %v", err)
	}
	if !slices.Equal(got, []string{"Faculdade", "Trabalho"}) {
		t.Errorf("Categories = %v", got)
	}
}

func TestService_ItemsAreSyncedOnUpdate(t *testing.T) {
	ctx := context.Background()
	s, _ := newService(t)
	in := input(time.Now())
	in.Items = items("arroz", "feijão", "café")
	created := mustCreate(t, s, in)

	if err := s.SetItemDone(ctx, created.Items[0].ID, nil, true); err != nil {
		t.Fatal(err)
	}

	// Reorder, rename "feijão", drop "café", add "leite".
	in.Items = []tasks.ItemInput{
		{ID: created.Items[1].ID, Title: "feijão preto"},
		{ID: created.Items[0].ID, Title: "arroz"},
		{Title: "leite"},
	}
	updated, err := s.Update(ctx, created.ID, in)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got := titles(updated.Items); !slices.Equal(got, []string{"feijão preto", "arroz", "leite"}) {
		t.Errorf("items = %v", got)
	}
	if !updated.Items[1].Done || updated.Items[0].Done {
		t.Errorf("done state not kept: %+v", updated.Items)
	}
}

func TestService_ItemsCompleteAndReopenTask(t *testing.T) {
	ctx := context.Background()
	s, q := newService(t)
	in := input(time.Now().Add(time.Hour))
	in.Items = items("arroz", "feijão")
	created := mustCreate(t, s, in)

	isDone := func() bool {
		row, err := q.GetTask(ctx, created.ID)
		if err != nil {
			t.Fatal(err)
		}
		return row.DoneAt != nil
	}

	if err := s.SetDone(ctx, created.ID, nil, true); !errors.Is(err, tasks.ErrDoneByItems) {
		t.Errorf("manual SetDone: err = %v, want ErrDoneByItems", err)
	}

	steps := []struct {
		item     int
		done     bool
		wantTask bool
	}{
		{0, true, false},
		{1, true, true},
		{0, false, false},
	}
	for _, step := range steps {
		if err := s.SetItemDone(ctx, created.Items[step.item].ID, nil, step.done); err != nil {
			t.Fatal(err)
		}
		if got := isDone(); got != step.wantTask {
			t.Errorf("after item %d = %v: task done = %v, want %v", step.item, step.done, got, step.wantTask)
		}
	}
}

func TestService_RecurringOccurrenceProgress(t *testing.T) {
	ctx := context.Background()
	s, _ := newService(t)
	in := input(today(7))
	in.Repeat = tasks.Repeat{Kind: tasks.RepeatDaily}
	in.Items = items("alongar", "correr")
	created := mustCreate(t, s, in)

	todayDay := tasks.StartOfDay(time.Now())
	tomorrow := todayDay.AddDate(0, 0, 1)

	for _, it := range created.Items {
		if err := s.SetItemDone(ctx, it.ID, &todayDay, true); err != nil {
			t.Fatalf("SetItemDone: %v", err)
		}
	}

	occToday, ok := findInAgenda(t, s, created.ID, todayDay)
	if !ok || !occToday.Done() || !occToday.Items[0].Done {
		t.Errorf("today's occurrence = %+v, want done with checked items", occToday)
	}
	occTomorrow, ok := findInAgenda(t, s, created.ID, tomorrow)
	if !ok || occTomorrow.Done() || occTomorrow.Items[0].Done {
		t.Errorf("tomorrow's occurrence = %+v, want untouched", occTomorrow)
	}

	if err := s.SetItemDone(ctx, created.Items[1].ID, &todayDay, false); err != nil {
		t.Fatal(err)
	}
	if occToday, _ := findInAgenda(t, s, created.ID, todayDay); occToday.Done() {
		t.Error("unchecking an item should reopen the occurrence")
	}
}

func TestService_RecurringSetDoneValidatesDay(t *testing.T) {
	ctx := context.Background()
	s, _ := newService(t)
	in := input(today(7))
	in.Repeat = tasks.Repeat{Kind: tasks.RepeatDaily}
	created := mustCreate(t, s, in)

	if err := s.SetDone(ctx, created.ID, nil, true); !errors.Is(err, tasks.ErrOccurrenceRequired) {
		t.Errorf("without day: err = %v, want ErrOccurrenceRequired", err)
	}
	yesterday := tasks.StartOfDay(time.Now()).AddDate(0, 0, -1)
	if err := s.SetDone(ctx, created.ID, &yesterday, true); !errors.Is(err, tasks.ErrInvalidOccurrence) {
		t.Errorf("before the series: err = %v, want ErrInvalidOccurrence", err)
	}

	todayDay := tasks.StartOfDay(time.Now())
	if err := s.SetDone(ctx, created.ID, &todayDay, true); err != nil {
		t.Fatalf("SetDone today: %v", err)
	}
	if occ, _ := findInAgenda(t, s, created.ID, todayDay); !occ.Done() {
		t.Error("today's occurrence should be done")
	}
}

func TestService_AgendaFilter(t *testing.T) {
	ctx := context.Background()
	s, _ := newService(t)

	work := input(time.Now().Add(time.Hour))
	work.Category = "Trabalho"
	work.Urgency = tasks.UrgencyHigh
	w := mustCreate(t, s, work)

	home := input(time.Now().Add(time.Hour))
	home.Category = "Casa"
	home.Urgency = tasks.UrgencyHigh
	mustCreate(t, s, home)

	agenda, err := s.Agenda(ctx, time.Now(), tasks.Filter{Category: "trabalho", Urgency: tasks.UrgencyHigh})
	if err != nil {
		t.Fatal(err)
	}
	var ids []int64
	for _, d := range agenda.Days {
		for _, tk := range d.Tasks {
			ids = append(ids, tk.ID)
		}
	}
	if len(ids) == 0 || slices.ContainsFunc(ids, func(id int64) bool { return id != w.ID }) {
		t.Errorf("filtered ids = %v, want only %d", ids, w.ID)
	}
}
