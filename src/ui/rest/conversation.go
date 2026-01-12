package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/gofiber/fiber/v2"
)

type ConversationHandler struct {
	AgentService *usecase.AgentService
}

func InitRestConversation(app fiber.Router, agentService *usecase.AgentService) ConversationHandler {
	handler := ConversationHandler{AgentService: agentService}

	app.Get("/conversations", handler.GetConversations)
	app.Get("/conversations/:id/messages", handler.GetMessages)

	return handler
}

// ConversationResponse represents a conversation with additional info
type ConversationResponse struct {
	ID              string `json:"id"`
	AgentID         string `json:"agent_id"`
	AgentName       string `json:"agent_name,omitempty"`
	IntegrationID   string `json:"integration_id"`
	IntegrationType string `json:"integration_type,omitempty"`
	RemoteJID       string `json:"remote_jid"`
	LastMessage     string `json:"last_message,omitempty"`
	UnreadCount     int    `json:"unread_count"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// GetConversations returns all conversations, optionally filtered by agent
func (h *ConversationHandler) GetConversations(c *fiber.Ctx) error {
	// For now, return empty list - would need to implement in agent repository
	// This is a placeholder for the Live Chat feature
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Conversations retrieved",
		Results: []ConversationResponse{},
	})
}

// GetMessages returns messages for a conversation
func (h *ConversationHandler) GetMessages(c *fiber.Ctx) error {
	// conversationID := c.Params("id")
	
	// For now, return empty list - would need to implement in agent repository
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Messages retrieved",
		Results: []interface{}{},
	})
}

