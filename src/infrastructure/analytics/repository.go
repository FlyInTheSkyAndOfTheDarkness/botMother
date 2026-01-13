package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/analytics"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	agentDB *sql.DB
}

func NewSQLiteRepository(agentDB *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{agentDB: agentDB}
}

// GetDashboardStats returns overall dashboard statistics
func (r *SQLiteRepository) GetDashboardStats(ctx context.Context) (*analytics.DashboardStats, error) {
	stats := &analytics.DashboardStats{}

	// Total and active agents
	err := r.agentDB.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) as total_agents,
			SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END) as active_agents
		FROM agents
	`).Scan(&stats.TotalAgents, &stats.ActiveAgents)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent stats: %w", err)
	}

	// Messages today
	today := time.Now().Truncate(24 * time.Hour)
	err = r.agentDB.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM messages 
		WHERE timestamp >= ?
	`, today).Scan(&stats.MessagesToday)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages today: %w", err)
	}

	// Total messages
	err = r.agentDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM messages`).Scan(&stats.TotalMessages)
	if err != nil {
		return nil, fmt.Errorf("failed to get total messages: %w", err)
	}

	// Total conversations
	err = r.agentDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM conversations`).Scan(&stats.TotalConversations)
	if err != nil {
		return nil, fmt.Errorf("failed to get total conversations: %w", err)
	}

	// Active chats (conversations with messages in last 24 hours)
	last24h := time.Now().Add(-24 * time.Hour)
	err = r.agentDB.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT c.id)
		FROM conversations c
		INNER JOIN messages m ON m.conversation_id = c.id
		WHERE m.timestamp >= ?
	`, last24h).Scan(&stats.ActiveChats)
	if err != nil {
		return nil, fmt.Errorf("failed to get active chats: %w", err)
	}

	// Messages this week
	weekAgo := time.Now().AddDate(0, 0, -7)
	err = r.agentDB.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM messages 
		WHERE timestamp >= ?
	`, weekAgo).Scan(&stats.MessagesThisWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages this week: %w", err)
	}

	return stats, nil
}

// GetMessagesDaily returns message counts grouped by time period
func (r *SQLiteRepository) GetMessagesDaily(ctx context.Context, period string) ([]analytics.MessageTimeSeries, error) {
	var query string
	var startTime time.Time
	now := time.Now()

	switch period {
	case "today":
		startTime = now.Truncate(24 * time.Hour)
		query = `
			SELECT 
				strftime('%H', timestamp) as hour,
				COUNT(*) as count
			FROM messages
			WHERE timestamp >= ?
			GROUP BY strftime('%H', timestamp)
			ORDER BY hour ASC
		`
	case "7days", "30days":
		days := 7
		if period == "30days" {
			days = 30
		}
		startTime = now.AddDate(0, 0, -days)
		query = `
			SELECT 
				date(timestamp) as date,
				COUNT(*) as count
			FROM messages
			WHERE timestamp >= ?
			GROUP BY date(timestamp)
			ORDER BY date ASC
		`
	case "month":
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		query = `
			SELECT 
				date(timestamp) as date,
				COUNT(*) as count
			FROM messages
			WHERE timestamp >= ?
			GROUP BY date(timestamp)
			ORDER BY date ASC
		`
	default:
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	rows, err := r.agentDB.QueryContext(ctx, query, startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var result []analytics.MessageTimeSeries
	for rows.Next() {
		var item analytics.MessageTimeSeries
		if period == "today" {
			var hour int
			err := rows.Scan(&hour, &item.Count)
			if err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}
			item.Hour = hour
		} else {
			var dateStr string
			err := rows.Scan(&dateStr, &item.Count)
			if err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}
			item.Date = dateStr
		}
		result = append(result, item)
	}

	return result, rows.Err()
}

// GetRecentActivity returns recent activity
func (r *SQLiteRepository) GetRecentActivity(ctx context.Context, limit int) ([]analytics.Activity, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	query := `
		SELECT 
			m.id,
			m.role,
			m.content,
			m.timestamp,
			c.agent_id,
			a.name as agent_name
		FROM messages m
		INNER JOIN conversations c ON c.id = m.conversation_id
		LEFT JOIN agents a ON a.id = c.agent_id
		ORDER BY m.timestamp DESC
		LIMIT ?
	`

	rows, err := r.agentDB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query activity: %w", err)
	}
	defer rows.Close()

	var result []analytics.Activity
	for rows.Next() {
		var act analytics.Activity
		var role string
		var content string
		err := rows.Scan(&act.ID, &role, &content, &act.Timestamp, &act.AgentID, &act.AgentName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Determine type and description
		if role == "user" {
			act.Type = "message_received"
			act.Description = fmt.Sprintf("Received: %s", truncateString(content, 50))
		} else {
			act.Type = "message_sent"
			act.Description = fmt.Sprintf("Sent: %s", truncateString(content, 50))
		}

		result = append(result, act)
	}

	return result, rows.Err()
}

// GetAgentStats returns statistics per agent
func (r *SQLiteRepository) GetAgentStats(ctx context.Context, period string) ([]analytics.AgentStats, error) {
	var startTime time.Time
	now := time.Now()

	switch period {
	case "today":
		startTime = now.Truncate(24 * time.Hour)
	case "7days":
		startTime = now.AddDate(0, 0, -7)
	case "30days":
		startTime = now.AddDate(0, 0, -30)
	case "month":
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	default:
		startTime = time.Time{} // All time
	}

	var query string
	var args []interface{}

	if !startTime.IsZero() {
		query = `
			SELECT 
				a.id as agent_id,
				a.name as agent_name,
				COUNT(DISTINCT c.id) as total_conversations,
				COUNT(m.id) as total_messages,
				SUM(CASE WHEN m.role = 'user' THEN 1 ELSE 0 END) as messages_received,
				SUM(CASE WHEN m.role = 'assistant' THEN 1 ELSE 0 END) as messages_sent
			FROM agents a
			LEFT JOIN conversations c ON c.agent_id = a.id
			LEFT JOIN messages m ON m.conversation_id = c.id AND m.timestamp >= ?
			GROUP BY a.id, a.name
			ORDER BY total_messages DESC
		`
		args = []interface{}{startTime}
	} else {
		query = `
			SELECT 
				a.id as agent_id,
				a.name as agent_name,
				COUNT(DISTINCT c.id) as total_conversations,
				COUNT(m.id) as total_messages,
				SUM(CASE WHEN m.role = 'user' THEN 1 ELSE 0 END) as messages_received,
				SUM(CASE WHEN m.role = 'assistant' THEN 1 ELSE 0 END) as messages_sent
			FROM agents a
			LEFT JOIN conversations c ON c.agent_id = a.id
			LEFT JOIN messages m ON m.conversation_id = c.id
			GROUP BY a.id, a.name
			ORDER BY total_messages DESC
		`
		args = []interface{}{}
	}

	rows, err := r.agentDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query agent stats: %w", err)
	}
	defer rows.Close()

	var result []analytics.AgentStats
	for rows.Next() {
		var stat analytics.AgentStats
		err := rows.Scan(
			&stat.AgentID,
			&stat.AgentName,
			&stat.TotalConversations,
			&stat.TotalMessages,
			&stat.MessagesReceived,
			&stat.MessagesSent,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result = append(result, stat)
	}

	return result, rows.Err()
}

// Helper function to truncate string
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
