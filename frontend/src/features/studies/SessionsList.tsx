import {formatMinutes, StudySession} from './api';

interface Props {
    sessions: StudySession[];
    onDelete: (session: StudySession) => void;
}

const startLabel = (iso: string) =>
    new Date(iso).toLocaleString('pt-BR', {weekday: 'short', day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit'});

export function SessionsList({sessions, onDelete}: Props) {
    return (
        <section aria-label="Sessões da semana">
            <h2>Sessões da semana</h2>
            {sessions.length === 0 && <p>Nenhuma sessão nesta semana.</p>}
            <ul>
                {sessions.map((s) => (
                    <li key={s.id}>
                        {startLabel(s.startedAt)} · {s.subjectName}
                        {s.topicTitle && <> · {s.topicTitle}</>}
                        {' · '}{formatMinutes(Math.floor(s.elapsedSeconds / 60))}{' '}
                        <button type="button" onClick={() => onDelete(s)}>Excluir</button>
                    </li>
                ))}
            </ul>
        </section>
    );
}
