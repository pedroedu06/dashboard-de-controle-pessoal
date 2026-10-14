import {FormEvent, useState} from 'react';
import {formatMinutes, Goal, Subject} from './api';

interface Props {
    goals: Goal[];
    subjects: Subject[];
    onCreate: (subjectId: number, hours: number) => void;
    onUpdate: (id: number, hours: number) => void;
    onDelete: (goal: Goal) => void;
}

export function GoalsPanel({goals, subjects, onCreate, onUpdate, onDelete}: Props) {
    const [subjectId, setSubjectId] = useState(0);
    const [hours, setHours] = useState('');

    function handleSubmit(e: FormEvent) {
        e.preventDefault();
        const value = Number(hours);
        if (!(value > 0)) return;
        onCreate(subjectId, value);
        setHours('');
    }

    return (
        <section aria-label="Metas da semana">
            <h2>Metas da semana</h2>
            {goals.length === 0 && <p>Nenhuma meta ainda.</p>}
            <ul>
                {goals.map((goal) => (
                    <GoalRow key={goal.id} goal={goal} onUpdate={onUpdate} onDelete={onDelete}/>
                ))}
            </ul>

            <form onSubmit={handleSubmit}>
                <label>
                    Meta para{' '}
                    <select value={subjectId} onChange={(e) => setSubjectId(Number(e.target.value))}>
                        <option value={0}>Geral (todas as matérias)</option>
                        {subjects.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                    </select>
                </label>{' '}
                <label>
                    Horas por semana{' '}
                    <input type="number" min="0.5" step="0.5" value={hours} onChange={(e) => setHours(e.target.value)}/>
                </label>{' '}
                <button type="submit">Adicionar meta</button>
            </form>
        </section>
    );
}

function GoalRow({goal, onUpdate, onDelete}: {goal: Goal} & Pick<Props, 'onUpdate' | 'onDelete'>) {
    const [editing, setEditing] = useState(false);
    const [hours, setHours] = useState(String(goal.targetMinutes / 60));
    const name = goal.subjectId === 0 ? 'Geral' : goal.subjectName;

    function save() {
        const value = Number(hours);
        if (value > 0) onUpdate(goal.id, value);
        setEditing(false);
    }

    return (
        <li aria-label={`Meta ${name}`}>
            {name}{' '}
            <progress max={goal.targetMinutes} value={Math.min(goal.doneMinutes, goal.targetMinutes)}/>{' '}
            {formatMinutes(goal.doneMinutes)} de {formatMinutes(goal.targetMinutes)}
            {goal.completed && <> · Concluída ✓</>}{' '}
            {editing ? (
                <>
                    <input
                        type="number" min="0.5" step="0.5" aria-label={`Horas da meta ${name}`}
                        value={hours} onChange={(e) => setHours(e.target.value)}
                    />{' '}
                    <button type="button" onClick={save}>Salvar</button>
                </>
            ) : (
                <button type="button" onClick={() => setEditing(true)}>Editar</button>
            )}{' '}
            <button type="button" onClick={() => onDelete(goal)}>Excluir</button>
        </li>
    );
}
