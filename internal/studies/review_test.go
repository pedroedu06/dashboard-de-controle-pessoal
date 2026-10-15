package studies

import (
	"testing"
	"time"
)

func reviewDays(t *testing.T, sch Schedule, today time.Time) int {
	t.Helper()
	if sch.NextReviewOn == nil {
		t.Fatal("NextReviewOn is nil")
	}
	return daysBetween(startOfDay(today), *sch.NextReviewOn)
}

func TestStartSchedule_NextReviewTomorrow(t *testing.T) {
	today := at(13, 15, 30)
	sch := StartSchedule(today)

	if sch.Repetitions != 1 || sch.IntervalDays != 1 || sch.EaseFactor != InitialEaseFactor {
		t.Errorf("schedule = %+v, want 1 repetition, 1 day, EF 2.5", sch)
	}
	if !sch.NextReviewOn.Equal(at(14, 0, 0)) {
		t.Errorf("NextReviewOn = %v, want tomorrow at midnight", sch.NextReviewOn)
	}
}

func TestReview_ClassicIntervals(t *testing.T) {
	// Marked as reviewed, then three OK reviews: 1, 6, 15, 38 days.
	today := at(1, 9, 0)
	sch := StartSchedule(today)
	want := []int{6, 15, 38}
	for i, days := range want {
		today = *sch.NextReviewOn
		sch = sch.Review(GradeOK.Quality(), today)
		if got := reviewDays(t, sch, today); got != days || sch.Repetitions != i+2 {
			t.Errorf("review %d: interval = %d days (reps %d), want %d", i+1, got, sch.Repetitions, days)
		}
	}
	if sch.EaseFactor != InitialEaseFactor {
		t.Errorf("EF = %v, OK reviews keep it at 2.5", sch.EaseFactor)
	}
}

func TestReview_GradesAdjustEaseFactor(t *testing.T) {
	today := at(13, 9, 0)
	base := StartSchedule(today)

	easy := base.Review(GradeEasy.Quality(), today)
	hard := base.Review(GradeHard.Quality(), today)
	if easy.EaseFactor <= InitialEaseFactor || hard.EaseFactor >= InitialEaseFactor {
		t.Errorf("EF easy = %v, hard = %v; want above and below 2.5", easy.EaseFactor, hard.EaseFactor)
	}
	if easy.IntervalDays != 6 || hard.IntervalDays != 6 {
		t.Errorf("second review is always 6 days when passed: easy %d, hard %d", easy.IntervalDays, hard.IntervalDays)
	}
}

func TestReview_ForgotResetsCycle(t *testing.T) {
	today := at(13, 9, 0)
	sch := Schedule{IntervalDays: 38, Repetitions: 4, EaseFactor: 2.5}

	sch = sch.Review(GradeForgot.Quality(), today)

	if sch.Repetitions != 0 || reviewDays(t, sch, today) != 1 {
		t.Errorf("schedule = %+v, want reset to tomorrow", sch)
	}
	if sch.EaseFactor >= 2.5 {
		t.Errorf("EF = %v, forgetting should lower it", sch.EaseFactor)
	}
}

func TestReview_EaseFactorFloor(t *testing.T) {
	sch := StartSchedule(at(1, 0, 0))
	for range 20 {
		sch = sch.Review(GradeForgot.Quality(), at(1, 0, 0))
	}
	if sch.EaseFactor != MinEaseFactor {
		t.Errorf("EF = %v, want the floor %v", sch.EaseFactor, MinEaseFactor)
	}
}

func TestGrade(t *testing.T) {
	for g, q := range map[Grade]int{GradeForgot: 1, GradeHard: 3, GradeOK: 4, GradeEasy: 5} {
		if !g.Valid() || g.Quality() != q {
			t.Errorf("%s: valid = %v, quality = %d, want %d", g, g.Valid(), g.Quality(), q)
		}
	}
	if Grade("great").Valid() {
		t.Error("unknown grades must be invalid")
	}
}

func TestSchedule_DaysOverdue(t *testing.T) {
	due := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC) // as read from Postgres
	sch := Schedule{NextReviewOn: &due}

	if got := sch.DaysOverdue(at(13, 22, 0)); got != 3 {
		t.Errorf("DaysOverdue = %d, want 3", got)
	}
	if got := sch.DaysOverdue(at(9, 8, 0)); got != 0 {
		t.Errorf("not due yet: DaysOverdue = %d, want 0", got)
	}
}
