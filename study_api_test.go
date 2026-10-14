package main

import (
	"context"
	"errors"
	"testing"

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
