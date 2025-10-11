package waha_test

import (
	"testing"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/internal/domain/crm/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWhatsAppAck(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		wantAck   waha.WhatsAppAck
		wantError bool
	}{
		{
			name:      "valid ACK ERROR",
			value:     -1,
			wantAck:   waha.AckError,
			wantError: false,
		},
		{
			name:      "valid ACK PENDING",
			value:     0,
			wantAck:   waha.AckPending,
			wantError: false,
		},
		{
			name:      "valid ACK SERVER",
			value:     1,
			wantAck:   waha.AckServer,
			wantError: false,
		},
		{
			name:      "valid ACK DEVICE",
			value:     2,
			wantAck:   waha.AckDevice,
			wantError: false,
		},
		{
			name:      "valid ACK READ",
			value:     3,
			wantAck:   waha.AckRead,
			wantError: false,
		},
		{
			name:      "valid ACK PLAYED",
			value:     4,
			wantAck:   waha.AckPlayed,
			wantError: false,
		},
		{
			name:      "invalid ACK -2",
			value:     -2,
			wantError: true,
		},
		{
			name:      "invalid ACK 5",
			value:     5,
			wantError: true,
		},
		{
			name:      "invalid ACK 100",
			value:     100,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ack, err := waha.NewWhatsAppAck(tt.value)

			if tt.wantError {
				assert.Error(t, err)
				assert.Equal(t, waha.ErrInvalidAck, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantAck, ack)
			assert.Equal(t, tt.value, ack.Value())
		})
	}
}

func TestWhatsAppAck_String(t *testing.T) {
	tests := []struct {
		ack  waha.WhatsAppAck
		want string
	}{
		{waha.AckError, "ERROR"},
		{waha.AckPending, "PENDING"},
		{waha.AckServer, "SERVER"},
		{waha.AckDevice, "DEVICE"},
		{waha.AckRead, "READ"},
		{waha.AckPlayed, "PLAYED"},
		{waha.WhatsAppAck(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ack.String())
		})
	}
}

func TestWhatsAppAck_ToStatus(t *testing.T) {
	tests := []struct {
		name       string
		ack        waha.WhatsAppAck
		wantStatus message.Status
		wantError  bool
	}{
		{
			name:       "ERROR maps to failed",
			ack:        waha.AckError,
			wantStatus: message.StatusFailed,
		},
		{
			name:       "PENDING maps to queued",
			ack:        waha.AckPending,
			wantStatus: message.StatusQueued,
		},
		{
			name:       "SERVER maps to sent",
			ack:        waha.AckServer,
			wantStatus: message.StatusSent,
		},
		{
			name:       "DEVICE maps to delivered",
			ack:        waha.AckDevice,
			wantStatus: message.StatusDelivered,
		},
		{
			name:       "READ maps to read",
			ack:        waha.AckRead,
			wantStatus: message.StatusRead,
		},
		{
			name:       "PLAYED maps to read",
			ack:        waha.AckPlayed,
			wantStatus: message.StatusRead,
		},
		{
			name:      "invalid ACK returns error",
			ack:       waha.WhatsAppAck(99),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := tt.ack.ToStatus()

			if tt.wantError {
				assert.Error(t, err)
				assert.Equal(t, waha.ErrInvalidAck, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, status)
		})
	}
}

func TestWhatsAppAck_IsError(t *testing.T) {
	tests := []struct {
		ack  waha.WhatsAppAck
		want bool
	}{
		{waha.AckError, true},
		{waha.AckPending, false},
		{waha.AckServer, false},
		{waha.AckDevice, false},
		{waha.AckRead, false},
		{waha.AckPlayed, false},
	}

	for _, tt := range tests {
		t.Run(tt.ack.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ack.IsError())
		})
	}
}

func TestWhatsAppAck_IsPending(t *testing.T) {
	tests := []struct {
		ack  waha.WhatsAppAck
		want bool
	}{
		{waha.AckError, false},
		{waha.AckPending, true},
		{waha.AckServer, false},
		{waha.AckDevice, false},
		{waha.AckRead, false},
		{waha.AckPlayed, false},
	}

	for _, tt := range tests {
		t.Run(tt.ack.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ack.IsPending())
		})
	}
}

func TestWhatsAppAck_IsSent(t *testing.T) {
	tests := []struct {
		ack  waha.WhatsAppAck
		want bool
	}{
		{waha.AckError, false},
		{waha.AckPending, false},
		{waha.AckServer, true},
		{waha.AckDevice, true},
		{waha.AckRead, true},
		{waha.AckPlayed, true},
	}

	for _, tt := range tests {
		t.Run(tt.ack.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ack.IsSent())
		})
	}
}

func TestWhatsAppAck_IsDelivered(t *testing.T) {
	tests := []struct {
		ack  waha.WhatsAppAck
		want bool
	}{
		{waha.AckError, false},
		{waha.AckPending, false},
		{waha.AckServer, false},
		{waha.AckDevice, true},
		{waha.AckRead, true},
		{waha.AckPlayed, true},
	}

	for _, tt := range tests {
		t.Run(tt.ack.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ack.IsDelivered())
		})
	}
}

func TestWhatsAppAck_IsRead(t *testing.T) {
	tests := []struct {
		ack  waha.WhatsAppAck
		want bool
	}{
		{waha.AckError, false},
		{waha.AckPending, false},
		{waha.AckServer, false},
		{waha.AckDevice, false},
		{waha.AckRead, true},
		{waha.AckPlayed, true},
	}

	for _, tt := range tests {
		t.Run(tt.ack.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ack.IsRead())
		})
	}
}

func TestWhatsAppAck_IsPlayed(t *testing.T) {
	tests := []struct {
		ack  waha.WhatsAppAck
		want bool
	}{
		{waha.AckError, false},
		{waha.AckPending, false},
		{waha.AckServer, false},
		{waha.AckDevice, false},
		{waha.AckRead, false},
		{waha.AckPlayed, true},
	}

	for _, tt := range tests {
		t.Run(tt.ack.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ack.IsPlayed())
		})
	}
}

func TestWhatsAppAck_ShouldUpdateStatus(t *testing.T) {
	tests := []struct {
		ack  waha.WhatsAppAck
		want bool
	}{
		{waha.AckError, false},
		{waha.AckPending, false},
		{waha.AckServer, true},
		{waha.AckDevice, true},
		{waha.AckRead, true},
		{waha.AckPlayed, true},
	}

	for _, tt := range tests {
		t.Run(tt.ack.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ack.ShouldUpdateStatus())
		})
	}
}

func TestWhatsAppAck_IsValid(t *testing.T) {
	tests := []struct {
		name string
		ack  waha.WhatsAppAck
		want bool
	}{
		{"valid -1", waha.AckError, true},
		{"valid 0", waha.AckPending, true},
		{"valid 1", waha.AckServer, true},
		{"valid 2", waha.AckDevice, true},
		{"valid 3", waha.AckRead, true},
		{"valid 4", waha.AckPlayed, true},
		{"invalid -2", waha.WhatsAppAck(-2), false},
		{"invalid 5", waha.WhatsAppAck(5), false},
		{"invalid 100", waha.WhatsAppAck(100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ack.IsValid())
		})
	}
}

// TestWhatsAppAck_LifecycleCoverage testa o ciclo completo de vida de uma mensagem
func TestWhatsAppAck_LifecycleCoverage(t *testing.T) {
	// Simula o ciclo de vida de uma mensagem bem-sucedida
	lifecycle := []struct {
		ack            waha.WhatsAppAck
		expectedStatus message.Status
		shouldUpdate   bool
	}{
		{waha.AckPending, message.StatusQueued, false},   // Pendente
		{waha.AckServer, message.StatusSent, true},       // Enviada ao servidor
		{waha.AckDevice, message.StatusDelivered, true},  // Entregue no dispositivo
		{waha.AckRead, message.StatusRead, true},         // Lida pelo destinatário
	}

	for _, step := range lifecycle {
		t.Run(step.ack.String(), func(t *testing.T) {
			status, err := step.ack.ToStatus()
			require.NoError(t, err)
			assert.Equal(t, step.expectedStatus, status)
			assert.Equal(t, step.shouldUpdate, step.ack.ShouldUpdateStatus())
		})
	}
}

// TestWhatsAppAck_MediaLifecycle testa o ciclo de vida de mensagens de mídia
func TestWhatsAppAck_MediaLifecycle(t *testing.T) {
	// Para mensagens de mídia (áudio/vídeo), o ciclo pode ir até PLAYED
	lifecycle := []struct {
		ack            waha.WhatsAppAck
		expectedStatus message.Status
	}{
		{waha.AckServer, message.StatusSent},
		{waha.AckDevice, message.StatusDelivered},
		{waha.AckRead, message.StatusRead},
		{waha.AckPlayed, message.StatusRead}, // Played também é mapeado para "read"
	}

	for _, step := range lifecycle {
		t.Run(step.ack.String(), func(t *testing.T) {
			status, err := step.ack.ToStatus()
			require.NoError(t, err)
			assert.Equal(t, step.expectedStatus, status)

			// Todas as etapas após SERVER devem atualizar status
			assert.True(t, step.ack.ShouldUpdateStatus())
		})
	}
}

// TestWhatsAppAck_ErrorScenario testa cenário de erro
func TestWhatsAppAck_ErrorScenario(t *testing.T) {
	ack := waha.AckError

	assert.True(t, ack.IsError())
	assert.False(t, ack.IsSent())
	assert.False(t, ack.IsDelivered())
	assert.False(t, ack.IsRead())
	assert.False(t, ack.ShouldUpdateStatus())

	status, err := ack.ToStatus()
	require.NoError(t, err)
	assert.Equal(t, message.StatusFailed, status)
}
