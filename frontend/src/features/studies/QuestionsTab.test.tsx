import {fireEvent, render, screen, waitFor, within} from '@testing-library/react';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import {QuestionsTab} from './QuestionsTab';
import {QuestionReport, studiesApi, StudyOverview} from './api';

vi.mock('./api', async (importOriginal) => ({
    ...(await importOriginal<typeof import('./api')>()),
    studiesApi: {overview: vi.fn(), questionReport: vi.fn(), logQuestions: vi.fn(), deleteQuestionLog: vi.fn()},
}));

const api = vi.mocked(studiesApi);

const overview: StudyOverview = {
    subjects: [{id: 1, name: 'Português', progress: 0, topics: [{id: 10, title: 'Crase', status: 'studying', nextReviewOn: ''}]}],
    archived: [], active: null, goals: [], sessions: [],
};

const report: QuestionReport = {
    bySubject: [{id: 1, name: 'Português', subjectName: '', total: 40, correct: 28, rate: 0.7}],
    byTopic: [{id: 10, name: 'Crase', subjectName: 'Português', total: 20, correct: 6, rate: 0.3}],
    weaknesses: [{id: 10, name: 'Crase', subjectName: 'Português', total: 20, correct: 6, rate: 0.3}],
    logs: [{id: 5, subjectName: 'Português', topicTitle: 'Crase', doneOn: '2026-10-14', total: 20, correct: 6, source: 'FGV'}],
};

async function renderTab() {
    render(<QuestionsTab/>);
    return screen.findByRole('region', {name: 'Por matéria'});
}

const change = (label: string, value: string) => fireEvent.change(screen.getByLabelText(label), {target: {value}});

describe('QuestionsTab', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        api.overview.mockResolvedValue(overview);
        api.questionReport.mockResolvedValue(report);
        api.logQuestions.mockResolvedValue();
        api.deleteQuestionLog.mockResolvedValue();
    });

    it('shows accuracy per subject and topic, and weaknesses', async () => {
        const bySubject = await renderTab();

        expect(within(bySubject).getByRole('row', {name: /Português/})).toHaveTextContent('Português402870%');
        expect(screen.getByRole('region', {name: 'Por tópico'})).toHaveTextContent('Crase');
        expect(screen.getByRole('region', {name: 'Pontos fracos'})).toHaveTextContent('Português · Crase: 30% (6/20)');
        expect(screen.getByRole('region', {name: 'Registros'})).toHaveTextContent('14/10 · Português · Crase · 6/20 (30%) · FGV');
        expect(api.questionReport).toHaveBeenCalledWith(30);
    });

    it('changes the period', async () => {
        await renderTab();

        change('Período', '90');

        await waitFor(() => expect(api.questionReport).toHaveBeenLastCalledWith(90));
    });

    it('logs questions', async () => {
        await renderTab();

        change('Matéria', '1');
        change('Tópico', '10');
        change('Data', '2026-10-13');
        change('Feitas', '15');
        change('Acertos', '12');
        change('Origem (opcional)', ' Simulado 2 ');
        fireEvent.click(screen.getByRole('button', {name: 'Registrar'}));

        await waitFor(() => expect(api.logQuestions).toHaveBeenCalledWith({
            subjectId: 1, topicId: 10, doneOn: '2026-10-13', total: 15, correct: 12, source: 'Simulado 2',
        }));
    });

    it('blocks more correct answers than questions', async () => {
        await renderTab();

        change('Matéria', '1');
        change('Feitas', '5');
        change('Acertos', '6');
        fireEvent.click(screen.getByRole('button', {name: 'Registrar'}));

        expect(screen.getByRole('alert')).toHaveTextContent('Os acertos devem estar entre 0 e o número de questões.');
        expect(api.logQuestions).not.toHaveBeenCalled();
    });

    it('explains how to unlock weaknesses', async () => {
        api.questionReport.mockResolvedValue({...report, weaknesses: []});
        await renderTab();

        expect(screen.getByRole('region', {name: 'Pontos fracos'})).toHaveTextContent('pelo menos 10 questões');
    });

    it('deletes a log after confirmation', async () => {
        vi.spyOn(window, 'confirm').mockReturnValue(true);
        await renderTab();

        fireEvent.click(within(screen.getByRole('region', {name: 'Registros'})).getByRole('button', {name: 'Excluir'}));

        await waitFor(() => expect(api.deleteQuestionLog).toHaveBeenCalledWith(5));
    });
});
