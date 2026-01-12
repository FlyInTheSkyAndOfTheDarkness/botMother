package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/agent"
	agentRepo "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/agent"
	aiService "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/ai"
)

type AgentService struct {
	repo *agentRepo.SQLiteRepository
}

func NewAgentService(repo *agentRepo.SQLiteRepository) *AgentService {
	return &AgentService{repo: repo}
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func (s *AgentService) CreateAgent(ctx context.Context, req agent.CreateAgentRequest) (*agent.AgentResponse, error) {
	a := &agent.Agent{
		Name:           req.Name,
		Description:    req.Description,
		APIKey:         req.APIKey,
		Model:          req.Model,
		SystemPrompt:   req.SystemPrompt,
		WelcomeMessage: req.WelcomeMessage,
		IsActive:       true,
	}

	if err := s.repo.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return s.agentToResponse(ctx, a)
}

func (s *AgentService) GetAgent(ctx context.Context, id string) (*agent.AgentResponse, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.agentToResponse(ctx, a)
}

// GetAgentInternal returns the full agent with API key (for internal use only)
func (s *AgentService) GetAgentInternal(ctx context.Context, id string) (*agent.Agent, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AgentService) GetAllAgents(ctx context.Context) ([]*agent.AgentResponse, error) {
	agents, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var responses []*agent.AgentResponse
	for _, a := range agents {
		resp, err := s.agentToResponse(ctx, a)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

func (s *AgentService) UpdateAgent(ctx context.Context, id string, req agent.UpdateAgentRequest) (*agent.AgentResponse, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update only provided fields
	if req.Name != nil {
		a.Name = *req.Name
	}
	if req.Description != nil {
		a.Description = *req.Description
	}
	if req.APIKey != nil {
		a.APIKey = *req.APIKey
	}
	if req.Model != nil {
		a.Model = *req.Model
	}
	if req.SystemPrompt != nil {
		a.SystemPrompt = *req.SystemPrompt
	}
	if req.WelcomeMessage != nil {
		a.WelcomeMessage = *req.WelcomeMessage
	}
	if req.IsActive != nil {
		a.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	return s.agentToResponse(ctx, a)
}

func (s *AgentService) DeleteAgent(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *AgentService) agentToResponse(ctx context.Context, a *agent.Agent) (*agent.AgentResponse, error) {
	integrations, err := s.repo.GetIntegrationsByAgentID(ctx, a.ID)
	if err != nil {
		return nil, err
	}

	var integrationResponses []agent.IntegrationResponse
	for _, i := range integrations {
		status := "disconnected"
		if i.IsConnected {
			status = "connected"
		}
		integrationResponses = append(integrationResponses, agent.IntegrationResponse{
			ID:          i.ID,
			Type:        i.Type,
			IsConnected: i.IsConnected,
			Status:      status,
		})
	}

	return &agent.AgentResponse{
		ID:             a.ID,
		Name:           a.Name,
		Description:    a.Description,
		APIKeyMasked:   maskAPIKey(a.APIKey),
		Model:          a.Model,
		SystemPrompt:   a.SystemPrompt,
		WelcomeMessage: a.WelcomeMessage,
		IsActive:       a.IsActive,
		Integrations:   integrationResponses,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}, nil
}

// HandleIncomingMessage processes an incoming message and generates AI response
func (s *AgentService) HandleIncomingMessage(ctx context.Context, agentID, integrationID, remoteJID, userMessage string) (string, error) {
	// Get agent
	a, err := s.repo.GetByID(ctx, agentID)
	if err != nil {
		return "", fmt.Errorf("agent not found: %w", err)
	}

	if !a.IsActive {
		return "", fmt.Errorf("agent is not active")
	}

	// Get or create conversation
	conv, err := s.repo.GetOrCreateConversation(ctx, agentID, integrationID, remoteJID)
	if err != nil {
		return "", fmt.Errorf("failed to get conversation: %w", err)
	}

	// Check if in manual mode (manager took over)
	if conv.IsManualMode {
		// Store user message but don't generate AI response
		userMsg := &agent.Message{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        userMessage,
		}
		s.repo.AddMessage(ctx, userMsg)
		return "", nil // Return empty - no AI response in manual mode
	}

	// Store user message
	userMsg := &agent.Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        userMessage,
	}
	if err := s.repo.AddMessage(ctx, userMsg); err != nil {
		return "", fmt.Errorf("failed to store user message: %w", err)
	}

	var response string

	// Check if this is the first reply - send welcome message
	if !conv.IsFirstReply && a.WelcomeMessage != "" {
		response = a.WelcomeMessage
		conv.IsFirstReply = true
		if err := s.repo.UpdateConversation(ctx, conv); err != nil {
			return "", fmt.Errorf("failed to update conversation: %w", err)
		}
	} else {
		// Generate AI response
		aiSvc := aiService.NewService(a.APIKey)
		if aiSvc == nil {
			return "", fmt.Errorf("failed to initialize AI service")
		}

		// Get recent messages for context
		recentMessages, err := s.repo.GetRecentMessages(ctx, conv.ID, 10)
		if err != nil {
			return "", fmt.Errorf("failed to get recent messages: %w", err)
		}

		// Build context from recent messages
		var contextBuilder strings.Builder
		for _, msg := range recentMessages {
			if msg.Role == "user" {
				contextBuilder.WriteString(fmt.Sprintf("User: %s\n", msg.Content))
			} else if msg.Role == "assistant" {
				contextBuilder.WriteString(fmt.Sprintf("Assistant: %s\n", msg.Content))
			}
		}

		// If we have context, include it in the prompt
		finalPrompt := userMessage
		if contextBuilder.Len() > 0 {
			finalPrompt = fmt.Sprintf("Previous conversation:\n%s\nCurrent message: %s", contextBuilder.String(), userMessage)
		}

		response, err = aiSvc.GenerateResponse(ctx, finalPrompt, a.SystemPrompt, a.Model)
		if err != nil {
			return "", fmt.Errorf("failed to generate AI response: %w", err)
		}

		// Mark conversation as having had first reply
		if !conv.IsFirstReply {
			conv.IsFirstReply = true
			s.repo.UpdateConversation(ctx, conv)
		}
	}

	// Store assistant response
	assistantMsg := &agent.Message{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        response,
	}
	if err := s.repo.AddMessage(ctx, assistantMsg); err != nil {
		return "", fmt.Errorf("failed to store assistant message: %w", err)
	}

	return response, nil
}

// Integration management methods

func (s *AgentService) GetOrCreateIntegration(ctx context.Context, agentID, integrationType string) (*agent.Integration, error) {
	// Check if integration already exists
	existing, err := s.repo.GetIntegrationByAgentAndType(ctx, agentID, integrationType)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// Create new integration
	integration := &agent.Integration{
		AgentID:     agentID,
		Type:        integrationType,
		IsConnected: false,
		Config:      "{}",
	}
	if err := s.repo.CreateIntegration(ctx, integration); err != nil {
		return nil, err
	}
	return integration, nil
}

func (s *AgentService) UpdateIntegrationStatus(ctx context.Context, integrationID string, isConnected bool, config string) error {
	integration, err := s.repo.GetIntegrationByID(ctx, integrationID)
	if err != nil {
		return err
	}
	integration.IsConnected = isConnected
	if config != "" {
		integration.Config = config
	}
	return s.repo.UpdateIntegration(ctx, integration)
}

func (s *AgentService) GetIntegration(ctx context.Context, integrationID string) (*agent.Integration, error) {
	return s.repo.GetIntegrationByID(ctx, integrationID)
}

func (s *AgentService) DeleteIntegration(ctx context.Context, integrationID string) error {
	return s.repo.DeleteIntegration(ctx, integrationID)
}

// UpdateIntegrationConfig updates integration config and connection status
func (s *AgentService) UpdateIntegrationConfig(ctx context.Context, integrationID, config string, isConnected bool) error {
	integration, err := s.repo.GetIntegrationByID(ctx, integrationID)
	if err != nil {
		return err
	}
	integration.Config = config
	integration.IsConnected = isConnected
	return s.repo.UpdateIntegration(ctx, integration)
}

// Conversation methods for Live Chat

// GetAllConversations returns all conversations
func (s *AgentService) GetAllConversations(ctx context.Context) ([]*agent.Conversation, error) {
	return s.repo.GetAllConversations(ctx)
}

// GetConversation returns a single conversation by ID
func (s *AgentService) GetConversation(ctx context.Context, id string) (*agent.Conversation, error) {
	return s.repo.GetConversationByID(ctx, id)
}

// GetMessagesForConversation returns all messages for a conversation
func (s *AgentService) GetMessagesForConversation(ctx context.Context, conversationID string) ([]*agent.Message, error) {
	return s.repo.GetMessagesForConversation(ctx, conversationID)
}

// GetLastMessage returns the last message for a conversation
func (s *AgentService) GetLastMessage(ctx context.Context, conversationID string) (*agent.Message, error) {
	return s.repo.GetLastMessageForConversation(ctx, conversationID)
}

// SetConversationManualMode sets manual mode for a conversation (TakeOver/Release)
func (s *AgentService) SetConversationManualMode(ctx context.Context, conversationID string, isManual bool) error {
	return s.repo.SetConversationManualMode(ctx, conversationID, isManual)
}

// UpdateConversationNotes updates notes for a conversation
func (s *AgentService) UpdateConversationNotes(ctx context.Context, conversationID, notes string) error {
	return s.repo.UpdateConversationNotes(ctx, conversationID, notes)
}

// AddManualMessage adds a message sent by manager (not AI)
func (s *AgentService) AddManualMessage(ctx context.Context, conversationID, content string) (*agent.Message, error) {
	msg := &agent.Message{
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        content,
	}
	if err := s.repo.AddMessage(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

