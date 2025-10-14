# 🔗 INTEGRATION PLAN: Memory + Message Groups + Documents

> **Como message_groups, memory embeddings e documents se conectam**
> O agente AI "enxerga" PDFs/áudios/vídeos pelo banco vetorial como se estivesse vendo a aba de mídia do WhatsApp

---

## 📋 PROBLEMA ATUAL

```
SITUAÇÃO:
- Message groups existem (agrupamento de mensagens)
- Memory embeddings existem (busca vetorial)
- Documents são processados (PDF → OCR → chunks)
- MAS: São sistemas separados, não integrados

PROBLEMA:
Quando AI Agent busca contexto de um contato:
❌ NÃO vê que o contato enviou 3 PDFs na última semana
❌ NÃO vê que enviou áudio falando sobre "problema X"
❌ NÃO vê referências cruzadas (PDF menciona invoice-123)
❌ Context window limitado (não inclui conteúdo dos documentos)

O QUE QUEREMOS:
✅ AI Agent "enxerga" aba de mídia do WhatsApp virtualmente
✅ Busca vetorial traz PDFs/áudios relevantes automaticamente
✅ Message groups referenciam embeddings related
✅ References bidirecionais (doc ↔ contact ↔ invoice)
```

---

## 🎯 SOLUÇÃO: UNIFIED CONTEXT LAYER

### **Arquitetura Integrada:**

```
┌────────────────────────────────────────────────────────────┐
│                    PYTHON ADK AGENT                         │
│                                                              │
│  agent.run_async(                                           │
│    user_input="Qual o valor do contrato?"                  │
│    session=session                                          │
│  )                                                           │
└──────────────────────┬─────────────────────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────────────┐
│              VENTROS MEMORY SERVICE (Go)                    │
│                                                              │
│  GetMemoryContext(contact_id, lookback_days=30):           │
│                                                              │
│  1. SQL Baseline (recent messages)                         │
│     SELECT * FROM messages WHERE contact_id = X            │
│     → [msg1, msg2, msg3...]                                │
│                                                              │
│  2. Vector Search (semantic memory)                        │
│     SELECT * FROM memory_embeddings                        │
│     WHERE embedding <=> query_embedding                    │
│     → [embedding1, embedding2...]                          │
│                                                              │
│  3. Document References (NEW!)                             │
│     FOR EACH message in recent_messages:                   │
│       IF message.content_type IN (document, audio, video): │
│         → Get message_group                                │
│         → Get embeddings with source_message_id            │
│         → Include document metadata                        │
│                                                              │
│  4. Unified Context Assembly:                              │
│     {                                                        │
│       recent_messages: [...],      // SQL baseline         │
│       semantic_memory: [...],       // Vector search       │
│       documents: [                  // Documents sent      │
│         {                                                    │
│           type: "pdf",                                      │
│           title: "Contrato.pdf",                           │
│           sent_at: "2025-01-10",                           │
│           summary: "Contrato de R$ 10k...",  // AI summary│
│           chunks: [...],            // Relevant chunks     │
│           references: [invoice-123] // Extracted refs      │
│         },                                                   │
│         {                                                    │
│           type: "audio",                                    │
│           duration: "2:30",                                 │
│           transcription: "Sobre o problema X...",          │
│           key_phrases: ["problema", "urgente"],            │
│           sent_at: "2025-01-12"                            │
│         }                                                    │
│       ],                                                     │
│       graph_facts: [...],           // Knowledge graph     │
│       session_summary: "..."        // Session context     │
│     }                                                        │
└────────────────────────────────────────────────────────────┘
```

---

## 🏗️ DATABASE SCHEMA INTEGRATION

### **1. Link Message → Message Group → Embeddings**

```sql
-- Adicionar foreign key em messages
ALTER TABLE messages
ADD COLUMN message_group_id UUID REFERENCES message_groups(id);

-- Index para lookup rápido
CREATE INDEX idx_messages_group ON messages(message_group_id);

-- Adicionar flag para indicar se tem conteúdo vetorizado
ALTER TABLE messages
ADD COLUMN has_embeddings BOOLEAN DEFAULT FALSE;

-- Trigger para atualizar flag
CREATE OR REPLACE FUNCTION update_message_embeddings_flag()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE messages
    SET has_embeddings = TRUE
    WHERE id = NEW.metadata->>'source_message_id';
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_message_embeddings_flag
AFTER INSERT ON memory_embeddings
FOR EACH ROW
WHEN (NEW.metadata->>'source_message_id' IS NOT NULL)
EXECUTE FUNCTION update_message_embeddings_flag();
```

### **2. Enhanced Message Group Schema**

```sql
-- Atualizar message_groups para incluir embedding summary
ALTER TABLE message_groups
ADD COLUMN embedding_ids UUID[] DEFAULT '{}',  -- Array de embedding IDs
ADD COLUMN document_count INT DEFAULT 0,
ADD COLUMN audio_count INT DEFAULT 0,
ADD COLUMN video_count INT DEFAULT 0,
ADD COLUMN image_count INT DEFAULT 0,
ADD COLUMN total_tokens INT DEFAULT 0;  -- Total de tokens dos chunks

-- Index GIN para array search
CREATE INDEX idx_message_groups_embeddings ON message_groups USING GIN (embedding_ids);

-- View para facilitar queries
CREATE VIEW message_groups_with_content AS
SELECT
    mg.id,
    mg.tenant_id,
    mg.group_type,
    mg.title,
    mg.context,
    mg.ai_summary,

    -- Count by content type
    COUNT(DISTINCT me.id) FILTER (WHERE me.content_type = 'document') as doc_count,
    COUNT(DISTINCT me.id) FILTER (WHERE me.content_type = 'audio') as audio_count,
    COUNT(DISTINCT me.id) FILTER (WHERE me.content_type = 'video') as video_count,
    COUNT(DISTINCT me.id) FILTER (WHERE me.content_type = 'image') as image_count,

    -- Total tokens
    SUM(me.token_count) as total_tokens,

    -- Array of document titles
    ARRAY_AGG(DISTINCT me.metadata->>'document_title') FILTER (WHERE me.content_type = 'document') as document_titles

FROM message_groups mg
LEFT JOIN memory_embeddings me ON me.id = ANY(mg.embedding_ids)
GROUP BY mg.id;
```

### **3. Enhanced Memory Embeddings Metadata**

```sql
-- Adicionar campos para facilitar lookup
ALTER TABLE memory_embeddings
ADD COLUMN message_group_id UUID REFERENCES message_groups(id),
ADD COLUMN is_document BOOLEAN GENERATED ALWAYS AS (content_type = 'document') STORED,
ADD COLUMN is_media BOOLEAN GENERATED ALWAYS AS (content_type IN ('audio', 'video', 'image')) STORED;

-- Indexes
CREATE INDEX idx_memory_embeddings_message_group ON memory_embeddings(message_group_id);
CREATE INDEX idx_memory_embeddings_is_document ON memory_embeddings(is_document) WHERE is_document = TRUE;
CREATE INDEX idx_memory_embeddings_is_media ON memory_embeddings(is_media) WHERE is_media = TRUE;
```

---

## 🔄 FLUXO COMPLETO: Message → Group → Embeddings

### **Scenario: Usuário envia PDF**

```
┌─────────────────────────────────────────────────────────────┐
│  STEP 1: Message Received                                    │
│  WAHA webhook → Go backend                                   │
│                                                               │
│  INSERT INTO messages:                                       │
│  {                                                            │
│    id: "msg-001",                                            │
│    content_type: "document",                                 │
│    media_url: "s3://bucket/contrato.pdf",                   │
│    media_mimetype: "application/pdf",                        │
│    contact_id: "contact-123",                                │
│    session_id: "session-456",                                │
│    has_embeddings: FALSE  // Will be updated by trigger     │
│  }                                                            │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  STEP 2: Enqueue Enrichment Job                             │
│  RabbitMQ: message.document.received                        │
│                                                               │
│  Worker picks up job → LlamaParse OCR                       │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  STEP 3: OCR Complete → Chunking                            │
│                                                               │
│  Markdown: "# Contrato... Valor: R$ 10.000..."             │
│  → Split into 15 chunks (512 tokens each)                   │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  STEP 4: Generate Embeddings (batch)                        │
│  Vertex AI: text-embedding-005                              │
│                                                               │
│  FOR EACH chunk:                                             │
│    embedding = embed(chunk_text)                            │
│                                                               │
│    INSERT INTO memory_embeddings:                           │
│    {                                                          │
│      id: "emb-001",                                          │
│      tenant_id: "tenant-123",                                │
│      content_type: "document",                               │
│      content_text: "chunk 1 text...",                       │
│      embedding: vector<768>,                                 │
│      contact_id: "contact-123",                              │
│      session_id: "session-456",                              │
│      metadata: {                                             │
│        source_type: "message",                               │
│        source_message_id: "msg-001",                         │
│        document_title: "Contrato.pdf",                       │
│        document_type: "contract",                            │
│        page_number: 1,                                       │
│        chunk_index: 0,                                       │
│        total_chunks: 15,                                     │
│        entities: [...],                                      │
│        references: [                                         │
│          {type: "contact", id: "contact-123"},              │
│          {type: "invoice", id: "invoice-456"}               │
│        ]                                                      │
│      }                                                        │
│    }                                                          │
│                                                               │
│  → Trigger updates: messages.has_embeddings = TRUE          │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  STEP 5: Create/Update Message Group                        │
│                                                               │
│  INSERT INTO message_groups:                                │
│  {                                                            │
│    id: "group-001",                                          │
│    tenant_id: "tenant-123",                                  │
│    group_type: "document_analysis",                         │
│    title: "Contrato.pdf - 2025-01-10",                      │
│    message_ids: ["msg-001"],                                │
│    embedding_ids: ["emb-001", "emb-002", ..., "emb-015"],  │
│    enrichment_ids: ["enrich-001"],                          │
│    document_count: 1,                                        │
│    total_tokens: 7680,  // 512 * 15                        │
│    context: {                                                │
│      contact_id: "contact-123",                              │
│      session_id: "session-456",                              │
│      topic: "contrato prestação de serviços",               │
│      entities: ["Company A", "João Silva"],                 │
│      references: ["invoice-456"],                            │
│      summary: "Cliente enviou contrato de R$ 10k..."        │
│    },                                                         │
│    ai_summary: "Contrato de prestação entre Company A e..." │
│  }                                                            │
│                                                               │
│  UPDATE messages SET message_group_id = 'group-001'         │
│  WHERE id = 'msg-001';                                       │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  STEP 6: AI Agent Queries Context (LATER)                   │
│  Python ADK → Go Memory Service                              │
│                                                               │
│  GetMemoryContext(contact_id="contact-123"):                │
│                                                               │
│  1. Recent messages (SQL)                                    │
│  2. Vector search (pgvector)                                 │
│  3. Document context (NEW!):                                 │
│                                                               │
│     SELECT                                                    │
│       m.id as message_id,                                    │
│       m.timestamp,                                           │
│       mg.title as group_title,                               │
│       mg.ai_summary,                                         │
│       mg.context,                                            │
│       -- Aggregate embeddings                                │
│       ARRAY_AGG(me.content_text) as chunks,                 │
│       ARRAY_AGG(me.metadata) as chunk_metadata               │
│     FROM messages m                                          │
│     JOIN message_groups mg ON m.message_group_id = mg.id    │
│     JOIN memory_embeddings me ON me.id = ANY(mg.embedding_ids)│
│     WHERE m.contact_id = 'contact-123'                       │
│       AND m.has_embeddings = TRUE                            │
│       AND m.timestamp >= NOW() - INTERVAL '30 days'          │
│     GROUP BY m.id, mg.id                                     │
│                                                               │
│  4. Return unified context:                                  │
│     {                                                         │
│       recent_messages: [...],                                │
│       documents: [                                           │
│         {                                                     │
│           message_id: "msg-001",                             │
│           sent_at: "2025-01-10",                             │
│           type: "pdf",                                       │
│           title: "Contrato.pdf",                             │
│           summary: "Cliente enviou contrato de R$ 10k...",  │
│           chunks: ["chunk1", "chunk2", ...],  // Top 5      │
│           entities: ["Company A", "João Silva"],            │
│           references: ["invoice-456"]                        │
│         }                                                     │
│       ]                                                       │
│     }                                                         │
└─────────────────────────────────────────────────────────────┘
```

---

## 🎨 SCENARIO: Áudio/Transcrição

### **Diferença: Áudio não é chunked, é transcrição completa**

```sql
-- Message de áudio
INSERT INTO messages (
    id, content_type, media_url, media_mimetype,
    contact_id, session_id
) VALUES (
    'msg-002', 'audio', 's3://bucket/audio.ogg', 'audio/ogg',
    'contact-123', 'session-456'
);

-- Enrichment (Whisper transcription)
INSERT INTO message_enrichments (
    id, message_id, enrichment_type, status,
    transcription, metadata
) VALUES (
    'enrich-002', 'msg-002', 'audio', 'completed',
    'Estou ligando para falar sobre o problema X que está atrasando o projeto...',
    '{
        "duration_seconds": 150,
        "speaker_count": 1,
        "language": "pt-BR",
        "sentiment": "concerned",
        "key_phrases": ["problema X", "atraso", "projeto"],
        "confidence": 0.92
    }'
);

-- Memory embedding (SINGLE embedding, not chunked)
INSERT INTO memory_embeddings (
    id, tenant_id, content_type, content_text, embedding,
    contact_id, session_id, metadata
) VALUES (
    'emb-audio-001',
    'tenant-123',
    'audio',
    'Estou ligando para falar sobre o problema X que está atrasando o projeto...',
    VECTOR_EMBEDDING_HERE,
    'contact-123',
    'session-456',
    '{
        "source_type": "message",
        "source_message_id": "msg-002",
        "media_type": "audio",
        "duration_seconds": 150,
        "transcription_confidence": 0.92,
        "key_phrases": ["problema X", "atraso", "projeto"],
        "sentiment": "concerned",
        "entities": [
            {"type": "topic", "value": "problema X", "confidence": 0.95}
        ]
    }'
);

-- Message group
INSERT INTO message_groups (
    id, tenant_id, group_type, title,
    message_ids, embedding_ids, enrichment_ids,
    audio_count, context
) VALUES (
    'group-002',
    'tenant-123',
    'audio_transcription',
    'Áudio sobre problema X - 2025-01-12',
    ARRAY['msg-002'],
    ARRAY['emb-audio-001'],
    ARRAY['enrich-002'],
    1,
    '{
        "contact_id": "contact-123",
        "session_id": "session-456",
        "topic": "problema X",
        "sentiment": "concerned",
        "key_phrases": ["problema X", "atraso", "projeto"]
    }'
);
```

---

## 🔍 ENHANCED MEMORY SERVICE

### **Go: GetMemoryContext with Document/Media Integration**

```go
// infrastructure/memory/context_builder.go

package memory

type UnifiedContext struct {
	RecentMessages  []Message                `json:"recent_messages"`
	SemanticMemory  []MemoryEmbedding       `json:"semantic_memory"`
	Documents       []DocumentContext        `json:"documents"`
	MediaFiles      []MediaContext           `json:"media_files"`
	GraphFacts      []KnowledgeFact          `json:"graph_facts"`
	SessionSummary  string                   `json:"session_summary"`
}

type DocumentContext struct {
	MessageID    string                 `json:"message_id"`
	SentAt       time.Time              `json:"sent_at"`
	Type         string                 `json:"type"` // pdf, docx, xlsx
	Title        string                 `json:"title"`
	Summary      string                 `json:"summary"` // AI-generated summary
	TopChunks    []string               `json:"top_chunks"` // Top 5 relevant chunks
	Entities     []Entity               `json:"entities"`
	References   []Reference            `json:"references"`
	PageCount    int                    `json:"page_count"`
}

type MediaContext struct {
	MessageID      string    `json:"message_id"`
	SentAt         time.Time `json:"sent_at"`
	Type           string    `json:"type"` // audio, video, image
	DurationSec    float64   `json:"duration_sec,omitempty"`
	Transcription  string    `json:"transcription,omitempty"`
	Summary        string    `json:"summary,omitempty"`
	KeyPhrases     []string  `json:"key_phrases"`
	Objects        []string  `json:"objects,omitempty"` // For images/video
	Sentiment      string    `json:"sentiment,omitempty"`
}

func (s *MemoryService) GetMemoryContext(
	ctx context.Context,
	tenantID string,
	contactID string,
	knowledgeScope KnowledgeScope,
) (*UnifiedContext, error) {

	result := &UnifiedContext{}

	// 1. SQL Baseline (recent messages)
	result.RecentMessages = s.getRecentMessages(ctx, tenantID, contactID, knowledgeScope)

	// 2. Vector search (semantic memory)
	if knowledgeScope.IncludeSemanticMemory {
		result.SemanticMemory = s.semanticSearch(ctx, tenantID, contactID, knowledgeScope)
	}

	// 3. Document context (NEW!)
	if knowledgeScope.IncludeDocuments {
		result.Documents = s.getDocumentContext(ctx, tenantID, contactID, knowledgeScope)
	}

	// 4. Media context (NEW!)
	if knowledgeScope.IncludeMedia {
		result.MediaFiles = s.getMediaContext(ctx, tenantID, contactID, knowledgeScope)
	}

	// 5. Graph facts
	if knowledgeScope.IncludeGraphFacts {
		result.GraphFacts = s.getGraphFacts(ctx, tenantID, contactID)
	}

	// 6. Session summary
	if knowledgeScope.SessionID != "" {
		result.SessionSummary = s.getSessionSummary(ctx, knowledgeScope.SessionID)
	}

	return result, nil
}

func (s *MemoryService) getDocumentContext(
	ctx context.Context,
	tenantID string,
	contactID string,
	scope KnowledgeScope,
) []DocumentContext {

	query := `
		SELECT
			m.id as message_id,
			m.timestamp as sent_at,
			mg.title as group_title,
			mg.ai_summary,
			mg.context,

			-- First chunk metadata (for document info)
			(SELECT metadata FROM memory_embeddings
			 WHERE id = mg.embedding_ids[1]) as first_chunk_metadata,

			-- Top 5 relevant chunks (by similarity to recent messages)
			ARRAY(
				SELECT me.content_text
				FROM memory_embeddings me
				WHERE me.id = ANY(mg.embedding_ids)
				ORDER BY me.embedding <=> $4  -- Query embedding
				LIMIT 5
			) as top_chunks

		FROM messages m
		JOIN message_groups mg ON m.message_group_id = mg.id
		WHERE m.tenant_id = $1
			AND m.contact_id = $2
			AND m.content_type = 'document'
			AND m.has_embeddings = TRUE
			AND m.timestamp >= $3
		ORDER BY m.timestamp DESC
		LIMIT 10
	`

	// Generate query embedding from recent conversation
	queryText := s.buildQueryFromRecentMessages(ctx, contactID, scope)
	queryEmbedding, _ := s.embeddingService.Embed(ctx, queryText)

	rows, err := s.db.Raw(query,
		tenantID,
		contactID,
		time.Now().Add(-time.Duration(scope.LookbackDays)*24*time.Hour),
		pgvector.NewVector(queryEmbedding),
	).Rows()

	if err != nil {
		return []DocumentContext{}
	}
	defer rows.Close()

	var documents []DocumentContext
	for rows.Next() {
		var doc DocumentContext
		var metadataJSON string
		var contextJSON string

		rows.Scan(
			&doc.MessageID,
			&doc.SentAt,
			&doc.Title,
			&doc.Summary,
			&contextJSON,
			&metadataJSON,
			pq.Array(&doc.TopChunks),
		)

		// Parse metadata
		var metadata map[string]interface{}
		json.Unmarshal([]byte(metadataJSON), &metadata)

		doc.Type = metadata["mimetype"].(string)
		doc.PageCount = int(metadata["total_chunks"].(float64))

		// Parse entities
		if entitiesRaw, ok := metadata["entities"].([]interface{}); ok {
			for _, e := range entitiesRaw {
				entity := e.(map[string]interface{})
				doc.Entities = append(doc.Entities, Entity{
					Type:  entity["type"].(string),
					Value: entity["value"].(string),
				})
			}
		}

		// Parse references
		if refsRaw, ok := metadata["references"].([]interface{}); ok {
			for _, r := range refsRaw {
				ref := r.(map[string]interface{})
				doc.References = append(doc.References, Reference{
					Type: ref["type"].(string),
					ID:   ref["id"].(string),
				})
			}
		}

		documents = append(documents, doc)
	}

	return documents
}

func (s *MemoryService) getMediaContext(
	ctx context.Context,
	tenantID string,
	contactID string,
	scope KnowledgeScope,
) []MediaContext {

	query := `
		SELECT
			m.id as message_id,
			m.timestamp as sent_at,
			m.content_type as type,
			me_enrich.transcription,
			me_enrich.summary,
			me_enrich.metadata,
			mg.ai_summary

		FROM messages m
		LEFT JOIN message_enrichments me_enrich ON me_enrich.message_id = m.id
		LEFT JOIN message_groups mg ON m.message_group_id = mg.id
		WHERE m.tenant_id = $1
			AND m.contact_id = $2
			AND m.content_type IN ('audio', 'video', 'image')
			AND m.has_embeddings = TRUE
			AND m.timestamp >= $3
		ORDER BY m.timestamp DESC
		LIMIT 10
	`

	rows, err := s.db.Raw(query,
		tenantID,
		contactID,
		time.Now().Add(-time.Duration(scope.LookbackDays)*24*time.Hour),
	).Rows()

	if err != nil {
		return []MediaContext{}
	}
	defer rows.Close()

	var mediaFiles []MediaContext
	for rows.Next() {
		var media MediaContext
		var metadataJSON string

		rows.Scan(
			&media.MessageID,
			&media.SentAt,
			&media.Type,
			&media.Transcription,
			&media.Summary,
			&metadataJSON,
			&media.Summary, // ai_summary fallback
		)

		// Parse metadata
		var metadata map[string]interface{}
		json.Unmarshal([]byte(metadataJSON), &metadata)

		if duration, ok := metadata["duration_seconds"].(float64); ok {
			media.DurationSec = duration
		}

		if phrases, ok := metadata["key_phrases"].([]interface{}); ok {
			for _, p := range phrases {
				media.KeyPhrases = append(media.KeyPhrases, p.(string))
			}
		}

		if objects, ok := metadata["objects"].([]interface{}); ok {
			for _, o := range objects {
				media.Objects = append(media.Objects, o.(string))
			}
		}

		if sentiment, ok := metadata["sentiment"].(string); ok {
			media.Sentiment = sentiment
		}

		mediaFiles = append(mediaFiles, media)
	}

	return mediaFiles
}
```

---

## 🎯 PYTHON ADK: Using Unified Context

```python
# ventros-ai/memory/unified_context.py

from dataclasses import dataclass
from typing import List, Optional
from datetime import datetime

@dataclass
class DocumentContext:
    """Document sent by contact"""
    message_id: str
    sent_at: datetime
    type: str  # pdf, docx, xlsx
    title: str
    summary: str  # AI-generated
    top_chunks: List[str]  # Top 5 relevant chunks
    entities: List[dict]
    references: List[dict]
    page_count: int

@dataclass
class MediaContext:
    """Media (audio/video/image) sent by contact"""
    message_id: str
    sent_at: datetime
    type: str  # audio, video, image
    duration_sec: Optional[float]
    transcription: Optional[str]
    summary: Optional[str]
    key_phrases: List[str]
    objects: List[str]  # For images/video
    sentiment: Optional[str]

@dataclass
class UnifiedContext:
    """Complete context for AI agent"""
    recent_messages: List[dict]
    semantic_memory: List[dict]
    documents: List[DocumentContext]
    media_files: List[MediaContext]
    graph_facts: List[dict]
    session_summary: str

    def to_prompt(self) -> str:
        """Convert to natural language prompt for LLM"""
        prompt = []

        # Recent messages
        if self.recent_messages:
            prompt.append("## Recent Conversation:")
            for msg in self.recent_messages[-10:]:  # Last 10
                sender = "Customer" if not msg["from_me"] else "Agent"
                prompt.append(f"{sender}: {msg['text']}")

        # Documents
        if self.documents:
            prompt.append("\n## Documents Shared:")
            for doc in self.documents:
                prompt.append(f"\n### {doc.title} (sent {doc.sent_at.strftime('%Y-%m-%d')})")
                prompt.append(f"Summary: {doc.summary}")

                if doc.entities:
                    entities_str = ", ".join([f"{e['value']} ({e['type']})" for e in doc.entities])
                    prompt.append(f"Entities: {entities_str}")

                # Include top chunks
                prompt.append("Relevant excerpts:")
                for i, chunk in enumerate(doc.top_chunks[:3], 1):
                    prompt.append(f"  {i}. {chunk[:200]}...")

        # Media
        if self.media_files:
            prompt.append("\n## Media Files Shared:")
            for media in self.media_files:
                prompt.append(f"\n### {media.type.upper()} (sent {media.sent_at.strftime('%Y-%m-%d')})")

                if media.transcription:
                    prompt.append(f"Transcription: {media.transcription}")

                if media.summary:
                    prompt.append(f"Summary: {media.summary}")

                if media.key_phrases:
                    prompt.append(f"Key phrases: {', '.join(media.key_phrases)}")

        # Session summary
        if self.session_summary:
            prompt.append(f"\n## Session Summary:\n{self.session_summary}")

        return "\n".join(prompt)

# Usage in agent
class RetentionChurnAgent(LlmAgent):

    async def run_async(self, user_input: str, session: Session):
        # Get unified context from Go Memory Service
        unified_context = await self.memory_service.get_unified_context(
            tenant_id=session.state["tenant_id"],
            contact_id=session.state["contact_id"],
            knowledge_scope=session.state["knowledge_scope"]
        )

        # Convert to prompt
        context_prompt = unified_context.to_prompt()

        # Build full prompt
        full_prompt = f"""
        {self.instruction}

        # CONTEXT:
        {context_prompt}

        # USER INPUT:
        {user_input}

        # YOUR RESPONSE:
        """

        # Call LLM
        response = await self.model.generate_content_async(full_prompt)

        return response
```

---

## ✅ RESUMO DA INTEGRAÇÃO

### **O que mudou:**

1. **Messages agora linkam para Message Groups**
   - `messages.message_group_id` → `message_groups.id`
   - `messages.has_embeddings` flag

2. **Message Groups referenciam Embeddings**
   - `message_groups.embedding_ids` array
   - Counts por tipo (document_count, audio_count, etc)

3. **Memory Embeddings referenciam Message Groups**
   - `memory_embeddings.message_group_id`
   - Metadata com source_message_id

4. **GetMemoryContext retorna contexto unificado:**
   - Recent messages (SQL baseline)
   - Semantic memory (vector search)
   - **Documents sent** (PDFs com top chunks relevantes)
   - **Media files** (áudios/vídeos com transcrições)
   - Graph facts
   - Session summary

5. **AI Agent "enxerga" a aba de mídia:**
   - Vê PDFs enviados com sumário AI
   - Vê áudios com transcrição
   - Vê referências cruzadas (doc menciona invoice-123)
   - Context window otimizado (top chunks, não documento inteiro)

### **Benefícios:**

✅ **Context-aware**: Agent sabe exatamente o que foi compartilhado
✅ **Efficient**: Apenas top chunks relevantes (não documento inteiro)
✅ **Bidirectional**: References conectam docs ↔ contacts ↔ invoices
✅ **Scalable**: Vetorização permite busca semântica
✅ **Multimodal**: PDFs, áudios, vídeos, imagens integrados

---

## 🗄️ DATA ARCHITECTURE: PostgreSQL vs BigQuery

### **Design Decision: Quando usar colunas vs metadata**

#### **PostgreSQL (Operacional) - Colunas Tipadas**

```sql
-- Design: Dados frequentes em colunas, metadata para flexíveis
CREATE TABLE memory_embeddings (
    -- IDs estruturados (colunas para JOINs)
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    contact_id UUID NOT NULL,
    session_id UUID,
    document_id UUID,  -- ← COLUNA para JOIN eficiente

    -- Filtros comuns (colunas para índices)
    document_name TEXT,  -- ← COLUNA para ILIKE
    document_type TEXT,  -- ← COLUNA para filtering
    content_type TEXT NOT NULL,

    -- Dados primários
    content_text TEXT NOT NULL,
    embedding vector(768) NOT NULL,

    -- Metadata (apenas campos raros/flexíveis)
    metadata JSONB DEFAULT '{}',
    -- {
    --   "page_number": 3,
    --   "chunk_index": 2,
    --   "ocr_confidence": 0.98,
    --   "custom_field_X": "valor"  -- Campos customizados por tenant
    -- }

    created_at TIMESTAMP NOT NULL,

    -- Índices otimizados
    INDEX idx_contact (contact_id),
    INDEX idx_document (document_id),  -- Para JOINs
    INDEX idx_doc_name (document_name),  -- Para ILIKE
    INDEX idx_doc_type (document_type),  -- Para filtering
    INDEX idx_vector USING ivfflat (embedding)
);

-- Query operacional (rápida)
SELECT me.*, ce.summary
FROM memory_embeddings me
JOIN contact_events ce ON ce.metadata->>'document_id' = me.document_id::text
WHERE me.contact_id = 'contact-123'
    AND me.document_type = 'contract'
    AND me.document_name ILIKE '%contrato%';
-- Execution: 15-30ms (índices B-tree + foreign keys)
```

**Benefícios PostgreSQL:**
- ✅ Queries <50ms (índices otimizados)
- ✅ JOINs eficientes (foreign keys)
- ✅ Type safety (PostgreSQL valida)
- ✅ Storage eficiente (sem duplicação JSON)

#### **BigQuery (Analytical) - Metadata Estratégico**

```sql
-- Design: Schema flexível, tudo em JSON para análises
CREATE TABLE `project.dataset.embeddings_warehouse` (
    id STRING NOT NULL,
    tenant_id STRING NOT NULL,

    -- Vector
    embedding ARRAY<FLOAT64>,

    -- Metadata estratégico (TUDO aqui para flexibilidade)
    metadata JSON,
    -- {
    --   // Identifiers
    --   "contact_id": "contact-123",
    --   "document_id": "doc-789",
    --   "session_id": "session-456",
    --   "event_id": "event-123",
    --
    --   // Document info
    --   "document_name": "Contrato.pdf",
    --   "document_type": "contract",
    --   "content_type": "document",
    --
    --   // Business data
    --   "amount_extracted": 10000.00,
    --   "currency": "BRL",
    --   "date_extracted": "2025-01-01",
    --
    --   // Entities
    --   "entities": [
    --     {"type": "company", "value": "Company A"},
    --     {"type": "person", "value": "João Silva"}
    --   ],
    --
    --   // Dimensions (para BI)
    --   "campaign_source": "google_ads",
    --   "agent_type": "human",
    --   "channel_type": "whatsapp",
    --   "pipeline_stage": "qualified",
    --
    --   // Processing
    --   "tokens_used": 1200,
    --   "cost_usd": 0.0012,
    --   "provider": "llamaparse"
    -- }

    created_at TIMESTAMP NOT NULL,
    ingestion_date DATE NOT NULL  -- Partition key
)
PARTITION BY ingestion_date
CLUSTER BY tenant_id, JSON_VALUE(metadata.contact_id);

-- Query BI (analytics)
WITH document_stats AS (
    SELECT
        JSON_VALUE(metadata.document_type) as doc_type,
        JSON_VALUE(metadata.campaign_source) as campaign,
        JSON_VALUE(metadata.pipeline_stage) as stage,
        CAST(JSON_VALUE(metadata.amount_extracted) AS FLOAT64) as amount,
        COUNT(*) as doc_count,
        SUM(CAST(JSON_VALUE(metadata.tokens_used) AS INT64)) as total_tokens
    FROM `project.dataset.embeddings_warehouse`
    WHERE ingestion_date >= DATE_SUB(CURRENT_DATE(), INTERVAL 90 DAY)
        AND JSON_VALUE(metadata.content_type) = 'document'
        AND JSON_VALUE(metadata.tenant_id) = 'tenant-123'
    GROUP BY doc_type, campaign, stage, amount
)
SELECT
    doc_type,
    campaign,
    stage,
    COUNT(*) as unique_documents,
    SUM(doc_count) as total_chunks,
    AVG(amount) as avg_amount,
    SUM(total_tokens) as tokens
FROM document_stats
GROUP BY doc_type, campaign, stage
ORDER BY unique_documents DESC;
-- Execution: 2-5s (scans 90 days, clustering otimiza)
```

**Benefícios BigQuery:**
- ✅ Schema flexível (adiciona campos sem ALTER)
- ✅ Queries analíticas complexas (UNNEST, JSON_VALUE)
- ✅ Partitioning (reduz scan de TB para GB)
- ✅ Integração BI (Looker, DataStudio, Metabase)

### **Decision Matrix: Coluna vs Metadata**

| Critério | Coluna (PostgreSQL) | Metadata (ambos) |
|----------|---------------------|------------------|
| **Frequência de query** | Alta (>50% queries) | Baixa (<20% queries) |
| **Tipo de query** | Filtering, JOINs | Exploratory, flex |
| **Schema stability** | Estável | Evolutivo |
| **Performance requerida** | <50ms | 1-5s OK |
| **Cardinalidade** | Baixa (enums) | Alta (valores únicos) |
| **BI integration** | PostgreSQL views | BigQuery JSON |

**Exemplos:**
- `contact_id` → **Coluna** (query frequente, JOINs)
- `document_name` → **Coluna** (ILIKE frequente)
- `document_type` → **Coluna** (filtering frequente, baixa cardinalidade)
- `page_number` → **Metadata** (query rara)
- `ocr_confidence` → **Metadata** (query rara)
- `campaign_source` → **Metadata BigQuery** (BI analytics)

---

## 🗓️ CONTACT EVENTS AS DOCUMENT INDEX

### **Eventos criam índice temporal de documentos (PostgreSQL)**

```sql
-- 1. Quando documento é recebido, cria evento
INSERT INTO contact_events (
    id, tenant_id, contact_id, category, summary,
    priority, metadata
) VALUES (
    'event-123',
    'tenant-1',
    'contact-456',
    'document_received',
    'Cliente enviou contrato de prestação de serviços',
    'medium',
    '{
        "document_name": "Contrato.pdf",
        "document_id": "doc-uuid-789",
        "document_type": "contract",
        "page_count": 5,
        "file_size_mb": 2.3
    }'::jsonb
);

-- 2. Embeddings linkam ao document_id do evento
INSERT INTO memory_embeddings (
    id, tenant_id, content_type, content_text, embedding,
    contact_id, session_id, metadata
) VALUES (
    'emb-001', 'tenant-1', 'document', 'Chunk 1 text...', VECTOR_HERE,
    'contact-456', 'session-789',
    '{
        "source_type": "message",
        "source_message_id": "msg-001",
        "source_document_id": "doc-uuid-789",  ← LINK ao evento
        "source_event_id": "event-123",
        "document_title": "Contrato.pdf",
        "document_type": "contract",
        "chunk_index": 0,
        "total_chunks": 15
    }'::jsonb
);

-- 3. Query cross-reference: eventos → documentos vetorizados
SELECT
    ce.id as event_id,
    ce.summary as event_summary,
    ce.created_at as event_date,
    ce.metadata->>'document_name' as doc_name,
    ce.metadata->>'document_type' as doc_type,

    -- Count chunks
    COUNT(me.id) as chunk_count,

    -- Sample content
    STRING_AGG(
        SUBSTRING(me.content_text, 1, 100),
        ' | '
    ) as content_preview

FROM contact_events ce
LEFT JOIN memory_embeddings me
    ON me.metadata->>'source_document_id' = ce.metadata->>'document_id'
WHERE ce.contact_id = 'contact-456'
    AND ce.category IN ('document_received', 'document_shared')
    AND ce.created_at >= NOW() - INTERVAL '30 days'
GROUP BY ce.id
ORDER BY ce.created_at DESC;
```

### **MCP Tool: get_contact_events_with_documents**

```python
# Python ADK calls MCP tool
from mcp_client import MCPClient

client = MCPClient("http://localhost:8081", token)

# Get events with linked documents
events_with_docs = client.call_tool("get_contact_events_with_documents", {
    "contact_id": "contact-456",
    "event_categories": ["document_received", "document_shared"],
    "lookback_days": 30
})

# Result:
{
    "events": [
        {
            "event_id": "event-123",
            "category": "document_received",
            "summary": "Cliente enviou contrato",
            "created_at": "2025-01-10T10:30:00Z",
            "document_name": "Contrato.pdf",
            "document_id": "doc-uuid-789",
            "document_type": "contract",
            "chunk_count": 15,  # Vectorized chunks
            "top_chunks": [
                "# Contrato de Prestação... Partes: Company A...",
                "Valor: R$ 10.000,00 mensais...",
                "Vigência: 12 meses..."
            ]
        }
    ],
    "total": 1
}
```

### **AI Agent Usage**

```python
# Agent queries unified context
context = memory_service.get_unified_context(
    contact_id="contact-456",
    knowledge_scope={
        "include_contact_events": True,  # Include events
        "include_documents": True,        # Include vectorized docs
    }
)

# Context includes events timeline + document content
{
    "contact_events": [
        {
            "summary": "Cliente enviou contrato",
            "created_at": "2025-01-10",
            "metadata": {
                "document_name": "Contrato.pdf",
                "document_id": "doc-uuid-789"
            }
        }
    ],
    "documents": [
        {
            "document_id": "doc-uuid-789",  # ← Same as event
            "title": "Contrato.pdf",
            "summary": "Contrato de R$ 10k...",
            "chunks": [...],  # Top 5 relevant chunks
            "sent_at": "2025-01-10"
        }
    ]
}

# Agent prompt includes:
"""
## Timeline of Events:
- 2025-01-10: Cliente enviou contrato (Contrato.pdf)

## Documents Shared:
### Contrato.pdf (sent 2025-01-10)
Summary: Contrato de prestação de serviços entre Company A e João Silva.
Valor: R$ 10.000,00 mensais.
Vigência: 12 meses.

Relevant excerpts:
1. "Partes: Company A LTDA (CNPJ...) e João Silva (CPF...)"
2. "Valor: R$ 10.000,00 (dez mil reais) pagos mensalmente..."
3. "Vigência: 12 meses, iniciando em 01/01/2025..."
"""
```

### **Benefícios da Integração Events → Documents:**

✅ **Timeline visual**: Agent vê "quando" cada documento foi enviado
✅ **Busca por nome**: `WHERE metadata->>'document_name' ILIKE '%contrato%'`
✅ **Cross-reference**: Evento → document_id → chunks vetorizados
✅ **Contexto completo**: Evento (when) + Embeddings (what)
✅ **Efficient queries**: GIN index on metadata JSONB
✅ **Scalable**: Mesmo documento pode aparecer em múltiplos eventos

---

**Próximo:** Plano completo de integração finalizado!
