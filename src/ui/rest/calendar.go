package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/calendar"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

type CalendarHandler struct {
	Repo calendar.ICalendarRepository
}

func InitRestCalendar(app fiber.Router, repo calendar.ICalendarRepository) CalendarHandler {
	handler := CalendarHandler{Repo: repo}

	app.Get("/agents/:agentId/calendar", handler.GetCredential)
	app.Post("/agents/:agentId/calendar", handler.SaveCredential)
	app.Delete("/agents/:agentId/calendar", handler.DeleteCredential)

	return handler
}

// GetCredential returns calendar credential for an agent
func (h *CalendarHandler) GetCredential(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID required")
	}

	cred, err := h.Repo.GetCredentialByAgentID(c.UserContext(), agentID)
	if err != nil {
		// Return empty credential if not found
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "No calendar connected",
			Results: nil,
		})
	}

	// Mask sensitive fields
	cred.ClientSecret = "***"
	cred.AccessToken = ""
	cred.RefreshToken = ""

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Calendar credential retrieved",
		Results: cred,
	})
}

// SaveCredential creates or updates calendar credential
func (h *CalendarHandler) SaveCredential(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID required")
	}

	var cred calendar.CalendarCredential
	if err := c.BodyParser(&cred); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	cred.AgentID = agentID

	// Check if credential exists
	existing, err := h.Repo.GetCredentialByAgentID(c.UserContext(), agentID)
	if err == nil && existing != nil {
		// Update existing
		cred.ID = existing.ID
		if cred.ClientSecret == "***" {
			cred.ClientSecret = existing.ClientSecret
		}
		cred.AccessToken = existing.AccessToken
		cred.RefreshToken = existing.RefreshToken
		cred.TokenExpiry = existing.TokenExpiry
		if err := h.Repo.UpdateCredential(c.UserContext(), &cred); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	} else {
		// Create new
		if err := h.Repo.CreateCredential(c.UserContext(), &cred); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Calendar credential saved",
	})
}

// DeleteCredential removes calendar credential
func (h *CalendarHandler) DeleteCredential(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID required")
	}

	cred, err := h.Repo.GetCredentialByAgentID(c.UserContext(), agentID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Calendar not connected")
	}

	if err := h.Repo.DeleteCredential(c.UserContext(), cred.ID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Calendar disconnected",
	})
}

