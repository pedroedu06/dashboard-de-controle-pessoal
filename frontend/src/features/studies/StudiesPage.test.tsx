import {act, fireEvent, render, screen, waitFor, within} from '@testing-library/react';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {StudiesPage} from './StudiesPage';
import {StudyOverview, studiesApi, StudySession} from './api';

vi.mock('./api', async (importOriginal) => ({
    ...(await importOriginal<typeof import('./api')>()),
    studiesApi: {
        overview: vi.fn(), createSubject: vi.fn(), renameSubject: vi.fn(), setSubjectArchived: vi.fn(),
        createTopic: vi.fn(), renameTopic: vi.fn(), setTopicStatus: vi.fn(), deleteTopic: vi.fn(),
        start: vi.fn(), pause: vi.fn(), resume: vi.fn(), stop: vi.fn(), deleteSession: vi.fn(),
        createGoal: vi.fn(), updateGoal: vi.fn(), deleteGoal: vi.fn(),
    },
}));

const api = vi.mocked(studiesApi);

const session = (overrides: Partial<StudySession> = {}): StudySession => ({
    id: 1, subjectId: 10, subjectName: 'Cálculo', topicId: 100, topicTitle: 'Limites',
    startedAt: new Date(2026, 9, 14, 9, 0).toISOString(), paused: false, elapsedSeconds: 65, ...overrides,
});

const base: StudyOverview = {
    subjects: [
        {
            id: 10, name: 'Cálculo', progress: 0.5,
            topics: [{id: 100, title: 'Limites', status: 'studying'}, {id: 101, title: 'Derivadas', status: 'not_seen'}],
        },
        {id: 11, name: 'Física', progress: 0, topics: []},
    ],
    archived: [{id: 12, name: 'Química', progress: 0, topics: []}],
    active: null,
    goals: [
        {id: 1, subjectId: 0, subjectName: '', targetMinutes: 240, doneMinutes: 130, completed: false},
        {id: 2, subjectId: 10, subjectName: 'Cálculo', targetMinutes: 60, doneMinutes: 75, completed: true},
    ],
    sessions: [session({id: 5, elapsedSeconds: 45 * 60})],
};

function withOverview(overview: Partial<StudyOverview> = {}) {
    api.overview.mockResolvedValue({...base, ...overview});
}

async function renderPage() {
    render(<StudiesPage/>);
    return screen.findByRole('heading', {name: 'Timer'});
}

describe('StudiesPage', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        withOverview();
        for (const fn of Object.values(api)) {
            if (fn !== api.overview && fn !== api.stop) fn.mockResolvedValue(undefined as never);
        }
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    describe('timer', () => {
        it('starts with the chosen subject and topic', async () => {
            await renderPage();

            fireEvent.change(screen.getByLabelText('Matéria'), {target: {value: '10'}});
            fireEvent.change(screen.getByLabelText('Tópico'), {target: {value: '101'}});
            fireEvent.click(screen.getByRole('button', {name: 'Iniciar'}));

            await waitFor(() => expect(api.start).toHaveBeenCalledWith(10, 101));
        });

        it('starts without a topic', async () => {
            await renderPage();

            expect(screen.getByRole('button', {name: 'Iniciar'})).toBeDisabled();
            fireEvent.change(screen.getByLabelText('Matéria'), {target: {value: '11'}});
            fireEvent.click(screen.getByRole('button', {name: 'Iniciar'}));

            await waitFor(() => expect(api.start).toHaveBeenCalledWith(11, 0));
        });

        it('ticks while running and pauses', async () => {
            vi.useFakeTimers({shouldAdvanceTime: true});
            withOverview({active: session()});
            await renderPage();

            expect(screen.getByRole('timer')).toHaveTextContent('00:01:05');
            await act(() => vi.advanceTimersByTimeAsync(3000));
            expect(screen.getByRole('timer')).toHaveTextContent('00:01:08');

            fireEvent.click(screen.getByRole('button', {name: 'Pausar'}));
            await waitFor(() => expect(api.pause).toHaveBeenCalled());
        });

        it('stands still while paused and resumes', async () => {
            vi.useFakeTimers({shouldAdvanceTime: true});
            withOverview({active: session({paused: true})});
            await renderPage();

            await act(() => vi.advanceTimersByTimeAsync(3000));
            expect(screen.getByRole('timer')).toHaveTextContent('00:01:05');
            expect(screen.getByText(/pausado/)).toBeInTheDocument();

            fireEvent.click(screen.getByRole('button', {name: 'Retomar'}));
            await waitFor(() => expect(api.resume).toHaveBeenCalled());
        });

        it('tells when a short session is discarded', async () => {
            withOverview({active: session()});
            api.stop.mockResolvedValue({saved: false, minutes: 0});
            await renderPage();

            fireEvent.click(screen.getByRole('button', {name: 'Encerrar'}));

            expect(await screen.findByRole('status')).toHaveTextContent('Sessão descartada');
        });

        it('tells how long a saved session was', async () => {
            withOverview({active: session()});
            api.stop.mockResolvedValue({saved: true, minutes: 95});
            await renderPage();

            fireEvent.click(screen.getByRole('button', {name: 'Encerrar'}));

            expect(await screen.findByRole('status')).toHaveTextContent('Sessão salva: 1h35');
        });
    });

    describe('goals', () => {
        it('shows progress per goal', async () => {
            await renderPage();

            expect(screen.getByRole('listitem', {name: 'Meta Geral'})).toHaveTextContent('2h10 de 4h');
            const calc = screen.getByRole('listitem', {name: 'Meta Cálculo'});
            expect(calc).toHaveTextContent('1h15 de 1h');
            expect(calc).toHaveTextContent('Concluída');
        });

        it('creates a goal in hours', async () => {
            await renderPage();

            fireEvent.change(screen.getByLabelText('Meta para'), {target: {value: '10'}});
            fireEvent.change(screen.getByLabelText('Horas por semana'), {target: {value: '2.5'}});
            fireEvent.click(screen.getByRole('button', {name: 'Adicionar meta'}));

            await waitFor(() => expect(api.createGoal).toHaveBeenCalledWith(10, 2.5));
        });

        it('edits a goal target', async () => {
            await renderPage();

            const row = screen.getByRole('listitem', {name: 'Meta Geral'});
            fireEvent.click(within(row).getByRole('button', {name: 'Editar'}));
            fireEvent.change(within(row).getByLabelText('Horas da meta Geral'), {target: {value: '6'}});
            fireEvent.click(within(row).getByRole('button', {name: 'Salvar'}));

            await waitFor(() => expect(api.updateGoal).toHaveBeenCalledWith(1, 6));
        });
    });

    describe('subjects', () => {
        it('shows progress and topic status', async () => {
            await renderPage();

            const calc = screen.getByRole('article', {name: 'Cálculo'});
            expect(within(calc).getByLabelText('Progresso de Cálculo')).toHaveValue(50);
            expect(calc).toHaveTextContent('50%');
            expect(within(calc).getByLabelText('Status de Limites')).toHaveValue('studying');
        });

        it('changes a topic status', async () => {
            await renderPage();

            fireEvent.change(screen.getByLabelText('Status de Derivadas'), {target: {value: 'reviewed'}});

            await waitFor(() => expect(api.setTopicStatus).toHaveBeenCalledWith(101, 'reviewed'));
        });

        it('creates subjects and topics', async () => {
            await renderPage();

            fireEvent.change(screen.getByLabelText('Nova matéria'), {target: {value: ' Química Orgânica '}});
            fireEvent.click(screen.getByRole('button', {name: 'Adicionar matéria'}));
            await waitFor(() => expect(api.createSubject).toHaveBeenCalledWith('Química Orgânica'));

            const fisica = screen.getByRole('article', {name: 'Física'});
            fireEvent.change(within(fisica).getByLabelText('Novo tópico em Física'), {target: {value: 'Óptica'}});
            fireEvent.click(within(fisica).getByRole('button', {name: 'Adicionar tópico'}));
            await waitFor(() => expect(api.createTopic).toHaveBeenCalledWith(11, 'Óptica'));
        });

        it('archives after confirmation and restores', async () => {
            const confirm = vi.spyOn(window, 'confirm').mockReturnValue(true);
            await renderPage();

            const fisica = screen.getByRole('article', {name: 'Física'});
            fireEvent.click(within(fisica).getByRole('button', {name: 'Arquivar'}));
            expect(confirm).toHaveBeenCalledWith(expect.stringContaining('meta geral'));
            await waitFor(() => expect(api.setSubjectArchived).toHaveBeenCalledWith(11, true));

            fireEvent.click(screen.getByRole('button', {name: 'Desarquivar'}));
            await waitFor(() => expect(api.setSubjectArchived).toHaveBeenCalledWith(12, false));
        });

        it('renames a topic through a prompt', async () => {
            vi.spyOn(window, 'prompt').mockReturnValue('Limites laterais');
            await renderPage();

            const calc = screen.getByRole('article', {name: 'Cálculo'});
            fireEvent.click(within(calc).getAllByRole('button', {name: 'Renomear'})[1]); // first topic

            await waitFor(() => expect(api.renameTopic).toHaveBeenCalledWith(100, 'Limites laterais'));
        });
    });

    it('lists and deletes sessions of the week', async () => {
        vi.spyOn(window, 'confirm').mockReturnValue(true);
        await renderPage();

        const list = screen.getByRole('region', {name: 'Sessões da semana'});
        expect(list).toHaveTextContent('Cálculo · Limites · 45min');
        fireEvent.click(within(list).getByRole('button', {name: 'Excluir'}));

        await waitFor(() => expect(api.deleteSession).toHaveBeenCalledWith(5));
    });

    it('shows backend errors', async () => {
        api.createSubject.mockRejectedValue('já existe uma matéria com esse nome');
        await renderPage();

        fireEvent.change(screen.getByLabelText('Nova matéria'), {target: {value: 'Cálculo'}});
        fireEvent.click(screen.getByRole('button', {name: 'Adicionar matéria'}));

        expect(await screen.findByRole('alert')).toHaveTextContent('já existe uma matéria com esse nome');
    });
});
