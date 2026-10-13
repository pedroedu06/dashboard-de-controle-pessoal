import {render, screen} from '@testing-library/react';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import App from './App';
import {DBStatus} from '../wailsjs/go/main/App';

// Os bindings do Wails só existem dentro do app, então nos testes são mockados.
vi.mock('../wailsjs/go/main/App', () => ({
    DBStatus: vi.fn(),
}));

const mockDBStatus = vi.mocked(DBStatus);

describe('App', () => {
    beforeEach(() => {
        mockDBStatus.mockReset();
    });

    it('mostra "conectado" quando o banco responde', async () => {
        mockDBStatus.mockResolvedValue({connected: true});

        render(<App/>);

        expect(await screen.findByText('Banco: conectado')).toBeInTheDocument();
    });

    it('mostra o erro retornado pelo Go', async () => {
        mockDBStatus.mockResolvedValue({connected: false, error: 'DATABASE_URL não definida'});

        render(<App/>);

        expect(await screen.findByText(/DATABASE_URL não definida/)).toBeInTheDocument();
    });

    it('mostra o erro quando a chamada falha', async () => {
        mockDBStatus.mockRejectedValue(new Error('falhou'));

        render(<App/>);

        expect(await screen.findByText(/falhou/)).toBeInTheDocument();
    });
});
