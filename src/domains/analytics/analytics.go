package analytics

import "time"

// DashboardStats represents overall dashboard statistics
type DashboardStats struct {
	TotalAgents      int `json:"total_agents"`
	ActiveAgents     int `json:"active_agents"`
	MessagesToday    int `json:"messages_today"`
	TotalMessages    int `json:"total_messages"`
	TotalConversations int `json:"total_conversations"`
	ActiveChats      int `json:"active_chats"`
	MessagesThisWeek int `json:"messages_this_week"`
}

// MessageTimeSeries represents messages over time
type MessageTimeSeries struct {
	Date  string `json:"date,omitempty"`
	Hour  int    `json:"hour,omitempty"`
	Count int    `json:"count"`
}

// Activity represents recent activity
type Activity struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // message_received, message_sent
	Description string    `json:"description"`
	AgentID     string    `json:"agent_id,omitempty"`
	AgentName   string    `json:"agent_name,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// AgentStats represents statistics for a specific agent
type AgentStats struct {
	AgentID            string `json:"agent_id"`
	AgentName          string `json:"agent_name"`
	TotalConversations int    `json:"total_conversations"`
	TotalMessages      int    `json:"total_messages"`
	MessagesReceived   int    `json:"messages_received"`
	MessagesSent       int    `json:"messages_sent"`
}
