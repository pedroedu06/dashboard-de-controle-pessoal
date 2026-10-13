package tasks

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"dashboard/internal/db/sqlc"
)

type Service struct {
	q *sqlc.Queries
}

func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

func (s *Service) Create(ctx context.Context, in Input) (Task, error) {
	in, err := in.Normalize()
	if err != nil {
		return Task{}, err
	}
	row, err := s.q.CreateTask(ctx, sqlc.CreateTaskParams{
		Title:    in.Title,
		Urgency:  sqlc.TaskUrgency(in.Urgency),
		StartsAt: in.StartsAt,
		EndsAt:   in.EndsAt,
	})
	if err != nil {
		return Task{}, err
	}
	return fromRow(row), nil
}

func (s *Service) Update(ctx context.Context, id int64, in Input) (Task, error) {
	in, err := in.Normalize()
	if err != nil {
		return Task{}, err
	}
	row, err := s.q.UpdateTask(ctx, sqlc.UpdateTaskParams{
		ID:       id,
		Title:    in.Title,
		Urgency:  sqlc.TaskUrgency(in.Urgency),
		StartsAt: in.StartsAt,
		EndsAt:   in.EndsAt,
	})
	if err != nil {
		return Task{}, notFound(err)
	}
	return fromRow(row), nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	n, err := s.q.DeleteTask(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) SetDone(ctx context.Context, id int64, done bool) (Task, error) {
	row, err := s.q.SetTaskDone(ctx, sqlc.SetTaskDoneParams{ID: id, Done: done})
	if err != nil {
		return Task{}, notFound(err)
	}
	return fromRow(row), nil
}

// Agenda loads the tasks relevant from today onwards and groups them by day.
func (s *Service) Agenda(ctx context.Context, now time.Time) (Agenda, error) {
	rows, err := s.q.ListAgendaTasks(ctx, StartOfDay(now))
	if err != nil {
		return Agenda{}, err
	}
	list := make([]Task, len(rows))
	for i, r := range rows {
		list[i] = fromRow(r)
	}
	return BuildAgenda(list, now), nil
}

func fromRow(r sqlc.Task) Task {
	return Task{
		ID:       r.ID,
		Title:    r.Title,
		Urgency:  Urgency(r.Urgency),
		StartsAt: r.StartsAt,
		EndsAt:   r.EndsAt,
		DoneAt:   r.DoneAt,
	}
}

func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
