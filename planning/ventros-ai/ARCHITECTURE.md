# 🤖 PYTHON ADK MULTI-AGENT ARCHITECTURE (2025)

> **Arquitetura completa de agentes AI usando Google ADK**
> Baseado em: ADK Primitives, Multi-Agent Patterns, Event-Driven Architecture
> Stack: Python + ADK + Vertex AI + gRPC + RabbitMQ + OpenTelemetry

---

## 📋 ÍNDICE

1. [Visão Geral](#visão-geral)
2. [ADK Agent Types & Primitives](#adk-agent-types--primitives)
3. [Multi-Agent Orchestration Patterns](#multi-agent-orchestration-patterns)
4. [Memory Service Integration](#memory-service-integration)
5. [Event-Driven Architecture](#event-driven-architecture)
6. [Agent Implementation Examples](#agent-implementation-examples)
7. [Observability & Callbacks](#observability--callbacks)
8. [Production Deployment](#production-deployment)

---

## 🎯 VISÃO GERAL

### Responsabilidades do Python ADK Service

```
┌─────────────────────────────────────────────────────────────┐
│                  PYTHON ADK ORCHESTRATOR                     │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ✅ Agent Orchestration (Coordinator + Specialists)         │
│  ✅ Semantic Routing (Intent Classification)                │
│  ✅ Memory Service (BaseMemoryService implementation)       │
│  ✅ Tool Registry & Execution                                │
│  ✅ LLM Interaction (Gemini 2.0 Flash)                      │
│  ✅ Event Consumer/Publisher (RabbitMQ)                     │
│  ✅ gRPC Client (chama Go Memory Service)                   │
│  ✅ Callbacks & Observability (OpenTelemetry)               │
│  ✅ Session Management & State                               │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### Arquitetura de Comunicação

```
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│   RabbitMQ   │◄────────┤  Python ADK  ├────────►│  Go Memory   │
│  Event Bus   │ async   │ Orchestrator │  gRPC  │   Service    │
└──────────────┘         └──────┬───────┘         └──────────────┘
                                │
                                │ REST/gRPC
                                ▼
                         ┌──────────────┐
                         │   Frontend   │
                         │  (WebSocket) │
                         └──────────────┘
```

### Fluxo Completo (Event-Driven)

```
1. WAHA Webhook → Go API
     ↓
2. Go creates Message + publishes MessageReceived event → RabbitMQ
     ↓
3. Python ADK consumes event
     ↓
4. Semantic Router → Intent Classification
     ↓
5. Coordinator Agent selects Specialist Agent
     ↓
6. Specialist Agent calls Memory Service (gRPC → Go)
     ↓
7. Agent processes with LLM + Tools
     ↓
8. Agent publishes OutboundMessage event → RabbitMQ
     ↓
9. Go consumes event → sends via WAHA
     ↓
10. Background: Go updates embeddings + graph
```

---

## 🔧 ADK AGENT TYPES & PRIMITIVES

### **1. BaseAgent** (Foundation)

Todos os agentes herdam de `BaseAgent`:

```python
from adk import BaseAgent, Session

class BaseAgent:
    """
    Base class para todos os agentes ADK

    Primitives:
    - sub_agents: List[BaseAgent] - Hierarquia pai-filho
    - session: Session - Estado compartilhado
    - name: str - Identificador único
    - instruction: str - System prompt
    """

    def __init__(
        self,
        name: str,
        instruction: str = "",
        sub_agents: List[BaseAgent] = None,
        tools: List[Tool] = None,
        memory_service: BaseMemoryService = None,
    ):
        self.name = name
        self.instruction = instruction
        self.sub_agents = sub_agents or []
        self.tools = tools or []
        self.memory_service = memory_service
```

### **2. LlmAgent** (ReAct Pattern)

Agente que usa LLM para reasoning + tool calling:

```python
from adk import LlmAgent, Tool
from vertexai.generative_models import GenerativeModel

class RetentionChurnAgent(LlmAgent):
    """
    Specialist Agent para churn prevention

    Features:
    - ReAct loop (reasoning + acting)
    - Dynamic tool selection
    - Memory-aware (busca contexto via Memory Service)
    - Self-reflection capable
    """

    def __init__(
        self,
        memory_service: BaseMemoryService,
        tool_registry: ToolRegistry,
    ):
        super().__init__(
            name="retention_churn_agent",
            instruction=self._build_instruction(),
            model=GenerativeModel("gemini-2.0-flash"),
            tools=[
                tool_registry.get("create_retention_offer"),
                tool_registry.get("escalate_to_supervisor"),
                tool_registry.get("search_similar_churn_cases"),
                AgentTool(agent=SupervisorAgent()),  # Agent-as-Tool
            ],
            memory_service=memory_service,
        )

    def _build_instruction(self) -> str:
        return """
        Você é um especialista em retenção de clientes.

        OBJETIVO: Prevenir churn e manter clientes satisfeitos.

        CAPABILITIES:
        - Buscar histórico de interações e padrões de churn
        - Criar ofertas de retenção personalizadas
        - Escalar para supervisor quando necessário
        - Analisar sentiment e identificar sinais de insatisfação

        PROTOCOL:
        1. SEMPRE busque contexto na memória antes de responder
        2. Se sentiment < -0.5: escale imediatamente
        3. Se cliente mencionar "cancelar": ofereça retenção
        4. Seja empático e proativo

        CONSTRAINTS:
        - Nunca prometa algo que não pode cumprir
        - Máximo de 3 tentativas de retenção
        - Se rejeitar 3 vezes: respeite decisão e facilite offboarding
        """
```

### **3. SequentialAgent** (Deterministic Pipeline)

Executa sub-agents em sequência:

```python
from adk import SequentialAgent

class OnboardingPipeline(SequentialAgent):
    """
    Pipeline determinístico de onboarding

    Flow:
    1. WelcomeAgent → Envia boas-vindas
    2. ProfileSetupAgent → Coleta informações
    3. TutorialAgent → Ensina features
    4. ActivationAgent → Primeira ação guiada
    """

    def __init__(
        self,
        memory_service: BaseMemoryService,
    ):
        super().__init__(
            name="onboarding_pipeline",
            instruction="Execute onboarding completo do novo cliente",
            sub_agents=[
                WelcomeAgent(),
                ProfileSetupAgent(),
                TutorialAgent(),
                ActivationAgent(),
            ],
            memory_service=memory_service,
        )

    # SequentialAgent automatically:
    # - Executes agents in order
    # - Passes same session context sequentially
    # - Stops if any agent returns terminal=True
```

### **4. ParallelAgent** (Concurrent Execution)

Executa sub-agents em paralelo:

```python
from adk import ParallelAgent

class LeadEnrichmentAgent(ParallelAgent):
    """
    Enriquece lead com múltiplas fontes em paralelo

    Parallel Tasks:
    - CRM lookup
    - LinkedIn scraping
    - Email verification
    - Company data enrichment
    """

    def __init__(self):
        super().__init__(
            name="lead_enrichment",
            instruction="Enriqueça dados do lead usando múltiplas fontes",
            sub_agents=[
                CRMLookupAgent(),
                LinkedInAgent(),
                EmailVerifierAgent(),
                CompanyDataAgent(),
            ],
        )

    # ParallelAgent automatically:
    # - Runs all sub-agents concurrently
    # - All share same session state
    # - Waits for all to complete
    # - Aggregates results
```

### **5. LoopAgent** (Iterative Execution)

Executa sub-agents repetidamente até condição:

```python
from adk import LoopAgent

class QualityAssuranceAgent(LoopAgent):
    """
    Loop de QA até resposta passar critérios

    Loop:
    1. GenerateAgent → Cria resposta
    2. CriticAgent → Avalia qualidade
    3. If score < threshold: repeat
    4. Else: approve
    """

    def __init__(self):
        super().__init__(
            name="qa_loop",
            instruction="Refine resposta até qualidade adequada",
            sub_agents=[
                ResponseGeneratorAgent(),
                CriticAgent(),
            ],
            termination_condition=self._should_terminate,
            max_iterations=3,
        )

    def _should_terminate(self, session: Session) -> bool:
        """Termina se quality score >= 0.8 ou max iterations"""
        quality_score = session.state.get("quality_score", 0)
        return quality_score >= 0.8
```

---

## 🎭 MULTI-AGENT ORCHESTRATION PATTERNS

### **Pattern 1: Coordinator-Worker** (Recommended for CRM)

```python
from adk import LlmAgent, AgentTool

class CoordinatorAgent(LlmAgent):
    """
    Coordinator que roteia para especialistas

    Pattern:
    - Analisa intent da mensagem
    - Seleciona specialist agent apropriado
    - Delega execução
    - Agrega resultados
    """

    def __init__(
        self,
        memory_service: BaseMemoryService,
    ):
        # Specialist agents como tools
        specialists = [
            SalesProspectingAgent(memory_service),
            RetentionChurnAgent(memory_service),
            SupportTechnicalAgent(memory_service),
            SupportBillingAgent(memory_service),
        ]

        super().__init__(
            name="coordinator",
            instruction="""
            Você é o coordenador central do CRM.

            ROLE: Analisar mensagens e rotear para especialista correto.

            SPECIALISTS AVAILABLE:
            - sales_prospecting: Lead qualification, pricing questions
            - retention_churn: Customer wants to cancel, dissatisfaction
            - support_technical: Bugs, errors, technical issues
            - support_billing: Payment, invoices, billing questions

            PROTOCOL:
            1. Analyze user message intent
            2. Select appropriate specialist
            3. Delegate to specialist using AgentTool
            4. Return specialist's response

            CRITICAL: ALWAYS delegate. You are NOT the one who solves, you ROUTE.
            """,
            model=GenerativeModel("gemini-2.0-flash"),
            tools=[AgentTool(agent=agent) for agent in specialists],
            memory_service=memory_service,
        )
```

### **Pattern 2: Handoff** (Dynamic Transfer)

```python
class HandoffPattern:
    """
    Pattern where agents can dynamically transfer to each other

    Use case: Support → Escalation → Manager
    """

    def __init__(self):
        # Agents can call each other via AgentTool
        self.support_agent = SupportAgent(
            handoff_options=[
                AgentTool(agent=EscalationAgent(), name="escalate"),
                AgentTool(agent=BillingAgent(), name="transfer_billing"),
            ]
        )

class SupportAgent(LlmAgent):
    def __init__(self, handoff_options: List[AgentTool]):
        super().__init__(
            name="support_agent",
            instruction="""
            You are first-line support.

            HANDOFF RULES:
            - If customer is very angry (sentiment < -0.7): escalate
            - If question about billing: transfer_billing
            - Otherwise: solve yourself

            Use tools to handoff when needed.
            """,
            tools=handoff_options,
        )
```

### **Pattern 3: Reflection** (Self-Critique)

```python
class ReflectiveAgent(LlmAgent):
    """
    Agent with self-reflection loop

    Pattern:
    1. Generate initial response
    2. Critique own response
    3. If inadequate: regenerate
    4. Repeat until satisfied
    """

    def __init__(self, memory_service: BaseMemoryService):
        super().__init__(
            name="reflective_agent",
            instruction="""
            You are a careful agent that self-critiques.

            WORKFLOW:
            1. Generate initial response based on context
            2. CRITIQUE your response:
               - Is it accurate based on memory?
               - Is it empathetic?
               - Does it address all user points?
            3. If critique finds issues: REGENERATE
            4. Repeat until satisfied (max 3 iterations)

            Always show your reasoning in <thinking> tags.
            """,
            model=GenerativeModel("gemini-2.0-flash"),
            tools=[
                Tool(name="critique_response", function=self._critique),
                Tool(name="search_memory", function=self._search_memory),
            ],
            memory_service=memory_service,
        )

    def _critique(self, response: str) -> Dict[str, any]:
        """LLM-based self-critique"""
        critique_prompt = f"""
        Critique this response:
        "{response}"

        Evaluate:
        1. Accuracy (0-10)
        2. Empathy (0-10)
        3. Completeness (0-10)

        Return JSON: {{"score": X, "issues": ["..."], "should_regenerate": bool}}
        """
        # Call LLM for critique
        critique_result = self.model.generate_content(critique_prompt)
        return json.loads(critique_result.text)
```

### **Pattern 4: Hierarchical** (Tree Structure)

```python
class HierarchicalCRM(LlmAgent):
    """
    Hierarquia de agentes espelhando estrutura organizacional

    Tree:
    CEO Agent (strategy)
      ├─ Sales Director (sales team coordination)
      │   ├─ Prospecting Agent
      │   ├─ Negotiation Agent
      │   └─ Closing Agent
      ├─ Support Director (support coordination)
      │   ├─ Technical Agent
      │   └─ Billing Agent
      └─ Retention Director (retention coordination)
          ├─ Churn Agent
          └─ Upsell Agent
    """

    def __init__(self, memory_service: BaseMemoryService):
        # Build tree structure
        ceo = LlmAgent(
            name="ceo_agent",
            instruction="Strategic decisions and escalations",
            sub_agents=[
                SalesDirector(memory_service),
                SupportDirector(memory_service),
                RetentionDirector(memory_service),
            ],
        )

        self.root = ceo

class SalesDirector(LlmAgent):
    def __init__(self, memory_service: BaseMemoryService):
        super().__init__(
            name="sales_director",
            instruction="Coordinate sales team",
            sub_agents=[
                ProspectingAgent(memory_service),
                NegotiationAgent(memory_service),
                ClosingAgent(memory_service),
            ],
            tools=[
                AgentTool(agent=agent) for agent in self.sub_agents
            ],
        )
```

---

## 💾 MEMORY SERVICE INTEGRATION

### **Custom BaseMemoryService Implementation**

```python
from adk import BaseMemoryService, Session
import grpc
from typing import List, Dict
from google.protobuf.json_format import MessageToDict

# Import gRPC generated code
import memory_service_pb2
import memory_service_pb2_grpc

class VentrosMemoryService(BaseMemoryService):
    """
    Custom Memory Service that integrates with Go Memory Service via gRPC

    Implements ADK's BaseMemoryService interface:
    - add_session_to_memory(session: Session)
    - search_memory(query: str, session: Session) -> str
    """

    def __init__(
        self,
        grpc_host: str = "localhost:50051",
        default_agent_category: str = "balanced",
    ):
        self.grpc_host = grpc_host
        self.default_agent_category = default_agent_category
        self.channel = grpc.insecure_channel(grpc_host)
        self.stub = memory_service_pb2_grpc.MemoryServiceStub(self.channel)

    def add_session_to_memory(
        self,
        session: Session,
    ) -> None:
        """
        Adiciona sessão completa à memória

        Flow:
        1. Extrai dados relevantes do session
        2. Chama Go via gRPC para gerar embeddings
        3. Go persiste embeddings + atualiza graph

        Note: Isso é chamado ASYNC (não bloqueia agent execution)
        """
        try:
            request = memory_service_pb2.AddSessionRequest(
                tenant_id=session.state.get("tenant_id"),
                session_id=session.state.get("session_id"),
                contact_id=session.state.get("contact_id"),
                messages=self._extract_messages(session),
                metadata=self._extract_metadata(session),
            )

            # Async call (não espera embedding ser gerado)
            self.stub.AddSession(request)

        except grpc.RpcError as e:
            print(f"Failed to add session to memory: {e}")

    def search_memory(
        self,
        query: str,
        session: Session,
    ) -> str:
        """
        Busca memória relevante para query

        Flow:
        1. Extrai contexto do session (contact_id, agent_category)
        2. Chama Go via gRPC para hybrid search
        3. Go retorna contexto formatado
        4. Retorna como string para LLM

        Note: Isso é chamado SYNC (bloqueia até retornar)
        """
        try:
            request = memory_service_pb2.SearchMemoryRequest(
                tenant_id=session.state.get("tenant_id"),
                contact_id=session.state.get("contact_id"),
                query=query,
                agent_category=session.state.get("agent_category", self.default_agent_category),
                top_k=10,
            )

            response = self.stub.SearchMemory(request)

            # Format para LLM consumption
            return self._format_memory_context(response)

        except grpc.RpcError as e:
            print(f"Failed to search memory: {e}")
            return "No relevant memory found."

    def _extract_messages(self, session: Session) -> List[Dict]:
        """Extrai mensagens do session history"""
        messages = []
        for msg in session.history:
            messages.append({
                "role": msg.role,
                "content": msg.content,
                "timestamp": msg.timestamp,
            })
        return messages

    def _extract_metadata(self, session: Session) -> Dict:
        """Extrai metadata relevante"""
        return {
            "agent_id": session.state.get("agent_id"),
            "pipeline_id": session.state.get("pipeline_id"),
            "sentiment_score": session.state.get("sentiment_score"),
            "topics": session.state.get("topics", []),
        }

    def _format_memory_context(
        self,
        response: memory_service_pb2.SearchMemoryResponse,
    ) -> str:
        """
        Formata resposta do Go para consumo do LLM

        Output format:
        ```
        === RECENT MESSAGES ===
        [Last 20 messages from conversation]

        === SIMILAR PAST SESSIONS ===
        1. [Summary of similar session 1]
        2. [Summary of similar session 2]
        ...

        === CONTACT CONTEXT ===
        - Total sessions: X
        - Avg sentiment: Y
        - Last interaction: Z
        - Campaign source: W

        === MEMORY FACTS ===
        - Budget: R$ 5000
        - Preference: Prefer phone over email
        - Pain point: Slow response times
        ```
        """
        context_parts = []

        # Recent messages
        if response.recent_messages:
            context_parts.append("=== RECENT MESSAGES ===")
            for msg in response.recent_messages:
                context_parts.append(f"[{msg.timestamp}] {msg.role}: {msg.content}")

        # Similar sessions
        if response.similar_sessions:
            context_parts.append("\n=== SIMILAR PAST SESSIONS ===")
            for i, session in enumerate(response.similar_sessions, 1):
                context_parts.append(f"{i}. {session.summary} (similarity: {session.score:.2f})")

        # Contact stats
        if response.contact_stats:
            stats = response.contact_stats
            context_parts.append("\n=== CONTACT CONTEXT ===")
            context_parts.append(f"- Total sessions: {stats.total_sessions}")
            context_parts.append(f"- Avg sentiment: {stats.avg_sentiment:.2f}")
            context_parts.append(f"- Last interaction: {stats.last_interaction_at}")

        # Memory facts
        if response.memory_facts:
            context_parts.append("\n=== MEMORY FACTS ===")
            for fact in response.memory_facts:
                context_parts.append(f"- {fact.fact_type}: {fact.fact_text}")

        return "\n".join(context_parts)
```

### **Using Memory Service in Agents**

```python
class MemoryAwareAgent(LlmAgent):
    """
    Agent que usa Memory Service para contexto

    Flow:
    1. Recebe query do usuário
    2. Busca contexto na memória (search_memory)
    3. LLM processa com contexto
    4. Retorna resposta
    5. Sessão é adicionada à memória (add_session_to_memory)
    """

    def __init__(self, memory_service: VentrosMemoryService):
        super().__init__(
            name="memory_aware_agent",
            instruction="""
            You have access to long-term memory via search_memory tool.

            PROTOCOL:
            1. ALWAYS search memory before responding
            2. Use memory context to personalize responses
            3. Reference specific past interactions when relevant

            Memory will include:
            - Recent conversation history
            - Similar past sessions
            - Contact stats and context
            - Stored facts (preferences, constraints, etc)
            """,
            memory_service=memory_service,
            tools=[
                # ADK automatically adds MemoryTool when memory_service is provided
            ],
        )

# Usage
memory_service = VentrosMemoryService(grpc_host="localhost:50051")
agent = MemoryAwareAgent(memory_service)

# Run agent
session = Session(state={
    "tenant_id": "tenant-123",
    "contact_id": "contact-456",
    "agent_category": "retention_churn",
})

response = agent.run(
    user_input="Quero cancelar minha conta",
    session=session,
)

# After agent completes, add to memory (async)
memory_service.add_session_to_memory(session)
```

---

## 🔄 EVENT-DRIVEN ARCHITECTURE

### **RabbitMQ Event Consumer**

```python
import pika
import json
from typing import Callable, Dict
from dataclasses import dataclass

@dataclass
class MessageReceivedEvent:
    """Event published quando nova mensagem chega"""
    message_id: str
    contact_id: str
    session_id: str
    tenant_id: str
    text: str
    from_me: bool
    timestamp: str
    channel_id: str
    metadata: Dict

class EventConsumer:
    """
    Consome eventos do RabbitMQ e delega para handlers

    Pattern: Orchestrator-Worker com async event processing
    """

    def __init__(
        self,
        rabbitmq_url: str,
        exchange: str = "ventros.events",
    ):
        self.connection = pika.BlockingConnection(
            pika.URLParameters(rabbitmq_url)
        )
        self.channel = self.connection.channel()
        self.exchange = exchange

        # Declare exchange
        self.channel.exchange_declare(
            exchange=self.exchange,
            exchange_type='topic',
            durable=True,
        )

        # Event handlers registry
        self.handlers: Dict[str, Callable] = {}

    def register_handler(
        self,
        event_type: str,
        handler: Callable,
    ):
        """Registra handler para tipo de evento"""
        self.handlers[event_type] = handler

    def start(self):
        """Inicia consumo de eventos"""
        # Create queue
        result = self.channel.queue_declare(queue='', exclusive=True)
        queue_name = result.method.queue

        # Bind queue to exchange for all event types
        for event_type in self.handlers.keys():
            self.channel.queue_bind(
                exchange=self.exchange,
                queue=queue_name,
                routing_key=f"message.{event_type}",
            )

        # Start consuming
        self.channel.basic_consume(
            queue=queue_name,
            on_message_callback=self._handle_message,
            auto_ack=False,
        )

        print("Started consuming events...")
        self.channel.start_consuming()

    def _handle_message(
        self,
        ch,
        method,
        properties,
        body,
    ):
        """Handle incoming event"""
        try:
            # Parse event
            event_data = json.loads(body)
            event_type = event_data.get("event_type")

            # Find handler
            handler = self.handlers.get(event_type)
            if not handler:
                print(f"No handler for event type: {event_type}")
                ch.basic_ack(delivery_tag=method.delivery_tag)
                return

            # Execute handler
            handler(event_data)

            # Ack message
            ch.basic_ack(delivery_tag=method.delivery_tag)

        except Exception as e:
            print(f"Error handling event: {e}")
            # Nack message (will be requeued)
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)

# Usage
consumer = EventConsumer(rabbitmq_url="amqp://localhost:5672")

# Register handlers
consumer.register_handler("message.received", handle_message_received)
consumer.register_handler("session.ended", handle_session_ended)

# Start consuming
consumer.start()
```

### **Message Handler with Agent Orchestration**

```python
from adk import Session

async def handle_message_received(event_data: Dict):
    """
    Handler para MessageReceivedEvent

    Flow:
    1. Parse event
    2. Create/load session
    3. Semantic routing (intent classification)
    4. Dispatch to appropriate agent
    5. Publish outbound response event
    """

    # 1. Parse event
    event = MessageReceivedEvent(**event_data["payload"])

    # 2. Load or create session
    session = await load_or_create_session(
        contact_id=event.contact_id,
        session_id=event.session_id,
        tenant_id=event.tenant_id,
    )

    # Add incoming message to session
    session.history.append({
        "role": "user",
        "content": event.text,
        "timestamp": event.timestamp,
    })

    # 3. Semantic routing (intent classification)
    agent_category = await semantic_router.route(
        message=event.text,
        session=session,
    )

    session.state["agent_category"] = agent_category

    # 4. Get coordinator agent
    coordinator = get_coordinator_agent()

    # 5. Execute agent (async)
    response = await coordinator.run_async(
        user_input=event.text,
        session=session,
    )

    # 6. Publish outbound response event
    await publish_outbound_message_event(
        contact_id=event.contact_id,
        session_id=event.session_id,
        tenant_id=event.tenant_id,
        text=response.output,
        agent_id=response.agent_id,
        source="bot",
        metadata={
            "agent_category": agent_category,
            "confidence": response.confidence,
        },
    )

    # 7. Add session to memory (async, não bloqueia)
    await memory_service.add_session_to_memory_async(session)
```

### **Semantic Router Implementation**

```python
from semantic_router import SemanticRouter, Route
from typing import Dict

class VentrosSemanticRouter:
    """
    Semantic router para classificação de intent

    Uses:
    - Embedding-based similarity search
    - Zero-shot classification (no training needed)
    - Fast (<50ms) deterministic routing
    """

    def __init__(self, embedding_service: EmbeddingService):
        self.embedding_service = embedding_service

        # Define routes com examples
        self.router = SemanticRouter(
            routes=[
                Route(
                    name="retention_churn",
                    utterances=[
                        "quero cancelar",
                        "não quero mais",
                        "vou desistir",
                        "isso não tá funcionando",
                        "muito caro",
                        "insatisfeito",
                        "esperava mais",
                    ],
                ),
                Route(
                    name="sales_prospecting",
                    utterances=[
                        "quanto custa",
                        "qual o preço",
                        "tem desconto",
                        "quero saber valores",
                        "planos disponíveis",
                        "orçamento",
                    ],
                ),
                Route(
                    name="support_technical",
                    utterances=[
                        "não está funcionando",
                        "deu erro",
                        "bug",
                        "problema técnico",
                        "não carrega",
                        "travou",
                    ],
                ),
                Route(
                    name="support_billing",
                    utterances=[
                        "cobrança errada",
                        "fatura",
                        "pagamento",
                        "não recebi boleto",
                        "valor incorreto",
                    ],
                ),
            ],
            encoder=embedding_service,
        )

    async def route(
        self,
        message: str,
        session: Session,
    ) -> str:
        """
        Classifica intent e retorna agent_category

        Returns:
        - Agent category string (ex: "retention_churn")
        - Falls back to "balanced" if no match
        """
        # Semantic similarity search
        route = self.router.route(message)

        if route:
            return route.name

        # Fallback: balanced agent
        return "balanced"
```

---

## 📝 AGENT IMPLEMENTATION EXAMPLES

### **Complete Retention Churn Agent**

```python
from adk import LlmAgent, Tool, AgentTool, Session
from vertexai.generative_models import GenerativeModel
from typing import Dict, List

class RetentionChurnAgent(LlmAgent):
    """
    Production-ready Retention Churn Agent

    Features:
    - Memory-aware (busca contexto histórico)
    - Tool-enabled (create offers, escalate)
    - Self-reflective (critique próprias respostas)
    - Callback-instrumented (observability)
    """

    def __init__(
        self,
        memory_service: VentrosMemoryService,
        tool_registry: ToolRegistry,
        callback_manager: CallbackManager,
    ):
        # Define tools
        tools = [
            # Memory tool (auto-added by ADK)
            # MemoryTool is implicit when memory_service is provided

            # Custom tools
            tool_registry.get("create_retention_offer"),
            tool_registry.get("escalate_to_supervisor"),
            tool_registry.get("check_customer_value"),
            tool_registry.get("get_past_offers"),

            # Agent-as-tool (supervisor escalation)
            AgentTool(
                agent=SupervisorAgent(memory_service),
                name="escalate_supervisor",
                description="Escalate to human supervisor when situation is critical",
            ),
        ]

        super().__init__(
            name="retention_churn_agent",
            instruction=self._build_instruction(),
            model=GenerativeModel(
                "gemini-2.0-flash",
                generation_config={
                    "temperature": 0.7,
                    "top_p": 0.95,
                    "max_output_tokens": 2048,
                },
            ),
            tools=tools,
            memory_service=memory_service,
            callbacks=callback_manager.get_callbacks(),
        )

    def _build_instruction(self) -> str:
        return """
        # ROLE
        You are an expert Customer Retention Specialist for Ventros CRM.
        Your mission is to prevent churn and keep customers satisfied.

        # CAPABILITIES
        You have access to:
        - Long-term memory of all customer interactions
        - Tools to create personalized retention offers
        - Ability to escalate to human supervisor
        - Customer lifetime value (CLV) data
        - History of past retention attempts

        # PROTOCOL

        ## Step 1: Search Memory
        ALWAYS start by searching memory with query about:
        - Customer's interaction history
        - Past complaints or issues
        - Previous retention attempts
        - Sentiment trends

        ## Step 2: Assess Situation
        Analyze memory context to determine:
        - Severity (1-10): How urgent is the churn risk?
        - Root cause: What's driving dissatisfaction?
        - Customer value: Check CLV with check_customer_value tool
        - History: Any patterns of complaints?

        ## Step 3: Decision Tree

        ### If severity >= 8 OR CLV > R$ 50,000:
        - Escalate immediately to supervisor
        - Do NOT attempt retention yourself

        ### If severity 5-7:
        - Check past_offers to avoid repeating
        - Create personalized retention offer
        - Max 3 attempts per customer (check history)
        - Focus on addressing root cause

        ### If severity < 5:
        - Address concerns empathetically
        - Propose solutions without discounts
        - Document feedback for product team

        ## Step 4: Self-Reflection
        Before sending response:
        - Critique: Is this empathetic enough?
        - Critique: Does it address root cause?
        - Critique: Am I over-promising?

        If critique fails: regenerate response

        ## Step 5: Execute
        - Send response
        - If offering retention: call create_retention_offer tool
        - If escalating: call escalate_supervisor tool

        # CONSTRAINTS
        - NEVER promise features that don't exist
        - NEVER offer discounts > 30% without supervisor approval
        - NEVER attempt retention > 3 times (respect customer decision)
        - ALWAYS be honest and transparent

        # TONE
        - Empathetic and understanding
        - Professional but warm
        - Solution-oriented
        - Respectful of customer's autonomy

        # EXAMPLES

        User: "Quero cancelar, isso tá muito caro"
        Agent:
        <thinking>
        - Severity: 6 (price objection)
        - Need: Search memory for pricing discussions
        - Action: Check CLV, create offer if valuable
        </thinking>

        [Searches memory, finds customer has been with us 2 years, paying R$ 200/mo]

        "Entendo sua preocupação, [Nome]. Vi que você está conosco há 2 anos e
        valorizamos muito sua parceria. Posso oferecer 25% de desconto nos próximos
        3 meses enquanto ajustamos seu plano. O que acha?"
        """

    # Custom callback hooks (optional)
    def on_before_model_call(self, session: Session):
        """Called before each LLM call"""
        print(f"[RetentionAgent] Calling LLM with context: {len(session.history)} messages")

    def on_after_tool_call(self, tool_name: str, tool_result: any):
        """Called after each tool execution"""
        print(f"[RetentionAgent] Tool {tool_name} returned: {tool_result}")
```

### **Tool Implementations**

```python
from adk import Tool
from typing import Dict

def create_retention_offer_tool() -> Tool:
    """
    Tool para criar ofertas de retenção

    This calls Go backend via gRPC to:
    - Create discount offer in database
    - Track retention attempt
    - Update customer profile
    """

    async def execute(
        contact_id: str,
        discount_percent: int,
        duration_months: int,
        reason: str,
    ) -> Dict:
        """
        Create retention offer

        Args:
            contact_id: UUID of contact
            discount_percent: Discount percentage (1-30)
            duration_months: How many months (1-12)
            reason: Reason for offer (for tracking)

        Returns:
            Dict with offer_id and confirmation
        """
        # Call Go API via gRPC
        request = retention_service_pb2.CreateOfferRequest(
            contact_id=contact_id,
            discount_percent=discount_percent,
            duration_months=duration_months,
            reason=reason,
        )

        response = retention_service_stub.CreateOffer(request)

        return {
            "offer_id": response.offer_id,
            "discount_value": response.discount_value,
            "valid_until": response.valid_until,
            "message": f"Oferta criada: {discount_percent}% por {duration_months} meses",
        }

    return Tool(
        name="create_retention_offer",
        description="""
        Creates a retention offer for customer at risk of churning.
        Use when customer complains about price or wants to cancel.
        Maximum discount: 30% (needs supervisor approval above this).
        """,
        function=execute,
    )
```

---

## 📊 OBSERVABILITY & CALLBACKS

### **ADK Callbacks for Observability**

```python
from adk import (
    BeforeAgentCallback,
    AfterAgentCallback,
    BeforeModelCallback,
    AfterModelCallback,
    BeforeToolCallback,
    AfterToolCallback,
)
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

# Configure OpenTelemetry
tracer_provider = TracerProvider()
tracer_provider.add_span_processor(
    BatchSpanProcessor(OTLPSpanExporter(endpoint="http://localhost:4317"))
)
trace.set_tracer_provider(tracer_provider)
tracer = trace.get_tracer(__name__)

class ObservabilityCallbacks:
    """
    Callbacks para instrumentar agents com observability

    Features:
    - OpenTelemetry tracing
    - Metrics collection
    - Logging
    - Error tracking
    """

    @staticmethod
    def before_agent(agent_name: str, input_data: any):
        """Called before agent execution starts"""
        with tracer.start_as_current_span(f"agent.{agent_name}") as span:
            span.set_attribute("agent.name", agent_name)
            span.set_attribute("agent.input_length", len(str(input_data)))
            print(f"[Agent Start] {agent_name}")

    @staticmethod
    def after_agent(agent_name: str, output_data: any, duration_ms: float):
        """Called after agent execution completes"""
        with tracer.start_as_current_span(f"agent.{agent_name}.complete") as span:
            span.set_attribute("agent.name", agent_name)
            span.set_attribute("agent.duration_ms", duration_ms)
            span.set_attribute("agent.output_length", len(str(output_data)))
            print(f"[Agent Complete] {agent_name} ({duration_ms}ms)")

    @staticmethod
    def before_model(model_name: str, prompt: str, session_history_len: int):
        """Called before LLM API call"""
        with tracer.start_as_current_span(f"model.{model_name}") as span:
            span.set_attribute("model.name", model_name)
            span.set_attribute("model.prompt_length", len(prompt))
            span.set_attribute("model.history_length", session_history_len)
            print(f"[LLM Call] {model_name} (context: {session_history_len} msgs)")

    @staticmethod
    def after_model(
        model_name: str,
        response: str,
        tokens_used: int,
        duration_ms: float,
    ):
        """Called after LLM API call"""
        with tracer.start_as_current_span(f"model.{model_name}.complete") as span:
            span.set_attribute("model.name", model_name)
            span.set_attribute("model.tokens_used", tokens_used)
            span.set_attribute("model.duration_ms", duration_ms)
            span.set_attribute("model.response_length", len(response))
            print(f"[LLM Complete] {tokens_used} tokens ({duration_ms}ms)")

    @staticmethod
    def before_tool(tool_name: str, tool_args: Dict):
        """Called before tool execution"""
        with tracer.start_as_current_span(f"tool.{tool_name}") as span:
            span.set_attribute("tool.name", tool_name)
            span.set_attribute("tool.args", str(tool_args))
            print(f"[Tool Call] {tool_name}({tool_args})")

    @staticmethod
    def after_tool(tool_name: str, tool_result: any, duration_ms: float):
        """Called after tool execution"""
        with tracer.start_as_current_span(f"tool.{tool_name}.complete") as span:
            span.set_attribute("tool.name", tool_name)
            span.set_attribute("tool.duration_ms", duration_ms)
            span.set_attribute("tool.success", tool_result is not None)
            print(f"[Tool Complete] {tool_name} ({duration_ms}ms)")

# Usage: Register callbacks with agents
agent = RetentionChurnAgent(
    memory_service=memory_service,
    tool_registry=tool_registry,
    callbacks=[
        BeforeAgentCallback(ObservabilityCallbacks.before_agent),
        AfterAgentCallback(ObservabilityCallbacks.after_agent),
        BeforeModelCallback(ObservabilityCallbacks.before_model),
        AfterModelCallback(ObservabilityCallbacks.after_model),
        BeforeToolCallback(ObservabilityCallbacks.before_tool),
        AfterToolCallback(ObservabilityCallbacks.after_tool),
    ],
)
```

### **Metrics Collection**

```python
from prometheus_client import Counter, Histogram, Gauge
import time

# Define metrics
agent_requests_total = Counter(
    'agent_requests_total',
    'Total agent requests',
    ['agent_name', 'agent_category'],
)

agent_duration_seconds = Histogram(
    'agent_duration_seconds',
    'Agent execution duration',
    ['agent_name', 'agent_category'],
)

agent_errors_total = Counter(
    'agent_errors_total',
    'Total agent errors',
    ['agent_name', 'error_type'],
)

llm_tokens_used = Counter(
    'llm_tokens_used_total',
    'Total LLM tokens used',
    ['model_name', 'agent_name'],
)

tool_calls_total = Counter(
    'tool_calls_total',
    'Total tool calls',
    ['tool_name', 'agent_name'],
)

class MetricsCallbacks:
    """Callbacks for metrics collection"""

    @staticmethod
    def before_agent(agent_name: str, agent_category: str):
        agent_requests_total.labels(
            agent_name=agent_name,
            agent_category=agent_category,
        ).inc()

    @staticmethod
    def after_agent(agent_name: str, agent_category: str, duration_ms: float):
        agent_duration_seconds.labels(
            agent_name=agent_name,
            agent_category=agent_category,
        ).observe(duration_ms / 1000.0)

    @staticmethod
    def on_error(agent_name: str, error_type: str):
        agent_errors_total.labels(
            agent_name=agent_name,
            error_type=error_type,
        ).inc()

    @staticmethod
    def after_model(model_name: str, agent_name: str, tokens_used: int):
        llm_tokens_used.labels(
            model_name=model_name,
            agent_name=agent_name,
        ).inc(tokens_used)

    @staticmethod
    def after_tool(tool_name: str, agent_name: str):
        tool_calls_total.labels(
            tool_name=tool_name,
            agent_name=agent_name,
        ).inc()
```

---

## 🚀 PRODUCTION DEPLOYMENT

### **Agent Factory Pattern**

```python
from typing import Dict
from adk import LlmAgent

class AgentFactory:
    """
    Factory para criar agents configurados para produção

    Features:
    - Dependency injection
    - Configuration management
    - Callback registration
    - Error handling
    """

    def __init__(
        self,
        memory_service: VentrosMemoryService,
        tool_registry: ToolRegistry,
        callback_manager: CallbackManager,
        config: Dict,
    ):
        self.memory_service = memory_service
        self.tool_registry = tool_registry
        self.callback_manager = callback_manager
        self.config = config

        # Agent registry
        self._agents: Dict[str, LlmAgent] = {}

    def create_agent(self, agent_category: str) -> LlmAgent:
        """Creates or returns cached agent"""
        if agent_category in self._agents:
            return self._agents[agent_category]

        # Create agent based on category
        agent_class = self._get_agent_class(agent_category)
        agent = agent_class(
            memory_service=self.memory_service,
            tool_registry=self.tool_registry,
            callback_manager=self.callback_manager,
        )

        # Cache
        self._agents[agent_category] = agent
        return agent

    def _get_agent_class(self, agent_category: str):
        """Maps category to agent class"""
        mapping = {
            "sales_prospecting": SalesProspectingAgent,
            "sales_negotiation": SalesNegotiationAgent,
            "sales_closing": SalesClosingAgent,
            "retention_churn": RetentionChurnAgent,
            "retention_upsell": RetentionUpsellAgent,
            "support_technical": SupportTechnicalAgent,
            "support_billing": SupportBillingAgent,
            "operations_followup": OperationsFollowupAgent,
        }
        return mapping.get(agent_category, BalancedAgent)

    def create_coordinator(self) -> LlmAgent:
        """Creates coordinator agent with all specialists"""
        specialists = [
            self.create_agent(category)
            for category in self.config.get("enabled_categories", [])
        ]

        return CoordinatorAgent(
            specialists=specialists,
            memory_service=self.memory_service,
            callback_manager=self.callback_manager,
        )
```

### **Main Application**

```python
import asyncio
from fastapi import FastAPI, WebSocket
from typing import Dict

# FastAPI app
app = FastAPI(title="Ventros AI Agents")

# Global instances
agent_factory: AgentFactory = None
event_consumer: EventConsumer = None
memory_service: VentrosMemoryService = None

@app.on_event("startup")
async def startup():
    """Initialize services"""
    global agent_factory, event_consumer, memory_service

    # Initialize memory service
    memory_service = VentrosMemoryService(
        grpc_host="localhost:50051",
    )

    # Initialize agent factory
    agent_factory = AgentFactory(
        memory_service=memory_service,
        tool_registry=ToolRegistry(),
        callback_manager=CallbackManager(),
        config=load_config(),
    )

    # Initialize event consumer
    event_consumer = EventConsumer(
        rabbitmq_url="amqp://localhost:5672",
    )

    # Register event handlers
    event_consumer.register_handler(
        "message.received",
        handle_message_received,
    )

    # Start consuming events (background task)
    asyncio.create_task(event_consumer.start_async())

    print("✅ Ventros AI Agents started")

@app.post("/api/agents/message")
async def handle_direct_message(request: Dict):
    """
    Direct HTTP endpoint for testing
    (Production usa event-driven via RabbitMQ)
    """
    # Create session
    session = Session(state=request.get("session_state", {}))

    # Get coordinator
    coordinator = agent_factory.create_coordinator()

    # Execute
    response = await coordinator.run_async(
        user_input=request["message"],
        session=session,
    )

    return {
        "response": response.output,
        "agent_used": response.agent_name,
        "confidence": response.confidence,
    }

@app.websocket("/ws/agent")
async def websocket_endpoint(websocket: WebSocket):
    """
    WebSocket for real-time agent interaction
    """
    await websocket.accept()

    # Create session
    session = Session()

    # Get coordinator
    coordinator = agent_factory.create_coordinator()

    try:
        while True:
            # Receive message
            data = await websocket.receive_json()

            # Execute agent
            response = await coordinator.run_async(
                user_input=data["message"],
                session=session,
            )

            # Send response
            await websocket.send_json({
                "response": response.output,
                "agent_used": response.agent_name,
            })

    except Exception as e:
        print(f"WebSocket error: {e}")
        await websocket.close()

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

---

## 📦 REQUIREMENTS & SETUP

### **requirements.txt**

```txt
# ADK
google-cloud-adk==0.5.0
vertexai==1.60.0

# gRPC
grpcio==1.62.0
grpcio-tools==1.62.0
protobuf==4.25.0

# Event Bus
pika==1.3.2

# Web Framework
fastapi==0.110.0
uvicorn==0.29.0
websockets==12.0

# Observability
opentelemetry-api==1.24.0
opentelemetry-sdk==1.24.0
opentelemetry-exporter-otlp==1.24.0
prometheus-client==0.20.0

# Semantic Router
semantic-router==0.0.20

# Utils
python-dotenv==1.0.1
pydantic==2.6.0
asyncio==3.4.3
```

### **.env**

```bash
# Vertex AI
GOOGLE_CLOUD_PROJECT=your-project-id
GOOGLE_APPLICATION_CREDENTIALS=path/to/credentials.json

# Go Memory Service
MEMORY_SERVICE_GRPC_HOST=localhost:50051

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672

# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317

# Agent Config
AGENT_DEFAULT_MODEL=gemini-2.0-flash
AGENT_MAX_TOKENS=8192
AGENT_TEMPERATURE=0.7
```

---

## ✅ RESUMO EXECUTIVO

### **O que este serviço faz:**

1. **Consome eventos** do RabbitMQ (mensagens inbound)
2. **Classifica intent** com Semantic Router
3. **Orquestra agents** (Coordinator → Specialists)
4. **Busca contexto** no Go Memory Service (gRPC)
5. **Processa com LLM** + Tools
6. **Publica resposta** no RabbitMQ (mensagens outbound)
7. **Observa tudo** com OpenTelemetry + Prometheus

### **Agents disponíveis:**

- **Coordinator**: Roteia para especialistas
- **Sales**: Prospecting, Negotiation, Closing
- **Retention**: Churn, Upsell, Winback
- **Support**: Technical, Billing, Onboarding
- **Operations**: Followup, Schedule, QA

### **Patterns usados:**

- ✅ Coordinator-Worker (primary pattern)
- ✅ Handoff (escalation)
- ✅ Reflection (self-critique)
- ✅ Sequential (onboarding pipeline)
- ✅ Parallel (lead enrichment)
- ✅ Loop (QA refinement)

### **Production-ready features:**

- ✅ Event-driven async (RabbitMQ)
- ✅ gRPC communication (Go Memory)
- ✅ OpenTelemetry tracing
- ✅ Prometheus metrics
- ✅ Callback instrumentation
- ✅ Error handling & retries
- ✅ WebSocket support
- ✅ Agent factory pattern
- ✅ Dependency injection

---

## 🔭 PHOENIX OBSERVABILITY (2025)

### **Por que Phoenix em vez de apenas OpenTelemetry?**

Phoenix (Arize AI) é uma plataforma moderna de observabilidade específica para LLMs e AI agents:

```
OpenTelemetry (Generic)         Phoenix (AI-Native)
├─ Traces genéricos             ├─ LLM-aware traces
├─ Logs estruturados            ├─ Prompt/response tracking
├─ Métricas básicas             ├─ Token cost analysis
└─ Spans personalizados         ├─ Embedding similarity
                                 ├─ Hallucination detection
                                 ├─ Agent conversation flows
                                 └─ Tool call visualization
```

**Stack Completo:**
- **Phoenix**: LLM observability (prompts, responses, embeddings, hallucinations)
- **OpenTelemetry**: Infrastructure traces (gRPC, RabbitMQ, DB)
- **Prometheus**: Metrics (agent duration, errors, throughput)

### **Phoenix Setup**

```python
import phoenix as px
from phoenix.trace import using_project
from phoenix.trace.langchain import LangChainInstrumentor
from phoenix.trace.openai import OpenAIInstrumentor
from openinference.instrumentation.vertexai import VertexAIInstrumentor

# Launch Phoenix server (runs locally on port 6006)
session = px.launch_app()
print(f"📊 Phoenix UI: http://localhost:6006")

# Auto-instrument Vertex AI
VertexAIInstrumentor().instrument()

# Configure project
px.Client().create_project("ventros-crm-agents")

class PhoenixObservability:
    """
    Phoenix-based observability for ADK agents

    Features:
    - Automatic LLM call tracking (prompts + completions)
    - Token usage & cost tracking
    - Embedding vector tracking
    - Agent conversation flows
    - Tool call visualization
    - Hallucination detection
    - Latency analysis
    """

    def __init__(self, project_name: str = "ventros-crm-agents"):
        self.project_name = project_name
        self.client = px.Client()

    @contextmanager
    def trace_agent_execution(
        self,
        agent_name: str,
        session_id: str,
        contact_id: str,
    ):
        """
        Context manager for tracing complete agent execution

        Usage:
        with phoenix.trace_agent_execution("retention_churn", session_id, contact_id):
            response = agent.run(user_input, session)
        """
        with using_project(self.project_name):
            # Phoenix automatically tracks:
            # - All LLM calls (prompts + completions)
            # - All embeddings (search_memory)
            # - All tool calls
            # - Latencies
            # - Token counts

            span_attributes = {
                "agent.name": agent_name,
                "session.id": session_id,
                "contact.id": contact_id,
            }

            with px.span(
                name=f"agent.{agent_name}",
                attributes=span_attributes,
            ) as span:
                yield span

    def log_memory_search(
        self,
        query: str,
        results: List[Dict],
        retrieval_strategy: str,
    ):
        """
        Log memory search with embedding vectors

        Phoenix visualizes:
        - Query embedding vs result embeddings (UMAP projection)
        - Similarity scores
        - Retrieved context
        """
        px.log_retrievals(
            project_name=self.project_name,
            query_text=query,
            documents=[r["text"] for r in results],
            document_scores=[r["score"] for r in results],
            metadata={
                "retrieval_strategy": retrieval_strategy,
            },
        )

    def log_llm_interaction(
        self,
        agent_name: str,
        prompt: str,
        completion: str,
        model: str,
        tokens_used: int,
        latency_ms: float,
    ):
        """
        Log LLM interaction (Phoenix auto-captures, mas pode customizar)
        """
        px.log_evaluations(
            project_name=self.project_name,
            model_name=model,
            prompt=prompt,
            completion=completion,
            metadata={
                "agent_name": agent_name,
                "tokens_used": tokens_used,
                "latency_ms": latency_ms,
            },
        )

    def detect_hallucinations(
        self,
        agent_response: str,
        memory_context: str,
    ) -> Dict:
        """
        Detecta alucinações comparando resposta com contexto

        Phoenix tem built-in hallucination detection usando:
        - Semantic similarity
        - Fact verification
        - Context grounding
        """
        # Phoenix evaluator
        from phoenix.evals import llm_classify, HallucinationEvaluator

        evaluator = HallucinationEvaluator()
        result = evaluator.evaluate(
            input=agent_response,
            reference=memory_context,
        )

        return {
            "is_hallucination": result.label == "hallucinated",
            "confidence": result.score,
            "explanation": result.explanation,
        }

# Usage in Agent Callbacks
class PhoenixAgentCallbacks:
    """ADK callbacks integrated with Phoenix"""

    def __init__(self, phoenix: PhoenixObservability):
        self.phoenix = phoenix

    def before_agent(self, agent_name: str, session: Session):
        # Phoenix trace context
        self.trace_context = self.phoenix.trace_agent_execution(
            agent_name=agent_name,
            session_id=session.state.get("session_id"),
            contact_id=session.state.get("contact_id"),
        )
        self.trace_context.__enter__()

    def after_agent(self, agent_name: str, response: any):
        # Exit Phoenix trace
        if hasattr(self, 'trace_context'):
            self.trace_context.__exit__(None, None, None)

    def after_tool_call(self, tool_name: str, tool_result: any):
        # Phoenix automatically captures tool calls
        # Can add custom metadata
        if tool_name == "search_memory":
            self.phoenix.log_memory_search(
                query=tool_result.get("query"),
                results=tool_result.get("results", []),
                retrieval_strategy=tool_result.get("strategy"),
            )
```

### **Phoenix Dashboard Views**

Phoenix fornece dashboards automáticos:

1. **Agent Flow Visualization**
   - Waterfall de agent → sub-agent → tool calls
   - Latência por componente
   - Success/failure rates

2. **LLM Analytics**
   - Token usage por agent
   - Cost tracking (Gemini pricing)
   - Response quality scores
   - Hallucination rates

3. **Embedding Space**
   - UMAP projection de todos embeddings
   - Cluster de sessões similares
   - Retrieval quality (query vs retrieved docs)

4. **Conversation Inspector**
   - Full conversation history
   - Context window utilization
   - Memory retrieval effectiveness

### **Production Phoenix Deployment**

```yaml
# docker-compose.yml
services:
  phoenix:
    image: arizephoenix/phoenix:latest
    ports:
      - "6006:6006"  # UI
      - "4317:4317"  # OTLP receiver
    environment:
      - PHOENIX_SQL_DATABASE_URL=postgresql://user:pass@postgres/phoenix
    volumes:
      - phoenix_data:/phoenix
    restart: always

  ventros-ai:
    build: ./ventros-ai
    environment:
      - PHOENIX_COLLECTOR_ENDPOINT=http://phoenix:4317
      - PHOENIX_PROJECT_NAME=ventros-crm-agents
    depends_on:
      - phoenix
```

---

## 🏗️ AGENT ENTITY CREATION ARCHITECTURE

### **Decisão Arquitetural: Quem Cria o Quê?**

```
┌─────────────────────────────────────────────────────┐
│  GO (Ventros CRM) - OWNS AGENT ENTITIES             │
│  ✅ Domain: Agent aggregate (agent.go)              │
│  ✅ Persistence: agents table                        │
│  ✅ CRUD: Create/Update/Delete agent records        │
│  ✅ Metadata: AIAgentMetadata, KnowledgeScope       │
│  ✅ Registry: Agent discovery & routing             │
└─────────────────────────────────────────────────────┘
                    │
                    │ gRPC: GetAgent(agent_id)
                    ▼
┌─────────────────────────────────────────────────────┐
│  PYTHON ADK - ORCHESTRATES BEHAVIOR                 │
│  ✅ Agent runtime behavior (LlmAgent instances)     │
│  ✅ Multi-agent orchestration patterns              │
│  ✅ LLM interaction & tool execution                │
│  ✅ Memory service integration                       │
│  ✅ NO persistence (stateless behavior layer)       │
└─────────────────────────────────────────────────────┘
```

**Princípio:** Go é o **source of truth** para entidades, Python é **behavior orchestrator**

### **Agent Templates (Python → Go)**

Python expõe templates de agentes genéricos que Go pode usar:

```python
# ventros-ai/agent_templates.py

from dataclasses import dataclass
from typing import List, Dict, Optional
from enum import Enum

class AgentTemplate(Enum):
    """
    Templates de agentes pré-configurados

    Go pode instanciar esses templates criando agent records
    com metadata correspondente
    """

    # Sales
    SALES_PROSPECTING = "sales_prospecting"
    SALES_NEGOTIATION = "sales_negotiation"
    SALES_CLOSING = "sales_closing"

    # Retention
    RETENTION_CHURN = "retention_churn"
    RETENTION_UPSELL = "retention_upsell"
    RETENTION_WINBACK = "retention_winback"

    # Support
    SUPPORT_TECHNICAL = "support_technical"
    SUPPORT_BILLING = "support_billing"
    SUPPORT_ONBOARDING = "support_onboarding"

    # Operations
    OPERATIONS_FOLLOWUP = "operations_followup"
    OPERATIONS_SCHEDULE = "operations_schedule"
    OPERATIONS_QA = "operations_qa"

    # Marketing
    MARKETING_CAMPAIGN = "marketing_campaign"
    MARKETING_CONTENT = "marketing_content"
    MARKETING_EVENT = "marketing_event"

    # Generic
    BALANCED = "balanced"

@dataclass
class AgentTemplateConfig:
    """
    Configuração completa de um agent template

    Go usa isso para criar agent records com metadata correto
    """
    template_id: str
    name: str
    description: str
    category: str

    # Memory configuration
    knowledge_scope: Dict  # KnowledgeScope parameters
    retrieval_strategy: str  # ex: "retention_churn", "balanced"

    # Behavior configuration
    instruction_prompt: str
    model_name: str
    temperature: float
    max_tokens: int

    # Tools & capabilities
    required_tools: List[str]
    optional_tools: List[str]

    # Routing
    intent_examples: List[str]  # Para semantic router

class AgentTemplateRegistry:
    """
    Registry de todos agent templates disponíveis

    Exposto via gRPC para Go consultar
    """

    @staticmethod
    def get_all_templates() -> List[AgentTemplateConfig]:
        """Retorna todos templates disponíveis"""
        return [
            AgentTemplateRegistry.get_template(template)
            for template in AgentTemplate
        ]

    @staticmethod
    def get_template(template: AgentTemplate) -> AgentTemplateConfig:
        """Retorna configuração de um template específico"""

        templates = {
            AgentTemplate.RETENTION_CHURN: AgentTemplateConfig(
                template_id="retention_churn",
                name="Retention & Churn Prevention",
                description="Especialista em prevenir cancelamentos e retenção de clientes",
                category="retention",

                knowledge_scope={
                    "lookback_days": 90,
                    "include_sessions": True,
                    "include_contact_events": True,
                    "include_agent_transfers": True,
                    "include_campaigns": False,
                },

                retrieval_strategy="retention_churn",  # 50% vector, 20% keyword, 20% graph

                instruction_prompt="""
                Você é um especialista em retenção de clientes.

                OBJETIVO: Prevenir churn e manter clientes satisfeitos.

                PROTOCOL:
                1. SEMPRE busque contexto na memória primeiro
                2. Se sentiment < -0.5: escale imediatamente
                3. Se mencionar "cancelar": ofereça retenção
                4. Máximo 3 tentativas de retenção

                CONSTRAINTS:
                - Nunca prometa algo impossível
                - Máximo 30% desconto (acima disso: supervisor)
                - Respeite decisão após 3 rejeições
                """,

                model_name="gemini-2.0-flash",
                temperature=0.7,
                max_tokens=2048,

                required_tools=[
                    "search_memory",
                    "create_retention_offer",
                    "escalate_to_supervisor",
                ],

                optional_tools=[
                    "check_customer_value",
                    "get_past_offers",
                ],

                intent_examples=[
                    "quero cancelar",
                    "muito caro",
                    "insatisfeito",
                    "não quero mais",
                    "esperava mais",
                ],
            ),

            AgentTemplate.SALES_PROSPECTING: AgentTemplateConfig(
                template_id="sales_prospecting",
                name="Sales Prospecting",
                description="Especialista em qualificação de leads e prospecção",
                category="sales",

                knowledge_scope={
                    "lookback_days": 30,
                    "include_tracking": True,  # UTM, campaign source
                    "include_contact_events": True,
                    "include_pipeline": True,
                },

                retrieval_strategy="sales_prospecting",  # 20% vector, 30% keyword, 40% graph

                instruction_prompt="""
                Você é um especialista em prospecção e qualificação de leads.

                OBJETIVO: Qualificar leads e avançar pipeline.

                PROTOCOL:
                1. Busque origem do lead (UTM, campaign)
                2. Qualifique usando BANT (Budget, Authority, Need, Timeline)
                3. Avance pipeline se qualificado

                CONSTRAINTS:
                - Não seja pushy
                - Eduque antes de vender
                - Respeite timing do lead
                """,

                model_name="gemini-2.0-flash",
                temperature=0.8,
                max_tokens=1024,

                required_tools=[
                    "search_memory",
                    "update_pipeline_stage",
                    "create_follow_up_task",
                ],

                optional_tools=[
                    "get_campaign_info",
                    "check_similar_leads",
                ],

                intent_examples=[
                    "quanto custa",
                    "quero saber preços",
                    "planos disponíveis",
                    "orçamento",
                ],
            ),

            AgentTemplate.SUPPORT_TECHNICAL: AgentTemplateConfig(
                template_id="support_technical",
                name="Technical Support",
                description="Especialista em suporte técnico e resolução de bugs",
                category="support",

                knowledge_scope={
                    "lookback_days": 7,
                    "include_sessions": True,
                    "include_contact_events": False,
                },

                retrieval_strategy="support_technical",  # 30% vector, 50% keyword, 10% graph

                instruction_prompt="""
                Você é um especialista em suporte técnico.

                OBJETIVO: Resolver problemas técnicos rapidamente.

                PROTOCOL:
                1. Identifique o problema exato
                2. Busque casos similares na memória
                3. Forneça solução passo-a-passo
                4. Se não resolver: escale

                CONSTRAINTS:
                - Seja técnico mas claro
                - Screenshots ajudam
                - Sempre confirme resolução
                """,

                model_name="gemini-2.0-flash",
                temperature=0.5,
                max_tokens=2048,

                required_tools=[
                    "search_memory",
                    "create_ticket",
                    "escalate_to_engineering",
                ],

                optional_tools=[
                    "check_system_status",
                    "get_error_logs",
                ],

                intent_examples=[
                    "não funciona",
                    "deu erro",
                    "bug",
                    "problema técnico",
                ],
            ),

            AgentTemplate.BALANCED: AgentTemplateConfig(
                template_id="balanced",
                name="Balanced Agent",
                description="Agente genérico para casos não especializados",
                category="general",

                knowledge_scope={
                    "lookback_days": 30,
                    "include_sessions": True,
                    "include_contact_events": True,
                },

                retrieval_strategy="balanced",  # 33% vector, 33% keyword, 33% graph

                instruction_prompt="""
                Você é um agente versátil do Ventros CRM.

                OBJETIVO: Atender o cliente da melhor forma possível.

                PROTOCOL:
                1. Busque contexto na memória
                2. Seja profissional e prestativo
                3. Escale se necessário
                """,

                model_name="gemini-2.0-flash",
                temperature=0.7,
                max_tokens=1024,

                required_tools=["search_memory"],
                optional_tools=["escalate_to_human"],

                intent_examples=[],  # Fallback
            ),
        }

        return templates.get(template, templates[AgentTemplate.BALANCED])

# gRPC Service para Go consultar templates
class AgentTemplateService:
    """
    Expõe agent templates via gRPC para Go

    Go chama isso quando admin quer criar novo agent
    """

    def ListAgentTemplates(
        self,
        request: ListAgentTemplatesRequest,
        context,
    ) -> ListAgentTemplatesResponse:
        """Lista todos agent templates disponíveis"""
        templates = AgentTemplateRegistry.get_all_templates()

        return ListAgentTemplatesResponse(
            templates=[
                AgentTemplateProto(
                    template_id=t.template_id,
                    name=t.name,
                    description=t.description,
                    category=t.category,
                    retrieval_strategy=t.retrieval_strategy,
                    intent_examples=t.intent_examples,
                )
                for t in templates
            ]
        )

    def GetAgentTemplate(
        self,
        request: GetAgentTemplateRequest,
        context,
    ) -> AgentTemplateConfig:
        """Retorna configuração completa de um template"""
        template = AgentTemplate(request.template_id)
        return AgentTemplateRegistry.get_template(template)
```

### **Go Side: Creating Agents from Templates**

```go
// Go chama Python para listar templates
func (s *AgentService) ListAvailableAgentTemplates(
    ctx context.Context,
) ([]AgentTemplate, error) {
    // gRPC call to Python
    resp, err := s.pythonClient.ListAgentTemplates(ctx, &pb.ListAgentTemplatesRequest{})
    if err != nil {
        return nil, err
    }

    return resp.Templates, nil
}

// Admin cria agent baseado em template
func (s *AgentService) CreateAgentFromTemplate(
    ctx context.Context,
    templateID string,
    customName string,
    projectID uuid.UUID,
) (*domain.Agent, error) {
    // 1. Get template config from Python
    templateConfig, err := s.pythonClient.GetAgentTemplate(ctx, &pb.GetAgentTemplateRequest{
        TemplateId: templateID,
    })
    if err != nil {
        return nil, err
    }

    // 2. Create agent entity in Go
    agent := domain.NewAgent(
        customName,
        domain.AgentTypeAI,
        projectID,
    )

    // 3. Set AI metadata from template
    agent.SetAIMetadata(domain.AIAgentMetadata{
        Category:          templateConfig.Category,
        Instructions:      templateConfig.InstructionPrompt,
        ModelName:         templateConfig.ModelName,
        Temperature:       templateConfig.Temperature,
        MaxTokens:         templateConfig.MaxTokens,
        KnowledgeScope:    convertKnowledgeScope(templateConfig.KnowledgeScope),
        MemoryStrategy:    convertMemoryStrategy(templateConfig.RetrievalStrategy),
        Skills:            templateConfig.RequiredTools,
        IntentExamples:    templateConfig.IntentExamples,
    })

    // 4. Persist to DB
    if err := s.agentRepo.Save(ctx, agent); err != nil {
        return nil, err
    }

    return agent, nil
}
```

---

## ⚙️ TEMPORAL WORKFLOW INTEGRATION

### **Por que Temporal com ADK?**

**Temporal** é ideal para:
- ✅ Long-running agent workflows (multi-dia, multi-step)
- ✅ Saga patterns (compensation em caso de falha)
- ✅ Scheduled executions (follow-ups automáticos)
- ✅ Durable state (surviving crashes/deploys)
- ✅ Human-in-the-loop patterns (approval flows)

**ADK** é ideal para:
- ✅ Real-time agent interaction (< 5s response)
- ✅ Multi-agent orchestration (coordinator + specialists)
- ✅ Tool calling & LLM reasoning
- ✅ Memory-aware responses

**Combinados:**
- Temporal orquestra **workflows complexos**
- ADK fornece **comportamento inteligente** em cada step

### **Arquitetura Temporal + ADK**

```
┌──────────────────────────────────────────────────────┐
│  TEMPORAL (Workflow Orchestration)                    │
│                                                        │
│  LeadNurturingWorkflow:                               │
│    Day 1: SendWelcomeEmail (activity)                │
│    Day 3: CheckEngagement (activity)                 │
│      └─> If no engagement: TriggerAIOutreach ──────┐│
│    Day 7: QualificationCall (ADK agent)           ││
│    Day 14: RetentionCheck (ADK agent)             ││
│                                                     ││
└─────────────────────────────────────────────────────┘│
                                                        │
                                                        ▼
┌──────────────────────────────────────────────────────┐
│  PYTHON ADK (Agent Runtime)                          │
│                                                        │
│  TriggerAIOutreach calls:                             │
│    → LeadEngagementAgent                              │
│      → Searches memory for lead behavior              │
│      → Crafts personalized outreach message           │
│      → Publishes SendMessage event                    │
│                                                        │
└──────────────────────────────────────────────────────┘
```

### **Temporal Workflows com ADK Activities**

```python
# ventros-ai/temporal_workflows.py

from temporalio import workflow, activity
from datetime import timedelta
from typing import Dict
import asyncio

@workflow.defn
class LeadNurturingWorkflow:
    """
    Long-running workflow para nutrição de leads

    Duração: 30 dias
    Steps:
    - Day 1: Welcome email
    - Day 3: Check engagement → AI outreach se inativo
    - Day 7: AI qualification call
    - Day 14: AI retention check
    - Day 30: Human handoff

    Temporal mantém estado durante 30 dias, sobrevive a deploys
    """

    @workflow.run
    async def run(self, lead_id: str, project_id: str) -> Dict:
        # Day 1: Welcome email (activity simples)
        await workflow.execute_activity(
            send_welcome_email,
            args=[lead_id],
            start_to_close_timeout=timedelta(minutes=5),
        )

        # Wait 3 days
        await asyncio.sleep(timedelta(days=3).total_seconds())

        # Day 3: Check engagement
        engagement_score = await workflow.execute_activity(
            check_engagement,
            args=[lead_id],
            start_to_close_timeout=timedelta(minutes=2),
        )

        # If low engagement: AI outreach
        if engagement_score < 5:
            await workflow.execute_activity(
                trigger_ai_outreach,
                args=[lead_id, "low_engagement"],
                start_to_close_timeout=timedelta(minutes=10),
            )

        # Wait 4 more days
        await asyncio.sleep(timedelta(days=4).total_seconds())

        # Day 7: AI qualification (ADK agent)
        qualification_result = await workflow.execute_activity(
            ai_lead_qualification,
            args=[lead_id],
            start_to_close_timeout=timedelta(minutes=15),
        )

        # If qualified: move pipeline
        if qualification_result.get("is_qualified"):
            await workflow.execute_activity(
                move_pipeline_stage,
                args=[lead_id, "qualified"],
                start_to_close_timeout=timedelta(minutes=5),
            )

        # Wait 7 more days
        await asyncio.sleep(timedelta(days=7).total_seconds())

        # Day 14: Retention check
        await workflow.execute_activity(
            ai_retention_check,
            args=[lead_id],
            start_to_close_timeout=timedelta(minutes=10),
        )

        # Wait 16 more days
        await asyncio.sleep(timedelta(days=16).total_seconds())

        # Day 30: Human handoff
        await workflow.execute_activity(
            assign_to_human_agent,
            args=[lead_id],
            start_to_close_timeout=timedelta(minutes=5),
        )

        return {"status": "completed", "lead_id": lead_id}

@activity.defn
async def trigger_ai_outreach(lead_id: str, reason: str) -> Dict:
    """
    Activity que chama ADK agent para outreach

    Temporal executa isso como activity (retriable, monitorable)
    ADK fornece comportamento inteligente
    """

    # Get agent factory
    agent_factory = get_agent_factory()

    # Create session context
    session = Session(state={
        "contact_id": lead_id,
        "agent_category": "sales_prospecting",
        "context": f"Lead with {reason}, needs re-engagement",
    })

    # Get agent
    agent = agent_factory.create_agent("sales_prospecting")

    # Execute agent (ADK handles reasoning + memory + tools)
    response = await agent.run_async(
        user_input="Analyze lead behavior and craft personalized re-engagement message",
        session=session,
    )

    # Publish outbound message event (Go will send via WAHA)
    await publish_outbound_message(
        contact_id=lead_id,
        text=response.output,
        source="sequence",
        metadata={"workflow": "lead_nurturing", "reason": reason},
    )

    return {
        "message_sent": True,
        "agent_response": response.output,
    }

@activity.defn
async def ai_lead_qualification(lead_id: str) -> Dict:
    """
    Activity que usa ADK para qualificar lead

    ADK agent analisa:
    - Histórico de interações
    - Engagement metrics
    - Pipeline fit
    - BANT criteria
    """

    agent_factory = get_agent_factory()

    session = Session(state={
        "contact_id": lead_id,
        "agent_category": "sales_prospecting",
    })

    agent = agent_factory.create_agent("sales_prospecting")

    # Agent calls search_memory, analyzes, returns structured result
    response = await agent.run_async(
        user_input="""
        Analyze this lead and determine qualification:

        BANT Criteria:
        - Budget: Can they afford our solution?
        - Authority: Are they decision maker?
        - Need: Do they have clear pain point?
        - Timeline: Ready to buy soon?

        Return JSON: {"is_qualified": bool, "score": 0-10, "reasoning": "..."}
        """,
        session=session,
    )

    # Parse structured output
    import json
    result = json.loads(response.output)

    return result
```

### **Saga Pattern com Temporal + ADK**

```python
@workflow.defn
class CustomerOnboardingSaga:
    """
    Saga pattern para onboarding complexo

    Steps (com compensation):
    1. CreateAccount → [compensation: DeleteAccount]
    2. SetupBilling → [compensation: CancelBilling]
    3. AIWelcome → [compensation: SendApology]
    4. ProvisionResources → [compensation: DeprovisionResources]
    5. ActivateSubscription → [compensation: SuspendSubscription]

    Se qualquer step falhar: Temporal executa compensations em ordem reversa
    """

    def __init__(self):
        self.saga_state = {
            "completed_steps": [],
            "compensations": [],
        }

    @workflow.run
    async def run(self, customer_id: str) -> Dict:
        try:
            # Step 1: Create account
            account_id = await self._execute_with_compensation(
                forward=create_account_activity,
                compensation=delete_account_activity,
                args=[customer_id],
            )

            # Step 2: Setup billing
            billing_id = await self._execute_with_compensation(
                forward=setup_billing_activity,
                compensation=cancel_billing_activity,
                args=[customer_id, account_id],
            )

            # Step 3: AI Welcome (ADK agent)
            await self._execute_with_compensation(
                forward=ai_welcome_activity,
                compensation=send_apology_activity,
                args=[customer_id],
            )

            # Step 4: Provision resources
            resources = await self._execute_with_compensation(
                forward=provision_resources_activity,
                compensation=deprovision_resources_activity,
                args=[customer_id, account_id],
            )

            # Step 5: Activate subscription
            await self._execute_with_compensation(
                forward=activate_subscription_activity,
                compensation=suspend_subscription_activity,
                args=[customer_id, billing_id],
            )

            return {"status": "success", "customer_id": customer_id}

        except Exception as e:
            # Algo falhou: execute compensations em ordem reversa
            await self._compensate_all()
            return {"status": "failed", "error": str(e)}

    async def _execute_with_compensation(
        self,
        forward: callable,
        compensation: callable,
        args: list,
    ):
        """Executa activity e registra compensation"""
        result = await workflow.execute_activity(
            forward,
            args=args,
            start_to_close_timeout=timedelta(minutes=5),
        )

        # Register compensation
        self.saga_state["completed_steps"].append(forward.__name__)
        self.saga_state["compensations"].insert(0, (compensation, args))

        return result

    async def _compensate_all(self):
        """Executa todas compensations em ordem reversa"""
        for compensation, args in self.saga_state["compensations"]:
            try:
                await workflow.execute_activity(
                    compensation,
                    args=args,
                    start_to_close_timeout=timedelta(minutes=5),
                )
            except Exception as e:
                workflow.logger.error(f"Compensation failed: {e}")

@activity.defn
async def ai_welcome_activity(customer_id: str) -> None:
    """
    ADK agent envia welcome personalizado

    Se isso falhar, compensation envia apology
    """
    agent_factory = get_agent_factory()
    agent = agent_factory.create_agent("support_onboarding")

    session = Session(state={
        "contact_id": customer_id,
        "agent_category": "support_onboarding",
    })

    response = await agent.run_async(
        user_input="Send personalized welcome message to new customer",
        session=session,
    )

    await publish_outbound_message(
        contact_id=customer_id,
        text=response.output,
        source="system",
    )

@activity.defn
async def send_apology_activity(customer_id: str) -> None:
    """Compensation: se welcome falhou, envia apology"""
    await publish_outbound_message(
        contact_id=customer_id,
        text="Desculpe, tivemos um problema técnico. Nossa equipe entrará em contato em breve.",
        source="system",
    )
```

### **Integration: Temporal + RabbitMQ + Outbox**

```python
"""
Como Temporal, RabbitMQ, e Outbox coexistem:

1. TEMPORAL: Workflows de longo prazo (dias/semanas)
   - Lead nurturing
   - Customer onboarding
   - Retention campaigns
   - Scheduled follow-ups

2. RABBITMQ: Real-time events (segundos)
   - MessageReceived → AI response imediata
   - SessionEnded → Update stats
   - ContactCreated → Trigger workflow

3. OUTBOX: Garantia de entrega (at-least-once)
   - Go persiste eventos no outbox table
   - Background worker publica no RabbitMQ
   - Retry automático se RabbitMQ estiver down

INTEGRAÇÃO:

┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│  Go CRM API  │────────►│   Outbox     │────────►│  RabbitMQ    │
│              │ persist │   Table      │ publish │              │
└──────────────┘         └──────────────┘         └───┬──────────┘
                                                       │
                        ┌──────────────────────────────┤
                        │                              │
                        ▼                              ▼
                 ┌──────────────┐            ┌──────────────┐
                 │   Temporal   │            │  Python ADK  │
                 │   Workflows  │            │    Agents    │
                 └──────┬───────┘            └──────────────┘
                        │
                        │ (Temporal worker consome eventos específicos)
                        │
                        ▼
                 Start workflow se evento é:
                 - ContactCreated → LeadNurturingWorkflow
                 - SubscriptionCreated → OnboardingSaga
                 - ChurnRiskDetected → RetentionWorkflow
```

**Padrão de Integração:**

```python
# Temporal worker consome eventos RabbitMQ para iniciar workflows

class TemporalWorkflowStarter:
    """
    Consome eventos RabbitMQ e inicia Temporal workflows

    Routing:
    - ContactCreated → LeadNurturingWorkflow
    - SubscriptionCreated → CustomerOnboardingSaga
    - ChurnRiskDetected → RetentionCampaignWorkflow
    """

    def __init__(
        self,
        temporal_client: Client,
        rabbitmq_consumer: EventConsumer,
    ):
        self.temporal_client = temporal_client
        self.rabbitmq_consumer = rabbitmq_consumer

        # Register handlers
        rabbitmq_consumer.register_handler(
            "contact.created",
            self.on_contact_created,
        )
        rabbitmq_consumer.register_handler(
            "subscription.created",
            self.on_subscription_created,
        )

    async def on_contact_created(self, event: Dict):
        """
        Quando contact é criado, inicia workflow de nurturing

        Temporal workflow ID = "lead-nurturing-{contact_id}"
        Isso garante idempotência (não duplica workflows)
        """
        contact_id = event["payload"]["contact_id"]
        project_id = event["payload"]["project_id"]

        # Start Temporal workflow
        await self.temporal_client.start_workflow(
            LeadNurturingWorkflow.run,
            args=[contact_id, project_id],
            id=f"lead-nurturing-{contact_id}",  # Idempotent
            task_queue="ventros-workflows",
        )

    async def on_subscription_created(self, event: Dict):
        """Inicia onboarding saga"""
        customer_id = event["payload"]["customer_id"]

        await self.temporal_client.start_workflow(
            CustomerOnboardingSaga.run,
            args=[customer_id],
            id=f"onboarding-{customer_id}",
            task_queue="ventros-workflows",
        )

# Main app integra tudo
@app.on_event("startup")
async def startup():
    # 1. Initialize Temporal client
    temporal_client = await Client.connect("localhost:7233")

    # 2. Start Temporal worker (executa workflows + activities)
    worker = Worker(
        temporal_client,
        task_queue="ventros-workflows",
        workflows=[
            LeadNurturingWorkflow,
            CustomerOnboardingSaga,
            RetentionCampaignWorkflow,
        ],
        activities=[
            trigger_ai_outreach,
            ai_lead_qualification,
            ai_welcome_activity,
            # ... all activities
        ],
    )
    asyncio.create_task(worker.run())

    # 3. Start RabbitMQ consumer (eventos real-time)
    event_consumer = EventConsumer(rabbitmq_url=RABBITMQ_URL)
    event_consumer.register_handler("message.received", handle_message_received)
    asyncio.create_task(event_consumer.start_async())

    # 4. Start workflow starter (RabbitMQ → Temporal)
    workflow_starter = TemporalWorkflowStarter(temporal_client, event_consumer)

    print("✅ All systems operational")
    print("  - Temporal worker running")
    print("  - RabbitMQ consumer running")
    print("  - ADK agents ready")
    print("  - Phoenix observability: http://localhost:6006")
```

---

**Próximos passos:** Implemente gradualmente, começando pelo CoordinatorAgent + 1 specialist (RetentionChurnAgent). Depois adicione outros specialists conforme necessário.

---

## 🔌 MCP (MODEL CONTEXT PROTOCOL) INTEGRATION

### **O que é MCP?**

**MCP (Model Context Protocol)** é um protocolo aberto da Anthropic para conectar AI agents a ferramentas e dados externos de forma padronizada.

```
Traditional Approach:        MCP Approach:
┌─────────────────┐         ┌─────────────────┐
│  Python Agent   │         │  Python Agent   │
│  └─ Tool 1      │         │  └─ MCPToolset  │
│  └─ Tool 2      │         └────────┬────────┘
│  └─ Tool 3      │                  │
│  └─ Tool 4      │                  │ MCP Protocol
└─────────────────┘                  │
                                     ▼
                             ┌────────────────┐
                             │   MCP Server   │
                             │  (Go Backend)  │
                             │                │
                             │  ✅ Tool 1-8   │
                             │  ✅ Cached     │
                             │  ✅ Secure     │
                             │  ✅ Versioned  │
                             └────────────────┘
```

**Benefícios:**
- ✅ **Reutilização**: Same tools across multiple agents
- ✅ **Centralização**: Single source of truth for business logic
- ✅ **Performance**: Server-side caching (5 min cache on queries)
- ✅ **Segurança**: Centralized authentication & authorization
- ✅ **Versionamento**: API versioning sem quebrar agents
- ✅ **Observabilidade**: Unified logging & metrics

### **Decision Tree: MCP vs Direct Tools**

```python
"""
Use MCP Tools when:
✅ Tool acessa dados do CRM (contacts, messages, pipelines)
✅ Tool é reutilizado por múltiplos agents
✅ Tool precisa de caching (queries de analytics)
✅ Tool tem side-effects críticos (update database)
✅ Tool precisa de autenticação/autorização

Use Direct ADK Tools when:
✅ Tool é específico de um agent (parsing, formatting)
✅ Tool é lightweight (string manipulation, calculations)
✅ Tool não acessa databases
✅ Tool é stateless e determinístico
✅ Tool precisa de máxima performance (<10ms)

Examples:
- get_leads_count → MCP (database query, cacheable)
- format_bullet_list → Direct (lightweight, agent-specific)
- update_pipeline_stage → MCP (critical side-effect)
- calculate_sentiment → Direct (stateless, fast)
- analyze_agent_messages → MCP (complex query, reusable)
- parse_date_string → Direct (simple, deterministic)
"""
```

### **MCPToolset Implementation**

```python
# ventros-ai/mcp_integration.py

from adk import Tool
from typing import Dict, List, Any
import grpc
from dataclasses import dataclass
import json

# Import gRPC generated code
import mcp_service_pb2
import mcp_service_pb2_grpc

@dataclass
class MCPToolConfig:
    """Configuration for an MCP tool"""
    name: str
    description: str
    parameters: Dict[str, Any]  # JSON schema

class MCPToolset:
    """
    ADK-compatible toolset that calls Go MCP Server via gRPC

    Usage:
    mcp_toolset = MCPToolset(grpc_host="localhost:50052")
    tools = mcp_toolset.get_all_tools()

    agent = LlmAgent(
        name="bi_manager",
        tools=tools,  # MCP tools
        memory_service=memory_service,
    )
    """

    def __init__(
        self,
        grpc_host: str = "localhost:50052",
        auth_token: str = None,
    ):
        self.grpc_host = grpc_host
        self.auth_token = auth_token
        self.channel = grpc.insecure_channel(grpc_host)
        self.stub = mcp_service_pb2_grpc.MCPServerStub(self.channel)

        # Cache tool definitions
        self._tools_cache: Dict[str, MCPToolConfig] = {}
        self._load_tool_definitions()

    def _load_tool_definitions(self):
        """Load tool definitions from MCP server"""
        try:
            request = mcp_service_pb2.ListToolsRequest()
            response = self.stub.ListTools(request)

            for tool_proto in response.tools:
                self._tools_cache[tool_proto.name] = MCPToolConfig(
                    name=tool_proto.name,
                    description=tool_proto.description,
                    parameters=json.loads(tool_proto.parameters_json),
                )
        except grpc.RpcError as e:
            print(f"Failed to load MCP tool definitions: {e}")

    def get_all_tools(self) -> List[Tool]:
        """
        Returns all MCP tools as ADK Tool objects

        Each tool wraps the MCP gRPC call
        """
        return [
            self._create_adk_tool(tool_name, tool_config)
            for tool_name, tool_config in self._tools_cache.items()
        ]

    def get_tool(self, tool_name: str) -> Tool:
        """Get specific MCP tool by name"""
        tool_config = self._tools_cache.get(tool_name)
        if not tool_config:
            raise ValueError(f"MCP tool not found: {tool_name}")

        return self._create_adk_tool(tool_name, tool_config)

    def _create_adk_tool(
        self,
        tool_name: str,
        tool_config: MCPToolConfig,
    ) -> Tool:
        """
        Creates ADK Tool that wraps MCP gRPC call

        The tool function calls Go MCP Server
        """

        async def tool_function(**kwargs) -> Dict[str, Any]:
            """
            Generic tool function that calls MCP server

            Args: Dynamic based on tool's JSON schema
            Returns: Tool result from Go MCP server
            """
            try:
                # Build MCP request
                request = mcp_service_pb2.ExecuteToolRequest(
                    tool_name=tool_name,
                    arguments_json=json.dumps(kwargs),
                    auth_token=self.auth_token or "",
                )

                # Call MCP server via gRPC
                response = self.stub.ExecuteTool(request)

                # Parse result
                result = json.loads(response.result_json)

                return result

            except grpc.RpcError as e:
                return {
                    "error": f"MCP tool call failed: {e}",
                    "tool_name": tool_name,
                }

        # Create ADK Tool
        return Tool(
            name=tool_name,
            description=tool_config.description,
            function=tool_function,
            # ADK will infer parameters from function signature & JSON schema
        )

    def execute_tool_direct(
        self,
        tool_name: str,
        arguments: Dict[str, Any],
    ) -> Dict[str, Any]:
        """
        Direct tool execution without ADK wrapper

        Useful for programmatic calls (not from LLM)
        """
        try:
            request = mcp_service_pb2.ExecuteToolRequest(
                tool_name=tool_name,
                arguments_json=json.dumps(arguments),
                auth_token=self.auth_token or "",
            )

            response = self.stub.ExecuteTool(request)

            return json.loads(response.result_json)

        except grpc.RpcError as e:
            return {"error": str(e)}

# Global MCP toolset instance
_mcp_toolset: MCPToolset = None

def get_mcp_toolset() -> MCPToolset:
    """Get or create global MCP toolset instance"""
    global _mcp_toolset
    if _mcp_toolset is None:
        _mcp_toolset = MCPToolset(
            grpc_host=os.getenv("MCP_SERVER_HOST", "localhost:50052"),
            auth_token=os.getenv("MCP_AUTH_TOKEN"),
        )
    return _mcp_toolset
```

---

## 📊 BI MANAGER AGENT (Analytics & Insights)

### **Purpose**

BI Manager Agent é um agente especializado que **conhece todas as conversas, agentes, métricas** e pode responder perguntas de business intelligence:

- "Quantos leads tive hoje?"
- "Qual agente converteu mais?"
- "PORQUE o João converteu mais que a Maria?" (análise profunda)

**Características:**
- ✅ Acessa dados via MCP tools (get_leads_count, get_agent_conversion_stats)
- ✅ Dynamic KnowledgeScope (muda escopo conforme pergunta)
- ✅ Retorna respostas com guia de formatação (tabelas, gráficos)
- ✅ Pode delegar para Agent Analyzer para análises qualitativas

### **Implementation**

```python
# ventros-ai/agents/bi_manager_agent.py

from adk import LlmAgent, AgentTool, Session
from vertexai.generative_models import GenerativeModel
from typing import Dict, List
from ..mcp_integration import get_mcp_toolset

class BIManagerAgent(LlmAgent):
    """
    Business Intelligence Manager Agent

    Capabilities:
    - Answer quantitative questions (leads, conversion rates, etc)
    - Answer qualitative questions (why João performs better)
    - Delegate to Agent Analyzer for deep quality analysis
    - Format responses with tables, charts, markdown

    Memory Strategy:
    - Dynamic KnowledgeScope based on question type
    - For "quantos leads": narrow scope (today)
    - For "PORQUE João": broad scope (90 days, all messages)

    Tools:
    - MCP: get_leads_count, get_agent_conversion_stats, get_top_performing_agent
    - MCP: analyze_agent_messages, compare_agents
    - Direct: format_table, format_chart_config
    """

    def __init__(
        self,
        memory_service,
        callback_manager,
    ):
        # Get MCP tools
        mcp_toolset = get_mcp_toolset()

        # MCP tools for BI queries
        bi_tools = [
            mcp_toolset.get_tool("get_leads_count"),
            mcp_toolset.get_tool("get_agent_conversion_stats"),
            mcp_toolset.get_tool("get_top_performing_agent"),
            mcp_toolset.get_tool("analyze_agent_messages"),
            mcp_toolset.get_tool("compare_agents"),
        ]

        # Direct tools for formatting
        direct_tools = [
            Tool(name="format_table", function=self._format_table),
            Tool(name="format_chart_config", function=self._format_chart_config),
        ]

        # Agent Analyzer as tool (for qualitative analysis)
        agent_analyzer = AgentAnalyzerAgent(memory_service, callback_manager)
        agent_tools = [
            AgentTool(
                agent=agent_analyzer,
                name="deep_agent_analysis",
                description="Delegate to Agent Analyzer for deep qualitative analysis (grammar, tone, brand)",
            ),
        ]

        super().__init__(
            name="bi_manager_agent",
            instruction=self._build_instruction(),
            model=GenerativeModel(
                "gemini-2.0-flash",
                generation_config={
                    "temperature": 0.3,  # Low temp for factual accuracy
                    "top_p": 0.95,
                    "max_output_tokens": 4096,
                },
            ),
            tools=bi_tools + direct_tools + agent_tools,
            memory_service=memory_service,
            callbacks=callback_manager.get_callbacks(),
        )

    def _build_instruction(self) -> str:
        return """
        # ROLE
        You are the Business Intelligence Manager for Ventros CRM.
        You have access to ALL conversations, agents, metrics, and can answer any BI question.

        # CAPABILITIES
        You can answer:

        ## Quantitative Questions (use MCP tools directly):
        - "Quantos leads tive hoje?" → get_leads_count(period="today")
        - "Qual agente converteu mais?" → get_agent_conversion_stats() + compare
        - "Qual a taxa de conversão por canal?" → aggregate stats

        ## Qualitative Questions (delegate to Agent Analyzer):
        - "PORQUE o João converteu mais?" → deep_agent_analysis(agent_id=João, compare_with=Maria)
          - Analyzes: grammar, tone, response time, empathy, brand alignment
          - Returns structured comparison

        # PROTOCOL

        ## Step 1: Classify Question Type

        ### If QUANTITATIVE (numbers, counts, rates):
        1. Identify time period (today, week, month, all-time)
        2. Call appropriate MCP tool
        3. Format result with format_table or format_chart_config
        4. Return with ResponseFormatGuide

        ### If QUALITATIVE (why, how, what makes X better):
        1. Identify agents to analyze
        2. Set broad KnowledgeScope (90 days, all messages)
        3. Delegate to deep_agent_analysis tool
        4. Summarize findings with formatting guide

        ### If MIXED (quantitative + qualitative):
        1. Answer quantitative part first (MCP tools)
        2. Then answer qualitative part (Agent Analyzer)
        3. Combine results with clear sections

        ## Step 2: Dynamic KnowledgeScope

        Adjust memory scope based on question:

        ```python
        # For "quantos leads hoje":
        knowledge_scope = {
            "lookback_days": 1,
            "include_sessions": False,  # Don't need conversation details
            "include_tracking": True,   # UTM for lead source
        }

        # For "PORQUE João":
        knowledge_scope = {
            "lookback_days": 90,         # Broader time window
            "include_sessions": True,     # Need conversation details
            "include_messages": True,     # Need message quality analysis
            "agent_ids": [joão_id, maria_id],  # Compare specific agents
        }
        ```

        ## Step 3: Response Formatting

        ALWAYS return ResponseFormatGuide with your answer:

        ```json
        {
            "answer": "Your answer text here",
            "format_guide": {
                "format": "markdown",
                "structure": {
                    "sections": [
                        {
                            "title": "Resumo Executivo",
                            "type": "text",
                            "styling": {"bold_first_line": true}
                        },
                        {
                            "title": "Métricas",
                            "type": "table",
                            "columns": ["Agente", "Conversões", "Taxa"],
                            "data": [[...]]
                        },
                        {
                            "title": "Análise Qualitativa",
                            "type": "bullets",
                            "styling": {"highlight_winners": true}
                        }
                    ]
                },
                "chart_config": {
                    "type": "bar",
                    "x_axis": "agent_name",
                    "y_axis": "conversion_rate",
                    "title": "Taxa de Conversão por Agente"
                }
            }
        }
        ```

        # EXAMPLES

        ## Example 1: Quantitative

        User: "Quantos leads tive hoje?"

        Agent:
        <thinking>
        - Question type: Quantitative
        - Time period: Today
        - Tool: get_leads_count(period="today")
        - Format: Simple text + number highlight
        </thinking>

        [Calls get_leads_count(period="today")]

        Result: {"total_leads": 47, "qualified_leads": 12, "period": "2025-01-15"}

        Response:
        {
            "answer": "Hoje você teve **47 leads**, sendo **12 qualificados**.",
            "format_guide": {
                "format": "markdown",
                "structure": {
                    "highlight_numbers": true,
                    "emphasis": "bold"
                }
            }
        }

        ## Example 2: Qualitative

        User: "PORQUE o João converteu mais que a Maria?"

        Agent:
        <thinking>
        - Question type: Qualitative (why)
        - Agents: João vs Maria
        - Need: Deep analysis (grammar, tone, brand)
        - Tool: deep_agent_analysis
        - Scope: Broad (90 days, all messages)
        </thinking>

        [Updates KnowledgeScope to broad]
        [Calls deep_agent_analysis(agent_ids=["joão", "maria"])]

        Result: {
            "winner": "joão",
            "dimensions": {
                "grammar": {"joão": 9.2, "maria": 7.8},
                "tone": {"joão": 8.9, "maria": 8.1},
                "response_time": {"joão": "2.3min", "maria": "5.7min"},
                "brand_alignment": {"joão": 9.5, "maria": 8.0}
            },
            "insights": [
                "João responds 2.5x faster (2.3 min vs 5.7 min)",
                "João has better grammar scores (9.2 vs 7.8)",
                "João's tone is more aligned with brand (9.5 vs 8.0)"
            ]
        }

        Response:
        {
            "answer": "João converteu mais por 3 razões principais:\\n\\n**1. Tempo de Resposta**\\nJoão responde 2.5x mais rápido (2.3 min vs 5.7 min de Maria)\\n\\n**2. Qualidade Gramatical**\\nJoão tem pontuação 9.2/10 vs 7.8/10 de Maria\\n\\n**3. Alinhamento com Marca**\\nJoão está mais alinhado com tom da marca (9.5 vs 8.0)",
            "format_guide": {
                "format": "markdown",
                "structure": {
                    "sections": [
                        {"title": "Resumo", "type": "text"},
                        {"title": "Comparação Detalhada", "type": "table"},
                        {"title": "Recomendações", "type": "bullets"}
                    ]
                },
                "chart_config": {
                    "type": "radar",
                    "dimensions": ["grammar", "tone", "response_time", "brand_alignment"],
                    "agents": ["João", "Maria"]
                }
            }
        }

        # CONSTRAINTS
        - ALWAYS cite data sources (tool results, date ranges)
        - NEVER make up numbers (only use tool results)
        - ALWAYS provide format_guide with response
        - For "why" questions: delegate to Agent Analyzer
        - Be concise but complete

        # TONE
        - Executive-friendly (non-technical)
        - Data-driven (cite numbers)
        - Action-oriented (provide insights, not just data)
        """

    def _format_table(self, columns: List[str], rows: List[List[Any]]) -> str:
        """Format data as markdown table"""
        # Header
        table = "| " + " | ".join(columns) + " |\\n"
        table += "| " + " | ".join(["---"] * len(columns)) + " |\\n"

        # Rows
        for row in rows:
            table += "| " + " | ".join(str(cell) for cell in row) + " |\\n"

        return table

    def _format_chart_config(
        self,
        chart_type: str,
        x_axis: str,
        y_axis: str,
        data: List[Dict],
        title: str = "",
    ) -> Dict:
        """
        Generate chart configuration for frontend

        Frontend will render using this config (Chart.js, etc)
        """
        return {
            "type": chart_type,  # bar, line, pie, radar
            "title": title,
            "x_axis": x_axis,
            "y_axis": y_axis,
            "data": data,
            "options": {
                "responsive": True,
                "plugins": {
                    "legend": {"display": True},
                    "tooltip": {"enabled": True},
                }
            }
        }

# Usage example
async def handle_bi_question(question: str, tenant_id: str, user_id: str):
    """
    Handle BI question from user

    Flow:
    1. Create session with minimal context (BI questions don't need full history)
    2. Execute BI Manager agent
    3. Parse response + format_guide
    4. Return to frontend
    """

    # Create session
    session = Session(state={
        "tenant_id": tenant_id,
        "user_id": user_id,
        "agent_category": "bi_manager",
        "knowledge_scope": "dynamic",  # Agent will adjust
    })

    # Get agent
    agent_factory = get_agent_factory()
    bi_agent = agent_factory.create_agent("bi_manager")

    # Execute
    response = await bi_agent.run_async(
        user_input=question,
        session=session,
    )

    # Parse response (agent returns JSON with answer + format_guide)
    import json
    result = json.loads(response.output)

    return {
        "answer": result["answer"],
        "format_guide": result["format_guide"],
        "agent_used": "bi_manager",
        "execution_time_ms": response.execution_time_ms,
    }
```

---

## 🎯 SDR AGENT (Sales Development Representative)

### **Purpose**

SDR Agent faz atendimento inicial, qualifica leads, e atribui para agentes humanos quando necessário.

**Capabilities:**
- ✅ Initial qualification (BANT criteria)
- ✅ Lead scoring
- ✅ Pipeline stage updates via MCP tools
- ✅ Assignment to human agents
- ✅ Automated follow-ups

### **Implementation**

```python
# ventros-ai/agents/sdr_agent.py

from adk import LlmAgent, Tool, Session
from vertexai.generative_models import GenerativeModel
from typing import Dict
from ..mcp_integration import get_mcp_toolset

class SDRAgent(LlmAgent):
    """
    Sales Development Representative Agent

    Responsibilities:
    - Initial contact with leads
    - Qualify using BANT (Budget, Authority, Need, Timeline)
    - Update pipeline stages
    - Assign to human agents when qualified
    - Schedule follow-ups

    Tools:
    - MCP: qualify_lead, update_pipeline_stage, assign_to_agent
    - MCP: get_agent_conversion_stats (to pick best human agent)
    - Direct: calculate_lead_score
    """

    def __init__(
        self,
        memory_service,
        callback_manager,
    ):
        # Get MCP tools
        mcp_toolset = get_mcp_toolset()

        mcp_tools = [
            mcp_toolset.get_tool("qualify_lead"),
            mcp_toolset.get_tool("update_pipeline_stage"),
            mcp_toolset.get_tool("assign_to_agent"),
            mcp_toolset.get_tool("get_agent_conversion_stats"),
        ]

        direct_tools = [
            Tool(name="calculate_lead_score", function=self._calculate_lead_score),
        ]

        super().__init__(
            name="sdr_agent",
            instruction=self._build_instruction(),
            model=GenerativeModel(
                "gemini-2.0-flash",
                generation_config={
                    "temperature": 0.7,
                    "top_p": 0.95,
                    "max_output_tokens": 2048,
                },
            ),
            tools=mcp_tools + direct_tools,
            memory_service=memory_service,
            callbacks=callback_manager.get_callbacks(),
        )

    def _build_instruction(self) -> str:
        return """
        # ROLE
        You are an SDR (Sales Development Representative) for Ventros CRM.
        You handle initial contact, qualify leads, and assign to human closers.

        # QUALIFICATION FRAMEWORK (BANT)

        ## B - Budget
        - Can they afford our solution?
        - What's their budget range?
        - Who controls budget decisions?

        ## A - Authority
        - Are they the decision maker?
        - If not, who is?
        - What's the approval process?

        ## N - Need
        - What problem are they trying to solve?
        - How urgent is it?
        - What's the cost of not solving it?

        ## T - Timeline
        - When do they need solution?
        - Any specific deadlines?
        - What's driving the timeline?

        # LEAD SCORING

        Calculate lead score (0-100) based on:
        - BANT qualification (40 points)
        - Engagement level (20 points)
        - Company size/revenue (20 points)
        - Source quality (10 points)
        - Speed of response (10 points)

        Thresholds:
        - 80+: Hot lead → Assign to top human agent immediately
        - 60-79: Warm lead → 1-2 more touches, then assign
        - 40-59: Cold lead → Nurture sequence
        - <40: Unqualified → Polite disqualification

        # PROTOCOL

        ## Step 1: Initial Contact
        1. Warm greeting (personalized, not robotic)
        2. Ask discovery questions
        3. Listen actively (search memory for context)

        ## Step 2: Qualification
        1. Assess each BANT dimension
        2. Calculate lead score (calculate_lead_score tool)
        3. Update pipeline stage (qualify_lead MCP tool)

        ## Step 3: Routing Decision

        ### If score >= 80 (HOT):
        1. Get best human agent (get_agent_conversion_stats)
        2. Assign immediately (assign_to_agent)
        3. Brief handoff message

        ### If score 60-79 (WARM):
        1. Schedule follow-up (1-2 days)
        2. Provide value content
        3. Re-qualify after 2nd touch

        ### If score < 60 (COLD/UNQUALIFIED):
        1. Politely explain fit issues
        2. Offer alternative resources
        3. Keep door open for future

        ## Step 4: Handoff (if qualified)

        Prepare handoff notes:
        ```
        LEAD SUMMARY
        - Name: [name]
        - Company: [company]
        - Score: [score]/100

        BANT ASSESSMENT:
        - Budget: [budget notes]
        - Authority: [authority notes]
        - Need: [need description]
        - Timeline: [timeline notes]

        KEY INSIGHTS:
        - [insight 1]
        - [insight 2]

        NEXT STEPS:
        - [recommended actions]
        ```

        # EXAMPLES

        ## Example 1: Hot Lead

        Lead: "Preciso de CRM urgente, nossa equipe tem 50 vendedores e perdemos muitos leads"

        SDR:
        <thinking>
        - Budget: Likely good (50 vendedores = empresa média/grande)
        - Authority: Unknown, need to ask
        - Need: STRONG (perdendo leads = pain point claro)
        - Timeline: URGENT (palavra "urgente")

        Initial score estimate: 75+ (warm-to-hot)
        </thinking>

        "Entendo a urgência! Perder leads é caro. Algumas perguntas rápidas:

        1. Você é a pessoa que decide a compra, ou precisa de aprovação?
        2. Qual orçamento vocês têm em mente para CRM?
        3. Quanto vocês estimam estar perdendo por mês sem CRM adequado?"

        [After answers, calculates score = 88]
        [Calls get_agent_conversion_stats to find top closer]
        [Calls assign_to_agent(agent_id=top_closer, priority=high)]

        "Perfeito! Vou conectar você com [Nome], nosso especialista em implementações para equipes médias. Ele tem 92% de taxa de sucesso e vai preparar uma proposta personalizada. Pode conversar agora?"

        ## Example 2: Cold Lead

        Lead: "Quanto custa?"

        SDR:
        <thinking>
        - Very little context
        - Need to qualify before pricing
        - Risk: price shoppers
        </thinking>

        "Ótima pergunta! Nossos planos variam conforme o tamanho da operação. Para eu te dar o preço mais adequado:

        1. Quantas pessoas vão usar o CRM?
        2. Quais funcionalidades são prioridade? (vendas, suporte, marketing)
        3. Já usam algum CRM hoje?"

        [If evasive responses → Score < 40]
        [Calls qualify_lead(status="unqualified", reason="price_shopper")]

        "Entendo. Deixo nosso site com a tabela de preços: [link]. Se quiser uma análise personalizada, é só me chamar!"

        # CONSTRAINTS
        - NEVER be pushy or aggressive
        - NEVER lie about features or pricing
        - ALWAYS respect if lead says "not now"
        - NEVER assign to human if score < 60 (wastes closer's time)
        - ALWAYS provide value in every interaction

        # TONE
        - Friendly but professional
        - Consultative (not salesy)
        - Genuinely helpful
        - Respectful of lead's time
        """

    def _calculate_lead_score(
        self,
        budget_score: int,        # 0-10
        authority_score: int,     # 0-10
        need_score: int,          # 0-10
        timeline_score: int,      # 0-10
        engagement_score: int,    # 0-10
        company_score: int,       # 0-10
        source_score: int,        # 0-10
        response_speed_score: int,# 0-10
    ) -> Dict:
        """
        Calculate lead score (0-100) from individual dimensions

        Weights:
        - BANT: 40 points (10 each)
        - Engagement: 20 points
        - Company: 20 points
        - Source: 10 points
        - Response speed: 10 points
        """

        # Calculate weighted score
        bant_score = (budget_score + authority_score + need_score + timeline_score) * 1.0  # 40 points max
        engagement = engagement_score * 2.0  # 20 points max
        company = company_score * 2.0  # 20 points max
        source = source_score * 1.0  # 10 points max
        speed = response_speed_score * 1.0  # 10 points max

        total_score = bant_score + engagement + company + source + speed

        # Classification
        if total_score >= 80:
            classification = "hot"
            action = "assign_immediately"
        elif total_score >= 60:
            classification = "warm"
            action = "nurture_then_assign"
        elif total_score >= 40:
            classification = "cold"
            action = "long_nurture"
        else:
            classification = "unqualified"
            action = "disqualify"

        return {
            "total_score": round(total_score, 1),
            "classification": classification,
            "recommended_action": action,
            "breakdown": {
                "bant": round(bant_score, 1),
                "engagement": round(engagement, 1),
                "company": round(company, 1),
                "source": round(source, 1),
                "response_speed": round(speed, 1),
            }
        }
```

---

## 🔍 AGENT ANALYZER AGENT (Quality Analysis)

### **Purpose**

Agent Analyzer faz análise profunda da qualidade de agentes (humanos ou AI), comparando:
- Grammar e ortografia
- Tom de voz e empatia
- Alinhamento com marca
- Tempo de resposta
- Satisfação do cliente

**Use Cases:**
- BI Manager pergunta "PORQUE João é melhor?"
- Manager quer comparar performance de 2 agentes
- Quality assurance contínua

### **Implementation**

```python
# ventros-ai/agents/agent_analyzer_agent.py

from adk import LlmAgent, Tool, Session
from vertexai.generative_models import GenerativeModel
from typing import Dict, List
from ..mcp_integration import get_mcp_toolset

class AgentAnalyzerAgent(LlmAgent):
    """
    Agent Analyzer - Deep quality analysis of agent performance

    Capabilities:
    - Analyze individual agent quality (grammar, tone, brand alignment)
    - Compare multiple agents
    - Identify strengths and weaknesses
    - Provide actionable recommendations

    Tools:
    - MCP: analyze_agent_messages (LLM-based quality analysis)
    - MCP: compare_agents (side-by-side comparison)
    - Direct: aggregate_scores, generate_recommendations
    """

    def __init__(
        self,
        memory_service,
        callback_manager,
    ):
        # Get MCP tools
        mcp_toolset = get_mcp_toolset()

        mcp_tools = [
            mcp_toolset.get_tool("analyze_agent_messages"),
            mcp_toolset.get_tool("compare_agents"),
        ]

        direct_tools = [
            Tool(name="aggregate_scores", function=self._aggregate_scores),
            Tool(name="generate_recommendations", function=self._generate_recommendations),
        ]

        super().__init__(
            name="agent_analyzer",
            instruction=self._build_instruction(),
            model=GenerativeModel(
                "gemini-2.0-flash",
                generation_config={
                    "temperature": 0.2,  # Low temp for objective analysis
                    "top_p": 0.95,
                    "max_output_tokens": 4096,
                },
            ),
            tools=mcp_tools + direct_tools,
            memory_service=memory_service,
            callbacks=callback_manager.get_callbacks(),
        )

    def _build_instruction(self) -> str:
        return """
        # ROLE
        You are an Agent Quality Analyzer.
        You analyze agent performance across multiple dimensions and provide objective insights.

        # ANALYSIS DIMENSIONS

        ## 1. Grammar & Spelling (0-10)
        - Correct spelling
        - Proper punctuation
        - Sentence structure
        - Professional language

        ## 2. Tone & Empathy (0-10)
        - Empathetic responses
        - Appropriate formality
        - Warmth and friendliness
        - Active listening indicators

        ## 3. Brand Alignment (0-10)
        - Matches brand voice guidelines
        - Uses approved terminology
        - Follows messaging framework
        - Consistent brand personality

        ## 4. Response Time
        - Average time to first response
        - Time per message
        - Consistency across sessions

        ## 5. Customer Satisfaction
        - Resolution rate
        - Escalation rate
        - Positive sentiment in replies
        - CSAT scores (if available)

        # PROTOCOL

        ## For Single Agent Analysis:

        1. Call analyze_agent_messages(agent_id, start_date, end_date, sample_size)
           - MCP tool returns LLM-analyzed scores per dimension

        2. Call aggregate_scores to calculate averages

        3. Generate insights:
           - Top 3 strengths
           - Top 3 areas for improvement
           - Specific examples (best and worst messages)

        4. Call generate_recommendations
           - Actionable improvement steps
           - Training suggestions
           - Best practice examples

        ## For Multi-Agent Comparison:

        1. Call compare_agents(agent_ids, start_date, end_date)
           - MCP tool returns comparative analysis

        2. Identify winner per dimension

        3. Analyze WHY winner is better:
           - Specific behaviors
           - Message patterns
           - Response strategies

        4. Generate recommendations for lower performers:
           - What to learn from winner
           - Specific behaviors to adopt

        # OUTPUT FORMAT

        Always return structured analysis:

        ```json
        {
            "agent_id": "agent-uuid",
            "agent_name": "João Silva",
            "analysis_period": "2025-01-01 to 2025-01-15",
            "sample_size": 50,

            "scores": {
                "grammar": 9.2,
                "tone": 8.9,
                "brand_alignment": 9.5,
                "avg_response_time_seconds": 138,
                "customer_satisfaction": 8.7
            },

            "strengths": [
                "Excelente alinhamento com marca (9.5/10)",
                "Respostas rápidas (2.3 min média)",
                "Grammar impecável (9.2/10)"
            ],

            "improvements": [
                "Tom pode ser mais empático em situações de frustração",
                "Usar mais perguntas abertas para entender necessidades",
                "Reduzir uso de jargão técnico"
            ],

            "best_message_example": {
                "text": "[example of excellent message]",
                "why_good": "Empático, claro, resolveu problema rapidamente"
            },

            "worst_message_example": {
                "text": "[example of poor message]",
                "why_bad": "Muito técnico, não demonstrou empatia"
            },

            "recommendations": [
                "Treinamento: Comunicação empática em situações de conflito",
                "Observar: Mensagens do João (top performer) para aprender estratégias",
                "Praticar: Usar framework 'Acknowledge → Empathize → Solve'"
            ]
        }
        ```

        # COMPARISON OUTPUT FORMAT

        When comparing agents:

        ```json
        {
            "comparison_id": "comp-uuid",
            "agents": ["joão", "maria"],
            "period": "2025-01-01 to 2025-01-15",

            "winners_per_dimension": {
                "grammar": {"winner": "joão", "score": 9.2, "vs": 7.8},
                "tone": {"winner": "joão", "score": 8.9, "vs": 8.1},
                "response_time": {"winner": "joão", "avg_seconds": 138, "vs": 342},
                "brand_alignment": {"winner": "joão", "score": 9.5, "vs": 8.0}
            },

            "overall_winner": "joão",

            "why_winner_is_better": [
                "João responde 2.5x mais rápido (2.3 min vs 5.7 min)",
                "João tem melhor grammar (9.2 vs 7.8)",
                "João está mais alinhado com marca (9.5 vs 8.0)"
            ],

            "what_maria_should_learn": [
                "Observar como João estrutura respostas (intro empática + solução + confirmação)",
                "Adotar template de respostas rápidas para perguntas comuns",
                "Revisar guia de estilo da marca"
            ],

            "specific_examples": {
                "joao_best": "[example]",
                "maria_needs_improvement": "[example]"
            }
        }
        ```

        # CONSTRAINTS
        - ALWAYS be objective (data-driven)
        - NEVER be judgmental or harsh
        - ALWAYS provide specific examples
        - ALWAYS include actionable recommendations
        - NEVER compare agents publicly (sensitive info)

        # TONE
        - Objective and analytical
        - Constructive (focus on growth)
        - Specific (cite examples)
        - Encouraging (highlight strengths too)
        """

    def _aggregate_scores(
        self,
        scores_list: List[Dict],
    ) -> Dict:
        """
        Aggregate scores from multiple messages

        Input: List of per-message analysis from MCP tool
        Output: Aggregated statistics
        """
        if not scores_list:
            return {}

        # Calculate averages
        grammar_scores = [s["grammar"] for s in scores_list]
        tone_scores = [s["tone"] for s in scores_list]
        brand_scores = [s["brand_alignment"] for s in scores_list]

        return {
            "grammar": {
                "mean": sum(grammar_scores) / len(grammar_scores),
                "min": min(grammar_scores),
                "max": max(grammar_scores),
            },
            "tone": {
                "mean": sum(tone_scores) / len(tone_scores),
                "min": min(tone_scores),
                "max": max(tone_scores),
            },
            "brand_alignment": {
                "mean": sum(brand_scores) / len(brand_scores),
                "min": min(brand_scores),
                "max": max(brand_scores),
            },
            "sample_size": len(scores_list),
        }

    def _generate_recommendations(
        self,
        strengths: List[str],
        weaknesses: List[str],
        comparison_data: Dict = None,
    ) -> List[str]:
        """
        Generate actionable recommendations

        Based on:
        - Agent's weaknesses
        - Best practices from top performers (if comparison_data provided)
        - Standard training resources
        """
        recommendations = []

        # Address each weakness
        weakness_actions = {
            "grammar": "Revisar guia de redação empresarial + usar corretor ortográfico",
            "tone": "Treinamento: Comunicação empática + framework de escuta ativa",
            "brand_alignment": "Estudar brand guidelines + shadowing com top performer",
            "response_time": "Adotar templates para perguntas comuns + priorização de urgentes",
        }

        for weakness in weaknesses:
            for key, action in weakness_actions.items():
                if key in weakness.lower():
                    recommendations.append(action)

        # If comparison data: learn from winner
        if comparison_data and "winner_strategies" in comparison_data:
            recommendations.append(
                f"Observar estratégias de {comparison_data['winner_name']}: {', '.join(comparison_data['winner_strategies'])}"
            )

        return recommendations

# Integration example: BI Manager calls Agent Analyzer
async def handle_why_question(question: str, tenant_id: str):
    """
    Handle "PORQUE" questions from BI Manager

    Example: "Porque o João converteu mais que a Maria?"

    Flow:
    1. BI Manager detects qualitative question
    2. Delegates to Agent Analyzer
    3. Agent Analyzer calls MCP compare_agents
    4. Returns structured analysis
    5. BI Manager formats for user
    """

    # Parse agent names from question (NER or explicit)
    agent_names = extract_agent_names(question)  # ["joão", "maria"]

    # Create session
    session = Session(state={
        "tenant_id": tenant_id,
        "knowledge_scope": {
            "lookback_days": 90,
            "include_messages": True,
            "agent_ids": agent_names,
        }
    })

    # Get Agent Analyzer
    agent_factory = get_agent_factory()
    analyzer = agent_factory.create_agent("agent_analyzer")

    # Execute analysis
    response = await analyzer.run_async(
        user_input=f"Compare agents {agent_names[0]} and {agent_names[1]} and explain why one is better",
        session=session,
    )

    # Parse structured result
    import json
    analysis = json.loads(response.output)

    return analysis
```

---

## 🔄 DYNAMIC KNOWLEDGESCOPE PATTERN

### **How Coordinator Changes Scope**

```python
# ventros-ai/dynamic_knowledge_scope.py

class DynamicKnowledgeScopeManager:
    """
    Manages dynamic KnowledgeScope changes as coordinator delegates between agents

    Pattern:
    1. User message arrives with default scope
    2. Coordinator analyzes intent
    3. Coordinator updates scope based on specialist needs
    4. Specialist executes with new scope
    5. Result returns to coordinator
    """

    @staticmethod
    def adjust_scope_for_agent(
        agent_category: str,
        user_intent: str,
        current_session: Session,
    ) -> Dict:
        """
        Adjusts KnowledgeScope parameters based on agent category and user intent

        Returns updated scope dict
        """

        # Default scopes per agent category
        default_scopes = {
            "retention_churn": {
                "lookback_days": 90,
                "include_sessions": True,
                "include_contact_events": True,
                "include_agent_transfers": True,
            },
            "sales_prospecting": {
                "lookback_days": 30,
                "include_tracking": True,
                "include_pipeline": True,
            },
            "support_technical": {
                "lookback_days": 7,
                "include_sessions": True,
            },
            "bi_manager": {
                "lookback_days": 365,  # Can query all-time
                "include_all": True,   # Needs everything
            },
            "agent_analyzer": {
                "lookback_days": 90,
                "include_messages": True,  # Needs message content for analysis
                "include_agent_performance": True,
            },
        }

        # Get base scope
        base_scope = default_scopes.get(agent_category, {})

        # Intent-based adjustments
        if "today" in user_intent.lower() or "hoje" in user_intent.lower():
            base_scope["lookback_days"] = 1

        if "month" in user_intent.lower() or "mês" in user_intent.lower():
            base_scope["lookback_days"] = 30

        if "why" in user_intent.lower() or "porque" in user_intent.lower():
            # "Why" questions need broader context
            base_scope["lookback_days"] = max(base_scope.get("lookback_days", 30), 90)
            base_scope["include_messages"] = True

        if "compare" in user_intent.lower() or "comparar" in user_intent.lower():
            base_scope["include_agent_performance"] = True

        return base_scope

# Usage in Coordinator
class CoordinatorAgent(LlmAgent):
    """
    Enhanced coordinator with dynamic scope management
    """

    async def delegate_to_specialist(
        self,
        user_input: str,
        session: Session,
        specialist_category: str,
    ):
        """
        Delegate to specialist with adjusted knowledge scope
        """

        # 1. Adjust scope for specialist
        scope_manager = DynamicKnowledgeScopeManager()
        new_scope = scope_manager.adjust_scope_for_agent(
            agent_category=specialist_category,
            user_intent=user_input,
            current_session=session,
        )

        # 2. Update session state
        old_scope = session.state.get("knowledge_scope", {})
        session.state["knowledge_scope"] = new_scope
        session.state["agent_category"] = specialist_category

        # 3. Get specialist
        specialist = self.agent_factory.create_agent(specialist_category)

        # 4. Execute with new scope
        # (memory_service will use updated scope from session.state)
        response = await specialist.run_async(
            user_input=user_input,
            session=session,
        )

        # 5. Restore original scope (if needed for next delegation)
        session.state["knowledge_scope"] = old_scope

        return response

# Example flow
"""
User: "Quantos leads tive hoje?"

1. Coordinator analyzes → BI Manager needed
2. Coordinator adjusts scope:
   {
     "lookback_days": 1,  # "hoje" → 1 day
     "include_tracking": True,
     "include_pipeline": True,
   }
3. BI Manager executes with narrow scope (fast query)
4. Returns: "47 leads hoje"

---

User: "PORQUE o João é melhor que a Maria?"

1. Coordinator analyzes → Agent Analyzer needed
2. Coordinator adjusts scope:
   {
     "lookback_days": 90,  # "porque" → broader context
     "include_messages": True,  # Need message content
     "include_agent_performance": True,
     "agent_ids": ["joão", "maria"],
   }
3. Agent Analyzer executes with broad scope
4. Returns: Detailed comparison analysis
"""
```

---

## 📝 RESPONSE FORMATTING GUIDE GENERATION

### **Pattern: Agent Returns Format Guide, Go Formats**

```python
# ventros-ai/response_formatting.py

from dataclasses import dataclass
from typing import Dict, List, Any
from enum import Enum

class FormatType(Enum):
    """Supported output formats"""
    MARKDOWN = "markdown"
    HTML = "html"
    JSON = "json"
    PLAIN_TEXT = "plain_text"

@dataclass
class ResponseFormatGuide:
    """
    Formatting guide that agent returns with response

    Go backend will use this to format the final output
    """
    format: FormatType
    structure: Dict[str, Any]
    styling: Dict[str, Any] = None
    chart_config: Dict[str, Any] = None

    def to_dict(self) -> Dict:
        return {
            "format": self.format.value,
            "structure": self.structure,
            "styling": self.styling or {},
            "chart_config": self.chart_config,
        }

# Example: Agent generates response with format guide
class ResponseWithFormatting:
    """
    Wrapper for agent responses that include formatting guides
    """

    @staticmethod
    def create_simple_text_response(
        text: str,
        highlight_numbers: bool = False,
    ) -> Dict:
        """Simple text response with optional number highlighting"""
        return {
            "answer": text,
            "format_guide": ResponseFormatGuide(
                format=FormatType.MARKDOWN,
                structure={
                    "type": "single_paragraph",
                    "highlight_numbers": highlight_numbers,
                },
                styling={
                    "bold_numbers": highlight_numbers,
                }
            ).to_dict()
        }

    @staticmethod
    def create_table_response(
        summary: str,
        table_title: str,
        columns: List[str],
        rows: List[List[Any]],
        chart_type: str = None,
    ) -> Dict:
        """Response with summary + table + optional chart"""
        response = {
            "answer": summary,
            "format_guide": ResponseFormatGuide(
                format=FormatType.MARKDOWN,
                structure={
                    "sections": [
                        {
                            "type": "summary",
                            "content": summary,
                        },
                        {
                            "type": "table",
                            "title": table_title,
                            "columns": columns,
                            "rows": rows,
                        }
                    ]
                },
            ).to_dict()
        }

        # Add chart config if specified
        if chart_type:
            response["format_guide"]["chart_config"] = {
                "type": chart_type,
                "data_source": "table",
                "x_column": columns[0],
                "y_column": columns[1],
            }

        return response

    @staticmethod
    def create_comparison_response(
        winner: str,
        dimensions: Dict[str, Dict],
        insights: List[str],
        recommendations: List[str],
    ) -> Dict:
        """Response for agent comparisons (BI Manager + Agent Analyzer)"""
        return {
            "answer": f"Análise comparativa completa. Vencedor: {winner}",
            "format_guide": ResponseFormatGuide(
                format=FormatType.MARKDOWN,
                structure={
                    "sections": [
                        {
                            "title": "Resumo Executivo",
                            "type": "text",
                            "content": f"**{winner}** teve melhor desempenho.",
                            "styling": {"highlight_winner": True},
                        },
                        {
                            "title": "Comparação por Dimensão",
                            "type": "table",
                            "columns": ["Dimensão", "Vencedor", "Score"],
                            "rows": [
                                [dim, data["winner"], data["score"]]
                                for dim, data in dimensions.items()
                            ],
                        },
                        {
                            "title": "Insights Principais",
                            "type": "bullets",
                            "items": insights,
                        },
                        {
                            "title": "Recomendações",
                            "type": "bullets",
                            "items": recommendations,
                            "styling": {"icon": "💡"},
                        }
                    ]
                },
                chart_config={
                    "type": "radar",
                    "dimensions": list(dimensions.keys()),
                    "series": [
                        {
                            "name": agent,
                            "data": [
                                dimensions[dim].get(agent, 0)
                                for dim in dimensions.keys()
                            ]
                        }
                        for agent in [winner, "other"]
                    ]
                }
            ).to_dict()
        }

# Integration: Agent uses these helpers
class BIManagerAgentEnhanced(LlmAgent):
    """Enhanced BI Manager that returns formatted responses"""

    def format_quantitative_response(
        self,
        query_result: Dict,
        question: str,
    ) -> str:
        """
        Format quantitative results with guide

        Returns JSON string that Go will parse
        """
        import json

        # Example: "Quantos leads tive hoje?" → 47 leads
        if "total" in query_result or "count" in query_result:
            count = query_result.get("total", query_result.get("count", 0))
            response = ResponseWithFormatting.create_simple_text_response(
                text=f"Você teve **{count} leads** no período solicitado.",
                highlight_numbers=True,
            )

        # Example: "Qual agente converteu mais?" → Table
        elif "agents" in query_result and "stats" in query_result:
            columns = ["Agente", "Conversões", "Taxa (%)"]
            rows = [
                [agent["name"], agent["conversions"], agent["rate"]]
                for agent in query_result["agents"]
            ]

            response = ResponseWithFormatting.create_table_response(
                summary=f"**{rows[0][0]}** teve mais conversões ({rows[0][1]} conversões, {rows[0][2]}% taxa)",
                table_title="Desempenho por Agente",
                columns=columns,
                rows=rows,
                chart_type="bar",
            )

        else:
            # Generic fallback
            response = ResponseWithFormatting.create_simple_text_response(
                text=json.dumps(query_result, indent=2),
            )

        return json.dumps(response)

# Go Backend: Format Response Based on Guide
"""
// Go side: infrastructure/http/handlers/ai_response_formatter.go

package handlers

import (
    "encoding/json"
    "fmt"
    "strings"
)

type ResponseWithFormatGuide struct {
    Answer      string                 `json:"answer"`
    FormatGuide ResponseFormatGuide    `json:"format_guide"`
}

type ResponseFormatGuide struct {
    Format      string                 `json:"format"`
    Structure   map[string]interface{} `json:"structure"`
    Styling     map[string]interface{} `json:"styling"`
    ChartConfig map[string]interface{} `json:"chart_config"`
}

func FormatAIResponse(rawResponse string) (string, error) {
    var response ResponseWithFormatGuide
    if err := json.Unmarshal([]byte(rawResponse), &response); err != nil {
        // Fallback: return raw if not formatted
        return rawResponse, nil
    }

    // Format based on guide
    switch response.FormatGuide.Format {
    case "markdown":
        return formatMarkdown(response)
    case "html":
        return formatHTML(response)
    case "json":
        return response.Answer, nil
    default:
        return response.Answer, nil
    }
}

func formatMarkdown(response ResponseWithFormatGuide) (string, error) {
    var output strings.Builder

    // Check structure type
    structure := response.FormatGuide.Structure

    if sections, ok := structure["sections"].([]interface{}); ok {
        // Multi-section response
        for _, sec := range sections {
            section := sec.(map[string]interface{})

            // Title
            if title, ok := section["title"].(string); ok {
                output.WriteString(fmt.Sprintf("## %s\\n\\n", title))
            }

            // Content based on type
            switch section["type"] {
            case "text":
                output.WriteString(fmt.Sprintf("%s\\n\\n", section["content"]))

            case "table":
                output.WriteString(formatTable(section))

            case "bullets":
                items := section["items"].([]interface{})
                for _, item := range items {
                    output.WriteString(fmt.Sprintf("- %s\\n", item))
                }
                output.WriteString("\\n")
            }
        }
    } else {
        // Simple response
        output.WriteString(response.Answer)
    }

    return output.String(), nil
}

func formatTable(section map[string]interface{}) string {
    var output strings.Builder

    columns := section["columns"].([]interface{})
    rows := section["rows"].([]interface{})

    // Header
    output.WriteString("|")
    for _, col := range columns {
        output.WriteString(fmt.Sprintf(" %s |", col))
    }
    output.WriteString("\\n")

    // Separator
    output.WriteString("|")
    for range columns {
        output.WriteString(" --- |")
    }
    output.WriteString("\\n")

    // Rows
    for _, r := range rows {
        row := r.([]interface{})
        output.WriteString("|")
        for _, cell := range row {
            output.WriteString(fmt.Sprintf(" %v |", cell))
        }
        output.WriteString("\\n")
    }

    output.WriteString("\\n")
    return output.String()
}
"""
```

---

## 🔄 OUTBOX PATTERN, DLQ & ERROR HANDLING

### **Communication Architecture: RabbitMQ vs gRPC**

```
┌──────────────────────────────────────────────────────────────┐
│  PROTOCOLO          USO                   QUANDO USAR         │
├──────────────────────────────────────────────────────────────┤
│  gRPC              Go ↔ Python            Sync calls          │
│  (HTTP/2)          Direct communication   Memory queries      │
│                                           MCP tool calls       │
│                                           Template requests    │
│                                                                │
│  RabbitMQ          Async events           Message processing  │
│  (AMQP)            Pub/Sub pattern        Background jobs     │
│                    Decoupling services    Event-driven flows  │
│                                           Workflow triggers    │
└──────────────────────────────────────────────────────────────┘
```

**RabbitMQ NÃO é gRPC:**
- **RabbitMQ**: Message broker usando AMQP protocol (async, pub/sub)
- **gRPC**: RPC framework usando HTTP/2 (sync, request/response)

### **Outbox Pattern Implementation**

```python
# ventros-ai/outbox_pattern.py

"""
OUTBOX PATTERN:

Problem: What if we process a message but fail to publish response to RabbitMQ?
- Agent processes message successfully
- Tries to publish response to RabbitMQ
- RabbitMQ is down → Response is lost!

Solution: Transactional Outbox
1. Store response in database first (transactional)
2. Background worker publishes from outbox to RabbitMQ
3. Retry with exponential backoff
4. DLQ for permanent failures
"""

from dataclasses import dataclass
from datetime import datetime
from typing import Dict, Any
import asyncio

@dataclass
class OutboxMessage:
    """
    Message stored in outbox table before publishing to RabbitMQ
    """
    id: str
    event_type: str
    payload: Dict[str, Any]
    destination_queue: str
    status: str  # pending, published, failed, dlq
    retry_count: int = 0
    max_retries: int = 3
    created_at: datetime = None
    published_at: datetime = None
    error_message: str = None

class OutboxRepository:
    """
    Stores outbox messages in database (PostgreSQL)
    """

    async def save(self, message: OutboxMessage) -> None:
        """
        Save message to outbox table

        This is done in same transaction as agent processing
        """
        # INSERT INTO outbox_messages (...)
        pass

    async def get_pending(self, limit: int = 100) -> List[OutboxMessage]:
        """Get pending messages to publish"""
        # SELECT * FROM outbox_messages WHERE status = 'pending' LIMIT $1
        pass

    async def mark_published(self, message_id: str) -> None:
        """Mark message as successfully published"""
        # UPDATE outbox_messages SET status = 'published', published_at = NOW()
        pass

    async def increment_retry(
        self,
        message_id: str,
        error_message: str,
    ) -> None:
        """Increment retry count after failure"""
        # UPDATE outbox_messages SET retry_count = retry_count + 1, error_message = $1
        pass

    async def move_to_dlq(self, message_id: str, reason: str) -> None:
        """Move to DLQ after max retries"""
        # UPDATE outbox_messages SET status = 'dlq', error_message = $1
        pass

class OutboxPublisher:
    """
    Background worker that publishes from outbox to RabbitMQ

    Runs continuously in background
    """

    def __init__(
        self,
        outbox_repo: OutboxRepository,
        rabbitmq_publisher: RabbitMQPublisher,
        poll_interval_seconds: int = 5,
    ):
        self.outbox_repo = outbox_repo
        self.rabbitmq_publisher = rabbitmq_publisher
        self.poll_interval = poll_interval_seconds

    async def start(self):
        """
        Start outbox publisher worker

        Polls database for pending messages and publishes to RabbitMQ
        """
        print("✅ Outbox publisher started")

        while True:
            try:
                # Get pending messages
                pending = await self.outbox_repo.get_pending(limit=100)

                if not pending:
                    # No pending messages, wait
                    await asyncio.sleep(self.poll_interval)
                    continue

                # Process each message
                for message in pending:
                    await self._process_message(message)

            except Exception as e:
                print(f"❌ Outbox publisher error: {e}")
                await asyncio.sleep(self.poll_interval)

    async def _process_message(self, message: OutboxMessage):
        """
        Process single outbox message

        Flow:
        1. Try to publish to RabbitMQ
        2. If success: mark as published
        3. If failure:
           - If retry_count < max_retries: increment retry
           - Else: move to DLQ
        """
        try:
            # Try to publish to RabbitMQ
            await self.rabbitmq_publisher.publish(
                queue=message.destination_queue,
                event_type=message.event_type,
                payload=message.payload,
            )

            # Success: mark as published
            await self.outbox_repo.mark_published(message.id)
            print(f"✅ Published outbox message: {message.id}")

        except RabbitMQConnectionError as e:
            # RabbitMQ is down or connection failed
            if message.retry_count >= message.max_retries:
                # Max retries reached: move to DLQ
                await self.outbox_repo.move_to_dlq(
                    message.id,
                    reason=f"Max retries reached. Last error: {str(e)}",
                )
                print(f"⚠️  Message moved to DLQ: {message.id}")

                # Alert ops team (Slack, PagerDuty, etc)
                await self._alert_dlq(message, str(e))

            else:
                # Increment retry with exponential backoff
                await self.outbox_repo.increment_retry(
                    message.id,
                    error_message=str(e),
                )
                print(f"🔄 Retry {message.retry_count + 1}/{message.max_retries} for message: {message.id}")

                # Exponential backoff: wait before next retry
                backoff_seconds = 2 ** message.retry_count  # 2, 4, 8 seconds
                await asyncio.sleep(backoff_seconds)

        except Exception as e:
            # Unexpected error: move to DLQ immediately
            await self.outbox_repo.move_to_dlq(
                message.id,
                reason=f"Unexpected error: {str(e)}",
            )
            print(f"❌ Unexpected error, moved to DLQ: {message.id}")

    async def _alert_dlq(self, message: OutboxMessage, error: str):
        """
        Alert operations team when message moves to DLQ

        Integration: Slack, PagerDuty, email, etc
        """
        alert_message = f"""
        🚨 OUTBOX MESSAGE MOVED TO DLQ

        Message ID: {message.id}
        Event Type: {message.event_type}
        Destination: {message.destination_queue}
        Retry Count: {message.retry_count}
        Error: {error}

        Action Required: Investigate and manually reprocess if needed.
        """

        # Send alert (Slack example)
        # await slack_client.post_message(channel="#alerts", text=alert_message)
        print(alert_message)

class RabbitMQConnectionError(Exception):
    """Raised when RabbitMQ connection fails"""
    pass

# Integration with Agent Handler
async def handle_message_with_outbox(event_data: Dict):
    """
    Handle message with outbox pattern

    Flow:
    1. Process with agent
    2. Store response in outbox (transactional)
    3. Background worker publishes to RabbitMQ
    """

    # 1. Process with agent
    event = MessageReceivedEvent(**event_data["payload"])
    session = await load_or_create_session(event.contact_id, event.session_id)

    coordinator = get_coordinator_agent()
    response = await coordinator.run_async(
        user_input=event.text,
        session=session,
    )

    # 2. Store in outbox (instead of publishing directly)
    outbox_repo = get_outbox_repository()

    outbox_message = OutboxMessage(
        id=generate_uuid(),
        event_type="message.outbound",
        payload={
            "contact_id": event.contact_id,
            "session_id": event.session_id,
            "text": response.output,
            "agent_id": response.agent_id,
            "source": "bot",
        },
        destination_queue="message.outbound",
        status="pending",
        created_at=datetime.now(),
    )

    # Save to database (transactional with agent state)
    await outbox_repo.save(outbox_message)

    # 3. Background worker will publish to RabbitMQ
    # (OutboxPublisher running in background)

    print(f"✅ Response stored in outbox: {outbox_message.id}")
```

### **Dead Letter Queue (DLQ) Handling**

```python
# ventros-ai/dlq_handler.py

class DLQHandler:
    """
    Handles messages in Dead Letter Queue

    Responsibilities:
    - Monitor DLQ
    - Analyze failure patterns
    - Provide manual reprocessing interface
    - Alert operations team
    """

    def __init__(
        self,
        outbox_repo: OutboxRepository,
        alert_service: AlertService,
    ):
        self.outbox_repo = outbox_repo
        self.alert_service = alert_service

    async def get_dlq_messages(
        self,
        limit: int = 100,
        since: datetime = None,
    ) -> List[OutboxMessage]:
        """Get messages in DLQ"""
        # SELECT * FROM outbox_messages WHERE status = 'dlq'
        pass

    async def analyze_dlq_patterns(self) -> Dict:
        """
        Analyze DLQ for patterns

        Returns:
        - Most common error types
        - Affected event types
        - Time distribution
        """
        dlq_messages = await self.get_dlq_messages(limit=1000)

        # Group by error type
        error_counts = {}
        for msg in dlq_messages:
            error_type = self._extract_error_type(msg.error_message)
            error_counts[error_type] = error_counts.get(error_type, 0) + 1

        # Group by event type
        event_counts = {}
        for msg in dlq_messages:
            event_counts[msg.event_type] = event_counts.get(msg.event_type, 0) + 1

        return {
            "total_dlq_messages": len(dlq_messages),
            "error_types": error_counts,
            "event_types": event_counts,
            "oldest_message": min(dlq_messages, key=lambda m: m.created_at),
        }

    async def reprocess_message(
        self,
        message_id: str,
        force: bool = False,
    ) -> bool:
        """
        Manually reprocess a DLQ message

        Args:
            message_id: Message to reprocess
            force: Force reprocess even if error not resolved

        Returns:
            True if successful, False otherwise
        """
        # Get message from DLQ
        message = await self.outbox_repo.get_by_id(message_id)

        if message.status != "dlq":
            raise ValueError(f"Message {message_id} is not in DLQ")

        # Reset status and retry count
        message.status = "pending"
        message.retry_count = 0
        message.error_message = None

        await self.outbox_repo.update(message)

        # OutboxPublisher will pick it up
        print(f"🔄 Reprocessing DLQ message: {message_id}")

        return True

    async def bulk_reprocess(
        self,
        error_type: str = None,
        event_type: str = None,
    ) -> int:
        """
        Bulk reprocess DLQ messages matching criteria

        Returns: Number of messages reprocessed
        """
        dlq_messages = await self.get_dlq_messages()

        # Filter by criteria
        filtered = dlq_messages
        if error_type:
            filtered = [
                m for m in filtered
                if self._extract_error_type(m.error_message) == error_type
            ]
        if event_type:
            filtered = [m for m in filtered if m.event_type == event_type]

        # Reprocess each
        count = 0
        for message in filtered:
            try:
                await self.reprocess_message(message.id)
                count += 1
            except Exception as e:
                print(f"❌ Failed to reprocess {message.id}: {e}")

        return count

    def _extract_error_type(self, error_message: str) -> str:
        """Extract error type from error message"""
        if "connection" in error_message.lower():
            return "connection_error"
        elif "timeout" in error_message.lower():
            return "timeout_error"
        elif "auth" in error_message.lower():
            return "auth_error"
        else:
            return "unknown_error"

# CLI for DLQ Management
async def dlq_cli():
    """
    CLI for managing DLQ

    Commands:
    - dlq list: Show DLQ messages
    - dlq analyze: Analyze patterns
    - dlq reprocess <id>: Reprocess specific message
    - dlq reprocess-all --error-type connection_error: Bulk reprocess
    """
    import sys

    dlq_handler = DLQHandler(
        outbox_repo=get_outbox_repository(),
        alert_service=get_alert_service(),
    )

    command = sys.argv[1] if len(sys.argv) > 1 else "list"

    if command == "list":
        messages = await dlq_handler.get_dlq_messages(limit=10)
        for msg in messages:
            print(f"ID: {msg.id} | Type: {msg.event_type} | Error: {msg.error_message}")

    elif command == "analyze":
        analysis = await dlq_handler.analyze_dlq_patterns()
        print(json.dumps(analysis, indent=2))

    elif command == "reprocess":
        message_id = sys.argv[2]
        success = await dlq_handler.reprocess_message(message_id)
        print(f"Reprocess {'successful' if success else 'failed'}")

    elif command == "reprocess-all":
        error_type = sys.argv[3] if len(sys.argv) > 3 else None
        count = await dlq_handler.bulk_reprocess(error_type=error_type)
        print(f"Reprocessed {count} messages")

# Run CLI
# python -m ventros_ai.dlq_cli list
# python -m ventros_ai.dlq_cli analyze
# python -m ventros_ai.dlq_cli reprocess <message-id>
# python -m ventros_ai.dlq_cli reprocess-all connection_error
```

### **Error Handling Summary**

```
FLUXO COMPLETO (Com Outbox + DLQ):

1. Message chega no Python ADK (via RabbitMQ)
   ↓
2. Agent processa e gera resposta
   ↓
3. Resposta é salva no OUTBOX (database, transactional)
   ✅ GARANTIDO: Resposta persistida mesmo se RabbitMQ cair
   ↓
4. Background OutboxPublisher pega mensagens pending
   ↓
5. Tenta publicar no RabbitMQ:
   ├─ ✅ SUCESSO → marca como 'published'
   ├─ ❌ FALHA (RabbitMQ down):
   │   ├─ retry_count < 3 → incrementa retry (backoff exponencial)
   │   └─ retry_count >= 3 → move para DLQ + alerta ops team
   │
6. DLQ Handler:
   ├─ Monitora mensagens em DLQ
   ├─ Analisa padrões de erro
   ├─ Permite reprocessamento manual
   └─ Alerta equipe de ops

GARANTIAS:
✅ At-least-once delivery (mensagem nunca é perdida)
✅ Retry automático com backoff exponencial
✅ DLQ para investigação e reprocessamento manual
✅ Alertas para ops team
✅ Graceful degradation (RabbitMQ pode cair, sistema continua)
```

---

## ✅ RESUMO EXECUTIVO PYTHON ADK (ATUALIZADO 2025)

### **O que este serviço faz:**

1. **Consome eventos** do RabbitMQ (mensagens inbound, contact created, etc)
2. **Classifica intent** com Semantic Router (zero-shot)
3. **Orquestra agents** (Coordinator → Specialists via AgentTool)
4. **Busca contexto** no Go Memory Service via gRPC
5. **Processa com LLM** Gemini 2.0 Flash + Tools (MCP + Direct)
6. **Publica resposta** no Outbox (transactional) → RabbitMQ (async)
7. **Observa tudo** com Phoenix + OpenTelemetry + Prometheus

### **Agents disponíveis:**

#### Core Agents:
- **Coordinator**: Roteia para especialistas (AgentTool pattern)
- **Balanced**: Fallback genérico para casos não especializados

#### Sales Agents:
- **Sales Prospecting**: Qualificação de leads (BANT framework)
- **Sales Negotiation**: Negociação de preços e condições
- **Sales Closing**: Fechamento de vendas
- **SDR Agent**: Atendimento inicial + qualificação + assignment (NEW)

#### Retention Agents:
- **Retention Churn**: Prevenção de cancelamento (ofertas personalizadas)
- **Retention Upsell**: Cross-sell e upsell
- **Retention Winback**: Recuperação de clientes perdidos

#### Support Agents:
- **Support Technical**: Suporte técnico (bugs, errors)
- **Support Billing**: Questões de pagamento e fatura
- **Support Onboarding**: Onboarding de novos clientes

#### Operations Agents:
- **Operations Followup**: Follow-ups automáticos
- **Operations Schedule**: Agendamento de interações
- **Operations QA**: Quality assurance (LoopAgent pattern)

#### Analytics & BI Agents (NEW):
- **BI Manager**: Business intelligence queries
  - Quantitative: "Quantos leads tive hoje?" (MCP: get_leads_count)
  - Qualitative: "PORQUE João é melhor?" (delega para Agent Analyzer)
  - Dynamic KnowledgeScope (ajusta escopo conforme pergunta)
  - Returns ResponseFormatGuide (tabelas, gráficos, markdown)

- **Agent Analyzer**: Quality analysis de agentes (humanos ou AI)
  - Analisa: grammar, tone, brand alignment, response time
  - Compara múltiplos agentes (MCP: compare_agents)
  - Gera recomendações acionáveis
  - Usado pelo BI Manager para perguntas qualitativas

### **Multi-Agent Patterns usados:**

- ✅ **Coordinator-Worker** (primary pattern): Coordinator com specialists como AgentTools
- ✅ **Handoff**: Transferência dinâmica entre agentes (escalation)
- ✅ **Reflection**: Self-critique loops (QA agents)
- ✅ **Sequential**: Pipelines determinísticos (onboarding)
- ✅ **Parallel**: Execução concorrente (lead enrichment)
- ✅ **Loop**: Iteração até condição (QA refinement)
- ✅ **Hierarchical**: Tree structure (CEO → Directors → Specialists)

### **Tools Architecture:**

#### MCP Tools (8 tools via Go MCP Server):
1. `get_leads_count` - BI queries (cached 5 min)
2. `get_agent_conversion_stats` - Agent performance metrics
3. `get_top_performing_agent` - Best agent finder
4. `analyze_agent_messages` - LLM-based quality analysis
5. `compare_agents` - Side-by-side agent comparison
6. `qualify_lead` - BANT qualification
7. `update_pipeline_stage` - Pipeline mutations
8. `assign_to_agent` - Agent assignment

#### Direct ADK Tools:
- `calculate_lead_score` - Lead scoring (lightweight)
- `format_table` - Markdown table formatting
- `format_chart_config` - Chart.js config generation
- `aggregate_scores` - Statistical aggregation
- `generate_recommendations` - Recommendation engine

### **Production-ready features:**

- ✅ **Event-driven async** (RabbitMQ AMQP)
- ✅ **gRPC communication** (Go Memory Service + MCP Server)
- ✅ **Outbox Pattern** (transactional message persistence)
- ✅ **DLQ & Retry** (exponential backoff, manual reprocessing)
- ✅ **Phoenix observability** (LLM-native: prompts, embeddings, hallucinations)
- ✅ **OpenTelemetry** (infrastructure tracing)
- ✅ **Prometheus** (metrics: agent_requests_total, llm_tokens_used, etc)
- ✅ **Temporal workflows** (30-day lead nurturing, saga patterns)
- ✅ **Dynamic KnowledgeScope** (changes as coordinator delegates)
- ✅ **Response Formatting** (agent returns guide, Go formats)
- ✅ **MCP Integration** (hybrid approach: MCP + Direct tools)
- ✅ **Agent Templates** (Go queries Python for templates via gRPC)
- ✅ **Callback instrumentation** (before/after hooks)
- ✅ **Error handling** (graceful degradation, circuit breakers)
- ✅ **WebSocket support** (real-time agent interaction)
- ✅ **Agent factory pattern** (dependency injection)

### **Communication Protocols:**

```
┌─────────────────────────────────────────────────┐
│  Protocol    │  Purpose           │  Direction  │
├─────────────────────────────────────────────────┤
│  gRPC        │  Memory queries    │  Python→Go  │
│  (HTTP/2)    │  MCP tool calls    │  Python→Go  │
│              │  Agent templates   │  Go→Python  │
│                                                  │
│  RabbitMQ    │  Message events    │  Bi-dir     │
│  (AMQP)      │  Workflow triggers │  Bi-dir     │
│              │  Background jobs   │  Bi-dir     │
│                                                  │
│  Outbox      │  Transactional pub │  DB→RabbitMQ│
│  (Database)  │  At-least-once     │             │
└─────────────────────────────────────────────────┘
```

### **Database Tables (Python ADK owns):**

```sql
-- Outbox pattern
CREATE TABLE outbox_messages (
    id UUID PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    destination_queue VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL, -- pending, published, failed, dlq
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    created_at TIMESTAMP NOT NULL,
    published_at TIMESTAMP,
    error_message TEXT,
    INDEX idx_status (status) WHERE status = 'pending'
);

-- Session state (if Python owns session storage)
CREATE TABLE agent_sessions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    contact_id UUID NOT NULL,
    agent_category VARCHAR(50),
    state JSONB, -- Session.state
    history JSONB, -- Session.history
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### **Observability Dashboards:**

#### Phoenix Dashboard (http://localhost:6006):
- 🔍 **Agent Flow Visualization**: Waterfall de coordinator → specialist → tools
- 📊 **LLM Analytics**: Token usage, costs, response quality, hallucination rates
- 🗺️  **Embedding Space**: UMAP projection, cluster analysis
- 💬 **Conversation Inspector**: Full history, context window utilization

#### Prometheus Metrics (http://localhost:9090):
```
agent_requests_total{agent_name="coordinator", agent_category="balanced"}
agent_duration_seconds{agent_name="retention_churn"}
agent_errors_total{agent_name="bi_manager", error_type="mcp_timeout"}
llm_tokens_used_total{model_name="gemini-2.0-flash", agent_name="sdr_agent"}
tool_calls_total{tool_name="get_leads_count", agent_name="bi_manager"}
```

### **Error Handling Flow:**

```
Message Processing:
1. RabbitMQ → Python ADK consumer
2. Agent processes (with try/catch)
3. Response → Outbox (transactional ✅)
4. Background worker → RabbitMQ
   ├─ Success → mark published
   ├─ Failure (retry < 3) → exponential backoff
   └─ Failure (retry >= 3) → DLQ + alert ops

DLQ Management:
- CLI: `python -m ventros_ai.dlq_cli list`
- CLI: `python -m ventros_ai.dlq_cli analyze`
- CLI: `python -m ventros_ai.dlq_cli reprocess <id>`
- CLI: `python -m ventros_ai.dlq_cli reprocess-all connection_error`
```

### **Deployment Checklist:**

```bash
# 1. Dependencies
pip install -r requirements.txt

# 2. Environment variables
export GOOGLE_CLOUD_PROJECT=your-project
export GOOGLE_APPLICATION_CREDENTIALS=path/to/creds.json
export MEMORY_SERVICE_GRPC_HOST=localhost:50051
export MCP_SERVER_HOST=localhost:50052
export RABBITMQ_URL=amqp://guest:guest@localhost:5672
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317

# 3. Start Phoenix
python -m phoenix.server

# 4. Start Temporal worker (if using workflows)
python -m ventros_ai.temporal_worker

# 5. Start main application
python -m ventros_ai.main

# 6. Start outbox publisher (background)
python -m ventros_ai.outbox_publisher

# 7. Health check
curl http://localhost:8000/health
```

### **Agent Template Example (Go queries Python):**

```python
# Python exposes templates via gRPC
templates = [
    {
        "template_id": "bi_manager",
        "name": "BI Manager",
        "category": "analytics",
        "retrieval_strategy": "bi_manager", # 365 days, include_all
        "tools": ["get_leads_count", "get_agent_conversion_stats", "compare_agents"],
        "intent_examples": ["quantos leads", "qual agente", "porque"],
    },
    {
        "template_id": "sdr_agent",
        "name": "SDR Agent",
        "category": "sales",
        "retrieval_strategy": "sales_prospecting",
        "tools": ["qualify_lead", "update_pipeline_stage", "assign_to_agent"],
        "intent_examples": ["preço", "orçamento", "quero comprar"],
    },
    # ... 15 total templates
]

# Go calls: grpcClient.ListAgentTemplates()
# Go creates agent entity with metadata from template
```

### **Performance Targets:**

- **Agent Response Time**: < 2s (p95), < 5s (p99)
- **Memory Query Time**: < 200ms (p95)
- **MCP Tool Call**: < 500ms (p95)
- **Outbox Publish Latency**: < 100ms (p95)
- **Throughput**: 100 messages/sec (single instance), 1000+ (scaled)
- **Token Efficiency**: < 2000 tokens/interaction (with prompt caching)
- **Cost**: ~$0.001 per interaction (Gemini 2.0 Flash pricing)

---

**Próximos passos:**

1. **Implementar gradualmente**: Comece com Coordinator + 1 specialist (RetentionChurnAgent)
2. **Configurar observabilidade**: Phoenix + Prometheus + alertas
3. **Testar padrões de erro**: DLQ, retry, circuit breakers
4. **Temporal workflows**: Implementar 1 workflow simples (lead nurturing)
5. **MCP integration**: Conectar com Go MCP Server
6. **Agent templates**: Expor via gRPC para Go
7. **Tuning**: Ajustar KnowledgeScopes, retrieval strategies, prompts
8. **Monitoring**: Dashboard com métricas chave
9. **Documentação**: Runbooks para ops team
10. **Load testing**: Validar targets de performance

**Arquitetura de referência completa em:**
- `AI_MEMORY_GO_ARCHITECTURE.md` - Camada de memória (Go)
- `PYTHON_ADK_ARCHITECTURE.md` - Camada de agentes (Python) ← YOU ARE HERE
- `AI_ARCHITECTURE_EXECUTIVE_SUMMARY.md` - Visão geral integrada
# Python ADK Architecture - Part 2
**Multi-Agent Orchestration, Tools, Session Management, and Complete Examples**

---

## Table of Contents (Part 2)
1. [Multi-Agent Orchestration Patterns](#multi-agent-orchestration-patterns)
2. [Tool Calling Patterns](#tool-calling-patterns)
3. [Sophisticated Session Manager](#sophisticated-session-manager)
4. [ReAct Pattern Deep Dive](#react-pattern-deep-dive)
5. [Self-Reflection & Planning](#self-reflection--planning)
6. [Complete Implementation Examples](#complete-implementation-examples)
7. [Production Deployment](#production-deployment)

---

## Multi-Agent Orchestration Patterns

### Pattern 1: Coordinator/Dispatcher (Router Pattern)

**Purpose**: Route incoming messages to specialized agents based on intent classification.

**Use Case**: Customer messages come in → Router determines intent → Dispatches to specialized agent (Sales/Support/Retention/etc).

```python
# ventros_adk/orchestration/coordinator_agent.py

from google.adk import LlmAgent, BaseAgent
from google.adk.models import GeminiModel
from typing import Dict, Any, Optional
import logging

class CoordinatorAgent(BaseAgent):
    """
    Master coordinator that routes messages to specialized agents.

    Flow:
    1. Receive message from customer
    2. Use SemanticRouter to classify intent
    3. Select best specialized agent (with fallback)
    4. Dispatch to agent
    5. Collect response
    6. Return to customer
    """

    def __init__(
        self,
        memory_service: 'VentrosMemoryService',
        session_manager: 'SessionManager',
        agent_registry: Dict[str, BaseAgent],
    ):
        super().__init__(name="coordinator_agent")

        self.memory_service = memory_service
        self.session_manager = session_manager
        self.agent_registry = agent_registry  # Map of agent_id → agent instance
        self.logger = logging.getLogger(__name__)

        # Fallback agent for when routing fails
        self.fallback_agent = LlmAgent(
            name="general_assistant",
            model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.7),
            system_prompt="""
            Você é um assistente geral do Ventros CRM.
            Responda perguntas de forma útil e educada.
            Se precisar de especialista, sugira transferir para agente específico.
            """,
            memory=memory_service,
        )

    def run(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Coordinate message routing and agent execution.
        """
        contact_id = input_data.get("contact_id")
        message = input_data.get("message")
        session_id = input_data.get("session_id")

        self.logger.info(f"Coordinator received message from contact {contact_id}")

        # Step 1: Semantic routing to determine best agent
        routing_result = self.memory_service.route_to_agent(
            message=message,
            contact_id=contact_id,
            available_agents=list(self.agent_registry.keys()),
        )

        agent_id = routing_result.get("agent_id")
        confidence = routing_result.get("confidence", 0.0)
        category = routing_result.get("category")
        reasoning = routing_result.get("reasoning")

        self.logger.info(
            f"Routing: agent={agent_id}, category={category}, "
            f"confidence={confidence:.2f}, reasoning={reasoning}"
        )

        # Step 2: Select agent (with confidence threshold)
        if agent_id and confidence >= 0.7:
            selected_agent = self.agent_registry.get(agent_id)
            if selected_agent:
                self.logger.info(f"Dispatching to specialized agent: {agent_id}")
            else:
                self.logger.warning(f"Agent {agent_id} not found in registry, using fallback")
                selected_agent = self.fallback_agent
        else:
            self.logger.info(f"Low confidence ({confidence:.2f}), using fallback agent")
            selected_agent = self.fallback_agent

        # Step 3: Prepare context for agent
        agent_input = {
            "contact_id": contact_id,
            "message": message,
            "session_id": session_id,
            "routing": {
                "category": category,
                "confidence": confidence,
                "reasoning": reasoning,
            },
        }

        # Step 4: Execute selected agent
        try:
            result = selected_agent.run(agent_input)

            # Log agent execution
            self.session_manager.log_agent_execution(
                session_id=session_id,
                agent_id=agent_id or "fallback",
                category=category,
                confidence=confidence,
                input_data=agent_input,
                output_data=result,
            )

            return {
                "success": True,
                "agent_used": agent_id or "fallback",
                "category": category,
                "confidence": confidence,
                "response": result.get("response"),
                "metadata": result.get("metadata", {}),
            }

        except Exception as e:
            self.logger.error(f"Agent execution failed: {str(e)}", exc_info=True)

            # Fallback to general assistant if specialized agent fails
            if selected_agent != self.fallback_agent:
                self.logger.info("Retrying with fallback agent")
                result = self.fallback_agent.run(agent_input)

                return {
                    "success": True,
                    "agent_used": "fallback",
                    "category": "general",
                    "confidence": 0.5,
                    "response": result.get("response"),
                    "error_recovered": True,
                }

            # Complete failure
            return {
                "success": False,
                "error": str(e),
                "response": "Desculpe, ocorreu um erro. Um agente humano será notificado.",
            }


# Example usage with full agent registry
def create_production_coordinator(
    memory_service: 'VentrosMemoryService',
    session_manager: 'SessionManager',
) -> CoordinatorAgent:
    """
    Create coordinator with full suite of specialized agents.
    """
    from ventros_adk.agents import (
        SalesProspectingAgent,
        ChurnPreventionAgent,
        TechnicalSupportAgent,
        BillingAgent,
        OnboardingAgent,
        UpsellAgent,
    )

    # Initialize all specialized agents
    agent_registry = {
        # Sales Category
        "sales_prospecting": SalesProspectingAgent(memory_service, session_manager),
        "sales_negotiation": SalesNegotiationAgent(memory_service, session_manager),
        "sales_closing": SalesClosingAgent(memory_service, session_manager),

        # Support Category
        "support_technical": TechnicalSupportAgent(memory_service, session_manager),
        "support_billing": BillingAgent(memory_service, session_manager),
        "support_onboarding": OnboardingAgent(memory_service, session_manager),

        # Retention Category
        "retention_churn": ChurnPreventionAgent(memory_service, session_manager),
        "retention_upsell": UpsellAgent(memory_service, session_manager),
        "retention_winback": WinbackAgent(memory_service, session_manager),

        # Operations Category
        "operations_schedule": SchedulingAgent(memory_service, session_manager),
        "operations_followup": FollowupAgent(memory_service, session_manager),
        "operations_qa": QualityAssuranceAgent(memory_service, session_manager),

        # Marketing Category
        "marketing_campaign": CampaignAgent(memory_service, session_manager),
        "marketing_content": ContentAgent(memory_service, session_manager),
        "marketing_event": EventAgent(memory_service, session_manager),
    }

    coordinator = CoordinatorAgent(
        memory_service=memory_service,
        session_manager=session_manager,
        agent_registry=agent_registry,
    )

    return coordinator
```

---

### Pattern 2: Hierarchical Task Decomposition

**Purpose**: Break complex tasks into subtasks, delegating each to specialized sub-agents.

**Use Case**: Customer wants "complete sales analysis" → Coordinator decomposes into → Lead qualification + Competitive analysis + Pricing strategy + Timeline planning → Each handled by sub-agent → Results aggregated.

```python
# ventros_adk/orchestration/hierarchical_agent.py

from google.adk import BaseAgent, LlmAgent
from google.adk.models import GeminiModel
from typing import Dict, Any, List
import logging

class HierarchicalTaskAgent(BaseAgent):
    """
    Hierarchical task decomposition and delegation.

    Master Agent:
    - Receives complex task
    - Uses LLM to decompose into subtasks
    - Delegates each subtask to specialized sub-agent
    - Aggregates results
    - Synthesizes final answer
    """

    def __init__(
        self,
        memory_service: 'VentrosMemoryService',
        session_manager: 'SessionManager',
        sub_agents: Dict[str, BaseAgent],
    ):
        super().__init__(name="hierarchical_task_agent")

        self.memory_service = memory_service
        self.session_manager = session_manager
        self.sub_agents = sub_agents
        self.logger = logging.getLogger(__name__)

        # Task decomposer (uses LLM reasoning)
        self.decomposer = LlmAgent(
            name="task_decomposer",
            model=GeminiModel(
                model_name="gemini-2.0-flash-thinking-exp",  # Deep reasoning
                temperature=0.4,
            ),
            system_prompt="""
            Você é um especialista em decomposição de tarefas complexas.

            Dada uma tarefa complexa, você deve:
            1. Analisar a tarefa e identificar componentes
            2. Decompor em subtarefas atômicas e independentes
            3. Determinar ordem de execução (paralelo vs sequencial)
            4. Mapear cada subtask para agente especializado

            Agentes disponíveis:
            - lead_qualifier: Qualificação BANT de leads
            - competitor_analyst: Análise competitiva
            - pricing_strategist: Estratégia de pricing
            - timeline_planner: Planejamento de timeline
            - risk_assessor: Avaliação de riscos
            - roi_calculator: Cálculo de ROI

            Retorne JSON:
            {
                "subtasks": [
                    {
                        "id": "task_1",
                        "description": "...",
                        "agent": "lead_qualifier",
                        "dependencies": [],  // IDs de tasks que devem completar primeiro
                        "estimated_time_s": 10
                    },
                    ...
                ],
                "execution_strategy": "parallel" | "sequential" | "hybrid"
            }
            """,
        )

        # Result synthesizer
        self.synthesizer = LlmAgent(
            name="result_synthesizer",
            model=GeminiModel(
                model_name="gemini-2.0-flash-thinking-exp",
                temperature=0.5,
            ),
            system_prompt="""
            Você recebe resultados de múltiplas subtasks.

            Sua tarefa:
            1. Sintetizar todos os resultados numa resposta coerente
            2. Identificar conflitos ou inconsistências
            3. Priorizar informações mais relevantes
            4. Formatar resposta final para o usuário

            Mantenha:
            - Clareza e organização
            - Insights acionáveis
            - Referências às fontes (qual sub-agent forneceu)
            """,
        )

    def run(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Execute hierarchical task decomposition and delegation.
        """
        task_description = input_data.get("task")
        contact_id = input_data.get("contact_id")

        self.logger.info(f"Hierarchical task: {task_description}")

        # Step 1: Decompose task
        decomposition = self.decomposer.run({
            "task": task_description,
            "contact_id": contact_id,
            "available_agents": list(self.sub_agents.keys()),
        })

        subtasks = decomposition.get("subtasks", [])
        execution_strategy = decomposition.get("execution_strategy", "sequential")

        self.logger.info(f"Decomposed into {len(subtasks)} subtasks, strategy={execution_strategy}")

        # Step 2: Execute subtasks based on strategy
        if execution_strategy == "parallel":
            results = self._execute_parallel(subtasks, contact_id)
        elif execution_strategy == "sequential":
            results = self._execute_sequential(subtasks, contact_id)
        else:  # hybrid
            results = self._execute_hybrid(subtasks, contact_id)

        # Step 3: Synthesize final result
        synthesis = self.synthesizer.run({
            "original_task": task_description,
            "subtask_results": results,
            "contact_id": contact_id,
        })

        return {
            "task": task_description,
            "subtasks": subtasks,
            "execution_strategy": execution_strategy,
            "subtask_results": results,
            "final_answer": synthesis.get("response"),
            "metadata": {
                "total_subtasks": len(subtasks),
                "successful_subtasks": sum(1 for r in results if r.get("success")),
                "failed_subtasks": sum(1 for r in results if not r.get("success")),
            },
        }

    def _execute_parallel(self, subtasks: List[Dict], contact_id: str) -> List[Dict]:
        """
        Execute all subtasks in parallel (no dependencies).
        """
        import asyncio

        async def run_subtask(subtask: Dict) -> Dict:
            agent_name = subtask.get("agent")
            agent = self.sub_agents.get(agent_name)

            if not agent:
                return {
                    "subtask_id": subtask.get("id"),
                    "success": False,
                    "error": f"Agent {agent_name} not found",
                }

            try:
                result = agent.run({
                    "task": subtask.get("description"),
                    "contact_id": contact_id,
                })
                return {
                    "subtask_id": subtask.get("id"),
                    "agent_used": agent_name,
                    "success": True,
                    "result": result,
                }
            except Exception as e:
                return {
                    "subtask_id": subtask.get("id"),
                    "agent_used": agent_name,
                    "success": False,
                    "error": str(e),
                }

        # Run all subtasks concurrently
        async def run_all():
            tasks = [run_subtask(st) for st in subtasks]
            return await asyncio.gather(*tasks)

        results = asyncio.run(run_all())
        return results

    def _execute_sequential(self, subtasks: List[Dict], contact_id: str) -> List[Dict]:
        """
        Execute subtasks one by one in order.
        """
        results = []

        for subtask in subtasks:
            agent_name = subtask.get("agent")
            agent = self.sub_agents.get(agent_name)

            if not agent:
                results.append({
                    "subtask_id": subtask.get("id"),
                    "success": False,
                    "error": f"Agent {agent_name} not found",
                })
                continue

            try:
                result = agent.run({
                    "task": subtask.get("description"),
                    "contact_id": contact_id,
                    "previous_results": results,  # Sequential subtasks can access previous results
                })
                results.append({
                    "subtask_id": subtask.get("id"),
                    "agent_used": agent_name,
                    "success": True,
                    "result": result,
                })
            except Exception as e:
                results.append({
                    "subtask_id": subtask.get("id"),
                    "agent_used": agent_name,
                    "success": False,
                    "error": str(e),
                })

        return results

    def _execute_hybrid(self, subtasks: List[Dict], contact_id: str) -> List[Dict]:
        """
        Execute subtasks respecting dependencies (DAG execution).

        Build dependency graph and execute in topological order,
        parallelizing when possible.
        """
        # Build dependency graph
        graph = {st["id"]: st.get("dependencies", []) for st in subtasks}
        subtask_map = {st["id"]: st for st in subtasks}

        # Topological sort to find execution order
        execution_order = self._topological_sort(graph)

        results = {}

        # Execute in waves (parallel within each wave)
        for wave in execution_order:
            import asyncio

            async def run_subtask(subtask_id: str) -> Dict:
                subtask = subtask_map[subtask_id]
                agent_name = subtask.get("agent")
                agent = self.sub_agents.get(agent_name)

                if not agent:
                    return {
                        "subtask_id": subtask_id,
                        "success": False,
                        "error": f"Agent {agent_name} not found",
                    }

                # Gather results from dependencies
                dependency_results = [results[dep_id] for dep_id in subtask.get("dependencies", [])]

                try:
                    result = agent.run({
                        "task": subtask.get("description"),
                        "contact_id": contact_id,
                        "dependency_results": dependency_results,
                    })
                    return {
                        "subtask_id": subtask_id,
                        "agent_used": agent_name,
                        "success": True,
                        "result": result,
                    }
                except Exception as e:
                    return {
                        "subtask_id": subtask_id,
                        "agent_used": agent_name,
                        "success": False,
                        "error": str(e),
                    }

            # Execute wave in parallel
            async def run_wave():
                tasks = [run_subtask(st_id) for st_id in wave]
                return await asyncio.gather(*tasks)

            wave_results = asyncio.run(run_wave())

            # Store results
            for res in wave_results:
                results[res["subtask_id"]] = res

        return list(results.values())

    def _topological_sort(self, graph: Dict[str, List[str]]) -> List[List[str]]:
        """
        Topological sort with wave detection for parallel execution.

        Returns list of waves, where each wave can be executed in parallel.
        """
        from collections import defaultdict, deque

        # Calculate in-degrees
        in_degree = defaultdict(int)
        for node in graph:
            if node not in in_degree:
                in_degree[node] = 0
            for dep in graph[node]:
                in_degree[dep] += 1

        # Find nodes with no dependencies (wave 0)
        queue = deque([node for node in graph if in_degree[node] == 0])
        waves = []

        while queue:
            # Current wave (all nodes with no remaining dependencies)
            wave = []
            for _ in range(len(queue)):
                node = queue.popleft()
                wave.append(node)

                # Reduce in-degree for dependent nodes
                for dep in graph[node]:
                    in_degree[dep] -= 1
                    if in_degree[dep] == 0:
                        queue.append(dep)

            waves.append(wave)

        return waves
```

---

### Pattern 3: Consensus Building (Voting Pattern)

**Purpose**: Multiple agents analyze same input, results are aggregated via voting or confidence weighting.

**Use Case**: Critical decision (approve refund, escalate to manager) → 3 agents evaluate → Vote or weighted average → Final decision based on consensus.

```python
# ventros_adk/orchestration/consensus_agent.py

from google.adk import BaseAgent, LlmAgent, ParallelAgent
from google.adk.models import GeminiModel
from typing import Dict, Any, List
import statistics

class ConsensusAgent(BaseAgent):
    """
    Multiple agents vote on decision, consensus wins.

    Use for high-stakes decisions:
    - Approve refund/chargeback
    - Escalate to manager
    - Apply discount > 30%
    - Close high-value deal
    """

    def __init__(
        self,
        memory_service: 'VentrosMemoryService',
        session_manager: 'SessionManager',
        voting_agents: List[LlmAgent],
        consensus_threshold: float = 0.6,  # 60% agreement required
        weighting: str = "equal",  # "equal" or "confidence"
    ):
        super().__init__(name="consensus_agent")

        self.memory_service = memory_service
        self.session_manager = session_manager
        self.voting_agents = voting_agents
        self.consensus_threshold = consensus_threshold
        self.weighting = weighting

        # Parallel executor for voting
        self.parallel_executor = ParallelAgent(
            name="parallel_voters",
            agents=voting_agents,
        )

    def run(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Execute consensus voting.
        """
        decision_prompt = input_data.get("decision_prompt")
        contact_id = input_data.get("contact_id")

        # Step 1: All agents vote in parallel
        votes = self.parallel_executor.run({
            "prompt": decision_prompt,
            "contact_id": contact_id,
        })

        # Step 2: Aggregate votes
        if self.weighting == "equal":
            consensus = self._equal_weight_voting(votes)
        else:  # confidence-weighted
            consensus = self._confidence_weighted_voting(votes)

        # Step 3: Determine final decision
        decision_made = consensus["consensus_score"] >= self.consensus_threshold

        return {
            "decision": consensus["majority_vote"],
            "confidence": consensus["consensus_score"],
            "threshold_met": decision_made,
            "votes": votes,
            "reasoning": consensus["reasoning"],
        }

    def _equal_weight_voting(self, votes: Dict[str, Any]) -> Dict[str, Any]:
        """
        Simple majority voting (each agent has equal weight).
        """
        # Assume votes structure: {agent_name: {"vote": "approve"/"reject", "confidence": 0.8, ...}}

        vote_counts = {"approve": 0, "reject": 0, "abstain": 0}
        total_votes = len(votes)

        for agent_name, vote_data in votes.items():
            vote = vote_data.get("vote", "abstain")
            vote_counts[vote] += 1

        # Majority vote
        majority_vote = max(vote_counts, key=vote_counts.get)
        consensus_score = vote_counts[majority_vote] / total_votes

        return {
            "majority_vote": majority_vote,
            "consensus_score": consensus_score,
            "vote_counts": vote_counts,
            "reasoning": f"{vote_counts[majority_vote]}/{total_votes} agents voted {majority_vote}",
        }

    def _confidence_weighted_voting(self, votes: Dict[str, Any]) -> Dict[str, Any]:
        """
        Weighted voting based on each agent's confidence.
        """
        weighted_scores = {"approve": 0.0, "reject": 0.0, "abstain": 0.0}
        total_weight = 0.0

        for agent_name, vote_data in votes.items():
            vote = vote_data.get("vote", "abstain")
            confidence = vote_data.get("confidence", 0.5)

            weighted_scores[vote] += confidence
            total_weight += confidence

        # Normalize
        if total_weight > 0:
            for vote in weighted_scores:
                weighted_scores[vote] /= total_weight

        # Majority by weighted score
        majority_vote = max(weighted_scores, key=weighted_scores.get)
        consensus_score = weighted_scores[majority_vote]

        return {
            "majority_vote": majority_vote,
            "consensus_score": consensus_score,
            "weighted_scores": weighted_scores,
            "reasoning": f"Confidence-weighted: {majority_vote} = {consensus_score:.2%}",
        }


# Example: Refund approval consensus
def create_refund_approval_consensus(
    memory_service: 'VentrosMemoryService',
    session_manager: 'SessionManager',
) -> ConsensusAgent:
    """
    3 agents vote on refund approval:
    - Financial risk assessor
    - Customer satisfaction analyst
    - Policy compliance checker
    """

    # Agent 1: Financial Risk
    financial_agent = LlmAgent(
        name="financial_risk_assessor",
        model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.2),
        system_prompt="""
        Você avalia risco financeiro de reembolso.

        Considere:
        - Valor do reembolso vs LTV do cliente
        - Histórico de chargebacks
        - Margem de lucro
        - Impacto no cash flow

        Vote:
        - "approve" se risco baixo (< 5% LTV impact)
        - "reject" se risco alto (> 20% LTV impact)
        - Retorne confidence (0-1) e justificativa
        """,
        memory=memory_service,
    )

    # Agent 2: Customer Satisfaction
    satisfaction_agent = LlmAgent(
        name="customer_satisfaction_analyst",
        model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.3),
        system_prompt="""
        Você avalia impacto na satisfação do cliente.

        Considere:
        - Sentiment score histórico
        - NPS e CSAT
        - Probabilidade de churn se negar
        - Valor de relacionamento longo prazo

        Vote:
        - "approve" se negar causaria churn alto (> 60% prob)
        - "reject" se cliente tem histórico de abuso
        - Retorne confidence (0-1) e justificativa
        """,
        memory=memory_service,
    )

    # Agent 3: Policy Compliance
    policy_agent = LlmAgent(
        name="policy_compliance_checker",
        model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.1),
        system_prompt="""
        Você verifica compliance com políticas de reembolso.

        Políticas:
        - Reembolso integral: < 30 dias da compra
        - Reembolso parcial: 30-90 dias
        - Sem reembolso: > 90 dias (exceto defeito comprovado)
        - Limite: 2 reembolsos por cliente por ano

        Vote:
        - "approve" se dentro da política
        - "reject" se viola política (e não há exceção válida)
        - Retorne confidence (1.0 para regra clara, < 1.0 para ambíguo)
        """,
        memory=memory_service,
    )

    consensus = ConsensusAgent(
        memory_service=memory_service,
        session_manager=session_manager,
        voting_agents=[financial_agent, satisfaction_agent, policy_agent],
        consensus_threshold=0.66,  # 2 out of 3 must agree
        weighting="confidence",
    )

    return consensus
```

---

### Pattern 4: Human-in-the-Loop (HITL)

**Purpose**: Agent requests human approval/input for critical decisions.

**Use Case**: AI agent prepares action → Sends to human for approval → Human approves/rejects/modifies → Agent continues.

```python
# ventros_adk/orchestration/human_in_loop_agent.py

from google.adk import BaseAgent, LlmAgent
from google.adk.models import GeminiModel
from typing import Dict, Any, Optional
import time

class HumanInLoopAgent(BaseAgent):
    """
    Agent that requests human approval for critical actions.

    Flow:
    1. AI agent analyzes and prepares action
    2. Sends approval request to human (via webhook/queue)
    3. Waits for human response (with timeout)
    4. If approved: execute action
    5. If rejected: explain to customer and log
    6. If timeout: fallback behavior
    """

    def __init__(
        self,
        memory_service: 'VentrosMemoryService',
        session_manager: 'SessionManager',
        base_agent: LlmAgent,
        approval_callback: callable,
        timeout_seconds: int = 300,  # 5 minutes
        fallback_on_timeout: str = "reject",  # "reject", "approve", "escalate"
    ):
        super().__init__(name="human_in_loop_agent")

        self.memory_service = memory_service
        self.session_manager = session_manager
        self.base_agent = base_agent
        self.approval_callback = approval_callback  # Function to request human approval
        self.timeout_seconds = timeout_seconds
        self.fallback_on_timeout = fallback_on_timeout

    def run(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Execute agent with human approval step.
        """
        contact_id = input_data.get("contact_id")
        message = input_data.get("message")

        # Step 1: AI agent analyzes and prepares action
        ai_result = self.base_agent.run({
            "contact_id": contact_id,
            "message": message,
        })

        action_proposed = ai_result.get("action")
        confidence = ai_result.get("confidence", 0.0)
        reasoning = ai_result.get("reasoning")

        # Determine if human approval needed
        needs_approval = self._needs_human_approval(action_proposed, confidence)

        if not needs_approval:
            # Low-risk action, execute directly
            return {
                "action": action_proposed,
                "executed": True,
                "human_approved": False,
                "reasoning": "Low-risk action, no approval needed",
            }

        # Step 2: Request human approval
        approval_request = {
            "contact_id": contact_id,
            "message": message,
            "action_proposed": action_proposed,
            "ai_confidence": confidence,
            "ai_reasoning": reasoning,
            "requested_at": time.time(),
        }

        approval_request_id = self.approval_callback(approval_request)

        # Step 3: Wait for human response (polling or webhook)
        approval_result = self._wait_for_approval(
            approval_request_id,
            timeout_seconds=self.timeout_seconds,
        )

        # Step 4: Handle approval result
        if approval_result.get("status") == "approved":
            return {
                "action": action_proposed,
                "executed": True,
                "human_approved": True,
                "approved_by": approval_result.get("approved_by"),
                "approval_notes": approval_result.get("notes"),
            }

        elif approval_result.get("status") == "rejected":
            return {
                "action": None,
                "executed": False,
                "human_approved": False,
                "rejected_by": approval_result.get("rejected_by"),
                "rejection_reason": approval_result.get("reason"),
            }

        elif approval_result.get("status") == "modified":
            # Human modified the action
            modified_action = approval_result.get("modified_action")
            return {
                "action": modified_action,
                "executed": True,
                "human_approved": True,
                "modified_by": approval_result.get("modified_by"),
                "original_action": action_proposed,
            }

        else:  # timeout
            if self.fallback_on_timeout == "reject":
                return {
                    "action": None,
                    "executed": False,
                    "timeout": True,
                    "reasoning": "Human approval timeout, rejecting by default",
                }
            elif self.fallback_on_timeout == "approve":
                return {
                    "action": action_proposed,
                    "executed": True,
                    "timeout": True,
                    "reasoning": "Human approval timeout, approving by default",
                }
            else:  # escalate
                return {
                    "action": None,
                    "executed": False,
                    "timeout": True,
                    "escalated": True,
                    "reasoning": "Human approval timeout, escalated to manager",
                }

    def _needs_human_approval(self, action: Dict[str, Any], confidence: float) -> bool:
        """
        Determine if action requires human approval.

        Rules:
        - Discount > 30%: always require approval
        - Refund > $500: always require approval
        - Contract modification: always require approval
        - AI confidence < 70%: require approval
        - etc.
        """
        action_type = action.get("type")

        # High-risk actions always need approval
        if action_type in ["refund", "discount", "contract_modification", "escalate_to_legal"]:
            if action.get("amount", 0) > 500:
                return True

        # Low confidence needs approval
        if confidence < 0.7:
            return True

        return False

    def _wait_for_approval(
        self,
        approval_request_id: str,
        timeout_seconds: int,
    ) -> Dict[str, Any]:
        """
        Wait for human approval (polling or webhook-based).

        In production, this would:
        1. Send webhook to approval system (Slack, custom UI, etc)
        2. Poll database/queue for approval status
        3. Return result or timeout
        """
        start_time = time.time()

        while (time.time() - start_time) < timeout_seconds:
            # Poll approval status
            # In production: query database or check Redis
            status = self._check_approval_status(approval_request_id)

            if status.get("status") in ["approved", "rejected", "modified"]:
                return status

            time.sleep(5)  # Poll every 5 seconds

        # Timeout
        return {"status": "timeout"}

    def _check_approval_status(self, approval_request_id: str) -> Dict[str, Any]:
        """
        Check current approval status.

        In production:
        - Query approval_requests table in database
        - Check Redis cache
        - Listen to webhook callback

        Returns:
        {
            "status": "pending" | "approved" | "rejected" | "modified",
            "approved_by": "user_id",
            "notes": "...",
            "modified_action": {...} (if modified),
        }
        """
        # Placeholder - would query actual approval system
        return {"status": "pending"}
```

---

## Tool Calling Patterns

### Pattern 1: Function Tools (Inline Python Functions)

**Purpose**: Simple Python functions exposed as tools to LLM.

```python
# ventros_adk/tools/crm_tools.py

from google.adk.tools import FunctionTool
from typing import Dict, Any, List, Optional
import requests

class CRMTools:
    """
    Collection of CRM operation tools for agents.
    """

    def __init__(
        self,
        api_base_url: str,
        api_key: str,
        memory_service: 'VentrosMemoryService',
    ):
        self.api_base_url = api_base_url
        self.api_key = api_key
        self.memory_service = memory_service

    # === Contact Management Tools ===

    @FunctionTool
    def get_contact_profile(self, contact_id: str) -> Dict[str, Any]:
        """
        Get full contact profile with all metadata.

        Args:
            contact_id: UUID of contact

        Returns:
            Full contact profile including tags, custom fields, profile picture
        """
        response = requests.get(
            f"{self.api_base_url}/api/v1/contacts/{contact_id}",
            headers={"Authorization": f"Bearer {self.api_key}"},
            timeout=10,
        )
        response.raise_for_status()
        return response.json()

    @FunctionTool
    def update_contact_tags(
        self,
        contact_id: str,
        tags: List[str],
        operation: str = "add",  # "add", "remove", "replace"
    ) -> Dict[str, Any]:
        """
        Update contact tags.

        Args:
            contact_id: UUID of contact
            tags: List of tag strings
            operation: "add" (append), "remove" (delete), "replace" (overwrite)

        Returns:
            Updated contact with new tags
        """
        response = requests.patch(
            f"{self.api_base_url}/api/v1/contacts/{contact_id}/tags",
            json={"tags": tags, "operation": operation},
            headers={"Authorization": f"Bearer {self.api_key}"},
            timeout=10,
        )
        response.raise_for_status()
        return response.json()

    @FunctionTool
    def update_pipeline_status(
        self,
        contact_id: str,
        pipeline_id: str,
        status_id: str,
        reason: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Move contact to different pipeline status.

        Args:
            contact_id: UUID of contact
            pipeline_id: UUID of pipeline
            status_id: UUID of target status
            reason: Optional reason for status change

        Returns:
            Updated contact with new pipeline status
        """
        response = requests.post(
            f"{self.api_base_url}/api/v1/contacts/{contact_id}/pipeline-status",
            json={
                "pipeline_id": pipeline_id,
                "status_id": status_id,
                "reason": reason,
            },
            headers={"Authorization": f"Bearer {self.api_key}"},
            timeout=10,
        )
        response.raise_for_status()
        return response.json()

    # === Messaging Tools ===

    @FunctionTool
    def send_message(
        self,
        contact_id: str,
        text: str,
        channel_id: Optional[str] = None,
        content_type: str = "text",
        metadata: Optional[Dict] = None,
    ) -> Dict[str, Any]:
        """
        Send message to contact.

        Args:
            contact_id: UUID of contact
            text: Message text
            channel_id: Optional UUID of specific channel (auto-select if None)
            content_type: "text", "image", "video", "audio", "voice", "document"
            metadata: Optional metadata dict

        Returns:
            Created message object with ID and status
        """
        response = requests.post(
            f"{self.api_base_url}/api/v1/messages",
            json={
                "contact_id": contact_id,
                "text": text,
                "channel_id": channel_id,
                "content_type": content_type,
                "from_me": True,
                "source": "bot",  # AI-generated
                "metadata": metadata or {},
            },
            headers={"Authorization": f"Bearer {self.api_key}"},
            timeout=15,
        )
        response.raise_for_status()
        return response.json()

    @FunctionTool
    def send_template_message(
        self,
        contact_id: str,
        template_name: str,
        variables: Dict[str, str],
        channel_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Send WhatsApp template message (for initial outreach after 24h window).

        Args:
            contact_id: UUID of contact
            template_name: Name of approved WhatsApp template
            variables: Template variable substitutions
            channel_id: Optional UUID of specific channel

        Returns:
            Created message object
        """
        response = requests.post(
            f"{self.api_base_url}/api/v1/messages/template",
            json={
                "contact_id": contact_id,
                "template_name": template_name,
                "variables": variables,
                "channel_id": channel_id,
            },
            headers={"Authorization": f"Bearer {self.api_key}"},
            timeout=15,
        )
        response.raise_for_status()
        return response.json()

    # === Session Management Tools ===

    @FunctionTool
    def assign_agent_to_session(
        self,
        session_id: str,
        agent_id: str,
        reason: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Assign (or transfer) session to different agent.

        Args:
            session_id: UUID of session
            agent_id: UUID of agent to assign
            reason: Optional reason for assignment/transfer

        Returns:
            Updated session object
        """
        response = requests.post(
            f"{self.api_base_url}/api/v1/sessions/{session_id}/assign",
            json={
                "agent_id": agent_id,
                "reason": reason or "AI routing",
            },
            headers={"Authorization": f"Bearer {self.api_key}"},
            timeout=10,
        )
        response.raise_for_status()
        return response.json()

    @FunctionTool
    def close_session(
        self,
        session_id: str,
        summary: str,
        sentiment: Optional[str] = None,
        resolution: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Close session with summary.

        Args:
            session_id: UUID of session
            summary: AI-generated session summary
            sentiment: "positive", "neutral", "negative"
            resolution: "resolved", "escalated", "unresolved"

        Returns:
            Closed session object with final summary
        """
        response = requests.post(
            f"{self.api_base_url}/api/v1/sessions/{session_id}/close",
            json={
                "summary": summary,
                "sentiment": sentiment,
                "resolution": resolution,
            },
            headers={"Authorization": f"Bearer {self.api_key}"},
            timeout=10,
        )
        response.raise_for_status()
        return response.json()

    # === Event Tracking Tools ===

    @FunctionTool
    def create_contact_event(
        self,
        contact_id: str,
        category: str,
        event_type: str,
        description: Optional[str] = None,
        metadata: Optional[Dict] = None,
        priority: str = "normal",
    ) -> Dict[str, Any]:
        """
        Create contact event for tracking.

        Args:
            contact_id: UUID of contact
            category: "interaction", "milestone", "issue", "purchase", "engagement", "system"
            event_type: Specific event type within category
            description: Optional description
            metadata: Additional event data
            priority: "low", "normal", "high", "urgent"

        Returns:
            Created event object
        """
        response = requests.post(
            f"{self.api_base_url}/api/v1/contact-events",
            json={
                "contact_id": contact_id,
                "category": category,
                "event_type": event_type,
                "description": description,
                "metadata": metadata or {},
                "priority": priority,
            },
            headers={"Authorization": f"Bearer {self.api_key}"},
            timeout=10,
        )
        response.raise_for_status()
        return response.json()

    # === Note Management Tools ===

    @FunctionTool
    def create_note(
        self,
        contact_id: str,
        content: str,
        note_type: str = "general",
        visibility: str = "internal",
        agent_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Create note attached to contact.

        Args:
            contact_id: UUID of contact
            content: Note content (markdown supported)
            note_type: "general", "call_log", "meeting", "follow_up", "internal"
            visibility: "internal" (team only), "shared" (customer can see)
            agent_id: Optional UUID of agent creating note

        Returns:
            Created note object
        """
        response = requests.post(
            f"{self.api_base_url}/api/v1/notes",
            json={
                "contact_id": contact_id,
                "content": content,
                "note_type": note_type,
                "visibility": visibility,
                "agent_id": agent_id,
            },
            headers={"Authorization": f"Bearer {self.api_key}"},
            timeout=10,
        )
        response.raise_for_status()
        return response.json()

    # === Automation Tools ===

    @FunctionTool
    def trigger_sequence(
        self,
        contact_id: str,
        sequence_id: str,
        variables: Optional[Dict] = None,
    ) -> Dict[str, Any]:
        """
        Enroll contact in automation sequence.

        Args:
            contact_id: UUID of contact
            sequence_id: UUID of sequence to trigger
            variables: Optional variables for personalization

        Returns:
            Enrollment status
        """
        response = requests.post(
            f"{self.api_base_url}/api/v1/sequences/{sequence_id}/enroll",
            json={
                "contact_id": contact_id,
                "variables": variables or {},
            },
            headers={"Authorization": f"Bearer {self.api_key}"},
            timeout=10,
        )
        response.raise_for_status()
        return response.json()

    @FunctionTool
    def schedule_follow_up(
        self,
        contact_id: str,
        scheduled_at: str,  # ISO 8601 datetime
        message: str,
        channel_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Schedule follow-up message for future delivery.

        Args:
            contact_id: UUID of contact
            scheduled_at: ISO 8601 datetime string (e.g., "2025-02-01T14:00:00Z")
            message: Message to send
            channel_id: Optional channel ID

        Returns:
            Scheduled message object
        """
        response = requests.post(
            f"{self.api_base_url}/api/v1/messages/schedule",
            json={
                "contact_id": contact_id,
                "scheduled_at": scheduled_at,
                "text": message,
                "channel_id": channel_id,
                "source": "bot",
            },
            headers={"Authorization": f"Bearer {self.api_key}"},
            timeout=10,
        )
        response.raise_for_status()
        return response.json()


# Usage in agent
def create_agent_with_crm_tools(
    memory_service: 'VentrosMemoryService',
    session_manager: 'SessionManager',
    api_base_url: str,
    api_key: str,
) -> LlmAgent:
    """
    Create agent with full CRM tool suite.
    """
    crm_tools = CRMTools(
        api_base_url=api_base_url,
        api_key=api_key,
        memory_service=memory_service,
    )

    agent = LlmAgent(
        name="crm_agent_with_tools",
        model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.6),
        system_prompt="""
        Você é um agente de CRM com acesso a ferramentas de operação.

        Ferramentas disponíveis:
        - get_contact_profile: Buscar perfil completo do contato
        - update_contact_tags: Adicionar/remover tags
        - update_pipeline_status: Mover contato no pipeline
        - send_message: Enviar mensagem
        - send_template_message: Enviar template WhatsApp
        - assign_agent_to_session: Transferir para agente
        - close_session: Encerrar sessão
        - create_contact_event: Registrar evento
        - create_note: Criar nota interna
        - trigger_sequence: Iniciar automação
        - schedule_follow_up: Agendar follow-up

        Use as ferramentas apropriadamente baseado na conversa.
        Sempre confirme ações importantes antes de executar.
        """,
        tools=[
            crm_tools.get_contact_profile,
            crm_tools.update_contact_tags,
            crm_tools.update_pipeline_status,
            crm_tools.send_message,
            crm_tools.send_template_message,
            crm_tools.assign_agent_to_session,
            crm_tools.close_session,
            crm_tools.create_contact_event,
            crm_tools.create_note,
            crm_tools.trigger_sequence,
            crm_tools.schedule_follow_up,
        ],
        memory=memory_service,
    )

    return agent
```

---

### Pattern 2: Agent-as-Tool (Sub-Agent Composition)

**Purpose**: Use entire agents as tools for other agents.

```python
# ventros_adk/tools/agent_tools.py

from google.adk.tools import AgentTool
from google.adk import LlmAgent
from typing import Dict, Any

class ResearchAssistantAgent(LlmAgent):
    """
    Specialized research agent that can be used as tool by other agents.
    """

    def __init__(self, memory_service: 'VentrosMemoryService'):
        super().__init__(
            name="research_assistant",
            model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.3),
            system_prompt="""
            Você é um assistente de pesquisa especializado.
            Quando recebe uma pergunta, você:
            1. Busca informações relevantes na memória
            2. Sintetiza resposta concisa e factual
            3. Cita fontes quando possível

            Seja objetivo e preciso.
            """,
            tools=[FunctionTool(self._search_memory)],
            memory=memory_service,
        )

        self.memory_service = memory_service

    def _search_memory(self, query: str, contact_id: str) -> Dict[str, Any]:
        return self.memory_service.search_memory(
            query=query,
            contact_id=contact_id,
            strategy_name="vector_70_30",
            limit=15,
        )


class MainAgent(LlmAgent):
    """
    Main agent that uses research assistant as a tool.
    """

    def __init__(
        self,
        memory_service: 'VentrosMemoryService',
        research_assistant: ResearchAssistantAgent,
    ):
        # Convert research assistant to tool
        research_tool = AgentTool(
            agent=research_assistant,
            description="""
            Use this tool to research information about the contact.
            Input: question to research
            Output: factual answer with sources
            """
        )

        super().__init__(
            name="main_agent",
            model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.7),
            system_prompt="""
            Você é um agente principal de atendimento.

            Quando precisar de informações detalhadas sobre o contato,
            use a ferramenta research_assistant.

            Exemplo:
            User: "Qual foi a última compra desse cliente?"
            You: [call research_assistant with query="última compra do cliente"]
            Research Assistant: "Última compra: Plano Pro em 15/01/2025, R$ 299"
            You: "O cliente fez a última compra em 15 de janeiro..."
            """,
            tools=[research_tool],
            memory=memory_service,
        )
```

---

## Sophisticated Session Manager

```python
# ventros_adk/session/session_manager.py

from typing import Dict, Any, List, Optional
from datetime import datetime, timedelta
import uuid
import json
import logging

class SessionManager:
    """
    Sophisticated session manager for Ventros ADK agents.

    Responsibilities:
    1. Session lifecycle (create, get, update, close)
    2. Context assembly from Go Memory Service
    3. Conversation history tracking (within context window)
    4. Agent state persistence
    5. Session metadata management
    6. Timeout handling
    7. Automatic summary generation on close
    """

    def __init__(
        self,
        memory_service: 'VentrosMemoryService',
        context_window: int = 2_000_000,  # Gemini 2.0 Flash: 2M tokens
        session_timeout_minutes: int = 30,
        auto_summarize: bool = True,
    ):
        self.memory_service = memory_service
        self.context_window = context_window
        self.session_timeout_minutes = session_timeout_minutes
        self.auto_summarize = auto_summarize
        self.logger = logging.getLogger(__name__)

        # In-memory session cache (in production: Redis)
        self._sessions: Dict[str, 'Session'] = {}

    def get_or_create_session(
        self,
        contact_id: str,
        channel_id: Optional[str] = None,
        metadata: Optional[Dict] = None,
    ) -> 'Session':
        """
        Get existing active session or create new one.

        Sessions are scoped per contact (1 active session per contact).
        If previous session timed out, create new one.
        """
        # Check for active session
        active_session = self._find_active_session(contact_id)

        if active_session:
            # Extend session timeout
            active_session.extend_timeout(self.session_timeout_minutes)
            return active_session

        # Create new session
        session = Session(
            session_id=str(uuid.uuid4()),
            contact_id=contact_id,
            channel_id=channel_id,
            timeout_minutes=self.session_timeout_minutes,
            metadata=metadata or {},
        )

        self._sessions[session.session_id] = session
        self.logger.info(f"Created new session {session.session_id} for contact {contact_id}")

        return session

    def get_session(self, session_id: str) -> Optional['Session']:
        """Get session by ID."""
        session = self._sessions.get(session_id)

        if session and session.is_expired():
            self.logger.info(f"Session {session_id} expired, closing")
            self.close_session(session_id, reason="timeout")
            return None

        return session

    def close_session(
        self,
        session_id: str,
        reason: Optional[str] = None,
        summary: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Close session and add to memory.

        Steps:
        1. Generate summary (if not provided and auto_summarize=True)
        2. Extract sentiment, topics, key entities
        3. Call memory_service.add_session_to_memory()
        4. Remove from active sessions
        """
        session = self._sessions.get(session_id)

        if not session:
            return {"success": False, "error": "Session not found"}

        # Generate summary if needed
        if self.auto_summarize and not summary:
            summary = self._generate_summary(session)

        # Extract metadata
        sentiment, sentiment_score = self._analyze_sentiment(session)
        topics = self._extract_topics(session)
        key_entities = self._extract_entities(session)

        # Add to memory
        memory_result = self.memory_service.add_session_to_memory(
            session_id=session_id,
            contact_id=session.contact_id,
            summary=summary or "Session completed",
            sentiment=sentiment,
            sentiment_score=sentiment_score,
            topics=topics,
            key_entities=key_entities,
            metadata={
                "close_reason": reason,
                "message_count": len(session.messages),
                "agent_count": len(session.agents_involved),
                "duration_seconds": session.duration_seconds(),
            },
        )

        # Mark as closed
        session.status = "closed"
        session.closed_at = datetime.now()

        # Remove from active sessions (move to history)
        del self._sessions[session_id]

        self.logger.info(f"Closed session {session_id}, reason={reason}")

        return {
            "success": True,
            "session_id": session_id,
            "summary": summary,
            "memory_result": memory_result,
        }

    def add_message_to_session(
        self,
        session_id: str,
        role: str,  # "user", "assistant", "system"
        content: str,
        metadata: Optional[Dict] = None,
    ) -> None:
        """
        Add message to session conversation history.
        """
        session = self.get_session(session_id)

        if not session:
            self.logger.warning(f"Cannot add message: session {session_id} not found")
            return

        session.add_message(role, content, metadata)

        # Trim context if exceeding window
        self._trim_context_if_needed(session)

    def get_session_context(
        self,
        session_id: str,
        include_memory: bool = True,
        memory_query: Optional[str] = None,
        memory_strategy: str = "balanced",
    ) -> Dict[str, Any]:
        """
        Assemble full context for agent execution.

        Returns:
        {
            "session_id": "...",
            "contact_id": "...",
            "conversation_history": [...],  // Recent messages in session
            "memory_context": {...},         // Retrieved from Go Memory Service
            "session_metadata": {...},
            "token_count_estimate": 15000,
        }
        """
        session = self.get_session(session_id)

        if not session:
            return {"error": "Session not found"}

        context = {
            "session_id": session.session_id,
            "contact_id": session.contact_id,
            "conversation_history": session.get_conversation_history(),
            "session_metadata": session.metadata,
            "created_at": session.created_at.isoformat(),
            "agents_involved": session.agents_involved,
        }

        # Add memory context if requested
        if include_memory:
            # Use last user message as query if not provided
            if not memory_query and session.messages:
                last_user_msg = next(
                    (m for m in reversed(session.messages) if m["role"] == "user"),
                    None
                )
                memory_query = last_user_msg["content"] if last_user_msg else "recent context"

            memory_context = self.memory_service.search_memory(
                query=memory_query,
                contact_id=session.contact_id,
                strategy_name=memory_strategy,
                limit=20,
            )

            context["memory_context"] = memory_context

        # Estimate token count
        context["token_count_estimate"] = self._estimate_tokens(context)

        return context

    def log_agent_execution(
        self,
        session_id: str,
        agent_id: str,
        category: str,
        confidence: float,
        input_data: Dict,
        output_data: Dict,
    ) -> None:
        """
        Log agent execution in session for debugging and analysis.
        """
        session = self.get_session(session_id)

        if not session:
            return

        session.agents_involved.add(agent_id)
        session.agent_executions.append({
            "timestamp": datetime.now().isoformat(),
            "agent_id": agent_id,
            "category": category,
            "confidence": confidence,
            "input": input_data,
            "output": output_data,
        })

    # Private methods

    def _find_active_session(self, contact_id: str) -> Optional['Session']:
        """Find active session for contact."""
        for session in self._sessions.values():
            if session.contact_id == contact_id and not session.is_expired():
                return session
        return None

    def _generate_summary(self, session: 'Session') -> str:
        """
        Generate AI summary of session using LLM.
        """
        from google.adk.models import GeminiModel

        model = GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.3)

        conversation = session.get_conversation_history()

        prompt = f"""
        Summarize this customer conversation in 2-3 sentences.

        Conversation:
        {json.dumps(conversation, indent=2)}

        Focus on:
        - Main topic/purpose
        - Key outcomes or decisions
        - Next steps (if any)

        Summary:
        """

        summary = model.generate(prompt)
        return summary

    def _analyze_sentiment(self, session: 'Session') -> tuple[str, float]:
        """
        Analyze sentiment of conversation.

        Returns: (sentiment_label, confidence_score)
        """
        # Simplified - in production, use LLM or sentiment model
        # For now, return neutral
        return ("neutral", 0.5)

    def _extract_topics(self, session: 'Session') -> List[str]:
        """Extract main topics discussed."""
        # Simplified - in production, use LLM topic extraction
        return []

    def _extract_entities(self, session: 'Session') -> Dict[str, Any]:
        """Extract key entities (products, dates, amounts, etc)."""
        # Simplified - in production, use LLM NER
        return {}

    def _trim_context_if_needed(self, session: 'Session') -> None:
        """
        Trim conversation history if exceeding context window.

        Strategy: Keep first message (context) + recent N messages.
        """
        current_tokens = self._estimate_tokens({"conversation_history": session.messages})

        if current_tokens > self.context_window * 0.8:  # 80% threshold
            # Keep first message + last 50 messages
            if len(session.messages) > 51:
                first_msg = session.messages[0]
                recent_msgs = session.messages[-50:]
                session.messages = [first_msg] + recent_msgs

                self.logger.info(f"Trimmed session {session.session_id} context")

    def _estimate_tokens(self, data: Dict) -> int:
        """
        Rough token estimation (4 chars ≈ 1 token).
        """
        text = json.dumps(data)
        return len(text) // 4


class Session:
    """
    Session object representing conversation context.
    """

    def __init__(
        self,
        session_id: str,
        contact_id: str,
        channel_id: Optional[str],
        timeout_minutes: int,
        metadata: Dict,
    ):
        self.session_id = session_id
        self.contact_id = contact_id
        self.channel_id = channel_id
        self.created_at = datetime.now()
        self.last_activity_at = datetime.now()
        self.closed_at: Optional[datetime] = None
        self.status = "active"  # "active", "closed", "expired"
        self.timeout_minutes = timeout_minutes

        self.messages: List[Dict] = []
        self.agents_involved: set = set()
        self.agent_executions: List[Dict] = []
        self.metadata: Dict = metadata

    def add_message(self, role: str, content: str, metadata: Optional[Dict] = None) -> None:
        """Add message to conversation history."""
        self.messages.append({
            "role": role,
            "content": content,
            "timestamp": datetime.now().isoformat(),
            "metadata": metadata or {},
        })
        self.last_activity_at = datetime.now()

    def get_conversation_history(self, limit: Optional[int] = None) -> List[Dict]:
        """Get conversation history (optionally limited to last N messages)."""
        if limit:
            return self.messages[-limit:]
        return self.messages

    def is_expired(self) -> bool:
        """Check if session has expired."""
        if self.status != "active":
            return True

        timeout_threshold = datetime.now() - timedelta(minutes=self.timeout_minutes)
        return self.last_activity_at < timeout_threshold

    def extend_timeout(self, additional_minutes: int) -> None:
        """Extend session timeout."""
        self.last_activity_at = datetime.now()

    def duration_seconds(self) -> int:
        """Calculate session duration in seconds."""
        end_time = self.closed_at or datetime.now()
        return int((end_time - self.created_at).total_seconds())
```

---

**(File is getting long - continuing in Part 3 with ReAct, Self-Reflection, Complete Examples, and Production Deployment)**
# Python ADK Architecture - Part 3 (Final)
**ReAct, Self-Reflection, Complete Examples, and Production Deployment**

---

## Table of Contents (Part 3)
1. [ReAct Pattern Deep Dive](#react-pattern-deep-dive)
2. [Self-Reflection & Planning](#self-reflection--planning)
3. [Complete End-to-End Examples](#complete-end-to-end-examples)
4. [Production Deployment Guide](#production-deployment-guide)
5. [Performance Optimization](#performance-optimization)
6. [Monitoring & Observability](#monitoring--observability)
7. [Best Practices Summary](#best-practices-summary)

---

## ReAct Pattern Deep Dive

**ReAct = Reasoning + Acting**

The ReAct pattern interleaves reasoning traces with action execution, allowing agents to dynamically reason about what to do next.

### Core ReAct Loop

```
Thought → Action → Observation → Reflection → [Repeat until task complete]
```

### Implementation

```python
# ventros_adk/patterns/react_agent.py

from google.adk import BaseAgent, LlmAgent
from google.adk.models import GeminiModel
from google.adk.tools import FunctionTool
from typing import Dict, Any, List, Optional
import logging

class ReActAgent(BaseAgent):
    """
    Full ReAct (Reasoning + Acting) implementation.

    Process:
    1. THOUGHT: Reason about current state and what to do next
    2. ACTION: Execute chosen tool/action
    3. OBSERVATION: Observe and parse result
    4. REFLECTION: Evaluate if action was successful, adjust if needed
    5. Repeat until task complete or max iterations

    Based on: https://arxiv.org/abs/2210.03629
    """

    def __init__(
        self,
        memory_service: 'VentrosMemoryService',
        session_manager: 'SessionManager',
        tools: List[FunctionTool],
        max_iterations: int = 10,
        verbose: bool = True,
    ):
        super().__init__(name="react_agent")

        self.memory_service = memory_service
        self.session_manager = session_manager
        self.tools = {tool.__name__: tool for tool in tools}
        self.max_iterations = max_iterations
        self.verbose = verbose
        self.logger = logging.getLogger(__name__)

        # ReAct reasoning model (use thinking model for deep reasoning)
        self.model = GeminiModel(
            model_name="gemini-2.0-flash-thinking-exp-01-21",
            temperature=0.5,
            max_output_tokens=8192,
            thinking_mode="thinking",
        )

    def run(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Execute ReAct loop until task complete.
        """
        task = input_data.get("task")
        contact_id = input_data.get("contact_id")

        self.logger.info(f"ReAct Agent started: task='{task}'")

        # Initialize trace
        trace: List[Dict[str, Any]] = []
        iteration = 0
        task_complete = False
        final_answer = None

        while iteration < self.max_iterations and not task_complete:
            iteration += 1

            if self.verbose:
                print(f"\n{'='*60}")
                print(f"ITERATION {iteration}/{self.max_iterations}")
                print(f"{'='*60}")

            # === STEP 1: THOUGHT (Reasoning) ===
            thought_prompt = self._build_thought_prompt(
                task=task,
                contact_id=contact_id,
                trace=trace,
                iteration=iteration,
            )

            thought_response = self.model.generate(thought_prompt)
            thought = self._parse_thought(thought_response)

            if self.verbose:
                print(f"\n💭 THOUGHT:\n{thought['reasoning']}")

            # === STEP 2: ACTION (Tool Selection & Execution) ===
            action = thought.get("action")

            if not action or action == "FINISH":
                # Agent decided task is complete
                task_complete = True
                final_answer = thought.get("final_answer")

                if self.verbose:
                    print(f"\n✅ TASK COMPLETE")
                    print(f"Final Answer: {final_answer}")

                break

            if self.verbose:
                print(f"\n🔧 ACTION: {action['tool']}")
                print(f"Args: {action['args']}")

            # Execute action
            observation = self._execute_action(action)

            if self.verbose:
                print(f"\n👁️ OBSERVATION:")
                print(f"{observation['result']}")

            # === STEP 3: REFLECTION (Self-Evaluation) ===
            reflection = self._reflect_on_action(
                thought=thought,
                action=action,
                observation=observation,
            )

            if self.verbose:
                print(f"\n🤔 REFLECTION:")
                print(f"Success: {reflection['success']}")
                print(f"Reasoning: {reflection['reasoning']}")

            # Add to trace
            trace.append({
                "iteration": iteration,
                "thought": thought,
                "action": action,
                "observation": observation,
                "reflection": reflection,
            })

            # Check if reflection indicates task complete
            if reflection.get("task_complete"):
                task_complete = True
                final_answer = reflection.get("final_answer")

        # Return final result
        return {
            "task": task,
            "success": task_complete,
            "final_answer": final_answer,
            "iterations": iteration,
            "trace": trace,
            "reason_incomplete": "max_iterations_reached" if not task_complete else None,
        }

    def _build_thought_prompt(
        self,
        task: str,
        contact_id: str,
        trace: List[Dict],
        iteration: int,
    ) -> str:
        """
        Build prompt for reasoning step.
        """
        # Available tools description
        tools_desc = "\n".join([
            f"- {name}: {tool.__doc__}"
            for name, tool in self.tools.items()
        ])

        # Previous trace summary
        if trace:
            trace_summary = "\n\n".join([
                f"[Iteration {t['iteration']}]\n"
                f"Thought: {t['thought']['reasoning']}\n"
                f"Action: {t['action']['tool']}({t['action']['args']})\n"
                f"Observation: {t['observation']['result']}\n"
                f"Reflection: {t['reflection']['reasoning']}"
                for t in trace[-3:]  # Last 3 iterations
            ])
        else:
            trace_summary = "No previous actions."

        prompt = f"""
You are a ReAct agent solving a task through reasoning and acting.

**Task**: {task}
**Contact ID**: {contact_id}
**Current Iteration**: {iteration}/{self.max_iterations}

**Available Tools**:
{tools_desc}

**Previous Actions**:
{trace_summary}

---

**Instructions**:
1. ANALYZE the current situation:
   - What do you know so far?
   - What do you still need to discover?
   - Did previous actions succeed or fail?

2. REASON about the next step:
   - What is the most logical next action?
   - Why is this action appropriate?
   - What do you expect to learn/accomplish?

3. DECIDE on action:
   - Choose ONE tool to execute
   - OR decide task is complete (use "FINISH")

**Response Format** (JSON):
{{
    "reasoning": "Your detailed reasoning about what to do next...",
    "action": {{
        "tool": "tool_name",
        "args": {{"arg1": "value1", "arg2": "value2"}}
    }},
    "expected_outcome": "What you expect to happen..."
}}

OR if task is complete:
{{
    "reasoning": "Task is complete because...",
    "action": "FINISH",
    "final_answer": "The final answer to the task..."
}}

**Your Response**:
"""
        return prompt

    def _parse_thought(self, response: str) -> Dict[str, Any]:
        """
        Parse LLM thought response into structured format.
        """
        import json

        try:
            # Try to parse as JSON
            thought = json.loads(response)
            return thought
        except json.JSONDecodeError:
            # Fallback: extract reasoning and action manually
            # (In production, use more robust parsing)
            return {
                "reasoning": response,
                "action": None,
            }

    def _execute_action(self, action: Dict[str, Any]) -> Dict[str, Any]:
        """
        Execute tool action and capture result.
        """
        tool_name = action.get("tool")
        tool_args = action.get("args", {})

        if tool_name not in self.tools:
            return {
                "success": False,
                "result": f"Error: Tool '{tool_name}' not found",
                "error": "ToolNotFound",
            }

        try:
            tool = self.tools[tool_name]
            result = tool(**tool_args)

            return {
                "success": True,
                "result": result,
            }
        except Exception as e:
            self.logger.error(f"Tool execution error: {str(e)}", exc_info=True)
            return {
                "success": False,
                "result": str(e),
                "error": type(e).__name__,
            }

    def _reflect_on_action(
        self,
        thought: Dict,
        action: Dict,
        observation: Dict,
    ) -> Dict[str, Any]:
        """
        Reflect on whether action achieved intended goal.

        This is where self-correction happens.
        """
        reflection_prompt = f"""
Evaluate the result of this action:

**Intended Action**: {action['tool']}({action['args']})
**Expected Outcome**: {thought.get('expected_outcome')}
**Actual Observation**: {observation['result']}
**Success**: {observation['success']}

**Questions**:
1. Did the action succeed technically? (no errors)
2. Did the action achieve the intended goal?
3. Did we learn what we needed to learn?
4. Should we try a different approach?
5. Is the task now complete?

**Response Format** (JSON):
{{
    "success": true/false,
    "reasoning": "Detailed evaluation...",
    "task_complete": true/false,
    "suggested_next_action": "If action failed, suggest alternative...",
    "final_answer": "If task complete, provide final answer..."
}}

**Your Reflection**:
"""

        reflection_response = self.model.generate(reflection_prompt)

        try:
            import json
            reflection = json.loads(reflection_response)
            return reflection
        except json.JSONDecodeError:
            return {
                "success": observation["success"],
                "reasoning": reflection_response,
                "task_complete": False,
            }


# Example: Customer Support ReAct Agent

def create_support_react_agent(
    memory_service: 'VentrosMemoryService',
    session_manager: 'SessionManager',
    crm_api_base: str,
    crm_api_key: str,
) -> ReActAgent:
    """
    Create support agent using ReAct pattern.
    """

    # Define tools for support agent
    @FunctionTool
    def search_knowledge_base(query: str) -> Dict[str, Any]:
        """
        Search internal knowledge base for solutions.

        Args:
            query: Search query (error message, issue description)

        Returns:
            List of relevant articles/solutions
        """
        results = memory_service.search_memory(
            query=query,
            strategy_name="support_technical",
            limit=10,
        )
        return results

    @FunctionTool
    def search_past_tickets(contact_id: str, query: str) -> List[Dict]:
        """
        Search past support tickets for this contact.

        Args:
            contact_id: UUID of contact
            query: Issue description

        Returns:
            List of similar past tickets with resolutions
        """
        # Graph query: Contact -> HAS_TICKET -> Ticket
        tickets = memory_service.query_graph(
            node_id=contact_id,
            edge_type="HAS_TICKET",
            depth=1,
        )

        # Filter by relevance (simple keyword match)
        relevant = [
            t for t in tickets
            if query.lower() in t["target"].get("description", "").lower()
        ]

        return relevant

    @FunctionTool
    def get_product_version(contact_id: str) -> Dict[str, Any]:
        """
        Get customer's current product version and configuration.

        Args:
            contact_id: UUID of contact

        Returns:
            Product version, plan, and configuration details
        """
        import requests

        response = requests.get(
            f"{crm_api_base}/api/v1/contacts/{contact_id}/subscription",
            headers={"Authorization": f"Bearer {crm_api_key}"},
            timeout=10,
        )
        return response.json()

    @FunctionTool
    def check_service_status() -> Dict[str, Any]:
        """
        Check current status of services/APIs.

        Returns:
            System status, known incidents, planned maintenance
        """
        # This would call status page API
        return {
            "status": "operational",
            "incidents": [],
            "maintenance": [],
        }

    @FunctionTool
    def create_engineering_ticket(
        contact_id: str,
        title: str,
        description: str,
        severity: str,
    ) -> Dict[str, Any]:
        """
        Escalate issue to engineering team.

        Args:
            contact_id: UUID of contact
            title: Brief issue title
            description: Detailed description with reproduction steps
            severity: "low", "medium", "high", "critical"

        Returns:
            Created ticket with tracking number
        """
        import requests

        response = requests.post(
            f"{crm_api_base}/api/v1/engineering-tickets",
            json={
                "contact_id": contact_id,
                "title": title,
                "description": description,
                "severity": severity,
                "source": "ai_agent",
            },
            headers={"Authorization": f"Bearer {crm_api_key}"},
            timeout=10,
        )
        return response.json()

    @FunctionTool
    def send_solution_to_customer(
        contact_id: str,
        solution: str,
    ) -> Dict[str, Any]:
        """
        Send solution message to customer.

        Args:
            contact_id: UUID of contact
            solution: Solution steps (markdown supported)

        Returns:
            Sent message confirmation
        """
        import requests

        response = requests.post(
            f"{crm_api_base}/api/v1/messages",
            json={
                "contact_id": contact_id,
                "text": solution,
                "from_me": True,
                "source": "bot",
            },
            headers={"Authorization": f"Bearer {crm_api_key}"},
            timeout=15,
        )
        return response.json()

    # Create ReAct agent with these tools
    agent = ReActAgent(
        memory_service=memory_service,
        session_manager=session_manager,
        tools=[
            search_knowledge_base,
            search_past_tickets,
            get_product_version,
            check_service_status,
            create_engineering_ticket,
            send_solution_to_customer,
        ],
        max_iterations=10,
        verbose=True,
    )

    return agent


# Usage Example
"""
agent = create_support_react_agent(memory_service, session_manager, api_base, api_key)

result = agent.run({
    "task": "Customer reports 'Error 500' when trying to export contacts. Help diagnose and resolve.",
    "contact_id": "550e8400-e29b-41d4-a716-446655440000",
})

print(result["final_answer"])
print(f"Iterations: {result['iterations']}")
"""
```

### ReAct Example Output

```
============================================================
ITERATION 1/10
============================================================

💭 THOUGHT:
Customer is experiencing Error 500 during contact export. This is a server-side error.
First, I should check if this is a known system-wide issue or specific to this customer.
I'll start by checking service status.

🔧 ACTION: check_service_status
Args: {}

👁️ OBSERVATION:
{'status': 'operational', 'incidents': [], 'maintenance': []}

🤔 REFLECTION:
Success: True
Reasoning: Service status is operational, so this is likely a customer-specific issue.
Next, I should check their product version and configuration.

============================================================
ITERATION 2/10
============================================================

💭 THOUGHT:
No system-wide issues. Need to check customer's specific configuration.
Their product version might be outdated or they might have a configuration problem.

🔧 ACTION: get_product_version
Args: {'contact_id': '550e8400-e29b-41d4-a716-446655440000'}

👁️ OBSERVATION:
{'plan': 'Enterprise', 'version': '2.3.1', 'max_contacts': 50000, 'current_contacts': 48500}

🤔 REFLECTION:
Success: True
Reasoning: Customer is on Enterprise plan with 48,500 contacts (near limit).
This might be causing memory/timeout issues during export. Let me search knowledge base
for similar issues.

============================================================
ITERATION 3/10
============================================================

💭 THOUGHT:
Customer has nearly 50,000 contacts and is hitting Error 500 on export.
This is likely a timeout or memory issue. Let me search knowledge base for
"Error 500 export large dataset".

🔧 ACTION: search_knowledge_base
Args: {'query': 'Error 500 export large dataset timeout'}

👁️ OBSERVATION:
{
    'results': [
        {
            'title': 'Handling Large Exports',
            'solution': 'For datasets > 40K records, use batch export API instead of full export.
                        Max batch size: 10K records per request.',
            'score': 0.92
        }
    ]
}

🤔 REFLECTION:
Success: True
Reasoning: Found the root cause! Customer should use batch export API for large datasets.
I now have enough information to provide a solution.

✅ TASK COMPLETE
Final Answer: Error 500 is caused by attempting to export 48,500 contacts in a single request,
which exceeds the system's timeout limit. Solution: Use the batch export API with max 10K
records per request. Send customer instructions for batch export.
```

---

## Self-Reflection & Planning

### Pattern 1: Self-Critique and Refinement

```python
# ventros_adk/patterns/self_critique_agent.py

from google.adk import LlmAgent, SequentialAgent
from google.adk.models import GeminiModel

class SelfCritiqueAgent(SequentialAgent):
    """
    Generate-Critique-Refine pattern with self-reflection.

    Flow:
    1. Generator: Create initial output
    2. Critic: Evaluate output, identify weaknesses
    3. Refiner: Apply feedback and improve
    4. Validator: Final quality check
    """

    def __init__(
        self,
        memory_service: 'VentrosMemoryService',
        quality_threshold: float = 0.85,
        max_refinement_cycles: int = 3,
    ):
        # Generator Agent
        generator = LlmAgent(
            name="generator",
            model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.8),
            system_prompt="""
            You are a creative content generator.
            Generate engaging, personalized customer messages.
            Be creative and empathetic.
            """,
            memory=memory_service,
        )

        # Critic Agent (uses thinking model for deep analysis)
        critic = LlmAgent(
            name="critic",
            model=GeminiModel(
                model_name="gemini-2.0-flash-thinking-exp",
                temperature=0.2,
            ),
            system_prompt="""
            You are a harsh critic evaluating customer messages.

            Evaluate on:
            1. Personalization (0-25): Uses customer context appropriately
            2. Clarity (0-25): Clear and concise, no jargon
            3. Tone (0-25): Appropriate tone for situation
            4. Effectiveness (0-25): Strong call-to-action, persuasive

            Total score: 0-100

            Return:
            {
                "score": 85,
                "breakdown": {
                    "personalization": 22,
                    "clarity": 23,
                    "tone": 20,
                    "effectiveness": 20
                },
                "strengths": ["Good use of customer name", "Clear CTA"],
                "weaknesses": ["Too formal", "Lacks urgency"],
                "specific_improvements": [
                    "Use more conversational tone",
                    "Add deadline to create urgency"
                ]
            }
            """,
            memory=memory_service,
        )

        # Refiner Agent
        refiner = LlmAgent(
            name="refiner",
            model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.6),
            system_prompt="""
            You receive:
            - Original message
            - Critic feedback with specific improvements

            Apply ALL suggested improvements while maintaining core message.
            Be precise in implementing feedback.
            """,
            memory=memory_service,
        )

        # Validator Agent
        validator = LlmAgent(
            name="validator",
            model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.1),
            system_prompt="""
            Final quality gate. Ensure:
            - No grammatical errors
            - No placeholder text (like "[Customer Name]")
            - Appropriate length (not too long/short)
            - Professional presentation

            Return: {"approved": true/false, "issues": [...]}
            """,
            memory=memory_service,
        )

        super().__init__(
            name="self_critique_agent",
            agents=[generator, critic, refiner, validator],
        )

        self.quality_threshold = quality_threshold
        self.max_refinement_cycles = max_refinement_cycles

    def run(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Run generate-critique-refine loop until quality threshold met.
        """
        contact_id = input_data.get("contact_id")
        prompt = input_data.get("prompt")

        cycles = 0
        current_output = None

        while cycles < self.max_refinement_cycles:
            cycles += 1

            # Execute pipeline
            result = super().run({
                "contact_id": contact_id,
                "prompt": prompt,
                "previous_output": current_output,
                "previous_critique": result.get("critique") if cycles > 1 else None,
            })

            critique = result.get("critique")
            current_output = result.get("refined_output")
            validation = result.get("validation")

            # Check quality threshold
            if critique.get("score", 0) >= self.quality_threshold and validation.get("approved"):
                return {
                    "success": True,
                    "output": current_output,
                    "score": critique["score"],
                    "cycles": cycles,
                    "critique": critique,
                }

        # Max cycles reached
        return {
            "success": False,
            "output": current_output,
            "score": critique.get("score", 0),
            "cycles": cycles,
            "reason": "max_cycles_reached",
        }
```

### Pattern 2: Meta-Cognitive Planning

```python
# ventros_adk/patterns/planning_agent.py

from google.adk import LlmAgent
from google.adk.models import GeminiModel
from typing import Dict, Any, List

class PlanningAgent(LlmAgent):
    """
    Meta-cognitive agent that plans before acting.

    Process:
    1. Understand task and constraints
    2. Decompose into sub-goals
    3. Create step-by-step plan
    4. Identify potential risks/blockers
    5. Execute plan with monitoring
    """

    def __init__(self, memory_service: 'VentrosMemoryService'):
        super().__init__(
            name="planning_agent",
            model=GeminiModel(
                model_name="gemini-2.0-flash-thinking-exp-01-21",
                temperature=0.4,
                thinking_mode="thinking",
            ),
            system_prompt="""
            You are a strategic planning agent.

            Before taking any action, you must:
            1. UNDERSTAND: Analyze the task, goals, and constraints
            2. DECOMPOSE: Break down into achievable sub-goals
            3. PLAN: Create detailed step-by-step execution plan
            4. ANTICIPATE: Identify potential risks and mitigation strategies
            5. MONITOR: Track progress and adjust plan as needed

            Always think multiple steps ahead.
            Consider dependencies and order of operations.
            Plan for failure scenarios.
            """,
            memory=memory_service,
        )

    def create_plan(
        self,
        task: str,
        contact_id: str,
        constraints: Optional[Dict] = None,
    ) -> Dict[str, Any]:
        """
        Create comprehensive execution plan for task.
        """
        planning_prompt = f"""
Task: {task}
Contact ID: {contact_id}
Constraints: {constraints or 'None specified'}

Create a comprehensive execution plan following this structure:

1. GOAL ANALYSIS
   - Primary goal
   - Success criteria
   - Constraints and limitations

2. DECOMPOSITION
   - Break task into sub-goals (ordered)
   - Identify dependencies between sub-goals
   - Estimate time/effort for each

3. EXECUTION PLAN
   - Step-by-step actions
   - Tools/resources needed
   - Decision points and conditionals

4. RISK ASSESSMENT
   - Potential failure points
   - Mitigation strategies
   - Fallback plans

5. MONITORING STRATEGY
   - Progress indicators
   - Quality checkpoints
   - When to abort/pivot

Return as structured JSON.
"""

        response = self.model.generate(planning_prompt)

        import json
        try:
            plan = json.loads(response)
            return plan
        except json.JSONDecodeError:
            return {"error": "Failed to parse plan", "raw_response": response}

    def execute_plan_with_monitoring(
        self,
        plan: Dict[str, Any],
        executor_callback: callable,
    ) -> Dict[str, Any]:
        """
        Execute plan with real-time monitoring and adjustment.
        """
        steps = plan.get("execution_plan", {}).get("steps", [])
        results = []

        for i, step in enumerate(steps):
            # Execute step
            result = executor_callback(step)

            # Monitor result
            monitoring = self._monitor_step(step, result, plan)

            results.append({
                "step": i + 1,
                "action": step,
                "result": result,
                "monitoring": monitoring,
            })

            # Check if should continue
            if not monitoring.get("continue"):
                return {
                    "completed": False,
                    "reason": monitoring.get("reason"),
                    "results": results,
                }

        return {
            "completed": True,
            "results": results,
        }

    def _monitor_step(
        self,
        step: Dict,
        result: Any,
        overall_plan: Dict,
    ) -> Dict[str, Any]:
        """
        Monitor step execution and decide if plan should continue.
        """
        monitoring_prompt = f"""
Step Executed: {step}
Result: {result}
Overall Plan: {overall_plan}

Evaluate:
1. Did step succeed?
2. Is result as expected?
3. Should we continue to next step?
4. Do we need to adjust plan?

Return:
{{
    "success": true/false,
    "continue": true/false,
    "reason": "...",
    "plan_adjustment": "..."
}}
"""

        response = self.model.generate(monitoring_prompt)

        import json
        try:
            return json.loads(response)
        except json.JSONDecodeError:
            return {"success": True, "continue": True}
```

---

## Complete End-to-End Examples

### Example 1: Complete Customer Onboarding Flow

```python
# examples/customer_onboarding_flow.py

"""
Complete customer onboarding flow using multiple agent patterns.

Scenario:
- New customer signs up
- AI agent guides onboarding
- Multi-step process with personalization
- Human handoff if needed
"""

from ventros_adk.memory import VentrosMemoryService
from ventros_adk.session import SessionManager
from ventros_adk.agents import OnboardingAgent
from ventros_adk.orchestration import CoordinatorAgent
from google.adk import SequentialAgent, LlmAgent
from google.adk.models import GeminiModel

# Initialize services
memory_service = VentrosMemoryService(
    grpc_host="localhost",
    grpc_port=50051,
)

session_manager = SessionManager(
    memory_service=memory_service,
    session_timeout_minutes=30,
)

# Create onboarding pipeline
class CustomerOnboardingPipeline(SequentialAgent):
    """
    Complete onboarding flow:
    1. Welcome & Introduction
    2. Profile Completion
    3. Product Tour
    4. Integration Setup
    5. Success Confirmation
    """

    def __init__(self, memory_service, session_manager):
        # Step 1: Welcome
        welcome_agent = LlmAgent(
            name="welcome_agent",
            model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.8),
            system_prompt="""
            You are a friendly onboarding specialist.
            Welcome new customer warmly, introduce yourself and the platform.
            Set expectations for onboarding process.
            Ask about their primary goal.
            """,
            memory=memory_service,
        )

        # Step 2: Profile Completion
        profile_agent = LlmAgent(
            name="profile_agent",
            model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.6),
            system_prompt="""
            Guide customer through completing their profile.
            Collect: company size, industry, use case, team size.
            Be conversational, don't feel like a form.
            """,
            memory=memory_service,
        )

        # Step 3: Product Tour
        tour_agent = LlmAgent(
            name="tour_agent",
            model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.7),
            system_prompt="""
            Provide personalized product tour based on use case.
            Highlight features relevant to their industry/goal.
            Use screen sharing or interactive demo.
            """,
            memory=memory_service,
        )

        # Step 4: Integration Setup
        integration_agent = LlmAgent(
            name="integration_agent",
            model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.4),
            system_prompt="""
            Help customer connect integrations (email, CRM, calendar).
            Provide step-by-step technical instructions.
            Troubleshoot any connection issues.
            """,
            memory=memory_service,
        )

        # Step 5: Success Confirmation
        success_agent = LlmAgent(
            name="success_agent",
            model=GeminiModel(model_name="gemini-2.0-flash-exp", temperature=0.7),
            system_prompt="""
            Confirm successful onboarding.
            Set up first real task together.
            Provide next steps and resources.
            Schedule follow-up check-in.
            """,
            memory=memory_service,
        )

        super().__init__(
            name="customer_onboarding_pipeline",
            agents=[
                welcome_agent,
                profile_agent,
                tour_agent,
                integration_agent,
                success_agent,
            ],
        )

        self.memory_service = memory_service
        self.session_manager = session_manager

    def run(self, contact_id: str) -> Dict[str, Any]:
        """
        Execute full onboarding pipeline.
        """
        # Create session
        session = self.session_manager.get_or_create_session(contact_id)

        # Execute pipeline
        result = super().run({
            "contact_id": contact_id,
            "session_id": session.session_id,
        })

        # Close session with summary
        self.session_manager.close_session(
            session_id=session.session_id,
            summary="Customer onboarding completed successfully",
        )

        return result


# Usage
onboarding_pipeline = CustomerOnboardingPipeline(memory_service, session_manager)

result = onboarding_pipeline.run(contact_id="550e8400-e29b-41d4-a716-446655440000")

print(f"Onboarding Status: {result['status']}")
print(f"Steps Completed: {len(result['steps'])}")
```

### Example 2: Sales Qualification & Closing

```python
# examples/sales_qualification_flow.py

"""
Complete sales flow from lead to close.
"""

from ventros_adk.agents import (
    SalesProspectingAgent,
    SalesNegotiationAgent,
    SalesClosingAgent,
)
from ventros_adk.orchestration import HierarchicalTaskAgent

class SalesClosingOrchestrator(HierarchicalTaskAgent):
    """
    Hierarchical sales process:
    1. Lead Qualification (BANT)
    2. Needs Analysis
    3. Solution Presentation
    4. Objection Handling
    5. Proposal Generation
    6. Negotiation
    7. Close
    """

    def __init__(self, memory_service, session_manager):
        # Initialize specialized agents
        sub_agents = {
            "qualifier": SalesProspectingAgent(memory_service, session_manager),
            "needs_analyst": NeedsAnalysisAgent(memory_service, session_manager),
            "presenter": SolutionPresenterAgent(memory_service, session_manager),
            "objection_handler": ObjectionHandlerAgent(memory_service, session_manager),
            "proposal_generator": ProposalGeneratorAgent(memory_service, session_manager),
            "negotiator": SalesNegotiationAgent(memory_service, session_manager),
            "closer": SalesClosingAgent(memory_service, session_manager),
        }

        super().__init__(
            memory_service=memory_service,
            session_manager=session_manager,
            sub_agents=sub_agents,
        )

    def run(self, contact_id: str, opportunity_id: str) -> Dict[str, Any]:
        """
        Execute full sales cycle.
        """
        task = f"""
        Close sales opportunity {opportunity_id} for contact {contact_id}.

        Process:
        1. Qualify lead using BANT framework
        2. If qualified, analyze needs deeply
        3. Present tailored solution
        4. Handle any objections
        5. Generate proposal
        6. Negotiate terms if needed
        7. Ask for commitment (close)

        Use sub-agents appropriately for each step.
        If customer isn't ready, schedule follow-up.
        """

        result = super().run({
            "task": task,
            "contact_id": contact_id,
            "opportunity_id": opportunity_id,
        })

        return result


# Usage
sales_orchestrator = SalesClosingOrchestrator(memory_service, session_manager)

result = sales_orchestrator.run(
    contact_id="550e8400-e29b-41d4-a716-446655440000",
    opportunity_id="opp_12345",
)

if result.get("closed"):
    print(f"🎉 Deal closed! Value: ${result['deal_value']}")
else:
    print(f"Next step: {result['next_action']}")
```

---

## Production Deployment Guide

### Docker Deployment

```dockerfile
# Dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application
COPY . .

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Run application
CMD ["uvicorn", "ventros_adk.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  adk-service:
    build: .
    ports:
      - "8000:8000"
    environment:
      - GO_MEMORY_SERVICE_HOST=go-memory-service
      - GO_MEMORY_SERVICE_PORT=50051
      - GEMINI_API_KEY=${GEMINI_API_KEY}
      - CRM_API_BASE=http://go-api:8080
      - CRM_API_KEY=${CRM_API_KEY}
    depends_on:
      - go-memory-service
      - redis
    networks:
      - ventros-network

  go-memory-service:
    image: ventros/memory-service:latest
    ports:
      - "50051:50051"
    environment:
      - DATABASE_URL=postgresql://user:pass@postgres:5432/ventros
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis
    networks:
      - ventros-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    networks:
      - ventros-network

networks:
  ventros-network:
    driver: bridge
```

### Configuration Management

```python
# ventros_adk/config.py

from pydantic import BaseSettings
from typing import Optional

class Settings(BaseSettings):
    """
    Application settings (loads from environment variables).
    """

    # Go Memory Service
    go_memory_service_host: str = "localhost"
    go_memory_service_port: int = 50051

    # Gemini API
    gemini_api_key: str
    gemini_model: str = "gemini-2.0-flash-exp"

    # CRM API
    crm_api_base: str
    crm_api_key: str

    # Redis
    redis_url: str = "redis://localhost:6379"

    # Session Management
    session_timeout_minutes: int = 30
    context_window_tokens: int = 2_000_000

    # Logging
    log_level: str = "INFO"
    log_format: str = "json"  # "json" or "text"

    # Performance
    max_concurrent_agents: int = 10
    request_timeout_seconds: int = 120

    # Feature Flags
    enable_prompt_caching: bool = True
    enable_reranking: bool = True
    enable_thinking_mode: bool = True

    class Config:
        env_file = ".env"


settings = Settings()
```

### Application Entrypoint

```python
# ventros_adk/main.py

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Dict, Any
import logging

from ventros_adk.config import settings
from ventros_adk.memory import VentrosMemoryService
from ventros_adk.session import SessionManager
from ventros_adk.orchestration import CoordinatorAgent, create_production_coordinator

# Initialize logging
logging.basicConfig(
    level=settings.log_level,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Initialize FastAPI
app = FastAPI(
    title="Ventros ADK Service",
    description="AI Agent service for Ventros CRM",
    version="1.0.0",
)

# Initialize services (singleton)
memory_service = VentrosMemoryService(
    grpc_host=settings.go_memory_service_host,
    grpc_port=settings.go_memory_service_port,
    enable_caching=settings.enable_prompt_caching,
)

session_manager = SessionManager(
    memory_service=memory_service,
    context_window=settings.context_window_tokens,
    session_timeout_minutes=settings.session_timeout_minutes,
)

# Initialize coordinator with all agents
coordinator = create_production_coordinator(
    memory_service=memory_service,
    session_manager=session_manager,
)

# Request/Response Models
class MessageRequest(BaseModel):
    contact_id: str
    message: str
    channel_id: str = None
    session_id: str = None
    metadata: Dict[str, Any] = {}

class MessageResponse(BaseModel):
    success: bool
    response: str
    agent_used: str
    category: str
    confidence: float
    session_id: str

# API Endpoints
@app.post("/v1/messages", response_model=MessageResponse)
async def process_message(request: MessageRequest):
    """
    Process incoming customer message.
    """
    try:
        # Get or create session
        session = session_manager.get_or_create_session(
            contact_id=request.contact_id,
            channel_id=request.channel_id,
            metadata=request.metadata,
        )

        # Add user message to session
        session_manager.add_message_to_session(
            session_id=session.session_id,
            role="user",
            content=request.message,
        )

        # Process with coordinator
        result = coordinator.run({
            "contact_id": request.contact_id,
            "message": request.message,
            "session_id": session.session_id,
        })

        # Add assistant response to session
        session_manager.add_message_to_session(
            session_id=session.session_id,
            role="assistant",
            content=result.get("response", ""),
        )

        return MessageResponse(
            success=result.get("success", False),
            response=result.get("response", ""),
            agent_used=result.get("agent_used", "unknown"),
            category=result.get("category", "general"),
            confidence=result.get("confidence", 0.0),
            session_id=session.session_id,
        )

    except Exception as e:
        logger.error(f"Error processing message: {str(e)}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health_check():
    """Health check endpoint."""
    return {
        "status": "healthy",
        "memory_service": "connected" if memory_service.channel else "disconnected",
    }

@app.on_event("shutdown")
async def shutdown():
    """Cleanup on shutdown."""
    memory_service.close()
    logger.info("Application shutdown complete")
```

---

## Best Practices Summary

### 1. Memory & Context Management
✅ **Always include recent messages baseline** (SQL aggregation)
✅ **Use appropriate retrieval strategy** per agent category
✅ **Enable reranking for critical decisions** (churn, support)
✅ **Implement prompt caching** (5min TTL, 90% cost reduction)
✅ **Trim context proactively** to stay within window

### 2. Agent Design
✅ **Single responsibility** per agent
✅ **Clear system prompts** with examples
✅ **Appropriate temperature** (low for consistency, high for creativity)
✅ **Tool usage patterns** (ReAct for complex reasoning)
✅ **Error handling** with graceful fallbacks

### 3. Orchestration
✅ **Coordinator pattern** for routing
✅ **Hierarchical decomposition** for complex tasks
✅ **Parallel execution** when possible (latency reduction)
✅ **Human-in-the-loop** for high-stakes decisions
✅ **Consensus voting** for critical approvals

### 4. Production Readiness
✅ **Comprehensive logging** (structured JSON)
✅ **Health checks** and monitoring
✅ **Timeout handling** with retries
✅ **Rate limiting** (respect API quotas)
✅ **Graceful degradation** (fallback agents)

### 5. Performance
✅ **Async operations** where possible
✅ **Connection pooling** (gRPC, HTTP)
✅ **Caching layers** (Redis for prompts, results)
✅ **Batch operations** (group similar requests)
✅ **Monitoring latency** (P50, P95, P99)

---

## Architecture Diagram (Complete System)

```
┌──────────────────────────────────────────────────────────────────┐
│                         CUSTOMER                                  │
│                 (WhatsApp, Email, SMS, Web)                       │
└───────────────────────────────┬──────────────────────────────────┘
                                │
                                ↓
┌──────────────────────────────────────────────────────────────────┐
│                    GO CRM SERVICE (Main)                          │
│  • Message ingestion (WAHA webhook)                               │
│  • Contact/Session management                                     │
│  • Channel orchestration                                          │
│  • Event sourcing & outbox                                        │
└───────────────────────────────┬──────────────────────────────────┘
                                │ REST/Webhook
                                ↓
┌──────────────────────────────────────────────────────────────────┐
│                   PYTHON ADK SERVICE                              │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │            CoordinatorAgent (Router)                       │  │
│  │  • Semantic intent classification                          │  │
│  │  • Agent selection & dispatch                              │  │
│  │  • Fallback handling                                       │  │
│  └──────────────────┬─────────────────────────────────────────┘  │
│                     │                                             │
│  ┌──────────────────┴─────────────────────────────────────────┐  │
│  │         Specialized Agents (15+)                           │  │
│  │  • SalesProspectingAgent                                   │  │
│  │  • ChurnPreventionAgent                                    │  │
│  │  • TechnicalSupportAgent                                   │  │
│  │  • OnboardingAgent                                         │  │
│  │  • ... (sales, support, retention, ops, marketing)         │  │
│  └──────────────────┬─────────────────────────────────────────┘  │
│                     │                                             │
│  ┌──────────────────┴─────────────────────────────────────────┐  │
│  │           SessionManager                                   │  │
│  │  • Context assembly                                        │  │
│  │  • Conversation history tracking                           │  │
│  │  • Auto-summarization                                      │  │
│  └──────────────────┬─────────────────────────────────────────┘  │
│                     │ gRPC                                        │
└─────────────────────┼─────────────────────────────────────────────┘
                      ↓
┌──────────────────────────────────────────────────────────────────┐
│              GO MEMORY SERVICE (from Part 3)                      │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  HybridSearchService                                       │  │
│  │  • Vector search (pgvector HNSW)                           │  │
│  │  • Keyword search (pg_trgm + BM25)                         │  │
│  │  • Graph traversal (Apache AGE)                            │  │
│  │  • SQL aggregation (baseline)                              │  │
│  │  • RRF fusion + Jina reranking                             │  │
│  └────────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  TemporalKnowledgeGraphService                             │  │
│  │  • Bi-temporal edges (valid_from, valid_to)                │  │
│  │  • Point-in-time queries                                   │  │
│  │  • CYPHER graph queries                                    │  │
│  └────────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  SemanticRouterService                                     │  │
│  │  • Intent classification (zero-shot)                       │  │
│  │  • Route scoring & prioritization                          │  │
│  └────────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  MemoryFactService                                         │  │
│  │  • Structured fact extraction                              │  │
│  │  • Contradiction resolution                                │  │
│  │  • Temporal validity tracking                              │  │
│  └────────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  ContextManager                                            │  │
│  │  • Prompt caching (Redis 5min TTL)                         │  │
│  │  • Context assembly optimization                           │  │
│  └────────────────────────────────────────────────────────────┘  │
└────────────────────────────┬─────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────┐
│         PostgreSQL + pgvector + Apache AGE + pg_trgm             │
│  • messages (recent baseline - ALWAYS included)                  │
│  • sessions (summaries with contextual embeddings)               │
│  • memory_embeddings (HNSW index, 768-dim)                       │
│  • memory_facts (temporal validity)                              │
│  • temporal_edges (bi-temporal graph)                            │
│  • agent_ai_metadata (routing rules, skills)                     │
│  • semantic_routes (intent classification)                       │
└──────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│                      EXTERNAL SERVICES                            │
│  • Vertex AI (text-embedding-005, Gemini 2.0 Flash)              │
│  • Jina Reranker v2 (multilingual, function-calling)             │
│  • Redis (prompt caching, session state)                         │
│  • RabbitMQ (event bus, async processing)                        │
└──────────────────────────────────────────────────────────────────┘
```

---

## Conclusion

This comprehensive Python ADK architecture provides:

✅ **All ADK agent types** (LlmAgent, SequentialAgent, ParallelAgent, LoopAgent, Custom BaseAgent)
✅ **Multi-agent orchestration patterns** (Coordinator, Hierarchical, Consensus, HITL, Parallel)
✅ **Advanced reasoning** (ReAct, self-reflection, planning)
✅ **Production-ready** (Docker, FastAPI, health checks, monitoring)
✅ **Seamless Go integration** (gRPC to Memory Service)
✅ **Sophisticated session management** (context assembly, auto-summarization)
✅ **Tool ecosystem** (Function Tools, Agent-as-Tool, CRM operations)
✅ **Performance optimized** (async operations, caching, parallel execution)

The architecture leverages **2025 state-of-the-art patterns**:
- **OmniRAG** (dynamic retrieval method selection)
- **Contextual Retrieval** (Anthropic 2025, 67% error reduction)
- **Temporal Knowledge Graphs** (Zep/Graphiti bi-temporal model)
- **Semantic Routing** (Aurelio Labs zero-shot classification)
- **Memory Bank** (Google contradiction resolution)
- **Prompt Caching** (90% cost reduction)
- **Gemini 2.0 Flash Thinking** (deep reasoning mode)

This system can handle **dozens to hundreds of specialized agents** with sophisticated routing, memory, and orchestration - exactly as requested for Ventros CRM's vision of rich AI-powered customer interactions.
