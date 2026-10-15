import {FormEvent, useCallback, useEffect, useState} from 'react';
import {formatDay, formatPercent, QuestionLogInput, QuestionReport, RateRow, studiesApi, Subject} from './api';

const periods = [
    {days: 30, label: 'Últimos 30 dias'},
    {days: 90, label: 'Últimos 90 dias'},
    {days: 0, label: 'Tudo'},
];

const pad = (n: number) => String(n).padStart(2, '0');
const todayISO = () => {
    const d = new Date();
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
};

export function QuestionsTab() {
    const [subjects, setSubjects] = useState<Subject[]>([]);
    const [period, setPeriod] = useState(30);
    const [report, setReport] = useState<QuestionReport | null>(null);
    const [error, setError] = useState<string | null>(null);

    const reload = useCallback(async () => {
        try {
            const [overview, nextReport] = await Promise.all([studiesApi.overview(), studiesApi.questionReport(period)]);
            setSubjects(overview.subjects);
            setReport(nextReport);
            setError(null);
        } catch (err) {
            setError(String(err));
        }
    }, [period]);

    useEffect(() => {
        reload();
    }, [reload]);

    async function run(action: () => Promise<unknown>) {
        try {
            await action();
        } catch (err) {
            setError(String(err));
            return false;
        }
        await reload();
        return true;
    }

    return (
        <div>
            {error && <p role="alert">{error}</p>}
            <LogForm subjects={subjects} onSubmit={(input) => run(() => studiesApi.logQuestions(input))}/>

            <p>
                <label>
                    Período{' '}
                    <select value={period} onChange={(e) => setPeriod(Number(e.target.value))}>
                        {periods.map((p) => <option key={p.days} value={p.days}>{p.label}</option>)}
                    </select>
                </label>
            </p>

            {report && (
                <>
                    <RateTable title="Por matéria" rows={report.bySubject}/>
                    <RateTable title="Por tópico" rows={report.byTopic} showSubject/>

                    <section aria-label="Pontos fracos">
                        <h2>Pontos fracos</h2>
                        {report.weaknesses.length === 0
                            ? <p>Registre pelo menos 10 questões num tópico para ver seus pontos fracos.</p>
                            : (
                                <ol>
                                    {report.weaknesses.map((w) => (
                                        <li key={w.id}>
                                            {w.subjectName} · {w.name}: {formatPercent(w.rate)} ({w.correct}/{w.total})
                                        </li>
                                    ))}
                                </ol>
                            )}
                    </section>

                    <section aria-label="Registros">
                        <h2>Registros</h2>
                        {report.logs.length === 0 && <p>Nenhum registro no período.</p>}
                        <ul>
                            {report.logs.map((l) => (
                                <li key={l.id}>
                                    {formatDay(l.doneOn)} · {l.subjectName}
                                    {l.topicTitle && <> · {l.topicTitle}</>}
                                    {' · '}{l.correct}/{l.total} ({formatPercent(l.correct / l.total)})
                                    {l.source && <> · {l.source}</>}{' '}
                                    <button
                                        type="button"
                                        onClick={() => window.confirm('Excluir este registro?') && run(() => studiesApi.deleteQuestionLog(l.id))}
                                    >
                                        Excluir
                                    </button>
                                </li>
                            ))}
                        </ul>
                    </section>
                </>
            )}
        </div>
    );
}

function RateTable({title, rows, showSubject = false}: {title: string; rows: RateRow[]; showSubject?: boolean}) {
    return (
        <section aria-label={title}>
            <h2>{title}</h2>
            {rows.length === 0 ? <p>Sem dados no período.</p> : (
                <table>
                    <thead>
                        <tr>
                            {showSubject && <th>Matéria</th>}
                            <th>{showSubject ? 'Tópico' : 'Matéria'}</th>
                            <th>Feitas</th>
                            <th>Acertos</th>
                            <th>%</th>
                        </tr>
                    </thead>
                    <tbody>
                        {rows.map((r) => (
                            <tr key={r.id}>
                                {showSubject && <td>{r.subjectName}</td>}
                                <td>{r.name}</td>
                                <td>{r.total}</td>
                                <td>{r.correct}</td>
                                <td>{formatPercent(r.rate)}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            )}
        </section>
    );
}

function LogForm({subjects, onSubmit}: {subjects: Subject[]; onSubmit: (input: QuestionLogInput) => Promise<boolean>}) {
    const [subjectId, setSubjectId] = useState(0);
    const [topicId, setTopicId] = useState(0);
    const [doneOn, setDoneOn] = useState(todayISO);
    const [total, setTotal] = useState('');
    const [correct, setCorrect] = useState('');
    const [source, setSource] = useState('');
    const [invalid, setInvalid] = useState<string | null>(null);
    const subject = subjects.find((s) => s.id === subjectId);

    async function handleSubmit(e: FormEvent) {
        e.preventDefault();
        const t = Number(total);
        const c = Number(correct);
        if (!subject) return setInvalid('Escolha a matéria.');
        if (!Number.isInteger(t) || t < 1) return setInvalid('Informe quantas questões fez.');
        if (!Number.isInteger(c) || c < 0 || c > t) return setInvalid('Os acertos devem estar entre 0 e o número de questões.');
        setInvalid(null);
        const saved = await onSubmit({subjectId, topicId, doneOn, total: t, correct: c, source: source.trim()});
        if (saved) {
            setTotal('');
            setCorrect('');
        }
    }

    return (
        <form onSubmit={handleSubmit} aria-label="Registrar questões">
            <h2>Registrar questões</h2>
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
            <label>
                Data <input type="date" value={doneOn} max={todayISO()} onChange={(e) => setDoneOn(e.target.value)}/>
            </label>{' '}
            <label>
                Feitas <input type="number" min="1" value={total} onChange={(e) => setTotal(e.target.value)}/>
            </label>{' '}
            <label>
                Acertos <input type="number" min="0" value={correct} onChange={(e) => setCorrect(e.target.value)}/>
            </label>{' '}
            <label>
                Origem (opcional) <input placeholder="Ex.: Simulado FGV" value={source} onChange={(e) => setSource(e.target.value)}/>
            </label>{' '}
            <button type="submit">Registrar</button>
            {invalid && <p role="alert">{invalid}</p>}
        </form>
    );
}
