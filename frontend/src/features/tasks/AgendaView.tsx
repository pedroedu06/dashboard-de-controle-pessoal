import {Agenda, Item, repeatLabel, Task, urgencyLabels} from './api';
import {dayLabel, scheduleLabel} from './dates';

interface Props {
    agenda: Agenda;
    onToggleDone: (task: Task) => void;
    onToggleItem: (task: Task, item: Item) => void;
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

function TaskGroup({title, tasks, ...actions}: GroupProps) {
    return (
        <section aria-label={title}>
            <h3>{title}</h3>
            <ul>
                {tasks.map((task) => (
                    <TaskRow key={`${task.id}-${task.occursOn}`} task={task} {...actions}/>
                ))}
            </ul>
        </section>
    );
}

type RowProps = Omit<Props, 'agenda'> & {task: Task};

function TaskRow({task, onToggleDone, onToggleItem, onEdit, onDelete}: RowProps) {
    const hasItems = task.items.length > 0;
    const doneItems = task.items.filter((it) => it.done).length;

    return (
        <li>
            <label>
                <input
                    type="checkbox"
                    checked={task.done}
                    disabled={hasItems}
                    title={hasItems ? 'Conclua marcando os itens da lista' : undefined}
                    onChange={() => onToggleDone(task)}
                    aria-label={`Concluir ${task.title}`}
                />{' '}
                {task.done ? <s>{task.title}</s> : task.title}
            </label>
            {task.category && <> · [{task.category}]</>}
            {' · '}{scheduleLabel(task.startsAt, task.endsAt)}
            {' · '}Urgência {urgencyLabels[task.urgency]}
            {task.repeat.kind !== 'none' && <> · {repeatLabel(task.repeat)}</>}
            {hasItems && <> · Itens {doneItems}/{task.items.length}</>}{' '}
            <button type="button" onClick={() => onEdit(task)}>Editar</button>{' '}
            <button type="button" onClick={() => onDelete(task)}>Excluir</button>

            {hasItems && (
                <ul aria-label={`Itens de ${task.title}`}>
                    {task.items.map((item) => (
                        <li key={item.id}>
                            <label>
                                <input type="checkbox" checked={item.done} onChange={() => onToggleItem(task, item)}/>
                                {' '}{item.done ? <s>{item.title}</s> : item.title}
                            </label>
                        </li>
                    ))}
                </ul>
            )}
        </li>
    );
}
