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

// MessageResponse represents a message
type MessageResponse struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// GetConversations returns all conversations
func (h *ConversationHandler) GetConversations(c *fiber.Ctx) error {
	conversations, err := h.AgentService.GetAllConversations(c.UserContext())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var results []ConversationResponse
	for _, conv := range conversations {
		// Get last message
		lastMsg, _ := h.AgentService.GetLastMessage(c.UserContext(), conv.ID)
		lastMsgContent := ""
		if lastMsg != nil {
			lastMsgContent = lastMsg.Content
			if len(lastMsgContent) > 50 {
				lastMsgContent = lastMsgContent[:50] + "..."
			}
		}

		// Get agent name
		agentName := ""
		if agent, err := h.AgentService.GetAgent(c.UserContext(), conv.AgentID); err == nil {
			agentName = agent.Name
		}

		// Get integration type
		integrationType := ""
		if integration, err := h.AgentService.GetIntegration(c.UserContext(), conv.IntegrationID); err == nil {
			integrationType = integration.Type
		}

		results = append(results, ConversationResponse{
			ID:              conv.ID,
			AgentID:         conv.AgentID,
			AgentName:       agentName,
			IntegrationID:   conv.IntegrationID,
			IntegrationType: integrationType,
			RemoteJID:       conv.RemoteJID,
			LastMessage:     lastMsgContent,
			UnreadCount:     0,
			CreatedAt:       conv.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:       conv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Conversations retrieved",
		Results: results,
	})
}

// GetMessages returns messages for a conversation
func (h *ConversationHandler) GetMessages(c *fiber.Ctx) error {
	conversationID := c.Params("id")
	if conversationID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Conversation ID is required")
	}

	messages, err := h.AgentService.GetMessagesForConversation(c.UserContext(), conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var results []MessageResponse
	for _, msg := range messages {
		results = append(results, MessageResponse{
			ID:        msg.ID,
			Role:      msg.Role,
			Content:   msg.Content,
			Timestamp: msg.Timestamp.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Messages retrieved",
		Results: results,
	})
}

