package usecase

import (
	"context"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/analytics"
)

type AnalyticsService struct {
	repo analytics.IAnalyticsRepository
}

func NewAnalyticsService(repo analytics.IAnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

// GetDashboardStats returns overall dashboard statistics
func (s *AnalyticsService) GetDashboardStats(ctx context.Context) (*analytics.DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx)
}

// GetMessagesDaily returns message counts grouped by time period
func (s *AnalyticsService) GetMessagesDaily(ctx context.Context, period string) ([]analytics.MessageTimeSeries, error) {
	return s.repo.GetMessagesDaily(ctx, period)
}

// GetRecentActivity returns recent activity
func (s *AnalyticsService) GetRecentActivity(ctx context.Context, limit int) ([]analytics.Activity, error) {
	return s.repo.GetRecentActivity(ctx, limit)
}

// GetAgentStats returns statistics per agent
func (s *AnalyticsService) GetAgentStats(ctx context.Context, period string) ([]analytics.AgentStats, error) {
	return s.repo.GetAgentStats(ctx, period)
}
