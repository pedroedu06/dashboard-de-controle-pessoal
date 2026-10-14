package studies

import "time"

// MinSessionDuration is the shortest session kept when the timer stops.
const MinSessionDuration = time.Minute

func (s Session) Running() bool { return s.EndedAt == nil }
func (s Session) Paused() bool  { return s.PausedAt != nil }

// Elapsed is the studied time: from start to end (or now while running),
// minus every pause, including one still open.
func (s Session) Elapsed(now time.Time) time.Duration {
	end := now
	if s.EndedAt != nil {
		end = *s.EndedAt
	}
	d := end.Sub(s.StartedAt) - time.Duration(s.PausedSeconds)*time.Second
	if s.PausedAt != nil {
		d -= end.Sub(*s.PausedAt)
	}
	return max(d, 0)
}

func (s Session) Pause(now time.Time) (Session, error) {
	if !s.Running() {
		return s, ErrNoActiveSession
	}
	if s.Paused() {
		return s, ErrAlreadyPaused
	}
	s.PausedAt = &now
	return s, nil
}

func (s Session) Resume(now time.Time) (Session, error) {
	if !s.Running() {
		return s, ErrNoActiveSession
	}
	if !s.Paused() {
		return s, ErrNotPaused
	}
	s.PausedSeconds += seconds(now.Sub(*s.PausedAt))
	s.PausedAt = nil
	return s, nil
}

// Stop ends the session, closing an open pause first.
func (s Session) Stop(now time.Time) (Session, error) {
	if !s.Running() {
		return s, ErrNoActiveSession
	}
	if s.Paused() {
		s, _ = s.Resume(now)
	}
	s.EndedAt = &now
	return s, nil
}

func seconds(d time.Duration) int {
	return int(max(d, 0).Round(time.Second) / time.Second)
}
