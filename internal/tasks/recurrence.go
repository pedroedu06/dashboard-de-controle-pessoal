package tasks

import (
	"slices"
	"time"
)

// StartOfDay returns midnight of t's day, in t's location.
func StartOfDay(t time.Time) time.Time {
	return CivilDay(t, t.Location())
}

// CivilDay returns midnight in loc of t's calendar date, ignoring t's own
// location. Postgres dates come back as UTC midnight, so this is how they
// become local days.
func CivilDay(t time.Time, loc *time.Location) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

// DayKey identifies a calendar day, independent of location.
func DayKey(day time.Time) string {
	return day.Format(time.DateOnly)
}

// OccursOnDay reports whether a recurring task has an occurrence on day.
func (t Task) OccursOnDay(day time.Time) bool {
	if !t.Repeat.Recurring() {
		return false
	}
	day = StartOfDay(day)
	loc := day.Location()

	if day.Before(StartOfDay(t.StartsAt.In(loc))) {
		return false
	}
	if t.Repeat.Until != nil && day.After(CivilDay(*t.Repeat.Until, loc)) {
		return false
	}
	if t.Repeat.Kind == RepeatWeekly {
		return slices.Contains(t.Repeat.Weekdays, day.Weekday())
	}
	return true
}

// Occurrences lists the days in [from, to] on which the task occurs.
func (t Task) Occurrences(from, to time.Time) []time.Time {
	var days []time.Time
	for day := StartOfDay(from); !day.After(to); day = day.AddDate(0, 0, 1) {
		if t.OccursOnDay(day) {
			days = append(days, day)
		}
	}
	return days
}

// At returns the occurrence of the task on day: same clock time and duration
// as the first occurrence, with OccursOn set.
func (t Task) At(day time.Time) Task {
	day = StartOfDay(day)
	first := t.StartsAt.In(day.Location())
	h, m, s := first.Clock()

	occ := t
	occ.OccursOn = day
	occ.SeriesStartsAt, occ.SeriesEndsAt = t.StartsAt, t.EndsAt
	occ.StartsAt = time.Date(day.Year(), day.Month(), day.Day(), h, m, s, first.Nanosecond(), day.Location())
	occ.EndsAt = occ.StartsAt.Add(t.EndsAt.Sub(t.StartsAt))
	occ.DoneAt = nil
	occ.Items = slices.Clone(t.Items)
	for i := range occ.Items {
		occ.Items[i].Done = false
	}
	return occ
}
