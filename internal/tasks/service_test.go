package tasks_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"dashboard/internal/db/sqlc"
	"dashboard/internal/tasks"
	"dashboard/internal/testutil"
)

func newService(t *testing.T) (*tasks.Service, *sqlc.Queries) {
	q := testutil.Queries(t)
	return tasks.NewService(q), q
}

// input truncates to microseconds, the precision Postgres stores.
func input(start time.Time) tasks.Input {
	start = start.Truncate(time.Microsecond)
	return tasks.Input{Title: "Estudar", Urgency: tasks.UrgencyHigh, StartsAt: start, EndsAt: start.Add(time.Hour)}
}

func TestService_CreateAndAgenda(t *testing.T) {
	ctx := context.Background()
	s, _ := newService(t)
	now := time.Now()

	created, err := s.Create(ctx, input(now.Add(time.Hour)))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 || created.Urgency != tasks.UrgencyHigh || created.Done() {
		t.Errorf("created = %+v", created)
	}

	agenda, err := s.Agenda(ctx, now)
	if err != nil {
		t.Fatalf("Agenda: %v", err)
	}
	found := false
	for _, d := range agenda.Days {
		for _, tk := range d.Tasks {
			found = found || tk.ID == created.ID
		}
	}
	if !found {
		t.Errorf("created task not in agenda: %+v", agenda)
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
	created, err := s.Create(ctx, input(time.Now()))
	if err != nil {
		t.Fatal(err)
	}

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
	s, _ := newService(t)
	created, err := s.Create(ctx, input(time.Now()))
	if err != nil {
		t.Fatal(err)
	}

	got, err := s.SetDone(ctx, created.ID, true)
	if err != nil || !got.Done() {
		t.Fatalf("SetDone(true) = %+v, %v", got, err)
	}
	got, err = s.SetDone(ctx, created.ID, false)
	if err != nil || got.Done() {
		t.Fatalf("SetDone(false) = %+v, %v", got, err)
	}

	if _, err := s.SetDone(ctx, -1, true); !errors.Is(err, tasks.ErrNotFound) {
		t.Errorf("SetDone missing: err = %v, want ErrNotFound", err)
	}
}

func TestService_Delete(t *testing.T) {
	ctx := context.Background()
	s, _ := newService(t)
	created, err := s.Create(ctx, input(time.Now()))
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := s.Delete(ctx, created.ID); !errors.Is(err, tasks.ErrNotFound) {
		t.Errorf("second Delete: err = %v, want ErrNotFound", err)
	}
}

func TestDatabase_RejectsEndBeforeStart(t *testing.T) {
	_, q := newService(t)
	start := time.Now()

	_, err := q.CreateTask(context.Background(), sqlc.CreateTaskParams{
		Title:    "x",
		Urgency:  sqlc.TaskUrgencyLow,
		StartsAt: start,
		EndsAt:   start.Add(-time.Hour),
	})
	if err == nil {
		t.Fatal("expected the database check constraint to reject the row")
	}
}
