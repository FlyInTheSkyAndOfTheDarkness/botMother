package rest

import (
	"encoding/json"
	"fmt"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/agent"
	domainApp "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/app"
	telegramBot "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/telegram"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/gofiber/fiber/v2"
)

type AgentHandler struct {
	Service    *usecase.AgentService
	AppService domainApp.IAppUsecase
}

func InitRestAgent(app fiber.Router, service *usecase.AgentService, appService domainApp.IAppUsecase) AgentHandler {
	handler := AgentHandler{Service: service, AppService: appService}

	// Agent CRUD
	app.Get("/agents", handler.GetAllAgents)
	app.Post("/agents", handler.CreateAgent)
	app.Get("/agents/:id", handler.GetAgent)
	app.Put("/agents/:id", handler.UpdateAgent)
	app.Delete("/agents/:id", handler.DeleteAgent)

	// Integration endpoints
	app.Post("/agents/:id/integrations/:type", handler.CreateIntegration)
	app.Delete("/agents/:id/integrations/:integrationId", handler.DeleteIntegration)
	app.Post("/agents/:id/integrations/:integrationId/connect", handler.ConnectIntegration)
	app.Post("/agents/:id/integrations/:integrationId/disconnect", handler.DisconnectIntegration)
	
	// WhatsApp QR endpoint
	app.Get("/whatsapp/qr", handler.GetWhatsAppQR)

	return handler
}

// GetAllAgents returns all agents
func (h *AgentHandler) GetAllAgents(c *fiber.Ctx) error {
	agents, err := h.Service.GetAllAgents(c.UserContext())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Agents retrieved successfully",
		Results: agents,
	})
}

// CreateAgent creates a new agent
func (h *AgentHandler) CreateAgent(c *fiber.Ctx) error {
	var req agent.CreateAgentRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// Basic validation
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name is required")
	}
	if req.APIKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "API key is required")
	}
	if req.Model == "" {
		req.Model = "gpt-4o-mini"
	}
	if req.SystemPrompt == "" {
		return fiber.NewError(fiber.StatusBadRequest, "System prompt is required")
	}

	result, err := h.Service.CreateAgent(c.UserContext(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(utils.ResponseData{
		Status:  201,
		Code:    "SUCCESS",
		Message: "Agent created successfully",
		Results: result,
	})
}

// GetAgent returns a single agent by ID
func (h *AgentHandler) GetAgent(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID is required")
	}

	result, err := h.Service.GetAgent(c.UserContext(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Agent retrieved successfully",
		Results: result,
	})
}

// UpdateAgent updates an agent
func (h *AgentHandler) UpdateAgent(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID is required")
	}

	var req agent.UpdateAgentRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	result, err := h.Service.UpdateAgent(c.UserContext(), id, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Agent updated successfully",
		Results: result,
	})
}

// DeleteAgent deletes an agent
func (h *AgentHandler) DeleteAgent(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID is required")
	}

	if err := h.Service.DeleteAgent(c.UserContext(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Agent deleted successfully",
		Results: nil,
	})
}

// CreateIntegration creates a new integration for an agent
func (h *AgentHandler) CreateIntegration(c *fiber.Ctx) error {
	agentID := c.Params("id")
	integrationType := c.Params("type")

	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID is required")
	}

	// Validate integration type
	validTypes := map[string]bool{
		"whatsapp":  true,
		"telegram":  true,
		"instagram": true,
	}
	if !validTypes[integrationType] {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid integration type. Must be: whatsapp, telegram, or instagram")
	}

	// Check if agent exists
	_, err := h.Service.GetAgent(c.UserContext(), agentID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Agent not found")
	}

	integration, err := h.Service.GetOrCreateIntegration(c.UserContext(), agentID, integrationType)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(utils.ResponseData{
		Status:  201,
		Code:    "SUCCESS",
		Message: "Integration created successfully",
		Results: integration,
	})
}

// DeleteIntegration removes an integration
func (h *AgentHandler) DeleteIntegration(c *fiber.Ctx) error {
	integrationID := c.Params("integrationId")
	if integrationID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Integration ID is required")
	}

	// Stop Telegram bot if running
	if botMgr := telegramBot.GetBotManager(); botMgr != nil {
		botMgr.StopBot(integrationID)
	}

	if err := h.Service.DeleteIntegration(c.UserContext(), integrationID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Integration deleted successfully",
		Results: nil,
	})
}

// ConnectIntegration connects an integration (e.g., starts Telegram bot)
func (h *AgentHandler) ConnectIntegration(c *fiber.Ctx) error {
	agentID := c.Params("id")
	integrationID := c.Params("integrationId")
	
	if agentID == "" || integrationID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID and Integration ID are required")
	}

	// Get config from body
	var configBody map[string]interface{}
	if err := c.BodyParser(&configBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Get integration
	integration, err := h.Service.GetIntegration(c.UserContext(), integrationID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Integration not found")
	}

	// Handle based on integration type
	switch integration.Type {
	case agent.IntegrationTypeTelegram:
		botToken, ok := configBody["bot_token"].(string)
		if !ok || botToken == "" {
			return fiber.NewError(fiber.StatusBadRequest, "bot_token is required for Telegram")
		}

		// Update integration config
		config := agent.TelegramConfig{BotToken: botToken}
		configJSON, _ := json.Marshal(config)
		
		if err := h.Service.UpdateIntegrationConfig(c.UserContext(), integrationID, string(configJSON), true); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		// Start Telegram bot
		if botMgr := telegramBot.GetBotManager(); botMgr != nil {
			if err := botMgr.StartBot(integrationID, agentID, botToken); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to start Telegram bot: "+err.Error())
			}
		}

	case agent.IntegrationTypeWhatsApp:
		// WhatsApp connection - save device_id
		deviceID, _ := configBody["device_id"].(string)
		config := agent.WhatsAppConfig{DeviceID: deviceID}
		configJSON, _ := json.Marshal(config)
		
		if err := h.Service.UpdateIntegrationConfig(c.UserContext(), integrationID, string(configJSON), true); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

	case agent.IntegrationTypeInstagram:
		// Instagram requires OAuth - placeholder
		return fiber.NewError(fiber.StatusNotImplemented, "Instagram integration not yet implemented")
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Integration connected successfully",
		Results: nil,
	})
}

// DisconnectIntegration disconnects an integration
func (h *AgentHandler) DisconnectIntegration(c *fiber.Ctx) error {
	integrationID := c.Params("integrationId")
	if integrationID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Integration ID is required")
	}

	integration, err := h.Service.GetIntegration(c.UserContext(), integrationID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Integration not found")
	}

	// Stop services based on type
	if integration.Type == agent.IntegrationTypeTelegram {
		if botMgr := telegramBot.GetBotManager(); botMgr != nil {
			botMgr.StopBot(integrationID)
		}
	}

	// Update integration status
	if err := h.Service.UpdateIntegrationConfig(c.UserContext(), integrationID, integration.Config, false); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Integration disconnected successfully",
		Results: nil,
	})
}

// GetWhatsAppQR generates a QR code for WhatsApp login
func (h *AgentHandler) GetWhatsAppQR(c *fiber.Ctx) error {
	// Get device manager
	dm := whatsapp.GetDeviceManager()
	if dm == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Device manager not available")
	}

	// Get or create a device
	devices := dm.ListDevices()
	var deviceID string
	
	if len(devices) > 0 {
		deviceID = devices[0].ID()
	} else {
		// Create new device
		device, err := dm.CreateDevice(c.UserContext(), "")
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to create device: "+err.Error())
		}
		deviceID = device.ID()
	}

	// Generate QR using AppService
	if h.AppService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "App service not available")
	}

	response, err := h.AppService.Login(c.UserContext(), deviceID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate QR: "+err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "QR code generated",
		Results: map[string]any{
			"device_id": deviceID,
			"qr_link":   fmt.Sprintf("%s://%s%s/%s", c.Protocol(), c.Hostname(), config.AppBasePath, response.ImagePath),
			"qr_duration": response.Duration,
		},
	})
}

