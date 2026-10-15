import {FormEvent, useState} from 'react';
import {formatDay, statusLabels, Subject, Topic, TopicStatus} from './api';

export interface SubjectActions {
    onCreateSubject: (name: string) => void;
    onRenameSubject: (subject: Subject) => void;
    onArchive: (subject: Subject, archived: boolean) => void;
    onCreateTopic: (subjectId: number, title: string) => void;
    onRenameTopic: (topic: Topic) => void;
    onSetStatus: (topic: Topic, status: TopicStatus) => void;
    onDeleteTopic: (topic: Topic) => void;
}

interface Props extends SubjectActions {
    subjects: Subject[];
    archived: Subject[];
}

export function SubjectsPanel({subjects, archived, ...actions}: Props) {
    return (
        <section aria-label="Matérias">
            <h2>Matérias</h2>
            {subjects.length === 0 && <p>Nenhuma matéria ainda.</p>}
            {subjects.map((s) => <SubjectCard key={s.id} subject={s} {...actions}/>)}

            <AddForm label="Nova matéria" button="Adicionar matéria" onAdd={actions.onCreateSubject}/>

            {archived.length > 0 && (
                <details>
                    <summary>Arquivadas ({archived.length})</summary>
                    <ul>
                        {archived.map((s) => (
                            <li key={s.id}>
                                {s.name}{' '}
                                <button type="button" onClick={() => actions.onArchive(s, false)}>Desarquivar</button>
                            </li>
                        ))}
                    </ul>
                </details>
            )}
        </section>
    );
}

function SubjectCard({subject, ...actions}: {subject: Subject} & SubjectActions) {
    const percent = Math.round(subject.progress * 100);

    return (
        <article aria-label={subject.name}>
            <h3>{subject.name}</h3>
            <p>
                <progress max={100} value={percent} aria-label={`Progresso de ${subject.name}`}/> {percent}%{' '}
                <button type="button" onClick={() => actions.onRenameSubject(subject)}>Renomear</button>{' '}
                <button type="button" onClick={() => actions.onArchive(subject, true)}>Arquivar</button>
            </p>
            <ul>
                {subject.topics.map((topic) => (
                    <li key={topic.id}>
                        {topic.title}{' '}
                        <select
                            aria-label={`Status de ${topic.title}`}
                            value={topic.status}
                            onChange={(e) => actions.onSetStatus(topic, e.target.value as TopicStatus)}
                        >
                            {Object.entries(statusLabels).map(([value, label]) => <option key={value} value={value}>{label}</option>)}
                        </select>{' '}
                        {topic.nextReviewOn && <>próxima revisão {formatDay(topic.nextReviewOn)}{' '}</>}
                        <button type="button" onClick={() => actions.onRenameTopic(topic)}>Renomear</button>{' '}
                        <button type="button" onClick={() => actions.onDeleteTopic(topic)}>Excluir</button>
                    </li>
                ))}
            </ul>
            <AddForm
                label={`Novo tópico em ${subject.name}`}
                button="Adicionar tópico"
                onAdd={(title) => actions.onCreateTopic(subject.id, title)}
            />
        </article>
    );
}

function AddForm({label, button, onAdd}: {label: string; button: string; onAdd: (text: string) => void}) {
    const [text, setText] = useState('');

    function handleSubmit(e: FormEvent) {
        e.preventDefault();
        if (!text.trim()) return;
        onAdd(text.trim());
        setText('');
    }

    return (
        <form onSubmit={handleSubmit}>
            <input aria-label={label} placeholder={label} value={text} onChange={(e) => setText(e.target.value)}/>{' '}
            <button type="submit">{button}</button>
        </form>
    );
}
