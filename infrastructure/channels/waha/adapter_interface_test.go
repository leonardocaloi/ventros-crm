package waha_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ventros/crm/infrastructure/channels/waha"
	"github.com/ventros/crm/internal/domain/crm/message"
)

// TestWhatsAppAck_ImplementsDeliveryStatusAdapter verifica que WhatsAppAck
// implementa corretamente a interface message.DeliveryStatusAdapter
func TestWhatsAppAck_ImplementsDeliveryStatusAdapter(t *testing.T) {
	// Cria um WhatsAppAck
	wahaAck, err := waha.NewWhatsAppAck(2) // DEVICE
	require.NoError(t, err)

	// Verifica que pode ser atribuído à interface
	var adapter message.DeliveryStatusAdapter = wahaAck

	// Verifica que a interface funciona
	assert.True(t, adapter.IsValid())
	assert.Equal(t, "DEVICE", adapter.String())

	status, err := adapter.ToMessageStatus()
	require.NoError(t, err)
	assert.Equal(t, message.StatusDelivered, status)
}

// TestDeliveryStatusAdapter_PolymorphicUsage demonstra uso polimórfico
func TestDeliveryStatusAdapter_PolymorphicUsage(t *testing.T) {
	// Simula processamento genérico de múltiplos provedores
	adapters := []struct {
		name    string
		adapter message.DeliveryStatusAdapter
		want    message.Status
	}{
		{
			name:    "WhatsApp ACK DEVICE",
			adapter: mustCreateAck(t, 2),
			want:    message.StatusDelivered,
		},
		{
			name:    "WhatsApp ACK READ",
			adapter: mustCreateAck(t, 3),
			want:    message.StatusRead,
		},
		{
			name:    "WhatsApp ACK SERVER",
			adapter: mustCreateAck(t, 1),
			want:    message.StatusSent,
		},
	}

	for _, tt := range adapters {
		t.Run(tt.name, func(t *testing.T) {
			// Código genérico que funciona com qualquer DeliveryStatusAdapter
			status, err := processDeliveryStatus(tt.adapter)
			require.NoError(t, err)
			assert.Equal(t, tt.want, status)
		})
	}
}

// TestDeliveryStatusAdapter_WithMultipleProviders demonstra como
// diferentes provedores podem usar a mesma interface
func TestDeliveryStatusAdapter_WithMultipleProviders(t *testing.T) {
	// Array de adapters de diferentes provedores (futuro: Telegram, SMS, etc.)
	var adapters []message.DeliveryStatusAdapter

	// WhatsApp provider
	wahaAck, err := waha.NewWhatsAppAck(2)
	require.NoError(t, err)
	adapters = append(adapters, wahaAck)

	// Future: Telegram provider
	// telegramUpdate := telegram.NewUpdate(telegram.UpdateDelivered)
	// adapters = append(adapters, telegramUpdate)

	// Future: SMS provider
	// smsReceipt := sms.NewDeliveryReceipt(sms.ReceiptDelivered)
	// adapters = append(adapters, smsReceipt)

	// Processa todos os adapters de forma genérica
	for i, adapter := range adapters {
		t.Run(adapter.String(), func(t *testing.T) {
			// Validação genérica
			assert.True(t, adapter.IsValid(), "adapter %d should be valid", i)

			// Conversão genérica
			status, err := adapter.ToMessageStatus()
			require.NoError(t, err, "adapter %d conversion failed", i)

			// Status deve ser um dos valores válidos do domain
			validStatuses := []message.Status{
				message.StatusQueued,
				message.StatusSent,
				message.StatusDelivered,
				message.StatusRead,
				message.StatusFailed,
			}
			assert.Contains(t, validStatuses, status, "adapter %d returned invalid status", i)
		})
	}
}

// TestDeliveryStatusAdapter_ErrorHandling testa tratamento de erros via interface
func TestDeliveryStatusAdapter_ErrorHandling(t *testing.T) {
	// ACK inválido (fora do range -1 a 4)
	invalidAck := waha.WhatsAppAck(99)

	var adapter message.DeliveryStatusAdapter = invalidAck

	// Verifica que IsValid retorna false
	assert.False(t, adapter.IsValid())

	// Verifica que ToMessageStatus retorna erro
	_, err := adapter.ToMessageStatus()
	assert.Error(t, err)
}

// Helper: processDeliveryStatus é uma função genérica que processa
// qualquer DeliveryStatusAdapter, demonstrando o poder da interface
func processDeliveryStatus(adapter message.DeliveryStatusAdapter) (message.Status, error) {
	// Validação genérica
	if !adapter.IsValid() {
		return "", assert.AnError
	}

	// Conversão genérica
	return adapter.ToMessageStatus()
}

// Helper: mustCreateAck cria um WhatsAppAck ou falha o teste
func mustCreateAck(t *testing.T, value int) waha.WhatsAppAck {
	t.Helper()
	ack, err := waha.NewWhatsAppAck(value)
	require.NoError(t, err)
	return ack
}
