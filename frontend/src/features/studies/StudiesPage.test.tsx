import {fireEvent, render, screen} from '@testing-library/react';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import {StudiesPage} from './StudiesPage';
import {studiesApi} from './api';

vi.mock('./api', async (importOriginal) => ({
    ...(await importOriginal<typeof import('./api')>()),
    studiesApi: {overview: vi.fn(), dueReviews: vi.fn(), questionReport: vi.fn()},
}));

const api = vi.mocked(studiesApi);

describe('StudiesPage', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        api.overview.mockResolvedValue({subjects: [], archived: [], active: null, goals: [], sessions: []});
        api.questionReport.mockResolvedValue({bySubject: [], byTopic: [], weaknesses: [], logs: []});
        api.dueReviews.mockResolvedValue([
            {topicId: 1, topicTitle: 'Crase', subjectName: 'Português', dueOn: '2026-10-15', daysOverdue: 0, repetitions: 1, intervalDays: 1},
        ]);
    });

    it('starts on the overview and shows the pending review count', async () => {
        render(<StudiesPage/>);

        expect(await screen.findByRole('button', {name: 'Revisões de hoje (1)'})).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Visão geral'})).toHaveAttribute('aria-current', 'page');
        expect(await screen.findByRole('heading', {name: 'Timer'})).toBeInTheDocument();
    });

    it('switches between tabs', async () => {
        render(<StudiesPage/>);

        fireEvent.click(await screen.findByRole('button', {name: 'Revisões de hoje (1)'}));
        expect(await screen.findByRole('heading', {name: 'Revisões de hoje'})).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', {name: 'Questões'}));
        expect(await screen.findByRole('heading', {name: 'Registrar questões'})).toBeInTheDocument();
    });
});
