import {fireEvent, render, screen, waitFor, within} from '@testing-library/react';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import {TasksPage} from './TasksPage';
import {Agenda, Task, tasksApi} from './api';
import {toLocalInput} from './dates';

vi.mock('./api', async (importOriginal) => ({
    ...(await importOriginal<typeof import('./api')>()),
    tasksApi: {agenda: vi.fn(), create: vi.fn(), update: vi.fn(), remove: vi.fn(), setDone: vi.fn()},
}));

const api = vi.mocked(tasksApi);

const today = toLocalInput(new Date()).slice(0, 10);

function makeTask(overrides: Partial<Task>): Task {
    return {
        id: 1,
        title: 'Tarefa',
        urgency: 'medium',
        startsAt: new Date().toISOString(),
        endsAt: new Date(Date.now() + 3_600_000).toISOString(),
        done: false,
        ...overrides,
    };
}

const agenda: Agenda = {
    overdue: [makeTask({id: 1, title: 'Pagar conta', urgency: 'high'})],
    days: [{date: today, tasks: [makeTask({id: 2, title: 'Estudar Go'}), makeTask({id: 3, title: 'Treinar', done: true})]}],
};

describe('TasksPage', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        api.agenda.mockResolvedValue(agenda);
    });

    it('renders overdue and day groups', async () => {
        render(<TasksPage/>);

        const overdue = await screen.findByRole('region', {name: 'Atrasadas'});
        expect(within(overdue).getByText('Pagar conta')).toBeInTheDocument();
        expect(within(overdue).getByText(/Urgência Alta/)).toBeInTheDocument();

        const todayGroup = screen.getByRole('region', {name: 'Hoje'});
        expect(within(todayGroup).getByText('Estudar Go')).toBeInTheDocument();
        expect(within(todayGroup).getByText('Treinar').tagName).toBe('S');
    });

    it('toggles done and reloads', async () => {
        api.setDone.mockResolvedValue(makeTask({}));
        render(<TasksPage/>);

        fireEvent.click(await screen.findByRole('checkbox', {name: 'Concluir Treinar'}));

        await waitFor(() => expect(api.setDone).toHaveBeenCalledWith(3, false));
        await waitFor(() => expect(api.agenda).toHaveBeenCalledTimes(2));
    });

    it('deletes after confirmation', async () => {
        vi.spyOn(window, 'confirm').mockReturnValue(true);
        api.remove.mockResolvedValue();
        render(<TasksPage/>);

        const overdue = await screen.findByRole('region', {name: 'Atrasadas'});
        fireEvent.click(within(overdue).getByRole('button', {name: 'Excluir'}));

        await waitFor(() => expect(api.remove).toHaveBeenCalledWith(1));
    });

    it('does not delete when cancelled', async () => {
        vi.spyOn(window, 'confirm').mockReturnValue(false);
        render(<TasksPage/>);

        const overdue = await screen.findByRole('region', {name: 'Atrasadas'});
        fireEvent.click(within(overdue).getByRole('button', {name: 'Excluir'}));

        expect(api.remove).not.toHaveBeenCalled();
    });

    it('creates a task through the modal', async () => {
        api.create.mockResolvedValue(makeTask({}));
        render(<TasksPage/>);
        await screen.findByRole('region', {name: 'Hoje'});

        fireEvent.click(screen.getByRole('button', {name: 'Nova tarefa'}));
        fireEvent.change(screen.getByLabelText('Nome'), {target: {value: 'Ler'}});
        fireEvent.click(screen.getByRole('button', {name: 'Salvar'}));

        await waitFor(() => expect(api.create).toHaveBeenCalledWith(expect.objectContaining({title: 'Ler', urgency: 'medium'})));
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });

    it('edits a task through the modal', async () => {
        api.update.mockResolvedValue(makeTask({}));
        render(<TasksPage/>);

        const overdue = await screen.findByRole('region', {name: 'Atrasadas'});
        fireEvent.click(within(overdue).getByRole('button', {name: 'Editar'}));
        expect(screen.getByLabelText('Nome')).toHaveValue('Pagar conta');
        fireEvent.change(screen.getByLabelText('Nome'), {target: {value: 'Pagar luz'}});
        fireEvent.click(screen.getByRole('button', {name: 'Salvar'}));

        await waitFor(() => expect(api.update).toHaveBeenCalledWith(1, expect.objectContaining({title: 'Pagar luz', urgency: 'high'})));
    });

    it('shows the error when loading fails', async () => {
        api.agenda.mockRejectedValue('banco não conectado');

        render(<TasksPage/>);

        expect(await screen.findByRole('alert')).toHaveTextContent('banco não conectado');
    });
});
