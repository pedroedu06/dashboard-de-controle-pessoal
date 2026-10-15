import {fireEvent, render, screen, waitFor, within} from '@testing-library/react';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import {ReviewsTab} from './ReviewsTab';
import {DueReview, studiesApi} from './api';

vi.mock('./api', async (importOriginal) => ({
    ...(await importOriginal<typeof import('./api')>()),
    studiesApi: {dueReviews: vi.fn(), reviewTopic: vi.fn()},
}));

const api = vi.mocked(studiesApi);

const review = (overrides: Partial<DueReview>): DueReview => ({
    topicId: 1, topicTitle: 'Crase', subjectName: 'Português', dueOn: '2026-10-12',
    daysOverdue: 3, repetitions: 1, intervalDays: 1, ...overrides,
});

describe('ReviewsTab', () => {
    beforeEach(() => {
        vi.resetAllMocks();
    });

    it('lists due reviews with how late they are', async () => {
        api.dueReviews.mockResolvedValue([review({}), review({topicId: 2, topicTitle: 'Regência', daysOverdue: 0})]);

        render(<ReviewsTab onReviewed={vi.fn()}/>);

        expect(await screen.findByRole('listitem', {name: 'Revisão de Crase'})).toHaveTextContent('Português · Crase · atrasada há 3 dias');
        expect(screen.getByRole('listitem', {name: 'Revisão de Regência'})).not.toHaveTextContent('atrasada');
    });

    it('grades a review and tells when the next one is', async () => {
        const onReviewed = vi.fn();
        api.dueReviews.mockResolvedValueOnce([review({})]).mockResolvedValueOnce([]);
        api.reviewTopic.mockResolvedValue({nextReviewOn: '2026-10-21', intervalDays: 6});
        render(<ReviewsTab onReviewed={onReviewed}/>);

        const item = await screen.findByRole('listitem', {name: 'Revisão de Crase'});
        fireEvent.click(within(item).getByRole('button', {name: 'Fácil'}));

        await waitFor(() => expect(api.reviewTopic).toHaveBeenCalledWith(1, 'easy'));
        expect(await screen.findByRole('status')).toHaveTextContent('Crase: próxima revisão em 6 dias (21/10).');
        expect(screen.getByText('Nenhuma revisão para hoje.')).toBeInTheDocument();
        expect(onReviewed).toHaveBeenCalled();
    });

    it('offers the four grades', async () => {
        api.dueReviews.mockResolvedValue([review({})]);
        render(<ReviewsTab onReviewed={vi.fn()}/>);

        const item = await screen.findByRole('listitem', {name: 'Revisão de Crase'});
        const buttons = within(item).getAllByRole('button').map((b) => b.textContent);
        expect(buttons).toEqual(['Esqueci', 'Difícil', 'OK', 'Fácil']);
    });

    it('says when tomorrow', async () => {
        api.dueReviews.mockResolvedValue([review({})]);
        api.reviewTopic.mockResolvedValue({nextReviewOn: '2026-10-16', intervalDays: 1});
        render(<ReviewsTab onReviewed={vi.fn()}/>);

        fireEvent.click(await screen.findByRole('button', {name: 'Esqueci'}));

        expect(await screen.findByRole('status')).toHaveTextContent('próxima revisão amanhã');
    });
});
