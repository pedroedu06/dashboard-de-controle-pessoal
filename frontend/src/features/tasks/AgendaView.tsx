import {Agenda, Task, urgencyLabels} from './api';
import {dayLabel, scheduleLabel} from './dates';

interface Props {
    agenda: Agenda;
    onToggleDone: (task: Task) => void;
    onEdit: (task: Task) => void;
    onDelete: (task: Task) => void;
}

export function AgendaView({agenda, ...actions}: Props) {
    if (agenda.overdue.length === 0 && agenda.days.length === 0) {
        return <p>Nenhuma tarefa por aqui.</p>;
    }

    return (
        <div>
            {agenda.overdue.length > 0 && (
                <TaskGroup title="Atrasadas" tasks={agenda.overdue} {...actions}/>
            )}
            {agenda.days.map((day) => (
                <TaskGroup key={day.date} title={dayLabel(day.date)} tasks={day.tasks} {...actions}/>
            ))}
        </div>
    );
}

type GroupProps = Omit<Props, 'agenda'> & {title: string; tasks: Task[]};

function TaskGroup({title, tasks, onToggleDone, onEdit, onDelete}: GroupProps) {
    return (
        <section aria-label={title}>
            <h3>{title}</h3>
            <ul>
                {tasks.map((task) => (
                    <li key={task.id}>
                        <label>
                            <input
                                type="checkbox"
                                checked={task.done}
                                onChange={() => onToggleDone(task)}
                                aria-label={`Concluir ${task.title}`}
                            />{' '}
                            {task.done ? <s>{task.title}</s> : task.title}
                        </label>
                        {' · '}{scheduleLabel(task.startsAt, task.endsAt)}
                        {' · '}Urgência {urgencyLabels[task.urgency]}{' '}
                        <button type="button" onClick={() => onEdit(task)}>Editar</button>{' '}
                        <button type="button" onClick={() => onDelete(task)}>Excluir</button>
                    </li>
                ))}
            </ul>
        </section>
    );
}
