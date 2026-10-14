package tasks

import (
	"slices"
	"testing"
	"time"
)

func recurring(start, end time.Time, repeat Repeat) Task {
	t := task(1, start, end)
	t.Repeat = repeat
	return t
}

func days(list []time.Time) []int {
	out := []int{}
	for _, d := range list {
		out = append(out, d.Day())
	}
	return out
}

// Oct 13 2026 is a Tuesday.
func TestOccurrences_Daily(t *testing.T) {
	task := recurring(at(14, 7, 0), at(14, 8, 0), Repeat{Kind: RepeatDaily})

	got := days(task.Occurrences(at(13, 0, 0), at(17, 0, 0)))
	if !slices.Equal(got, []int{14, 15, 16, 17}) {
		t.Errorf("days = %v, want from the start day on", got)
	}
}

func TestOccurrences_WeeklyPicksWeekdaysAndSkipsStartDay(t *testing.T) {
	// Starts on Tuesday, but repeats on Monday, Wednesday and Friday.
	task := recurring(at(13, 7, 0), at(13, 8, 0), Repeat{
		Kind:     RepeatWeekly,
		Weekdays: []time.Weekday{time.Monday, time.Wednesday, time.Friday},
	})

	got := days(task.Occurrences(at(13, 0, 0), at(20, 0, 0)))
	if !slices.Equal(got, []int{14, 16, 19}) {
		t.Errorf("days = %v, want [14 16 19]", got)
	}
}

func TestOccurrences_RespectsUntil(t *testing.T) {
	until := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC) // as read from Postgres
	task := recurring(at(13, 7, 0), at(13, 8, 0), Repeat{Kind: RepeatDaily, Until: &until})

	got := days(task.Occurrences(at(13, 0, 0), at(20, 0, 0)))
	if !slices.Equal(got, []int{13, 14, 15}) {
		t.Errorf("days = %v, want through the 15th", got)
	}
}

func TestOccursOnDay_OneOffNeverOccurs(t *testing.T) {
	if task(1, at(13, 7, 0), at(13, 8, 0)).OccursOnDay(at(13, 0, 0)) {
		t.Error("one-off tasks have no occurrences")
	}
}

func TestAt_KeepsClockAndDurationAndResetsProgress(t *testing.T) {
	base := recurring(at(13, 7, 30), at(13, 9, 0), Repeat{Kind: RepeatDaily})
	base.Items = []Item{{ID: 1, Title: "x", Done: true}}
	doneAt := at(13, 9, 0)
	base.DoneAt = &doneAt

	occ := base.At(at(16, 0, 0))

	if !occ.StartsAt.Equal(at(16, 7, 30)) || !occ.EndsAt.Equal(at(16, 9, 0)) {
		t.Errorf("occurrence = %v – %v", occ.StartsAt, occ.EndsAt)
	}
	if !occ.OccursOn.Equal(at(16, 0, 0)) || occ.Done() || occ.Items[0].Done {
		t.Errorf("occurrence progress not reset: %+v", occ)
	}
	if !base.Items[0].Done {
		t.Error("At must not modify the original task's items")
	}
}
