package whatsapp

import (
	"context"
	"strings"
	"sync"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/agent"
	agentRepo "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/agent"
	aiService "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/ai"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// AgentMessageHandler handles incoming messages for agents with WhatsApp integrations
type AgentMessageHandler struct {
	agentRepo    *agentRepo.SQLiteRepository
	agentService *usecase.AgentService
	mu           sync.RWMutex
}

var (
	agentHandler     *AgentMessageHandler
	agentHandlerOnce sync.Once
)

// InitAgentHandler initializes the global agent message handler
func InitAgentHandler(repo *agentRepo.SQLiteRepository, agentSvc *usecase.AgentService) {
	agentHandlerOnce.Do(func() {
		if repo != nil && agentSvc != nil {
			agentHandler = &AgentMessageHandler{
				agentRepo:    repo,
				agentService: agentSvc,
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
		logrus.Errorf("Failed to get agents: %v", err)
		return
	}

	remoteJID := evt.Info.Sender.String()

	for _, ag := range agents {
		if !ag.IsActive {
			continue
		}

		// Get integrations for this agent
		integrations, err := h.agentRepo.GetIntegrationsByAgentID(ctx, ag.ID)
		if err != nil {
			continue
		}

		// Find WhatsApp integration that matches this device
		var matchingIntegration *agent.Integration
		for _, integration := range integrations {
			if integration.Type == agent.IntegrationTypeWhatsApp && integration.IsConnected {
				// For now, use first connected WhatsApp integration
				// TODO: Match by device JID stored in config
				waConfig, _ := agentRepo.ParseWhatsAppConfig(integration.Config)
				if waConfig != nil && waConfig.DeviceID == deviceJID {
					matchingIntegration = integration
					break
				}
				// If no specific device configured, use any connected integration
				if matchingIntegration == nil {
					matchingIntegration = integration
				}
			}
		}

		if matchingIntegration == nil {
			continue
		}

		// Process message for this agent
		go h.processMessageForAgent(ctx, ag, matchingIntegration, remoteJID, userMessage, chatStorageRepo, client)
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
	if h.agentService == nil {
		logrus.Error("Agent service not initialized")
		return
	}

	// Handle audio transcription if needed
	if strings.HasPrefix(userMessage, "[AUDIO:") && strings.HasSuffix(userMessage, "]") {
		audioPath := userMessage[7 : len(userMessage)-1]
		aiSvc := aiService.NewService(ag.APIKey, ag.SerpAPIKey)
		if aiSvc != nil {
			transcription, err := aiSvc.TranscribeAudio(ctx, audioPath)
			if err != nil {
				logrus.Errorf("Failed to transcribe audio for agent %s: %v", ag.ID, err)
				return
			}
			userMessage = transcription
			logrus.Infof("Agent %s transcribed audio: %s", ag.ID, userMessage)
		}
	}

	// Use AgentService to handle the message (same as Telegram)
	response, err := h.agentService.HandleIncomingMessage(ctx, ag.ID, integration.ID, remoteJID, userMessage)
	if err != nil {
		logrus.Errorf("Failed to get AI response for agent %s: %v", ag.ID, err)
		return
	}

	if response == "" {
		logrus.Warnf("Agent %s: AI returned empty response (manual mode or error)", ag.ID)
		return
	}

	// Send the response via WhatsApp
	recipientJID := utils.FormatJID(remoteJID)
	sendResp, err := client.SendMessage(
		ctx,
		recipientJID,
		&waE2E.Message{Conversation: proto.String(response)},
	)

	if err != nil {
		logrus.Errorf("Failed to send agent %s response: %v", ag.ID, err)
		return
	}

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

	logrus.Infof("Agent %s sent response to %s", ag.Name, remoteJID)
}

