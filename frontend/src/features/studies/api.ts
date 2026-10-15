import * as StudyAPI from '../../../wailsjs/go/main/StudyAPI';
import {main} from '../../../wailsjs/go/models';

export type TopicStatus = 'not_seen' | 'studying' | 'reviewed';

export interface Topic {
    id: number;
    title: string;
    status: TopicStatus;
    nextReviewOn: string; // YYYY-MM-DD, empty when not in the review queue
}

export interface Subject {
    id: number;
    name: string;
    progress: number; // 0..1
    topics: Topic[];
}

export interface StudySession {
    id: number;
    subjectId: number;
    subjectName: string;
    topicId: number; // 0 = no topic
    topicTitle: string;
    startedAt: string; // RFC 3339
    paused: boolean;
    elapsedSeconds: number; // as of the response
}

export interface Goal {
    id: number;
    subjectId: number; // 0 = all subjects
    subjectName: string;
    targetMinutes: number;
    doneMinutes: number;
    completed: boolean;
}

export interface StudyOverview {
    subjects: Subject[];
    archived: Subject[];
    active: StudySession | null;
    goals: Goal[];
    sessions: StudySession[];
}

export interface StopResult {
    saved: boolean;
    minutes: number;
}

export type Grade = 'forgot' | 'hard' | 'ok' | 'easy';

export interface DueReview {
    topicId: number;
    topicTitle: string;
    subjectName: string;
    dueOn: string; // YYYY-MM-DD
    daysOverdue: number;
    repetitions: number;
    intervalDays: number;
}

export interface ReviewResult {
    nextReviewOn: string;
    intervalDays: number;
}

export interface QuestionLogInput {
    subjectId: number;
    topicId: number; // 0 = no topic
    doneOn: string; // YYYY-MM-DD
    total: number;
    correct: number;
    source: string;
}

export interface QuestionLog {
    id: number;
    subjectName: string;
    topicTitle: string;
    doneOn: string;
    total: number;
    correct: number;
    source: string;
}

export interface RateRow {
    id: number;
    name: string;
    subjectName: string;
    total: number;
    correct: number;
    rate: number; // 0..1
}

export interface QuestionReport {
    bySubject: RateRow[];
    byTopic: RateRow[];
    weaknesses: RateRow[];
    logs: QuestionLog[];
}

export const gradeLabels: Record<Grade, string> = {
    forgot: 'Esqueci',
    hard: 'Difícil',
    ok: 'OK',
    easy: 'Fácil',
};

export const statusLabels: Record<TopicStatus, string> = {
    not_seen: 'Não visto',
    studying: 'Estudando',
    reviewed: 'Revisado',
};

const pad = (n: number) => String(n).padStart(2, '0');

/** 3725 → "01:02:05" */
export function formatClock(totalSeconds: number): string {
    const s = Math.max(0, Math.floor(totalSeconds));
    return `${pad(Math.floor(s / 3600))}:${pad(Math.floor(s / 60) % 60)}:${pad(s % 60)}`;
}

/** 130 → "2h10", 45 → "45min", 120 → "2h" */
export function formatMinutes(minutes: number): string {
    const h = Math.floor(minutes / 60);
    const m = minutes % 60;
    if (h === 0) return `${m}min`;
    return m === 0 ? `${h}h` : `${h}h${pad(m)}`;
}

/** 0.756 → "76%" */
export function formatPercent(rate: number): string {
    return `${Math.round(rate * 100)}%`;
}

/** "2026-10-16" → "16/10" */
export function formatDay(day: string): string {
    const [, m, d] = day.split('-');
    return `${d}/${m}`;
}

export const studiesApi = {
    overview: () => StudyAPI.Overview() as Promise<StudyOverview>,
    createSubject: (name: string) => StudyAPI.CreateSubject(name),
    renameSubject: (id: number, name: string) => StudyAPI.RenameSubject(id, name),
    setSubjectArchived: (id: number, archived: boolean) => StudyAPI.SetSubjectArchived(id, archived),
    createTopic: (subjectId: number, title: string) => StudyAPI.CreateTopic(subjectId, title),
    renameTopic: (id: number, title: string) => StudyAPI.RenameTopic(id, title),
    setTopicStatus: (id: number, status: TopicStatus) => StudyAPI.SetTopicStatus(id, status),
    deleteTopic: (id: number) => StudyAPI.DeleteTopic(id),
    start: (subjectId: number, topicId: number) => StudyAPI.StartSession(subjectId, topicId),
    pause: () => StudyAPI.PauseSession(),
    resume: () => StudyAPI.ResumeSession(),
    stop: () => StudyAPI.StopSession() as Promise<StopResult>,
    deleteSession: (id: number) => StudyAPI.DeleteSession(id),
    createGoal: (subjectId: number, hours: number) => StudyAPI.CreateGoal(subjectId, hours),
    updateGoal: (id: number, hours: number) => StudyAPI.UpdateGoal(id, hours),
    deleteGoal: (id: number) => StudyAPI.DeleteGoal(id),
    dueReviews: () => StudyAPI.DueReviews() as Promise<DueReview[]>,
    reviewTopic: (topicId: number, grade: Grade) => StudyAPI.ReviewTopic(topicId, grade) as Promise<ReviewResult>,
    logQuestions: (input: QuestionLogInput) => StudyAPI.LogQuestions(main.QuestionLogInputDTO.createFrom(input)),
    deleteQuestionLog: (id: number) => StudyAPI.DeleteQuestionLog(id),
    questionReport: (periodDays: number) => StudyAPI.QuestionReport(periodDays) as Promise<QuestionReport>,
};
