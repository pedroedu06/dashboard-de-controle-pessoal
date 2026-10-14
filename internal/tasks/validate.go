package tasks

import (
	"slices"
	"strings"
	"time"
)

// Normalize trims text fields, applies defaults and validates the input.
// Starting in the past is allowed. Calendar rules (same day, until) use
// StartsAt's location.
func (in Input) Normalize() (Input, error) {
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return Input{}, ErrTitleRequired
	}
	in.Category = strings.TrimSpace(in.Category)

	if in.Urgency == "" {
		in.Urgency = UrgencyMedium
	}
	if !in.Urgency.Valid() {
		return Input{}, ErrInvalidUrgency
	}

	if in.StartsAt.IsZero() || in.EndsAt.IsZero() {
		return Input{}, ErrMissingSchedule
	}
	if !in.EndsAt.After(in.StartsAt) {
		return Input{}, ErrEndBeforeStart
	}

	repeat, err := normalizeRepeat(in.Repeat, in.StartsAt, in.EndsAt)
	if err != nil {
		return Input{}, err
	}
	in.Repeat = repeat

	items := make([]ItemInput, len(in.Items))
	for i, it := range in.Items {
		it.Title = strings.TrimSpace(it.Title)
		if it.Title == "" {
			return Input{}, ErrItemTitleRequired
		}
		items[i] = it
	}
	in.Items = items

	return in, nil
}

func normalizeRepeat(r Repeat, starts, ends time.Time) (Repeat, error) {
	switch r.Kind {
	case "", RepeatNone:
		return Repeat{Kind: RepeatNone}, nil
	case RepeatDaily:
		r.Weekdays = nil
	case RepeatWeekly:
		days := slices.Clone(r.Weekdays)
		for _, d := range days {
			if d < time.Sunday || d > time.Saturday {
				return Repeat{}, ErrInvalidWeekday
			}
		}
		slices.Sort(days)
		r.Weekdays = slices.Compact(days)
		if len(r.Weekdays) == 0 {
			return Repeat{}, ErrWeekdaysRequired
		}
	default:
		return Repeat{}, ErrInvalidRepeat
	}

	loc := starts.Location()
	firstDay := StartOfDay(starts)
	if !StartOfDay(ends.In(loc)).Equal(firstDay) {
		return Repeat{}, ErrRecurringSpansDays
	}
	if r.Until != nil {
		until := CivilDay(*r.Until, loc)
		if until.Before(firstDay) {
			return Repeat{}, ErrUntilBeforeStart
		}
		r.Until = &until
	}
	return r, nil
}
