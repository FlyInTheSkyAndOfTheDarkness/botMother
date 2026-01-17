package rest

import (
	"encoding/json"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/flow"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/gofiber/fiber/v2"
)

type FlowHandler struct {
	Service *usecase.FlowService
}

func InitRestFlow(app fiber.Router, service *usecase.FlowService) FlowHandler {
	handler := FlowHandler{Service: service}

	// Flow CRUD
	app.Get("/flows", handler.GetAllFlows)
	app.Post("/flows", handler.CreateFlow)
	app.Get("/flows/:id", handler.GetFlow)
	app.Put("/flows/:id", handler.UpdateFlow)
	app.Delete("/flows/:id", handler.DeleteFlow)

	// Credential CRUD
	app.Get("/credentials", handler.GetAllCredentials)
	app.Post("/credentials", handler.CreateCredential)
	app.Get("/credentials/:id", handler.GetCredential)
	app.Put("/credentials/:id", handler.UpdateCredential)
	app.Delete("/credentials/:id", handler.DeleteCredential)
	app.Post("/credentials/test-database", handler.TestDatabaseConnection)

	return handler
}

// === Flow Handlers ===

// GetAllFlows returns all flows for an agent
func (h *FlowHandler) GetAllFlows(c *fiber.Ctx) error {
	agentID := c.Query("agent_id")
	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "agent_id query parameter is required")
	}

	flows, err := h.Service.GetFlowsByAgent(c.UserContext(), agentID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Flows retrieved successfully",
		Results: flows,
	})
}

// CreateFlow creates a new flow
func (h *FlowHandler) CreateFlow(c *fiber.Ctx) error {
	var req flow.CreateFlowRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if req.AgentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "agent_id is required")
	}
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}

	result, err := h.Service.CreateFlow(c.UserContext(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(utils.ResponseData{
		Status:  201,
		Code:    "SUCCESS",
		Message: "Flow created successfully",
		Results: result,
	})
}

// GetFlow returns a single flow by ID
func (h *FlowHandler) GetFlow(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Flow ID is required")
	}

	result, err := h.Service.GetFlow(c.UserContext(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Flow retrieved successfully",
		Results: result,
	})
}

// UpdateFlow updates a flow
func (h *FlowHandler) UpdateFlow(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Flow ID is required")
	}

	var req flow.UpdateFlowRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	result, err := h.Service.UpdateFlow(c.UserContext(), id, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Flow updated successfully",
		Results: result,
	})
}

// DeleteFlow deletes a flow
func (h *FlowHandler) DeleteFlow(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Flow ID is required")
	}

	if err := h.Service.DeleteFlow(c.UserContext(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Flow deleted successfully",
		Results: nil,
	})
}

// === Credential Handlers ===

// GetAllCredentials returns all credentials for an agent
func (h *FlowHandler) GetAllCredentials(c *fiber.Ctx) error {
	agentID := c.Query("agent_id")
	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "agent_id query parameter is required")
	}

	credentials, err := h.Service.GetCredentialsByAgent(c.UserContext(), agentID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Credentials retrieved successfully",
		Results: credentials,
	})
}

// CreateCredential creates a new credential
func (h *FlowHandler) CreateCredential(c *fiber.Ctx) error {
	var req flow.CreateCredentialRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if req.AgentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "agent_id is required")
	}
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if req.Type == "" {
		return fiber.NewError(fiber.StatusBadRequest, "type is required")
	}
	if req.Config == "" {
		return fiber.NewError(fiber.StatusBadRequest, "config is required")
	}

	result, err := h.Service.CreateCredential(c.UserContext(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(utils.ResponseData{
		Status:  201,
		Code:    "SUCCESS",
		Message: "Credential created successfully",
		Results: result,
	})
}

// GetCredential returns a single credential by ID
func (h *FlowHandler) GetCredential(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Credential ID is required")
	}

	result, err := h.Service.GetCredential(c.UserContext(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Credential retrieved successfully",
		Results: result,
	})
}

// UpdateCredential updates a credential
func (h *FlowHandler) UpdateCredential(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Credential ID is required")
	}

	var req flow.UpdateCredentialRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	result, err := h.Service.UpdateCredential(c.UserContext(), id, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Credential updated successfully",
		Results: result,
	})
}

// DeleteCredential deletes a credential
func (h *FlowHandler) DeleteCredential(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Credential ID is required")
	}

	if err := h.Service.DeleteCredential(c.UserContext(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Credential deleted successfully",
		Results: nil,
	})
}

// TestDatabaseConnection tests a database connection
func (h *FlowHandler) TestDatabaseConnection(c *fiber.Ctx) error {
	var config flow.DatabaseCredential
	if err := c.BodyParser(&config); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if config.Host == "" {
		return fiber.NewError(fiber.StatusBadRequest, "host is required")
	}
	if config.Port == 0 {
		config.Port = 5432
	}
	if config.Database == "" {
		return fiber.NewError(fiber.StatusBadRequest, "database is required")
	}
	if config.User == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user is required")
	}
	if config.SSLMode == "" {
		config.SSLMode = "require"
	}

	if err := h.Service.TestDatabaseConnection(c.UserContext(), config); err != nil {
		return c.JSON(utils.ResponseData{
			Status:  400,
			Code:    "CONNECTION_FAILED",
			Message: err.Error(),
			Results: nil,
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Database connection successful",
		Results: map[string]interface{}{
			"connected": true,
		},
	})
}

// Helper to get credential config as JSON for specific node types
func (h *FlowHandler) GetCredentialConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Credential ID is required")
	}

	cred, err := h.Service.GetCredentialFull(c.UserContext(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	// Parse and return config based on type
	var config interface{}
	switch cred.Type {
	case flow.CredentialTypeDatabase:
		var dbConfig flow.DatabaseCredential
		if err := json.Unmarshal([]byte(cred.Config), &dbConfig); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse config")
		}
		// Mask password
		dbConfig.Password = "********"
		config = dbConfig
	default:
		config = map[string]string{"type": cred.Type}
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Credential config retrieved",
		Results: config,
	})
}




