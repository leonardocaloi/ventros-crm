package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ventros/crm/internal/domain/crm/agent"
)

func TestGetAgentUseCase_NewGetAgentUseCase(t *testing.T) {
	// Arrange
	agentRepo := new(MockAgentRepository)

	// Act
	useCase := NewGetAgentUseCase(agentRepo)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, agentRepo, useCase.agentRepo)
}

func TestGetAgentUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	agentRepo := new(MockAgentRepository)
	useCase := NewGetAgentUseCase(agentRepo)

	agentID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()
	tenantID := "tenant-1"
	name := "John Agent"
	email := "john@example.com"

	// Create a mock agent
	foundAgent := agent.ReconstructAgent(
		agentID,
		1, // version
		projectID,
		&userID,
		tenantID,
		name,
		email,
		agent.AgentTypeHuman,
		agent.AgentStatusAvailable,
		agent.RoleHumanAgent,
		true,
		make(map[string]interface{}),
		make(map[string]bool),
		make(map[string]interface{}),
		0,
		0,
		nil,
		nil,
		testTime(),
		testTime(),
		nil,
	)

	req := GetAgentRequest{
		AgentID: agentID,
	}

	// Mock: FindByID
	agentRepo.On("FindByID", ctx, agentID).Return(foundAgent, nil)

	// Act
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, agentID, result.AgentID)
	assert.Equal(t, tenantID, result.TenantID)
	assert.Equal(t, name, result.Name)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, agent.AgentTypeHuman, result.AgentType)
	assert.Equal(t, agent.AgentStatusAvailable, result.Status)
	assert.True(t, result.IsActive)
	assert.NotEmpty(t, result.CreatedAt)
	assert.NotEmpty(t, result.UpdatedAt)

	agentRepo.AssertExpectations(t)
}

func TestGetAgentUseCase_Execute_AgentNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	agentRepo := new(MockAgentRepository)
	useCase := NewGetAgentUseCase(agentRepo)

	agentID := uuid.New()

	req := GetAgentRequest{
		AgentID: agentID,
	}

	// Mock: FindByID returns not found
	agentRepo.On("FindByID", ctx, agentID).Return(nil, agent.ErrAgentNotFound)

	// Act
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "agent not found")

	agentRepo.AssertExpectations(t)
}

func TestGetAgentUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	agentRepo := new(MockAgentRepository)
	useCase := NewGetAgentUseCase(agentRepo)

	agentID := uuid.New()

	req := GetAgentRequest{
		AgentID: agentID,
	}

	// Mock: FindByID returns generic error
	dbError := errors.New("database connection error")
	agentRepo.On("FindByID", ctx, agentID).Return(nil, dbError)

	// Act
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find agent")

	agentRepo.AssertExpectations(t)
}

func TestGetAgentUseCase_Execute_InactiveAgent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	agentRepo := new(MockAgentRepository)
	useCase := NewGetAgentUseCase(agentRepo)

	agentID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()
	tenantID := "tenant-1"
	name := "Inactive Agent"
	email := "inactive@example.com"

	// Create an inactive agent
	foundAgent := agent.ReconstructAgent(
		agentID,
		1, // version
		projectID,
		&userID,
		tenantID,
		name,
		email,
		agent.AgentTypeHuman,
		agent.AgentStatusOffline,
		agent.RoleHumanAgent,
		false, // inactive
		make(map[string]interface{}),
		make(map[string]bool),
		make(map[string]interface{}),
		0,
		0,
		nil,
		nil,
		testTime(),
		testTime(),
		nil,
	)

	req := GetAgentRequest{
		AgentID: agentID,
	}

	// Mock: FindByID
	agentRepo.On("FindByID", ctx, agentID).Return(foundAgent, nil)

	// Act
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, agentID, result.AgentID)
	assert.False(t, result.IsActive)

	agentRepo.AssertExpectations(t)
}

func TestGetAgentUseCase_Execute_AIAgent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	agentRepo := new(MockAgentRepository)
	useCase := NewGetAgentUseCase(agentRepo)

	agentID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "AI Assistant"
	email := "ai@example.com"

	// Create an AI agent
	foundAgent := agent.ReconstructAgent(
		agentID,
		1, // version
		projectID,
		nil, // AI agents don't have userID
		tenantID,
		name,
		email,
		agent.AgentTypeAI,
		agent.AgentStatusAvailable,
		agent.RoleAIAgent,
		true,
		make(map[string]interface{}),
		make(map[string]bool),
		make(map[string]interface{}),
		0,
		0,
		nil,
		nil,
		testTime(),
		testTime(),
		nil,
	)

	req := GetAgentRequest{
		AgentID: agentID,
	}

	// Mock: FindByID
	agentRepo.On("FindByID", ctx, agentID).Return(foundAgent, nil)

	// Act
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, agentID, result.AgentID)
	assert.Equal(t, agent.AgentTypeAI, result.AgentType)
	assert.Equal(t, agent.AgentStatusAvailable, result.Status) // Status field holds the agent status
	assert.True(t, result.IsActive)

	agentRepo.AssertExpectations(t)
}

func TestListAgentsUseCase_NewListAgentsUseCase(t *testing.T) {
	// Arrange
	agentRepo := new(MockAgentRepository)

	// Act
	useCase := NewListAgentsUseCase(agentRepo)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, agentRepo, useCase.agentRepo)
}

func TestListAgentsUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	agentRepo := new(MockAgentRepository)
	useCase := NewListAgentsUseCase(agentRepo)

	tenantID := "tenant-1"
	projectID := uuid.New()

	// Create mock agents
	userID1 := uuid.New()
	agent1 := agent.ReconstructAgent(
		uuid.New(),
		1, // version
		projectID,
		&userID1,
		tenantID,
		"Agent 1",
		"agent1@example.com",
		agent.AgentTypeHuman,
		agent.AgentStatusAvailable,
		agent.RoleHumanAgent,
		true,
		make(map[string]interface{}),
		make(map[string]bool),
		make(map[string]interface{}),
		0,
		0,
		nil,
		nil,
		testTime(),
		testTime(),
		nil,
	)

	userID2 := uuid.New()
	agent2 := agent.ReconstructAgent(
		uuid.New(),
		1, // version
		projectID,
		&userID2,
		tenantID,
		"Agent 2",
		"agent2@example.com",
		agent.AgentTypeHuman,
		agent.AgentStatusBusy,
		agent.RoleHumanAgent,
		true,
		make(map[string]interface{}),
		make(map[string]bool),
		make(map[string]interface{}),
		0,
		0,
		nil,
		nil,
		testTime(),
		testTime(),
		nil,
	)

	agents := []*agent.Agent{agent1, agent2}

	req := ListAgentsRequest{
		TenantID:   tenantID,
		ActiveOnly: false,
		Limit:      20,
		Offset:     0,
	}

	// Mock: FindByTenant
	agentRepo.On("FindByTenant", ctx, tenantID).Return(agents, nil)

	// Act
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.Agents))
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 20, result.Limit)
	assert.Equal(t, 0, result.Offset)

	agentRepo.AssertExpectations(t)
}

func TestListAgentsUseCase_Execute_ActiveOnly(t *testing.T) {
	// Arrange
	ctx := context.Background()
	agentRepo := new(MockAgentRepository)
	useCase := NewListAgentsUseCase(agentRepo)

	tenantID := "tenant-1"
	projectID := uuid.New()

	// Create mock active agents
	userID1 := uuid.New()
	agent1 := agent.ReconstructAgent(
		uuid.New(),
		1, // version
		projectID,
		&userID1,
		tenantID,
		"Active Agent 1",
		"active1@example.com",
		agent.AgentTypeHuman,
		agent.AgentStatusAvailable,
		agent.RoleHumanAgent,
		true,
		make(map[string]interface{}),
		make(map[string]bool),
		make(map[string]interface{}),
		0,
		0,
		nil,
		nil,
		testTime(),
		testTime(),
		nil,
	)

	agents := []*agent.Agent{agent1}

	req := ListAgentsRequest{
		TenantID:   tenantID,
		ActiveOnly: true,
		Limit:      20,
		Offset:     0,
	}

	// Mock: FindActiveByTenant
	agentRepo.On("FindActiveByTenant", ctx, tenantID).Return(agents, nil)

	// Act
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result.Agents))
	assert.Equal(t, 1, result.Total)

	agentRepo.AssertExpectations(t)
	agentRepo.AssertNotCalled(t, "FindByTenant")
}

func TestListAgentsUseCase_Execute_EmptyResult(t *testing.T) {
	// Arrange
	ctx := context.Background()
	agentRepo := new(MockAgentRepository)
	useCase := NewListAgentsUseCase(agentRepo)

	tenantID := "tenant-1"

	req := ListAgentsRequest{
		TenantID:   tenantID,
		ActiveOnly: false,
		Limit:      20,
		Offset:     0,
	}

	// Mock: FindByTenant returns empty slice
	agents := []*agent.Agent{}
	agentRepo.On("FindByTenant", ctx, tenantID).Return(agents, nil)

	// Act
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.Agents))
	assert.Equal(t, 0, result.Total)

	agentRepo.AssertExpectations(t)
}

func TestListAgentsUseCase_Execute_DefaultLimit(t *testing.T) {
	// Arrange
	ctx := context.Background()
	agentRepo := new(MockAgentRepository)
	useCase := NewListAgentsUseCase(agentRepo)

	tenantID := "tenant-1"

	req := ListAgentsRequest{
		TenantID:   tenantID,
		ActiveOnly: false,
		Limit:      0, // Should default to 20
		Offset:     0,
	}

	// Mock: FindByTenant
	agents := []*agent.Agent{}
	agentRepo.On("FindByTenant", ctx, tenantID).Return(agents, nil)

	// Act
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 20, result.Limit) // Should be default 20

	agentRepo.AssertExpectations(t)
}

func TestListAgentsUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	agentRepo := new(MockAgentRepository)
	useCase := NewListAgentsUseCase(agentRepo)

	tenantID := "tenant-1"

	req := ListAgentsRequest{
		TenantID:   tenantID,
		ActiveOnly: false,
		Limit:      20,
		Offset:     0,
	}

	// Mock: FindByTenant returns error
	dbError := errors.New("database connection error")
	agentRepo.On("FindByTenant", ctx, tenantID).Return(nil, dbError)

	// Act
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find agents")

	agentRepo.AssertExpectations(t)
}

func TestListAgentsUseCase_Execute_ActiveOnlyRepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	agentRepo := new(MockAgentRepository)
	useCase := NewListAgentsUseCase(agentRepo)

	tenantID := "tenant-1"

	req := ListAgentsRequest{
		TenantID:   tenantID,
		ActiveOnly: true,
		Limit:      20,
		Offset:     0,
	}

	// Mock: FindActiveByTenant returns error
	dbError := errors.New("database connection error")
	agentRepo.On("FindActiveByTenant", ctx, tenantID).Return(nil, dbError)

	// Act
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find agents")

	agentRepo.AssertExpectations(t)
}

func TestListAgentsUseCase_Execute_WithPagination(t *testing.T) {
	// Arrange
	ctx := context.Background()
	agentRepo := new(MockAgentRepository)
	useCase := NewListAgentsUseCase(agentRepo)

	tenantID := "tenant-1"
	projectID := uuid.New()

	// Create mock agents
	userID1 := uuid.New()
	agent1 := agent.ReconstructAgent(
		uuid.New(),
		1, // version
		projectID,
		&userID1,
		tenantID,
		"Agent 1",
		"agent1@example.com",
		agent.AgentTypeHuman,
		agent.AgentStatusAvailable,
		agent.RoleHumanAgent,
		true,
		make(map[string]interface{}),
		make(map[string]bool),
		make(map[string]interface{}),
		0,
		0,
		nil,
		nil,
		testTime(),
		testTime(),
		nil,
	)

	agents := []*agent.Agent{agent1}

	req := ListAgentsRequest{
		TenantID:   tenantID,
		ActiveOnly: false,
		Limit:      10,
		Offset:     20,
	}

	// Mock: FindByTenant
	agentRepo.On("FindByTenant", ctx, tenantID).Return(agents, nil)

	// Act
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 10, result.Limit)
	assert.Equal(t, 20, result.Offset)

	agentRepo.AssertExpectations(t)
}

func TestListAgentsUseCase_Execute_MixedAgentTypes(t *testing.T) {
	// Arrange
	ctx := context.Background()
	agentRepo := new(MockAgentRepository)
	useCase := NewListAgentsUseCase(agentRepo)

	tenantID := "tenant-1"
	projectID := uuid.New()

	// Create mixed agent types
	userID := uuid.New()
	humanAgent := agent.ReconstructAgent(
		uuid.New(),
		1, // version
		projectID,
		&userID,
		tenantID,
		"Human Agent",
		"human@example.com",
		agent.AgentTypeHuman,
		agent.AgentStatusAvailable,
		agent.RoleHumanAgent,
		true,
		make(map[string]interface{}),
		make(map[string]bool),
		make(map[string]interface{}),
		0,
		0,
		nil,
		nil,
		testTime(),
		testTime(),
		nil,
	)

	aiAgent := agent.ReconstructAgent(
		uuid.New(),
		1, // version
		projectID,
		nil,
		tenantID,
		"AI Agent",
		"ai@example.com",
		agent.AgentTypeAI,
		agent.AgentStatusAvailable,
		agent.RoleAIAgent,
		true,
		make(map[string]interface{}),
		make(map[string]bool),
		make(map[string]interface{}),
		0,
		0,
		nil,
		nil,
		testTime(),
		testTime(),
		nil,
	)

	botAgent := agent.ReconstructAgent(
		uuid.New(),
		1, // version
		projectID,
		nil,
		tenantID,
		"Bot Agent",
		"bot@example.com",
		agent.AgentTypeBot,
		agent.AgentStatusAvailable,
		agent.RoleChannelBot,
		true,
		make(map[string]interface{}),
		make(map[string]bool),
		make(map[string]interface{}),
		0,
		0,
		nil,
		nil,
		testTime(),
		testTime(),
		nil,
	)

	agents := []*agent.Agent{humanAgent, aiAgent, botAgent}

	req := ListAgentsRequest{
		TenantID:   tenantID,
		ActiveOnly: false,
		Limit:      20,
		Offset:     0,
	}

	// Mock: FindByTenant
	agentRepo.On("FindByTenant", ctx, tenantID).Return(agents, nil)

	// Act
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, len(result.Agents))
	assert.Equal(t, 3, result.Total)

	// Verify agent types
	assert.Equal(t, agent.AgentTypeHuman, result.Agents[0].AgentType)
	assert.Equal(t, agent.AgentTypeAI, result.Agents[1].AgentType)
	assert.Equal(t, agent.AgentTypeBot, result.Agents[2].AgentType)

	agentRepo.AssertExpectations(t)
}
