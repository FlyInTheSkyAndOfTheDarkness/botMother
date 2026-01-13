package calendar

import (
	"context"
	"time"
)

// CalendarCredential represents Google Calendar API credentials
type CalendarCredential struct {
	ID           string    `json:"id"`
	AgentID      string    `json:"agent_id"`
	Name         string    `json:"name"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenExpiry  time.Time `json:"token_expiry,omitempty"`
	CalendarID   string    `json:"calendar_id"` // Default: "primary"
	IsConnected  bool      `json:"is_connected"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CalendarEvent represents a calendar event
type CalendarEvent struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Location    string    `json:"location,omitempty"`
	Attendees   []string  `json:"attendees,omitempty"`
	AllDay      bool      `json:"all_day"`
	Status      string    `json:"status"` // confirmed, tentative, cancelled
}

// CreateEventRequest represents a request to create a calendar event
type CreateEventRequest struct {
	AgentID     string    `json:"agent_id" validate:"required"`
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description,omitempty"`
	StartTime   time.Time `json:"start_time" validate:"required"`
	EndTime     time.Time `json:"end_time" validate:"required"`
	Location    string    `json:"location,omitempty"`
	Attendees   []string  `json:"attendees,omitempty"`
	AllDay      bool      `json:"all_day"`
}

// GetEventsRequest represents a request to get calendar events
type GetEventsRequest struct {
	AgentID   string    `json:"agent_id" validate:"required"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
	Query     string    `json:"query,omitempty"` // Search query
}

// AvailabilityRequest represents a request to check availability
type AvailabilityRequest struct {
	AgentID   string    `json:"agent_id" validate:"required"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
	Duration  int       `json:"duration"` // Duration in minutes
}

// TimeSlot represents an available time slot
type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// ICalendarRepository defines database operations for calendar credentials
type ICalendarRepository interface {
	CreateCredential(ctx context.Context, cred *CalendarCredential) error
	GetCredential(ctx context.Context, id string) (*CalendarCredential, error)
	GetCredentialByAgentID(ctx context.Context, agentID string) (*CalendarCredential, error)
	UpdateCredential(ctx context.Context, cred *CalendarCredential) error
	DeleteCredential(ctx context.Context, id string) error
}

// ICalendarService defines business logic for calendar operations
type ICalendarService interface {
	// Credential management
	SaveCredential(ctx context.Context, cred *CalendarCredential) error
	GetCredential(ctx context.Context, agentID string) (*CalendarCredential, error)
	DeleteCredential(ctx context.Context, id string) error
	
	// OAuth flow
	GetAuthURL(ctx context.Context, agentID string) (string, error)
	HandleCallback(ctx context.Context, agentID string, code string) error
	
	// Calendar operations
	CreateEvent(ctx context.Context, req CreateEventRequest) (*CalendarEvent, error)
	GetEvents(ctx context.Context, req GetEventsRequest) ([]*CalendarEvent, error)
	GetAvailableSlots(ctx context.Context, req AvailabilityRequest) ([]TimeSlot, error)
	CancelEvent(ctx context.Context, agentID string, eventID string) error
}

// NaturalLanguageParser helps parse natural language into calendar requests
type NaturalLanguageParser struct{}

// ParseEventRequest attempts to parse natural language into an event request
func (p *NaturalLanguageParser) ParseEventRequest(text string) (*CreateEventRequest, error) {
	// This would use AI to parse natural language like:
	// "Schedule a meeting tomorrow at 3pm for 1 hour with john@example.com"
	// For now, return nil - actual implementation would use OpenAI
	return nil, nil
}

// ParseTimeQuery attempts to parse natural language into a time query
func (p *NaturalLanguageParser) ParseTimeQuery(text string) (*GetEventsRequest, error) {
	// This would parse things like:
	// "What do I have scheduled for tomorrow?"
	// "Show my meetings next week"
	return nil, nil
}


