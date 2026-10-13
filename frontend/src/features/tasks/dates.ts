const pad = (n: number) => String(n).padStart(2, '0');

/** Formats a Date as the value of an `<input type="datetime-local">`. */
export function toLocalInput(date: Date): string {
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

/** Converts a `datetime-local` value (local time) to RFC 3339. */
export function localInputToISO(value: string): string {
    return new Date(value).toISOString();
}

/** Default schedule for a new task: next full hour, lasting one hour. */
export function defaultSchedule(now = new Date()): {startsAt: string; endsAt: string} {
    const start = new Date(now);
    start.setHours(start.getHours() + 1, 0, 0, 0);
    const end = new Date(start);
    end.setHours(end.getHours() + 1);
    return {startsAt: toLocalInput(start), endsAt: toLocalInput(end)};
}

/** Parses a `YYYY-MM-DD` day as local midnight. */
export function parseDay(day: string): Date {
    const [y, m, d] = day.split('-').map(Number);
    return new Date(y, m - 1, d);
}

export function dayLabel(day: string, now = new Date()): string {
    const date = parseDay(day);
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const diffDays = Math.round((date.getTime() - today.getTime()) / 86_400_000);
    if (diffDays === 0) return 'Hoje';
    if (diffDays === 1) return 'Amanhã';
    return date.toLocaleDateString('pt-BR', {weekday: 'long', day: '2-digit', month: '2-digit'});
}

const time = (d: Date) => `${pad(d.getHours())}:${pad(d.getMinutes())}`;
const dateTime = (d: Date) => `${pad(d.getDate())}/${pad(d.getMonth() + 1)} ${time(d)}`;

/** "14:00 – 15:00" on the same day, otherwise with dates. */
export function scheduleLabel(startsAt: string, endsAt: string): string {
    const start = new Date(startsAt);
    const end = new Date(endsAt);
    if (start.toDateString() === end.toDateString()) {
        return `${time(start)} – ${time(end)}`;
    }
    return `${dateTime(start)} – ${dateTime(end)}`;
}
