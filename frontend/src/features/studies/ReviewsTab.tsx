import {useCallback, useEffect, useState} from 'react';
import {DueReview, formatDay, Grade, gradeLabels, studiesApi} from './api';

interface Props {
    onReviewed: () => void; // lets the page refresh its pending count
}

export function ReviewsTab({onReviewed}: Props) {
    const [due, setDue] = useState<DueReview[] | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [notice, setNotice] = useState<string | null>(null);

    const reload = useCallback(async () => {
        try {
            setDue(await studiesApi.dueReviews());
            setError(null);
        } catch (err) {
            setError(String(err));
        }
    }, []);

    useEffect(() => {
        reload();
    }, [reload]);

    async function review(item: DueReview, grade: Grade) {
        setNotice(null);
        try {
            const result = await studiesApi.reviewTopic(item.topicId, grade);
            const days = result.intervalDays === 1 ? 'amanhã' : `em ${result.intervalDays} dias`;
            setNotice(`${item.topicTitle}: próxima revisão ${days} (${formatDay(result.nextReviewOn)}).`);
        } catch (err) {
            setError(String(err));
        }
        await reload();
        onReviewed();
    }

    if (due === null) {
        return error ? <p role="alert">{error}</p> : <p>Carregando...</p>;
    }

    return (
        <section aria-label="Revisões de hoje">
            <h2>Revisões de hoje</h2>
            {error && <p role="alert">{error}</p>}
            {notice && <p role="status">{notice}</p>}
            {due.length === 0 && <p>Nenhuma revisão para hoje.</p>}
            <ul>
                {due.map((item) => (
                    <li key={item.topicId} aria-label={`Revisão de ${item.topicTitle}`}>
                        {item.subjectName} · <strong>{item.topicTitle}</strong>
                        {item.daysOverdue > 0 && <> · atrasada há {item.daysOverdue} {item.daysOverdue === 1 ? 'dia' : 'dias'}</>}
                        {' · '}Como foi?{' '}
                        {(Object.keys(gradeLabels) as Grade[]).map((grade) => (
                            <button key={grade} type="button" onClick={() => review(item, grade)}>
                                {gradeLabels[grade]}
                            </button>
                        ))}
                    </li>
                ))}
            </ul>
        </section>
    );
}
