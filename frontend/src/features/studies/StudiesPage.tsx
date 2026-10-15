import {useCallback, useEffect, useState} from 'react';
import {studiesApi} from './api';
import {OverviewTab} from './OverviewTab';
import {QuestionsTab} from './QuestionsTab';
import {ReviewsTab} from './ReviewsTab';

type Tab = 'overview' | 'reviews' | 'questions';

export function StudiesPage() {
    const [tab, setTab] = useState<Tab>('overview');
    const [dueCount, setDueCount] = useState<number | null>(null);

    const refreshCount = useCallback(async () => {
        try {
            setDueCount((await studiesApi.dueReviews()).length);
        } catch {
            setDueCount(null); // the tabs show the error themselves
        }
    }, []);

    // Status changes in the overview can add or remove reviews, so refresh on every tab switch.
    useEffect(() => {
        refreshCount();
    }, [tab, refreshCount]);

    const tabs: {key: Tab; label: string}[] = [
        {key: 'overview', label: 'Visão geral'},
        {key: 'reviews', label: dueCount ? `Revisões de hoje (${dueCount})` : 'Revisões de hoje'},
        {key: 'questions', label: 'Questões'},
    ];

    return (
        <div>
            <h1>Estudos</h1>
            <nav aria-label="Seções de estudos">
                {tabs.map((t) => (
                    <button
                        key={t.key}
                        type="button"
                        aria-current={t.key === tab ? 'page' : undefined}
                        onClick={() => setTab(t.key)}
                    >
                        {t.label}
                    </button>
                ))}
            </nav>

            {tab === 'overview' && <OverviewTab/>}
            {tab === 'reviews' && <ReviewsTab onReviewed={refreshCount}/>}
            {tab === 'questions' && <QuestionsTab/>}
        </div>
    );
}
