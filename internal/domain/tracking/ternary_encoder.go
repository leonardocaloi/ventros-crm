package tracking

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// TernaryEncoder implementa codificação ternária com caracteres invisíveis para WhatsApp
type TernaryEncoder struct {
	// 3 caracteres invisíveis seguros para representar 0, 1, 2
	safeChars []rune
}

// NewTernaryEncoder cria nova instância do encoder
func NewTernaryEncoder() *TernaryEncoder {
	return &TernaryEncoder{
		safeChars: []rune{
			'\u200B', // Zero Width Space (ZWSP) = 0
			'\u2060', // Word Joiner (WJ) = 1
			'\uFEFF', // Zero Width No-Break Space (ZWNBSP) = 2
		},
	}
}

// EncodeTernary codifica string ternária (0,1,2) em caracteres invisíveis
func (e *TernaryEncoder) EncodeTernary(ternaryString string) (string, error) {
	// Validar entrada: deve ser string com dígitos 0-2 e 1-7 caracteres
	matched, _ := regexp.MatchString(`^[0-2]{1,7}$`, ternaryString)
	if !matched {
		return "", fmt.Errorf("entrada deve ser string ternária (0,1,2) com 1-7 dígitos")
	}

	// Garantir exatamente 7 dígitos com padding de zeros à esquerda
	paddedTernary := fmt.Sprintf("%07s", ternaryString)
	if len(paddedTernary) > 7 {
		paddedTernary = paddedTernary[:7]
	}

	var result strings.Builder
	for _, digit := range paddedTernary {
		index := digit - '0'
		if index < 0 || index > 2 {
			return "", fmt.Errorf("dígito ternário inválido: %c", digit)
		}
		result.WriteRune(e.safeChars[index])
	}

	return result.String(), nil
}

// DecodeTernary decodifica caracteres invisíveis de volta para ternário
func (e *TernaryEncoder) DecodeTernary(text string) *string {
	if len(text) < 8 {
		return nil
	}

	// O código invisível está entre o 1º e 2º caracteres (posições 1-7)
	runes := []rune(text)
	if len(runes) < 8 {
		return nil
	}

	encodedSection := runes[1:8]

	var ternaryResult strings.Builder
	foundChars := 0

	// Processar cada caractere
	for i := 0; i < len(encodedSection) && foundChars < 7; i++ {
		char := encodedSection[i]

		// Tentar decodificar diretamente
		digitValue := e.findSafeCharIndex(char)

		if digitValue != -1 {
			// Caractere preservado corretamente
			ternaryResult.WriteString(strconv.Itoa(digitValue))
			foundChars++
		} else {
			// Caractere foi corrompido pelo WhatsApp
			// Aplicar mapeamento de recuperação
			digitValue = e.recoverCorruptedChar(char)
			if digitValue != -1 {
				ternaryResult.WriteString(strconv.Itoa(digitValue))
				foundChars++
			}
		}
	}

	// Se não conseguiu recuperar 7 dígitos, usar fallback
	if foundChars < 7 {
		result := e.fallbackRecovery(text, ternaryResult.String())
		return result
	}

	result := ternaryResult.String()
	return &result
}

// findSafeCharIndex encontra índice do caractere nos safe chars
func (e *TernaryEncoder) findSafeCharIndex(char rune) int {
	for i, safeChar := range e.safeChars {
		if char == safeChar {
			return i
		}
	}
	return -1
}

// recoverCorruptedChar mapeia caracteres corrompidos de volta para valores ternários
func (e *TernaryEncoder) recoverCorruptedChar(char rune) int {
	charCode := int(char)

	// ZWSP (8203) às vezes vira espaço normal (32) ou NBSP (160)
	if charCode == 32 || charCode == 160 {
		return 0
	}

	// Word Joiner (8288) pode virar outros joiners
	if charCode == 8204 || charCode == 8205 {
		return 1
	}

	// ZWNBSP (65279) geralmente preservado, mas pode virar BOM mark
	if charCode == 65279 || charCode == 8206 || charCode == 8207 {
		return 2
	}

	// Caracteres Hangul/Braille
	if charCode == 12644 { // Hangul Filler
		return 2 // Geralmente vem de ZWNBSP
	}

	if charCode == 10240 || charCode == 8192 { // Braille ou En Quad
		return 1 // Geralmente vem de WJ
	}

	// Heurística: caracteres com código alto tendem a ser 2
	if charCode > 10000 {
		return 2
	} else if charCode > 8000 {
		return 1
	}

	// Default para 0
	return 0
}

// fallbackRecovery recuperação usando padrões conhecidos
func (e *TernaryEncoder) fallbackRecovery(text string, partialTernary string) *string {
	// Se temos dígitos parciais, completar com zeros
	if len(partialTernary) > 0 {
		padded := fmt.Sprintf("%-7s", partialTernary)
		padded = strings.ReplaceAll(padded, " ", "0")
		return &padded
	}

	return nil
}

// TernaryToDecimal converte ternário para decimal
func (e *TernaryEncoder) TernaryToDecimal(ternaryString string) (int64, error) {
	if ternaryString == "" {
		return 0, fmt.Errorf("string ternária vazia")
	}

	var decimal int64 = 0
	for _, digit := range ternaryString {
		digitValue := int64(digit - '0')
		if digitValue < 0 || digitValue > 2 {
			return 0, fmt.Errorf("dígito ternário inválido: %c", digit)
		}
		decimal = decimal*3 + digitValue
	}
	return decimal, nil
}

// DecimalToTernary converte decimal para ternário
func (e *TernaryEncoder) DecimalToTernary(decimal int64) (string, error) {
	maxValue := int64(math.Pow(3, 7) - 1) // 2187 - 1 = 2186
	if decimal > maxValue {
		return "", fmt.Errorf("número muito grande! Máximo é %d (2222222 em ternário)", maxValue)
	}

	if decimal == 0 {
		return "0000000", nil
	}

	var ternary strings.Builder
	num := decimal

	for num > 0 {
		digit := num % 3
		ternary.WriteString(strconv.FormatInt(digit, 10))
		num = num / 3
	}

	// Reverter e fazer padding
	ternaryStr := reverseString(ternary.String())
	return fmt.Sprintf("%07s", ternaryStr), nil
}

// EncodeMessage insere código invisível após o primeiro caractere da mensagem
func (e *TernaryEncoder) EncodeMessage(message string, trackingID int64) (string, error) {
	// Converter ID para ternário
	ternary, err := e.DecimalToTernary(trackingID)
	if err != nil {
		return "", err
	}

	// Codificar em caracteres invisíveis
	encodedChars, err := e.EncodeTernary(ternary)
	if err != nil {
		return "", err
	}

	// Inserir código invisível ENTRE o 1º e 2º caracteres da mensagem
	runes := []rune(message)
	if len(runes) >= 1 {
		firstChar := string(runes[0])
		restOfMessage := ""
		if len(runes) > 1 {
			restOfMessage = string(runes[1:])
		}
		return firstChar + encodedChars + restOfMessage, nil
	}

	// Se mensagem vazia, apenas o código
	return encodedChars, nil
}

// DecodeMessage extrai código invisível e retorna ID e mensagem limpa
func (e *TernaryEncoder) DecodeMessage(message string) (*int64, string, error) {
	// Verificar comprimento mínimo
	if len([]rune(message)) < 8 {
		return nil, message, fmt.Errorf("mensagem muito curta para conter código")
	}

	// Decodificar ternário
	ternaryPtr := e.DecodeTernary(message)
	if ternaryPtr == nil {
		return nil, message, fmt.Errorf("não foi possível decodificar código ternário")
	}

	ternary := *ternaryPtr

	// Converter para decimal
	decimal, err := e.TernaryToDecimal(ternary)
	if err != nil {
		return nil, message, err
	}

	// Limpar mensagem (remover código invisível das posições 1-7)
	runes := []rune(message)
	cleanMessage := string(runes[0])
	if len(runes) > 8 {
		cleanMessage += string(runes[8:])
	}

	return &decimal, cleanMessage, nil
}

// HasInvisibleCode verifica se mensagem contém código invisível
func (e *TernaryEncoder) HasInvisibleCode(message string) bool {
	runes := []rune(message)
	if len(runes) < 8 {
		return false
	}

	// Verificar se há pelo menos 3 caracteres invisíveis nas posições 1-7
	invisibleCount := 0
	for i := 1; i <= 7 && i < len(runes); i++ {
		char := runes[i]
		if e.findSafeCharIndex(char) != -1 || e.isInvisibleChar(char) {
			invisibleCount++
		}
	}

	return invisibleCount >= 3
}

// isInvisibleChar verifica se caractere é invisível (incluindo corrompidos)
func (e *TernaryEncoder) isInvisibleChar(char rune) bool {
	// Caracteres invisíveis originais
	if e.findSafeCharIndex(char) != -1 {
		return true
	}

	// Caracteres que podem ser resultado de corrupção
	charCode := int(char)
	invisibleCodes := []int{
		32, 160, // Espaços
		8203, 8204, 8205, // Zero-width chars
		8206, 8207, // Bi-directional marks
		8288,        // Word joiner
		65279,       // Zero-width no-break
		12644,       // Hangul filler
		10240, 8192, // Braille/En Quad
	}

	for _, code := range invisibleCodes {
		if charCode == code {
			return true
		}
	}

	return false
}

// GetAnalysis retorna análise detalhada do código na mensagem
func (e *TernaryEncoder) GetAnalysis(message string) map[string]interface{} {
	runes := []rune(message)
	if len(runes) < 8 {
		return map[string]interface{}{
			"error": "mensagem muito curta",
		}
	}

	firstChar := string(runes[0])
	extractedChars := string(runes[1:8])

	var charCodes []int
	var charAnalysis []string

	for _, char := range runes[1:8] {
		code := int(char)
		charCodes = append(charCodes, code)

		safeIndex := e.findSafeCharIndex(char)
		if safeIndex != -1 {
			charAnalysis = append(charAnalysis, fmt.Sprintf("PRESERVED: SAFE_%d (U+%04X)", safeIndex, code))
		} else {
			recovered := e.recoverCorruptedChar(char)
			charAnalysis = append(charAnalysis, fmt.Sprintf("CORRUPTED: U+%04X → Recovered as %d", code, recovered))
		}
	}

	remainingMessage := ""
	if len(runes) > 8 {
		end := 28
		if len(runes) < end {
			end = len(runes)
		}
		remainingMessage = string(runes[8:end])
		if len(runes) > 28 {
			remainingMessage += "..."
		}
	}

	ternaryPtr := e.DecodeTernary(message)
	var decodedTernary string
	var decodedDecimal int64

	if ternaryPtr != nil {
		decodedTernary = *ternaryPtr
		decodedDecimal, _ = e.TernaryToDecimal(decodedTernary)
	}

	return map[string]interface{}{
		"first_char":        firstChar,
		"extracted_chars":   extractedChars,
		"char_codes":        charCodes,
		"char_analysis":     charAnalysis,
		"decoded_ternary":   decodedTernary,
		"decoded_decimal":   decodedDecimal,
		"remaining_message": remainingMessage,
	}
}

// reverseString inverte uma string
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
