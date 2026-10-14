// Package tasks implements the to-do feature: validation, recurrence, the
// agenda rules and persistence.
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

type RepeatKind string

const (
	RepeatNone   RepeatKind = "none"
	RepeatDaily  RepeatKind = "daily"
	RepeatWeekly RepeatKind = "weekly"
)

// Repeat describes how a task recurs. Until is a calendar day (inclusive).
type Repeat struct {
	Kind     RepeatKind
	Weekdays []time.Weekday // weekly only, sorted
	Until    *time.Time
}

func (r Repeat) Recurring() bool {
	return r.Kind == RepeatDaily || r.Kind == RepeatWeekly
}

// User-facing errors, shown as-is by the frontend.
var (
	ErrTitleRequired        = errors.New("informe o nome da tarefa")
	ErrEndBeforeStart       = errors.New("o fim da tarefa deve ser depois do início")
	ErrInvalidUrgency       = errors.New("urgência inválida: use baixa, média ou alta")
	ErrNotFound             = errors.New("tarefa não encontrada")
	ErrMissingSchedule      = errors.New("informe o início e o fim da tarefa")
	ErrInvalidRepeat        = errors.New("repetição inválida: use diária ou semanal")
	ErrWeekdaysRequired     = errors.New("escolha pelo menos um dia da semana")
	ErrInvalidWeekday       = errors.New("dia da semana inválido")
	ErrRecurringSpansDays   = errors.New("uma tarefa recorrente deve começar e terminar no mesmo dia")
	ErrUntilBeforeStart     = errors.New("a data final da repetição deve ser depois do início")
	ErrItemTitleRequired    = errors.New("todo item da lista precisa de um nome")
	ErrItemNotFound         = errors.New("item da lista não encontrado")
	ErrDoneByItems          = errors.New("esta tarefa é concluída marcando os itens da lista")
	ErrInvalidOccurrence    = errors.New("esta data não é uma ocorrência da tarefa")
	ErrOccurrenceRequired   = errors.New("informe o dia da ocorrência da tarefa recorrente")
	ErrUnexpectedOccurrence = errors.New("só tarefas recorrentes têm ocorrências")
)

type Item struct {
	ID    int64
	Title string
	Done  bool
}

type Task struct {
	ID       int64
	Title    string
	Category string // empty when uncategorized
	Urgency  Urgency
	StartsAt time.Time
	EndsAt   time.Time
	DoneAt   *time.Time
	Repeat   Repeat
	Items    []Item

	// OccursOn is the day of this occurrence of a recurring task; zero otherwise.
	OccursOn time.Time
	// SeriesStartsAt and SeriesEndsAt keep the first occurrence's schedule on
	// occurrences, so editing one edits the series; zero otherwise.
	SeriesStartsAt time.Time
	SeriesEndsAt   time.Time
}

func (t Task) Done() bool {
	return t.DoneAt != nil
}

// ItemInput is an item as edited in the form; ID is 0 for new items.
type ItemInput struct {
	ID    int64
	Title string
}

// Input holds the user-editable fields of a task.
type Input struct {
	Title    string
	Category string
	Urgency  Urgency
	StartsAt time.Time
	EndsAt   time.Time
	Repeat   Repeat
	Items    []ItemInput
}
