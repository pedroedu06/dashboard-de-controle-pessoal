package main

import (
	"errors"
	"fmt"
	"time"

	"dashboard/internal/tasks"
)

var errDBUnavailable = errors.New("banco não conectado")

// TaskAPI exposes the to-do feature to the frontend. Instants cross the
// boundary as RFC 3339 strings and days as YYYY-MM-DD, since Wails types
// time.Time as `any`.
type TaskAPI struct {
	app *App
}

type ItemDTO struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type RepeatDTO struct {
	Kind     string `json:"kind"`     // none, daily or weekly
	Weekdays []int  `json:"weekdays"` // 0 = Sunday .. 6 = Saturday
	Until    string `json:"until"`    // YYYY-MM-DD or empty
}

type TaskDTO struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Urgency  string `json:"urgency"`
	StartsAt string `json:"startsAt"`
	EndsAt   string `json:"endsAt"`
	Done     bool   `json:"done"`
	// SeriesStartsAt and SeriesEndsAt are the schedule to edit: the first
	// occurrence for recurring tasks, the same as StartsAt/EndsAt otherwise.
	SeriesStartsAt string    `json:"seriesStartsAt"`
	SeriesEndsAt   string    `json:"seriesEndsAt"`
	Repeat         RepeatDTO `json:"repeat"`
	Items          []ItemDTO `json:"items"`
	OccursOn       string    `json:"occursOn"` // YYYY-MM-DD for occurrences of recurring tasks
}

type ItemInputDTO struct {
	ID    int64  `json:"id"` // 0 for new items
	Title string `json:"title"`
}

type TaskInputDTO struct {
	Title    string         `json:"title"`
	Category string         `json:"category"`
	Urgency  string         `json:"urgency"`
	StartsAt string         `json:"startsAt"`
	EndsAt   string         `json:"endsAt"`
	Repeat   RepeatDTO      `json:"repeat"`
	Items    []ItemInputDTO `json:"items"`
}

type FilterDTO struct {
	Category string `json:"category"`
	Urgency  string `json:"urgency"`
}

type DayDTO struct {
	Date  string    `json:"date"` // YYYY-MM-DD, local time
	Tasks []TaskDTO `json:"tasks"`
}

type AgendaDTO struct {
	Overdue []TaskDTO `json:"overdue"`
	Days    []DayDTO  `json:"days"`
}

func (api *TaskAPI) Agenda(filter FilterDTO) (AgendaDTO, error) {
	s, err := api.service()
	if err != nil {
		return AgendaDTO{}, err
	}
	urgency := tasks.Urgency(filter.Urgency)
	if urgency != "" && !urgency.Valid() {
		return AgendaDTO{}, tasks.ErrInvalidUrgency
	}
	agenda, err := s.Agenda(api.app.ctx, time.Now(), tasks.Filter{Category: filter.Category, Urgency: urgency})
	if err != nil {
		return AgendaDTO{}, err
	}

	out := AgendaDTO{Overdue: toDTOs(agenda.Overdue), Days: make([]DayDTO, len(agenda.Days))}
	for i, d := range agenda.Days {
		out.Days[i] = DayDTO{Date: tasks.DayKey(d.Date), Tasks: toDTOs(d.Tasks)}
	}
	return out, nil
}

func (api *TaskAPI) Categories() ([]string, error) {
	s, err := api.service()
	if err != nil {
		return nil, err
	}
	return s.Categories(api.app.ctx)
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

// SetDone completes a task; occursOn is "" for one-off tasks.
func (api *TaskAPI) SetDone(id int64, occursOn string, done bool) error {
	s, err := api.service()
	if err != nil {
		return err
	}
	day, err := parseDay(occursOn, "ocorrência")
	if err != nil {
		return err
	}
	return s.SetDone(api.app.ctx, id, day, done)
}

// SetItemDone checks a list item; occursOn is "" for one-off tasks.
func (api *TaskAPI) SetItemDone(itemID int64, occursOn string, done bool) error {
	s, err := api.service()
	if err != nil {
		return err
	}
	day, err := parseDay(occursOn, "ocorrência")
	if err != nil {
		return err
	}
	return s.SetItemDone(api.app.ctx, itemID, day, done)
}

func (api *TaskAPI) service() (*tasks.Service, error) {
	if api.app.tasks == nil {
		return nil, errDBUnavailable
	}
	return api.app.tasks, nil
}

// parseInput converts the DTO. Instants are moved to local time because the
// calendar rules (same day, until) are local. Empty dates stay zero so
// Normalize reports them.
func parseInput(in TaskInputDTO) (tasks.Input, error) {
	starts, err := parseInstant(in.StartsAt, "início")
	if err != nil {
		return tasks.Input{}, err
	}
	ends, err := parseInstant(in.EndsAt, "fim")
	if err != nil {
		return tasks.Input{}, err
	}
	until, err := parseDay(in.Repeat.Until, "fim da repetição")
	if err != nil {
		return tasks.Input{}, err
	}

	input := tasks.Input{
		Title:    in.Title,
		Category: in.Category,
		Urgency:  tasks.Urgency(in.Urgency),
		StartsAt: starts,
		EndsAt:   ends,
		Repeat:   tasks.Repeat{Kind: tasks.RepeatKind(in.Repeat.Kind), Until: until},
	}
	for _, d := range in.Repeat.Weekdays {
		input.Repeat.Weekdays = append(input.Repeat.Weekdays, time.Weekday(d))
	}
	for _, it := range in.Items {
		input.Items = append(input.Items, tasks.ItemInput{ID: it.ID, Title: it.Title})
	}
	return input, nil
}

func parseInstant(value, field string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("data de %s inválida: %q", field, value)
	}
	return t.In(time.Local), nil
}

func parseDay(value, field string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	d, err := time.ParseInLocation(time.DateOnly, value, time.Local)
	if err != nil {
		return nil, fmt.Errorf("data de %s inválida: %q", field, value)
	}
	return &d, nil
}

func toDTO(t tasks.Task) TaskDTO {
	dto := TaskDTO{
		ID:       t.ID,
		Title:    t.Title,
		Category: t.Category,
		Urgency:  string(t.Urgency),
		StartsAt: t.StartsAt.Local().Format(time.RFC3339),
		EndsAt:   t.EndsAt.Local().Format(time.RFC3339),
		Done:     t.Done(),
		Repeat:   RepeatDTO{Kind: string(t.Repeat.Kind), Weekdays: []int{}},
		Items:    make([]ItemDTO, len(t.Items)),
	}
	if dto.Repeat.Kind == "" {
		dto.Repeat.Kind = string(tasks.RepeatNone)
	}
	for _, d := range t.Repeat.Weekdays {
		dto.Repeat.Weekdays = append(dto.Repeat.Weekdays, int(d))
	}
	if t.Repeat.Until != nil {
		dto.Repeat.Until = tasks.DayKey(*t.Repeat.Until)
	}
	for i, it := range t.Items {
		dto.Items[i] = ItemDTO{ID: it.ID, Title: it.Title, Done: it.Done}
	}
	dto.SeriesStartsAt, dto.SeriesEndsAt = dto.StartsAt, dto.EndsAt
	if !t.OccursOn.IsZero() {
		dto.OccursOn = tasks.DayKey(t.OccursOn)
		dto.SeriesStartsAt = t.SeriesStartsAt.Local().Format(time.RFC3339)
		dto.SeriesEndsAt = t.SeriesEndsAt.Local().Format(time.RFC3339)
	}
	return dto
}

func toDTOs(list []tasks.Task) []TaskDTO {
	out := make([]TaskDTO, len(list))
	for i, t := range list {
		out[i] = toDTO(t)
	}
	return out
}
