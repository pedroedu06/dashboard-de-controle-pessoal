package studies

import (
	"cmp"
	"slices"
	"strings"
	"time"
)

const (
	MaxQuestionsPerLog   = 1000
	WeaknessMinQuestions = 10 // a topic needs this many questions to count as a weakness
	WeaknessLimit        = 5
)

// QuestionLog records a batch of practice questions.
type QuestionLog struct {
	ID          int64
	SubjectID   int64
	SubjectName string
	TopicID     *int64
	TopicTitle  string
	DoneOn      time.Time // calendar day
	Total       int
	Correct     int
	Source      string // optional, e.g. "Simulado FGV 03"
}

// Validate trims the source and checks the numbers and the day.
func (l QuestionLog) Validate(today time.Time) (QuestionLog, error) {
	l.Source = strings.TrimSpace(l.Source)
	if l.Total < 1 || l.Total > MaxQuestionsPerLog {
		return QuestionLog{}, ErrInvalidQuestionCount
	}
	if l.Correct < 0 || l.Correct > l.Total {
		return QuestionLog{}, ErrCorrectOverTotal
	}
	if l.DoneOn.IsZero() {
		l.DoneOn = startOfDay(today)
	}
	if civilDay(l.DoneOn, today.Location()).After(startOfDay(today)) {
		return QuestionLog{}, ErrFutureDate
	}
	return l, nil
}

// RateRow is the accuracy of a subject or topic.
type RateRow struct {
	ID          int64
	Name        string
	SubjectName string // for topics
	Total       int
	Correct     int
}

// Rate is the share of correct answers, from 0 to 1.
func (r RateRow) Rate() float64 {
	if r.Total == 0 {
		return 0
	}
	return float64(r.Correct) / float64(r.Total)
}

type QuestionStats struct {
	BySubject  []RateRow
	ByTopic    []RateRow
	Weaknesses []RateRow // lowest accuracy topics with enough questions
}

// Stats aggregates logs per subject and per topic. Logs without a topic
// count only for their subject.
func Stats(logs []QuestionLog) QuestionStats {
	subjects := map[int64]*RateRow{}
	topics := map[int64]*RateRow{}
	for _, l := range logs {
		add(subjects, l.SubjectID, RateRow{ID: l.SubjectID, Name: l.SubjectName}, l)
		if l.TopicID != nil {
			add(topics, *l.TopicID, RateRow{ID: *l.TopicID, Name: l.TopicTitle, SubjectName: l.SubjectName}, l)
		}
	}

	stats := QuestionStats{BySubject: values(subjects), ByTopic: values(topics), Weaknesses: []RateRow{}}
	slices.SortFunc(stats.BySubject, func(a, b RateRow) int {
		return cmp.Or(compareNames(a.Name, b.Name), cmp.Compare(a.ID, b.ID))
	})
	slices.SortFunc(stats.ByTopic, func(a, b RateRow) int {
		return cmp.Or(compareNames(a.SubjectName, b.SubjectName), compareNames(a.Name, b.Name), cmp.Compare(a.ID, b.ID))
	})

	for _, t := range stats.ByTopic {
		if t.Total >= WeaknessMinQuestions {
			stats.Weaknesses = append(stats.Weaknesses, t)
		}
	}
	slices.SortStableFunc(stats.Weaknesses, func(a, b RateRow) int {
		return cmp.Or(cmp.Compare(a.Rate(), b.Rate()), cmp.Compare(b.Total, a.Total))
	})
	if len(stats.Weaknesses) > WeaknessLimit {
		stats.Weaknesses = stats.Weaknesses[:WeaknessLimit]
	}
	return stats
}

func add(rows map[int64]*RateRow, id int64, blank RateRow, l QuestionLog) {
	row, ok := rows[id]
	if !ok {
		row = &blank
		rows[id] = row
	}
	row.Total += l.Total
	row.Correct += l.Correct
}

func values(rows map[int64]*RateRow) []RateRow {
	out := make([]RateRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, *r)
	}
	return out
}

func compareNames(a, b string) int {
	return strings.Compare(strings.ToLower(a), strings.ToLower(b))
}
