import {fireEvent, render, screen} from '@testing-library/react';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import App from './App';
import {DBStatus} from '../wailsjs/go/main/App';
import {tasksApi} from './features/tasks/api';

// Wails bindings only exist inside the running app, so tests mock them.
vi.mock('../wailsjs/go/main/App', () => ({DBStatus: vi.fn()}));
vi.mock('./features/tasks/api', async (importOriginal) => ({
    ...(await importOriginal<typeof import('./features/tasks/api')>()),
    tasksApi: {agenda: vi.fn()},
}));

describe('App', () => {
    beforeEach(() => {
        vi.mocked(DBStatus).mockResolvedValue({connected: true});
        vi.mocked(tasksApi.agenda).mockResolvedValue({overdue: [], days: []});
    });

    it('starts on Home', async () => {
        render(<App/>);

        expect(await screen.findByText('Banco: conectado')).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Home'})).toHaveAttribute('aria-current', 'page');
    });

    it('navigates to Tarefas', async () => {
        render(<App/>);

        fireEvent.click(screen.getByRole('button', {name: 'Tarefas'}));

        expect(await screen.findByText('Nenhuma tarefa por aqui.')).toBeInTheDocument();
        expect(screen.getByRole('heading', {name: 'Tarefas'})).toBeInTheDocument();
    });
});
