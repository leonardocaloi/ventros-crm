package ai

// VisionPromptContext defines the context where image analysis is needed
type VisionPromptContext string

const (
	// Chat contexts - mensagens de conversa
	ContextChatMessage        VisionPromptContext = "chat_message"           // Imagem enviada em conversa
	ContextChatProduct        VisionPromptContext = "chat_product"           // Produto compartilhado em chat
	ContextChatDocument       VisionPromptContext = "chat_document"          // Documento/screenshot em chat
	ContextChatReceipt        VisionPromptContext = "chat_receipt"           // Nota fiscal/recibo

	// Profile contexts - dados de perfil
	ContextProfilePicture     VisionPromptContext = "profile_picture"        // Foto de perfil do contato
	ContextProfileDocument    VisionPromptContext = "profile_document"       // Documento de identificação
	ContextProfileBusiness    VisionPromptContext = "profile_business"       // Logo/imagem de empresa

	// Pipeline contexts - funil de vendas
	ContextPipelineProduct    VisionPromptContext = "pipeline_product"       // Produto no pipeline
	ContextPipelineCatalog    VisionPromptContext = "pipeline_catalog"       // Catálogo de produtos
	ContextPipelineContract   VisionPromptContext = "pipeline_contract"      // Contrato/proposta

	// Automation contexts - automações
	ContextAutomationForm     VisionPromptContext = "automation_form"        // Formulário preenchido
	ContextAutomationID       VisionPromptContext = "automation_id"          // Documento de identidade
	ContextAutomationInvoice  VisionPromptContext = "automation_invoice"     // Nota fiscal/fatura
)

// VisionPrompt representa um prompt específico para análise de imagem
type VisionPrompt struct {
	Context     VisionPromptContext
	Prompt      string
	Description string
	OutputType  string // "text", "json", "structured"
	MaxTokens   int
}

// VisionPromptRegistry gerencia prompts por contexto
type VisionPromptRegistry struct {
	prompts map[VisionPromptContext]VisionPrompt
}

// NewVisionPromptRegistry cria um novo registry com prompts padrão
func NewVisionPromptRegistry() *VisionPromptRegistry {
	registry := &VisionPromptRegistry{
		prompts: make(map[VisionPromptContext]VisionPrompt),
	}

	// Registrar prompts padrão
	registry.registerDefaultPrompts()
	return registry
}

// registerDefaultPrompts registra todos os prompts padrão
func (r *VisionPromptRegistry) registerDefaultPrompts() {
	// Chat Message - Imagem em conversa (caso mais comum)
	r.Register(VisionPrompt{
		Context:     ContextChatMessage,
		Description: "Imagem enviada em conversa do WhatsApp/Instagram",
		OutputType:  "text",
		MaxTokens:   300,
		Prompt: `Extraia TODO o texto visível nesta imagem (OCR completo).
Se houver elementos visuais relevantes (produtos, pessoas, marcas), mencione brevemente.
Seja conciso e objetivo - foco no conteúdo útil para entender o contexto da conversa.`,
	})

	// Chat Product - Produto compartilhado
	r.Register(VisionPrompt{
		Context:     ContextChatProduct,
		Description: "Produto compartilhado em conversa",
		OutputType:  "json",
		MaxTokens:   400,
		Prompt: `Analise esta imagem de produto e retorne JSON com:
{
  "texto_visivel": "qualquer texto/marca na imagem",
  "produto": "nome/tipo do produto",
  "marca": "marca identificada (se houver)",
  "condicao": "novo/usado/etc (se visível)",
  "preco": "preço visível (se houver)"
}
Seja objetivo. Se não identificar algo, use null.`,
	})

	// Chat Document - Screenshot/documento em chat
	r.Register(VisionPrompt{
		Context:     ContextChatDocument,
		Description: "Screenshot ou documento enviado em chat",
		OutputType:  "text",
		MaxTokens:   500,
		Prompt: `Extraia TODO o texto desta imagem preservando a estrutura.
Mantenha formatação, números, datas, valores.
Use quebras de linha para separar seções.
Seja completo - este é um documento importante.`,
	})

	// Chat Receipt - Nota fiscal/recibo
	r.Register(VisionPrompt{
		Context:     ContextChatReceipt,
		Description: "Nota fiscal ou recibo",
		OutputType:  "json",
		MaxTokens:   500,
		Prompt: `Extraia informações deste recibo/nota fiscal em JSON:
{
  "empresa": "nome da empresa",
  "cnpj": "CNPJ (se visível)",
  "data": "data da compra",
  "valor_total": "valor total",
  "items": ["lista de produtos/serviços"],
  "numero_nota": "número da nota fiscal"
}
Extraia apenas informações visíveis. Use null se não encontrar.`,
	})

	// Profile Picture - Foto de perfil do contato
	r.Register(VisionPrompt{
		Context:     ContextProfilePicture,
		Description: "Foto de perfil do contato no CRM",
		OutputType:  "json",
		MaxTokens:   200,
		Prompt: `Analise esta foto de perfil:
{
  "tipo": "pessoa/logo/grupo/outro",
  "genero": "masculino/feminino/indefinido (se pessoa)",
  "idade_aproximada": "faixa etária (se pessoa)",
  "profissional": true/false (aparência profissional),
  "contexto": "descrição breve do cenário"
}
Seja respeitoso e objetivo. Foco em dados úteis para CRM.`,
	})

	// Profile Business - Logo/imagem de empresa
	r.Register(VisionPrompt{
		Context:     ContextProfileBusiness,
		Description: "Logo ou imagem de empresa",
		OutputType:  "json",
		MaxTokens:   300,
		Prompt: `Analise esta imagem de empresa/logo:
{
  "nome_empresa": "nome visível no logo/imagem",
  "tipo_negocio": "setor/tipo de negócio (inferido)",
  "texto_visivel": "slogan/texto adicional",
  "cores_principais": ["lista de cores dominantes"],
  "estilo": "moderno/clássico/minimalista/etc"
}
Extraia informações úteis para perfil empresarial.`,
	})

	// Pipeline Product - Produto no funil de vendas
	r.Register(VisionPrompt{
		Context:     ContextPipelineProduct,
		Description: "Produto sendo negociado no pipeline",
		OutputType:  "json",
		MaxTokens:   400,
		Prompt: `Analise este produto para o pipeline de vendas:
{
  "categoria": "categoria do produto",
  "descricao": "descrição objetiva",
  "marca": "marca (se visível)",
  "modelo": "modelo/referência",
  "estado": "novo/usado/etc",
  "diferenciais": ["características únicas"],
  "preco_visivel": "preço se houver"
}
Foco em informações úteis para venda.`,
	})

	// Automation Form - Formulário preenchido
	r.Register(VisionPrompt{
		Context:     ContextAutomationForm,
		Description: "Formulário preenchido (automação)",
		OutputType:  "json",
		MaxTokens:   600,
		Prompt: `Extraia dados deste formulário em JSON estruturado.
Use formato: {"campo": "valor"}
Preserve TODOS os campos visíveis.
Mantenha valores exatamente como escritos.
Use null para campos vazios.`,
	})

	// Automation ID - Documento de identificação
	r.Register(VisionPrompt{
		Context:     ContextAutomationID,
		Description: "Documento de identidade (RG, CNH, etc)",
		OutputType:  "json",
		MaxTokens:   400,
		Prompt: `Extraia informações do documento de identidade:
{
  "tipo_documento": "RG/CNH/RNE/Passaporte/etc",
  "numero": "número do documento",
  "nome": "nome completo",
  "data_nascimento": "data de nascimento",
  "cpf": "CPF (se visível)",
  "orgao_emissor": "órgão emissor",
  "validade": "data de validade (se houver)"
}
ATENÇÃO: Dados sensíveis - manuseie com cuidado.`,
	})

	// Automation Invoice - Nota fiscal/fatura
	r.Register(VisionPrompt{
		Context:     ContextAutomationInvoice,
		Description: "Nota fiscal ou fatura (automação)",
		OutputType:  "json",
		MaxTokens:   800,
		Prompt: `Extraia dados completos desta nota fiscal/fatura:
{
  "numero": "número da nota",
  "serie": "série",
  "data_emissao": "data de emissão",
  "empresa_emitente": {
    "razao_social": "",
    "cnpj": "",
    "endereco": ""
  },
  "cliente": {
    "nome": "",
    "cpf_cnpj": ""
  },
  "items": [
    {"descricao": "", "quantidade": 0, "valor_unitario": 0, "valor_total": 0}
  ],
  "valores": {
    "subtotal": 0,
    "descontos": 0,
    "impostos": 0,
    "total": 0
  }
}
Extraia TODOS os dados visíveis. Precisão é crítica.`,
	})
}

// Register adiciona ou atualiza um prompt
func (r *VisionPromptRegistry) Register(prompt VisionPrompt) {
	r.prompts[prompt.Context] = prompt
}

// Get retorna o prompt para um contexto específico
func (r *VisionPromptRegistry) Get(context VisionPromptContext) (VisionPrompt, bool) {
	prompt, exists := r.prompts[context]
	return prompt, exists
}

// GetPromptText retorna apenas o texto do prompt para um contexto
func (r *VisionPromptRegistry) GetPromptText(context VisionPromptContext) string {
	prompt, exists := r.prompts[context]
	if !exists {
		// Fallback para chat message (mais genérico)
		prompt, _ = r.prompts[ContextChatMessage]
	}
	return prompt.Prompt
}

// ListContexts retorna todos os contextos disponíveis
func (r *VisionPromptRegistry) ListContexts() []VisionPromptContext {
	contexts := make([]VisionPromptContext, 0, len(r.prompts))
	for context := range r.prompts {
		contexts = append(contexts, context)
	}
	return contexts
}

// GetDefaultContext retorna o contexto padrão (chat message)
func (r *VisionPromptRegistry) GetDefaultContext() VisionPromptContext {
	return ContextChatMessage
}
