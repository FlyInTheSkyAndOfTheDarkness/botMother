package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/gofiber/fiber/v2"
)

type AnalyticsHandler struct {
	Service *usecase.AnalyticsService
}

func InitRestAnalytics(app fiber.Router, service *usecase.AnalyticsService) AnalyticsHandler {
	handler := AnalyticsHandler{Service: service}

	app.Get("/analytics/dashboard", handler.GetDashboard)
	app.Get("/analytics/agents", handler.GetAllAgentsStats)
	app.Get("/analytics/agents/:id", handler.GetAgentStats)
	app.Get("/analytics/messages/hourly", handler.GetMessagesByHour)
	app.Get("/analytics/messages/daily", handler.GetMessagesByDay)
	app.Get("/analytics/activity", handler.GetRecentActivity)

	return handler
}

// GetDashboard returns overall dashboard statistics
func (h *AnalyticsHandler) GetDashboard(c *fiber.Ctx) error {
	stats, err := h.Service.GetDashboard(c.UserContext())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Dashboard stats retrieved",
		Results: stats,
	})
}

// GetAllAgentsStats returns statistics for all agents
func (h *AnalyticsHandler) GetAllAgentsStats(c *fiber.Ctx) error {
	period := c.Query("period", "7days")

	stats, err := h.Service.GetAllAgentsAnalytics(c.UserContext(), period)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Agent stats retrieved",
		Results: stats,
	})
}

// GetAgentStats returns statistics for a single agent
func (h *AnalyticsHandler) GetAgentStats(c *fiber.Ctx) error {
	agentID := c.Params("id")
	period := c.Query("period", "7days")

	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID is required")
	}

	stats, err := h.Service.GetAgentAnalytics(c.UserContext(), agentID, period)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Agent stats retrieved",
		Results: stats,
	})
}

// GetMessagesByHour returns message counts grouped by hour
func (h *AnalyticsHandler) GetMessagesByHour(c *fiber.Ctx) error {
	period := c.Query("period", "today")

	data, err := h.Service.GetMessagesByHour(c.UserContext(), period)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Hourly message stats retrieved",
		Results: data,
	})
}

// GetMessagesByDay returns message counts grouped by day
func (h *AnalyticsHandler) GetMessagesByDay(c *fiber.Ctx) error {
	period := c.Query("period", "30days")

	data, err := h.Service.GetMessagesByDay(c.UserContext(), period)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Daily message stats retrieved",
		Results: data,
	})
}

// GetRecentActivity returns recent activity items
func (h *AnalyticsHandler) GetRecentActivity(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)

	data, err := h.Service.GetRecentActivity(c.UserContext(), limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Recent activity retrieved",
		Results: data,
	})
}


