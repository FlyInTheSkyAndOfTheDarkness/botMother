package usecase

import (
	"context"
	"fmt"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/agent"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/health"
	agentRepo "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/agent"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/sirupsen/logrus"
)

type HealthService struct {
	agentRepo        *agentRepo.SQLiteRepository
	telegramChecker  func(integrationID string) bool // Function to check if Telegram bot is running
}

func NewHealthService(agentRepo *agentRepo.SQLiteRepository) *HealthService {
	return &HealthService{agentRepo: agentRepo}
}

// SetTelegramChecker sets the function to check Telegram bot status (to avoid import cycle)
func (s *HealthService) SetTelegramChecker(checker func(integrationID string) bool) {
	s.telegramChecker = checker
}

// GetSystemHealth returns overall system health status
func (s *HealthService) GetSystemHealth(ctx context.Context) (*health.SystemHealth, error) {
	systemHealth := &health.SystemHealth{
		Status: "healthy",
	}

	// Check WhatsApp devices
	whatsappHealth := s.checkWhatsAppHealth()
	systemHealth.WhatsApp = whatsappHealth

	// Check Telegram bots
	telegramHealth := s.checkTelegramHealth()
	systemHealth.Telegram = telegramHealth

	// Check all agents and their integrations
	agents, err := s.agentRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents: %w", err)
	}

	systemHealth.TotalAgents = len(agents)
	activeCount := 0
	var agentHealthList []health.AgentHealth

	for _, agent := range agents {
		if agent.IsActive {
			activeCount++
		}

		agentHealth := s.checkAgentHealth(ctx, agent)
		agentHealthList = append(agentHealthList, agentHealth)
	}

	systemHealth.ActiveAgents = activeCount
	systemHealth.Agents = agentHealthList

	// Determine overall status
	if systemHealth.WhatsApp.ConnectedDevices == 0 && systemHealth.Telegram.RunningBots == 0 {
		systemHealth.Status = "unhealthy"
	} else if systemHealth.WhatsApp.ConnectedDevices < systemHealth.WhatsApp.TotalDevices || 
		systemHealth.Telegram.RunningBots < systemHealth.Telegram.TotalBots {
		systemHealth.Status = "degraded"
	}

	return systemHealth, nil
}

// GetAgentHealth returns health status for a specific agent
func (s *HealthService) GetAgentHealth(ctx context.Context, agentID string) (*health.AgentHealth, error) {
	agent, err := s.agentRepo.GetByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	agentHealth := s.checkAgentHealth(ctx, agent)
	return &agentHealth, nil
}

// checkWhatsAppHealth checks WhatsApp devices status
func (s *HealthService) checkWhatsAppHealth() health.WhatsAppHealth {
	health := health.WhatsAppHealth{}

	deviceManager := whatsapp.GetDeviceManager()
	if deviceManager == nil {
		return health
	}

	devices := deviceManager.ListDevices()
	health.TotalDevices = len(devices)

	for _, device := range devices {
		device.UpdateStateFromClient()
		client := device.GetClient()
		if client != nil {
			if client.IsConnected() {
				health.ConnectedDevices++
			}
			if client.IsLoggedIn() {
				health.LoggedInDevices++
			}
		}
	}

	return health
}

// checkTelegramHealth checks Telegram bots status
func (s *HealthService) checkTelegramHealth() health.TelegramHealth {
	health := health.TelegramHealth{}

	// Count Telegram integrations from agents
	ctx := context.Background()
	agents, err := s.agentRepo.GetAll(ctx)
	if err != nil {
		return health
	}

	for _, agent := range agents {
		if !agent.IsActive {
			continue
		}
		integrations, err := s.agentRepo.GetIntegrationsByAgentID(ctx, agent.ID)
		if err != nil {
			continue
		}
		for _, integration := range integrations {
			if integration.Type == "telegram" {
				health.TotalBots++
				if integration.IsConnected && s.telegramChecker != nil && s.telegramChecker(integration.ID) {
					health.RunningBots++
				}
			}
		}
	}

	return health
}

// checkAgentHealth checks health status for a specific agent
func (s *HealthService) checkAgentHealth(ctx context.Context, agent *agent.Agent) health.AgentHealth {
	agentHealth := health.AgentHealth{
		AgentID:   agent.ID,
		AgentName: agent.Name,
		IsActive:  agent.IsActive,
	}

	// Get integrations
	integrations, err := s.agentRepo.GetIntegrationsByAgentID(ctx, agent.ID)
	if err != nil {
		logrus.Warnf("⚠️  [Health] Failed to get integrations for agent %s: %v", agent.ID, err)
		return agentHealth
	}

	var integrationStatuses []health.IntegrationStatus
	for _, integration := range integrations {
		status := s.checkIntegrationStatus(ctx, integration)
		integrationStatuses = append(integrationStatuses, status)
	}

	agentHealth.Integrations = integrationStatuses
	return agentHealth
}

// checkIntegrationStatus checks status of a specific integration
func (s *HealthService) checkIntegrationStatus(ctx context.Context, integration *agent.Integration) health.IntegrationStatus {
	status := health.IntegrationStatus{
		ID:          integration.ID,
		Type:        integration.Type,
		IsConnected: integration.IsConnected,
		IsActive:    true, // Integration itself is always "active" if it exists
		Status:      "disconnected",
	}

	if !integration.IsConnected {
		status.Message = "Integration is not connected"
		return status
	}

	switch integration.Type {
	case agent.IntegrationTypeWhatsApp:
		status = s.checkWhatsAppIntegration(ctx, integration, status)
	case agent.IntegrationTypeTelegram:
		status = s.checkTelegramIntegration(ctx, integration, status)
	case agent.IntegrationTypeInstagram:
		status.Status = "not_implemented"
		status.Message = "Instagram integration not yet implemented"
	default:
		status.Status = "unknown"
		status.Message = fmt.Sprintf("Unknown integration type: %s", integration.Type)
	}

	return status
}

// checkWhatsAppIntegration checks WhatsApp integration status
func (s *HealthService) checkWhatsAppIntegration(ctx context.Context, integration *agent.Integration, status health.IntegrationStatus) health.IntegrationStatus {
	// Parse WhatsApp config
	waConfig, err := agentRepo.ParseWhatsAppConfig(integration.Config)
	if err != nil || waConfig == nil {
		status.Status = "error"
		status.Message = "Failed to parse WhatsApp config"
		return status
	}

	if waConfig.DeviceID == "" {
		status.Status = "error"
		status.Message = "No device ID configured"
		return status
	}

	status.DeviceID = waConfig.DeviceID

	// Get device manager
	deviceManager := whatsapp.GetDeviceManager()
	if deviceManager == nil {
		status.Status = "error"
		status.Message = "Device manager not initialized"
		return status
	}

	// Get device instance
	deviceInstance, ok := deviceManager.GetDevice(waConfig.DeviceID)
	if !ok || deviceInstance == nil {
		status.Status = "error"
		status.Message = fmt.Sprintf("Device %s not found", waConfig.DeviceID)
		return status
	}

	// Update state and check connection
	deviceInstance.UpdateStateFromClient()
	client := deviceInstance.GetClient()
	if client == nil {
		status.Status = "disconnected"
		status.Message = "WhatsApp client not initialized"
		return status
	}

	if client.IsLoggedIn() {
		status.Status = "connected"
		status.Message = "Connected and logged in"
	} else if client.IsConnected() {
		status.Status = "connecting"
		status.Message = "Connected but not logged in"
	} else {
		status.Status = "disconnected"
		status.Message = "Not connected"
	}

	return status
}

// checkTelegramIntegration checks Telegram integration status
func (s *HealthService) checkTelegramIntegration(ctx context.Context, integration *agent.Integration, status health.IntegrationStatus) health.IntegrationStatus {
	// Parse Telegram config
	tgConfig, err := agentRepo.ParseTelegramConfig(integration.Config)
	if err != nil || tgConfig == nil {
		status.Status = "error"
		status.Message = "Failed to parse Telegram config"
		return status
	}

	// Mask bot token for security
	if len(tgConfig.BotToken) > 8 {
		status.BotToken = tgConfig.BotToken[:4] + "..." + tgConfig.BotToken[len(tgConfig.BotToken)-4:]
	}

	// Get bot manager
	// Check if bot is running using checker function (to avoid import cycle)
	if s.telegramChecker != nil && s.telegramChecker(integration.ID) {
		status.Status = "connected"
		status.Message = "Bot is running"
	} else if integration.IsConnected {
		status.Status = "disconnected"
		status.Message = "Bot should be running but is not"
	} else {
		status.Status = "disconnected"
		status.Message = "Bot is not running"
	}

	return status
}

