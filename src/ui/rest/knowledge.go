package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/knowledge"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/gofiber/fiber/v2"
)

type KnowledgeHandler struct {
	Service *usecase.KnowledgeService
}

func InitRestKnowledge(app fiber.Router, service *usecase.KnowledgeService) KnowledgeHandler {
	handler := KnowledgeHandler{Service: service}

	app.Get("/agents/:agentId/knowledge", handler.GetDocuments)
	app.Post("/agents/:agentId/knowledge", handler.UploadDocument)
	app.Delete("/knowledge/:id", handler.DeleteDocument)
	app.Post("/agents/:agentId/knowledge/search", handler.Search)

	return handler
}

// GetDocuments returns all documents for an agent
func (h *KnowledgeHandler) GetDocuments(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID required")
	}

	docs, err := h.Service.GetDocuments(c.UserContext(), agentID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Documents retrieved",
		Results: docs,
	})
}

// UploadDocument creates a new document
func (h *KnowledgeHandler) UploadDocument(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID required")
	}

	var req knowledge.CreateDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	req.AgentID = agentID

	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Document name required")
	}

	if req.Content == "" && req.URL == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Content or URL required")
	}

	doc, err := h.Service.UploadDocument(c.UserContext(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  201,
		Code:    "SUCCESS",
		Message: "Document uploaded and processing",
		Results: doc,
	})
}

// DeleteDocument removes a document
func (h *KnowledgeHandler) DeleteDocument(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Document ID required")
	}

	if err := h.Service.DeleteDocument(c.UserContext(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Document deleted",
	})
}

// Search performs a RAG search
func (h *KnowledgeHandler) Search(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Agent ID required")
	}

	var req knowledge.SearchRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	req.AgentID = agentID

	if req.Query == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Query required")
	}

	results, err := h.Service.Search(c.UserContext(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Search completed",
		Results: results,
	})
}

