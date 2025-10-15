package contact_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ventros/crm/internal/domain/crm/contact"
)

func TestNormalizeWhatsAppID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantError bool
	}{
		{
			name:  "normalize ID with @c.us suffix",
			input: "5511999999999@c.us",
			want:  "5511999999999",
		},
		{
			name:  "normalize ID with @s.whatsapp.net suffix",
			input: "5511999999999@s.whatsapp.net",
			want:  "5511999999999",
		},
		{
			name:  "normalize ID with @lid suffix",
			input: "5511999999999@lid",
			want:  "5511999999999",
		},
		{
			name:  "already normalized ID",
			input: "5511999999999",
			want:  "5511999999999",
		},
		{
			name:  "ID with spaces",
			input: "  5511999999999  ",
			want:  "5511999999999",
		},
		{
			name:      "empty ID",
			input:     "",
			wantError: true,
		},
		{
			name:      "invalid characters",
			input:     "abc123@c.us",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := contact.NormalizeWhatsAppID(tt.input)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewWhatsAppIdentifiers(t *testing.T) {
	tests := []struct {
		name      string
		wid       string
		lid       *string
		jid       *string
		wantWID   string
		wantError bool
	}{
		{
			name:    "valid WID only",
			wid:     "5511999999999@c.us",
			wantWID: "5511999999999",
		},
		{
			name:    "valid WID with LID",
			wid:     "5511999999999@c.us",
			lid:     strPtr("5511888888888@lid"),
			wantWID: "5511999999999",
		},
		{
			name:    "valid WID with JID",
			wid:     "5511999999999@c.us",
			jid:     strPtr("5511777777777@s.whatsapp.net"),
			wantWID: "5511999999999",
		},
		{
			name:    "valid WID with LID and JID",
			wid:     "5511999999999@c.us",
			lid:     strPtr("5511888888888@lid"),
			jid:     strPtr("5511777777777@s.whatsapp.net"),
			wantWID: "5511999999999",
		},
		{
			name:      "empty WID",
			wid:       "",
			wantError: true,
		},
		{
			name:      "invalid WID",
			wid:       "invalid@c.us",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := contact.NewWhatsAppIdentifiers(tt.wid, tt.lid, tt.jid)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantWID, ids.WID())

			if tt.lid != nil {
				assert.True(t, ids.HasLID())
				assert.NotNil(t, ids.LID())
			} else {
				assert.False(t, ids.HasLID())
			}

			if tt.jid != nil {
				assert.True(t, ids.HasJID())
				assert.NotNil(t, ids.JID())
			} else {
				assert.False(t, ids.HasJID())
			}
		})
	}
}

func TestWhatsAppIdentifiers_ToCustomFields(t *testing.T) {
	tests := []struct {
		name         string
		wid          string
		lid          *string
		jid          *string
		expectedKeys []string
	}{
		{
			name:         "WID only",
			wid:          "5511999999999@c.us",
			expectedKeys: []string{"waha_wid"},
		},
		{
			name:         "WID with LID",
			wid:          "5511999999999@c.us",
			lid:          strPtr("5511888888888@lid"),
			expectedKeys: []string{"waha_wid", "waha_lid"},
		},
		{
			name:         "WID with JID",
			wid:          "5511999999999@c.us",
			jid:          strPtr("5511777777777@s.whatsapp.net"),
			expectedKeys: []string{"waha_wid", "waha_jid"},
		},
		{
			name:         "WID with LID and JID",
			wid:          "5511999999999@c.us",
			lid:          strPtr("5511888888888@lid"),
			jid:          strPtr("5511777777777@s.whatsapp.net"),
			expectedKeys: []string{"waha_wid", "waha_lid", "waha_jid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := contact.NewWhatsAppIdentifiers(tt.wid, tt.lid, tt.jid)
			require.NoError(t, err)

			fields := ids.ToCustomFields()

			// Verifica que todas as chaves esperadas existem
			for _, key := range tt.expectedKeys {
				_, exists := fields[key]
				assert.True(t, exists, "expected key %s not found", key)
			}

			// Verifica que não há chaves extras
			assert.Equal(t, len(tt.expectedKeys), len(fields))
		})
	}
}

func TestWhatsAppIdentifiers_Equals(t *testing.T) {
	lid1 := "5511888888888"
	jid1 := "5511777777777"

	tests := []struct {
		name  string
		ids1  *contact.WhatsAppIdentifiers
		ids2  *contact.WhatsAppIdentifiers
		equal bool
	}{
		{
			name:  "equal WID only",
			ids1:  mustCreateIdentifiers(t, "5511999999999", nil, nil),
			ids2:  mustCreateIdentifiers(t, "5511999999999", nil, nil),
			equal: true,
		},
		{
			name:  "different WID",
			ids1:  mustCreateIdentifiers(t, "5511999999999", nil, nil),
			ids2:  mustCreateIdentifiers(t, "5511888888888", nil, nil),
			equal: false,
		},
		{
			name:  "equal WID and LID",
			ids1:  mustCreateIdentifiers(t, "5511999999999", &lid1, nil),
			ids2:  mustCreateIdentifiers(t, "5511999999999", &lid1, nil),
			equal: true,
		},
		{
			name:  "equal WID and JID",
			ids1:  mustCreateIdentifiers(t, "5511999999999", nil, &jid1),
			ids2:  mustCreateIdentifiers(t, "5511999999999", nil, &jid1),
			equal: true,
		},
		{
			name:  "one has LID, other doesn't",
			ids1:  mustCreateIdentifiers(t, "5511999999999", &lid1, nil),
			ids2:  mustCreateIdentifiers(t, "5511999999999", nil, nil),
			equal: false,
		},
		{
			name:  "compare with nil",
			ids1:  mustCreateIdentifiers(t, "5511999999999", nil, nil),
			ids2:  nil,
			equal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.equal, tt.ids1.Equals(tt.ids2))
		})
	}
}

func TestWhatsAppIdentifiers_String(t *testing.T) {
	lid := "5511888888888"
	jid := "5511777777777"

	tests := []struct {
		name     string
		wid      string
		lid      *string
		jid      *string
		contains []string
	}{
		{
			name:     "WID only",
			wid:      "5511999999999",
			contains: []string{"WID: 5511999999999"},
		},
		{
			name:     "WID with LID",
			wid:      "5511999999999",
			lid:      &lid,
			contains: []string{"WID: 5511999999999", "LID: 5511888888888"},
		},
		{
			name:     "WID with JID",
			wid:      "5511999999999",
			jid:      &jid,
			contains: []string{"WID: 5511999999999", "JID: 5511777777777"},
		},
		{
			name:     "all identifiers",
			wid:      "5511999999999",
			lid:      &lid,
			jid:      &jid,
			contains: []string{"WID: 5511999999999", "LID: 5511888888888", "JID: 5511777777777"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := contact.NewWhatsAppIdentifiers(tt.wid, tt.lid, tt.jid)
			require.NoError(t, err)

			str := ids.String()
			for _, expected := range tt.contains {
				assert.Contains(t, str, expected)
			}
		})
	}
}

// Helper functions

func strPtr(s string) *string {
	return &s
}

func mustCreateIdentifiers(t *testing.T, wid string, lid, jid *string) *contact.WhatsAppIdentifiers {
	t.Helper()
	ids, err := contact.NewWhatsAppIdentifiers(wid, lid, jid)
	require.NoError(t, err)
	return ids
}
