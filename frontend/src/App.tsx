import {useState} from 'react';
import {Home} from './pages/Home';
import {TasksPage} from './features/tasks/TasksPage';

const pages = {
    home: {label: 'Home', component: Home},
    tasks: {label: 'Tarefas', component: TasksPage},
};

type PageKey = keyof typeof pages;

function App() {
    const [page, setPage] = useState<PageKey>('home');
    const Page = pages[page].component;

    return (
        <div id="App">
            <nav>
                {(Object.keys(pages) as PageKey[]).map((key) => (
                    <button
                        key={key}
                        type="button"
                        aria-current={key === page ? 'page' : undefined}
                        onClick={() => setPage(key)}
                    >
                        {pages[key].label}
                    </button>
                ))}
            </nav>
            <main>
                <Page/>
            </main>
        </div>
    );
}

export default App
