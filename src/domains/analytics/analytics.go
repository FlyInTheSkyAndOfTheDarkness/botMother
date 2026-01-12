package analytics

import (
	"context"
	"time"
)

// DashboardStats represents overall dashboard statistics
type DashboardStats struct {
	TotalAgents        int     `json:"total_agents"`
	ActiveAgents       int     `json:"active_agents"`
	TotalConversations int     `json:"total_conversations"`
	TotalMessages      int     `json:"total_messages"`
	MessagesToday      int     `json:"messages_today"`
	MessagesThisWeek   int     `json:"messages_this_week"`
	AvgResponseTime    float64 `json:"avg_response_time_ms"`
	ActiveChats        int     `json:"active_chats"` // Chats with activity in last 24h
}

// AgentStats represents statistics for a single agent
type AgentStats struct {
	AgentID            string  `json:"agent_id"`
	AgentName          string  `json:"agent_name"`
	TotalConversations int     `json:"total_conversations"`
	TotalMessages      int     `json:"total_messages"`
	MessagesReceived   int     `json:"messages_received"`
	MessagesSent       int     `json:"messages_sent"`
	AvgResponseTime    float64 `json:"avg_response_time_ms"`
	SatisfactionScore  float64 `json:"satisfaction_score"` // 0-5 based on sentiment
}

// MessagesByHour represents message count per hour
type MessagesByHour struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// MessagesByDay represents message count per day
type MessagesByDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// ConversationStats represents conversation-level analytics
type ConversationStats struct {
	TotalConversations   int     `json:"total_conversations"`
	AvgMessagesPerConvo  float64 `json:"avg_messages_per_conversation"`
	AvgConvoDuration     float64 `json:"avg_conversation_duration_min"`
	ResolvedConversations int    `json:"resolved_conversations"`
	UnresolvedConversations int  `json:"unresolved_conversations"`
}

// TopKeyword represents frequently used keywords
type TopKeyword struct {
	Keyword string `json:"keyword"`
	Count   int    `json:"count"`
}

// SentimentBreakdown represents sentiment distribution
type SentimentBreakdown struct {
	Positive int `json:"positive"`
	Neutral  int `json:"neutral"`
	Negative int `json:"negative"`
}

// IntegrationStats represents stats per integration type
type IntegrationStats struct {
	Type           string `json:"type"` // whatsapp, telegram, instagram
	MessagesCount  int    `json:"messages_count"`
	ConversationsCount int `json:"conversations_count"`
	IsConnected    bool   `json:"is_connected"`
}

// RecentActivity represents recent activity item
type RecentActivity struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // message_received, message_sent, agent_created, flow_triggered
	Description string    `json:"description"`
	AgentID     string    `json:"agent_id,omitempty"`
	AgentName   string    `json:"agent_name,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// AnalyticsTimeRange represents time range for queries
type AnalyticsTimeRange struct {
	Start time.Time
	End   time.Time
}

// Predefined time ranges
func TimeRangeToday() AnalyticsTimeRange {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return AnalyticsTimeRange{Start: start, End: now}
}

func TimeRangeYesterday() AnalyticsTimeRange {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return AnalyticsTimeRange{Start: start, End: end}
}

func TimeRangeThisWeek() AnalyticsTimeRange {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	return AnalyticsTimeRange{Start: start, End: now}
}

func TimeRangeThisMonth() AnalyticsTimeRange {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return AnalyticsTimeRange{Start: start, End: now}
}

func TimeRangeLast7Days() AnalyticsTimeRange {
	now := time.Now()
	start := now.AddDate(0, 0, -7)
	return AnalyticsTimeRange{Start: start, End: now}
}

func TimeRangeLast30Days() AnalyticsTimeRange {
	now := time.Now()
	start := now.AddDate(0, 0, -30)
	return AnalyticsTimeRange{Start: start, End: now}
}

// IAnalyticsRepository defines database operations for analytics
type IAnalyticsRepository interface {
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
	GetAgentStats(ctx context.Context, agentID string, timeRange AnalyticsTimeRange) (*AgentStats, error)
	GetAllAgentsStats(ctx context.Context, timeRange AnalyticsTimeRange) ([]*AgentStats, error)
	GetMessagesByHour(ctx context.Context, timeRange AnalyticsTimeRange) ([]MessagesByHour, error)
	GetMessagesByDay(ctx context.Context, timeRange AnalyticsTimeRange) ([]MessagesByDay, error)
	GetRecentActivity(ctx context.Context, limit int) ([]RecentActivity, error)
	GetTopKeywords(ctx context.Context, agentID string, limit int) ([]TopKeyword, error)
	GetIntegrationStats(ctx context.Context, agentID string) ([]IntegrationStats, error)
}

// IAnalyticsService defines business logic for analytics
type IAnalyticsService interface {
	GetDashboard(ctx context.Context) (*DashboardStats, error)
	GetAgentAnalytics(ctx context.Context, agentID string, period string) (*AgentStats, error)
	GetAllAgentsAnalytics(ctx context.Context, period string) ([]*AgentStats, error)
	GetMessageChart(ctx context.Context, period string, groupBy string) (interface{}, error)
	GetRecentActivity(ctx context.Context, limit int) ([]RecentActivity, error)
}

