package tasks

import (
	"slices"
	"testing"
	"time"
)

var brt = time.FixedZone("BRT", -3*60*60)

// at builds a time on October 2026 in BRT.
func at(day, hour, minute int) time.Time {
	return time.Date(2026, 10, day, hour, minute, 0, 0, brt)
}

func task(id int64, start, end time.Time) Task {
	return Task{ID: id, Title: "t", Urgency: UrgencyMedium, StartsAt: start, EndsAt: end}
}

func done(t Task) Task {
	doneAt := t.EndsAt
	t.DoneAt = &doneAt
	return t
}

func ids(tasks []Task) []int64 {
	out := []int64{}
	for _, t := range tasks {
		out = append(out, t.ID)
	}
	return out
}

func dayIDs(a Agenda) map[int][]int64 {
	out := map[int][]int64{}
	for _, d := range a.Days {
		out[d.Date.Day()] = ids(d.Tasks)
	}
	return out
}

// now is Oct 13 2026, 12:00 BRT.
var now = at(13, 12, 0)

func TestBuildAgenda_SameDayOrderedByStart(t *testing.T) {
	a := BuildAgenda([]Task{
		task(1, at(13, 18, 0), at(13, 19, 0)),
		task(2, at(13, 14, 0), at(13, 15, 0)),
		task(3, at(13, 14, 0), at(13, 16, 0)), // overlaps task 2
	}, now)

	if got := dayIDs(a)[13]; !slices.Equal(got, []int64{2, 3, 1}) {
		t.Errorf("day 13 = %v, want [2 3 1]", got)
	}
}

func TestBuildAgenda_MultiDayTaskAppearsOnEveryDay(t *testing.T) {
	a := BuildAgenda([]Task{task(1, at(14, 22, 0), at(16, 10, 0))}, now)

	days := dayIDs(a)
	for _, d := range []int{14, 15, 16} {
		if !slices.Equal(days[d], []int64{1}) {
			t.Errorf("day %d = %v, want [1]", d, days[d])
		}
	}
	if len(a.Days) != 3 {
		t.Errorf("got %d days, want 3", len(a.Days))
	}
}

func TestBuildAgenda_EndingAtMidnightDoesNotOccupyNextDay(t *testing.T) {
	a := BuildAgenda([]Task{task(1, at(14, 20, 0), at(15, 0, 0))}, now)

	if len(a.Days) != 1 || a.Days[0].Date.Day() != 14 {
		t.Errorf("days = %v, want only day 14", dayIDs(a))
	}
}

func TestBuildAgenda_OngoingTaskStartedInThePastShowsFromToday(t *testing.T) {
	a := BuildAgenda([]Task{task(1, at(10, 9, 0), at(14, 9, 0))}, now)

	days := dayIDs(a)
	if len(a.Days) != 2 || days[13] == nil || days[14] == nil {
		t.Errorf("days = %v, want 13 and 14", days)
	}
}

func TestBuildAgenda_OverdueOnlyInOverdueGroup(t *testing.T) {
	a := BuildAgenda([]Task{
		task(1, at(13, 8, 0), at(13, 9, 0)),   // earlier today, not done
		task(2, at(10, 8, 0), at(10, 9, 0)),   // days ago, not done
		task(3, at(13, 13, 0), at(13, 14, 0)), // later today
	}, now)

	if got := ids(a.Overdue); !slices.Equal(got, []int64{2, 1}) {
		t.Errorf("overdue = %v, want [2 1]", got)
	}
	if got := dayIDs(a)[13]; !slices.Equal(got, []int64{3}) {
		t.Errorf("day 13 = %v, want [3]", got)
	}
}

func TestBuildAgenda_DoneTasks(t *testing.T) {
	a := BuildAgenda([]Task{
		done(task(1, at(13, 8, 0), at(13, 9, 0))), // done earlier today: stays on today
		done(task(2, at(10, 8, 0), at(10, 9, 0))), // done days ago: dropped
	}, now)

	if len(a.Overdue) != 0 {
		t.Errorf("overdue = %v, want none", ids(a.Overdue))
	}
	if got := dayIDs(a); len(got) != 1 || !slices.Equal(got[13], []int64{1}) {
		t.Errorf("days = %v, want only day 13 with [1]", got)
	}
}

func TestBuildAgenda_EmptyDaysAreOmittedAndSorted(t *testing.T) {
	a := BuildAgenda([]Task{
		task(1, at(20, 9, 0), at(20, 10, 0)),
		task(2, at(15, 9, 0), at(15, 10, 0)),
	}, now)

	var got []int
	for _, d := range a.Days {
		got = append(got, d.Date.Day())
	}
	if !slices.Equal(got, []int{15, 20}) {
		t.Errorf("days = %v, want [15 20]", got)
	}
}

func TestBuildAgenda_UsesNowLocation(t *testing.T) {
	// 01:00 UTC on the 14th is still the 13th in BRT.
	a := BuildAgenda([]Task{task(1, time.Date(2026, 10, 14, 1, 0, 0, 0, time.UTC), time.Date(2026, 10, 14, 2, 0, 0, 0, time.UTC))}, now)

	if len(a.Days) != 1 || a.Days[0].Date.Day() != 13 || a.Days[0].Date.Location() != brt {
		t.Errorf("days = %v, want day 13 in BRT", a.Days)
	}
}

func TestBuildAgenda_Empty(t *testing.T) {
	a := BuildAgenda(nil, now)

	if a.Overdue == nil || a.Days == nil || len(a.Overdue)+len(a.Days) != 0 {
		t.Errorf("agenda = %+v, want empty non-nil slices", a)
	}
}
