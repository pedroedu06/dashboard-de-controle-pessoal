# Tarefas (To-do)

Regras de produto da primeira versão. A implementação fica em `internal/tasks/` (Go), e os testes em `internal/tasks/*_test.go` cobrem cada regra abaixo.

## Tarefa

| Campo | Regra |
|---|---|
| Nome | Obrigatório. Espaços nas pontas são removidos. |
| Início e fim | Data + hora, obrigatórios. **O fim deve ser depois do início.** |
| Urgência | Baixa, Média ou Alta. Padrão: **Média**. |
| Concluída | Sim/não. Dá para desmarcar. |

- Pode ter várias tarefas no mesmo dia, inclusive em horários sobrepostos.
- **Início no passado é permitido** (ex.: registrar algo que já começou). *A validar nos testes manuais.*
- Ao abrir o modal de nova tarefa: início na próxima hora cheia, fim uma hora depois.

## Ações

Criar (botão **Nova tarefa** → modal), editar (mesmo modal, preenchido), excluir (pede confirmação) e marcar/desmarcar como concluída.

## Agenda

Lista agrupada por dia, de **hoje em diante**, no fuso do computador.

1. **Atrasadas** (no topo): não concluídas cujo fim já passou. Aparecem **só** aqui.
2. **Um grupo por dia** ("Hoje", "Amanhã" ou a data), só para dias que têm tarefas:
   - Uma tarefa aparece em **todos os dias que ocupa**. Ex.: 14/10 22:00 → 16/10 10:00 aparece em 14, 15 e 16.
   - Uma tarefa que termina exatamente à meia-noite **não** aparece no dia seguinte.
   - Uma tarefa que começou antes de hoje e ainda não terminou aparece a partir de hoje.
   - Dentro do dia, a ordem é pelo horário de início.
3. **Concluídas** aparecem riscadas no dia delas. As que terminaram antes de hoje somem da agenda (continuam no banco).

## Em aberto

- Início no passado: confirmar se faz sentido na prática.
- Tarefas concluídas antigas: hoje elas só ficam guardadas, sem tela de histórico.
