package tasks

import (
	"errors"
	"slices"
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

func TestNormalize_TrimsCategoryAndItems(t *testing.T) {
	in := validInput()
	in.Category = "  Faculdade "
	in.Items = []ItemInput{{Title: " arroz "}, {ID: 7, Title: "feijão"}}

	got, err := in.Normalize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Category != "Faculdade" || got.Items[0].Title != "arroz" || got.Items[1].ID != 7 {
		t.Errorf("got = %+v", got)
	}
}

func TestNormalize_DefaultsToNoRepeat(t *testing.T) {
	got, err := validInput().Normalize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Repeat.Kind != RepeatNone || got.Repeat.Recurring() {
		t.Errorf("Repeat = %+v, want none", got.Repeat)
	}
}

func TestNormalize_WeeklySortsAndDedupesWeekdays(t *testing.T) {
	in := validInput()
	in.Repeat = Repeat{Kind: RepeatWeekly, Weekdays: []time.Weekday{time.Friday, time.Monday, time.Friday}}

	got, err := in.Normalize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []time.Weekday{time.Monday, time.Friday}; !slices.Equal(got.Repeat.Weekdays, want) {
		t.Errorf("Weekdays = %v, want %v", got.Repeat.Weekdays, want)
	}
}

func TestNormalize_DailyDropsWeekdays(t *testing.T) {
	in := validInput()
	in.Repeat = Repeat{Kind: RepeatDaily, Weekdays: []time.Weekday{time.Monday}}

	got, err := in.Normalize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Repeat.Weekdays) != 0 {
		t.Errorf("Weekdays = %v, want none", got.Repeat.Weekdays)
	}
}

func TestNormalize_RepeatErrors(t *testing.T) {
	dayBefore := validInput().StartsAt.AddDate(0, 0, -1)
	tests := []struct {
		name   string
		modify func(*Input)
		want   error
	}{
		{"unknown kind", func(in *Input) { in.Repeat.Kind = "monthly" }, ErrInvalidRepeat},
		{"weekly without days", func(in *Input) { in.Repeat = Repeat{Kind: RepeatWeekly} }, ErrWeekdaysRequired},
		{"invalid weekday", func(in *Input) { in.Repeat = Repeat{Kind: RepeatWeekly, Weekdays: []time.Weekday{9}} }, ErrInvalidWeekday},
		{"spans midnight", func(in *Input) {
			in.Repeat = Repeat{Kind: RepeatDaily}
			in.EndsAt = in.StartsAt.Add(12 * time.Hour)
		}, ErrRecurringSpansDays},
		{"until before start", func(in *Input) { in.Repeat = Repeat{Kind: RepeatDaily, Until: &dayBefore} }, ErrUntilBeforeStart},
		{"blank item", func(in *Input) { in.Items = []ItemInput{{Title: "  "}} }, ErrItemTitleRequired},
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

func TestNormalize_UntilOnStartDayIsAllowed(t *testing.T) {
	in := validInput()
	until := in.StartsAt
	in.Repeat = Repeat{Kind: RepeatDaily, Until: &until}

	got, err := in.Normalize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Repeat.Until.Equal(StartOfDay(in.StartsAt)) {
		t.Errorf("Until = %v, want the start day at midnight", got.Repeat.Until)
	}
}
