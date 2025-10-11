package tracking

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type TernaryEncoder struct {
	safeChars []rune
}

func NewTernaryEncoder() *TernaryEncoder {
	return &TernaryEncoder{
		safeChars: []rune{
			'\u200B',
			'\u2060',
			'\uFEFF',
		},
	}
}

func (e *TernaryEncoder) EncodeTernary(ternaryString string) (string, error) {
	matched, _ := regexp.MatchString(`^[0-2]{1,7}$`, ternaryString)
	if !matched {
		return "", fmt.Errorf("entrada deve ser string ternária (0,1,2) com 1-7 dígitos")
	}

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

func (e *TernaryEncoder) DecodeTernary(text string) *string {
	if len(text) < 8 {
		return nil
	}

	runes := []rune(text)
	if len(runes) < 8 {
		return nil
	}

	encodedSection := runes[1:8]

	var ternaryResult strings.Builder
	foundChars := 0

	for i := 0; i < len(encodedSection) && foundChars < 7; i++ {
		char := encodedSection[i]

		digitValue := e.findSafeCharIndex(char)

		if digitValue != -1 {

			ternaryResult.WriteString(strconv.Itoa(digitValue))
			foundChars++
		} else {

			digitValue = e.recoverCorruptedChar(char)
			if digitValue != -1 {
				ternaryResult.WriteString(strconv.Itoa(digitValue))
				foundChars++
			}
		}
	}

	if foundChars < 7 {
		result := e.fallbackRecovery(text, ternaryResult.String())
		return result
	}

	result := ternaryResult.String()
	return &result
}

func (e *TernaryEncoder) findSafeCharIndex(char rune) int {
	for i, safeChar := range e.safeChars {
		if char == safeChar {
			return i
		}
	}
	return -1
}

func (e *TernaryEncoder) recoverCorruptedChar(char rune) int {
	charCode := int(char)

	if charCode == 32 || charCode == 160 {
		return 0
	}

	if charCode == 8204 || charCode == 8205 {
		return 1
	}

	if charCode == 65279 || charCode == 8206 || charCode == 8207 {
		return 2
	}

	if charCode == 12644 {
		return 2
	}

	if charCode == 10240 || charCode == 8192 {
		return 1
	}

	if charCode > 10000 {
		return 2
	} else if charCode > 8000 {
		return 1
	}

	return 0
}

func (e *TernaryEncoder) fallbackRecovery(text string, partialTernary string) *string {

	if len(partialTernary) > 0 {
		padded := fmt.Sprintf("%-7s", partialTernary)
		padded = strings.ReplaceAll(padded, " ", "0")
		return &padded
	}

	return nil
}

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

func (e *TernaryEncoder) DecimalToTernary(decimal int64) (string, error) {
	maxValue := int64(math.Pow(3, 7) - 1)
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

	ternaryStr := reverseString(ternary.String())
	return fmt.Sprintf("%07s", ternaryStr), nil
}

func (e *TernaryEncoder) EncodeMessage(message string, trackingID int64) (string, error) {

	ternary, err := e.DecimalToTernary(trackingID)
	if err != nil {
		return "", err
	}

	encodedChars, err := e.EncodeTernary(ternary)
	if err != nil {
		return "", err
	}

	runes := []rune(message)
	if len(runes) >= 1 {
		firstChar := string(runes[0])
		restOfMessage := ""
		if len(runes) > 1 {
			restOfMessage = string(runes[1:])
		}
		return firstChar + encodedChars + restOfMessage, nil
	}

	return encodedChars, nil
}

func (e *TernaryEncoder) DecodeMessage(message string) (*int64, string, error) {
	if len([]rune(message)) < 8 {
		return nil, message, fmt.Errorf("mensagem muito curta para conter código")
	}

	ternaryPtr := e.DecodeTernary(message)
	if ternaryPtr == nil {
		return nil, message, fmt.Errorf("não foi possível decodificar código ternário")
	}

	ternary := *ternaryPtr

	decimal, err := e.TernaryToDecimal(ternary)
	if err != nil {
		return nil, message, err
	}

	runes := []rune(message)
	cleanMessage := string(runes[0])
	if len(runes) > 8 {
		cleanMessage += string(runes[8:])
	}

	return &decimal, cleanMessage, nil
}

func (e *TernaryEncoder) HasInvisibleCode(message string) bool {
	runes := []rune(message)
	if len(runes) < 8 {
		return false
	}

	invisibleCount := 0
	for i := 1; i <= 7 && i < len(runes); i++ {
		char := runes[i]
		if e.findSafeCharIndex(char) != -1 || e.isInvisibleChar(char) {
			invisibleCount++
		}
	}

	return invisibleCount >= 3
}

func (e *TernaryEncoder) isInvisibleChar(char rune) bool {

	if e.findSafeCharIndex(char) != -1 {
		return true
	}

	charCode := int(char)
	invisibleCodes := []int{
		32, 160,
		8203, 8204, 8205,
		8206, 8207,
		8288,
		65279,
		12644,
		10240, 8192,
	}

	for _, code := range invisibleCodes {
		if charCode == code {
			return true
		}
	}

	return false
}

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

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
