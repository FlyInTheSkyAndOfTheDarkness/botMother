package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/settings"
	broadcastWorker "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/gofiber/fiber/v2"
)

type SettingsHandler struct {
	Service      *usecase.SettingsService
	BroadcastWorker *broadcastWorker.BroadcastWorker
}

func InitRestSettings(app fiber.Router, service *usecase.SettingsService, worker *broadcastWorker.BroadcastWorker) SettingsHandler {
	handler := SettingsHandler{Service: service, BroadcastWorker: worker}

	app.Get("/agents/:agentId/settings", handler.GetAgentSettings)
	app.Put("/agents/:agentId/settings", handler.UpdateAgentSettings)
	
	// Broadcast endpoints
	app.Get("/agents/:agentId/broadcasts", handler.GetBroadcasts)
	app.Post("/agents/:agentId/broadcasts", handler.CreateBroadcast)
	app.Post("/broadcasts/:id/send", handler.SendBroadcast)

	return handler
}

// GetAgentSettings returns settings for an agent
func (h *SettingsHandler) GetAgentSettings(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID required")
	}

	agentSettings, err := h.Service.GetAgentSettings(c.UserContext(), agentID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Settings retrieved",
		Results: agentSettings,
	})
}

// UpdateAgentSettings updates settings for an agent
func (h *SettingsHandler) UpdateAgentSettings(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID required")
	}

	var agentSettings settings.AgentSettings
	if err := c.BodyParser(&agentSettings); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	agentSettings.AgentID = agentID

	if err := h.Service.UpdateAgentSettings(c.UserContext(), &agentSettings); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Settings updated",
	})
}

// GetBroadcasts returns all broadcasts for an agent
func (h *SettingsHandler) GetBroadcasts(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID required")
	}

	broadcasts, err := h.Service.GetBroadcasts(c.UserContext(), agentID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Broadcasts retrieved",
		Results: broadcasts,
	})
}

// CreateBroadcast creates a new broadcast
func (h *SettingsHandler) CreateBroadcast(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID required")
	}

	var req settings.CreateBroadcastRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	req.AgentID = agentID

	if req.Message == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Message required")
	}
	if len(req.Recipients) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "At least one recipient required")
	}

	broadcast, err := h.Service.CreateBroadcast(c.UserContext(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  201,
		Code:    "SUCCESS",
		Message: "Broadcast created",
		Results: broadcast,
	})
}

// SendBroadcast starts sending a broadcast
func (h *SettingsHandler) SendBroadcast(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Broadcast ID required")
	}

	// Get broadcast
	broadcast, err := h.Service.GetBroadcast(c.UserContext(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Broadcast not found")
	}

	if broadcast.Status != "pending" {
		return fiber.NewError(fiber.StatusBadRequest, "Broadcast already started or completed")
	}

	// Update status to sending
	if err := h.Service.UpdateBroadcastStatus(c.UserContext(), id, "sending", 0, 0); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Send messages in background using worker
	if h.BroadcastWorker == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Broadcast worker not initialized")
	}
	
	if err := h.BroadcastWorker.ExecuteBroadcast(c.UserContext(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to start broadcast: "+err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Broadcast started",
	})
}


