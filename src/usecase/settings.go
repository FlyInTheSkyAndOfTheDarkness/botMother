package usecase

import (
	"context"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/settings"
	settingsRepo "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/settings"
)

type SettingsService struct {
	repo *settingsRepo.SQLiteRepository
}

func NewSettingsService(repo *settingsRepo.SQLiteRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

func (s *SettingsService) GetAgentSettings(ctx context.Context, agentID string) (*settings.AgentSettings, error) {
	return s.repo.GetAgentSettings(ctx, agentID)
}

func (s *SettingsService) UpdateAgentSettings(ctx context.Context, agentSettings *settings.AgentSettings) error {
	return s.repo.SaveAgentSettings(ctx, agentSettings)
}

func (s *SettingsService) IsWithinWorkingHours(ctx context.Context, agentID string) (bool, string, error) {
	agentSettings, err := s.repo.GetAgentSettings(ctx, agentID)
	if err != nil {
		return true, "", err // Default to working if error
	}

	if !agentSettings.WorkingHours.Enabled {
		return true, "", nil // Working hours not enabled, always available
	}

	// Get current time in agent's timezone
	loc, err := time.LoadLocation(agentSettings.WorkingHours.Timezone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	
	// Find today's schedule
	dayOfWeek := int(now.Weekday())
	for _, schedule := range agentSettings.WorkingHours.Schedule {
		if schedule.Day == dayOfWeek {
			if !schedule.IsWorking {
				return false, agentSettings.WorkingHours.AwayMessage, nil
			}

			// Parse start and end times
			startTime, _ := time.Parse("15:04", schedule.StartTime)
			endTime, _ := time.Parse("15:04", schedule.EndTime)

			// Create times for today
			startToday := time.Date(now.Year(), now.Month(), now.Day(), 
				startTime.Hour(), startTime.Minute(), 0, 0, loc)
			endToday := time.Date(now.Year(), now.Month(), now.Day(), 
				endTime.Hour(), endTime.Minute(), 0, 0, loc)

			if now.Before(startToday) || now.After(endToday) {
				return false, agentSettings.WorkingHours.AwayMessage, nil
			}

			return true, "", nil
		}
	}

	return true, "", nil
}

func (s *SettingsService) CreateBroadcast(ctx context.Context, req settings.CreateBroadcastRequest) (*settings.BroadcastMessage, error) {
	broadcast := &settings.BroadcastMessage{
		AgentID:         req.AgentID,
		IntegrationType: req.IntegrationType,
		Message:         req.Message,
		MediaURL:        req.MediaURL,
		Recipients:      req.Recipients,
		Status:          "pending",
		TotalRecipients: len(req.Recipients),
		ScheduledAt:     req.ScheduledAt,
	}

	if err := s.repo.CreateBroadcast(ctx, broadcast); err != nil {
		return nil, err
	}

	return broadcast, nil
}

func (s *SettingsService) GetBroadcasts(ctx context.Context, agentID string) ([]*settings.BroadcastMessage, error) {
	return s.repo.GetBroadcastsByAgentID(ctx, agentID)
}

func (s *SettingsService) GetBroadcast(ctx context.Context, id string) (*settings.BroadcastMessage, error) {
	return s.repo.GetBroadcast(ctx, id)
}

func (s *SettingsService) UpdateBroadcastStatus(ctx context.Context, id string, status string, sentCount, failedCount int) error {
	broadcast, err := s.repo.GetBroadcast(ctx, id)
	if err != nil {
		return err
	}
	
	broadcast.Status = status
	broadcast.SentCount = sentCount
	broadcast.FailedCount = failedCount
	
	if status == "sending" && broadcast.StartedAt == nil {
		now := time.Now()
		broadcast.StartedAt = &now
	}
	if status == "completed" || status == "failed" {
		now := time.Now()
		broadcast.CompletedAt = &now
	}
	
	return s.repo.UpdateBroadcast(ctx, broadcast)
}


