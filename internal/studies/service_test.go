package studies_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"dashboard/internal/studies"
	"dashboard/internal/testutil"
)

var ctx = context.Background()

func newService(t *testing.T) *studies.Service {
	return studies.NewService(testutil.Tx(t))
}

func mustSubject(t *testing.T, s *studies.Service, name string) studies.Subject {
	t.Helper()
	sub, err := s.CreateSubject(ctx, name)
	if err != nil {
		t.Fatalf("CreateSubject(%q): %v", name, err)
	}
	return sub
}

func mustTopic(t *testing.T, s *studies.Service, subjectID int64, title string) studies.Topic {
	t.Helper()
	topic, err := s.CreateTopic(ctx, subjectID, title)
	if err != nil {
		t.Fatalf("CreateTopic(%q): %v", title, err)
	}
	return topic
}

func overview(t *testing.T, s *studies.Service, now time.Time) studies.Overview {
	t.Helper()
	ov, err := s.Overview(ctx, now)
	if err != nil {
		t.Fatalf("Overview: %v", err)
	}
	return ov
}

func findSubject(ov studies.Overview, id int64) (studies.Subject, bool) {
	for _, sub := range ov.Subjects {
		if sub.ID == id {
			return sub, true
		}
	}
	return studies.Subject{}, false
}

// thisWeek returns a time inside the current week (Monday 09:00 + offset).
func thisWeek(offset time.Duration) time.Time {
	return studies.WeekStart(time.Now()).Add(9*time.Hour + offset).Truncate(time.Microsecond)
}

func TestSubjects_NameRules(t *testing.T) {
	s := newService(t)
	calc := mustSubject(t, s, "  Cálculo ")
	if calc.Name != "Cálculo" {
		t.Errorf("Name = %q, want trimmed", calc.Name)
	}

	if _, err := s.CreateSubject(ctx, "cálculo"); !errors.Is(err, studies.ErrSubjectExists) {
		t.Errorf("duplicate: err = %v, want ErrSubjectExists", err)
	}
	if _, err := s.CreateSubject(ctx, " "); !errors.Is(err, studies.ErrNameRequired) {
		t.Errorf("blank: err = %v, want ErrNameRequired", err)
	}

	// Archived names can be reused, but then the old one can't be restored.
	if err := s.SetSubjectArchived(ctx, calc.ID, true); err != nil {
		t.Fatal(err)
	}
	mustSubject(t, s, "Cálculo")
	if err := s.SetSubjectArchived(ctx, calc.ID, false); !errors.Is(err, studies.ErrSubjectExists) {
		t.Errorf("restore over a taken name: err = %v, want ErrSubjectExists", err)
	}
}

func TestTopics_StatusAndProgressInOverview(t *testing.T) {
	s := newService(t)
	sub := mustSubject(t, s, "Física")
	a := mustTopic(t, s, sub.ID, "Cinemática")
	mustTopic(t, s, sub.ID, "Dinâmica")

	if err := s.SetTopicStatus(ctx, a.ID, studies.StatusReviewed); err != nil {
		t.Fatal(err)
	}
	if err := s.SetTopicStatus(ctx, a.ID, "done"); !errors.Is(err, studies.ErrInvalidStatus) {
		t.Errorf("invalid status: err = %v", err)
	}

	got, ok := findSubject(overview(t, s, time.Now()), sub.ID)
	if !ok || len(got.Topics) != 2 || studies.Progress(got.Topics) != 0.5 {
		t.Errorf("subject = %+v, want 2 topics at 50%%", got)
	}
}

func TestTimer_FullCycle(t *testing.T) {
	s := newService(t)
	sub := mustSubject(t, s, "Química")
	topic := mustTopic(t, s, sub.ID, "Estequiometria")
	start := thisWeek(0)

	session, err := s.Start(ctx, start, sub.ID, &topic.ID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if session.SubjectName != "Química" || session.TopicTitle != "Estequiometria" {
		t.Errorf("session = %+v", session)
	}

	got, _ := findSubject(overview(t, s, start), sub.ID)
	if got.Topics[0].Status != studies.StatusStudying {
		t.Errorf("topic status = %s, want studying after starting the timer", got.Topics[0].Status)
	}

	if _, err := s.Start(ctx, start, sub.ID, nil); !errors.Is(err, studies.ErrSessionRunning) {
		t.Errorf("second Start: err = %v, want ErrSessionRunning", err)
	}

	if _, err := s.Pause(ctx, start.Add(20*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Resume(ctx, start.Add(30*time.Minute)); err != nil {
		t.Fatal(err)
	}

	// The running timer counts towards goals.
	goalID, err := s.CreateGoal(ctx, nil, 60)
	if err != nil {
		t.Fatal(err)
	}
	ov := overview(t, s, start.Add(50*time.Minute))
	if ov.Active == nil || ov.Goals[0].ID != goalID || ov.Goals[0].DoneMinutes != 40 {
		t.Errorf("active = %+v, goals = %+v, want 40 min done", ov.Active, ov.Goals)
	}

	stopped, saved, err := s.Stop(ctx, start.Add(80*time.Minute))
	if err != nil || !saved {
		t.Fatalf("Stop: saved = %v, err = %v", saved, err)
	}
	if got := stopped.Elapsed(time.Now()); got != 70*time.Minute {
		t.Errorf("Elapsed = %v, want 70m", got)
	}

	ov = overview(t, s, start.Add(2*time.Hour))
	if ov.Active != nil || len(ov.Sessions) != 1 || !ov.Goals[0].Completed {
		t.Errorf("after stop: active = %+v, sessions = %d, goals = %+v", ov.Active, len(ov.Sessions), ov.Goals)
	}
}

func TestTimer_DiscardsShortSessions(t *testing.T) {
	s := newService(t)
	sub := mustSubject(t, s, "História")
	start := thisWeek(0)

	if _, err := s.Start(ctx, start, sub.ID, nil); err != nil {
		t.Fatal(err)
	}
	_, saved, err := s.Stop(ctx, start.Add(59*time.Second))
	if err != nil || saved {
		t.Fatalf("Stop: saved = %v, err = %v, want discarded", saved, err)
	}
	if ov := overview(t, s, start); len(ov.Sessions) != 0 || ov.Active != nil {
		t.Errorf("sessions = %+v, want none", ov.Sessions)
	}
	if _, _, err := s.Stop(ctx, start); !errors.Is(err, studies.ErrNoActiveSession) {
		t.Errorf("Stop without timer: err = %v", err)
	}
}

func TestTimer_RejectsTopicOfAnotherSubject(t *testing.T) {
	s := newService(t)
	a := mustSubject(t, s, "Inglês")
	b := mustSubject(t, s, "Espanhol")
	topic := mustTopic(t, s, b.ID, "Verbos")

	if _, err := s.Start(ctx, thisWeek(0), a.ID, &topic.ID); !errors.Is(err, studies.ErrTopicOtherSubject) {
		t.Errorf("err = %v, want ErrTopicOtherSubject", err)
	}
}

func TestArchivedSubject(t *testing.T) {
	s := newService(t)
	sub := mustSubject(t, s, "Biologia")
	start := thisWeek(0)

	if _, err := s.Start(ctx, start, sub.ID, nil); err != nil {
		t.Fatal(err)
	}
	if err := s.SetSubjectArchived(ctx, sub.ID, true); !errors.Is(err, studies.ErrArchiveRunning) {
		t.Errorf("archive while running: err = %v, want ErrArchiveRunning", err)
	}
	if _, _, err := s.Stop(ctx, start.Add(30*time.Minute)); err != nil {
		t.Fatal(err)
	}

	overall, _ := s.CreateGoal(ctx, nil, 60)
	if _, err := s.CreateGoal(ctx, &sub.ID, 60); err != nil {
		t.Fatal(err)
	}
	if err := s.SetSubjectArchived(ctx, sub.ID, true); err != nil {
		t.Fatal(err)
	}

	ov := overview(t, s, start.Add(time.Hour))
	if _, ok := findSubject(ov, sub.ID); ok || len(ov.Archived) != 1 {
		t.Errorf("subjects = %+v, archived = %+v", ov.Subjects, ov.Archived)
	}
	if len(ov.Goals) != 1 || ov.Goals[0].ID != overall || ov.Goals[0].DoneMinutes != 30 {
		t.Errorf("goals = %+v, want only the overall goal with 30 min", ov.Goals)
	}
	if _, err := s.Start(ctx, start, sub.ID, nil); !errors.Is(err, studies.ErrSubjectArchived) {
		t.Errorf("Start archived: err = %v", err)
	}

	if err := s.SetSubjectArchived(ctx, sub.ID, false); err != nil {
		t.Fatal(err)
	}
	if ov := overview(t, s, start.Add(time.Hour)); len(ov.Goals) != 2 {
		t.Errorf("after restoring: goals = %+v, want the subject goal back", ov.Goals)
	}
}

func TestDeleteTopicKeepsSessions(t *testing.T) {
	s := newService(t)
	sub := mustSubject(t, s, "Artes")
	topic := mustTopic(t, s, sub.ID, "Barroco")
	start := thisWeek(0)

	if _, err := s.Start(ctx, start, sub.ID, &topic.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Stop(ctx, start.Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteTopic(ctx, topic.ID); err != nil {
		t.Fatal(err)
	}

	ov := overview(t, s, start)
	if len(ov.Sessions) != 1 || ov.Sessions[0].TopicID != nil {
		t.Errorf("sessions = %+v, want the session kept without topic", ov.Sessions)
	}
	if err := s.DeleteSession(ctx, ov.Sessions[0].ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteSession(ctx, ov.Sessions[0].ID); !errors.Is(err, studies.ErrSessionNotFound) {
		t.Errorf("second DeleteSession: err = %v", err)
	}
}

func TestGoals_Validation(t *testing.T) {
	s := newService(t)

	for _, minutes := range []int{0, studies.MaxGoalMinutes + 1} {
		if _, err := s.CreateGoal(ctx, nil, minutes); !errors.Is(err, studies.ErrInvalidGoalTarget) {
			t.Errorf("CreateGoal(%d): err = %v", minutes, err)
		}
	}
	missing := int64(-1)
	if _, err := s.CreateGoal(ctx, &missing, 60); !errors.Is(err, studies.ErrSubjectNotFound) {
		t.Errorf("unknown subject: err = %v", err)
	}

	id, err := s.CreateGoal(ctx, nil, 60)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateGoal(ctx, id, 120); err != nil {
		t.Fatal(err)
	}
	if ov := overview(t, s, time.Now()); ov.Goals[0].TargetMinutes != 120 {
		t.Errorf("target = %d, want 120", ov.Goals[0].TargetMinutes)
	}
	if err := s.DeleteGoal(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteGoal(ctx, id); !errors.Is(err, studies.ErrGoalNotFound) {
		t.Errorf("second DeleteGoal: err = %v", err)
	}
}
