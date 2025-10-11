package waha

import (
	"testing"

	"go.uber.org/zap"
)

func TestIdentifierExtractor_ExtractFromMessageEvent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	extractor := NewIdentifierExtractor(logger)

	tests := []struct {
		name           string
		event          WAHAMessageEvent
		expectedWID    string
		expectedHasLID bool
		expectedHasJID bool
		expectedIDType string
	}{
		{
			name: "Complete IDs - WID, LID and JID",
			event: WAHAMessageEvent{
				Payload: WAHAPayload{
					From: "5511999999999@c.us",
					Data: WAHAData{
						Info: WAHAInfo{
							Chat:   "5511999999999@c.us",
							Sender: "5511999999999@c.us",
						},
					},
				},
				Me: WAHAMe{
					ID:  "5511888888888@s.whatsapp.net",
					LID: "5511999999999@lid",
					JID: "5511999999999@s.whatsapp.net",
				},
			},
			expectedWID:    "5511999999999", // Normalizado
			expectedHasLID: true,
			expectedHasJID: true,
			expectedIDType: "BOTH_AVAILABLE",
		},
		{
			name: "Only JID available - constructs WID from JID",
			event: WAHAMessageEvent{
				Payload: WAHAPayload{
					From: "5511999999999@s.whatsapp.net",
				},
				Me: WAHAMe{
					JID: "5511999999999@s.whatsapp.net",
				},
			},
			expectedWID:    "5511999999999", // Construído do JID
			expectedHasLID: false,
			expectedHasJID: true,
			expectedIDType: "PHONE_ONLY",
		},
		{
			name: "Only LID available",
			event: WAHAMessageEvent{
				Payload: WAHAPayload{
					From: "5511999999999@lid",
				},
				Me: WAHAMe{
					LID: "5511999999999@lid",
				},
			},
			expectedWID:    "5511999999999", // Normalizado do LID
			expectedHasLID: true,
			expectedHasJID: false,
			expectedIDType: "LID_ONLY",
		},
		{
			name: "Standard @c.us format",
			event: WAHAMessageEvent{
				Payload: WAHAPayload{
					From: "5511999999999@c.us",
				},
			},
			expectedWID:    "5511999999999",
			expectedHasLID: false,
			expectedHasJID: false,
			expectedIDType: "UNKNOWN",
		},
		{
			name: "Multiple candidates - picks correct format",
			event: WAHAMessageEvent{
				Payload: WAHAPayload{
					From: "5511999999999@c.us",
					Data: WAHAData{
						Info: WAHAInfo{
							Chat:   "5511888888888@s.whatsapp.net", // JID alternativo
							Sender: "5511777777777@lid",            // LID alternativo
						},
					},
				},
				Me: WAHAMe{
					ID:  "5511666666666@c.us",
					LID: "5511999999999@lid",
					JID: "5511999999999@s.whatsapp.net",
				},
			},
			expectedWID:    "5511999999999",
			expectedHasLID: true,
			expectedHasJID: true,
			expectedIDType: "BOTH_AVAILABLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identifiers, err := extractor.ExtractFromMessageEvent(tt.event)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
				return
			}

			if identifiers == nil {
				t.Error("Expected identifiers, got nil")
				return
			}

			// Verifica WID normalizado
			if identifiers.WID() != tt.expectedWID {
				t.Errorf("WID mismatch: expected %s, got %s", tt.expectedWID, identifiers.WID())
			}

			// Verifica LID
			if identifiers.HasLID() != tt.expectedHasLID {
				t.Errorf("LID presence mismatch: expected %v, got %v", tt.expectedHasLID, identifiers.HasLID())
			}

			// Verifica JID
			if identifiers.HasJID() != tt.expectedHasJID {
				t.Errorf("JID presence mismatch: expected %v, got %v", tt.expectedHasJID, identifiers.HasJID())
			}
		})
	}
}

func TestIdentifierExtractor_FindByFormat(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	extractor := NewIdentifierExtractor(logger)

	candidates := []string{
		"5511999999999@c.us",
		"5511888888888@s.whatsapp.net",
		"5511777777777@lid",
		"",
		"5511666666666@g.us",
	}

	tests := []struct {
		format   string
		expected string
	}{
		{"@c.us", "5511999999999@c.us"},
		{"@s.whatsapp.net", "5511888888888@s.whatsapp.net"},
		{"@lid", "5511777777777@lid"},
		{"@g.us", "5511666666666@g.us"},
		{"@unknown", ""},
	}

	for _, tt := range tests {
		t.Run("format_"+tt.format, func(t *testing.T) {
			result := extractor.findByFormat(candidates, tt.format)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIdentifierExtractor_ExtractPhoneNumber(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	extractor := NewIdentifierExtractor(logger)

	tests := []struct {
		input    string
		expected string
	}{
		{"5511999999999@c.us", "5511999999999"},
		{"5511999999999@s.whatsapp.net", "5511999999999"},
		{"5511999999999@lid", "5511999999999"},
		{"5511999999999@g.us", "5511999999999"},
		{"5511999999999", "5511999999999"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractor.extractPhoneNumber(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIdentifierExtractor_ContainsFormat(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	extractor := NewIdentifierExtractor(logger)

	tests := []struct {
		identifier string
		format     string
		expected   bool
	}{
		{"5511999999999@c.us", "@c.us", true},
		{"5511999999999@s.whatsapp.net", "@s.whatsapp.net", true},
		{"5511999999999@lid", "@lid", true},
		{"5511999999999", "@c.us", false},
		{"", "@c.us", false},
		{"5511999999999@c.us", "", false},
		{"@c.us", "@c.us", true},
		{"5511999999999@c.us", "@s.whatsapp.net", false},
	}

	for _, tt := range tests {
		t.Run(tt.identifier+"_"+tt.format, func(t *testing.T) {
			result := extractor.containsFormat(tt.identifier, tt.format)
			if result != tt.expected {
				t.Errorf("containsFormat(%s, %s): expected %v, got %v",
					tt.identifier, tt.format, tt.expected, result)
			}
		})
	}
}
