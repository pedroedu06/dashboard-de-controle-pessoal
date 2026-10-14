import {Filter, Urgency, urgencyLabels} from './api';

interface Props {
    categories: string[];
    value: Filter;
    onChange: (filter: Filter) => void;
}

export function TaskFilters({categories, value, onChange}: Props) {
    return (
        <p>
            <label>
                Filtrar categoria{' '}
                <select value={value.category} onChange={(e) => onChange({...value, category: e.target.value})}>
                    <option value="">Todas</option>
                    {categories.map((c) => <option key={c} value={c}>{c}</option>)}
                </select>
            </label>{' '}
            <label>
                Filtrar urgência{' '}
                <select value={value.urgency} onChange={(e) => onChange({...value, urgency: e.target.value as Urgency | ''})}>
                    <option value="">Todas</option>
                    {Object.entries(urgencyLabels).map(([v, label]) => <option key={v} value={v}>{label}</option>)}
                </select>
            </label>
        </p>
    );
}
