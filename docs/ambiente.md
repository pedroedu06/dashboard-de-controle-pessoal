# Variáveis de ambiente

Crie um arquivo `.env` na raiz do projeto com o conteúdo abaixo. O `.env` está no `.gitignore` e **nunca** deve ser commitado.

```env
DATABASE_URL=postgresql://postgres.<ref>:<senha>@aws-0-sa-east-1.pooler.supabase.com:5432/postgres
TEST_DATABASE_URL=postgres://postgres:postgres@localhost:54329/dashboard_test?sslmode=disable
```

| Variável | Para quê | Onde conseguir |
|---|---|---|
| `DATABASE_URL` | Banco do app (Supabase) | Supabase → projeto → **Connect** → **Connection String** → Method: **Session pooler** |
| `TEST_DATABASE_URL` | Banco dos testes de integração | Fixa: é o Postgres do `docker-compose.yml`. Copie como está. |

## Detalhes

- **Use o Session pooler (porta 5432).** A conexão *Direct* só funciona em IPv6, e o *Transaction pooler* (6543) quebra os prepared statements do pgx.
- **Senha com caracteres especiais** (`@ # / ? %`) precisa ser codificada na URL: `@` → `%40`, `#` → `%23`, `/` → `%2F`, `?` → `%3F`, `%` → `%25`.
- **`TEST_DATABASE_URL` nunca pode ser igual a `DATABASE_URL`.** Os testes se recusam a rodar nesse caso, para proteger os dados reais.
- **Onde o app procura:** primeiro nas variáveis de ambiente, depois no `.env` do diretório atual e, por último, em `%AppData%\dashboard\.env` (para o executável compilado).
