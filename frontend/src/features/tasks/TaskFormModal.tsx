import {FormEvent, useState} from 'react';
import {Task, TaskInput, Urgency, urgencyLabels} from './api';
import {defaultSchedule, localInputToISO, toLocalInput} from './dates';

interface Props {
    task?: Task; // set when editing
    onSubmit: (input: TaskInput) => Promise<void>;
    onClose: () => void;
}

function initialValues(task?: Task) {
    if (!task) {
        return {title: '', urgency: 'medium' as Urgency, ...defaultSchedule()};
    }
    return {
        title: task.title,
        urgency: task.urgency,
        startsAt: toLocalInput(new Date(task.startsAt)),
        endsAt: toLocalInput(new Date(task.endsAt)),
    };
}

/** Client-side checks for quick feedback; the Go side is the source of truth. */
function validate(values: ReturnType<typeof initialValues>): string | null {
    if (!values.title.trim()) return 'Informe o nome da tarefa.';
    if (!values.startsAt || !values.endsAt) return 'Informe o início e o fim da tarefa.';
    if (new Date(values.endsAt) <= new Date(values.startsAt)) return 'O fim da tarefa deve ser depois do início.';
    return null;
}

export function TaskFormModal({task, onSubmit, onClose}: Props) {
    const [values, setValues] = useState(() => initialValues(task));
    const [error, setError] = useState<string | null>(null);
    const [saving, setSaving] = useState(false);

    const set = (field: keyof typeof values) => (e: {target: {value: string}}) =>
        setValues((v) => ({...v, [field]: e.target.value}));

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
            await onSubmit({
                title: values.title.trim(),
                urgency: values.urgency,
                startsAt: localInputToISO(values.startsAt),
                endsAt: localInputToISO(values.endsAt),
            });
        } catch (err) {
            setError(String(err));
            setSaving(false);
        }
    }

    const title = task ? 'Editar tarefa' : 'Nova tarefa';

    return (
        <dialog open aria-modal="true" aria-label={title}>
            <form onSubmit={handleSubmit}>
                <h2>{title}</h2>

                <p>
                    <label>
                        Nome <input value={values.title} onChange={set('title')} autoFocus/>
                    </label>
                </p>
                <p>
                    <label>
                        Início <input type="datetime-local" value={values.startsAt} onChange={set('startsAt')}/>
                    </label>
                </p>
                <p>
                    <label>
                        Fim <input type="datetime-local" value={values.endsAt} onChange={set('endsAt')}/>
                    </label>
                </p>
                <p>
                    <label>
                        Urgência{' '}
                        <select value={values.urgency} onChange={set('urgency')}>
                            {Object.entries(urgencyLabels).map(([value, label]) => (
                                <option key={value} value={value}>{label}</option>
                            ))}
                        </select>
                    </label>
                </p>

                {error && <p role="alert">{error}</p>}

                <button type="submit" disabled={saving}>Salvar</button>{' '}
                <button type="button" onClick={onClose}>Cancelar</button>
            </form>
        </dialog>
    );
}
