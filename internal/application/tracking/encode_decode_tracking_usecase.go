package tracking

import (
	"fmt"
	"net/url"

	"github.com/caloi/ventros-crm/internal/domain/crm/tracking"
	"go.uber.org/zap"
)

// EncodeTrackingUseCase codifica tracking ID em mensagem com caracteres invisíveis
type EncodeTrackingUseCase struct {
	encoder *tracking.TernaryEncoder
	logger  *zap.Logger
}

// NewEncodeTrackingUseCase cria novo use case de encode
func NewEncodeTrackingUseCase(logger *zap.Logger) *EncodeTrackingUseCase {
	return &EncodeTrackingUseCase{
		encoder: tracking.NewTernaryEncoder(),
		logger:  logger,
	}
}

// Execute executa a codificação
func (uc *EncodeTrackingUseCase) Execute(req EncodeTrackingRequest) (*EncodeTrackingResponse, error) {
	// Converter tracking ID para ternário
	ternary, err := uc.encoder.DecimalToTernary(req.TrackingID)
	if err != nil {
		uc.logger.Error("failed to convert tracking ID to ternary",
			zap.Int64("tracking_id", req.TrackingID),
			zap.Error(err))
		return nil, fmt.Errorf("erro ao converter tracking ID: %w", err)
	}

	// Codificar ternário em caracteres invisíveis
	encodedChars, err := uc.encoder.EncodeTernary(ternary)
	if err != nil {
		uc.logger.Error("failed to encode ternary",
			zap.String("ternary", ternary),
			zap.Error(err))
		return nil, fmt.Errorf("erro ao codificar ternário: %w", err)
	}

	// Inserir código na mensagem
	messageWithCode, err := uc.encoder.EncodeMessage(req.Message, req.TrackingID)
	if err != nil {
		uc.logger.Error("failed to encode message",
			zap.String("message", req.Message),
			zap.Error(err))
		return nil, fmt.Errorf("erro ao codificar mensagem: %w", err)
	}

	// Gerar link do WhatsApp se telefone fornecido
	var whatsappLink string
	if req.Phone != "" {
		whatsappLink = fmt.Sprintf("https://wa.me/%s?text=%s",
			req.Phone,
			url.QueryEscape(messageWithCode))
	}

	// Informações de debug
	debugInfo := uc.buildDebugInfo(req.TrackingID, ternary, encodedChars, req.Message)

	response := &EncodeTrackingResponse{
		Success:         true,
		TrackingID:      req.TrackingID,
		OriginalMessage: req.Message,
		TernaryEncoded:  ternary,
		DecimalValue:    req.TrackingID,
		Phone:           req.Phone,
		InvisibleCode:   encodedChars,
		MessageWithCode: messageWithCode,
		WhatsAppLink:    whatsappLink,
		Debug:           debugInfo,
	}

	uc.logger.Info("tracking encoded successfully",
		zap.Int64("tracking_id", req.TrackingID),
		zap.String("ternary", ternary),
		zap.Int("invisible_code_length", len(encodedChars)))

	return response, nil
}

// buildDebugInfo constrói informações de debug
func (uc *EncodeTrackingUseCase) buildDebugInfo(trackingID int64, ternary, encodedChars, originalMessage string) map[string]interface{} {
	charCodes := make([]int, 0, len(encodedChars))
	charMapping := make([]string, 0, len(encodedChars))

	for i, char := range encodedChars {
		code := int(char)
		charCodes = append(charCodes, code)

		if i < len(ternary) {
			digitValue := int(ternary[i] - '0')
			charMapping = append(charMapping, fmt.Sprintf("Digit %d → SAFE_CHAR[%d] (U+%04X)",
				digitValue, digitValue, code))
		}
	}

	return map[string]interface{}{
		"input_original":     trackingID,
		"ternary_value":      ternary,
		"decimal_equivalent": trackingID,
		"encoded_length":     len(encodedChars),
		"original_message":   originalMessage,
		"char_codes":         charCodes,
		"char_mapping":       charMapping,
	}
}

// DecodeTrackingUseCase decodifica mensagem e extrai tracking ID
type DecodeTrackingUseCase struct {
	encoder *tracking.TernaryEncoder
	logger  *zap.Logger
}

// NewDecodeTrackingUseCase cria novo use case de decode
func NewDecodeTrackingUseCase(logger *zap.Logger) *DecodeTrackingUseCase {
	return &DecodeTrackingUseCase{
		encoder: tracking.NewTernaryEncoder(),
		logger:  logger,
	}
}

// Execute executa a decodificação
func (uc *DecodeTrackingUseCase) Execute(req DecodeTrackingRequest) (*DecodeTrackingResponse, error) {
	// Verificar comprimento mínimo
	if len([]rune(req.Message)) < 8 {
		return &DecodeTrackingResponse{
			Success:         false,
			Confidence:      "none",
			CleanMessage:    req.Message,
			OriginalMessage: req.Message,
			Error:           "mensagem muito curta - deve conter pelo menos 8 caracteres",
		}, nil
	}

	// Tentar decodificar
	trackingIDPtr, cleanMessage, err := uc.encoder.DecodeMessage(req.Message)
	if err != nil {
		// Não conseguiu decodificar, mas não é erro fatal
		return &DecodeTrackingResponse{
			Success:         false,
			Confidence:      "none",
			CleanMessage:    req.Message,
			OriginalMessage: req.Message,
			Error:           err.Error(),
		}, nil
	}

	if trackingIDPtr == nil {
		return &DecodeTrackingResponse{
			Success:         false,
			Confidence:      "low",
			CleanMessage:    cleanMessage,
			OriginalMessage: req.Message,
			Error:           "não foi possível decodificar código invisível",
		}, nil
	}

	trackingID := *trackingIDPtr

	// Obter análise detalhada
	analysis := uc.encoder.GetAnalysis(req.Message)

	// Reconverter para ternário para validação
	ternary, err := uc.encoder.DecimalToTernary(trackingID)
	if err != nil {
		return &DecodeTrackingResponse{
			Success:         false,
			Confidence:      "low",
			CleanMessage:    cleanMessage,
			OriginalMessage: req.Message,
			Error:           fmt.Sprintf("tracking ID inválido: %v", err),
		}, nil
	}

	response := &DecodeTrackingResponse{
		Success:         true,
		DecodedTernary:  ternary,
		DecodedDecimal:  trackingID,
		Confidence:      "high",
		Analysis:        analysis,
		CleanMessage:    cleanMessage,
		OriginalMessage: req.Message,
	}

	uc.logger.Info("tracking decoded successfully",
		zap.Int64("tracking_id", trackingID),
		zap.String("ternary", ternary),
		zap.String("clean_message", cleanMessage))

	return response, nil
}

// CheckForInvisibleCode verifica se mensagem contém código invisível
func (uc *DecodeTrackingUseCase) CheckForInvisibleCode(message string) bool {
	return uc.encoder.HasInvisibleCode(message)
}
