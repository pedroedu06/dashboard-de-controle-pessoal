# Estudos

Regras de produto do módulo de estudos. A implementação fica em `internal/studies/` (Go), e os testes em `internal/studies/*_test.go` cobrem cada regra abaixo.

## Matérias e tópicos

- A **matéria** tem nome (obrigatório) e uma lista de **tópicos**.
- Não podem existir duas matérias ativas com o mesmo nome, sem diferenciar maiúsculas ("Cálculo" e "cálculo" são a mesma).
- O **tópico** tem nome (obrigatório) e um status:

  | Status | Significa |
  |---|---|
  | Não visto | Ainda não começou |
  | Estudando | Já começou a estudar |
  | Revisado | Estudou e revisou |

- O status é escolhido num select e pode ir e voltar livremente.
- **Automático:** ao estudar no timer um tópico "Não visto", ele passa para "Estudando". "Revisado" é sempre você que marca.
- Matérias e tópicos podem ser renomeados.
- **Excluir um tópico** é definitivo. As sessões dele continuam, só que sem tópico.

### Arquivar matéria

Matérias não são apagadas, são **arquivadas**:

- A matéria some da lista, do timer e das metas por matéria.
- As horas já estudadas nela **continuam contando na meta geral**.
- Em "Arquivadas", dá para desarquivar, e tudo volta (inclusive as metas dela).
- Não dá para arquivar a matéria com o timer ligado nela.
- O nome de uma matéria arquivada pode ser usado por uma nova. Nesse caso, a antiga só pode ser desarquivada depois que a nova for renomeada ou arquivada.

## Progresso por matéria

A barra de progresso da matéria vai de 0% a 100%:

- Tópico **revisado** vale 1; **estudando** vale ½; **não visto** vale 0.
- Progresso = soma ÷ número de tópicos. Ex.: 4 tópicos, 1 revisado + 2 estudando → (1 + ½ + ½) ÷ 4 = **50%**.
- Matéria sem tópicos fica em 0%.

## Timer de estudo

- Cronômetro: **Iniciar → Pausar/Retomar → Encerrar**.
- Ao iniciar, escolha a matéria (obrigatória) e, se quiser, um tópico dela.
- Só existe **um timer ligado por vez**.
- **Pausas não contam** no tempo estudado.
- O timer fica salvo no banco: fechar o app não perde nada. Ao abrir de novo, ele continua de onde estava (inclusive pausado).
- Ao encerrar, sessões com **menos de 1 minuto são descartadas** (ex.: iniciou sem querer).

## Metas semanais

- A meta é em **horas por semana**: **geral** (todas as matérias) ou **de uma matéria**.
- A semana vai de **segunda 00:00 a domingo 23:59**, no fuso do computador.
- A meta vale toda semana até ser excluída. O alvo pode ser editado.
- O progresso soma as sessões da semana, **incluindo o timer ligado**. Ao atingir o alvo, a meta aparece como **Concluída**.
- Uma sessão conta na semana em que **começou**.

## Sessões da semana

Lista das sessões da semana atual (dia, matéria, tópico e duração), da mais recente para a mais antiga. Dá para **excluir** uma sessão errada (ex.: esqueceu o timer ligado), e ela deixa de contar nas metas.

## Telas

Estudos tem três subabas: **Visão geral** (matérias, timer, metas e sessões), **Revisões de hoje (N)**, com a quantidade pendente, e **Questões**.

## Revisão espaçada (SM-2)

Cada tópico guarda o seu ciclo de revisão: a data da próxima revisão, o intervalo em dias, a "facilidade" (começa em 2,5) e quantas revisões seguidas deram certo.

- **Marcar o tópico como "Revisado"** coloca ele na fila: a primeira revisão é **amanhã**.
- **Voltar o tópico para "Estudando" ou "Não visto"** tira ele da fila e zera o ciclo. Marcar "Revisado" de novo recomeça do zero.
- Em **Revisões de hoje** aparecem os tópicos com revisão para hoje ou atrasada (os mais atrasados primeiro), de matérias ativas. Para cada um, você diz como foi:

  | Botão | Nota SM-2 | Efeito |
  |---|---|---|
  | Esqueci | 1 | Volta para o começo: revisa **amanhã** de novo, e a facilidade cai |
  | Difícil | 3 | Avança, mas a facilidade cai (os próximos intervalos crescem menos) |
  | OK | 4 | Avança normalmente |
  | Fácil | 5 | Avança, e a facilidade sobe (os próximos intervalos crescem mais) |

- **Intervalos** (SM-2 clássico), contando a partir do dia da revisão:
  - Depois de marcar "Revisado": 1 dia.
  - Na revisão seguinte que der certo: 6 dias.
  - A partir daí: intervalo anterior × facilidade. Respondendo sempre **OK**: 1 → 6 → 15 → 38 → 95 dias...
- A facilidade nunca fica abaixo de 1,3.
- O tópico continua "Revisado" mesmo quando você esquece; só a data da próxima revisão muda.
- Na Visão geral, cada tópico revisado mostra a data da **próxima revisão**.

## Registro de questões

Em **Questões** você registra blocos de questões/simulados:

| Campo | Regra |
|---|---|
| Data | Padrão hoje; não pode ser no futuro |
| Matéria | Obrigatória (ativa) |
| Tópico | Opcional, da mesma matéria |
| Feitas | De 1 a 1000 |
| Acertos | De 0 até o número de feitas |
| Origem | Opcional (ex.: "Simulado FGV 03", "QConcursos") |

O relatório usa o período escolhido (**últimos 30 dias**, **90 dias** ou **tudo**):

- **Por matéria** e **por tópico**: feitas, acertos e % de acerto. Registros sem tópico contam só na matéria.
- **Pontos fracos**: os **5 tópicos com menor % de acerto** entre os que têm **pelo menos 10 questões** no período. No empate, aparece primeiro o que tem mais questões.
- **Registros**: a lista do período, com excluir.
- Matérias arquivadas ficam fora do relatório.

## Em aberto

- Lançar uma sessão manual (estudo feito sem o timer).
- Caderno de erros (anotar cada questão errada e o porquê).
- Histórico de semanas anteriores.
