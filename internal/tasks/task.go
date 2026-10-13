// Package tasks implements the to-do feature: validation, the agenda rules
// and persistence.
package tasks

import (
	"errors"
	"time"
)

type Urgency string

const (
	UrgencyLow    Urgency = "low"
	UrgencyMedium Urgency = "medium"
	UrgencyHigh   Urgency = "high"
)

func (u Urgency) Valid() bool {
	switch u {
	case UrgencyLow, UrgencyMedium, UrgencyHigh:
		return true
	}
	return false
}

// User-facing errors, shown as-is by the frontend.
var (
	ErrTitleRequired   = errors.New("informe o nome da tarefa")
	ErrEndBeforeStart  = errors.New("o fim da tarefa deve ser depois do início")
	ErrInvalidUrgency  = errors.New("urgência inválida: use baixa, média ou alta")
	ErrNotFound        = errors.New("tarefa não encontrada")
	ErrMissingSchedule = errors.New("informe o início e o fim da tarefa")
)

type Task struct {
	ID       int64
	Title    string
	Urgency  Urgency
	StartsAt time.Time
	EndsAt   time.Time
	DoneAt   *time.Time
}

func (t Task) Done() bool {
	return t.DoneAt != nil
}

// Input holds the user-editable fields of a task.
type Input struct {
	Title    string
	Urgency  Urgency
	StartsAt time.Time
	EndsAt   time.Time
}
