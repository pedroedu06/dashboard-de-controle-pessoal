import {KeyboardEvent, useState} from 'react';

export interface EditorItem {
    key: string; // stable React key, also for unsaved items
    id: number; // 0 for new items
    title: string;
}

interface Props {
    items: EditorItem[];
    onChange: (items: EditorItem[]) => void;
}

let nextKey = 0;
export const newEditorItem = (title: string, id = 0): EditorItem => ({key: `item-${nextKey++}`, id, title});

export function ItemsEditor({items, onChange}: Props) {
    const [draft, setDraft] = useState('');

    function add() {
        if (!draft.trim()) return;
        onChange([...items, newEditorItem(draft.trim())]);
        setDraft('');
    }

    function handleKeyDown(e: KeyboardEvent) {
        if (e.key === 'Enter') {
            e.preventDefault(); // don't submit the task form
            add();
        }
    }

    function move(index: number, delta: number) {
        const next = [...items];
        [next[index], next[index + delta]] = [next[index + delta], next[index]];
        onChange(next);
    }

    const rename = (index: number, title: string) =>
        onChange(items.map((it, i) => (i === index ? {...it, title} : it)));

    return (
        <fieldset>
            <legend>Itens da lista</legend>
            <ol>
                {items.map((item, i) => (
                    <li key={item.key}>
                        <input
                            aria-label={`Item ${i + 1}`}
                            value={item.title}
                            onChange={(e) => rename(i, e.target.value)}
                        />{' '}
                        <button type="button" aria-label={`Subir item ${i + 1}`} disabled={i === 0} onClick={() => move(i, -1)}>↑</button>
                        <button type="button" aria-label={`Descer item ${i + 1}`} disabled={i === items.length - 1} onClick={() => move(i, 1)}>↓</button>{' '}
                        <button type="button" aria-label={`Remover item ${i + 1}`} onClick={() => onChange(items.filter((_, j) => j !== i))}>Remover</button>
                    </li>
                ))}
            </ol>
            <input
                aria-label="Novo item"
                placeholder="Novo item"
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
                onKeyDown={handleKeyDown}
            />{' '}
            <button type="button" onClick={add}>Adicionar</button>
        </fieldset>
    );
}
