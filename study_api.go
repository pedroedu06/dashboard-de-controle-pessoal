package main

import (
	"time"

	"dashboard/internal/studies"
)

// StudyAPI exposes the studies feature to the frontend. Instants cross the
// boundary as RFC 3339 strings.
type StudyAPI struct {
	app *App
}

type TopicDTO struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Status       string `json:"status"`       // not_seen, studying or reviewed
	NextReviewOn string `json:"nextReviewOn"` // YYYY-MM-DD, empty when not scheduled
}

type SubjectDTO struct {
	ID       int64      `json:"id"`
	Name     string     `json:"name"`
	Progress float64    `json:"progress"` // 0..1
	Topics   []TopicDTO `json:"topics"`
}

type StudySessionDTO struct {
	ID             int64  `json:"id"`
	SubjectID      int64  `json:"subjectId"`
	SubjectName    string `json:"subjectName"`
	TopicID        int64  `json:"topicId"` // 0 = no topic
	TopicTitle     string `json:"topicTitle"`
	StartedAt      string `json:"startedAt"`
	Paused         bool   `json:"paused"`
	ElapsedSeconds int    `json:"elapsedSeconds"` // as of the response
}

type GoalDTO struct {
	ID            int64  `json:"id"`
	SubjectID     int64  `json:"subjectId"` // 0 = all subjects
	SubjectName   string `json:"subjectName"`
	TargetMinutes int    `json:"targetMinutes"`
	DoneMinutes   int    `json:"doneMinutes"`
	Completed     bool   `json:"completed"`
}

type StudyOverviewDTO struct {
	Subjects []SubjectDTO      `json:"subjects"`
	Archived []SubjectDTO      `json:"archived"`
	Active   *StudySessionDTO  `json:"active"`
	Goals    []GoalDTO         `json:"goals"`
	Sessions []StudySessionDTO `json:"sessions"`
}

type StopResultDTO struct {
	Saved   bool `json:"saved"`
	Minutes int  `json:"minutes"`
}

func (api *StudyAPI) Overview() (StudyOverviewDTO, error) {
	s, err := api.service()
	if err != nil {
		return StudyOverviewDTO{}, err
	}
	now := time.Now()
	ov, err := s.Overview(api.app.ctx, now)
	if err != nil {
		return StudyOverviewDTO{}, err
	}

	out := StudyOverviewDTO{
		Subjects: make([]SubjectDTO, len(ov.Subjects)),
		Archived: make([]SubjectDTO, len(ov.Archived)),
		Goals:    make([]GoalDTO, len(ov.Goals)),
		Sessions: make([]StudySessionDTO, len(ov.Sessions)),
	}
	for i, sub := range ov.Subjects {
		out.Subjects[i] = toSubjectDTO(sub)
	}
	for i, sub := range ov.Archived {
		out.Archived[i] = toSubjectDTO(sub)
	}
	if ov.Active != nil {
		dto := toSessionDTO(*ov.Active, now)
		out.Active = &dto
	}
	for i, g := range ov.Goals {
		out.Goals[i] = GoalDTO{
			ID: g.ID, SubjectID: deref(g.SubjectID), SubjectName: g.SubjectName,
			TargetMinutes: g.TargetMinutes, DoneMinutes: g.DoneMinutes, Completed: g.Completed,
		}
	}
	for i, sess := range ov.Sessions {
		out.Sessions[i] = toSessionDTO(sess, now)
	}
	return out, nil
}

func (api *StudyAPI) CreateSubject(name string) (SubjectDTO, error) {
	s, err := api.service()
	if err != nil {
		return SubjectDTO{}, err
	}
	sub, err := s.CreateSubject(api.app.ctx, name)
	return toSubjectDTO(sub), err
}

func (api *StudyAPI) RenameSubject(id int64, name string) error {
	return api.do(func(s *studies.Service) error { return s.RenameSubject(api.app.ctx, id, name) })
}

func (api *StudyAPI) SetSubjectArchived(id int64, archived bool) error {
	return api.do(func(s *studies.Service) error { return s.SetSubjectArchived(api.app.ctx, id, archived) })
}

func (api *StudyAPI) CreateTopic(subjectID int64, title string) error {
	return api.do(func(s *studies.Service) error {
		_, err := s.CreateTopic(api.app.ctx, subjectID, title)
		return err
	})
}

func (api *StudyAPI) RenameTopic(id int64, title string) error {
	return api.do(func(s *studies.Service) error { return s.RenameTopic(api.app.ctx, id, title) })
}

func (api *StudyAPI) SetTopicStatus(id int64, status string) error {
	return api.do(func(s *studies.Service) error {
		return s.SetTopicStatus(api.app.ctx, time.Now(), id, studies.Status(status))
	})
}

func (api *StudyAPI) DeleteTopic(id int64) error {
	return api.do(func(s *studies.Service) error { return s.DeleteTopic(api.app.ctx, id) })
}

// StartSession starts the timer; topicID 0 means no topic.
func (api *StudyAPI) StartSession(subjectID, topicID int64) error {
	return api.do(func(s *studies.Service) error {
		_, err := s.Start(api.app.ctx, time.Now(), subjectID, optionalID(topicID))
		return err
	})
}

func (api *StudyAPI) PauseSession() error {
	return api.do(func(s *studies.Service) error {
		_, err := s.Pause(api.app.ctx, time.Now())
		return err
	})
}

func (api *StudyAPI) ResumeSession() error {
	return api.do(func(s *studies.Service) error {
		_, err := s.Resume(api.app.ctx, time.Now())
		return err
	})
}

// StopSession ends the timer; sessions under a minute are discarded.
func (api *StudyAPI) StopSession() (StopResultDTO, error) {
	s, err := api.service()
	if err != nil {
		return StopResultDTO{}, err
	}
	now := time.Now()
	session, saved, err := s.Stop(api.app.ctx, now)
	if err != nil {
		return StopResultDTO{}, err
	}
	return StopResultDTO{Saved: saved, Minutes: int(session.Elapsed(now) / time.Minute)}, nil
}

func (api *StudyAPI) DeleteSession(id int64) error {
	return api.do(func(s *studies.Service) error { return s.DeleteSession(api.app.ctx, id) })
}

// CreateGoal adds a weekly goal in hours; subjectID 0 means all subjects.
func (api *StudyAPI) CreateGoal(subjectID int64, hours float64) error {
	return api.do(func(s *studies.Service) error {
		_, err := s.CreateGoal(api.app.ctx, optionalID(subjectID), studies.HoursToMinutes(hours))
		return err
	})
}

func (api *StudyAPI) UpdateGoal(id int64, hours float64) error {
	return api.do(func(s *studies.Service) error { return s.UpdateGoal(api.app.ctx, id, studies.HoursToMinutes(hours)) })
}

func (api *StudyAPI) DeleteGoal(id int64) error {
	return api.do(func(s *studies.Service) error { return s.DeleteGoal(api.app.ctx, id) })
}

type ReviewDTO struct {
	TopicID      int64  `json:"topicId"`
	TopicTitle   string `json:"topicTitle"`
	SubjectName  string `json:"subjectName"`
	DueOn        string `json:"dueOn"` // YYYY-MM-DD
	DaysOverdue  int    `json:"daysOverdue"`
	Repetitions  int    `json:"repetitions"`
	IntervalDays int    `json:"intervalDays"`
}

type ReviewResultDTO struct {
	NextReviewOn string `json:"nextReviewOn"`
	IntervalDays int    `json:"intervalDays"`
}

type QuestionLogInputDTO struct {
	SubjectID int64  `json:"subjectId"`
	TopicID   int64  `json:"topicId"` // 0 = no topic
	DoneOn    string `json:"doneOn"`  // YYYY-MM-DD, empty = today
	Total     int    `json:"total"`
	Correct   int    `json:"correct"`
	Source    string `json:"source"`
}

type QuestionLogDTO struct {
	ID          int64  `json:"id"`
	SubjectName string `json:"subjectName"`
	TopicTitle  string `json:"topicTitle"`
	DoneOn      string `json:"doneOn"`
	Total       int    `json:"total"`
	Correct     int    `json:"correct"`
	Source      string `json:"source"`
}

type RateDTO struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	SubjectName string  `json:"subjectName"`
	Total       int     `json:"total"`
	Correct     int     `json:"correct"`
	Rate        float64 `json:"rate"` // 0..1
}

type QuestionReportDTO struct {
	BySubject  []RateDTO        `json:"bySubject"`
	ByTopic    []RateDTO        `json:"byTopic"`
	Weaknesses []RateDTO        `json:"weaknesses"`
	Logs       []QuestionLogDTO `json:"logs"`
}

// DueReviews lists the topic reviews due today, most overdue first.
func (api *StudyAPI) DueReviews() ([]ReviewDTO, error) {
	s, err := api.service()
	if err != nil {
		return nil, err
	}
	due, err := s.DueReviews(api.app.ctx, time.Now())
	if err != nil {
		return nil, err
	}
	out := make([]ReviewDTO, len(due))
	for i, d := range due {
		out[i] = ReviewDTO{
			TopicID: d.ID, TopicTitle: d.Title, SubjectName: d.SubjectName, DueOn: dayString(d.Schedule.NextReviewOn),
			DaysOverdue: d.DaysOverdue, Repetitions: d.Schedule.Repetitions, IntervalDays: d.Schedule.IntervalDays,
		}
	}
	return out, nil
}

// ReviewTopic grades a review: forgot, hard, ok or easy.
func (api *StudyAPI) ReviewTopic(topicID int64, grade string) (ReviewResultDTO, error) {
	s, err := api.service()
	if err != nil {
		return ReviewResultDTO{}, err
	}
	next, err := s.Review(api.app.ctx, time.Now(), topicID, studies.Grade(grade))
	if err != nil {
		return ReviewResultDTO{}, err
	}
	return ReviewResultDTO{NextReviewOn: dayString(next.NextReviewOn), IntervalDays: next.IntervalDays}, nil
}

func (api *StudyAPI) LogQuestions(in QuestionLogInputDTO) error {
	return api.do(func(s *studies.Service) error {
		log, err := parseQuestionLog(in)
		if err != nil {
			return err
		}
		_, err = s.LogQuestions(api.app.ctx, time.Now(), log)
		return err
	})
}

func (api *StudyAPI) DeleteQuestionLog(id int64) error {
	return api.do(func(s *studies.Service) error { return s.DeleteQuestionLog(api.app.ctx, id) })
}

// QuestionReport covers the last periodDays days, or everything when 0.
func (api *StudyAPI) QuestionReport(periodDays int) (QuestionReportDTO, error) {
	s, err := api.service()
	if err != nil {
		return QuestionReportDTO{}, err
	}
	report, err := s.QuestionReport(api.app.ctx, time.Now(), periodDays)
	if err != nil {
		return QuestionReportDTO{}, err
	}
	out := QuestionReportDTO{
		BySubject:  toRateDTOs(report.BySubject),
		ByTopic:    toRateDTOs(report.ByTopic),
		Weaknesses: toRateDTOs(report.Weaknesses),
		Logs:       make([]QuestionLogDTO, len(report.Logs)),
	}
	for i, l := range report.Logs {
		out.Logs[i] = QuestionLogDTO{
			ID: l.ID, SubjectName: l.SubjectName, TopicTitle: l.TopicTitle, DoneOn: l.DoneOn.Format(time.DateOnly),
			Total: l.Total, Correct: l.Correct, Source: l.Source,
		}
	}
	return out, nil
}

func parseQuestionLog(in QuestionLogInputDTO) (studies.QuestionLog, error) {
	log := studies.QuestionLog{
		SubjectID: in.SubjectID, TopicID: optionalID(in.TopicID),
		Total: in.Total, Correct: in.Correct, Source: in.Source,
	}
	day, err := parseDay(in.DoneOn, "questões")
	if err != nil {
		return studies.QuestionLog{}, err
	}
	if day != nil {
		log.DoneOn = *day
	}
	return log, nil
}

func toRateDTOs(rows []studies.RateRow) []RateDTO {
	out := make([]RateDTO, len(rows))
	for i, r := range rows {
		out[i] = RateDTO{ID: r.ID, Name: r.Name, SubjectName: r.SubjectName, Total: r.Total, Correct: r.Correct, Rate: r.Rate()}
	}
	return out
}

// dayString formats a calendar day as YYYY-MM-DD, or "" when nil.
func dayString(day *time.Time) string {
	if day == nil {
		return ""
	}
	return day.Format(time.DateOnly)
}

func (api *StudyAPI) service() (*studies.Service, error) {
	if api.app.studies == nil {
		return nil, errDBUnavailable
	}
	return api.app.studies, nil
}

func (api *StudyAPI) do(fn func(*studies.Service) error) error {
	s, err := api.service()
	if err != nil {
		return err
	}
	return fn(s)
}

func toSubjectDTO(sub studies.Subject) SubjectDTO {
	dto := SubjectDTO{ID: sub.ID, Name: sub.Name, Progress: studies.Progress(sub.Topics), Topics: make([]TopicDTO, len(sub.Topics))}
	for i, t := range sub.Topics {
		dto.Topics[i] = TopicDTO{ID: t.ID, Title: t.Title, Status: string(t.Status), NextReviewOn: dayString(t.Schedule.NextReviewOn)}
	}
	return dto
}

func toSessionDTO(s studies.Session, now time.Time) StudySessionDTO {
	return StudySessionDTO{
		ID:             s.ID,
		SubjectID:      s.SubjectID,
		SubjectName:    s.SubjectName,
		TopicID:        deref(s.TopicID),
		TopicTitle:     s.TopicTitle,
		StartedAt:      s.StartedAt.Local().Format(time.RFC3339),
		Paused:         s.Paused(),
		ElapsedSeconds: int(s.Elapsed(now) / time.Second),
	}
}

func optionalID(id int64) *int64 {
	if id == 0 {
		return nil
	}
	return &id
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
