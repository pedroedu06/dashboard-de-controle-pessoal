import {fireEvent, render, screen, waitFor, within} from '@testing-library/react';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import {TasksPage} from './TasksPage';
import {Agenda, tasksApi} from './api';
import {toLocalInput} from './dates';
import {makeTask} from './fixtures';

vi.mock('./api', async (importOriginal) => ({
    ...(await importOriginal<typeof import('./api')>()),
    tasksApi: {
        agenda: vi.fn(),
        categories: vi.fn(),
        create: vi.fn(),
        update: vi.fn(),
        remove: vi.fn(),
        setDone: vi.fn(),
        setItemDone: vi.fn(),
    },
}));

const api = vi.mocked(tasksApi);

const today = toLocalInput(new Date()).slice(0, 10);

const shopping = makeTask({
    id: 4,
    title: 'Mercado',
    category: 'Casa',
    items: [{id: 40, title: 'arroz', done: true}, {id: 41, title: 'café', done: false}],
});
const gym = makeTask({
    id: 5,
    title: 'Academia',
    repeat: {kind: 'weekly', weekdays: [1, 3], until: ''},
    occursOn: today,
});

const agenda: Agenda = {
    overdue: [makeTask({id: 1, title: 'Pagar conta', urgency: 'high'})],
    days: [{
        date: today,
        tasks: [makeTask({id: 2, title: 'Estudar Go', category: 'Faculdade'}), makeTask({id: 3, title: 'Treinar', done: true}), shopping, gym],
    }],
};

async function renderPage() {
    render(<TasksPage/>);
    return screen.findByRole('region', {name: 'Hoje'});
}

describe('TasksPage', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        api.agenda.mockResolvedValue(agenda);
        api.categories.mockResolvedValue(['Casa', 'Faculdade']);
        api.setDone.mockResolvedValue();
        api.setItemDone.mockResolvedValue();
    });

    it('renders overdue and day groups with task details', async () => {
        const todayGroup = await renderPage();

        const overdue = screen.getByRole('region', {name: 'Atrasadas'});
        expect(within(overdue).getByText(/Urgência Alta/)).toBeInTheDocument();
        expect(within(todayGroup).getByText(/\[Faculdade\]/)).toBeInTheDocument();
        expect(within(todayGroup).getByText('Treinar').tagName).toBe('S');
        expect(within(todayGroup).getByText(/Repete: seg, qua/)).toBeInTheDocument();
        expect(within(todayGroup).getByText(/Itens 1\/2/)).toBeInTheDocument();
    });

    it('lists the checklist and disables manual completion', async () => {
        await renderPage();

        const list = screen.getByRole('list', {name: 'Itens de Mercado'});
        expect(within(list).getByRole('checkbox', {name: 'arroz'})).toBeChecked();
        expect(screen.getByRole('checkbox', {name: 'Concluir Mercado'})).toBeDisabled();
    });

    it('checks an item and reloads', async () => {
        await renderPage();

        fireEvent.click(screen.getByRole('checkbox', {name: 'café'}));

        await waitFor(() => expect(api.setItemDone).toHaveBeenCalledWith(shopping, shopping.items[1], true));
        await waitFor(() => expect(api.agenda).toHaveBeenCalledTimes(2));
    });

    it('completes a recurring occurrence with its task (and day)', async () => {
        await renderPage();

        fireEvent.click(screen.getByRole('checkbox', {name: 'Concluir Academia'}));

        await waitFor(() => expect(api.setDone).toHaveBeenCalledWith(gym, true));
    });

    it('filters by category and urgency', async () => {
        await renderPage();

        fireEvent.change(screen.getByLabelText('Filtrar categoria'), {target: {value: 'Casa'}});
        await waitFor(() => expect(api.agenda).toHaveBeenLastCalledWith({category: 'Casa', urgency: ''}));

        fireEvent.change(screen.getByLabelText('Filtrar urgência'), {target: {value: 'high'}});
        await waitFor(() => expect(api.agenda).toHaveBeenLastCalledWith({category: 'Casa', urgency: 'high'}));
    });

    it('deletes after confirmation, mentioning the whole series', async () => {
        const confirm = vi.spyOn(window, 'confirm').mockReturnValue(true);
        api.remove.mockResolvedValue();
        const todayGroup = await renderPage();

        const gymRow = within(todayGroup).getByText(/Academia/).closest('li')!;
        fireEvent.click(within(gymRow).getByRole('button', {name: 'Excluir'}));

        expect(confirm).toHaveBeenCalledWith(expect.stringContaining('todas as ocorrências'));
        await waitFor(() => expect(api.remove).toHaveBeenCalledWith(5));
    });

    it('does not delete when cancelled', async () => {
        vi.spyOn(window, 'confirm').mockReturnValue(false);
        await renderPage();

        const overdue = screen.getByRole('region', {name: 'Atrasadas'});
        fireEvent.click(within(overdue).getByRole('button', {name: 'Excluir'}));

        expect(api.remove).not.toHaveBeenCalled();
    });

    it('creates a task through the modal', async () => {
        api.create.mockResolvedValue(makeTask());
        await renderPage();

        fireEvent.click(screen.getByRole('button', {name: 'Nova tarefa'}));
        fireEvent.change(screen.getByLabelText('Nome'), {target: {value: 'Ler'}});
        fireEvent.click(screen.getByRole('button', {name: 'Salvar'}));

        await waitFor(() => expect(api.create).toHaveBeenCalledWith(expect.objectContaining({title: 'Ler', urgency: 'medium'})));
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });

    it('edits a task through the modal', async () => {
        api.update.mockResolvedValue(makeTask());
        await renderPage();

        const overdue = screen.getByRole('region', {name: 'Atrasadas'});
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
