import * as StudyAPI from '../../../wailsjs/go/main/StudyAPI';

export type TopicStatus = 'not_seen' | 'studying' | 'reviewed';

export interface Topic {
    id: number;
    title: string;
    status: TopicStatus;
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
};
