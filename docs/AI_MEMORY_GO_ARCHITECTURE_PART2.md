# 🧠 AI MEMORY GO ARCHITECTURE - PART 2

## 📚 TEMPORAL KNOWLEDGE GRAPH SERVICE

```go
package memory

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
)

// TemporalKnowledgeGraphService - Gerencia grafo de conhecimento temporal (Apache AGE)
type TemporalKnowledgeGraphService struct {
    ageClient *ApacheAGEClient
    repo      TemporalGraphRepository
}

// TemporalEdge - Aresta com validade temporal (bi-temporal model)
type TemporalEdge struct {
    ID             uuid.UUID              `json:"id"`
    FromNodeID     uuid.UUID              `json:"from_node_id"`
    FromNodeType   string                 `json:"from_node_type"`   // "contact", "session", "agent"
    ToNodeID       uuid.UUID              `json:"to_node_id"`
    ToNodeType     string                 `json:"to_node_type"`
    EdgeType       string                 `json:"edge_type"`        // "HAS_SESSION", "ASSIGNED_TO", etc

    // === BI-TEMPORAL MODEL (Zep/Graphiti pattern) ===
    ValidFrom      time.Time              `json:"valid_from"`       // Quando evento OCORREU
    ValidTo        *time.Time             `json:"valid_to"`         // Quando evento TERMINOU (NULL = ainda válido)
    TransactionAt  time.Time              `json:"transaction_at"`   // Quando foi INGERIDO no sistema

    Properties     map[string]interface{} `json:"properties"`
    TenantID       string                 `json:"tenant_id"`
}

// EdgeType constants
const (
    EdgeHasSession            = "HAS_SESSION"             // Contact → Session
    EdgeAssignedTo            = "ASSIGNED_TO"             // Session → Agent
    EdgeTransferredTo         = "TRANSFERRED_TO"          // Agent → Agent (chain)
    EdgeRepliedTo             = "REPLIED_TO"              // Message → Message (threading)
    EdgeMentions              = "MENTIONS"                // Message → Contact (social graph)
    EdgeDiscussedTopic        = "DISCUSSED_TOPIC"         // Session → Topic
    EdgeCameFromCampaign      = "CAME_FROM_CAMPAIGN"      // Contact → Campaign
    EdgeHasPlatform           = "HAS_PLATFORM"            // Campaign → Platform
    EdgeInPipeline            = "IN_PIPELINE"             // Contact → Pipeline
    EdgeHasBudgetConstraint   = "HAS_BUDGET_CONSTRAINT"   // Contact → Budget (memory fact)
    EdgeHasPreference         = "HAS_PREFERENCE"          // Contact → Preference
    EdgeHasGoal               = "HAS_GOAL"                // Contact → Goal
    EdgeHasObjection          = "HAS_OBJECTION"           // Session → Objection
)

// AddTemporalEdge - Adiciona edge com validade temporal
func (t *TemporalKnowledgeGraphService) AddTemporalEdge(
    ctx context.Context,
    edge TemporalEdge,
) error {
    // Validate
    if edge.FromNodeID == uuid.Nil || edge.ToNodeID == uuid.Nil {
        return fmt.Errorf("invalid node IDs")
    }

    now := time.Now()
    if edge.TransactionAt.IsZero() {
        edge.TransactionAt = now
    }
    if edge.ValidFrom.IsZero() {
        edge.ValidFrom = now
    }

    // Apache AGE CYPHER query
    query := `
    MATCH (from {id: $from_id, type: $from_type})
    MATCH (to {id: $to_id, type: $to_type})

    // Invalida edges anteriores do mesmo tipo (se aplicável)
    OPTIONAL MATCH (from)-[old_edge:` + edge.EdgeType + `]->(to)
    WHERE old_edge.valid_to IS NULL
      AND old_edge.from_node_id = $from_id
      AND old_edge.to_node_id = $to_id
    SET old_edge.valid_to = $valid_from

    // Cria nova edge com bi-temporal tracking
    CREATE (from)-[new_edge:` + edge.EdgeType + `]->(to)
    SET new_edge.id = $edge_id,
        new_edge.valid_from = $valid_from,
        new_edge.valid_to = NULL,
        new_edge.transaction_at = $transaction_at,
        new_edge.properties = $properties,
        new_edge.tenant_id = $tenant_id

    RETURN new_edge
    `

    params := map[string]interface{}{
        "from_id":        edge.FromNodeID.String(),
        "from_type":      edge.FromNodeType,
        "to_id":          edge.ToNodeID.String(),
        "to_type":        edge.ToNodeType,
        "edge_id":        edge.ID.String(),
        "valid_from":     edge.ValidFrom,
        "transaction_at": edge.TransactionAt,
        "properties":     edge.Properties,
        "tenant_id":      edge.TenantID,
    }

    return t.ageClient.Execute(ctx, query, params)
}

// QueryTemporalGraph - Query point-in-time (válido em determinado momento)
func (t *TemporalKnowledgeGraphService) QueryTemporalGraph(
    ctx context.Context,
    nodeID uuid.UUID,
    edgeType string,
    asOf *time.Time,  // NULL = agora
) ([]TemporalEdge, error) {
    if asOf == nil {
        now := time.Now()
        asOf = &now
    }

    query := `
    MATCH (node {id: $node_id})-[edge:` + edgeType + `]->(target)
    WHERE edge.valid_from <= $as_of
      AND (edge.valid_to IS NULL OR edge.valid_to > $as_of)
    RETURN edge, target
    ORDER BY edge.valid_from DESC
    `

    params := map[string]interface{}{
        "node_id": nodeID.String(),
        "as_of":   *asOf,
    }

    // TODO: Parse results from Apache AGE
    results := []TemporalEdge{}

    return results, t.ageClient.Query(ctx, query, params, &results)
}

// GetAgentTransferChain - Retorna chain completa de agent transfers
func (t *TemporalKnowledgeGraphService) GetAgentTransferChain(
    ctx context.Context,
    sessionID uuid.UUID,
) (*AgentTransferChain, error) {
    query := `
    MATCH path = (session:Session {id: $session_id})-[:ASSIGNED_TO*]->(agent:Agent)
    RETURN path
    ORDER BY length(path) DESC
    LIMIT 1
    `

    params := map[string]interface{}{
        "session_id": sessionID.String(),
    }

    // Parse path from graph
    chain := &AgentTransferChain{
        SessionID: sessionID,
        Agents:    []AgentNode{},
    }

    // TODO: Parse from Apache AGE result
    err := t.ageClient.Query(ctx, query, params, &chain)

    return chain, err
}

// GetCampaignAttribution - Retorna grafo de atribuição completo
func (t *TemporalKnowledgeGraphService) GetCampaignAttribution(
    ctx context.Context,
    contactID uuid.UUID,
) (*CampaignAttributionGraph, error) {
    query := `
    MATCH (contact:Contact {id: $contact_id})
          -[:CAME_FROM_CAMPAIGN]->(campaign:Campaign)
          -[:HAS_PLATFORM]->(platform:Platform)
    OPTIONAL MATCH (campaign)-[:HAS_AD]->(ad:Ad)
    RETURN contact, campaign, platform, ad
    `

    params := map[string]interface{}{
        "contact_id": contactID.String(),
    }

    attribution := &CampaignAttributionGraph{
        ContactID: contactID,
    }

    err := t.ageClient.Query(ctx, query, params, &attribution)

    return attribution, err
}

// GetSocialGraph - Retorna grafo de menções (quem menciona quem)
func (t *TemporalKnowledgeGraphService) GetSocialGraph(
    ctx context.Context,
    contactID uuid.UUID,
    depth int,
) (*SocialGraph, error) {
    query := fmt.Sprintf(`
    MATCH path = (contact:Contact {id: $contact_id})
                 -[:MENTIONS*1..%d]-(other:Contact)
    RETURN path, other
    `, depth)

    params := map[string]interface{}{
        "contact_id": contactID.String(),
    }

    socialGraph := &SocialGraph{
        CenterContactID: contactID,
        Connections:     []SocialConnection{},
    }

    err := t.ageClient.Query(ctx, query, params, &socialGraph)

    return socialGraph, err
}

// Supporting types
type AgentTransferChain struct {
    SessionID uuid.UUID   `json:"session_id"`
    Agents    []AgentNode `json:"agents"`
    Transfers int         `json:"transfers"`
}

type AgentNode struct {
    AgentID      uuid.UUID `json:"agent_id"`
    AgentName    string    `json:"agent_name"`
    AssignedAt   time.Time `json:"assigned_at"`
    TransferredAt *time.Time `json:"transferred_at"`
}

type CampaignAttributionGraph struct {
    ContactID    uuid.UUID `json:"contact_id"`
    Campaign     string    `json:"campaign"`
    Platform     string    `json:"platform"`
    AdID         *string   `json:"ad_id"`
    AdCreative   *string   `json:"ad_creative"`
    UTMSource    string    `json:"utm_source"`
    UTMMedium    string    `json:"utm_medium"`
    UTMCampaign  string    `json:"utm_campaign"`
}

type SocialGraph struct {
    CenterContactID uuid.UUID           `json:"center_contact_id"`
    Connections     []SocialConnection  `json:"connections"`
}

type SocialConnection struct {
    ContactID   uuid.UUID `json:"contact_id"`
    ContactName string    `json:"contact_name"`
    Mentions    int       `json:"mentions"`
    Depth       int       `json:"depth"`  // Grau de separação
}
```

---

## 🎯 AGENT REGISTRY & SEMANTIC ROUTING

```go
package memory

import (
    "context"
    "fmt"
    "math"
    "sort"

    "github.com/google/uuid"
)

// AgentRegistry - Gerencia registro e routing de agentes
type AgentRegistry struct {
    agentRepo          AgentRepository
    semanticRouter     *SemanticRouterService
    embeddingClient    *genai.Client
    routeEmbeddings    map[string][]float32  // Pre-computed embeddings de routes
}

// RegisterAgent - Registra novo agente AI
func (r *AgentRegistry) RegisterAgent(
    ctx context.Context,
    agent *Agent,
    aiMetadata *AIAgentMetadata,
) error {
    // Validate
    if agent.Type() != AgentTypeAI && agent.Type() != AgentTypeBot {
        return fmt.Errorf("only AI/Bot agents can be registered")
    }

    // Store in repository
    if err := r.agentRepo.Save(ctx, agent); err != nil {
        return fmt.Errorf("failed to save agent: %w", err)
    }

    // Pre-compute embeddings para routing rules (se existirem)
    if len(aiMetadata.RoutingRules) > 0 {
        for _, rule := range aiMetadata.RoutingRules {
            embedding, err := r.embeddingClient.Embed(ctx, rule.Condition)
            if err != nil {
                continue // Log erro mas não falha
            }
            r.routeEmbeddings[agent.ID().String()+":"+rule.Condition] = embedding
        }
    }

    return nil
}

// RouteToAgent - Semantic routing baseado em mensagem
func (r *AgentRegistry) RouteToAgent(
    ctx context.Context,
    message *Message,
    session *Session,
    tenantID string,
) (*Agent, error) {
    // 1. Detecta intent usando Semantic Router
    intent, err := r.semanticRouter.ClassifyIntent(ctx, *message.Text())
    if err != nil {
        return nil, fmt.Errorf("failed to classify intent: %w", err)
    }

    // 2. Busca agentes compatíveis com o intent
    candidates, err := r.findCandidateAgents(ctx, intent, session, tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to find candidate agents: %w", err)
    }

    if len(candidates) == 0 {
        // Fallback: agente default/operations
        return r.getDefaultAgent(ctx, tenantID)
    }

    // 3. Score e ranking de candidates
    scored := r.scoreAgents(candidates, message, session, intent)

    // 4. Retorna melhor match
    return scored[0].Agent, nil
}

// SemanticRouterService - Implementa Semantic Router (Aurelio Labs pattern)
type SemanticRouterService struct {
    embeddingClient *genai.Client
    routes          []SemanticRoute
    routeEmbeddings map[string][]float32
    threshold       float64  // Similarity threshold (ex: 0.75)
}

// SemanticRoute - Route com examples para matching
type SemanticRoute struct {
    Name        string                 `json:"name"`         // "churn_risk", "sales_inquiry"
    Category    AgentCategory          `json:"category"`     // Maps to agent category
    Utterances  []string               `json:"utterances"`   // Example phrases
    Priority    int                    `json:"priority"`     // Tie-breaking
    Metadata    map[string]interface{} `json:"metadata"`
}

// Semantic Routes predefinidas
var DefaultSemanticRoutes = []SemanticRoute{
    {
        Name:     "churn_risk",
        Category: CategoryRetentionChurn,
        Utterances: []string{
            "quero cancelar",
            "não quero mais",
            "vou desistir",
            "isso não está funcionando",
            "muito caro pra mim",
            "não vale a pena",
            "vou procurar outro",
        },
        Priority: 10, // ALTA prioridade
    },
    {
        Name:     "sales_inquiry",
        Category: CategorySalesProspecting,
        Utterances: []string{
            "quanto custa",
            "qual o preço",
            "tem desconto",
            "quero saber valores",
            "como funciona a cobrança",
            "aceita cartão",
        },
        Priority: 7,
    },
    {
        Name:     "technical_support",
        Category: CategorySupportTechnical,
        Utterances: []string{
            "não está funcionando",
            "deu erro",
            "bug",
            "problema técnico",
            "não consigo acessar",
            "tela em branco",
        },
        Priority: 9, // ALTA (problemas técnicos são urgentes)
    },
    {
        Name:     "billing_support",
        Category: CategorySupportBilling,
        Utterances: []string{
            "não recebi a fatura",
            "cobrança errada",
            "problema com pagamento",
            "cartão não passou",
            "reembolso",
        },
        Priority: 8,
    },
    {
        Name:     "objection_handling",
        Category: CategorySalesNegotiation,
        Utterances: []string{
            "muito caro",
            "não cabe no orçamento",
            "preciso pensar",
            "vou conversar com o time",
            "não é o momento certo",
        },
        Priority: 7,
    },
    {
        Name:     "feature_request",
        Category: CategoryOperationsFollowup,
        Utterances: []string{
            "seria bom se tivesse",
            "gostaria de sugerir",
            "falta essa funcionalidade",
            "quando vocês vão lançar",
        },
        Priority: 5,
    },
}

// NewSemanticRouterService - Cria novo router com routes pré-computadas
func NewSemanticRouterService(
    embeddingClient *genai.Client,
    routes []SemanticRoute,
    threshold float64,
) (*SemanticRouterService, error) {
    router := &SemanticRouterService{
        embeddingClient: embeddingClient,
        routes:          routes,
        routeEmbeddings: make(map[string][]float32),
        threshold:       threshold,
    }

    // Pre-compute embeddings dos utterances
    for _, route := range routes {
        for _, utterance := range route.Utterances {
            embedding, err := embeddingClient.Embed(context.Background(), utterance)
            if err != nil {
                continue // Log mas não falha
            }
            key := fmt.Sprintf("%s:%s", route.Name, utterance)
            router.routeEmbeddings[key] = embedding
        }
    }

    return router, nil
}

// ClassifyIntent - Classifica intent usando semantic similarity
func (s *SemanticRouterService) ClassifyIntent(
    ctx context.Context,
    text string,
) (*Intent, error) {
    // 1. Gera embedding do texto de entrada
    textEmbedding, err := s.embeddingClient.Embed(ctx, text)
    if err != nil {
        return nil, fmt.Errorf("failed to embed text: %w", err)
    }

    // 2. Compara com todos utterance embeddings
    type ScoredRoute struct {
        Route      SemanticRoute
        Similarity float64
        MatchedUtterance string
    }

    scored := []ScoredRoute{}

    for _, route := range s.routes {
        maxSimilarity := 0.0
        matchedUtterance := ""

        for _, utterance := range route.Utterances {
            key := fmt.Sprintf("%s:%s", route.Name, utterance)
            utteranceEmbedding, exists := s.routeEmbeddings[key]
            if !exists {
                continue
            }

            similarity := cosineSimilarity(textEmbedding, utteranceEmbedding)
            if similarity > maxSimilarity {
                maxSimilarity = similarity
                matchedUtterance = utterance
            }
        }

        if maxSimilarity >= s.threshold {
            scored = append(scored, ScoredRoute{
                Route:            route,
                Similarity:       maxSimilarity,
                MatchedUtterance: matchedUtterance,
            })
        }
    }

    if len(scored) == 0 {
        // No match above threshold
        return &Intent{
            Name:       "unknown",
            Category:   CategoryOperationsFollowup, // Fallback
            Confidence: 0.0,
        }, nil
    }

    // 3. Sort by similarity DESC, then priority DESC
    sort.Slice(scored, func(i, j int) bool {
        if math.Abs(scored[i].Similarity-scored[j].Similarity) < 0.01 {
            // Similaridade igual: usa priority
            return scored[i].Route.Priority > scored[j].Route.Priority
        }
        return scored[i].Similarity > scored[j].Similarity
    })

    // 4. Retorna best match
    best := scored[0]
    return &Intent{
        Name:             best.Route.Name,
        Category:         best.Route.Category,
        Confidence:       best.Similarity,
        MatchedUtterance: best.MatchedUtterance,
    }, nil
}

// Intent - Classificação de intent
type Intent struct {
    Name             string        `json:"name"`
    Category         AgentCategory `json:"category"`
    Confidence       float64       `json:"confidence"`
    MatchedUtterance string        `json:"matched_utterance"`
}

// findCandidateAgents - Busca agentes compatíveis com intent
func (r *AgentRegistry) findCandidateAgents(
    ctx context.Context,
    intent *Intent,
    session *Session,
    tenantID string,
) ([]*Agent, error) {
    // Busca agentes ativos da categoria
    agents, err := r.agentRepo.FindByCategory(ctx, intent.Category, tenantID)
    if err != nil {
        return nil, err
    }

    // Filtra por disponibilidade
    available := []*Agent{}
    for _, agent := range agents {
        if agent.IsActive() && agent.Status() == AgentStatusAvailable {
            // TODO: Check MaxConcurrentSessions
            available = append(available, agent)
        }
    }

    return available, nil
}

// scoreAgents - Pontua agentes baseado em múltiplos fatores
func (r *AgentRegistry) scoreAgents(
    agents []*Agent,
    message *Message,
    session *Session,
    intent *Intent,
) []ScoredAgent {
    scored := make([]ScoredAgent, len(agents))

    for i, agent := range agents {
        score := 0.0

        // 1. Intent match confidence (peso: 40%)
        score += intent.Confidence * 0.40

        // 2. Agent priority (peso: 20%)
        // TODO: Get from AIAgentMetadata
        priority := 5.0 // Default
        score += (priority / 10.0) * 0.20

        // 3. Session history (peso: 20%)
        // Se agente já atendeu este contato antes, +bonus
        if containsAgent(session.AgentIDs(), agent.ID()) {
            score += 0.20
        }

        // 4. Load balancing (peso: 20%)
        // TODO: Get current session count
        currentSessions := 0
        maxSessions := 10 // TODO: Get from AIAgentMetadata
        loadFactor := 1.0 - (float64(currentSessions) / float64(maxSessions))
        score += loadFactor * 0.20

        scored[i] = ScoredAgent{
            Agent: agent,
            Score: score,
        }
    }

    // Sort by score DESC
    sort.Slice(scored, func(i, j int) bool {
        return scored[i].Score > scored[j].Score
    })

    return scored
}

type ScoredAgent struct {
    Agent *Agent  `json:"agent"`
    Score float64 `json:"score"`
}

// getDefaultAgent - Retorna agente default/fallback
func (r *AgentRegistry) getDefaultAgent(
    ctx context.Context,
    tenantID string,
) (*Agent, error) {
    agents, err := r.agentRepo.FindByCategory(ctx, CategoryOperationsFollowup, tenantID)
    if err != nil {
        return nil, err
    }

    if len(agents) == 0 {
        return nil, fmt.Errorf("no default agent found")
    }

    return agents[0], nil
}

// Helper functions
func cosineSimilarity(a, b []float32) float64 {
    if len(a) != len(b) {
        return 0.0
    }

    var dotProduct, normA, normB float64
    for i := 0; i < len(a); i++ {
        dotProduct += float64(a[i]) * float64(b[i])
        normA += float64(a[i]) * float64(a[i])
        normB += float64(b[i]) * float64(b[i])
    }

    if normA == 0 || normB == 0 {
        return 0.0
    }

    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func containsAgent(agentIDs []uuid.UUID, agentID uuid.UUID) bool {
    for _, id := range agentIDs {
        if id == agentID {
            return true
        }
    }
    return false
}
```

---

## 🧬 MEMORY FACT SERVICE (Google Memory Bank pattern)

```go
package memory

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
)

// MemoryFactService - Gerencia facts com contradiction resolution
type MemoryFactService struct {
    factRepo    FactRepository
    llmClient   *genai.Client  // Para extraction e contradiction detection
}

// MemoryFact - Fact extraído de conversas
type MemoryFact struct {
    ID           uuid.UUID              `db:"id"`
    ContactID    uuid.UUID              `db:"contact_id"`
    TenantID     string                 `db:"tenant_id"`
    FactType     FactType               `db:"fact_type"`
    FactText     string                 `db:"fact_text"`         // Texto original
    FactValue    interface{}            `db:"fact_value"`        // Valor estruturado
    Confidence   float64                `db:"confidence"`        // 0.0-1.0
    ValidFrom    time.Time              `db:"valid_from"`
    ValidTo      *time.Time             `db:"valid_to"`          // NULL = current
    Supersedes   *uuid.UUID             `db:"supersedes"`        // FK to previous fact
    Source       string                 `db:"source"`            // "message", "note", "annotation"
    SourceID     *uuid.UUID             `db:"source_id"`         // ID do source (message_id, note_id)
    CreatedAt    time.Time              `db:"created_at"`
    Metadata     map[string]interface{} `db:"metadata"`          // JSONB
}

// FactType - Tipos de facts
type FactType string

const (
    FactTypeBudgetConstraint   FactType = "budget_constraint"
    FactTypePreference          FactType = "preference"
    FactTypeGoal                FactType = "goal"
    FactTypeObjection           FactType = "objection"
    FactTypePainPoint           FactType = "pain_point"
    FactTypeTechnicalIssue      FactType = "technical_issue"
    FactTypeEnvironmentInfo     FactType = "environment_info"
    FactTypeDecisionMaker       FactType = "decision_maker"
    FactTypeTimeline            FactType = "timeline"
    FactTypeCompetitor          FactType = "competitor"
)

// AddFact - Adiciona novo fact com contradiction detection
func (m *MemoryFactService) AddFact(
    ctx context.Context,
    contactID uuid.UUID,
    tenantID string,
    factText string,
    source string,
    sourceID *uuid.UUID,
) (*MemoryFact, error) {
    // 1. Extract structured information usando LLM
    extracted, err := m.extractStructuredFact(ctx, factText)
    if err != nil {
        return nil, fmt.Errorf("failed to extract fact: %w", err)
    }

    // 2. Busca facts existentes do mesmo tipo
    existingFacts, err := m.factRepo.FindActiveFactsByType(
        ctx,
        contactID,
        extracted.FactType,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to find existing facts: %w", err)
    }

    // 3. Contradiction detection
    for _, existing := range existingFacts {
        contradicts, resolution, err := m.detectContradiction(
            ctx,
            existing,
            extracted,
        )
        if err != nil {
            continue // Log mas não falha
        }

        if contradicts {
            if err := m.resolveContradiction(ctx, existing, extracted, resolution); err != nil {
                return nil, fmt.Errorf("failed to resolve contradiction: %w", err)
            }
        }
    }

    // 4. Persiste novo fact
    newFact := &MemoryFact{
        ID:         uuid.New(),
        ContactID:  contactID,
        TenantID:   tenantID,
        FactType:   extracted.FactType,
        FactText:   factText,
        FactValue:  extracted.FactValue,
        Confidence: extracted.Confidence,
        ValidFrom:  time.Now(),
        ValidTo:    nil,  // Current
        Source:     source,
        SourceID:   sourceID,
        CreatedAt:  time.Now(),
        Metadata:   extracted.Metadata,
    }

    if err := m.factRepo.Save(ctx, newFact); err != nil {
        return nil, fmt.Errorf("failed to save fact: %w", err)
    }

    return newFact, nil
}

// extractStructuredFact - Usa LLM para extrair fact estruturado
func (m *MemoryFactService) extractStructuredFact(
    ctx context.Context,
    factText string,
) (*ExtractedFact, error) {
    prompt := fmt.Sprintf(`
Extraia informações estruturadas deste fact:

"%s"

Retorne JSON no formato:
{
    "fact_type": "budget_constraint|preference|goal|objection|pain_point|technical_issue|environment_info|decision_maker|timeline|competitor",
    "fact_value": <value>,  // Valor estruturado (número, string, objeto)
    "confidence": <0-1>,     // Quão confiante você está na extração
    "metadata": {}           // Informações adicionais relevantes
}

Exemplos:
- "Meu orçamento é R$5000" → {"fact_type": "budget_constraint", "fact_value": 5000.0, "confidence": 0.95}
- "Prefiro ser chamado de João" → {"fact_type": "preference", "fact_value": "name=João", "confidence": 0.90}
- "Quero fechar até sexta" → {"fact_type": "timeline", "fact_value": "deadline=2025-01-17", "confidence": 0.85}
`, factText)

    // TODO: Call LLM with structured output
    extracted := &ExtractedFact{}
    // err := m.llmClient.GenerateStructured(ctx, prompt, extracted)

    return extracted, nil
}

// detectContradiction - Detecta contradição entre facts
func (m *MemoryFactService) detectContradiction(
    ctx context.Context,
    existing *MemoryFact,
    newFact *ExtractedFact,
) (bool, ContradictionResolution, error) {
    prompt := fmt.Sprintf(`
Analise se estes dois facts são contraditórios:

Fact 1 (existente):
- Tipo: %s
- Texto: "%s"
- Valor: %v
- Data: %s

Fact 2 (novo):
- Tipo: %s
- Texto: "%s"
- Valor: %v

Retorne JSON:
{
    "contradicts": true/false,
    "resolution": "keep_new|keep_old|merge|both_valid",
    "explanation": "explicação da contradição ou compatibilidade"
}

Exemplos de contradição:
- "Orçamento R$5000" vs "Orçamento R$3000" → contradicts=true, resolution=keep_new
- "Prefiro João" vs "Prefiro João Pedro" → contradicts=true, resolution=keep_new
- "Quer fechar em 30 dias" vs "Quer fechar urgente" → contradicts=false, resolution=both_valid (complementares)
`,
        existing.FactType,
        existing.FactText,
        existing.FactValue,
        existing.ValidFrom.Format("02/01/2006"),
        newFact.FactType,
        newFact.FactText,
        newFact.FactValue,
    )

    // TODO: Call LLM
    resolution := ContradictionResolution{
        Contradicts: false,
        Resolution:  "keep_new",
    }

    // err := m.llmClient.GenerateStructured(ctx, prompt, &resolution)

    return resolution.Contradicts, resolution, nil
}

// resolveContradiction - Resolve contradição baseado em strategy
func (m *MemoryFactService) resolveContradiction(
    ctx context.Context,
    existing *MemoryFact,
    newFact *ExtractedFact,
    resolution ContradictionResolution,
) error {
    now := time.Now()

    switch resolution.Resolution {
    case "keep_new":
        // Invalida fact antigo
        existing.ValidTo = &now
        newFact.Supersedes = &existing.ID
        return m.factRepo.Update(ctx, existing)

    case "keep_old":
        // Não faz nada (descarta novo fact)
        return nil

    case "merge":
        // Merge values (caso específico)
        // TODO: Implement merge logic
        return nil

    case "both_valid":
        // Ambos válidos (não são realmente contraditórios)
        return nil

    default:
        return fmt.Errorf("unknown resolution: %s", resolution.Resolution)
    }
}

// GetActiveFacts - Retorna facts válidos point-in-time
func (m *MemoryFactService) GetActiveFacts(
    ctx context.Context,
    contactID uuid.UUID,
    factTypes []FactType,
    asOf *time.Time,
) ([]MemoryFact, error) {
    if asOf == nil {
        now := time.Now()
        asOf = &now
    }

    return m.factRepo.FindFactsByValidity(ctx, contactID, factTypes, *asOf)
}

// Supporting types
type ExtractedFact struct {
    FactType   FactType               `json:"fact_type"`
    FactValue  interface{}            `json:"fact_value"`
    Confidence float64                `json:"confidence"`
    Metadata   map[string]interface{} `json:"metadata"`
}

type ContradictionResolution struct {
    Contradicts bool   `json:"contradicts"`
    Resolution  string `json:"resolution"`  // "keep_new", "keep_old", "merge", "both_valid"
    Explanation string `json:"explanation"`
}
```

---

## 💾 CONTEXT MANAGER (Prompt Caching)

```go
package memory

import (
    "context"
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
)

// ContextManager - Gerencia caching de contexto (Redis)
type ContextManager struct {
    redisClient *redis.Client
}

// NewContextManager cria novo manager
func NewContextManager(redisClient *redis.Client) *ContextManager {
    return &ContextManager{
        redisClient: redisClient,
    }
}

// GetCached - Busca resultado cacheado
func (c *ContextManager) GetCached(
    ctx context.Context,
    req SearchRequest,
) (*SearchResult, error) {
    // Generate cache key
    cacheKey := c.generateCacheKey(req)

    // Get from Redis
    data, err := c.redisClient.Get(ctx, cacheKey).Bytes()
    if err == redis.Nil {
        return nil, nil  // Cache miss
    }
    if err != nil {
        return nil, fmt.Errorf("redis error: %w", err)
    }

    // Deserialize
    var result SearchResult
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, fmt.Errorf("failed to unmarshal: %w", err)
    }

    return &result, nil
}

// SetCached - Cacheia resultado
func (c *ContextManager) SetCached(
    ctx context.Context,
    req SearchRequest,
    result *SearchResult,
    ttl int,  // seconds
) error {
    // Generate cache key
    cacheKey := c.generateCacheKey(req)

    // Serialize
    data, err := json.Marshal(result)
    if err != nil {
        return fmt.Errorf("failed to marshal: %w", err)
    }

    // Set in Redis with TTL
    return c.redisClient.Set(ctx, cacheKey, data, time.Duration(ttl)*time.Second).Err()
}

// generateCacheKey - Gera key determinística
func (c *ContextManager) generateCacheKey(req SearchRequest) string {
    // Hash based on: contact_id + agent_category + strategy
    key := fmt.Sprintf("memory:%s:%s:%s",
        req.TenantID,
        req.ContactID.String(),
        req.AgentCategory,
    )

    // Se strategy é custom, inclui weights no hash
    if req.MemoryStrategy.Strategy == StrategyCustom {
        key += fmt.Sprintf(":%.2f:%.2f:%.2f:%.2f",
            req.MemoryStrategy.VectorWeight,
            req.MemoryStrategy.KeywordWeight,
            req.MemoryStrategy.GraphWeight,
            req.MemoryStrategy.RecentWeight,
        )
    } else {
        key += ":" + string(req.MemoryStrategy.Strategy)
    }

    // Hash final (para evitar keys muito longas)
    hash := sha256.Sum256([]byte(key))
    return fmt.Sprintf("ctx:%x", hash[:16])  // 32 chars
}

// InvalidateContactCache - Invalida cache de um contato
func (c *ContextManager) InvalidateContactCache(
    ctx context.Context,
    tenantID string,
    contactID uuid.UUID,
) error {
    // Pattern matching no Redis
    pattern := fmt.Sprintf("ctx:*%s*%s*", tenantID, contactID.String())

    iter := c.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
    for iter.Next(ctx) {
        if err := c.redisClient.Del(ctx, iter.Val()).Err(); err != nil {
            continue // Log mas não falha
        }
    }

    return iter.Err()
}
```

Continua na **PART 3** com gRPC API e Database Schema! Quer que eu continue ou prefere que eu crie o documento Python ADK agora?
