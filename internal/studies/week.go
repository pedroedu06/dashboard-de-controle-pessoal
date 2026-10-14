package studies

import "time"

// WeekStart returns Monday 00:00 of now's week, in now's location.
func WeekStart(now time.Time) time.Time {
	y, m, d := now.Date()
	daysSinceMonday := (int(now.Weekday()) + 6) % 7
	return time.Date(y, m, d-daysSinceMonday, 0, 0, 0, 0, now.Location())
}

type GoalStatus struct {
	Goal
	DoneMinutes int
	Completed   bool
}

// GoalProgress sums this week's study time per goal. A session counts in the
// week it started; the running session (if any) counts up to now. Overall
// goals count every subject, archived ones included.
func GoalProgress(goals []Goal, sessions []Session, active *Session, now time.Time) []GoalStatus {
	weekStart := WeekStart(now)
	weekEnd := weekStart.AddDate(0, 0, 7)

	all := sessions
	if active != nil {
		all = append(all[:len(all):len(all)], *active)
	}

	out := make([]GoalStatus, len(goals))
	for i, g := range goals {
		var total time.Duration
		for _, s := range all {
			if s.StartedAt.Before(weekStart) || !s.StartedAt.Before(weekEnd) {
				continue
			}
			if g.SubjectID != nil && *g.SubjectID != s.SubjectID {
				continue
			}
			total += s.Elapsed(now)
		}
		done := int(total / time.Minute)
		out[i] = GoalStatus{Goal: g, DoneMinutes: done, Completed: done >= g.TargetMinutes}
	}
	return out
}
