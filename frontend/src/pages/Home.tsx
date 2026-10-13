import {useEffect, useState} from 'react';
import {DBStatus} from '../../wailsjs/go/main/App';
import {main} from '../../wailsjs/go/models';

type Status = main.DBStatus | 'loading';

export function Home() {
    const [status, setStatus] = useState<Status>('loading');

    useEffect(() => {
        DBStatus()
            .then(setStatus)
            .catch((err) => setStatus({connected: false, error: String(err)}));
    }, []);

    return (
        <div>
            <h1>Bem-vindo ao Dashboard</h1>
            <p role="status">{statusText(status)}</p>
        </div>
    );
}

function statusText(status: Status): string {
    if (status === 'loading') return 'Banco: verificando...';
    if (status.connected) return 'Banco: conectado';
    return `Banco: erro: ${status.error ?? 'desconhecido'}`;
}
