package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	Service *usecase.HealthService
}

func InitRestHealth(app fiber.Router, service *usecase.HealthService) HealthHandler {
	handler := HealthHandler{Service: service}

	app.Get("/health", handler.GetSystemHealth)
	app.Get("/health/agents/:agentId", handler.GetAgentHealth)

	return handler
}

// GetSystemHealth returns overall system health status
func (h *HealthHandler) GetSystemHealth(c *fiber.Ctx) error {
	health, err := h.Service.GetSystemHealth(c.UserContext())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "System health retrieved",
		Results: health,
	})
}

// GetAgentHealth returns health status for a specific agent
func (h *HealthHandler) GetAgentHealth(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID required")
	}

	health, err := h.Service.GetAgentHealth(c.UserContext(), agentID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Agent health retrieved",
		Results: health,
	})
}

