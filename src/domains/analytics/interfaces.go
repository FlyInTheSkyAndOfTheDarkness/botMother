package analytics

import "context"

type IAnalyticsRepository interface {
	// Dashboard stats
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
	
	// Message time series
	GetMessagesDaily(ctx context.Context, period string) ([]MessageTimeSeries, error)
	
	// Recent activity
	GetRecentActivity(ctx context.Context, limit int) ([]Activity, error)
	
	// Agent statistics
	GetAgentStats(ctx context.Context, period string) ([]AgentStats, error)
}

