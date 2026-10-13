package main

import (
	"errors"
	"fmt"
	"time"

	"dashboard/internal/tasks"
)

var errDBUnavailable = errors.New("banco não conectado")

// TaskAPI exposes the to-do feature to the frontend. Dates cross the
// boundary as RFC 3339 strings, since Wails types time.Time as `any`.
type TaskAPI struct {
	app *App
}

type TaskDTO struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Urgency  string `json:"urgency"`
	StartsAt string `json:"startsAt"`
	EndsAt   string `json:"endsAt"`
	Done     bool   `json:"done"`
}

type TaskInputDTO struct {
	Title    string `json:"title"`
	Urgency  string `json:"urgency"`
	StartsAt string `json:"startsAt"`
	EndsAt   string `json:"endsAt"`
}

type DayDTO struct {
	Date  string    `json:"date"` // YYYY-MM-DD, local time
	Tasks []TaskDTO `json:"tasks"`
}

type AgendaDTO struct {
	Overdue []TaskDTO `json:"overdue"`
	Days    []DayDTO  `json:"days"`
}

func (api *TaskAPI) Agenda() (AgendaDTO, error) {
	s, err := api.service()
	if err != nil {
		return AgendaDTO{}, err
	}
	agenda, err := s.Agenda(api.app.ctx, time.Now())
	if err != nil {
		return AgendaDTO{}, err
	}

	out := AgendaDTO{Overdue: toDTOs(agenda.Overdue), Days: make([]DayDTO, len(agenda.Days))}
	for i, d := range agenda.Days {
		out.Days[i] = DayDTO{Date: d.Date.Format(time.DateOnly), Tasks: toDTOs(d.Tasks)}
	}
	return out, nil
}

func (api *TaskAPI) Create(in TaskInputDTO) (TaskDTO, error) {
	s, err := api.service()
	if err != nil {
		return TaskDTO{}, err
	}
	input, err := parseInput(in)
	if err != nil {
		return TaskDTO{}, err
	}
	t, err := s.Create(api.app.ctx, input)
	return toDTO(t), err
}

func (api *TaskAPI) Update(id int64, in TaskInputDTO) (TaskDTO, error) {
	s, err := api.service()
	if err != nil {
		return TaskDTO{}, err
	}
	input, err := parseInput(in)
	if err != nil {
		return TaskDTO{}, err
	}
	t, err := s.Update(api.app.ctx, id, input)
	return toDTO(t), err
}

func (api *TaskAPI) Delete(id int64) error {
	s, err := api.service()
	if err != nil {
		return err
	}
	return s.Delete(api.app.ctx, id)
}

func (api *TaskAPI) SetDone(id int64, done bool) (TaskDTO, error) {
	s, err := api.service()
	if err != nil {
		return TaskDTO{}, err
	}
	t, err := s.SetDone(api.app.ctx, id, done)
	return toDTO(t), err
}

func (api *TaskAPI) service() (*tasks.Service, error) {
	if api.app.tasks == nil {
		return nil, errDBUnavailable
	}
	return api.app.tasks, nil
}

// parseInput converts the DTO; empty dates stay zero so Normalize reports them.
func parseInput(in TaskInputDTO) (tasks.Input, error) {
	starts, err := parseTime(in.StartsAt, "início")
	if err != nil {
		return tasks.Input{}, err
	}
	ends, err := parseTime(in.EndsAt, "fim")
	if err != nil {
		return tasks.Input{}, err
	}
	return tasks.Input{Title: in.Title, Urgency: tasks.Urgency(in.Urgency), StartsAt: starts, EndsAt: ends}, nil
}

func parseTime(value, field string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("data de %s inválida: %q", field, value)
	}
	return t, nil
}

func toDTO(t tasks.Task) TaskDTO {
	return TaskDTO{
		ID:       t.ID,
		Title:    t.Title,
		Urgency:  string(t.Urgency),
		StartsAt: t.StartsAt.Local().Format(time.RFC3339),
		EndsAt:   t.EndsAt.Local().Format(time.RFC3339),
		Done:     t.Done(),
	}
}

func toDTOs(list []tasks.Task) []TaskDTO {
	out := make([]TaskDTO, len(list))
	for i, t := range list {
		out[i] = toDTO(t)
	}
	return out
}
