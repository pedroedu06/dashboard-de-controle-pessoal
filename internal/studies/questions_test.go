package studies

import (
	"errors"
	"slices"
	"testing"
	"time"
)

func qlog(subject int64, subjectName string, topic int64, topicName string, total, correct int) QuestionLog {
	l := QuestionLog{SubjectID: subject, SubjectName: subjectName, Total: total, Correct: correct}
	if topic != 0 {
		l.TopicID = ptr(topic)
		l.TopicTitle = topicName
	}
	return l
}

func names(rows []RateRow) []string {
	out := []string{}
	for _, r := range rows {
		out = append(out, r.Name)
	}
	return out
}

func TestStats_AggregatesBySubjectAndTopic(t *testing.T) {
	stats := Stats([]QuestionLog{
		qlog(1, "Português", 10, "Crase", 10, 4),
		qlog(1, "Português", 10, "Crase", 10, 6),
		qlog(1, "Português", 0, "", 20, 18), // no topic: subject only
		qlog(2, "Direito", 20, "Atos", 5, 5),
	})

	if got := names(stats.BySubject); !slices.Equal(got, []string{"Direito", "Português"}) {
		t.Errorf("subjects = %v", got)
	}
	pt := stats.BySubject[1]
	if pt.Total != 40 || pt.Correct != 28 || pt.Rate() != 0.7 {
		t.Errorf("Português = %+v (rate %v)", pt, pt.Rate())
	}

	if got := names(stats.ByTopic); !slices.Equal(got, []string{"Atos", "Crase"}) {
		t.Errorf("topics = %v", got)
	}
	if crase := stats.ByTopic[1]; crase.Total != 20 || crase.Correct != 10 || crase.SubjectName != "Português" {
		t.Errorf("Crase = %+v", crase)
	}
}

func TestStats_Weaknesses(t *testing.T) {
	var logs []QuestionLog
	add := func(id int64, name string, total, correct int) {
		logs = append(logs, qlog(1, "Mat", id, name, total, correct))
	}
	add(1, "A", 10, 9)  // 90%
	add(2, "B", 10, 3)  // 30%
	add(3, "C", 9, 0)   // too few questions
	add(4, "D", 20, 10) // 50%
	add(5, "E", 10, 5)  // 50%, fewer questions than D
	add(6, "F", 10, 7)
	add(7, "G", 10, 8)

	got := names(Stats(logs).Weaknesses)
	if want := []string{"B", "D", "E", "F", "G"}; !slices.Equal(got, want) {
		t.Errorf("weaknesses = %v, want %v", got, want)
	}
}

func TestStats_Empty(t *testing.T) {
	stats := Stats(nil)
	if stats.BySubject == nil || stats.ByTopic == nil || stats.Weaknesses == nil {
		t.Errorf("stats = %+v, want empty non-nil slices", stats)
	}
}

func TestQuestionLog_Validate(t *testing.T) {
	today := at(13, 10, 0)
	valid := QuestionLog{SubjectID: 1, DoneOn: at(13, 0, 0), Total: 10, Correct: 7, Source: "  Simulado 1 "}

	got, err := valid.Validate(today)
	if err != nil || got.Source != "Simulado 1" {
		t.Fatalf("Validate = %+v, %v", got, err)
	}

	noDate := valid
	noDate.DoneOn = time.Time{}
	if got, _ := noDate.Validate(today); !got.DoneOn.Equal(at(13, 0, 0)) {
		t.Errorf("empty date: DoneOn = %v, want today", got.DoneOn)
	}

	tests := []struct {
		name   string
		modify func(*QuestionLog)
		want   error
	}{
		{"no questions", func(l *QuestionLog) { l.Total = 0 }, ErrInvalidQuestionCount},
		{"too many", func(l *QuestionLog) { l.Total = MaxQuestionsPerLog + 1 }, ErrInvalidQuestionCount},
		{"correct over total", func(l *QuestionLog) { l.Correct = 11 }, ErrCorrectOverTotal},
		{"negative correct", func(l *QuestionLog) { l.Correct = -1 }, ErrCorrectOverTotal},
		{"future", func(l *QuestionLog) { l.DoneOn = at(14, 0, 0) }, ErrFutureDate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := valid
			tt.modify(&l)
			if _, err := l.Validate(today); !errors.Is(err, tt.want) {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}
}
