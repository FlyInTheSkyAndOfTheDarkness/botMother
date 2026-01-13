package analytics

import (
	"context"
	"database/sql"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/analytics"
)

type SQLiteRepository struct {
	agentDB *sql.DB
	chatDB  *sql.DB
}

func NewSQLiteRepository(agentDB, chatDB *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		agentDB: agentDB,
		chatDB:  chatDB,
	}
}

func (r *SQLiteRepository) GetDashboardStats(ctx context.Context) (*analytics.DashboardStats, error) {
	stats := &analytics.DashboardStats{}

	// Get agent counts
	if r.agentDB != nil {
		row := r.agentDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM agents")
		row.Scan(&stats.TotalAgents)

		row = r.agentDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM agents WHERE is_active = 1")
		row.Scan(&stats.ActiveAgents)

		// Get conversation counts
		row = r.agentDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM conversations")
		row.Scan(&stats.TotalConversations)

		// Get message counts
		row = r.agentDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM messages")
		row.Scan(&stats.TotalMessages)

		// Messages today
		today := time.Now().Format("2006-01-02")
		row = r.agentDB.QueryRowContext(ctx, 
			"SELECT COUNT(*) FROM messages WHERE date(timestamp) = ?", today)
		row.Scan(&stats.MessagesToday)

		// Messages this week
		weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		row = r.agentDB.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM messages WHERE date(timestamp) >= ?", weekAgo)
		row.Scan(&stats.MessagesThisWeek)

		// Active chats (last 24h)
		yesterday := time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
		row = r.agentDB.QueryRowContext(ctx,
			"SELECT COUNT(DISTINCT conversation_id) FROM messages WHERE timestamp >= ?", yesterday)
		row.Scan(&stats.ActiveChats)
	}

	return stats, nil
}

func (r *SQLiteRepository) GetAgentStats(ctx context.Context, agentID string, timeRange analytics.AnalyticsTimeRange) (*analytics.AgentStats, error) {
	stats := &analytics.AgentStats{AgentID: agentID}

	if r.agentDB == nil {
		return stats, nil
	}

	// Get agent name
	row := r.agentDB.QueryRowContext(ctx, "SELECT name FROM agents WHERE id = ?", agentID)
	row.Scan(&stats.AgentName)

	startStr := timeRange.Start.Format("2006-01-02 15:04:05")
	endStr := timeRange.End.Format("2006-01-02 15:04:05")

	// Get conversation count
	row = r.agentDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM conversations WHERE agent_id = ? AND created_at >= ? AND created_at <= ?`,
		agentID, startStr, endStr)
	row.Scan(&stats.TotalConversations)

	// Get message counts
	row = r.agentDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM messages m 
		 JOIN conversations c ON m.conversation_id = c.id 
		 WHERE c.agent_id = ? AND m.timestamp >= ? AND m.timestamp <= ?`,
		agentID, startStr, endStr)
	row.Scan(&stats.TotalMessages)

	// Messages by role
	row = r.agentDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM messages m 
		 JOIN conversations c ON m.conversation_id = c.id 
		 WHERE c.agent_id = ? AND m.role = 'user' AND m.timestamp >= ? AND m.timestamp <= ?`,
		agentID, startStr, endStr)
	row.Scan(&stats.MessagesReceived)

	row = r.agentDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM messages m 
		 JOIN conversations c ON m.conversation_id = c.id 
		 WHERE c.agent_id = ? AND m.role = 'assistant' AND m.timestamp >= ? AND m.timestamp <= ?`,
		agentID, startStr, endStr)
	row.Scan(&stats.MessagesSent)

	return stats, nil
}

func (r *SQLiteRepository) GetAllAgentsStats(ctx context.Context, timeRange analytics.AnalyticsTimeRange) ([]*analytics.AgentStats, error) {
	var allStats []*analytics.AgentStats

	if r.agentDB == nil {
		return allStats, nil
	}

	rows, err := r.agentDB.QueryContext(ctx, "SELECT id FROM agents")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agentIDs []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		agentIDs = append(agentIDs, id)
	}

	for _, id := range agentIDs {
		stats, err := r.GetAgentStats(ctx, id, timeRange)
		if err != nil {
			continue
		}
		allStats = append(allStats, stats)
	}

	return allStats, nil
}

func (r *SQLiteRepository) GetMessagesByHour(ctx context.Context, timeRange analytics.AnalyticsTimeRange) ([]analytics.MessagesByHour, error) {
	var result []analytics.MessagesByHour

	if r.agentDB == nil {
		// Return empty 24-hour structure
		for i := 0; i < 24; i++ {
			result = append(result, analytics.MessagesByHour{Hour: i, Count: 0})
		}
		return result, nil
	}

	startStr := timeRange.Start.Format("2006-01-02 15:04:05")
	endStr := timeRange.End.Format("2006-01-02 15:04:05")

	rows, err := r.agentDB.QueryContext(ctx,
		`SELECT strftime('%H', timestamp) as hour, COUNT(*) as count 
		 FROM messages 
		 WHERE timestamp >= ? AND timestamp <= ?
		 GROUP BY hour 
		 ORDER BY hour`,
		startStr, endStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hourMap := make(map[int]int)
	for rows.Next() {
		var hourStr string
		var count int
		rows.Scan(&hourStr, &count)
		var hour int
		if _, err := time.Parse("15", hourStr); err == nil {
			hour = int(hourStr[0]-'0')*10 + int(hourStr[1]-'0')
		}
		hourMap[hour] = count
	}

	// Fill all 24 hours
	for i := 0; i < 24; i++ {
		result = append(result, analytics.MessagesByHour{Hour: i, Count: hourMap[i]})
	}

	return result, nil
}

func (r *SQLiteRepository) GetMessagesByDay(ctx context.Context, timeRange analytics.AnalyticsTimeRange) ([]analytics.MessagesByDay, error) {
	var result []analytics.MessagesByDay

	if r.agentDB == nil {
		return result, nil
	}

	startStr := timeRange.Start.Format("2006-01-02 15:04:05")
	endStr := timeRange.End.Format("2006-01-02 15:04:05")

	rows, err := r.agentDB.QueryContext(ctx,
		`SELECT date(timestamp) as day, COUNT(*) as count 
		 FROM messages 
		 WHERE timestamp >= ? AND timestamp <= ?
		 GROUP BY day 
		 ORDER BY day`,
		startStr, endStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var day string
		var count int
		rows.Scan(&day, &count)
		result = append(result, analytics.MessagesByDay{Date: day, Count: count})
	}

	return result, nil
}

func (r *SQLiteRepository) GetRecentActivity(ctx context.Context, limit int) ([]analytics.RecentActivity, error) {
	var result []analytics.RecentActivity

	if r.agentDB == nil {
		return result, nil
	}

	// Get recent messages as activity
	rows, err := r.agentDB.QueryContext(ctx,
		`SELECT m.id, m.role, m.content, m.timestamp, c.agent_id, a.name
		 FROM messages m
		 JOIN conversations c ON m.conversation_id = c.id
		 LEFT JOIN agents a ON c.agent_id = a.id
		 ORDER BY m.timestamp DESC
		 LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, role, content, agentID string
		var timestamp time.Time
		var agentName sql.NullString

		rows.Scan(&id, &role, &content, &timestamp, &agentID, &agentName)

		activityType := "message_received"
		if role == "assistant" {
			activityType = "message_sent"
		}

		// Truncate content
		desc := content
		if len(desc) > 50 {
			desc = desc[:50] + "..."
		}

		result = append(result, analytics.RecentActivity{
			ID:          id,
			Type:        activityType,
			Description: desc,
			AgentID:     agentID,
			AgentName:   agentName.String,
			Timestamp:   timestamp,
		})
	}

	return result, nil
}

func (r *SQLiteRepository) GetTopKeywords(ctx context.Context, agentID string, limit int) ([]analytics.TopKeyword, error) {
	// This would require more sophisticated text analysis
	// For now, return empty
	return []analytics.TopKeyword{}, nil
}

func (r *SQLiteRepository) GetIntegrationStats(ctx context.Context, agentID string) ([]analytics.IntegrationStats, error) {
	var result []analytics.IntegrationStats

	if r.agentDB == nil {
		return result, nil
	}

	rows, err := r.agentDB.QueryContext(ctx,
		`SELECT type, is_connected FROM integrations WHERE agent_id = ?`, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var intType string
		var isConnected bool
		rows.Scan(&intType, &isConnected)
		result = append(result, analytics.IntegrationStats{
			Type:        intType,
			IsConnected: isConnected,
		})
	}

	return result, nil
}


