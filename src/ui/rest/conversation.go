package rest

import (
	"encoding/csv"
	"fmt"
	"strings"

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
	app.Get("/conversations/:id", handler.GetConversation)
	app.Get("/conversations/:id/messages", handler.GetMessages)
	app.Post("/conversations/:id/messages", handler.SendMessage)
	app.Post("/conversations/:id/takeover", handler.TakeOver)
	app.Post("/conversations/:id/release", handler.Release)
	app.Post("/conversations/:id/notes", handler.AddNote)
	app.Get("/conversations/:id/export", handler.ExportChat)

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
	IsManualMode    bool   `json:"is_manual_mode"`
	Notes           string `json:"notes,omitempty"`
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
			IsManualMode:    conv.IsManualMode,
			Notes:           conv.Notes,
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

// GetConversation returns a single conversation
func (h *ConversationHandler) GetConversation(c *fiber.Ctx) error {
	conversationID := c.Params("id")
	if conversationID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Conversation ID is required")
	}

	conv, err := h.AgentService.GetConversation(c.UserContext(), conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	// Get agent name
	agentName := ""
	if agent, err := h.AgentService.GetAgent(c.UserContext(), conv.AgentID); err == nil {
		agentName = agent.Name
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Conversation retrieved",
		Results: ConversationResponse{
			ID:            conv.ID,
			AgentID:       conv.AgentID,
			AgentName:     agentName,
			IntegrationID: conv.IntegrationID,
			RemoteJID:     conv.RemoteJID,
			IsManualMode:  conv.IsManualMode,
			Notes:         conv.Notes,
			CreatedAt:     conv.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     conv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
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

// SendMessage sends a manual message (from manager)
func (h *ConversationHandler) SendMessage(c *fiber.Ctx) error {
	conversationID := c.Params("id")
	if conversationID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Conversation ID is required")
	}

	var req struct {
		Content  string `json:"content"`
		IsManual bool   `json:"is_manual"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Content == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Content is required")
	}

	// Add message to conversation
	msg, err := h.AgentService.AddManualMessage(c.UserContext(), conversationID, req.Content)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// TODO: Actually send the message via WhatsApp/Telegram
	// This requires integration with the messaging platform

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Message sent",
		Results: MessageResponse{
			ID:        msg.ID,
			Role:      msg.Role,
			Content:   msg.Content,
			Timestamp: msg.Timestamp.Format("2006-01-02T15:04:05Z"),
		},
	})
}

// TakeOver puts conversation in manual mode (pauses AI)
func (h *ConversationHandler) TakeOver(c *fiber.Ctx) error {
	conversationID := c.Params("id")
	if conversationID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Conversation ID is required")
	}

	if err := h.AgentService.SetConversationManualMode(c.UserContext(), conversationID, true); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Conversation taken over - AI paused",
		Results: nil,
	})
}

// Release puts conversation back to AI mode
func (h *ConversationHandler) Release(c *fiber.Ctx) error {
	conversationID := c.Params("id")
	if conversationID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Conversation ID is required")
	}

	if err := h.AgentService.SetConversationManualMode(c.UserContext(), conversationID, false); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Conversation released - AI active",
		Results: nil,
	})
}

// AddNote adds or updates notes for a conversation
func (h *ConversationHandler) AddNote(c *fiber.Ctx) error {
	conversationID := c.Params("id")
	if conversationID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Conversation ID is required")
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.AgentService.UpdateConversationNotes(c.UserContext(), conversationID, req.Notes); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Notes updated",
		Results: nil,
	})
}

// ExportChat exports conversation as CSV
func (h *ConversationHandler) ExportChat(c *fiber.Ctx) error {
	conversationID := c.Params("id")
	if conversationID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Conversation ID is required")
	}

	// Get conversation
	conv, err := h.AgentService.GetConversation(c.UserContext(), conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	// Get messages
	messages, err := h.AgentService.GetMessagesForConversation(c.UserContext(), conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Generate CSV
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	
	// Header
	writer.Write([]string{"Timestamp", "Role", "Content"})
	
	// Data
	for _, msg := range messages {
		writer.Write([]string{
			msg.Timestamp.Format("2006-01-02 15:04:05"),
			msg.Role,
			msg.Content,
		})
	}
	writer.Flush()

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=chat_%s_%s.csv", conv.RemoteJID, conv.ID[:8]))
	
	return c.SendString(builder.String())
}

