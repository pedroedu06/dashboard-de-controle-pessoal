package studies

import (
	"errors"
	"testing"
	"time"
)

var brt = time.FixedZone("BRT", -3*60*60)

// at builds a time on October 2026 in BRT. Oct 12 2026 is a Monday.
func at(day, hour, minute int) time.Time {
	return time.Date(2026, 10, day, hour, minute, 0, 0, brt)
}

func ptr[T any](v T) *T { return &v }

func TestProgress(t *testing.T) {
	topic := func(s Status) Topic { return Topic{Status: s} }
	tests := []struct {
		name   string
		topics []Topic
		want   float64
	}{
		{"no topics", nil, 0},
		{"nothing seen", []Topic{topic(StatusNotSeen), topic(StatusNotSeen)}, 0},
		{"mixed", []Topic{topic(StatusReviewed), topic(StatusStudying), topic(StatusStudying), topic(StatusNotSeen)}, 0.5},
		{"all reviewed", []Topic{topic(StatusReviewed), topic(StatusReviewed)}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Progress(tt.topics); got != tt.want {
				t.Errorf("Progress = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_ElapsedWithPauses(t *testing.T) {
	s := Session{StartedAt: at(13, 10, 0)}

	s, _ = s.Pause(at(13, 10, 20))
	if got := s.Elapsed(at(13, 10, 30)); got != 20*time.Minute {
		t.Errorf("while paused: Elapsed = %v, want 20m (open pause excluded)", got)
	}

	s, _ = s.Resume(at(13, 10, 30))
	s, _ = s.Pause(at(13, 10, 40))
	s, _ = s.Resume(at(13, 10, 45))
	if s.PausedSeconds != 15*60 {
		t.Errorf("PausedSeconds = %d, want 900", s.PausedSeconds)
	}
	if got := s.Elapsed(at(13, 11, 0)); got != 45*time.Minute {
		t.Errorf("running: Elapsed = %v, want 45m", got)
	}
}

func TestSession_StopClosesOpenPause(t *testing.T) {
	s := Session{StartedAt: at(13, 10, 0)}
	s, _ = s.Pause(at(13, 10, 30))

	s, err := s.Stop(at(13, 11, 0))
	if err != nil {
		t.Fatal(err)
	}
	if s.Running() || s.Paused() {
		t.Errorf("session = %+v, want stopped and not paused", s)
	}
	if got := s.Elapsed(at(13, 23, 0)); got != 30*time.Minute {
		t.Errorf("Elapsed = %v, want 30m regardless of now", got)
	}
}

func TestSession_StateErrors(t *testing.T) {
	running := Session{StartedAt: at(13, 10, 0)}
	paused, _ := running.Pause(at(13, 10, 5))
	stopped, _ := running.Stop(at(13, 10, 10))
	now := at(13, 10, 20)

	if _, err := paused.Pause(now); !errors.Is(err, ErrAlreadyPaused) {
		t.Errorf("pause twice: err = %v", err)
	}
	if _, err := running.Resume(now); !errors.Is(err, ErrNotPaused) {
		t.Errorf("resume running: err = %v", err)
	}
	for name, op := range map[string]func(time.Time) (Session, error){
		"pause": stopped.Pause, "resume": stopped.Resume, "stop": stopped.Stop,
	} {
		if _, err := op(now); !errors.Is(err, ErrNoActiveSession) {
			t.Errorf("%s stopped: err = %v", name, err)
		}
	}
}

func TestWeekStart(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
	}{
		{"monday midnight", at(12, 0, 0)},
		{"wednesday", at(14, 15, 30)},
		{"sunday night", at(18, 23, 59)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WeekStart(tt.now); !got.Equal(at(12, 0, 0)) {
				t.Errorf("WeekStart = %v, want Monday Oct 12 00:00", got)
			}
		})
	}
}

func TestGoalProgress(t *testing.T) {
	now := at(14, 12, 0) // Wednesday
	session := func(subject int64, start time.Time, minutes int) Session {
		end := start.Add(time.Duration(minutes) * time.Minute)
		return Session{SubjectID: subject, StartedAt: start, EndedAt: &end}
	}
	sessions := []Session{
		session(1, at(12, 9, 0), 60),  // math, this week
		session(2, at(13, 9, 0), 30),  // physics, this week
		session(1, at(11, 23, 0), 90), // math, last Sunday: ignored
	}
	active := &Session{SubjectID: 1, StartedAt: at(14, 11, 30)} // 30 min so far

	got := GoalProgress([]Goal{
		{ID: 1, TargetMinutes: 120},                          // overall
		{ID: 2, SubjectID: ptr[int64](1), TargetMinutes: 90}, // math
		{ID: 3, SubjectID: ptr[int64](2), TargetMinutes: 60}, // physics
	}, sessions, active, now)

	want := []struct {
		done      int
		completed bool
	}{{120, true}, {90, true}, {30, false}}
	for i, w := range want {
		if got[i].DoneMinutes != w.done || got[i].Completed != w.completed {
			t.Errorf("goal %d = %d min (completed %v), want %d (%v)", got[i].ID, got[i].DoneMinutes, got[i].Completed, w.done, w.completed)
		}
	}
	if len(sessions) != 3 {
		t.Error("GoalProgress must not modify the sessions slice")
	}
}

func TestHoursToMinutes(t *testing.T) {
	if got := HoursToMinutes(1.5); got != 90 {
		t.Errorf("HoursToMinutes(1.5) = %d", got)
	}
	if got := HoursToMinutes(0.01); got != 1 {
		t.Errorf("HoursToMinutes(0.01) = %d, want rounding to 1", got)
	}
}
