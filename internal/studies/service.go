package studies

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"dashboard/internal/db/sqlc"
)

// DB is what the service needs from the database: queries plus transactions.
// Both *pgxpool.Pool and pgx.Tx (nested as a savepoint) satisfy it.
type DB interface {
	sqlc.DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Service struct {
	db DB
}

func NewService(db DB) *Service {
	return &Service{db: db}
}

// Overview is everything the studies page shows.
type Overview struct {
	Subjects []Subject // active, with topics
	Archived []Subject // without topics
	Active   *Session
	Goals    []GoalStatus
	Sessions []Session // finished sessions of this week, newest first
}

func (s *Service) Overview(ctx context.Context, now time.Time) (Overview, error) {
	q := sqlc.New(s.db)

	rows, err := q.ListSubjects(ctx, true)
	if err != nil {
		return Overview{}, err
	}
	ov := Overview{Subjects: []Subject{}, Archived: []Subject{}}
	var activeIDs []int64
	for _, r := range rows {
		if r.ArchivedAt != nil {
			ov.Archived = append(ov.Archived, Subject{ID: r.ID, Name: r.Name, Archived: true, Topics: []Topic{}})
			continue
		}
		ov.Subjects = append(ov.Subjects, Subject{ID: r.ID, Name: r.Name, Topics: []Topic{}})
		activeIDs = append(activeIDs, r.ID)
	}

	topics, err := q.ListTopicsBySubjects(ctx, activeIDs)
	if err != nil {
		return Overview{}, err
	}
	index := map[int64]int{}
	for i, sub := range ov.Subjects {
		index[sub.ID] = i
	}
	for _, t := range topics {
		i := index[t.SubjectID]
		ov.Subjects[i].Topics = append(ov.Subjects[i].Topics, fromTopic(t))
	}

	if ov.Active, err = activeSession(ctx, q); err != nil {
		return Overview{}, err
	}

	weekStart := WeekStart(now)
	sessionRows, err := q.ListSessionsBetween(ctx, sqlc.ListSessionsBetweenParams{FromTime: weekStart, ToTime: weekStart.AddDate(0, 0, 7)})
	if err != nil {
		return Overview{}, err
	}
	ov.Sessions = make([]Session, len(sessionRows))
	for i, r := range sessionRows {
		ov.Sessions[i] = Session{
			ID: r.ID, SubjectID: r.SubjectID, SubjectName: r.SubjectName, TopicID: r.TopicID, TopicTitle: deref(r.TopicTitle),
			StartedAt: r.StartedAt, EndedAt: r.EndedAt, PausedAt: r.PausedAt, PausedSeconds: int(r.PausedSeconds),
		}
	}

	goalRows, err := q.ListGoals(ctx)
	if err != nil {
		return Overview{}, err
	}
	goals := make([]Goal, len(goalRows))
	for i, g := range goalRows {
		goals[i] = Goal{ID: g.ID, SubjectID: g.SubjectID, SubjectName: deref(g.SubjectName), TargetMinutes: int(g.TargetMinutes)}
	}
	ov.Goals = GoalProgress(goals, ov.Sessions, ov.Active, now)

	return ov, nil
}

// Subjects

func (s *Service) CreateSubject(ctx context.Context, name string) (Subject, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Subject{}, ErrNameRequired
	}
	var row sqlc.Subject
	err := s.inTx(ctx, func(q *sqlc.Queries) (err error) {
		row, err = q.CreateSubject(ctx, name)
		return err
	})
	if err != nil {
		return Subject{}, uniqueName(err)
	}
	return Subject{ID: row.ID, Name: row.Name, Topics: []Topic{}}, nil
}

func (s *Service) RenameSubject(ctx context.Context, id int64, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrNameRequired
	}
	return s.inTx(ctx, func(q *sqlc.Queries) error {
		n, err := q.RenameSubject(ctx, sqlc.RenameSubjectParams{ID: id, Name: name})
		return affected(n, uniqueName(err), ErrSubjectNotFound)
	})
}

// SetSubjectArchived archives or restores a subject. Restoring fails if an
// active subject took its name meanwhile.
func (s *Service) SetSubjectArchived(ctx context.Context, id int64, archived bool) error {
	return s.inTx(ctx, func(q *sqlc.Queries) error {
		if archived {
			active, err := activeSession(ctx, q)
			if err != nil {
				return err
			}
			if active != nil && active.SubjectID == id {
				return ErrArchiveRunning
			}
		}
		n, err := q.SetSubjectArchived(ctx, sqlc.SetSubjectArchivedParams{ID: id, Archived: archived})
		return affected(n, uniqueName(err), ErrSubjectNotFound)
	})
}

// Topics

func (s *Service) CreateTopic(ctx context.Context, subjectID int64, title string) (Topic, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Topic{}, ErrTitleRequired
	}
	q := sqlc.New(s.db)
	if _, err := activeSubject(ctx, q, subjectID); err != nil {
		return Topic{}, err
	}
	row, err := q.CreateTopic(ctx, sqlc.CreateTopicParams{SubjectID: subjectID, Title: title})
	if err != nil {
		return Topic{}, err
	}
	return fromTopic(row), nil
}

func (s *Service) RenameTopic(ctx context.Context, id int64, title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrTitleRequired
	}
	n, err := sqlc.New(s.db).RenameTopic(ctx, sqlc.RenameTopicParams{ID: id, Title: title})
	return affected(n, err, ErrTopicNotFound)
}

func (s *Service) SetTopicStatus(ctx context.Context, id int64, status Status) error {
	if !status.Valid() {
		return ErrInvalidStatus
	}
	n, err := sqlc.New(s.db).SetTopicStatus(ctx, sqlc.SetTopicStatusParams{ID: id, Status: sqlc.TopicStatus(status)})
	return affected(n, err, ErrTopicNotFound)
}

// DeleteTopic removes a topic; its sessions are kept without a topic.
func (s *Service) DeleteTopic(ctx context.Context, id int64) error {
	n, err := sqlc.New(s.db).DeleteTopic(ctx, id)
	return affected(n, err, ErrTopicNotFound)
}

// Timer

// Start runs the timer for a subject and, optionally, one of its topics. A
// topic not seen yet becomes "studying".
func (s *Service) Start(ctx context.Context, now time.Time, subjectID int64, topicID *int64) (Session, error) {
	var session *Session
	err := s.inTx(ctx, func(q *sqlc.Queries) error {
		if _, err := activeSubject(ctx, q, subjectID); err != nil {
			return err
		}
		if topicID != nil {
			topic, err := q.GetTopic(ctx, *topicID)
			if err != nil {
				return notFound(err, ErrTopicNotFound)
			}
			if topic.SubjectID != subjectID {
				return ErrTopicOtherSubject
			}
			if topic.Status == sqlc.TopicStatusNotSeen {
				if _, err := q.SetTopicStatus(ctx, sqlc.SetTopicStatusParams{ID: topic.ID, Status: sqlc.TopicStatusStudying}); err != nil {
					return err
				}
			}
		}
		if _, err := q.StartSession(ctx, sqlc.StartSessionParams{SubjectID: subjectID, TopicID: topicID, StartedAt: now}); err != nil {
			if isUniqueViolation(err) {
				return ErrSessionRunning
			}
			return err
		}
		var err error
		session, err = activeSession(ctx, q)
		return err
	})
	if err != nil {
		return Session{}, err
	}
	return *session, nil
}

func (s *Service) Pause(ctx context.Context, now time.Time) (Session, error) {
	return s.updateActive(ctx, func(a Session) (Session, error) { return a.Pause(now) })
}

func (s *Service) Resume(ctx context.Context, now time.Time) (Session, error) {
	return s.updateActive(ctx, func(a Session) (Session, error) { return a.Resume(now) })
}

// Stop ends the timer. Sessions shorter than MinSessionDuration are
// discarded, and saved reports whether it was kept.
func (s *Service) Stop(ctx context.Context, now time.Time) (session Session, saved bool, err error) {
	err = s.inTx(ctx, func(q *sqlc.Queries) error {
		active, err := activeSession(ctx, q)
		if err != nil {
			return err
		}
		if active == nil {
			return ErrNoActiveSession
		}
		session, _ = active.Stop(now)
		if session.Elapsed(now) < MinSessionDuration {
			_, err := q.DeleteSession(ctx, session.ID)
			return err
		}
		saved = true
		return saveTimer(ctx, q, session)
	})
	return session, saved, err
}

func (s *Service) DeleteSession(ctx context.Context, id int64) error {
	n, err := sqlc.New(s.db).DeleteSession(ctx, id)
	return affected(n, err, ErrSessionNotFound)
}

// Goals

// CreateGoal adds a weekly goal; subjectID nil means all subjects.
func (s *Service) CreateGoal(ctx context.Context, subjectID *int64, targetMinutes int) (int64, error) {
	if targetMinutes < 1 || targetMinutes > MaxGoalMinutes {
		return 0, ErrInvalidGoalTarget
	}
	q := sqlc.New(s.db)
	if subjectID != nil {
		if _, err := activeSubject(ctx, q, *subjectID); err != nil {
			return 0, err
		}
	}
	return q.CreateGoal(ctx, sqlc.CreateGoalParams{SubjectID: subjectID, TargetMinutes: int32(targetMinutes)})
}

func (s *Service) UpdateGoal(ctx context.Context, id int64, targetMinutes int) error {
	if targetMinutes < 1 || targetMinutes > MaxGoalMinutes {
		return ErrInvalidGoalTarget
	}
	n, err := sqlc.New(s.db).UpdateGoalTarget(ctx, sqlc.UpdateGoalTargetParams{ID: id, TargetMinutes: int32(targetMinutes)})
	return affected(n, err, ErrGoalNotFound)
}

func (s *Service) DeleteGoal(ctx context.Context, id int64) error {
	n, err := sqlc.New(s.db).DeleteGoal(ctx, id)
	return affected(n, err, ErrGoalNotFound)
}

// HoursToMinutes converts a goal given in hours, rounding to whole minutes.
func HoursToMinutes(hours float64) int {
	return int(math.Round(hours * 60))
}

// helpers

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

func (s *Service) updateActive(ctx context.Context, change func(Session) (Session, error)) (Session, error) {
	var session Session
	err := s.inTx(ctx, func(q *sqlc.Queries) error {
		active, err := activeSession(ctx, q)
		if err != nil {
			return err
		}
		if active == nil {
			return ErrNoActiveSession
		}
		if session, err = change(*active); err != nil {
			return err
		}
		return saveTimer(ctx, q, session)
	})
	return session, err
}

func saveTimer(ctx context.Context, q *sqlc.Queries, s Session) error {
	return q.UpdateSessionTimer(ctx, sqlc.UpdateSessionTimerParams{
		ID: s.ID, PausedAt: s.PausedAt, PausedSeconds: int32(s.PausedSeconds), EndedAt: s.EndedAt,
	})
}

func activeSession(ctx context.Context, q *sqlc.Queries) (*Session, error) {
	r, err := q.GetActiveSession(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &Session{
		ID: r.ID, SubjectID: r.SubjectID, SubjectName: r.SubjectName, TopicID: r.TopicID, TopicTitle: deref(r.TopicTitle),
		StartedAt: r.StartedAt, EndedAt: r.EndedAt, PausedAt: r.PausedAt, PausedSeconds: int(r.PausedSeconds),
	}, nil
}

func activeSubject(ctx context.Context, q *sqlc.Queries, id int64) (sqlc.Subject, error) {
	sub, err := q.GetSubject(ctx, id)
	if err != nil {
		return sqlc.Subject{}, notFound(err, ErrSubjectNotFound)
	}
	if sub.ArchivedAt != nil {
		return sqlc.Subject{}, ErrSubjectArchived
	}
	return sub, nil
}

func fromTopic(t sqlc.Topic) Topic {
	return Topic{ID: t.ID, SubjectID: t.SubjectID, Title: t.Title, Status: Status(t.Status)}
}

// affected maps "no rows changed" to notFoundErr.
func affected(n int64, err, notFoundErr error) error {
	if err != nil {
		return err
	}
	if n == 0 {
		return notFoundErr
	}
	return nil
}

func notFound(err, target error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return target
	}
	return err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func uniqueName(err error) error {
	if isUniqueViolation(err) {
		return ErrSubjectExists
	}
	return err
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
