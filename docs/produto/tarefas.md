# Tarefas (To-do)

Regras de produto do módulo de tarefas. A implementação fica em `internal/tasks/` (Go), e os testes em `internal/tasks/*_test.go` cobrem cada regra abaixo.

## Tarefa

| Campo | Regra |
|---|---|
| Nome | Obrigatório. Espaços nas pontas são removidos. |
| Categoria | Opcional, uma por tarefa (ex.: Pessoal, Faculdade, Trabalho). |
| Início e fim | Data + hora, obrigatórios. **O fim deve ser depois do início.** |
| Urgência | Baixa, Média ou Alta. Padrão: **Média**. |
| Lista de itens | Opcional (ex.: lista de compras). |
| Repetição | Nenhuma, diária ou semanal. |
| Concluída | Sim/não. |

- Pode ter várias tarefas no mesmo dia, inclusive em horários sobrepostos.
- **Início no passado é permitido** (ex.: registrar algo que já começou). *A validar nos testes manuais.*
- Ao abrir o modal de nova tarefa: início na próxima hora cheia, fim uma hora depois.

## Ações

Criar (botão **Nova tarefa** → modal), editar (mesmo modal, preenchido), excluir (pede confirmação) e marcar/desmarcar como concluída.

## Categorias

- São digitadas livremente no modal, com sugestões das categorias já usadas.
- **Maiúsculas/minúsculas não duplicam:** digitar "faculdade" quando "Faculdade" já existe salva como "Faculdade".
- Não existe tela de gerenciar categorias. Uma categoria some das sugestões quando nenhuma tarefa a usa.

## Filtros

- Dois filtros no topo da agenda: **categoria** e **urgência**, cada um com a opção "Todas".
- Os dois se combinam (ex.: Faculdade + Alta) e valem para toda a agenda, inclusive as Atrasadas.

## Lista de itens (subtarefas)

- Os itens são as coisas a fazer dentro da tarefa. Ex.: na tarefa "Mercado", os itens são o que precisa comprar.
- No modal, marque **Tem lista de itens** para adicionar, renomear, remover e reordenar os itens (↑ ↓). Desmarcar apaga os itens ao salvar.
- Na agenda, os itens aparecem abaixo da tarefa, com check e contador (ex.: "Itens 2/5").
- **Conclusão automática nos dois sentidos:**
  - Todos os itens marcados → a tarefa é concluída.
  - Desmarcou um item → a tarefa volta a pendente.
  - Por isso, uma tarefa com itens **não** pode ser concluída na mão: o check dela fica desabilitado.
- Ao editar, os itens que continuam mantêm o check.

## Tarefas recorrentes

- **Diária** ou **semanal** com os dias escolhidos (ex.: seg, qua e sex). Data final ("repetir até") opcional.
- Todas as ocorrências têm o mesmo horário da primeira, e **começam e terminam no mesmo dia**.
- A série começa na data do início. No semanal, se esse dia não é um dos escolhidos, a primeira ocorrência é o próximo dia válido.
- **Cada ocorrência é independente:** concluir a de hoje não conclui a de amanhã. A lista de itens também zera em cada ocorrência.
- **Sem atraso:** uma ocorrência que passou sem ser feita simplesmente some. A de hoje continua em "Hoje" mesmo depois do horário.
- **Editar** muda a série inteira (o que já foi concluído continua concluído). **Excluir** apaga a série inteira.

## Agenda

Lista agrupada por dia, de **hoje em diante**, no fuso do computador.

1. **Atrasadas** (no topo): tarefas não recorrentes e não concluídas cujo fim já passou. Aparecem **só** aqui.
2. **Um grupo por dia** ("Hoje", "Amanhã" ou a data), só para dias que têm tarefas:
   - Uma tarefa aparece em **todos os dias que ocupa**. Ex.: 14/10 22:00 → 16/10 10:00 aparece em 14, 15 e 16.
   - Uma tarefa que termina exatamente à meia-noite **não** aparece no dia seguinte.
   - Uma tarefa que começou antes de hoje e ainda não terminou aparece a partir de hoje.
   - As recorrentes aparecem só nos **próximos 7 dias** (hoje + 6).
   - Dentro do dia, a ordem é pelo horário de início.
3. **Concluídas** aparecem riscadas no dia delas. As não recorrentes que terminaram antes de hoje somem da agenda (continuam no banco).

## Em aberto

- Início no passado: confirmar se faz sentido na prática.
- Tarefas concluídas antigas: hoje elas só ficam guardadas, sem tela de histórico.
- Editar a lista de itens de uma recorrente não recalcula as ocorrências que já foram concluídas.
