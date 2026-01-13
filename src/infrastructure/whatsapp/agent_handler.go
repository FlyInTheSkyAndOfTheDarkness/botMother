package whatsapp

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/agent"
	agentRepo "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/agent"
	aiService "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/ai"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// AgentMessageHandler handles incoming messages for agents with WhatsApp integrations
type AgentMessageHandler struct {
	agentRepo *agentRepo.SQLiteRepository
	mu        sync.RWMutex
}

var (
	agentHandler     *AgentMessageHandler
	agentHandlerOnce sync.Once
)

// InitAgentHandler initializes the global agent message handler
func InitAgentHandler(repo *agentRepo.SQLiteRepository) {
	agentHandlerOnce.Do(func() {
		if repo != nil {
			agentHandler = &AgentMessageHandler{
				agentRepo: repo,
			}
			logrus.Infof("Agent message handler initialized")
		}
	})
}

// GetAgentHandler returns the global agent handler
func GetAgentHandler() *AgentMessageHandler {
	return agentHandler
}

// HandleIncomingMessage processes incoming WhatsApp messages for all active agents
func (h *AgentMessageHandler) HandleIncomingMessage(
	ctx context.Context,
	evt *events.Message,
	chatStorageRepo domainChatStorage.IChatStorageRepository,
	client *whatsmeow.Client,
) {
	if h == nil || h.agentRepo == nil || client == nil {
		return
	}

	// Skip groups, broadcasts, and self messages
	if utils.IsGroupJID(evt.Info.Chat.String()) || evt.Info.IsIncomingBroadcast() || evt.Info.IsFromMe {
		return
	}

	// Only reply to direct 1:1 chats
	if evt.Info.Chat.Server != types.DefaultUserServer {
		return
	}

	// Extra safety: skip any broadcast/status contexts
	source := evt.Info.SourceString()
	if strings.Contains(source, "broadcast") ||
		strings.HasSuffix(evt.Info.Chat.String(), "@broadcast") ||
		strings.HasPrefix(evt.Info.Chat.String(), "status@") {
		return
	}

	// Extract user message
	userMessage := h.extractMessage(ctx, evt, client)
	if userMessage == "" {
		return
	}

	// Get current device phone number to find matching integration
	deviceJID := ""
	if client.Store.ID != nil {
		deviceJID = client.Store.ID.String()
	}

	// Find agents with active WhatsApp integrations matching this device
	agents, err := h.agentRepo.GetAll(ctx)
	if err != nil {
		logrus.Errorf("❌ [WhatsApp Agent] Failed to get agents: %v", err)
		return
	}

	remoteJID := evt.Info.Sender.String()
	logrus.Infof("📱 [WhatsApp Agent] Processing message from %s (device: %s), found %d agents", remoteJID, deviceJID, len(agents))

	foundAgent := false
	for _, ag := range agents {
		if !ag.IsActive {
			logrus.Debugf("⏭️  [WhatsApp Agent] Skipping agent %s (%s) - not active", ag.ID, ag.Name)
			continue
		}

		logrus.Infof("✅ [WhatsApp Agent] Checking agent %s (%s) - is active", ag.ID, ag.Name)

		// Get integrations for this agent
		integrations, err := h.agentRepo.GetIntegrationsByAgentID(ctx, ag.ID)
		if err != nil {
			logrus.Warnf("⚠️  [WhatsApp Agent] Failed to get integrations for agent %s: %v", ag.ID, err)
			continue
		}

		logrus.Debugf("📋 [WhatsApp Agent] Agent %s has %d integrations", ag.ID, len(integrations))

		// Find WhatsApp integration that matches this device
		var matchingIntegration *agent.Integration
		for _, integration := range integrations {
			if integration.Type != agent.IntegrationTypeWhatsApp {
				continue
			}
			
			if !integration.IsConnected {
				logrus.Debugf("⏭️  [WhatsApp Agent] Integration %s not connected", integration.ID)
				continue
			}

			logrus.Infof("🔍 [WhatsApp Agent] Checking WhatsApp integration %s (connected: %v)", integration.ID, integration.IsConnected)
			
			// For now, use first connected WhatsApp integration
			// TODO: Match by device JID stored in config
			waConfig, _ := agentRepo.ParseWhatsAppConfig(integration.Config)
			if waConfig != nil {
				logrus.Debugf("📱 [WhatsApp Agent] Integration config: device_id=%s, current_device=%s", waConfig.DeviceID, deviceJID)
				if waConfig.DeviceID == deviceJID {
					matchingIntegration = integration
					logrus.Infof("✅ [WhatsApp Agent] Found matching integration %s for device %s", integration.ID, deviceJID)
					break
				}
			}
			// If no specific device configured, use any connected integration
			if matchingIntegration == nil {
				matchingIntegration = integration
				logrus.Infof("✅ [WhatsApp Agent] Using first connected integration %s (no device match required)", integration.ID)
			}
		}

		if matchingIntegration == nil {
			logrus.Debugf("⏭️  [WhatsApp Agent] No matching WhatsApp integration for agent %s", ag.ID)
			continue
		}

		foundAgent = true
		logrus.Infof("🚀 [WhatsApp Agent] Processing message for agent %s (%s) with integration %s", ag.ID, ag.Name, matchingIntegration.ID)
		// Process message for this agent
		go h.processMessageForAgent(ctx, ag, matchingIntegration, remoteJID, userMessage, chatStorageRepo, client)
	}

	if !foundAgent {
		logrus.Warnf("⚠️  [WhatsApp Agent] No active agent with connected WhatsApp integration found for message from %s", remoteJID)
	}
}

func (h *AgentMessageHandler) extractMessage(ctx context.Context, evt *events.Message, client *whatsmeow.Client) string {
	// Unwrap FutureProof wrappers
	innerMsg := evt.Message
	for i := 0; i < 3; i++ {
		if vm := innerMsg.GetViewOnceMessage(); vm != nil && vm.GetMessage() != nil {
			innerMsg = vm.GetMessage()
			continue
		}
		if em := innerMsg.GetEphemeralMessage(); em != nil && em.GetMessage() != nil {
			innerMsg = em.GetMessage()
			continue
		}
		if vm2 := innerMsg.GetViewOnceMessageV2(); vm2 != nil && vm2.GetMessage() != nil {
			innerMsg = vm2.GetMessage()
			continue
		}
		if vm2e := innerMsg.GetViewOnceMessageV2Extension(); vm2e != nil && vm2e.GetMessage() != nil {
			innerMsg = vm2e.GetMessage()
			continue
		}
		break
	}

	// Extract text from message
	if conv := innerMsg.GetConversation(); conv != "" {
		return conv
	} else if ext := innerMsg.GetExtendedTextMessage(); ext != nil && ext.GetText() != "" {
		return ext.GetText()
	} else if protoMsg := innerMsg.GetProtocolMessage(); protoMsg != nil {
		if edited := protoMsg.GetEditedMessage(); edited != nil {
			if ext := edited.GetExtendedTextMessage(); ext != nil && ext.GetText() != "" {
				return ext.GetText()
			} else if conv := edited.GetConversation(); conv != "" {
				return conv
			}
		}
	} else if audioMsg := innerMsg.GetAudioMessage(); audioMsg != nil {
		// Handle audio messages - transcribe them
		if !config.WhatsappAutoDownloadMedia {
			logrus.Warnf("Agent received audio message but auto-download-media is disabled")
			return ""
		}

		extractedMedia, err := utils.ExtractMedia(ctx, client, config.PathMedia, audioMsg)
		if err != nil {
			logrus.Errorf("Failed to download audio for transcription: %v", err)
			return ""
		}

		// We'll transcribe in processMessageForAgent using the agent's API key
		return "[AUDIO:" + extractedMedia.MediaPath + "]"
	}

	return ""
}

func (h *AgentMessageHandler) processMessageForAgent(
	ctx context.Context,
	ag *agent.Agent,
	integration *agent.Integration,
	remoteJID string,
	userMessage string,
	chatStorageRepo domainChatStorage.IChatStorageRepository,
	client *whatsmeow.Client,
) {
	logrus.Infof("🤖 [WhatsApp Agent] Processing message for agent %s (%s): %s", ag.ID, ag.Name, userMessage[:min(50, len(userMessage))])
	
	// Create AI service with agent's API key
	aiSvc := aiService.NewService(ag.APIKey, ag.SerpAPIKey)
	if aiSvc == nil {
		logrus.Errorf("❌ [WhatsApp Agent] Failed to create AI service for agent %s (API key: %v)", ag.ID, ag.APIKey != "")
		return
	}
	logrus.Debugf("✅ [WhatsApp Agent] AI service created for agent %s", ag.ID)

	// Handle audio transcription if needed
	if strings.HasPrefix(userMessage, "[AUDIO:") && strings.HasSuffix(userMessage, "]") {
		audioPath := userMessage[7 : len(userMessage)-1]
		transcription, err := aiSvc.TranscribeAudio(ctx, audioPath)
		if err != nil {
			logrus.Errorf("Failed to transcribe audio for agent %s: %v", ag.ID, err)
			return
		}
		userMessage = transcription
		logrus.Infof("Agent %s transcribed audio: %s", ag.ID, userMessage)
	}

	// Get or create conversation
	conv, err := h.agentRepo.GetOrCreateConversation(ctx, ag.ID, integration.ID, remoteJID)
	if err != nil {
		logrus.Errorf("❌ [WhatsApp Agent] Failed to get conversation for agent %s: %v", ag.ID, err)
		return
	}
	logrus.Debugf("💬 [WhatsApp Agent] Conversation %s found/created for agent %s", conv.ID, ag.ID)

	// Check if in manual mode (manager took over)
	if conv.IsManualMode {
		// Store user message but don't generate AI response
		userMsg := &agent.Message{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        userMessage,
		}
		h.agentRepo.AddMessage(ctx, userMsg)
		logrus.Infof("⏸️  [WhatsApp Agent] Conversation %s is in manual mode, skipping AI response", conv.ID)
		return
	}

	// Store user message
	userMsg := &agent.Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        userMessage,
	}
	if err := h.agentRepo.AddMessage(ctx, userMsg); err != nil {
		logrus.Errorf("Failed to store user message: %v", err)
	}

	var response string

	// Check if this is the first reply - send welcome message
	if !conv.IsFirstReply && ag.WelcomeMessage != "" {
		response = ag.WelcomeMessage
		conv.IsFirstReply = true
		if err := h.agentRepo.UpdateConversation(ctx, conv); err != nil {
			logrus.Errorf("Failed to update conversation: %v", err)
		}
	} else {
		// Get recent messages for context (limited to 5 for efficiency)
		recentMessages, err := h.agentRepo.GetRecentMessages(ctx, conv.ID, 5)
		if err != nil {
			logrus.Errorf("Failed to get recent messages: %v", err)
		}

		// Build context from recent messages
		var contextBuilder strings.Builder
		for _, msg := range recentMessages {
			if msg.Role == "user" {
				contextBuilder.WriteString(fmt.Sprintf("User: %s\n", msg.Content))
			} else if msg.Role == "assistant" {
				contextBuilder.WriteString(fmt.Sprintf("Assistant: %s\n", msg.Content))
			}
		}

		// If we have context, include it in the prompt
		finalPrompt := userMessage
		if contextBuilder.Len() > 0 {
			finalPrompt = fmt.Sprintf("Previous conversation:\n%s\nCurrent message: %s", contextBuilder.String(), userMessage)
		}

	response, err = aiSvc.GenerateResponse(ctx, finalPrompt, ag.SystemPrompt, ag.Model)
	if err != nil {
		logrus.Errorf("❌ [WhatsApp Agent] Failed to generate AI response for agent %s: %v", ag.ID, err)
		return
	}

	// Mark conversation as having had first reply
	if !conv.IsFirstReply {
		conv.IsFirstReply = true
		h.agentRepo.UpdateConversation(ctx, conv)
	}
}

if response == "" {
	logrus.Warnf("⚠️  [WhatsApp Agent] Agent %s: AI returned empty response", ag.ID)
	return
}

logrus.Infof("💡 [WhatsApp Agent] AI response generated for agent %s: %s", ag.ID, response[:min(100, len(response))])

	// Store assistant response
	assistantMsg := &agent.Message{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        response,
	}
	if err := h.agentRepo.AddMessage(ctx, assistantMsg); err != nil {
		logrus.Errorf("Failed to store assistant message: %v", err)
	}

	// Send the response via WhatsApp
	recipientJID := utils.FormatJID(remoteJID)
	logrus.Infof("📤 [WhatsApp Agent] Sending response to %s via agent %s", recipientJID, ag.ID)
	sendResp, err := client.SendMessage(
		ctx,
		recipientJID,
		&waE2E.Message{Conversation: proto.String(response)},
	)

	if err != nil {
		logrus.Errorf("❌ [WhatsApp Agent] Failed to send agent %s response to %s: %v", ag.ID, recipientJID, err)
		return
	}
	logrus.Infof("✅ [WhatsApp Agent] Response sent successfully (message ID: %s)", sendResp.ID)

	// Store sent message in chat storage
	if chatStorageRepo != nil {
		senderJID := ""
		if client.Store.ID != nil {
			senderJID = client.Store.ID.String()
		}

		if err := chatStorageRepo.StoreSentMessageWithContext(
			ctx,
			sendResp.ID,
			senderJID,
			recipientJID.String(),
			response,
			sendResp.Timestamp,
		); err != nil {
			logrus.Errorf("Failed to store agent response in chat storage: %v", err)
		}
	}

	logrus.Infof("✅ [WhatsApp Agent] Agent %s successfully processed and sent response to %s", ag.Name, remoteJID)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

