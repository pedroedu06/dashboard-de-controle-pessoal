import {noRepeat, Task} from './api';

/** Builds a task for tests: one hour starting now, no items, no repeat. */
export function makeTask(overrides: Partial<Task> = {}): Task {
    const startsAt = overrides.startsAt ?? new Date().toISOString();
    const endsAt = overrides.endsAt ?? new Date(new Date(startsAt).getTime() + 3_600_000).toISOString();
    return {
        id: 1,
        title: 'Tarefa',
        category: '',
        urgency: 'medium',
        startsAt,
        endsAt,
        seriesStartsAt: startsAt,
        seriesEndsAt: endsAt,
        done: false,
        repeat: noRepeat,
        items: [],
        occursOn: '',
        ...overrides,
    };
}
