import * as TaskAPI from '../../../wailsjs/go/main/TaskAPI';

export type Urgency = 'low' | 'medium' | 'high';

export interface Task {
    id: number;
    title: string;
    urgency: Urgency;
    startsAt: string; // RFC 3339
    endsAt: string;
    done: boolean;
}

export interface Day {
    date: string; // YYYY-MM-DD
    tasks: Task[];
}

export interface Agenda {
    overdue: Task[];
    days: Day[];
}

export interface TaskInput {
    title: string;
    urgency: Urgency;
    startsAt: string; // RFC 3339
    endsAt: string;
}

export const urgencyLabels: Record<Urgency, string> = {
    low: 'Baixa',
    medium: 'Média',
    high: 'Alta',
};

export const tasksApi = {
    agenda: () => TaskAPI.Agenda() as Promise<Agenda>,
    create: (input: TaskInput) => TaskAPI.Create(input) as Promise<Task>,
    update: (id: number, input: TaskInput) => TaskAPI.Update(id, input) as Promise<Task>,
    remove: (id: number) => TaskAPI.Delete(id),
    setDone: (id: number, done: boolean) => TaskAPI.SetDone(id, done) as Promise<Task>,
};
