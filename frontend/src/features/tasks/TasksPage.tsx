import {useCallback, useEffect, useState} from 'react';
import {Agenda, Filter, noFilter, Task, TaskInput, tasksApi} from './api';
import {AgendaView} from './AgendaView';
import {TaskFilters} from './TaskFilters';
import {TaskFormModal} from './TaskFormModal';

type ModalState = {mode: 'closed'} | {mode: 'create'} | {mode: 'edit'; task: Task};

export function TasksPage() {
    const [agenda, setAgenda] = useState<Agenda | null>(null);
    const [categories, setCategories] = useState<string[]>([]);
    const [filter, setFilter] = useState<Filter>(noFilter);
    const [error, setError] = useState<string | null>(null);
    const [modal, setModal] = useState<ModalState>({mode: 'closed'});

    const reload = useCallback(async () => {
        try {
            const [nextAgenda, nextCategories] = await Promise.all([tasksApi.agenda(filter), tasksApi.categories()]);
            setAgenda(nextAgenda);
            setCategories(nextCategories);
            setError(null);
        } catch (err) {
            setError(String(err));
        }
    }, [filter]);

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
        const what = task.repeat.kind !== 'none' ? `a tarefa recorrente "${task.title}" (todas as ocorrências)` : `a tarefa "${task.title}"`;
        if (window.confirm(`Excluir ${what}?`)) {
            run(() => tasksApi.remove(task.id));
        }
    }

    return (
        <div>
            <h1>Tarefas</h1>
            <button type="button" onClick={() => setModal({mode: 'create'})}>Nova tarefa</button>
            <TaskFilters categories={categories} value={filter} onChange={setFilter}/>

            {error && <p role="alert">{error}</p>}
            {agenda === null && !error && <p>Carregando...</p>}
            {agenda && (
                <AgendaView
                    agenda={agenda}
                    onToggleDone={(task) => run(() => tasksApi.setDone(task, !task.done))}
                    onToggleItem={(task, item) => run(() => tasksApi.setItemDone(task, item, !item.done))}
                    onEdit={(task) => setModal({mode: 'edit', task})}
                    onDelete={handleDelete}
                />
            )}

            {modal.mode !== 'closed' && (
                <TaskFormModal
                    task={modal.mode === 'edit' ? modal.task : undefined}
                    categories={categories}
                    onSubmit={handleSubmit}
                    onClose={() => setModal({mode: 'closed'})}
                />
            )}
        </div>
    );
}
