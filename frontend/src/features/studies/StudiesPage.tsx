import {useCallback, useEffect, useState} from 'react';
import {formatMinutes, studiesApi, StudyOverview} from './api';
import {GoalsPanel} from './GoalsPanel';
import {SessionsList} from './SessionsList';
import {SubjectsPanel} from './SubjectsPanel';
import {TimerPanel} from './TimerPanel';

export function StudiesPage() {
    const [overview, setOverview] = useState<StudyOverview | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [notice, setNotice] = useState<string | null>(null);

    const reload = useCallback(async () => {
        try {
            setOverview(await studiesApi.overview());
            setError(null);
        } catch (err) {
            setError(String(err));
        }
    }, []);

    useEffect(() => {
        reload();
    }, [reload]);

    async function run(action: () => Promise<unknown>) {
        setNotice(null);
        try {
            await action();
        } catch (err) {
            setError(String(err));
            return;
        }
        await reload();
    }

    const confirmThen = (message: string, action: () => Promise<unknown>) => {
        if (window.confirm(message)) run(action);
    };

    const promptThen = (message: string, current: string, action: (value: string) => Promise<unknown>) => {
        const value = window.prompt(message, current);
        if (value !== null && value.trim() && value.trim() !== current) run(() => action(value.trim()));
    };

    async function stop() {
        await run(async () => {
            const result = await studiesApi.stop();
            setNotice(result.saved
                ? `Sessão salva: ${formatMinutes(result.minutes)}.`
                : 'Sessão descartada: menos de 1 minuto.');
        });
    }

    if (!overview) {
        return (
            <div>
                <h1>Estudos</h1>
                {error ? <p role="alert">{error}</p> : <p>Carregando...</p>}
            </div>
        );
    }

    return (
        <div>
            <h1>Estudos</h1>
            {error && <p role="alert">{error}</p>}
            {notice && <p role="status">{notice}</p>}

            <TimerPanel
                subjects={overview.subjects}
                active={overview.active}
                onStart={(subjectId, topicId) => run(() => studiesApi.start(subjectId, topicId))}
                onPause={() => run(studiesApi.pause)}
                onResume={() => run(studiesApi.resume)}
                onStop={stop}
            />

            <GoalsPanel
                goals={overview.goals}
                subjects={overview.subjects}
                onCreate={(subjectId, hours) => run(() => studiesApi.createGoal(subjectId, hours))}
                onUpdate={(id, hours) => run(() => studiesApi.updateGoal(id, hours))}
                onDelete={(goal) => confirmThen('Excluir esta meta?', () => studiesApi.deleteGoal(goal.id))}
            />

            <SubjectsPanel
                subjects={overview.subjects}
                archived={overview.archived}
                onCreateSubject={(name) => run(() => studiesApi.createSubject(name))}
                onRenameSubject={(s) => promptThen('Novo nome da matéria:', s.name, (name) => studiesApi.renameSubject(s.id, name))}
                onArchive={(s, archived) => archived
                    ? confirmThen(`Arquivar "${s.name}"? As horas continuam contando na meta geral.`, () => studiesApi.setSubjectArchived(s.id, true))
                    : run(() => studiesApi.setSubjectArchived(s.id, false))}
                onCreateTopic={(subjectId, title) => run(() => studiesApi.createTopic(subjectId, title))}
                onRenameTopic={(t) => promptThen('Novo nome do tópico:', t.title, (title) => studiesApi.renameTopic(t.id, title))}
                onSetStatus={(t, status) => run(() => studiesApi.setTopicStatus(t.id, status))}
                onDeleteTopic={(t) => confirmThen(`Excluir o tópico "${t.title}"? As sessões dele são mantidas.`, () => studiesApi.deleteTopic(t.id))}
            />

            <SessionsList
                sessions={overview.sessions}
                onDelete={(s) => confirmThen('Excluir esta sessão?', () => studiesApi.deleteSession(s.id))}
            />
        </div>
    );
}
