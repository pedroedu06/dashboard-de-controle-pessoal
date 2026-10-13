import {useCallback, useEffect, useState} from 'react';
import {Agenda, Task, TaskInput, tasksApi} from './api';
import {AgendaView} from './AgendaView';
import {TaskFormModal} from './TaskFormModal';

type ModalState = {mode: 'closed'} | {mode: 'create'} | {mode: 'edit'; task: Task};

export function TasksPage() {
    const [agenda, setAgenda] = useState<Agenda | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [modal, setModal] = useState<ModalState>({mode: 'closed'});

    const reload = useCallback(async () => {
        try {
            setAgenda(await tasksApi.agenda());
            setError(null);
        } catch (err) {
            setError(String(err));
        }
    }, []);

    useEffect(() => {
        reload();
    }, [reload]);

    async function run(action: () => Promise<unknown>) {
        try {
            await action();
        } catch (err) {
            setError(String(err));
        }
        await reload();
    }

    async function handleSubmit(input: TaskInput) {
        if (modal.mode === 'edit') {
            await tasksApi.update(modal.task.id, input);
        } else {
            await tasksApi.create(input);
        }
        setModal({mode: 'closed'});
        await reload();
    }

    function handleDelete(task: Task) {
        if (window.confirm(`Excluir a tarefa "${task.title}"?`)) {
            run(() => tasksApi.remove(task.id));
        }
    }

    return (
        <div>
            <h1>Tarefas</h1>
            <button type="button" onClick={() => setModal({mode: 'create'})}>Nova tarefa</button>

            {error && <p role="alert">{error}</p>}
            {agenda === null && !error && <p>Carregando...</p>}
            {agenda && (
                <AgendaView
                    agenda={agenda}
                    onToggleDone={(task) => run(() => tasksApi.setDone(task.id, !task.done))}
                    onEdit={(task) => setModal({mode: 'edit', task})}
                    onDelete={handleDelete}
                />
            )}

            {modal.mode !== 'closed' && (
                <TaskFormModal
                    task={modal.mode === 'edit' ? modal.task : undefined}
                    onSubmit={handleSubmit}
                    onClose={() => setModal({mode: 'closed'})}
                />
            )}
        </div>
    );
}
