import {FormEvent, useState} from 'react';
import {RepeatKind, Task, TaskInput, Urgency, urgencyLabels, weekdayLabels} from './api';
import {defaultSchedule, localInputToISO, toLocalInput} from './dates';
import {EditorItem, ItemsEditor, newEditorItem} from './ItemsEditor';

interface Props {
    task?: Task; // set when editing
    categories: string[];
    onSubmit: (input: TaskInput) => Promise<void>;
    onClose: () => void;
}

interface Values {
    title: string;
    category: string;
    urgency: Urgency;
    startsAt: string; // datetime-local
    endsAt: string;
    hasItems: boolean;
    items: EditorItem[];
    repeats: boolean;
    repeatKind: Exclude<RepeatKind, 'none'>;
    weekdays: number[];
    until: string; // YYYY-MM-DD or empty
}

function initialValues(task?: Task): Values {
    if (!task) {
        return {
            title: '', category: '', urgency: 'medium', ...defaultSchedule(),
            hasItems: false, items: [],
            repeats: false, repeatKind: 'daily', weekdays: [], until: '',
        };
    }
    // Recurring tasks are edited as a series, from their first occurrence.
    return {
        title: task.title,
        category: task.category,
        urgency: task.urgency,
        startsAt: toLocalInput(new Date(task.seriesStartsAt)),
        endsAt: toLocalInput(new Date(task.seriesEndsAt)),
        hasItems: task.items.length > 0,
        items: task.items.map((it) => newEditorItem(it.title, it.id)),
        repeats: task.repeat.kind !== 'none',
        repeatKind: task.repeat.kind === 'weekly' ? 'weekly' : 'daily',
        weekdays: task.repeat.weekdays,
        until: task.repeat.until,
    };
}

/** Client-side checks for quick feedback; the Go side is the source of truth. */
function validate(v: Values): string | null {
    if (!v.title.trim()) return 'Informe o nome da tarefa.';
    if (!v.startsAt || !v.endsAt) return 'Informe o início e o fim da tarefa.';
    if (new Date(v.endsAt) <= new Date(v.startsAt)) return 'O fim da tarefa deve ser depois do início.';
    if (v.hasItems && v.items.some((it) => !it.title.trim())) return 'Todo item da lista precisa de um nome.';
    if (v.repeats) {
        if (v.startsAt.slice(0, 10) !== v.endsAt.slice(0, 10)) return 'Uma tarefa recorrente deve começar e terminar no mesmo dia.';
        if (v.repeatKind === 'weekly' && v.weekdays.length === 0) return 'Escolha pelo menos um dia da semana.';
        if (v.until && v.until < v.startsAt.slice(0, 10)) return 'A data final da repetição deve ser depois do início.';
    }
    return null;
}

function toInput(v: Values): TaskInput {
    return {
        title: v.title.trim(),
        category: v.category.trim(),
        urgency: v.urgency,
        startsAt: localInputToISO(v.startsAt),
        endsAt: localInputToISO(v.endsAt),
        repeat: v.repeats
            ? {kind: v.repeatKind, weekdays: v.repeatKind === 'weekly' ? [...v.weekdays].sort((a, b) => a - b) : [], until: v.until}
            : {kind: 'none', weekdays: [], until: ''},
        items: v.hasItems ? v.items.map((it) => ({id: it.id, title: it.title.trim()})) : [],
    };
}

export function TaskFormModal({task, categories, onSubmit, onClose}: Props) {
    const [values, setValues] = useState(() => initialValues(task));
    const [error, setError] = useState<string | null>(null);
    const [saving, setSaving] = useState(false);

    const update = (patch: Partial<Values>) => setValues((v) => ({...v, ...patch}));
    const toggleWeekday = (day: number) => update({
        weekdays: values.weekdays.includes(day) ? values.weekdays.filter((d) => d !== day) : [...values.weekdays, day],
    });

    async function handleSubmit(e: FormEvent) {
        e.preventDefault();
        const invalid = validate(values);
        if (invalid) {
            setError(invalid);
            return;
        }
        setSaving(true);
        setError(null);
        try {
            await onSubmit(toInput(values));
        } catch (err) {
            setError(String(err));
            setSaving(false);
        }
    }

    const heading = task ? 'Editar tarefa' : 'Nova tarefa';

    return (
        <dialog open aria-modal="true" aria-label={heading}>
            <form onSubmit={handleSubmit}>
                <h2>{heading}</h2>

                <p>
                    <label>
                        Nome <input value={values.title} onChange={(e) => update({title: e.target.value})} autoFocus/>
                    </label>
                </p>
                <p>
                    <label>
                        Categoria{' '}
                        <input
                            list="task-categories"
                            placeholder="Ex.: Faculdade"
                            value={values.category}
                            onChange={(e) => update({category: e.target.value})}
                        />
                    </label>
                    <datalist id="task-categories">
                        {categories.map((c) => <option key={c} value={c}/>)}
                    </datalist>
                </p>
                <p>
                    <label>
                        Início <input type="datetime-local" value={values.startsAt} onChange={(e) => update({startsAt: e.target.value})}/>
                    </label>
                </p>
                <p>
                    <label>
                        Fim <input type="datetime-local" value={values.endsAt} onChange={(e) => update({endsAt: e.target.value})}/>
                    </label>
                </p>
                <p>
                    <label>
                        Urgência{' '}
                        <select value={values.urgency} onChange={(e) => update({urgency: e.target.value as Urgency})}>
                            {Object.entries(urgencyLabels).map(([value, label]) => (
                                <option key={value} value={value}>{label}</option>
                            ))}
                        </select>
                    </label>
                </p>

                <p>
                    <label>
                        <input type="checkbox" checked={values.hasItems} onChange={(e) => update({hasItems: e.target.checked})}/>
                        {' '}Tem lista de itens
                    </label>
                </p>
                {values.hasItems && <ItemsEditor items={values.items} onChange={(items) => update({items})}/>}

                <p>
                    <label>
                        <input type="checkbox" checked={values.repeats} onChange={(e) => update({repeats: e.target.checked})}/>
                        {' '}Repetir
                    </label>
                </p>
                {values.repeats && (
                    <fieldset>
                        <legend>Repetição</legend>
                        <label>
                            Frequência{' '}
                            <select value={values.repeatKind} onChange={(e) => update({repeatKind: e.target.value as Values['repeatKind']})}>
                                <option value="daily">Diária</option>
                                <option value="weekly">Semanal</option>
                            </select>
                        </label>
                        {values.repeatKind === 'weekly' && (
                            <p>
                                {weekdayLabels.map((label, day) => (
                                    <label key={day}>
                                        <input type="checkbox" checked={values.weekdays.includes(day)} onChange={() => toggleWeekday(day)}/>
                                        {label}{' '}
                                    </label>
                                ))}
                            </p>
                        )}
                        <p>
                            <label>
                                Repetir até (opcional){' '}
                                <input type="date" value={values.until} onChange={(e) => update({until: e.target.value})}/>
                            </label>
                        </p>
                    </fieldset>
                )}

                {error && <p role="alert">{error}</p>}

                <button type="submit" disabled={saving}>Salvar</button>{' '}
                <button type="button" onClick={onClose}>Cancelar</button>
            </form>
        </dialog>
    );
}
