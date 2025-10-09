package waha

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ProfileService gerencia operações de perfil do WhatsApp via WAHA
type ProfileService struct {
	baseURL string
	apiKey  string
	client  *http.Client
	logger  *zap.Logger
}

// NewProfileService cria um novo serviço de perfil
func NewProfileService(baseURL, apiKey string, logger *zap.Logger) *ProfileService {
	return &ProfileService{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// CheckContactExistsResponse resposta do endpoint check-exists
type CheckContactExistsResponse struct {
	NumberExists bool   `json:"numberExists"`
	ChatID       string `json:"chatId"`
}

// ProfilePictureResponse resposta do endpoint profile-picture
type ProfilePictureResponse struct {
	URL string `json:"url"`
}

// CheckContactExists verifica se um contato existe no WhatsApp e retorna o chatId
func (s *ProfileService) CheckContactExists(ctx context.Context, phone, session string) (*CheckContactExistsResponse, error) {
	url := fmt.Sprintf("%s/api/contacts/check-exists?phone=%s&session=%s", s.baseURL, phone, session)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("WAHA API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result CheckContactExistsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetProfilePicture busca a URL da foto de perfil de um contato
func (s *ProfileService) GetProfilePicture(ctx context.Context, chatID, session string) (string, error) {
	url := fmt.Sprintf("%s/api/contacts/profile-picture?contactId=%s&session=%s", s.baseURL, chatID, session)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Contato não tem foto de perfil
		return "", nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("WAHA API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result ProfilePictureResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.URL, nil
}

// FetchAndUpdateContactProfilePicture busca e atualiza a foto de perfil de um contato
func (s *ProfileService) FetchAndUpdateContactProfilePicture(ctx context.Context, phone, session string) (string, error) {
	// 1. Verificar se o contato existe e obter chatId
	existsResp, err := s.CheckContactExists(ctx, phone, session)
	if err != nil {
		s.logger.Error("Failed to check if contact exists",
			zap.String("phone", phone),
			zap.Error(err))
		return "", err
	}

	if !existsResp.NumberExists {
		s.logger.Debug("Contact does not exist on WhatsApp",
			zap.String("phone", phone))
		return "", fmt.Errorf("contact does not exist on WhatsApp")
	}

	// 2. Buscar foto de perfil
	profilePictureURL, err := s.GetProfilePicture(ctx, existsResp.ChatID, session)
	if err != nil {
		s.logger.Error("Failed to get profile picture",
			zap.String("chat_id", existsResp.ChatID),
			zap.Error(err))
		return "", err
	}

	if profilePictureURL == "" {
		s.logger.Debug("Contact has no profile picture",
			zap.String("chat_id", existsResp.ChatID))
		return "", nil
	}

	s.logger.Info("Profile picture fetched successfully",
		zap.String("phone", phone),
		zap.String("chat_id", existsResp.ChatID),
		zap.String("url", profilePictureURL))

	return profilePictureURL, nil
}
