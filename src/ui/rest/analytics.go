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
	app.Get("/analytics/messages/daily", handler.GetMessagesDaily)
	app.Get("/analytics/activity", handler.GetRecentActivity)
	app.Get("/analytics/agents", handler.GetAgentStats)

	return handler
}

// GetDashboard returns overall dashboard statistics
func (h *AnalyticsHandler) GetDashboard(c *fiber.Ctx) error {
	stats, err := h.Service.GetDashboardStats(c.UserContext())
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

// GetMessagesDaily returns message counts grouped by time period
func (h *AnalyticsHandler) GetMessagesDaily(c *fiber.Ctx) error {
	period := c.Query("period", "7days")
	if period != "today" && period != "7days" && period != "30days" && period != "month" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid period. Must be: today, 7days, 30days, or month")
	}

	data, err := h.Service.GetMessagesDaily(c.UserContext(), period)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Messages daily stats retrieved",
		Results: data,
	})
}

// GetRecentActivity returns recent activity
func (h *AnalyticsHandler) GetRecentActivity(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	activity, err := h.Service.GetRecentActivity(c.UserContext(), limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Recent activity retrieved",
		Results: activity,
	})
}

// GetAgentStats returns statistics per agent
func (h *AnalyticsHandler) GetAgentStats(c *fiber.Ctx) error {
	period := c.Query("period", "7days")
	if period != "today" && period != "7days" && period != "30days" && period != "month" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid period. Must be: today, 7days, 30days, or month")
	}

	stats, err := h.Service.GetAgentStats(c.UserContext(), period)
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
