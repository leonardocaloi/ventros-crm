package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/caloi/ventros-crm/infrastructure/ai"
	"github.com/caloi/ventros-crm/internal/domain/crm/contact"
	"github.com/caloi/ventros-crm/internal/domain/crm/message_enrichment"
	"github.com/caloi/ventros-crm/internal/domain/crm/pipeline"
)

// LeadQualificationConsumer processa eventos de ProfilePictureReceived
// Arquitetura: EVENT-DRIVEN (sem polling)
// Fluxo:
// 1. Contact recebe foto de perfil → evento ProfilePictureReceived disparado
// 2. Este consumer recebe o evento via RabbitMQ
// 3. Verifica se contato está em pipeline com qualificação ativada
// 4. Dispara análise com Vertex AI Vision usando prompt customizado
// 5. Calcula score (0-10) baseado nas respostas
// 6. Salva no metadata do contato
// 7. Dispara evento LeadQualified
type LeadQualificationConsumer struct {
	logger         *zap.Logger
	eventBus       *DomainEventBus
	contactRepo    contact.Repository
	pipelineRepo   pipeline.Repository
	enrichmentRepo message_enrichment.Repository
	visionProvider *ai.VertexVisionProvider
}

// NewLeadQualificationConsumer cria novo consumer
func NewLeadQualificationConsumer(
	logger *zap.Logger,
	eventBus *DomainEventBus,
	contactRepo contact.Repository,
	pipelineRepo pipeline.Repository,
	enrichmentRepo message_enrichment.Repository,
	visionProvider *ai.VertexVisionProvider,
) *LeadQualificationConsumer {
	return &LeadQualificationConsumer{
		logger:         logger,
		eventBus:       eventBus,
		contactRepo:    contactRepo,
		pipelineRepo:   pipelineRepo,
		enrichmentRepo: enrichmentRepo,
		visionProvider: visionProvider,
	}
}

// Start registra consumer no event bus
// TODO: Implementar Subscribe method no DomainEventBus
func (c *LeadQualificationConsumer) Start(ctx context.Context) error {
	c.logger.Info("Starting LeadQualificationConsumer (event-driven)")

	// TODO: Registrar handler para ProfilePictureReceived
	// return c.eventBus.Subscribe(ctx, "contact.profile_picture_received", c.handleProfilePictureReceived)
	c.logger.Warn("LeadQualificationConsumer not fully implemented - Subscribe method pending")
	return nil
}

// handleProfilePictureReceived processa evento quando contato recebe foto
func (c *LeadQualificationConsumer) handleProfilePictureReceived(ctx context.Context, eventData map[string]interface{}) error {
	c.logger.Info("Processing ProfilePictureReceived event", zap.Any("event_data", eventData))

	// Parse event
	event, err := c.parseEvent(eventData)
	if err != nil {
		return fmt.Errorf("failed to parse event: %w", err)
	}

	// 1. Buscar contato
	cont, err := c.contactRepo.FindByID(ctx, event.ContactID)
	if err != nil {
		return fmt.Errorf("failed to find contact: %w", err)
	}

	// 2. Verificar se contato está em pipeline
	if event.PipelineID == nil {
		c.logger.Debug("Contact not in pipeline, skipping qualification",
			zap.String("contact_id", event.ContactID.String()))
		return nil
	}

	// 3. Buscar pipeline
	pip, err := c.pipelineRepo.FindPipelineByID(ctx, *event.PipelineID)
	if err != nil {
		return fmt.Errorf("failed to find pipeline: %w", err)
	}

	// 4. Verificar se pipeline tem qualificação ativada
	if !pip.HasLeadQualification() {
		c.logger.Debug("Pipeline does not have lead qualification enabled",
			zap.String("pipeline_id", pip.ID().String()))
		return nil
	}

	config := pip.LeadQualificationConfig()

	// 5. WARNING: Verificar se tem foto de perfil
	hasProfilePhoto := event.ProfilePictureURL != ""
	if !hasProfilePhoto {
		c.logger.Warn("Contact has no profile photo - qualification may be limited",
			zap.String("contact_id", event.ContactID.String()))
		// Continuar mesmo sem foto - pode usar foto genérica ou dados limitados
	}

	// 6. Processar com Vision AI
	score, err := c.qualifyLead(ctx, config, event.ProfilePictureURL, hasProfilePhoto)
	if err != nil {
		c.logger.Error("Failed to qualify lead",
			zap.Error(err),
			zap.String("contact_id", event.ContactID.String()))
		return err
	}

	// 7. Salvar score no metadata do contato
	if err := c.saveQualificationScore(ctx, cont, score); err != nil {
		return fmt.Errorf("failed to save qualification score: %w", err)
	}

	// 8. Disparar evento LeadQualified
	// TODO: Fix event publishing - DomainEventBus.Publish signature
	/*
		qualifiedEvent := pipeline.LeadQualifiedEvent{
			ContactID:   event.ContactID,
			PipelineID:  *event.PipelineID,
			Score:       score.Score(),
			Qualified:   score.IsQualified(),
			Answers:     score.Answers(),
			Confidence:  score.Confidence(),
			QualifiedAt: time.Now(),
		}

		if err := c.eventBus.Publish(ctx, &qualifiedEvent); err != nil {
			c.logger.Error("Failed to publish LeadQualified event",
				zap.Error(err),
				zap.String("contact_id", event.ContactID.String()))
		}
	*/

	c.logger.Info("Lead qualification completed",
		zap.String("contact_id", event.ContactID.String()),
		zap.Int("score", score.Score()),
		zap.Bool("qualified", score.IsQualified()),
		zap.String("confidence", score.Confidence()))

	return nil
}

// qualifyLead processa foto com Vertex AI Vision e calcula score
func (c *LeadQualificationConsumer) qualifyLead(
	ctx context.Context,
	config *pipeline.LeadQualificationConfig,
	profilePictureURL string,
	hasProfilePhoto bool,
) (*pipeline.LeadQualificationScore, error) {

	// 1. Gerar prompt customizado baseado nas perguntas configuradas
	aiPrompt := config.GeneratePrompt()

	c.logger.Debug("Generated AI prompt for lead qualification",
		zap.String("prompt", aiPrompt))

	// 2. Contexto: "profile_picture" para foto de perfil
	enrichmentContentType := message_enrichment.EnrichmentTypeImage
	contextStr := "profile_picture"

	// 3. Processar com Vertex AI Vision
	result, err := c.visionProvider.Process(ctx, profilePictureURL, enrichmentContentType, &contextStr)
	if err != nil {
		return nil, fmt.Errorf("failed to process image with Vision AI: %w", err)
	}

	c.logger.Debug("Vision AI processing completed",
		zap.String("extracted_text", result.ExtractedText),
		zap.Any("metadata", result.Metadata))

	// 4. Parse respostas da IA (esperamos JSON)
	var aiAnswers map[string]string
	if err := json.Unmarshal([]byte(result.ExtractedText), &aiAnswers); err != nil {
		// Se não for JSON válido, criar respostas "indefinido"
		c.logger.Warn("Failed to parse AI answers as JSON, using fallback",
			zap.Error(err),
			zap.String("extracted_text", result.ExtractedText))

		aiAnswers = make(map[string]string)
		for _, question := range config.Questions() {
			aiAnswers[question.Key()] = "indefinido"
		}
	}

	// 5. Calcular score baseado nas respostas
	score, err := pipeline.NewLeadQualificationScore(config, aiAnswers, hasProfilePhoto)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate qualification score: %w", err)
	}

	return score, nil
}

// saveQualificationScore salva score no metadata do contato
// TODO: Implement Contact.SetCustomField or use alternative storage
func (c *LeadQualificationConsumer) saveQualificationScore(
	ctx context.Context,
	cont *contact.Contact,
	score *pipeline.LeadQualificationScore,
) error {

	// TODO: Serializar score para JSON e salvar
	// Por enquanto, apenas log
	c.logger.Info("Would save qualification score",
		zap.String("contact_id", cont.ID().String()),
		zap.Int("score", score.Score()),
		zap.Bool("qualified", score.IsQualified()),
		zap.String("confidence", score.Confidence()))

	/*
		// Serializar score para JSON
		scoreJSON, err := score.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to serialize score: %w", err)
		}

		// Salvar no custom field (ou metadata interno)
		// Usando custom field "lead_qualification_score"
		cont.SetCustomField("lead_qualification_score", string(scoreJSON))

		// Também salvar score simples para facilitar queries/filters
		cont.SetCustomField("lead_score", fmt.Sprintf("%d", score.Score()))
		cont.SetCustomField("lead_qualified", fmt.Sprintf("%t", score.IsQualified()))
		cont.SetCustomField("lead_confidence", score.Confidence())

		// Salvar contato
		if err := c.contactRepo.Save(ctx, cont); err != nil {
			return fmt.Errorf("failed to save contact: %w", err)
		}
	*/

	return nil
}

// parseEvent converte map para ProfilePictureReceivedEvent
func (c *LeadQualificationConsumer) parseEvent(eventData map[string]interface{}) (*pipeline.ProfilePictureReceivedEvent, error) {
	// Marshal para JSON e unmarshal para struct
	jsonData, err := json.Marshal(eventData)
	if err != nil {
		return nil, err
	}

	var event pipeline.ProfilePictureReceivedEvent
	if err := json.Unmarshal(jsonData, &event); err != nil {
		return nil, err
	}

	return &event, nil
}
