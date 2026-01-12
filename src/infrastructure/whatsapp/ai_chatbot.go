package whatsapp

import (
	"context"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	aiService "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/ai"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

var aiServiceInstance *aiService.Service

func initAIService() {
	if config.AIChatbotEnabled && config.AIChatbotAPIToken != "" {
		aiServiceInstance = aiService.NewService(config.AIChatbotAPIToken, "") // Legacy chatbot doesn't use SerpAPI
	} else {
		aiServiceInstance = nil
	}
}

func getAIService() *aiService.Service {
	// Reinitialize if config changed
	if config.AIChatbotEnabled && config.AIChatbotAPIToken != "" {
		if aiServiceInstance == nil {
			aiServiceInstance = aiService.NewService(config.AIChatbotAPIToken, "") // Legacy chatbot doesn't use SerpAPI
		}
		return aiServiceInstance
	}
	return nil
}

func handleAIChatbot(ctx context.Context, evt *events.Message, chatStorageRepo domainChatStorage.IChatStorageRepository, client *whatsmeow.Client) {
	// Check if AI chatbot is enabled
	if !config.AIChatbotEnabled {
		return
	}

	service := getAIService()
	if service == nil {
		return
	}

	if client == nil {
		return
	}

	// Skip groups, broadcasts, and self messages (same logic as auto-reply)
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

	var userMessage string

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
		userMessage = conv
	} else if ext := innerMsg.GetExtendedTextMessage(); ext != nil && ext.GetText() != "" {
		userMessage = ext.GetText()
	} else if protoMsg := innerMsg.GetProtocolMessage(); protoMsg != nil {
		if edited := protoMsg.GetEditedMessage(); edited != nil {
			if ext := edited.GetExtendedTextMessage(); ext != nil && ext.GetText() != "" {
				userMessage = ext.GetText()
			} else if conv := edited.GetConversation(); conv != "" {
				userMessage = conv
			}
		}
	} else if audioMsg := innerMsg.GetAudioMessage(); audioMsg != nil {
		// Handle audio messages - transcribe them
		if !config.WhatsappAutoDownloadMedia {
			log.Warnf("AI chatbot received audio message but auto-download-media is disabled. Cannot transcribe.")
			return
		}

		extractedMedia, err := utils.ExtractMedia(ctx, client, config.PathMedia, audioMsg)
		if err != nil {
			log.Errorf("Failed to download audio for transcription: %v", err)
			return
		}

		// Transcribe audio using the file path
		transcription, err := service.TranscribeAudio(ctx, extractedMedia.MediaPath)
		if err != nil {
			log.Errorf("Failed to transcribe audio: %v", err)
			return
		}

		userMessage = transcription
		log.Infof("Transcribed audio message: %s", userMessage)
	}

	if userMessage == "" {
		return
	}

	// Generate AI response
	aiResponse, err := service.GenerateResponse(ctx, userMessage, config.AIChatbotSystemPrompt, config.AIChatbotModel)
	if err != nil {
		log.Errorf("Failed to generate AI response: %v", err)
		return
	}

	if aiResponse == "" {
		log.Warnf("AI service returned empty response")
		return
	}

	// Format recipient JID
	recipientJID := utils.FormatJID(evt.Info.Sender.String())

	// Send the AI response
	response, err := client.SendMessage(
		ctx,
		recipientJID,
		&waE2E.Message{Conversation: proto.String(aiResponse)},
	)

	if err != nil {
		log.Errorf("Failed to send AI chatbot response: %v", err)
		return
	}

	// Store the AI response in chat storage
	if chatStorageRepo != nil {
		senderJID := ""
		if client.Store.ID != nil {
			senderJID = client.Store.ID.String()
		}

		if err := chatStorageRepo.StoreSentMessageWithContext(
			ctx,
			response.ID,
			senderJID,
			recipientJID.String(),
			aiResponse,
			response.Timestamp,
		); err != nil {
			log.Errorf("Failed to store AI chatbot response in chat storage: %v", err)
		} else {
			log.Debugf("AI chatbot response %s stored successfully in chat storage", response.ID)
		}
	}
}

