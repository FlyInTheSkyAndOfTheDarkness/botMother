package flow

import (
	"context"
	"time"
)

// Node types
const (
	// Triggers
	NodeTypeTriggerWhatsApp  = "trigger_whatsapp"
	NodeTypeTriggerTelegram  = "trigger_telegram"
	NodeTypeTriggerInstagram = "trigger_instagram"
	NodeTypeTriggerWebhook   = "trigger_webhook"
	NodeTypeTriggerSchedule  = "trigger_schedule"

	// AI & Logic
	NodeTypeAIAgent   = "ai_agent"
	NodeTypeCondition = "condition"
	NodeTypeSwitch    = "switch"
	NodeTypeDelay     = "delay"
	NodeTypeCode      = "code"

	// Integrations
	NodeTypeHTTPRequest  = "http_request"
	NodeTypeDatabase     = "database"
	NodeTypeGoogleSheets = "google_sheets"
	NodeTypeSerpAPI      = "serp_api"
	NodeTypeEmail        = "email"
	NodeTypeWebhookOut   = "webhook_out"

	// Actions
	NodeTypeSendMessage = "send_message"
	NodeTypeSendImage   = "send_image"
	NodeTypeSendFile    = "send_file"
	NodeTypeSetVariable = "set_variable"
)

// Credential types
const (
	CredentialTypeDatabase     = "database"
	CredentialTypeOpenAI       = "openai"
	CredentialTypeGoogleSheets = "google_sheets"
	CredentialTypeSMTP         = "smtp"
	CredentialTypeSerpAPI      = "serp_api"
	CredentialTypeCustomAPI    = "custom_api"
)

// Flow represents a workflow/automation
type Flow struct {
	ID          string    `json:"id"`
	AgentID     string    `json:"agent_id"`     // Which agent this flow belongs to
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	Nodes       []Node    `json:"nodes"`
	Edges       []Edge    `json:"edges"`
	Variables   []Variable `json:"variables,omitempty"` // Flow-level variables
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Node represents a single node in the flow
type Node struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`     // Node type constant
	Label    string                 `json:"label"`    // Display name
	Position Position               `json:"position"` // X, Y coordinates
	Data     map[string]interface{} `json:"data"`     // Node-specific configuration
}

// Position represents X,Y coordinates for a node
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Edge represents a connection between two nodes
type Edge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`        // Source node ID
	Target       string `json:"target"`        // Target node ID
	SourceHandle string `json:"source_handle,omitempty"` // For nodes with multiple outputs (condition)
	TargetHandle string `json:"target_handle,omitempty"`
	Label        string `json:"label,omitempty"`
}

// Variable represents a flow variable
type Variable struct {
	Name    string      `json:"name"`
	Value   interface{} `json:"value"`
	Type    string      `json:"type"` // string, number, boolean, object, array
}

// Credential stores external service credentials (encrypted)
type Credential struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`   // Credentials are per-agent
	Name      string    `json:"name"`       // Display name
	Type      string    `json:"type"`       // Credential type constant
	Config    string    `json:"config"`     // JSON encrypted config
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DatabaseCredential holds database connection settings
type DatabaseCredential struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"` // disable, require, verify-ca, verify-full
}

// OpenAICredential holds OpenAI API settings
type OpenAICredential struct {
	APIKey       string `json:"api_key"`
	Organization string `json:"organization,omitempty"`
	BaseURL      string `json:"base_url,omitempty"` // For proxies or Azure
}

// SMTPCredential holds email server settings
type SMTPCredential struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	FromName string `json:"from_name"`
	FromEmail string `json:"from_email"`
	UseTLS   bool   `json:"use_tls"`
}

// GoogleSheetsCredential holds Google API settings
type GoogleSheetsCredential struct {
	ServiceAccountJSON string `json:"service_account_json"`
}

// CustomAPICredential holds custom API settings
type CustomAPICredential struct {
	BaseURL string            `json:"base_url"`
	Headers map[string]string `json:"headers"`
	AuthType string           `json:"auth_type"` // none, basic, bearer, api_key
	AuthValue string          `json:"auth_value"`
}

// === Node Data Structures ===

// TriggerNodeData for trigger nodes
type TriggerNodeData struct {
	IntegrationID string   `json:"integration_id,omitempty"` // WhatsApp/Telegram integration
	MessageTypes  []string `json:"message_types,omitempty"`  // text, image, audio, etc.
	FilterKeywords []string `json:"filter_keywords,omitempty"`
}

// AIAgentNodeData for AI agent nodes
type AIAgentNodeData struct {
	CredentialID string `json:"credential_id"` // OpenAI credential
	Model        string `json:"model"`
	SystemPrompt string `json:"system_prompt"`
	MaxTokens    int    `json:"max_tokens"`
	Temperature  float64 `json:"temperature"`
	UseContext   bool   `json:"use_context"` // Use conversation history
}

// HTTPRequestNodeData for HTTP request nodes
type HTTPRequestNodeData struct {
	Method      string            `json:"method"` // GET, POST, PUT, DELETE, PATCH
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	QueryParams map[string]string `json:"query_params,omitempty"`
	Body        string            `json:"body,omitempty"`
	BodyType    string            `json:"body_type"` // json, form, raw
	Timeout     int               `json:"timeout"`   // seconds
	CredentialID string           `json:"credential_id,omitempty"` // Optional API credential
}

// DatabaseNodeData for database nodes
type DatabaseNodeData struct {
	CredentialID string `json:"credential_id"` // Database credential
	Operation    string `json:"operation"`     // select, insert, update, delete, raw
	Table        string `json:"table,omitempty"`
	Query        string `json:"query,omitempty"`      // For raw SQL
	Columns      []string `json:"columns,omitempty"`  // For select
	Where        string `json:"where,omitempty"`      // WHERE clause
	Values       map[string]interface{} `json:"values,omitempty"` // For insert/update
	Limit        int    `json:"limit,omitempty"`
}

// ConditionNodeData for condition nodes
type ConditionNodeData struct {
	Conditions []Condition `json:"conditions"`
	CombineWith string     `json:"combine_with"` // and, or
}

// Condition represents a single condition
type Condition struct {
	Field    string      `json:"field"`    // Variable or path to check
	Operator string      `json:"operator"` // eq, ne, gt, lt, gte, lte, contains, starts_with, ends_with
	Value    interface{} `json:"value"`
}

// DelayNodeData for delay nodes
type DelayNodeData struct {
	Duration int    `json:"duration"` // Amount
	Unit     string `json:"unit"`     // seconds, minutes, hours
}

// SendMessageNodeData for send message nodes
type SendMessageNodeData struct {
	IntegrationID string `json:"integration_id,omitempty"` // If not set, reply to trigger
	Message       string `json:"message"`                  // Can include {{variables}}
	ReplyToTrigger bool  `json:"reply_to_trigger"`
}

// CodeNodeData for custom code nodes
type CodeNodeData struct {
	Language string `json:"language"` // javascript
	Code     string `json:"code"`
}

// === Request/Response DTOs ===

// CreateFlowRequest for creating a new flow
type CreateFlowRequest struct {
	AgentID     string     `json:"agent_id" validate:"required"`
	Name        string     `json:"name" validate:"required"`
	Description string     `json:"description,omitempty"`
	Nodes       []Node     `json:"nodes"`
	Edges       []Edge     `json:"edges"`
}

// UpdateFlowRequest for updating a flow
type UpdateFlowRequest struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
	Nodes       []Node     `json:"nodes,omitempty"`
	Edges       []Edge     `json:"edges,omitempty"`
	Variables   []Variable `json:"variables,omitempty"`
}

// FlowResponse for API responses
type FlowResponse struct {
	ID          string     `json:"id"`
	AgentID     string     `json:"agent_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	IsActive    bool       `json:"is_active"`
	Nodes       []Node     `json:"nodes"`
	Edges       []Edge     `json:"edges"`
	Variables   []Variable `json:"variables"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateCredentialRequest for creating credentials
type CreateCredentialRequest struct {
	AgentID string `json:"agent_id" validate:"required"`
	Name    string `json:"name" validate:"required"`
	Type    string `json:"type" validate:"required"`
	Config  string `json:"config" validate:"required"` // JSON config
}

// UpdateCredentialRequest for updating credentials
type UpdateCredentialRequest struct {
	Name   *string `json:"name,omitempty"`
	Config *string `json:"config,omitempty"`
}

// CredentialResponse for API responses (masked sensitive data)
type CredentialResponse struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// === Interfaces ===

// IFlowRepository defines database operations for flows
type IFlowRepository interface {
	// Flow CRUD
	CreateFlow(ctx context.Context, flow *Flow) error
	GetFlowByID(ctx context.Context, id string) (*Flow, error)
	GetFlowsByAgentID(ctx context.Context, agentID string) ([]*Flow, error)
	UpdateFlow(ctx context.Context, flow *Flow) error
	DeleteFlow(ctx context.Context, id string) error

	// Credential CRUD
	CreateCredential(ctx context.Context, credential *Credential) error
	GetCredentialByID(ctx context.Context, id string) (*Credential, error)
	GetCredentialsByAgentID(ctx context.Context, agentID string) ([]*Credential, error)
	UpdateCredential(ctx context.Context, credential *Credential) error
	DeleteCredential(ctx context.Context, id string) error
}

// IFlowService defines business logic for flows
type IFlowService interface {
	// Flow operations
	CreateFlow(ctx context.Context, req CreateFlowRequest) (*FlowResponse, error)
	GetFlow(ctx context.Context, id string) (*FlowResponse, error)
	GetFlowsByAgent(ctx context.Context, agentID string) ([]*FlowResponse, error)
	UpdateFlow(ctx context.Context, id string, req UpdateFlowRequest) (*FlowResponse, error)
	DeleteFlow(ctx context.Context, id string) error
	
	// Credential operations
	CreateCredential(ctx context.Context, req CreateCredentialRequest) (*CredentialResponse, error)
	GetCredential(ctx context.Context, id string) (*CredentialResponse, error)
	GetCredentialsByAgent(ctx context.Context, agentID string) ([]*CredentialResponse, error)
	UpdateCredential(ctx context.Context, id string, req UpdateCredentialRequest) (*CredentialResponse, error)
	DeleteCredential(ctx context.Context, id string) error
	TestDatabaseConnection(ctx context.Context, config DatabaseCredential) error
}

// IFlowExecutor defines flow execution
type IFlowExecutor interface {
	// Execute a flow with given input
	Execute(ctx context.Context, flow *Flow, input map[string]interface{}) (map[string]interface{}, error)
	
	// Execute a single node
	ExecuteNode(ctx context.Context, node *Node, input map[string]interface{}, credentials map[string]*Credential) (map[string]interface{}, error)
}




