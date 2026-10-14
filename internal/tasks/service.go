package tasks

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"dashboard/internal/db/sqlc"
)

// DB is what the service needs from the database: queries plus transactions.
// Both *pgxpool.Pool and pgx.Tx (nested as a savepoint) satisfy it.
type DB interface {
	sqlc.DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Service struct {
	db  DB
	loc *time.Location // location used for calendar days
}

func NewService(db DB) *Service {
	return &Service{db: db, loc: time.Local}
}

func (s *Service) Create(ctx context.Context, in Input) (Task, error) {
	in, err := in.Normalize()
	if err != nil {
		return Task{}, err
	}
	var task Task
	err = s.inTx(ctx, func(q *sqlc.Queries) error {
		category, err := canonicalCategory(ctx, q, in.Category)
		if err != nil {
			return err
		}
		row, err := q.CreateTask(ctx, sqlc.CreateTaskParams{
			Title:          in.Title,
			Urgency:        sqlc.TaskUrgency(in.Urgency),
			StartsAt:       in.StartsAt,
			EndsAt:         in.EndsAt,
			Category:       category,
			Repeat:         sqlc.TaskRepeat(in.Repeat.Kind),
			RepeatWeekdays: weekdaysToDB(in.Repeat.Weekdays),
			RepeatUntil:    in.Repeat.Until,
		})
		if err != nil {
			return err
		}
		if err := syncItems(ctx, q, row.ID, in.Items); err != nil {
			return err
		}
		task, err = s.saved(ctx, q, row.ID)
		return err
	})
	return task, err
}

// Update edits the task (the whole series, for recurring tasks).
func (s *Service) Update(ctx context.Context, id int64, in Input) (Task, error) {
	in, err := in.Normalize()
	if err != nil {
		return Task{}, err
	}
	var task Task
	err = s.inTx(ctx, func(q *sqlc.Queries) error {
		category, err := canonicalCategory(ctx, q, in.Category)
		if err != nil {
			return err
		}
		_, err = q.UpdateTask(ctx, sqlc.UpdateTaskParams{
			ID:             id,
			Title:          in.Title,
			Urgency:        sqlc.TaskUrgency(in.Urgency),
			StartsAt:       in.StartsAt,
			EndsAt:         in.EndsAt,
			Category:       category,
			Repeat:         sqlc.TaskRepeat(in.Repeat.Kind),
			RepeatWeekdays: weekdaysToDB(in.Repeat.Weekdays),
			RepeatUntil:    in.Repeat.Until,
		})
		if err != nil {
			return notFound(err, ErrNotFound)
		}
		if err := syncItems(ctx, q, id, in.Items); err != nil {
			return err
		}
		task, err = s.saved(ctx, q, id)
		return err
	})
	return task, err
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	n, err := sqlc.New(s.db).DeleteTask(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// SetDone completes a task without items. occursOn is required for
// recurring tasks and must be nil otherwise.
func (s *Service) SetDone(ctx context.Context, id int64, occursOn *time.Time, done bool) error {
	return s.inTx(ctx, func(q *sqlc.Queries) error {
		task, err := s.load(ctx, q, id)
		if err != nil {
			return err
		}
		if len(task.Items) > 0 {
			return ErrDoneByItems
		}
		if !task.Repeat.Recurring() {
			if occursOn != nil {
				return ErrUnexpectedOccurrence
			}
			_, err := q.SetTaskDone(ctx, sqlc.SetTaskDoneParams{ID: id, Done: done})
			return err
		}
		day, err := s.occurrenceDay(task, occursOn)
		if err != nil {
			return err
		}
		return setOccurrenceDone(ctx, q, id, day, done)
	})
}

// SetItemDone checks an item and completes or reopens its task when all
// items are (or stop being) checked.
func (s *Service) SetItemDone(ctx context.Context, itemID int64, occursOn *time.Time, done bool) error {
	return s.inTx(ctx, func(q *sqlc.Queries) error {
		item, err := q.GetItem(ctx, itemID)
		if err != nil {
			return notFound(err, ErrItemNotFound)
		}
		task, err := s.load(ctx, q, item.TaskID)
		if err != nil {
			return err
		}

		if !task.Repeat.Recurring() {
			if occursOn != nil {
				return ErrUnexpectedOccurrence
			}
			if err := q.SetItemDone(ctx, sqlc.SetItemDoneParams{ID: itemID, Done: done}); err != nil {
				return err
			}
			return refreshDone(ctx, q, task.ID)
		}

		day, err := s.occurrenceDay(task, occursOn)
		if err != nil {
			return err
		}
		if done {
			err = q.InsertItemCheck(ctx, sqlc.InsertItemCheckParams{ItemID: itemID, OccursOn: day})
		} else {
			err = q.DeleteItemCheck(ctx, sqlc.DeleteItemCheckParams{ItemID: itemID, OccursOn: day})
		}
		if err != nil {
			return err
		}
		checks, err := q.ListItemChecks(ctx, sqlc.ListItemChecksParams{ItemIds: itemIDs(task.Items), FromDay: day, ToDay: day})
		if err != nil {
			return err
		}
		return setOccurrenceDone(ctx, q, task.ID, day, len(checks) == len(task.Items))
	})
}

// Agenda loads the tasks relevant from today onwards and groups them by day.
func (s *Service) Agenda(ctx context.Context, now time.Time, filter Filter) (Agenda, error) {
	q := sqlc.New(s.db)
	today := StartOfDay(now)
	lastDay := today.AddDate(0, 0, RecurringHorizonDays-1)

	rows, err := q.ListAgendaTasks(ctx, today)
	if err != nil {
		return Agenda{}, err
	}
	list, err := withItems(ctx, q, rows)
	if err != nil {
		return Agenda{}, err
	}

	state := OccurrenceState{Done: map[OccurrenceKey]bool{}, ItemChecks: map[OccurrenceKey]bool{}}
	var taskIDs, items []int64
	for _, t := range list {
		if t.Repeat.Recurring() {
			taskIDs = append(taskIDs, t.ID)
			items = append(items, itemIDs(t.Items)...)
		}
	}
	occurrences, err := q.ListOccurrences(ctx, sqlc.ListOccurrencesParams{TaskIds: taskIDs, FromDay: today, ToDay: lastDay})
	if err != nil {
		return Agenda{}, err
	}
	for _, o := range occurrences {
		state.Done[OccurrenceKey{o.TaskID, DayKey(o.OccursOn)}] = true
	}
	checks, err := q.ListItemChecks(ctx, sqlc.ListItemChecksParams{ItemIds: items, FromDay: today, ToDay: lastDay})
	if err != nil {
		return Agenda{}, err
	}
	for _, c := range checks {
		state.ItemChecks[OccurrenceKey{c.ItemID, DayKey(c.OccursOn)}] = true
	}

	return BuildAgenda(list, state, now, filter), nil
}

// Categories lists the categories in use, one spelling per name.
func (s *Service) Categories(ctx context.Context) ([]string, error) {
	return sqlc.New(s.db).ListCategories(ctx)
}

func (s *Service) inTx(ctx context.Context, fn func(q *sqlc.Queries) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(sqlc.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) load(ctx context.Context, q *sqlc.Queries, id int64) (Task, error) {
	row, err := q.GetTask(ctx, id)
	if err != nil {
		return Task{}, notFound(err, ErrNotFound)
	}
	list, err := withItems(ctx, q, []sqlc.Task{row})
	if err != nil {
		return Task{}, err
	}
	return list[0], nil
}

// saved reloads a task after a write, refreshing its completion from items.
func (s *Service) saved(ctx context.Context, q *sqlc.Queries, id int64) (Task, error) {
	if err := refreshDone(ctx, q, id); err != nil {
		return Task{}, err
	}
	return s.load(ctx, q, id)
}

// occurrenceDay validates that occursOn is an occurrence of the task.
func (s *Service) occurrenceDay(task Task, occursOn *time.Time) (time.Time, error) {
	if occursOn == nil {
		return time.Time{}, ErrOccurrenceRequired
	}
	day := CivilDay(*occursOn, s.loc)
	if !task.OccursOnDay(day) {
		return time.Time{}, ErrInvalidOccurrence
	}
	return day, nil
}

// canonicalCategory reuses an existing spelling of the category, if any.
func canonicalCategory(ctx context.Context, q *sqlc.Queries, name string) (*string, error) {
	if name == "" {
		return nil, nil
	}
	existing, err := q.FindCategory(ctx, name)
	if errors.Is(err, pgx.ErrNoRows) {
		return &name, nil
	}
	if err != nil {
		return nil, err
	}
	return &existing, nil
}

// syncItems makes the task's items match the form: removed items are
// deleted, kept items are renamed and reordered, new ones are inserted.
func syncItems(ctx context.Context, q *sqlc.Queries, taskID int64, items []ItemInput) error {
	keep := []int64{}
	for _, it := range items {
		if it.ID != 0 {
			keep = append(keep, it.ID)
		}
	}
	if err := q.DeleteItemsNotIn(ctx, sqlc.DeleteItemsNotInParams{TaskID: taskID, KeepIds: keep}); err != nil {
		return err
	}

	for pos, it := range items {
		if it.ID == 0 {
			if _, err := q.InsertItem(ctx, sqlc.InsertItemParams{TaskID: taskID, Title: it.Title, Position: int32(pos)}); err != nil {
				return err
			}
			continue
		}
		n, err := q.UpdateItem(ctx, sqlc.UpdateItemParams{ID: it.ID, TaskID: taskID, Title: it.Title, Position: int32(pos)})
		if err != nil {
			return err
		}
		if n == 0 {
			return ErrItemNotFound
		}
	}
	return nil
}

// refreshDone completes a one-off task with items when all are checked, and
// reopens it otherwise. Tasks without items are left alone.
func refreshDone(ctx context.Context, q *sqlc.Queries, taskID int64) error {
	task, err := q.GetTask(ctx, taskID)
	if err != nil {
		return notFound(err, ErrNotFound)
	}
	if task.Repeat != sqlc.TaskRepeatNone {
		return nil
	}
	items, err := q.ListItemsByTasks(ctx, []int64{taskID})
	if err != nil || len(items) == 0 {
		return err
	}
	allDone := true
	for _, it := range items {
		allDone = allDone && it.DoneAt != nil
	}
	_, err = q.SetTaskDone(ctx, sqlc.SetTaskDoneParams{ID: taskID, Done: allDone})
	return err
}

func setOccurrenceDone(ctx context.Context, q *sqlc.Queries, taskID int64, day time.Time, done bool) error {
	if done {
		return q.UpsertOccurrence(ctx, sqlc.UpsertOccurrenceParams{TaskID: taskID, OccursOn: day})
	}
	return q.DeleteOccurrence(ctx, sqlc.DeleteOccurrenceParams{TaskID: taskID, OccursOn: day})
}

func withItems(ctx context.Context, q *sqlc.Queries, rows []sqlc.Task) ([]Task, error) {
	ids := make([]int64, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	itemRows, err := q.ListItemsByTasks(ctx, ids)
	if err != nil {
		return nil, err
	}
	byTask := map[int64][]Item{}
	for _, it := range itemRows {
		byTask[it.TaskID] = append(byTask[it.TaskID], Item{ID: it.ID, Title: it.Title, Done: it.DoneAt != nil})
	}

	list := make([]Task, len(rows))
	for i, r := range rows {
		list[i] = fromRow(r, byTask[r.ID])
	}
	return list, nil
}

func fromRow(r sqlc.Task, items []Item) Task {
	t := Task{
		ID:       r.ID,
		Title:    r.Title,
		Urgency:  Urgency(r.Urgency),
		StartsAt: r.StartsAt,
		EndsAt:   r.EndsAt,
		DoneAt:   r.DoneAt,
		Repeat:   Repeat{Kind: RepeatKind(r.Repeat), Until: r.RepeatUntil},
		Items:    items,
	}
	if items == nil {
		t.Items = []Item{}
	}
	if r.Category != nil {
		t.Category = *r.Category
	}
	for _, d := range r.RepeatWeekdays {
		t.Repeat.Weekdays = append(t.Repeat.Weekdays, time.Weekday(d))
	}
	return t
}

func weekdaysToDB(days []time.Weekday) []int16 {
	out := make([]int16, len(days))
	for i, d := range days {
		out[i] = int16(d)
	}
	return out
}

func itemIDs(items []Item) []int64 {
	ids := make([]int64, len(items))
	for i, it := range items {
		ids[i] = it.ID
	}
	return ids
}

func notFound(err, target error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return target
	}
	return err
}
