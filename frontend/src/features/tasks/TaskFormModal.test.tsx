import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import {describe, expect, it, vi} from 'vitest';
import {TaskFormModal} from './TaskFormModal';
import {Task} from './api';

function setup(task?: Task) {
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    const onClose = vi.fn();
    render(<TaskFormModal task={task} onSubmit={onSubmit} onClose={onClose}/>);
    return {onSubmit, onClose};
}

const field = (label: string) => screen.getByLabelText(label);
const change = (label: string, value: string) => fireEvent.change(field(label), {target: {value}});
const save = () => fireEvent.click(screen.getByRole('button', {name: 'Salvar'}));

describe('TaskFormModal', () => {
    it('requires a name', () => {
        const {onSubmit} = setup();

        save();

        expect(screen.getByRole('alert')).toHaveTextContent('Informe o nome da tarefa.');
        expect(onSubmit).not.toHaveBeenCalled();
    });

    it('requires the end after the start', () => {
        const {onSubmit} = setup();
        change('Nome', 'Estudar');
        change('Início', '2026-10-13T15:00');
        change('Fim', '2026-10-13T15:00');

        save();

        expect(screen.getByRole('alert')).toHaveTextContent('O fim da tarefa deve ser depois do início.');
        expect(onSubmit).not.toHaveBeenCalled();
    });

    it('submits a trimmed name, the urgency and ISO dates', async () => {
        const {onSubmit} = setup();
        change('Nome', '  Estudar Go  ');
        change('Início', '2026-10-13T14:00');
        change('Fim', '2026-10-13T16:00');
        change('Urgência', 'high');

        save();

        await waitFor(() => expect(onSubmit).toHaveBeenCalledWith({
            title: 'Estudar Go',
            urgency: 'high',
            startsAt: new Date('2026-10-13T14:00').toISOString(),
            endsAt: new Date('2026-10-13T16:00').toISOString(),
        }));
    });

    it('defaults to medium urgency', () => {
        setup();

        expect(field('Urgência')).toHaveValue('medium');
    });

    it('shows the error returned by Go', async () => {
        const {onSubmit} = setup();
        onSubmit.mockRejectedValue('tarefa não encontrada');
        change('Nome', 'Estudar');

        save();

        expect(await screen.findByRole('alert')).toHaveTextContent('tarefa não encontrada');
    });

    it('prefills the fields when editing', () => {
        setup({
            id: 1,
            title: 'Revisar',
            urgency: 'low',
            startsAt: new Date(2026, 9, 13, 9, 0).toISOString(),
            endsAt: new Date(2026, 9, 13, 10, 0).toISOString(),
            done: false,
        });

        expect(screen.getByRole('heading', {name: 'Editar tarefa'})).toBeInTheDocument();
        expect(field('Nome')).toHaveValue('Revisar');
        expect(field('Urgência')).toHaveValue('low');
        expect(field('Início')).toHaveValue('2026-10-13T09:00');
        expect(field('Fim')).toHaveValue('2026-10-13T10:00');
    });

    it('closes on cancel', () => {
        const {onClose} = setup();

        fireEvent.click(screen.getByRole('button', {name: 'Cancelar'}));

        expect(onClose).toHaveBeenCalled();
    });
});
