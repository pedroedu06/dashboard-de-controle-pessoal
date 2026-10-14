import {useEffect, useState} from 'react';
import {formatClock, StudySession, Subject} from './api';

interface Props {
    subjects: Subject[];
    active: StudySession | null;
    onStart: (subjectId: number, topicId: number) => void;
    onPause: () => void;
    onResume: () => void;
    onStop: () => void;
}

export function TimerPanel({subjects, active, ...actions}: Props) {
    return (
        <section aria-label="Timer">
            <h2>Timer</h2>
            {active ? <RunningTimer session={active} {...actions}/> : <StartForm subjects={subjects} onStart={actions.onStart}/>}
        </section>
    );
}

function StartForm({subjects, onStart}: Pick<Props, 'subjects' | 'onStart'>) {
    const [subjectId, setSubjectId] = useState(0);
    const [topicId, setTopicId] = useState(0);
    const subject = subjects.find((s) => s.id === subjectId);

    if (subjects.length === 0) {
        return <p>Crie uma matéria para começar a estudar.</p>;
    }

    return (
        <p>
            <label>
                Matéria{' '}
                <select value={subjectId} onChange={(e) => { setSubjectId(Number(e.target.value)); setTopicId(0); }}>
                    <option value={0}>Escolha...</option>
                    {subjects.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
            </label>{' '}
            <label>
                Tópico{' '}
                <select value={topicId} onChange={(e) => setTopicId(Number(e.target.value))} disabled={!subject}>
                    <option value={0}>Sem tópico</option>
                    {subject?.topics.map((t) => <option key={t.id} value={t.id}>{t.title}</option>)}
                </select>
            </label>{' '}
            <button type="button" disabled={!subject} onClick={() => onStart(subjectId, topicId)}>Iniciar</button>
        </p>
    );
}

type RunningProps = Omit<Props, 'subjects' | 'active' | 'onStart'> & {session: StudySession};

function RunningTimer({session, onPause, onResume, onStop}: RunningProps) {
    const elapsed = useElapsed(session);

    return (
        <div>
            <p>
                Estudando <strong>{session.subjectName}</strong>
                {session.topicTitle && <> · {session.topicTitle}</>}
                {session.paused && <> (pausado)</>}
            </p>
            <p role="timer" aria-label="Tempo estudado">{formatClock(elapsed)}</p>
            {session.paused
                ? <button type="button" onClick={onResume}>Retomar</button>
                : <button type="button" onClick={onPause}>Pausar</button>}{' '}
            <button type="button" onClick={onStop}>Encerrar</button>
        </div>
    );
}

/** Seconds studied, ticking every second from the server's value unless paused. */
function useElapsed(session: StudySession): number {
    const [elapsed, setElapsed] = useState(session.elapsedSeconds);

    useEffect(() => {
        setElapsed(session.elapsedSeconds);
        if (session.paused) return;
        const loadedAt = Date.now();
        const id = setInterval(() => {
            setElapsed(session.elapsedSeconds + Math.floor((Date.now() - loadedAt) / 1000));
        }, 1000);
        return () => clearInterval(id);
    }, [session]);

    return elapsed;
}
