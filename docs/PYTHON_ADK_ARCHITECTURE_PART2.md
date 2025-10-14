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
