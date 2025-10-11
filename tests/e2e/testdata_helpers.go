package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// LoadWAHAEvent carrega um arquivo de evento WAHA do testdata
func LoadWAHAEvent(t *testing.T, filename string) map[string]interface{} {
	// Tenta vários caminhos possíveis (relativo ao diretório de execução do teste)
	paths := []string{
		filepath.Join("testdata", "events_waha", filename),
		filepath.Join("tests", "e2e", "testdata", "events_waha", filename),
		filepath.Join("../../testdata/events_waha", filename),
		filepath.Join("../testdata/events_waha", filename),
	}

	var data []byte
	var err error
	var successPath string

	for _, path := range paths {
		data, err = os.ReadFile(path)
		if err == nil {
			successPath = path
			break
		}
	}

	if err != nil {
		t.Fatalf("Failed to load WAHA event file %s (tried paths: %v): %v", filename, paths, err)
	}

	var event map[string]interface{}
	err = json.Unmarshal(data, &event)
	assert.NoError(t, err, "Failed to parse event JSON from %s", successPath)

	return event
}

// UpdateWAHAEventSession atualiza o session_id no evento WAHA
func UpdateWAHAEventSession(event map[string]interface{}, sessionID string) {
	event["session"] = sessionID

	// Atualiza também nos metadados internos
	if payload, ok := event["payload"].(map[string]interface{}); ok {
		if data, ok := payload["_data"].(map[string]interface{}); ok {
			if info, ok := data["Info"].(map[string]interface{}); ok {
				info["Chat"] = sessionID + "@s.whatsapp.net"
			}
		}
	}
}

// GetAvailableWAHAEvents retorna lista de eventos WAHA disponíveis
func GetAvailableWAHAEvents() []string {
	return []string{
		"message_text.json",
		"message_image.json",
		"message_audio.json",
		"message_recorded_audio.json",
		"message_document_pdf.json",
		"message_location.json",
		"message_contact.json",
		"message_image_text.json",
		"message_sticker.json",
		"fb_ads_message.json",
	}
}
