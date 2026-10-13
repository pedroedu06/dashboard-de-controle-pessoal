package tasks

import "strings"

// Normalize trims the title, defaults the urgency to medium and validates
// the input. Starting in the past is allowed.
func (in Input) Normalize() (Input, error) {
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return Input{}, ErrTitleRequired
	}

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
	return in, nil
}
