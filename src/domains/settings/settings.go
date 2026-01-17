package settings

import (
	"context"
	"time"
)

// WorkingHours represents agent's working hours configuration
type WorkingHours struct {
	Enabled     bool          `json:"enabled"`
	Timezone    string        `json:"timezone"` // e.g., "Europe/Moscow"
	Schedule    []DaySchedule `json:"schedule"`
	AwayMessage string        `json:"away_message"` // Message when outside working hours
}

// DaySchedule represents working hours for a specific day
type DaySchedule struct {
	Day       int    `json:"day"`        // 0=Sunday, 1=Monday, etc.
	IsWorking bool   `json:"is_working"` // false = day off
	StartTime string `json:"start_time"` // "09:00"
	EndTime   string `json:"end_time"`   // "18:00"
}

// TranslationSettings represents auto-translation configuration
type TranslationSettings struct {
	Enabled           bool   `json:"enabled"`
	SourceLanguage    string `json:"source_language"`    // Agent's language (e.g., "en")
	AutoDetect        bool   `json:"auto_detect"`        // Auto-detect incoming language
	TranslateIncoming bool   `json:"translate_incoming"` // Translate user messages to agent language
	TranslateOutgoing bool   `json:"translate_outgoing"` // Translate agent responses to user language
}

// FollowUpSettings represents follow-up automation
type FollowUpSettings struct {
	Enabled       bool     `json:"enabled"`
	DelayMinutes  int      `json:"delay_minutes"`    // Time after last message
	MaxFollowUps  int      `json:"max_follow_ups"`   // Max follow-ups per conversation
	Messages      []string `json:"messages"`         // Follow-up message templates
	OnlyIfNoReply bool     `json:"only_if_no_reply"` // Only send if user hasn't replied
}

// SentimentSettings represents sentiment analysis configuration
type SentimentSettings struct {
	Enabled                bool    `json:"enabled"`
	AlertOnNegative        bool    `json:"alert_on_negative"`         // Send alert for negative sentiment
	NegativeThreshold      float64 `json:"negative_threshold"`        // Threshold for negative (0-1)
	EscalateOnVeryNegative bool    `json:"escalate_on_very_negative"` // Auto-escalate to human
}

// AgentSettings represents all configurable settings for an agent
type AgentSettings struct {
	ID              string              `json:"id"`
	AgentID         string              `json:"agent_id"`
	WorkingHours    WorkingHours        `json:"working_hours"`
	Translation     TranslationSettings `json:"translation"`
	FollowUp        FollowUpSettings    `json:"follow_up"`
	Sentiment       SentimentSettings   `json:"sentiment"`
	MaxTokensPerMsg int                 `json:"max_tokens_per_msg"` // Max response length
	Temperature     float64             `json:"temperature"`        // AI creativity (0-1)
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

// BroadcastMessage represents a broadcast/bulk message
type BroadcastMessage struct {
	ID              string     `json:"id"`
	AgentID         string     `json:"agent_id"`
	IntegrationType string     `json:"integration_type"` // whatsapp, telegram, etc.
	Message         string     `json:"message"`
	MediaURL        string     `json:"media_url,omitempty"`
	Recipients      []string   `json:"recipients"` // List of JIDs/chat IDs
	Status          string     `json:"status"`     // pending, sending, completed, failed
	TotalRecipients int        `json:"total_recipients"`
	SentCount       int        `json:"sent_count"`
	FailedCount     int        `json:"failed_count"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreateBroadcastRequest represents request to create a broadcast
type CreateBroadcastRequest struct {
	AgentID         string     `json:"agent_id" validate:"required"`
	IntegrationType string     `json:"integration_type" validate:"required"`
	Message         string     `json:"message" validate:"required"`
	MediaURL        string     `json:"media_url,omitempty"`
	Recipients      []string   `json:"recipients" validate:"required,min=1"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
}

// ISettingsRepository defines database operations for settings
type ISettingsRepository interface {
	GetAgentSettings(ctx context.Context, agentID string) (*AgentSettings, error)
	SaveAgentSettings(ctx context.Context, settings *AgentSettings) error

	CreateBroadcast(ctx context.Context, broadcast *BroadcastMessage) error
	GetBroadcast(ctx context.Context, id string) (*BroadcastMessage, error)
	GetBroadcastsByAgentID(ctx context.Context, agentID string) ([]*BroadcastMessage, error)
	UpdateBroadcast(ctx context.Context, broadcast *BroadcastMessage) error
}

// ISettingsService defines business logic for settings
type ISettingsService interface {
	GetAgentSettings(ctx context.Context, agentID string) (*AgentSettings, error)
	UpdateAgentSettings(ctx context.Context, settings *AgentSettings) error
	IsWithinWorkingHours(ctx context.Context, agentID string) (bool, string, error)

	CreateBroadcast(ctx context.Context, req CreateBroadcastRequest) (*BroadcastMessage, error)
	GetBroadcasts(ctx context.Context, agentID string) ([]*BroadcastMessage, error)
	GetBroadcast(ctx context.Context, id string) (*BroadcastMessage, error)
	UpdateBroadcastStatus(ctx context.Context, id string, status string, sentCount, failedCount int) error
}

// DefaultAgentSettings returns default settings for a new agent
func DefaultAgentSettings(agentID string) *AgentSettings {
	return &AgentSettings{
		AgentID: agentID,
		WorkingHours: WorkingHours{
			Enabled:     false,
			Timezone:    "UTC",
			AwayMessage: "We're currently outside of working hours. We'll get back to you soon!",
			Schedule: []DaySchedule{
				{Day: 0, IsWorking: false},
				{Day: 1, IsWorking: true, StartTime: "09:00", EndTime: "18:00"},
				{Day: 2, IsWorking: true, StartTime: "09:00", EndTime: "18:00"},
				{Day: 3, IsWorking: true, StartTime: "09:00", EndTime: "18:00"},
				{Day: 4, IsWorking: true, StartTime: "09:00", EndTime: "18:00"},
				{Day: 5, IsWorking: true, StartTime: "09:00", EndTime: "18:00"},
				{Day: 6, IsWorking: false},
			},
		},
		Translation: TranslationSettings{
			Enabled:           false,
			SourceLanguage:    "en",
			AutoDetect:        true,
			TranslateIncoming: true,
			TranslateOutgoing: true,
		},
		FollowUp: FollowUpSettings{
			Enabled:       false,
			DelayMinutes:  30,
			MaxFollowUps:  2,
			OnlyIfNoReply: true,
			Messages: []string{
				"Hi! Just checking if you have any questions?",
				"Is there anything else I can help you with?",
			},
		},
		Sentiment: SentimentSettings{
			Enabled:                false,
			AlertOnNegative:        true,
			NegativeThreshold:      0.3,
			EscalateOnVeryNegative: false,
		},
		MaxTokensPerMsg: 500,
		Temperature:     0.7,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}
