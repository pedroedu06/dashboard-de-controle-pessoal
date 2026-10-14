import {describe, expect, it} from 'vitest';
import {formatClock, formatMinutes} from './api';

describe('formatters', () => {
    it('formats the timer clock', () => {
        expect(formatClock(0)).toBe('00:00:00');
        expect(formatClock(3725)).toBe('01:02:05');
        expect(formatClock(-5)).toBe('00:00:00');
    });

    it('formats minutes as hours', () => {
        expect(formatMinutes(45)).toBe('45min');
        expect(formatMinutes(120)).toBe('2h');
        expect(formatMinutes(130)).toBe('2h10');
    });
});
