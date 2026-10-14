import * as TaskAPI from '../../../wailsjs/go/main/TaskAPI';
import {main} from '../../../wailsjs/go/models';

export type Urgency = 'low' | 'medium' | 'high';
export type RepeatKind = 'none' | 'daily' | 'weekly';

export interface Repeat {
    kind: RepeatKind;
    weekdays: number[]; // 0 = Sunday .. 6 = Saturday
    until: string; // YYYY-MM-DD or empty
}

export interface Item {
    id: number;
    title: string;
    done: boolean;
}

export interface Task {
    id: number;
    title: string;
    category: string;
    urgency: Urgency;
    startsAt: string; // RFC 3339
    endsAt: string;
    seriesStartsAt: string; // schedule to edit (first occurrence for recurring tasks)
    seriesEndsAt: string;
    done: boolean;
    repeat: Repeat;
    items: Item[];
    occursOn: string; // YYYY-MM-DD for occurrences of recurring tasks, else empty
}

export interface Day {
    date: string; // YYYY-MM-DD
    tasks: Task[];
}

export interface Agenda {
    overdue: Task[];
    days: Day[];
}

export interface ItemInput {
    id: number; // 0 for new items
    title: string;
}

export interface TaskInput {
    title: string;
    category: string;
    urgency: Urgency;
    startsAt: string; // RFC 3339
    endsAt: string;
    repeat: Repeat;
    items: ItemInput[];
}

export interface Filter {
    category: string; // empty = all
    urgency: Urgency | ''; // empty = all
}

export const noRepeat: Repeat = {kind: 'none', weekdays: [], until: ''};
export const noFilter: Filter = {category: '', urgency: ''};

export const urgencyLabels: Record<Urgency, string> = {
    low: 'Baixa',
    medium: 'Média',
    high: 'Alta',
};

export const weekdayLabels = ['Dom', 'Seg', 'Ter', 'Qua', 'Qui', 'Sex', 'Sáb'];

export function repeatLabel(repeat: Repeat): string {
    if (repeat.kind === 'daily') return 'Repete: diária';
    if (repeat.kind === 'weekly') return `Repete: ${repeat.weekdays.map((d) => weekdayLabels[d].toLowerCase()).join(', ')}`;
    return '';
}

export const tasksApi = {
    agenda: (filter: Filter) => TaskAPI.Agenda(main.FilterDTO.createFrom(filter)) as Promise<Agenda>,
    categories: () => TaskAPI.Categories(),
    create: (input: TaskInput) => TaskAPI.Create(main.TaskInputDTO.createFrom(input)) as Promise<Task>,
    update: (id: number, input: TaskInput) => TaskAPI.Update(id, main.TaskInputDTO.createFrom(input)) as Promise<Task>,
    remove: (id: number) => TaskAPI.Delete(id),
    setDone: (task: Task, done: boolean) => TaskAPI.SetDone(task.id, task.occursOn, done),
    setItemDone: (task: Task, item: Item, done: boolean) => TaskAPI.SetItemDone(item.id, task.occursOn, done),
};
