import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import {describe, expect, it, vi} from 'vitest';
import {TaskFormModal} from './TaskFormModal';
import {Task} from './api';
import {makeTask} from './fixtures';

function setup(task?: Task) {
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    const onClose = vi.fn();
    render(<TaskFormModal task={task} categories={['Faculdade', 'Trabalho']} onSubmit={onSubmit} onClose={onClose}/>);
    return {onSubmit, onClose};
}

const field = (label: string) => screen.getByLabelText(label);
const change = (label: string, value: string) => fireEvent.change(field(label), {target: {value}});
const click = (name: string) => fireEvent.click(screen.getByRole('button', {name}));
const save = () => click('Salvar');

/** Fills a valid one-hour task on Oct 13 2026, 14:00. */
function fillBasics(title = 'Estudar') {
    change('Nome', title);
    change('Início', '2026-10-13T14:00');
    change('Fim', '2026-10-13T15:00');
}

describe('TaskFormModal', () => {
    it('requires a name', () => {
        const {onSubmit} = setup();

        save();

        expect(screen.getByRole('alert')).toHaveTextContent('Informe o nome da tarefa.');
        expect(onSubmit).not.toHaveBeenCalled();
    });

    it('requires the end after the start', () => {
        const {onSubmit} = setup();
        fillBasics();
        change('Fim', '2026-10-13T14:00');

        save();

        expect(screen.getByRole('alert')).toHaveTextContent('O fim da tarefa deve ser depois do início.');
        expect(onSubmit).not.toHaveBeenCalled();
    });

    it('submits a basic task with defaults', async () => {
        const {onSubmit} = setup();
        fillBasics('  Estudar Go  ');
        change('Urgência', 'high');

        save();

        await waitFor(() => expect(onSubmit).toHaveBeenCalledWith({
            title: 'Estudar Go',
            category: '',
            urgency: 'high',
            startsAt: new Date('2026-10-13T14:00').toISOString(),
            endsAt: new Date('2026-10-13T15:00').toISOString(),
            repeat: {kind: 'none', weekdays: [], until: ''},
            items: [],
        }));
    });

    it('suggests existing categories and submits the typed one', async () => {
        const {onSubmit} = setup();
        fillBasics();
        change('Categoria', ' Faculdade ');

        expect(field('Categoria')).toHaveAttribute('list', 'task-categories');
        expect(document.querySelectorAll('#task-categories option')).toHaveLength(2);

        save();

        await waitFor(() => expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({category: 'Faculdade'})));
    });

    it('builds a checklist: add, reorder and remove items', async () => {
        const {onSubmit} = setup();
        fillBasics('Mercado');
        fireEvent.click(screen.getByLabelText('Tem lista de itens'));

        for (const item of ['arroz', 'feijão', 'café']) {
            change('Novo item', item);
            click('Adicionar');
        }
        click('Subir item 3'); // café above feijão
        click('Remover item 1'); // drop arroz
        change('Novo item', 'leite');
        fireEvent.keyDown(field('Novo item'), {key: 'Enter'});

        save();

        await waitFor(() => expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({
            items: [{id: 0, title: 'café'}, {id: 0, title: 'feijão'}, {id: 0, title: 'leite'}],
        })));
    });

    it('rejects blank items', () => {
        const {onSubmit} = setup();
        fillBasics();
        fireEvent.click(screen.getByLabelText('Tem lista de itens'));
        change('Novo item', 'arroz');
        click('Adicionar');
        change('Item 1', '   ');

        save();

        expect(screen.getByRole('alert')).toHaveTextContent('Todo item da lista precisa de um nome.');
        expect(onSubmit).not.toHaveBeenCalled();
    });

    it('drops items when the list is unchecked', async () => {
        const {onSubmit} = setup();
        fillBasics();
        fireEvent.click(screen.getByLabelText('Tem lista de itens'));
        change('Novo item', 'arroz');
        click('Adicionar');
        fireEvent.click(screen.getByLabelText('Tem lista de itens'));

        save();

        await waitFor(() => expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({items: []})));
    });

    it('submits a weekly repeat with sorted weekdays and until', async () => {
        const {onSubmit} = setup();
        fillBasics('Academia');
        fireEvent.click(screen.getByLabelText('Repetir'));
        change('Frequência', 'weekly');
        fireEvent.click(screen.getByLabelText('Sex'));
        fireEvent.click(screen.getByLabelText('Seg'));
        change('Repetir até (opcional)', '2026-12-31');

        save();

        await waitFor(() => expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({
            repeat: {kind: 'weekly', weekdays: [1, 5], until: '2026-12-31'},
        })));
    });

    it('validates repeat rules', () => {
        const {onSubmit} = setup();
        fillBasics();
        fireEvent.click(screen.getByLabelText('Repetir'));
        change('Frequência', 'weekly');

        save();
        expect(screen.getByRole('alert')).toHaveTextContent('Escolha pelo menos um dia da semana.');

        change('Frequência', 'daily');
        change('Fim', '2026-10-14T09:00');
        save();
        expect(screen.getByRole('alert')).toHaveTextContent('começar e terminar no mesmo dia');

        expect(onSubmit).not.toHaveBeenCalled();
    });

    it('shows the error returned by Go', async () => {
        const {onSubmit} = setup();
        onSubmit.mockRejectedValue('tarefa não encontrada');
        fillBasics();

        save();

        expect(await screen.findByRole('alert')).toHaveTextContent('tarefa não encontrada');
    });

    it('prefills an occurrence with the series schedule, items and repeat', async () => {
        const {onSubmit} = setup(makeTask({
            id: 1,
            title: 'Academia',
            category: 'Saúde',
            urgency: 'low',
            startsAt: new Date(2026, 9, 16, 7, 0).toISOString(),
            endsAt: new Date(2026, 9, 16, 8, 0).toISOString(),
            seriesStartsAt: new Date(2026, 9, 13, 7, 0).toISOString(),
            seriesEndsAt: new Date(2026, 9, 13, 8, 0).toISOString(),
            repeat: {kind: 'weekly', weekdays: [2, 5], until: ''},
            items: [{id: 9, title: 'alongar', done: true}],
            occursOn: '2026-10-16',
        }));

        expect(screen.getByRole('heading', {name: 'Editar tarefa'})).toBeInTheDocument();
        expect(field('Nome')).toHaveValue('Academia');
        expect(field('Categoria')).toHaveValue('Saúde');
        expect(field('Início')).toHaveValue('2026-10-13T07:00');
        expect(field('Fim')).toHaveValue('2026-10-13T08:00');
        expect(field('Item 1')).toHaveValue('alongar');
        expect(field('Ter')).toBeChecked();
        expect(field('Sex')).toBeChecked();

        save();

        await waitFor(() => expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({
            items: [{id: 9, title: 'alongar'}],
            repeat: {kind: 'weekly', weekdays: [2, 5], until: ''},
        })));
    });

    it('closes on cancel', () => {
        const {onClose} = setup();

        click('Cancelar');

        expect(onClose).toHaveBeenCalled();
    });
});
