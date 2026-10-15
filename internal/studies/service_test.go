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

	if err := s.SetTopicStatus(ctx, time.Now(), a.ID, studies.StatusReviewed); err != nil {
		t.Fatal(err)
	}
	if err := s.SetTopicStatus(ctx, time.Now(), a.ID, "done"); !errors.Is(err, studies.ErrInvalidStatus) {
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

// dayKey identifies a calendar day; Postgres dates come back as UTC midnight.
func dayKey(t time.Time) string { return t.Format(time.DateOnly) }

func findTopic(t *testing.T, s *studies.Service, now time.Time, id int64) studies.Topic {
	t.Helper()
	for _, sub := range overview(t, s, now).Subjects {
		for _, topic := range sub.Topics {
			if topic.ID == id {
				return topic
			}
		}
	}
	t.Fatalf("topic %d not found", id)
	return studies.Topic{}
}

func dueIDs(t *testing.T, s *studies.Service, now time.Time) []int64 {
	t.Helper()
	due, err := s.DueReviews(ctx, now)
	if err != nil {
		t.Fatalf("DueReviews: %v", err)
	}
	ids := []int64{}
	for _, d := range due {
		ids = append(ids, d.ID)
	}
	return ids
}

func TestReviews_StatusDrivesTheQueue(t *testing.T) {
	s := newService(t)
	sub := mustSubject(t, s, "Direito Constitucional")
	topic := mustTopic(t, s, sub.ID, "Direitos fundamentais")
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())

	if err := s.SetTopicStatus(ctx, now, topic.ID, studies.StatusReviewed); err != nil {
		t.Fatal(err)
	}
	got := findTopic(t, s, now, topic.ID).Schedule
	if got.NextReviewOn == nil || dayKey(*got.NextReviewOn) != dayKey(tomorrow) || got.Repetitions != 1 {
		t.Errorf("after reviewed: schedule = %+v, want next review tomorrow", got)
	}

	// Setting "reviewed" again keeps the schedule.
	if err := s.SetTopicStatus(ctx, now.Add(48*time.Hour), topic.ID, studies.StatusReviewed); err != nil {
		t.Fatal(err)
	}
	if again := findTopic(t, s, now, topic.ID).Schedule; dayKey(*again.NextReviewOn) != dayKey(tomorrow) {
		t.Errorf("re-setting reviewed moved the schedule to %v", again.NextReviewOn)
	}

	if err := s.SetTopicStatus(ctx, now, topic.ID, studies.StatusStudying); err != nil {
		t.Fatal(err)
	}
	if cleared := findTopic(t, s, now, topic.ID).Schedule; cleared.NextReviewOn != nil || cleared.Repetitions != 0 {
		t.Errorf("after leaving reviewed: schedule = %+v, want cleared", cleared)
	}
	if _, err := s.Review(ctx, now, topic.ID, studies.GradeOK); !errors.Is(err, studies.ErrNotScheduled) {
		t.Errorf("Review unscheduled: err = %v, want ErrNotScheduled", err)
	}
}

func TestReviews_DueListAndGrading(t *testing.T) {
	s := newService(t)
	sub := mustSubject(t, s, "Português")
	early := mustTopic(t, s, sub.ID, "Crase")
	late := mustTopic(t, s, sub.ID, "Regência")
	notYet := mustTopic(t, s, sub.ID, "Concordância")
	now := time.Now()

	// Reviewed on different days: Crase 5 days ago, Regência 2 days ago, Concordância today.
	for topic, daysAgo := range map[int64]int{early.ID: 5, late.ID: 2, notYet.ID: 0} {
		if err := s.SetTopicStatus(ctx, now.AddDate(0, 0, -daysAgo), topic, studies.StatusReviewed); err != nil {
			t.Fatal(err)
		}
	}

	due, err := s.DueReviews(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(due) != 2 || due[0].ID != early.ID || due[1].ID != late.ID {
		t.Fatalf("due = %+v, want [Crase, Regência]", due)
	}
	if due[0].DaysOverdue != 4 || due[1].DaysOverdue != 1 || due[0].SubjectName != "Português" {
		t.Errorf("overdue days = %d, %d; want 4, 1", due[0].DaysOverdue, due[1].DaysOverdue)
	}

	next, err := s.Review(ctx, now, early.ID, studies.GradeOK)
	if err != nil {
		t.Fatal(err)
	}
	if next.IntervalDays != 6 || next.Repetitions != 2 {
		t.Errorf("after OK: %+v, want 6 days", next)
	}
	if got := dueIDs(t, s, now); len(got) != 1 || got[0] != late.ID {
		t.Errorf("due after review = %v, want only Regência", got)
	}

	if _, err := s.Review(ctx, now, late.ID, "great"); !errors.Is(err, studies.ErrInvalidGrade) {
		t.Errorf("invalid grade: err = %v", err)
	}

	// Archived subjects leave the queue.
	if err := s.SetSubjectArchived(ctx, sub.ID, true); err != nil {
		t.Fatal(err)
	}
	if got := dueIDs(t, s, now); len(got) != 0 {
		t.Errorf("due with archived subject = %v, want none", got)
	}
}

func TestQuestions_LogAndReport(t *testing.T) {
	s := newService(t)
	pt := mustSubject(t, s, "Português")
	crase := mustTopic(t, s, pt.ID, "Crase")
	dir := mustSubject(t, s, "Direito")
	other := mustTopic(t, s, dir.ID, "Atos")
	now := time.Now()
	daysAgo := func(n int) time.Time { return now.AddDate(0, 0, -n) }

	logs := []studies.QuestionLog{
		{SubjectID: pt.ID, TopicID: &crase.ID, DoneOn: daysAgo(0), Total: 10, Correct: 3, Source: "QConcursos"},
		{SubjectID: pt.ID, DoneOn: daysAgo(10), Total: 20, Correct: 15},
		{SubjectID: pt.ID, TopicID: &crase.ID, DoneOn: daysAgo(60), Total: 10, Correct: 10},
		{SubjectID: dir.ID, DoneOn: daysAgo(1), Total: 5, Correct: 5},
	}
	for _, l := range logs {
		if _, err := s.LogQuestions(ctx, now, l); err != nil {
			t.Fatalf("LogQuestions(%+v): %v", l, err)
		}
	}

	last30, err := s.QuestionReport(ctx, now, 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(last30.Logs) != 3 || last30.Logs[0].Source != "QConcursos" {
		t.Errorf("30-day logs = %+v", last30.Logs)
	}
	if len(last30.ByTopic) != 1 || last30.ByTopic[0].Total != 10 || last30.ByTopic[0].Correct != 3 {
		t.Errorf("30-day topics = %+v", last30.ByTopic)
	}
	if len(last30.Weaknesses) != 1 || last30.Weaknesses[0].Name != "Crase" {
		t.Errorf("weaknesses = %+v, want Crase", last30.Weaknesses)
	}

	all, _ := s.QuestionReport(ctx, now, 0)
	if len(all.Logs) != 4 || all.ByTopic[0].Total != 20 {
		t.Errorf("all-time report = %+v", all)
	}

	// Archived subjects leave the report.
	if err := s.SetSubjectArchived(ctx, dir.ID, true); err != nil {
		t.Fatal(err)
	}
	if all, _ := s.QuestionReport(ctx, now, 0); len(all.BySubject) != 1 {
		t.Errorf("subjects with Direito archived = %+v", all.BySubject)
	}

	if err := s.DeleteQuestionLog(ctx, all.Logs[0].ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteQuestionLog(ctx, all.Logs[0].ID); !errors.Is(err, studies.ErrQuestionLogNotFound) {
		t.Errorf("second delete: err = %v", err)
	}

	bad := []struct {
		log  studies.QuestionLog
		want error
	}{
		{studies.QuestionLog{SubjectID: pt.ID, TopicID: &other.ID, Total: 1}, studies.ErrTopicOtherSubject},
		{studies.QuestionLog{SubjectID: pt.ID, DoneOn: now.AddDate(0, 0, 1), Total: 1}, studies.ErrFutureDate},
		{studies.QuestionLog{SubjectID: pt.ID, Total: 5, Correct: 6}, studies.ErrCorrectOverTotal},
		{studies.QuestionLog{SubjectID: dir.ID, Total: 1}, studies.ErrSubjectArchived},
	}
	for _, b := range bad {
		if _, err := s.LogQuestions(ctx, now, b.log); !errors.Is(err, b.want) {
			t.Errorf("LogQuestions(%+v): err = %v, want %v", b.log, err, b.want)
		}
	}
}
