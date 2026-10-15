package studies

import (
	"math"
	"time"
)

// Grade is how well a review went, mapped to SM-2's 0–5 quality scale.
type Grade string

const (
	GradeForgot Grade = "forgot"
	GradeHard   Grade = "hard"
	GradeOK     Grade = "ok"
	GradeEasy   Grade = "easy"
)

var gradeQuality = map[Grade]int{GradeForgot: 1, GradeHard: 3, GradeOK: 4, GradeEasy: 5}

func (g Grade) Valid() bool {
	_, ok := gradeQuality[g]
	return ok
}

// Quality is the SM-2 quality (0–5) for the grade.
func (g Grade) Quality() int {
	return gradeQuality[g]
}

const (
	InitialEaseFactor = 2.5
	MinEaseFactor     = 1.3
)

// Schedule is a topic's SM-2 state. NextReviewOn is a calendar day; nil means
// the topic is not in the review queue.
type Schedule struct {
	NextReviewOn *time.Time
	IntervalDays int
	EaseFactor   float64
	Repetitions  int
}

// NewSchedule is an unscheduled, fresh SM-2 state.
func NewSchedule() Schedule {
	return Schedule{EaseFactor: InitialEaseFactor}
}

// StartSchedule is the state right after a topic is marked as reviewed: it
// counts as a first good review, so the next one is tomorrow.
func StartSchedule(today time.Time) Schedule {
	return NewSchedule().Review(GradeOK.Quality(), today)
}

// Review applies SM-2 for a review of quality q (0–5) done today.
func (s Schedule) Review(q int, today time.Time) Schedule {
	if q >= 3 {
		switch s.Repetitions {
		case 0:
			s.IntervalDays = 1
		case 1:
			s.IntervalDays = 6
		default:
			s.IntervalDays = int(math.Round(float64(s.IntervalDays) * s.EaseFactor))
		}
		s.Repetitions++
	} else {
		s.Repetitions = 0
		s.IntervalDays = 1
	}

	miss := float64(5 - q)
	s.EaseFactor = max(MinEaseFactor, s.EaseFactor+0.1-miss*(0.08+miss*0.02))

	next := startOfDay(today).AddDate(0, 0, s.IntervalDays)
	s.NextReviewOn = &next
	return s
}

// DaysOverdue is how many days past due the review is on today (0 if due today).
func (s Schedule) DaysOverdue(today time.Time) int {
	if s.NextReviewOn == nil {
		return 0
	}
	due := civilDay(*s.NextReviewOn, today.Location())
	return max(0, daysBetween(due, startOfDay(today)))
}

func startOfDay(t time.Time) time.Time {
	return civilDay(t, t.Location())
}

// civilDay is midnight in loc of t's calendar date. Postgres dates come back
// as UTC midnight, so this is how they become local days.
func civilDay(t time.Time, loc *time.Location) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

// daysBetween counts calendar days from a to b (both midnights).
func daysBetween(a, b time.Time) int {
	return int(math.Round(b.Sub(a).Hours() / 24))
}
