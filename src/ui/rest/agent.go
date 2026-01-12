package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/agent"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/gofiber/fiber/v2"
)

type AgentHandler struct {
	Service *usecase.AgentService
}

func InitRestAgent(app fiber.Router, service *usecase.AgentService) AgentHandler {
	handler := AgentHandler{Service: service}

	// Agent CRUD
	app.Get("/agents", handler.GetAllAgents)
	app.Post("/agents", handler.CreateAgent)
	app.Get("/agents/:id", handler.GetAgent)
	app.Put("/agents/:id", handler.UpdateAgent)
	app.Delete("/agents/:id", handler.DeleteAgent)

	// Integration endpoints
	app.Post("/agents/:id/integrations/:type", handler.CreateIntegration)
	app.Delete("/agents/:id/integrations/:integrationId", handler.DeleteIntegration)

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

