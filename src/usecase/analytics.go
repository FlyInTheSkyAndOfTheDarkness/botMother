package usecase

import (
	"context"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/analytics"
	analyticsRepo "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/analytics"
)

type AnalyticsService struct {
	repo *analyticsRepo.SQLiteRepository
}

func NewAnalyticsService(repo *analyticsRepo.SQLiteRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) GetDashboard(ctx context.Context) (*analytics.DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx)
}

func (s *AnalyticsService) GetAgentAnalytics(ctx context.Context, agentID string, period string) (*analytics.AgentStats, error) {
	timeRange := s.periodToTimeRange(period)
	return s.repo.GetAgentStats(ctx, agentID, timeRange)
}

func (s *AnalyticsService) GetAllAgentsAnalytics(ctx context.Context, period string) ([]*analytics.AgentStats, error) {
	timeRange := s.periodToTimeRange(period)
	return s.repo.GetAllAgentsStats(ctx, timeRange)
}

func (s *AnalyticsService) GetMessagesByHour(ctx context.Context, period string) ([]analytics.MessagesByHour, error) {
	timeRange := s.periodToTimeRange(period)
	return s.repo.GetMessagesByHour(ctx, timeRange)
}

func (s *AnalyticsService) GetMessagesByDay(ctx context.Context, period string) ([]analytics.MessagesByDay, error) {
	timeRange := s.periodToTimeRange(period)
	return s.repo.GetMessagesByDay(ctx, timeRange)
}

func (s *AnalyticsService) GetRecentActivity(ctx context.Context, limit int) ([]analytics.RecentActivity, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.GetRecentActivity(ctx, limit)
}

func (s *AnalyticsService) periodToTimeRange(period string) analytics.AnalyticsTimeRange {
	switch period {
	case "today":
		return analytics.TimeRangeToday()
	case "yesterday":
		return analytics.TimeRangeYesterday()
	case "week":
		return analytics.TimeRangeThisWeek()
	case "month":
		return analytics.TimeRangeThisMonth()
	case "7days":
		return analytics.TimeRangeLast7Days()
	case "30days":
		return analytics.TimeRangeLast30Days()
	default:
		return analytics.TimeRangeLast7Days()
	}
}

