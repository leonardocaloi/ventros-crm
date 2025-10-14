package waha

import (
	"errors"

	"github.com/ventros/crm/internal/domain/crm/message"
)

// WhatsAppAck representa o valor de ACK (acknowledgment) do WhatsApp/WAHA.
//
// Este type implementa message.DeliveryStatusAdapter, servindo como adapter
// entre os valores específicos do WhatsApp e o Status agnóstico do domínio.
//
// O ciclo de vida de uma mensagem no WhatsApp passa por diferentes estados:
//
// -1 (ERROR)   → Erro ao enviar a mensagem
//
//	0 (PENDING) → Mensagem pendente (ainda não enviada)
//	1 (SERVER)  → Mensagem foi enviada ao servidor WhatsApp
//	2 (DEVICE)  → Mensagem foi entregue ao dispositivo do destinatário (✓✓)
//	3 (READ)    → Mensagem foi lida pelo destinatário (✓✓ azul)
//	4 (PLAYED)  → Mensagem de mídia foi reproduzida/visualizada
type WhatsAppAck int

const (
	// AckError indica que houve erro ao enviar a mensagem
	AckError WhatsAppAck = -1

	// AckPending indica que a mensagem está pendente (ainda não foi enviada)
	AckPending WhatsAppAck = 0

	// AckServer indica que a mensagem foi enviada ao servidor WhatsApp
	AckServer WhatsAppAck = 1

	// AckDevice indica que a mensagem foi entregue ao dispositivo do destinatário (✓✓)
	AckDevice WhatsAppAck = 2

	// AckRead indica que a mensagem foi lida pelo destinatário (✓✓ azul)
	AckRead WhatsAppAck = 3

	// AckPlayed indica que a mensagem de voz/áudio foi reproduzida (SOMENTE voice messages)
	AckPlayed WhatsAppAck = 4
)

var (
	ErrInvalidAck = errors.New("invalid WhatsApp ACK value")
)

// NewWhatsAppAck cria um WhatsAppAck a partir de um valor inteiro.
func NewWhatsAppAck(value int) (WhatsAppAck, error) {
	ack := WhatsAppAck(value)
	if !ack.IsValid() {
		return 0, ErrInvalidAck
	}
	return ack, nil
}

// IsValid verifica se o ACK é válido.
func (a WhatsAppAck) IsValid() bool {
	return a >= AckError && a <= AckPlayed
}

// String retorna a representação em string do ACK.
func (a WhatsAppAck) String() string {
	switch a {
	case AckError:
		return "ERROR"
	case AckPending:
		return "PENDING"
	case AckServer:
		return "SERVER"
	case AckDevice:
		return "DEVICE"
	case AckRead:
		return "READ"
	case AckPlayed:
		return "PLAYED"
	default:
		return "UNKNOWN"
	}
}

// ToMessageStatus implementa message.DeliveryStatusAdapter.
// Converte o ACK do WhatsApp/WAHA para o Status do domínio.
//
// Este é o adapter que traduz valores específicos do WhatsApp para o modelo agnóstico do domínio.
//
// Mapeamento:
// - ERROR (-1)   → message.StatusFailed
// - PENDING (0)  → message.StatusQueued
// - SERVER (1)   → message.StatusSent
// - DEVICE (2)   → message.StatusDelivered
// - READ (3)     → message.StatusRead
// - PLAYED (4)   → message.StatusPlayed (voz/áudio reproduzido - SOMENTE voice)
func (a WhatsAppAck) ToMessageStatus() (message.Status, error) {
	switch a {
	case AckError:
		return message.StatusFailed, nil
	case AckPending:
		return message.StatusQueued, nil
	case AckServer:
		return message.StatusSent, nil
	case AckDevice:
		return message.StatusDelivered, nil
	case AckRead:
		return message.StatusRead, nil
	case AckPlayed:
		return message.StatusPlayed, nil // Voz/áudio reproduzido (SOMENTE voice)
	default:
		return "", ErrInvalidAck
	}
}

// ToStatus é um alias para ToMessageStatus para manter compatibilidade.
// Deprecated: Use ToMessageStatus() que implementa a interface DeliveryStatusAdapter.
func (a WhatsAppAck) ToStatus() (message.Status, error) {
	return a.ToMessageStatus()
}

// Value retorna o valor inteiro do ACK.
func (a WhatsAppAck) Value() int {
	return int(a)
}

// IsError verifica se o ACK indica erro.
func (a WhatsAppAck) IsError() bool {
	return a == AckError
}

// IsPending verifica se a mensagem está pendente.
func (a WhatsAppAck) IsPending() bool {
	return a == AckPending
}

// IsSent verifica se a mensagem foi enviada ao servidor.
func (a WhatsAppAck) IsSent() bool {
	return a >= AckServer
}

// IsDelivered verifica se a mensagem foi entregue ao dispositivo.
func (a WhatsAppAck) IsDelivered() bool {
	return a >= AckDevice
}

// IsRead verifica se a mensagem foi lida.
func (a WhatsAppAck) IsRead() bool {
	return a >= AckRead
}

// IsPlayed verifica se a mensagem de voz/áudio foi reproduzida (SOMENTE voice).
func (a WhatsAppAck) IsPlayed() bool {
	return a == AckPlayed
}

// ShouldUpdateStatus verifica se o ACK deve atualizar o status da mensagem.
// ACKs de erro e pending geralmente não atualizam mensagens existentes.
func (a WhatsAppAck) ShouldUpdateStatus() bool {
	return a >= AckServer
}

// Verificação em tempo de compilação de que WhatsAppAck implementa DeliveryStatusAdapter
var _ message.DeliveryStatusAdapter = (*WhatsAppAck)(nil)
