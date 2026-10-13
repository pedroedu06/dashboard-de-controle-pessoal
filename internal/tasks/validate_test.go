package tasks

import (
	"errors"
	"testing"
	"time"
)

func validInput() Input {
	start := time.Date(2026, 10, 13, 14, 0, 0, 0, time.UTC)
	return Input{Title: "Estudar Go", Urgency: UrgencyHigh, StartsAt: start, EndsAt: start.Add(time.Hour)}
}

func TestNormalize_Valid(t *testing.T) {
	in := validInput()
	in.Title = "  Estudar Go  "

	got, err := in.Normalize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "Estudar Go" {
		t.Errorf("Title = %q, want trimmed", got.Title)
	}
}

func TestNormalize_DefaultsUrgencyToMedium(t *testing.T) {
	in := validInput()
	in.Urgency = ""

	got, err := in.Normalize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Urgency != UrgencyMedium {
		t.Errorf("Urgency = %q, want medium", got.Urgency)
	}
}

func TestNormalize_AllowsStartInThePast(t *testing.T) {
	in := validInput()
	in.StartsAt = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	if _, err := in.Normalize(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNormalize_Errors(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*Input)
		want   error
	}{
		{"empty title", func(in *Input) { in.Title = "" }, ErrTitleRequired},
		{"blank title", func(in *Input) { in.Title = "   " }, ErrTitleRequired},
		{"invalid urgency", func(in *Input) { in.Urgency = "urgent" }, ErrInvalidUrgency},
		{"missing start", func(in *Input) { in.StartsAt = time.Time{} }, ErrMissingSchedule},
		{"missing end", func(in *Input) { in.EndsAt = time.Time{} }, ErrMissingSchedule},
		{"end equals start", func(in *Input) { in.EndsAt = in.StartsAt }, ErrEndBeforeStart},
		{"end before start", func(in *Input) { in.EndsAt = in.StartsAt.Add(-time.Minute) }, ErrEndBeforeStart},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			tt.modify(&in)

			if _, err := in.Normalize(); !errors.Is(err, tt.want) {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}
}
