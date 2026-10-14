package tasks

import (
	"cmp"
	"slices"
	"strings"
	"time"
)

// RecurringHorizonDays is how many days (today included) recurring tasks are
// shown ahead.
const RecurringHorizonDays = 7

type Day struct {
	Date  time.Time // midnight, in the agenda's location
	Tasks []Task
}

type Agenda struct {
	Overdue []Task
	Days    []Day
}

// Filter narrows the agenda; zero values match everything.
type Filter struct {
	Category string // case-insensitive
	Urgency  Urgency
}

func (f Filter) Matches(t Task) bool {
	if f.Category != "" && !strings.EqualFold(f.Category, t.Category) {
		return false
	}
	return f.Urgency == "" || f.Urgency == t.Urgency
}

// OccurrenceKey identifies a task or item on a given day of a recurring task.
type OccurrenceKey struct {
	ID  int64
	Day string // DayKey
}

// OccurrenceState holds per-day progress of recurring tasks.
type OccurrenceState struct {
	Done       map[OccurrenceKey]bool // by task ID
	ItemChecks map[OccurrenceKey]bool // by item ID
}

// BuildAgenda groups tasks by day from today onwards, in now's location.
//
// One-off tasks:
//   - unfinished tasks that already ended go only to Overdue;
//   - a task is listed on every day it occupies; ending exactly at midnight
//     does not occupy the next day;
//   - tasks that ended before today are dropped.
//
// Recurring tasks are listed on each occurrence from today through
// RecurringHorizonDays, with that day's progress, and are never overdue.
//
// Within a day, tasks are ordered by start time.
func BuildAgenda(tasks []Task, state OccurrenceState, now time.Time, filter Filter) Agenda {
	today := StartOfDay(now)
	lastDay := today.AddDate(0, 0, RecurringHorizonDays-1)

	agenda := Agenda{Overdue: []Task{}, Days: []Day{}}
	dayIndex := map[string]int{}
	addToDay := func(day time.Time, t Task) {
		i, ok := dayIndex[DayKey(day)]
		if !ok {
			i = len(agenda.Days)
			dayIndex[DayKey(day)] = i
			agenda.Days = append(agenda.Days, Day{Date: day})
		}
		agenda.Days[i].Tasks = append(agenda.Days[i].Tasks, t)
	}

	for _, t := range tasks {
		if !filter.Matches(t) {
			continue
		}

		if t.Repeat.Recurring() {
			for _, day := range t.Occurrences(today, lastDay) {
				addToDay(day, withProgress(t.At(day), state))
			}
			continue
		}

		if !t.Done() && t.EndsAt.Before(now) {
			agenda.Overdue = append(agenda.Overdue, t)
			continue
		}
		if !t.EndsAt.After(today) {
			continue
		}
		day := StartOfDay(t.StartsAt.In(now.Location()))
		if day.Before(today) {
			day = today
		}
		for ; day.Before(t.EndsAt); day = day.AddDate(0, 0, 1) {
			addToDay(day, t)
		}
	}

	slices.SortStableFunc(agenda.Overdue, byStart)
	for _, d := range agenda.Days {
		slices.SortStableFunc(d.Tasks, byStart)
	}
	slices.SortFunc(agenda.Days, func(a, b Day) int { return a.Date.Compare(b.Date) })
	return agenda
}

// withProgress applies the stored completion of an occurrence and its items.
func withProgress(occ Task, state OccurrenceState) Task {
	day := DayKey(occ.OccursOn)
	if state.Done[OccurrenceKey{occ.ID, day}] {
		doneAt := occ.OccursOn
		occ.DoneAt = &doneAt
	}
	for i, item := range occ.Items {
		occ.Items[i].Done = state.ItemChecks[OccurrenceKey{item.ID, day}]
	}
	return occ
}

func byStart(a, b Task) int {
	if c := a.StartsAt.Compare(b.StartsAt); c != 0 {
		return c
	}
	return cmp.Compare(a.ID, b.ID)
}
