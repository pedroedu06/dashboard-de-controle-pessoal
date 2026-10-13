import {render, screen} from '@testing-library/react';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import {Home} from './Home';
import {DBStatus} from '../../wailsjs/go/main/App';

vi.mock('../../wailsjs/go/main/App', () => ({DBStatus: vi.fn()}));

const mockDBStatus = vi.mocked(DBStatus);

describe('Home', () => {
    beforeEach(() => {
        mockDBStatus.mockReset();
    });

    it('shows "conectado" when the database responds', async () => {
        mockDBStatus.mockResolvedValue({connected: true});

        render(<Home/>);

        expect(await screen.findByText('Banco: conectado')).toBeInTheDocument();
    });

    it('shows the error returned by Go', async () => {
        mockDBStatus.mockResolvedValue({connected: false, error: 'DATABASE_URL não definida'});

        render(<Home/>);

        expect(await screen.findByText(/DATABASE_URL não definida/)).toBeInTheDocument();
    });

    it('shows the error when the call fails', async () => {
        mockDBStatus.mockRejectedValue(new Error('falhou'));

        render(<Home/>);

        expect(await screen.findByText(/falhou/)).toBeInTheDocument();
    });
});
