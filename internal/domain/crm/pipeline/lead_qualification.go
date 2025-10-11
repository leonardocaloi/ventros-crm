package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
)

// LeadQualificationConfig configura a qualificação automática de leads por imagem de perfil
// O cliente configura perguntas que a IA vai responder analisando a foto de perfil
type LeadQualificationConfig struct {
	enabled   bool
	questions []QualificationQuestion
	minScore  int // Score mínimo para considerar lead qualificado (0-10)
}

// QualificationQuestion representa uma pergunta de qualificação
type QualificationQuestion struct {
	key         string   // Chave única (ex: "product_type", "ticket_size")
	label       string   // Label para o cliente (ex: "Tipo de Produto")
	description string   // Descrição/instrução para a IA
	options     []string // Opções possíveis (se aplicável)
	weight      int      // Peso na pontuação final (1-10)
}

// NewLeadQualificationConfig cria uma nova config de qualificação
func NewLeadQualificationConfig() *LeadQualificationConfig {
	return &LeadQualificationConfig{
		enabled:   false,
		questions: []QualificationQuestion{},
		minScore:  5, // Default: 5/10
	}
}

// NewLeadQualificationConfigWithDefaults cria config com perguntas padrão
// Perguntas focadas em vendas B2C (cocoara = fábrica de cobertores)
func NewLeadQualificationConfigWithDefaults() *LeadQualificationConfig {
	config := NewLeadQualificationConfig()

	// Máximo 5 perguntas (como solicitado)
	config.questions = []QualificationQuestion{
		{
			key:         "product_interest",
			label:       "Produto de Interesse",
			description: "Analisando a foto de perfil, qual tipo de produto essa pessoa provavelmente compraria? Considere aparência, ambiente visível, objetos no fundo.",
			options:     []string{"cobertores_premium", "cobertores_basicos", "enxovais_luxo", "enxovais_simples", "indefinido"},
			weight:      8, // Alto peso - define persona
		},
		{
			key:         "ticket_size",
			label:       "Ticket Médio Estimado",
			description: "Baseado em sinais visuais (qualidade da foto, ambiente, vestimenta, objetos visíveis), estime o poder aquisitivo.",
			options:     []string{"alto", "medio", "baixo", "indefinido"},
			weight:      10, // Peso máximo - crucial para vendas
		},
		{
			key:         "purchase_intent",
			label:       "Intenção de Compra",
			description: "A foto sugere alguém buscando produtos (ex: foto em loja, segurando produtos, ambiente comercial) ou é apenas social?",
			options:     []string{"comercial", "pessoal", "indefinido"},
			weight:      6,
		},
		{
			key:         "persona",
			label:       "Persona Identificada",
			description: "Qual persona melhor se encaixa? Mãe/pai de família, jovem casal, pessoa idosa, revendedor/lojista?",
			options:     []string{"familia", "casal_jovem", "senior", "revendedor", "indefinido"},
			weight:      7,
		},
		{
			key:         "visual_quality",
			label:       "Qualidade Visual do Perfil",
			description: "A foto é profissional, casual ou genérica (logo, imagem aleatória)? Fotos profissionais indicam perfil business.",
			options:     []string{"profissional", "casual", "generica", "sem_foto"},
			weight:      5, // Menor peso - apenas indicador
		},
	}

	return config
}

// Enable ativa a qualificação automática
func (c *LeadQualificationConfig) Enable() {
	c.enabled = true
}

// Disable desativa a qualificação automática
func (c *LeadQualificationConfig) Disable() {
	c.enabled = false
}

// IsEnabled retorna se está ativado
func (c *LeadQualificationConfig) IsEnabled() bool {
	return c.enabled
}

// SetMinScore define score mínimo (0-10)
func (c *LeadQualificationConfig) SetMinScore(score int) error {
	if score < 0 || score > 10 {
		return errors.New("min score must be between 0 and 10")
	}
	c.minScore = score
	return nil
}

// MinScore retorna score mínimo
func (c *LeadQualificationConfig) MinScore() int {
	return c.minScore
}

// Questions retorna as perguntas configuradas
func (c *LeadQualificationConfig) Questions() []QualificationQuestion {
	return append([]QualificationQuestion{}, c.questions...)
}

// AddQuestion adiciona uma pergunta customizada
func (c *LeadQualificationConfig) AddQuestion(question QualificationQuestion) error {
	if question.key == "" {
		return errors.New("question key cannot be empty")
	}
	if question.label == "" {
		return errors.New("question label cannot be empty")
	}
	if question.description == "" {
		return errors.New("question description cannot be empty")
	}
	if question.weight < 1 || question.weight > 10 {
		return errors.New("question weight must be between 1 and 10")
	}

	// Verificar duplicata
	for _, q := range c.questions {
		if q.key == question.key {
			return fmt.Errorf("question with key %s already exists", question.key)
		}
	}

	// Máximo 5 perguntas
	if len(c.questions) >= 5 {
		return errors.New("maximum 5 questions allowed")
	}

	c.questions = append(c.questions, question)
	return nil
}

// RemoveQuestion remove uma pergunta
func (c *LeadQualificationConfig) RemoveQuestion(key string) error {
	for i, q := range c.questions {
		if q.key == key {
			c.questions = append(c.questions[:i], c.questions[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("question %s not found", key)
}

// GeneratePrompt gera o prompt completo para a IA
// A IA vai receber a foto de perfil + este prompt
func (c *LeadQualificationConfig) GeneratePrompt() string {
	if !c.enabled || len(c.questions) == 0 {
		return ""
	}

	prompt := `Você é um especialista em qualificação de leads para vendas.
Analise a foto de perfil do contato e responda as perguntas abaixo de forma objetiva.

IMPORTANTE:
- Se a foto não tem informação suficiente, responda "indefinido"
- Seja conservador nas estimativas
- Foque em sinais visuais concretos
- Retorne APENAS JSON válido, sem texto adicional

Perguntas:

`

	for _, q := range c.questions {
		prompt += fmt.Sprintf("- %s: %s\n", q.label, q.description)
		if len(q.options) > 0 {
			prompt += fmt.Sprintf("  Opções: %v\n", q.options)
		}
		prompt += "\n"
	}

	prompt += `
Formato de resposta (JSON):
{
`
	for i, q := range c.questions {
		prompt += fmt.Sprintf(`  "%s": "valor"`, q.key)
		if i < len(c.questions)-1 {
			prompt += ","
		}
		prompt += "\n"
	}
	prompt += `}`

	return prompt
}

// GenerateSimpleScorePrompt gera um prompt simples para scoring 0-10
// Inspirado no exemplo n8n: "Avalie a foto de perfil de 0 a 10 para qualificação de clientes"
func (c *LeadQualificationConfig) GenerateSimpleScorePrompt() string {
	prompt := `Avalie esta foto de perfil de 0 a 10 para qualificação de clientes potenciais para compras de produtos de alto valor.

Critérios de avaliação:
- Aparência profissional e cuidado pessoal (indica poder aquisitivo)
- Ambiente visível (casa/escritório bem decorado, objetos de qualidade)
- Qualidade da foto (profissional vs casual vs genérica)
- Sinais de estilo de vida (roupas, acessórios, objetos de marca)

Escala:
- 0-3: Baixo potencial (foto genérica, sem informações, ambiente simples)
- 4-6: Médio potencial (foto casual, algumas informações úteis)
- 7-10: Alto potencial (foto profissional, sinais claros de poder aquisitivo)

Retorne APENAS um número de 0 a 10.`

	return prompt
}

// LeadQualificationScore representa o resultado da análise
type LeadQualificationScore struct {
	score           int               // Score final (0-10)
	answers         map[string]string // Respostas da IA para cada pergunta
	qualified       bool              // Se passou no score mínimo
	confidence      string            // high, medium, low (baseado em "indefinido")
	analysisDetails string            // Detalhes/justificativa da IA
	hasProfilePhoto bool              // Se tinha foto de perfil
	warningMessage  *string           // Warning se não tinha foto ou foto genérica
}

// NewLeadQualificationScore cria um novo score a partir das respostas da IA
func NewLeadQualificationScore(
	config *LeadQualificationConfig,
	aiAnswers map[string]string,
	hasProfilePhoto bool,
) (*LeadQualificationScore, error) {

	score := &LeadQualificationScore{
		answers:         aiAnswers,
		hasProfilePhoto: hasProfilePhoto,
	}

	// Warning se não tem foto
	if !hasProfilePhoto {
		warning := "Contato sem foto de perfil - qualificação baseada apenas em dados limitados"
		score.warningMessage = &warning
	}

	// Calcular score ponderado
	totalScore := 0
	totalWeight := 0
	undefinedCount := 0

	for _, question := range config.questions {
		answer, exists := aiAnswers[question.key]
		if !exists {
			continue
		}

		// Contar respostas "indefinido"
		if answer == "indefinido" || answer == "sem_foto" || answer == "generica" {
			undefinedCount++
			continue
		}

		// Atribuir pontos baseado na resposta
		points := calculateQuestionScore(question, answer)
		totalScore += points * question.weight
		totalWeight += question.weight
	}

	// Score final (0-10)
	if totalWeight > 0 {
		score.score = (totalScore * 10) / (totalWeight * 10)
	} else {
		score.score = 0
	}

	// Confidence baseado em respostas indefinidas
	if undefinedCount == 0 {
		score.confidence = "high"
	} else if undefinedCount <= 2 {
		score.confidence = "medium"
	} else {
		score.confidence = "low"
	}

	// Verificar se qualificou
	score.qualified = score.score >= config.minScore

	return score, nil
}

// calculateQuestionScore calcula pontuação para uma resposta
func calculateQuestionScore(question QualificationQuestion, answer string) int {
	// Lógica de pontuação baseada na resposta
	switch question.key {
	case "ticket_size":
		switch answer {
		case "alto":
			return 10
		case "medio":
			return 7
		case "baixo":
			return 4
		default:
			return 0
		}

	case "purchase_intent":
		switch answer {
		case "comercial":
			return 10
		case "pessoal":
			return 5
		default:
			return 0
		}

	case "visual_quality":
		switch answer {
		case "profissional":
			return 10
		case "casual":
			return 7
		case "generica":
			return 3
		default:
			return 0
		}

	default:
		// Perguntas genéricas: se respondeu algo válido = 10
		if answer != "indefinido" && answer != "" {
			return 10
		}
		return 0
	}
}

// Getters
func (s *LeadQualificationScore) Score() int                 { return s.score }
func (s *LeadQualificationScore) Answers() map[string]string { return s.answers }
func (s *LeadQualificationScore) IsQualified() bool          { return s.qualified }
func (s *LeadQualificationScore) Confidence() string         { return s.confidence }
func (s *LeadQualificationScore) AnalysisDetails() string    { return s.analysisDetails }
func (s *LeadQualificationScore) HasProfilePhoto() bool      { return s.hasProfilePhoto }
func (s *LeadQualificationScore) WarningMessage() *string    { return s.warningMessage }

// ToJSON serializa para JSON (para armazenar em metadata)
func (s *LeadQualificationScore) ToJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"score":             s.score,
		"answers":           s.answers,
		"qualified":         s.qualified,
		"confidence":        s.confidence,
		"analysis_details":  s.analysisDetails,
		"has_profile_photo": s.hasProfilePhoto,
		"warning_message":   s.warningMessage,
	})
}

// Getters para QualificationQuestion
func (q QualificationQuestion) Key() string         { return q.key }
func (q QualificationQuestion) Label() string       { return q.label }
func (q QualificationQuestion) Description() string { return q.description }
func (q QualificationQuestion) Options() []string   { return q.options }
func (q QualificationQuestion) Weight() int         { return q.weight }
