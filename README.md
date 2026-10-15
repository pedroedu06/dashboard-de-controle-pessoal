# Dashboard

App desktop pessoal para gerenciar a vida: tarefas, finanças, estudos e alimentação.

**Stack:** [Wails v2](https://wails.io) (Go + React/TypeScript/Vite) · Postgres no [Supabase](https://supabase.com) · [pgx](https://github.com/jackc/pgx) + [sqlc](https://sqlc.dev) + [goose](https://github.com/pressly/goose) · testes com `go test` e [Vitest](https://vitest.dev) · comandos com [Task](https://taskfile.dev).

O React chama funções Go (bindings do Wails), e o Go conversa com o banco. O app só funciona online.

**Funcionalidades:** [Tarefas](docs/produto/tarefas.md) (to-do com agenda, categorias, filtros, lista de itens e tarefas recorrentes) e [Estudos](docs/produto/estudos.md) (matérias, tópicos, timer, progresso, metas semanais, revisão espaçada e registro de questões). As regras de produto de cada módulo ficam em `docs/produto/`.

---

## Pré-requisitos

| Ferramenta | Versão | Para quê |
|---|---|---|
| [Go](https://go.dev/dl/) | 1.26+ | backend e ferramentas (`sqlc`, `goose`, `task` já vêm fixadas no `go.mod`) |
| [Node.js](https://nodejs.org) | 20+ | frontend |
| Wails CLI | v2 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` (confira com `wails doctor`) |
| [Docker Desktop](https://www.docker.com/products/docker-desktop/) | — | só para os testes de integração |

Não precisa instalar `sqlc`, `goose` nem `task`: o projeto roda tudo com `go tool <nome>`.

## Rodando localmente

### 1. Crie o banco no Supabase

1. Em [supabase.com](https://supabase.com), crie um projeto (região **South America (São Paulo)**) e **guarde a senha do banco**.
2. No projeto, clique em **Connect** → aba **Connection String** → em *Method*, escolha **Session pooler**.
3. Copie a string. Ela tem o formato `postgresql://postgres.<ref>:<senha>@aws-0-sa-east-1.pooler.supabase.com:5432/postgres`.

> **Por que Session pooler?** A conexão *Direct* do Supabase só funciona em redes IPv6, e o *Transaction pooler* (porta 6543) não suporta os prepared statements que o pgx usa. O Session pooler (porta 5432) funciona em qualquer rede.

### 2. Configure o `.env`

Crie um `.env` na raiz seguindo [docs/ambiente.md](docs/ambiente.md), que lista as variáveis, de onde tirar cada uma e os cuidados com a senha.

### 3. Instale as dependências e rode

```bash
go tool task setup
```

O `setup` instala as dependências do frontend e gera os bindings do Wails (`frontend/wailsjs/go/`). Esses arquivos ficam fora do Git: são gerados a partir do código Go e mudam a cada alteração nele. O `wails dev` e o `wails build` também os geram.

```bash
go tool task dev
```

A janela do app abre com hot reload. Na primeira execução, as migrations são aplicadas automaticamente. Se aparecer **"Banco: conectado"**, está tudo certo. Se aparecer um erro, ele diz o que falta (por exemplo, `DATABASE_URL não definida`).

Também dá para abrir no navegador, em http://localhost:34115, com acesso às funções Go.

## Testes

```bash
go tool task test
```

Isso roda os testes Go (`go test ./...`) e os do frontend (Vitest).

| Tipo | Onde | Precisa de |
|---|---|---|
| Unitários Go | `internal/config`, `internal/tasks` (regras da agenda), `app_test.go`, `task_api_test.go` | nada |
| Integração Go | `internal/db`, `internal/tasks/service_test.go`, `app_test.go`, `task_api_test.go` | `TEST_DATABASE_URL` + Docker |
| Frontend | `frontend/src/**/*.test.tsx` | nada (bindings do Wails são mockados) |

### Testes de integração (banco real)

Eles rodam num **Postgres local no Docker**, nunca no Supabase:

```bash
go tool task db:test:up
```

```bash
go tool task test
```

- Sem `TEST_DATABASE_URL` (ou com o Docker desligado), esses testes aparecem como **SKIP**, e o resto continua rodando.
- Cada teste roda dentro de uma transação que é desfeita no final (`testutil.Queries`), então um teste não afeta o outro.
- Por segurança, os testes falham se `TEST_DATABASE_URL` for igual a `DATABASE_URL`.
- Para ver os detalhes: `go tool task test:go -- -v`.

## Banco de dados

O schema é versionado com **goose** (`internal/db/migrations`), e as queries são SQL puro, que o **sqlc** transforma em código Go tipado (`internal/db/sqlc`, **não edite à mão**).

Fluxo para mudar o banco:

1. Crie a migration: `go tool task migrate:new -- create_tasks`
2. Escreva o SQL em `internal/db/migrations/0000N_create_tasks.sql`, com as seções `-- +goose Up` e `-- +goose Down`.
3. Escreva as queries em `internal/db/queries/<tabela>.sql` (veja `settings.sql` como exemplo).
4. Gere o código: `go tool task generate`
5. Use no Go: `sqlc.New(pool).MinhaQuery(ctx, ...)`.

O app aplica as migrations pendentes ao iniciar. Para aplicar ou inspecionar manualmente no Supabase: `go tool task migrate:status`, `migrate:up`, `migrate:down`.

## Comandos

Todos rodam com `go tool task <nome>`. Para listar: `go tool task --list`.

| Comando | O que faz |
|---|---|
| `setup` | `npm install` no frontend + gera os bindings do Wails |
| `bindings` | gera os bindings TS do Wails (`frontend/wailsjs/go/`) |
| `dev` | abre o app em modo desenvolvimento |
| `build` | gera o executável em `build/bin` |
| `test` | todos os testes (Go + frontend) |
| `test:go` / `test:front` | só Go / só frontend |
| `lint` | `go vet`, `sqlc diff` (código gerado em dia?) e `tsc --noEmit` |
| `generate` | gera o código do sqlc e os bindings TS do Wails |
| `migrate:new -- nome` | cria uma migration nova |
| `migrate:up` / `migrate:down` / `migrate:status` | migrations no `DATABASE_URL` |
| `db:test:up` / `db:test:down` | sobe/derruba o Postgres de testes (porta 54329) |
| `db:test:reset` | apaga os dados do Postgres de testes (use se rodou o app contra ele) |

## Estrutura

```
main.go, app.go          entrada do Wails e status do banco
task_api.go              funções de tarefas expostas ao frontend
study_api.go             funções de estudos expostas ao frontend
internal/
  config/                carrega DATABASE_URL (ambiente / .env)
  tasks/                 regras de produto das tarefas (validação, agenda) e persistência
  studies/               regras de produto dos estudos (timer, progresso, metas, SM-2, questões) e persistência
  db/                    conexão (pgxpool) e migrations (goose, embutidas no binário)
    migrations/          arquivos .sql versionados
    queries/             queries SQL lidas pelo sqlc
    sqlc/                código Go gerado (não editar)
  testutil/              helpers dos testes de integração
frontend/
  src/
    pages/               telas simples (Home)
    features/tasks/      tela de tarefas: agenda, modal e chamadas ao Go
    features/studies/    tela de estudos: visão geral, revisões de hoje e questões
  wailsjs/               bindings gerados pelo Wails (go/ fica fora do Git)
sqlc.yaml                configuração do sqlc
Taskfile.yml             comandos do projeto
docker-compose.yml       Postgres de testes
docs/                    ambiente e regras de produto
```
