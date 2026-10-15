package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"dashboard/internal/studies"
	"dashboard/internal/testutil"
)

func TestStudyAPI_WithoutDatabase(t *testing.T) {
	api := &StudyAPI{app: &App{ctx: context.Background()}}

	if _, err := api.Overview(); !errors.Is(err, errDBUnavailable) {
		t.Errorf("Overview: err = %v, want errDBUnavailable", err)
	}
	if err := api.StartSession(1, 0); !errors.Is(err, errDBUnavailable) {
		t.Errorf("StartSession: err = %v, want errDBUnavailable", err)
	}
}

func TestOptionalID(t *testing.T) {
	if optionalID(0) != nil {
		t.Error("0 should mean no id")
	}
	if id := optionalID(7); id == nil || *id != 7 {
		t.Errorf("optionalID(7) = %v", id)
	}
}

func TestStudyAPI_Flow(t *testing.T) {
	app := &App{ctx: context.Background(), studies: studies.NewService(testutil.Tx(t))}
	api := &StudyAPI{app: app}

	sub, err := api.CreateSubject("Cálculo")
	if err != nil {
		t.Fatal(err)
	}
	if err := api.CreateTopic(sub.ID, "Limites"); err != nil {
		t.Fatal(err)
	}
	if err := api.CreateGoal(0, 1.5); err != nil {
		t.Fatal(err)
	}
	if err := api.StartSession(sub.ID, 0); err != nil {
		t.Fatal(err)
	}

	ov, err := api.Overview()
	if err != nil {
		t.Fatal(err)
	}
	if ov.Active == nil || ov.Active.SubjectName != "Cálculo" || ov.Active.TopicID != 0 || ov.Active.Paused {
		t.Errorf("active = %+v", ov.Active)
	}
	if len(ov.Goals) != 1 || ov.Goals[0].SubjectID != 0 || ov.Goals[0].TargetMinutes != 90 {
		t.Errorf("goals = %+v, want one overall goal of 90 min", ov.Goals)
	}
	if len(ov.Subjects) != 1 || len(ov.Subjects[0].Topics) != 1 || ov.Subjects[0].Topics[0].Status != "not_seen" {
		t.Errorf("subjects = %+v", ov.Subjects)
	}

	result, err := api.StopSession()
	if err != nil || result.Saved {
		t.Errorf("StopSession = %+v, %v, want discarded (under a minute)", result, err)
	}
}

func TestParseQuestionLog(t *testing.T) {
	log, err := parseQuestionLog(QuestionLogInputDTO{SubjectID: 1, TopicID: 0, DoneOn: "2026-10-10", Total: 10, Correct: 7, Source: "FGV"})
	if err != nil {
		t.Fatal(err)
	}
	if log.TopicID != nil || log.DoneOn.Format("2006-01-02") != "2026-10-10" || log.Total != 10 || log.Source != "FGV" {
		t.Errorf("log = %+v", log)
	}
	if log, _ := parseQuestionLog(QuestionLogInputDTO{SubjectID: 1, Total: 1}); !log.DoneOn.IsZero() {
		t.Errorf("empty date should stay zero (today is applied by the service), got %v", log.DoneOn)
	}
	if _, err := parseQuestionLog(QuestionLogInputDTO{DoneOn: "10/10/2026"}); err == nil {
		t.Error("expected an error for a non YYYY-MM-DD day")
	}
}

func TestDayString(t *testing.T) {
	if dayString(nil) != "" {
		t.Error("nil day should be empty")
	}
	day := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	if dayString(&day) != "2026-10-16" {
		t.Errorf("dayString = %q", dayString(&day))
	}
}

func TestStudyAPI_ReviewAndQuestions(t *testing.T) {
	app := &App{ctx: context.Background(), studies: studies.NewService(testutil.Tx(t))}
	api := &StudyAPI{app: app}

	sub, err := api.CreateSubject("Português")
	if err != nil {
		t.Fatal(err)
	}
	if err := api.CreateTopic(sub.ID, "Crase"); err != nil {
		t.Fatal(err)
	}
	ov, _ := api.Overview()
	topicID := ov.Subjects[0].Topics[0].ID

	if err := api.SetTopicStatus(topicID, "reviewed"); err != nil {
		t.Fatal(err)
	}
	ov, _ = api.Overview()
	if ov.Subjects[0].Topics[0].NextReviewOn == "" {
		t.Error("reviewed topic should have a next review day")
	}
	if due, _ := api.DueReviews(); len(due) != 0 {
		t.Errorf("due = %+v, want none until tomorrow", due)
	}
	if _, err := api.ReviewTopic(topicID, "great"); !errors.Is(err, studies.ErrInvalidGrade) {
		t.Errorf("invalid grade: err = %v", err)
	}

	if err := api.LogQuestions(QuestionLogInputDTO{SubjectID: sub.ID, TopicID: topicID, Total: 12, Correct: 3}); err != nil {
		t.Fatal(err)
	}
	report, err := api.QuestionReport(30)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Weaknesses) != 1 || report.Weaknesses[0].Rate != 0.25 || len(report.Logs) != 1 {
		t.Errorf("report = %+v", report)
	}
}
