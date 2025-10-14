package message

// Source representa a origem de uma mensagem
type Source string

const (
	SourceManual        Source = "manual"         // Enviada por agente humano
	SourceBroadcast     Source = "broadcast"      // Campanha broadcast
	SourceSequence      Source = "sequence"       // Sequência de automação
	SourceTrigger       Source = "trigger"        // Trigger/regra de pipeline
	SourceBot           Source = "bot"            // Bot/AI response
	SourceSystem        Source = "system"         // Sistema interno
	SourceWebhook       Source = "webhook"        // Resposta automática via webhook
	SourceScheduled     Source = "scheduled"      // Mensagem agendada
	SourceTest          Source = "test"           // Envio de teste (E2E, development)
	SourceHistoryImport Source = "history_import" // Importação de histórico
)

// IsValid verifica se a source é válida
func (s Source) IsValid() bool {
	switch s {
	case SourceManual, SourceBroadcast, SourceSequence,
		SourceTrigger, SourceBot, SourceSystem, SourceWebhook, SourceScheduled, SourceTest, SourceHistoryImport:
		return true
	}
	return false
}

// IsAutomated verifica se a source é de automação (não manual)
func (s Source) IsAutomated() bool {
	return s != SourceManual
}

// String retorna a representação em string da source
func (s Source) String() string {
	return string(s)
}
