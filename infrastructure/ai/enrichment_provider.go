package ai

import (
	"context"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/message_enrichment"
)

// EnrichmentResult representa o resultado do processamento de enriquecimento
type EnrichmentResult struct {
	ExtractedText  string                 // Texto extraído (transcrição, OCR, parsing)
	Metadata       map[string]interface{} // Metadados do provider
	ProcessingTime time.Duration          // Tempo de processamento
}

// EnrichmentProvider define a interface comum para todos os providers de enriquecimento
type EnrichmentProvider interface {
	// Process processa uma mídia e retorna o texto extraído + metadados
	// context é opcional e usado apenas por alguns providers (ex: Vision)
	Process(ctx context.Context, mediaURL string, contentType message_enrichment.EnrichmentContentType, processingContext *string) (*EnrichmentResult, error)

	// Name retorna o nome do provider (whisper, vision, etc)
	Name() string

	// SupportsContentType verifica se o provider suporta o tipo de conteúdo
	SupportsContentType(contentType message_enrichment.EnrichmentContentType) bool
}

// ProviderFactory cria providers baseado no tipo
type ProviderFactory struct {
	providers map[message_enrichment.EnrichmentProvider]EnrichmentProvider
}

// NewProviderFactory cria uma nova factory de providers
func NewProviderFactory() *ProviderFactory {
	factory := &ProviderFactory{
		providers: make(map[message_enrichment.EnrichmentProvider]EnrichmentProvider),
	}

	// Registrar providers disponíveis
	// Os providers serão registrados conforme forem implementados
	// Por enquanto, apenas estrutura

	return factory
}

// RegisterProvider registra um provider na factory
func (f *ProviderFactory) RegisterProvider(providerType message_enrichment.EnrichmentProvider, provider EnrichmentProvider) {
	f.providers[providerType] = provider
}

// GetProvider retorna o provider apropriado para o tipo
func (f *ProviderFactory) GetProvider(providerType message_enrichment.EnrichmentProvider) (EnrichmentProvider, error) {
	provider, exists := f.providers[providerType]
	if !exists {
		return nil, message_enrichment.ErrProviderUnavailable
	}
	return provider, nil
}

// HasProvider verifica se um provider está registrado
func (f *ProviderFactory) HasProvider(providerType message_enrichment.EnrichmentProvider) bool {
	_, exists := f.providers[providerType]
	return exists
}
