
endpoint de ocntoato events q filtra por tiopd eevento, eento terao categoraisa

Use Case: FetchContactProfilePicture
Recebe phone + session
Chama WAHA service
Atualiza contact
Dispara evento contact.profile_picture_updated
Consumer: ContactProfilePictureUpdated
Escuta evento
Cria ContactEvent na timeline
Calcular Métricas de Sessão
Ao registrar primeira mensagem do contato
Ao registrar primeira resposta do agente
Calcular tempos automaticamente
AI Session Summary Worker
Escuta evento session.ended
Verifica se pipeline tem enable_ai_summary = true
Cria AIProcessing com status pending
Processa resumo via IA
Atualiza session com summary

da file waha.events.message (que somente recebe) ele porcessa, é  campo personliado q tem q ser criado de imagem. e se tem imagem ou nao pela api da waha. puzar imagem tbm pela api da waha. puxa dados de contato, aqui entra contatct CREATE oR UPDATE entende?

webhook gerado do canal fraco

no final da session ele tem q calcular o tempo de atendimento e tempo de espera da primeira mensagem, no segundo caso, se a primeira mensagem for do agente, ele vai calcular o tempo de espera da primeira mensagem do contato, se for do contato ele vai calcular o tempo de espera da primeira mensagem do agente, um para lead score e outro para atendimento e feeback comercial.

a informcao, de ponta a ponta, vai estar protegida por fila se der erro, pela coregorafai com rmq, ou pelo temporal saga orquestraçao?

saga de compensaço de mensagem enviada com webhook recebido

ativar opçao de resumo inteligente de sessao ao final dentro de pipeline tambem e isso vai precisar de entidades de processaementos de ia dentro do go, correto? 

tem q gerar o reply to no waha


o agnete_id pode ser o 











📋 TASKS - Ventros CRM v0.1.0

## Validações Importantes

### Antes de Ativar Canal
- [ ] Sessão WAHA está rodando
- [ ] Sessão está com status `WORKING` (conectada ao WhatsApp)
- [ ] Webhook está configurado corretamente

### Antes de Importar Histórico (Temporal with SAGA)
- [ ] Canal está ativo
- [ ] Estratégia de importação está configurada (não `none`)
- [ ] Banco de dados tem espaço suficiente
- [ ] Timeout do HTTP está adequado (pode demorar)

### Após Importação (Temporal with SAGA)
- [ ] Todas mensagens têm `session_id` não nulo
- [ ] Sessões foram criadas corretamente
- [ ] Contatos foram criados/associados
- [ ] Mensagens estão ordenadas por timestamp
- [ ] Tipos de mídia foram mapeados corretamente


