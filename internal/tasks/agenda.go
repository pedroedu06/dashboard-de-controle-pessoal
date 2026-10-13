package tasks

import (
	"cmp"
	"slices"
	"time"
)

type Day struct {
	Date  time.Time // midnight, in the agenda's location
	Tasks []Task
}

type Agenda struct {
	Overdue []Task
	Days    []Day
}

// BuildAgenda groups tasks by day from today onwards, in now's location:
//   - unfinished tasks that already ended go only to Overdue;
//   - a task is listed on every day it occupies; ending exactly at midnight
//     does not occupy the next day;
//   - tasks that ended before today are dropped;
//   - within a day, tasks are ordered by start time.
func BuildAgenda(tasks []Task, now time.Time) Agenda {
	loc := now.Location()
	today := StartOfDay(now)

	sorted := slices.Clone(tasks)
	slices.SortStableFunc(sorted, func(a, b Task) int {
		if c := a.StartsAt.Compare(b.StartsAt); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})

	agenda := Agenda{Overdue: []Task{}, Days: []Day{}}
	dayIndex := map[int64]int{} // keyed by the day's Unix time

	for _, t := range sorted {
		if !t.Done() && t.EndsAt.Before(now) {
			agenda.Overdue = append(agenda.Overdue, t)
			continue
		}
		if !t.EndsAt.After(today) {
			continue
		}

		day := StartOfDay(t.StartsAt.In(loc))
		if day.Before(today) {
			day = today
		}
		for ; day.Before(t.EndsAt); day = day.AddDate(0, 0, 1) {
			i, ok := dayIndex[day.Unix()]
			if !ok {
				i = len(agenda.Days)
				dayIndex[day.Unix()] = i
				agenda.Days = append(agenda.Days, Day{Date: day})
			}
			agenda.Days[i].Tasks = append(agenda.Days[i].Tasks, t)
		}
	}

	slices.SortFunc(agenda.Days, func(a, b Day) int { return a.Date.Compare(b.Date) })
	return agenda
}

// StartOfDay returns midnight of t's day, in t's location.
func StartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
