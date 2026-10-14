// Package studies implements study management: subjects and topics, the
// study timer, topic progress and weekly goals.
package studies

import (
	"errors"
	"time"
)

type Status string

const (
	StatusNotSeen  Status = "not_seen"
	StatusStudying Status = "studying"
	StatusReviewed Status = "reviewed"
)

func (s Status) Valid() bool {
	switch s {
	case StatusNotSeen, StatusStudying, StatusReviewed:
		return true
	}
	return false
}

// User-facing errors, shown as-is by the frontend.
var (
	ErrNameRequired      = errors.New("informe o nome da matéria")
	ErrTitleRequired     = errors.New("informe o nome do tópico")
	ErrSubjectExists     = errors.New("já existe uma matéria com esse nome")
	ErrSubjectNotFound   = errors.New("matéria não encontrada")
	ErrSubjectArchived   = errors.New("esta matéria está arquivada")
	ErrTopicNotFound     = errors.New("tópico não encontrado")
	ErrTopicOtherSubject = errors.New("este tópico é de outra matéria")
	ErrInvalidStatus     = errors.New("status inválido: use não visto, estudando ou revisado")
	ErrSessionRunning    = errors.New("já existe um timer ligado: encerre-o antes de começar outro")
	ErrArchiveRunning    = errors.New("encerre o timer desta matéria antes de arquivá-la")
	ErrNoActiveSession   = errors.New("nenhum timer ligado")
	ErrAlreadyPaused     = errors.New("o timer já está pausado")
	ErrNotPaused         = errors.New("o timer não está pausado")
	ErrSessionNotFound   = errors.New("sessão não encontrada")
	ErrGoalNotFound      = errors.New("meta não encontrada")
	ErrInvalidGoalTarget = errors.New("a meta deve ter entre 1 minuto e 168 horas por semana")
)

type Topic struct {
	ID        int64
	SubjectID int64
	Title     string
	Status    Status
}

type Subject struct {
	ID       int64
	Name     string
	Archived bool
	Topics   []Topic
}

// Session is a study timer run. EndedAt is nil while it is running and
// PausedAt is set while it is paused.
type Session struct {
	ID            int64
	SubjectID     int64
	SubjectName   string
	TopicID       *int64
	TopicTitle    string
	StartedAt     time.Time
	EndedAt       *time.Time
	PausedAt      *time.Time
	PausedSeconds int
}

// Goal is a weekly study target; SubjectID nil means all subjects.
type Goal struct {
	ID            int64
	SubjectID     *int64
	SubjectName   string
	TargetMinutes int
}

// MaxGoalMinutes is a whole week.
const MaxGoalMinutes = 7 * 24 * 60
