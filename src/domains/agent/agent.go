package agent

import (
	"context"
	"time"
)

// Agent represents an AI chatbot agent
type Agent struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	APIKey         string    `json:"api_key"`          // OpenAI API key
	SerpAPIKey     string    `json:"serp_api_key"`    // SerpAPI key for internet access
	Model          string    `json:"model"`            // gpt-4o-mini, gpt-4o, etc.
	SystemPrompt   string    `json:"system_prompt"`    // AI behavior instructions
	WelcomeMessage string    `json:"welcome_message"`  // First response template
	IsActive       bool      `json:"is_active"`        // Enable/disable agent
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Integration types
const (
	IntegrationTypeWhatsApp  = "whatsapp"
	IntegrationTypeTelegram  = "telegram"
	IntegrationTypeInstagram = "instagram"
)

// Integration represents a messaging platform integration for an agent
type Integration struct {
	ID          string    `json:"id"`
	AgentID     string    `json:"agent_id"`
	Type        string    `json:"type"`        // whatsapp, telegram, instagram
	IsConnected bool      `json:"is_connected"`
	Config      string    `json:"config"`      // JSON config specific to integration type
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WhatsAppConfig holds WhatsApp-specific integration settings
type WhatsAppConfig struct {
	PhoneNumber string `json:"phone_number,omitempty"`
	DeviceID    string `json:"device_id,omitempty"`
	// Session data stored separately for security
}

// TelegramConfig holds Telegram-specific integration settings
type TelegramConfig struct {
	BotToken    string `json:"bot_token"`
	BotUsername string `json:"bot_username,omitempty"`
}

// InstagramConfig holds Instagram-specific integration settings
type InstagramConfig struct {
	AccessToken string `json:"access_token"`
	PageID      string `json:"page_id"`
	Username    string `json:"username,omitempty"`
}

// Conversation tracks message history for context
type Conversation struct {
	ID            string    `json:"id"`
	AgentID       string    `json:"agent_id"`
	IntegrationID string    `json:"integration_id"`
	RemoteJID     string    `json:"remote_jid"`      // User identifier (phone, chat_id, etc.)
	IsFirstReply  bool      `json:"is_first_reply"`  // Whether welcome message was sent
	IsManualMode  bool      `json:"is_manual_mode"`  // Whether AI is paused (manager takeover)
	Notes         string    `json:"notes"`           // Manager notes
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Message represents a single message in a conversation
type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           string    `json:"role"`    // user, assistant, system
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp"`
}

// CreateAgentRequest is the request body for creating an agent
type CreateAgentRequest struct {
	Name           string `json:"name" validate:"required,min=1,max=100"`
	Description    string `json:"description,omitempty"`
	APIKey         string `json:"api_key" validate:"required"`
	SerpAPIKey     string `json:"serp_api_key,omitempty"` // Optional SerpAPI key
	Model          string `json:"model" validate:"required"`
	SystemPrompt   string `json:"system_prompt" validate:"required"`
	WelcomeMessage string `json:"welcome_message,omitempty"`
}

// UpdateAgentRequest is the request body for updating an agent
type UpdateAgentRequest struct {
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	APIKey         *string `json:"api_key,omitempty"`
	SerpAPIKey     *string `json:"serp_api_key,omitempty"`
	Model          *string `json:"model,omitempty"`
	SystemPrompt   *string `json:"system_prompt,omitempty"`
	WelcomeMessage *string `json:"welcome_message,omitempty"`
	IsActive       *bool   `json:"is_active,omitempty"`
}

// AgentResponse is the response with masked sensitive data
type AgentResponse struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	APIKeyMasked   string                `json:"api_key_masked"` // sk-...xxxx
	SerpAPIKeyMasked string              `json:"serp_api_key_masked,omitempty"` // SerpAPI key masked
	Model          string                `json:"model"`
	SystemPrompt   string                `json:"system_prompt"`
	WelcomeMessage string                `json:"welcome_message"`
	IsActive       bool                  `json:"is_active"`
	Integrations   []IntegrationResponse `json:"integrations"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

// IntegrationResponse is the response for an integration
type IntegrationResponse struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	IsConnected bool   `json:"is_connected"`
	Status      string `json:"status"` // connected, disconnected, pending
	Details     string `json:"details,omitempty"`
}

// IAgentRepository defines database operations for agents
type IAgentRepository interface {
	// Agent CRUD
	Create(ctx context.Context, agent *Agent) error
	GetByID(ctx context.Context, id string) (*Agent, error)
	GetAll(ctx context.Context) ([]*Agent, error)
	Update(ctx context.Context, agent *Agent) error
	Delete(ctx context.Context, id string) error

	// Integration CRUD
	CreateIntegration(ctx context.Context, integration *Integration) error
	GetIntegrationsByAgentID(ctx context.Context, agentID string) ([]*Integration, error)
	GetIntegrationByID(ctx context.Context, id string) (*Integration, error)
	UpdateIntegration(ctx context.Context, integration *Integration) error
	DeleteIntegration(ctx context.Context, id string) error

	// Conversation management
	GetOrCreateConversation(ctx context.Context, agentID, integrationID, remoteJID string) (*Conversation, error)
	UpdateConversation(ctx context.Context, conversation *Conversation) error

	// Message history (for AI context)
	AddMessage(ctx context.Context, message *Message) error
	GetRecentMessages(ctx context.Context, conversationID string, limit int) ([]*Message, error)
}

// IAgentService defines business logic for agents
type IAgentService interface {
	CreateAgent(ctx context.Context, req CreateAgentRequest) (*AgentResponse, error)
	GetAgent(ctx context.Context, id string) (*AgentResponse, error)
	GetAllAgents(ctx context.Context) ([]*AgentResponse, error)
	UpdateAgent(ctx context.Context, id string, req UpdateAgentRequest) (*AgentResponse, error)
	DeleteAgent(ctx context.Context, id string) error

	// Integration management
	ConnectWhatsApp(ctx context.Context, agentID string) (qrCode string, err error)
	DisconnectWhatsApp(ctx context.Context, agentID string) error
	ConnectTelegram(ctx context.Context, agentID string, botToken string) error
	DisconnectTelegram(ctx context.Context, agentID string) error
	ConnectInstagram(ctx context.Context, agentID string, accessToken string) error
	DisconnectInstagram(ctx context.Context, agentID string) error

	// Handle incoming messages
	HandleIncomingMessage(ctx context.Context, agentID, integrationID, remoteJID, message string) (response string, err error)
}

