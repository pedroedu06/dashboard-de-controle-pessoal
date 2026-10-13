import {describe, expect, it} from 'vitest';
import {dayLabel, defaultSchedule, localInputToISO, scheduleLabel, toLocalInput} from './dates';

describe('dates', () => {
    it('round-trips datetime-local values', () => {
        const value = '2026-10-13T14:30';

        expect(toLocalInput(new Date(localInputToISO(value)))).toBe(value);
    });

    it('defaults to the next full hour, lasting one hour', () => {
        const schedule = defaultSchedule(new Date(2026, 9, 13, 14, 25));

        expect(schedule).toEqual({startsAt: '2026-10-13T15:00', endsAt: '2026-10-13T16:00'});
    });

    it('defaults across midnight', () => {
        const schedule = defaultSchedule(new Date(2026, 9, 13, 23, 10));

        expect(schedule).toEqual({startsAt: '2026-10-14T00:00', endsAt: '2026-10-14T01:00'});
    });

    it('labels today and tomorrow', () => {
        const now = new Date(2026, 9, 13, 12, 0);

        expect(dayLabel('2026-10-13', now)).toBe('Hoje');
        expect(dayLabel('2026-10-14', now)).toBe('Amanhã');
        expect(dayLabel('2026-10-16', now)).toMatch(/16\/10/);
    });

    it('shows only times for same-day tasks and dates otherwise', () => {
        const start = new Date(2026, 9, 13, 14, 0).toISOString();
        const sameDay = new Date(2026, 9, 13, 15, 30).toISOString();
        const nextDay = new Date(2026, 9, 15, 9, 0).toISOString();

        expect(scheduleLabel(start, sameDay)).toBe('14:00 – 15:30');
        expect(scheduleLabel(start, nextDay)).toBe('13/10 14:00 – 15/10 09:00');
    });
});
