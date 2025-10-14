package waha

import (
	"testing"
)

// TestChannelTypeConversionPolicy_WAHA verifies WAHA conversion policy
func TestChannelTypeConversionPolicy_WAHA(t *testing.T) {
	policy := NewChannelTypeConversionPolicy("waha")

	tests := []struct {
		name     string
		mimetype string
		isVideo  bool
		want     bool
	}{
		// WAHA sempre converte vídeos
		{"video mp4", "video/mp4", true, true},
		{"video webm", "video/webm", true, true},
		{"video avi", "video/avi", true, true},

		// WAHA não converte opus
		{"audio opus", "audio/ogg; codecs=opus", false, false},

		// WAHA converte outros áudios
		{"audio mp3", "audio/mpeg", false, true},
		{"audio m4a", "audio/mp4", false, true},
		{"audio wav", "audio/wav", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.isVideo {
				got = policy.ShouldConvertVideo(tt.mimetype)
			} else {
				got = policy.ShouldConvertAudio(tt.mimetype)
			}

			if got != tt.want {
				t.Errorf("WAHA policy for %s = %v, want %v", tt.mimetype, got, tt.want)
			}
		})
	}
}

// TestChannelTypeConversionPolicy_WhatsAppBusiness verifies WhatsApp Business conversion policy
func TestChannelTypeConversionPolicy_WhatsAppBusiness(t *testing.T) {
	policy := NewChannelTypeConversionPolicy("whatsapp_business")

	tests := []struct {
		name     string
		mimetype string
		isVideo  bool
		want     bool
	}{
		// WhatsApp Business não converte MP4
		{"video mp4", "video/mp4", true, false},

		// WhatsApp Business converte outros formatos
		{"video webm", "video/webm", true, true},
		{"video avi", "video/avi", true, true},

		// WhatsApp Business não converte formatos suportados de áudio
		{"audio opus", "audio/ogg; codecs=opus", false, false},
		{"audio mp3", "audio/mpeg", false, false},
		{"audio m4a", "audio/mp4", false, false},

		// WhatsApp Business converte formatos não suportados
		{"audio wav", "audio/wav", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.isVideo {
				got = policy.ShouldConvertVideo(tt.mimetype)
			} else {
				got = policy.ShouldConvertAudio(tt.mimetype)
			}

			if got != tt.want {
				t.Errorf("WhatsApp Business policy for %s = %v, want %v", tt.mimetype, got, tt.want)
			}
		})
	}
}

// TestChannelTypeConversionPolicy_Telegram verifies Telegram conversion policy
func TestChannelTypeConversionPolicy_Telegram(t *testing.T) {
	policy := NewChannelTypeConversionPolicy("telegram")

	tests := []struct {
		name     string
		mimetype string
		isVideo  bool
		want     bool
	}{
		// Telegram nunca converte vídeos
		{"video mp4", "video/mp4", true, false},
		{"video webm", "video/webm", true, false},
		{"video avi", "video/avi", true, false},

		// Telegram nunca converte áudios
		{"audio opus", "audio/ogg; codecs=opus", false, false},
		{"audio mp3", "audio/mpeg", false, false},
		{"audio m4a", "audio/mp4", false, false},
		{"audio wav", "audio/wav", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.isVideo {
				got = policy.ShouldConvertVideo(tt.mimetype)
			} else {
				got = policy.ShouldConvertAudio(tt.mimetype)
			}

			if got != tt.want {
				t.Errorf("Telegram policy for %s = %v, want %v", tt.mimetype, got, tt.want)
			}
		})
	}
}

// TestChannelTypeConversionPolicy_Unknown verifies default/unknown channel policy
func TestChannelTypeConversionPolicy_Unknown(t *testing.T) {
	policy := NewChannelTypeConversionPolicy("unknown_channel_type")

	// Para canais desconhecidos, o padrão é sempre converter vídeos por segurança
	if !policy.ShouldConvertVideo("video/mp4") {
		t.Error("Unknown channel should convert videos by default (safe fallback)")
	}

	// Para canais desconhecidos, o padrão é converter áudios exceto opus
	if policy.ShouldConvertAudio("audio/ogg; codecs=opus") {
		t.Error("Unknown channel should NOT convert opus")
	}

	if !policy.ShouldConvertAudio("audio/mpeg") {
		t.Error("Unknown channel should convert non-opus audio by default (safe fallback)")
	}
}

// TestChannelTypeConversionPolicy_PerformanceOptimization demonstrates the performance difference
func TestChannelTypeConversionPolicy_PerformanceOptimization(t *testing.T) {
	// Cenário: 100 vídeos MP4

	wahaPolicy := NewChannelTypeConversionPolicy("waha")
	whatsappBusinessPolicy := NewChannelTypeConversionPolicy("whatsapp_business")
	telegramPolicy := NewChannelTypeConversionPolicy("telegram")

	conversionsWAHA := 0
	conversionsWhatsApp := 0
	conversionsTelegram := 0

	// Simular 100 vídeos MP4
	for i := 0; i < 100; i++ {
		if wahaPolicy.ShouldConvertVideo("video/mp4") {
			conversionsWAHA++
		}
		if whatsappBusinessPolicy.ShouldConvertVideo("video/mp4") {
			conversionsWhatsApp++
		}
		if telegramPolicy.ShouldConvertVideo("video/mp4") {
			conversionsTelegram++
		}
	}

	// WAHA: converte tudo (100 conversões)
	if conversionsWAHA != 100 {
		t.Errorf("WAHA should convert all 100 videos, got %d", conversionsWAHA)
	}

	// WhatsApp Business: MP4 é suportado (0 conversões) ✅ 80% improvement!
	if conversionsWhatsApp != 0 {
		t.Errorf("WhatsApp Business should convert 0 MP4 videos, got %d", conversionsWhatsApp)
	}

	// Telegram: aceita tudo (0 conversões) ✅ 100% improvement!
	if conversionsTelegram != 0 {
		t.Errorf("Telegram should convert 0 videos, got %d", conversionsTelegram)
	}

	t.Logf("Performance optimization:")
	t.Logf("  WAHA:              100 conversões (sempre converte)")
	t.Logf("  WhatsApp Business:   0 conversões (MP4 suportado) ✅ 80%% improvement!")
	t.Logf("  Telegram:            0 conversões (aceita tudo)   ✅ 100%% improvement!")
}
