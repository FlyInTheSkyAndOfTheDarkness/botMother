package followup

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/agent"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/settings"
	agentRepo "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/agent"
	instagramPkg "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/instagram"
	telegramBot "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/telegram"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// FollowUpWorker handles automatic follow-up messages
type FollowUpWorker struct {
	settingsService settings.ISettingsService
	agentRepo       *agentRepo.SQLiteRepository
	checkInterval   time.Duration
	stopChan        chan struct{}
}

// NewFollowUpWorker creates a new follow-up worker
func NewFollowUpWorker(
	settingsService settings.ISettingsService,
	agentRepo *agentRepo.SQLiteRepository,
) *FollowUpWorker {
	return &FollowUpWorker{
		settingsService: settingsService,
		agentRepo:       agentRepo,
		checkInterval:   5 * time.Minute, // Check every 5 minutes
		stopChan:        make(chan struct{}),
	}
}

// Start begins the follow-up worker
func (w *FollowUpWorker) Start(ctx context.Context) {
	logrus.Info("🔄 Follow-up worker started")
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.processFollowUps(ctx)
		case <-w.stopChan:
			logrus.Info("🛑 Follow-up worker stopped")
			return
		}
	}
}

// Stop stops the follow-up worker
func (w *FollowUpWorker) Stop() {
	close(w.stopChan)
}

// processFollowUps checks all agents and sends follow-up messages where needed
func (w *FollowUpWorker) processFollowUps(ctx context.Context) {
	logrus.Debug("🔍 Checking for follow-up opportunities...")

	// Get all agents
	agents, err := w.agentRepo.GetAll(ctx)
	if err != nil {
		logrus.Errorf("❌ Follow-up worker: Failed to get agents: %v", err)
		return
	}

	for _, ag := range agents {
		if !ag.IsActive {
			continue
		}

		// Get agent settings
		agentSettings, err := w.settingsService.GetAgentSettings(ctx, ag.ID)
		if err != nil {
			logrus.Debugf("⚠️  Follow-up worker: Failed to get settings for agent %s: %v", ag.ID, err)
			continue
		}

		// Check if follow-up is enabled
		if !agentSettings.FollowUp.Enabled {
			continue
		}

		// Process follow-ups for this agent
		w.processAgentFollowUps(ctx, ag, agentSettings)
	}
}

// processAgentFollowUps processes follow-ups for a specific agent
func (w *FollowUpWorker) processAgentFollowUps(ctx context.Context, agent *agent.Agent, settings *settings.AgentSettings) {
	// Get all conversations for this agent
	conversations, err := w.getAgentConversations(ctx, agent.ID)
	if err != nil {
		logrus.Errorf("❌ Follow-up worker: Failed to get conversations for agent %s: %v", agent.ID, err)
		return
	}

	for _, conv := range conversations {
		// Skip if in manual mode
		if conv.IsManualMode {
			continue
		}

		// Get last message
		lastMessage, err := w.getLastMessage(ctx, conv.ID)
		if err != nil {
			logrus.Debugf("⚠️  Follow-up worker: Failed to get last message for conversation %s: %v", conv.ID, err)
			continue
		}

		// Check if last message was from assistant (agent)
		if lastMessage == nil || lastMessage.Role != "assistant" {
			continue // User hasn't replied yet, or no messages
		}

		// Check if enough time has passed
		timeSinceLastMessage := time.Since(lastMessage.Timestamp)
		requiredDelay := time.Duration(settings.FollowUp.DelayMinutes) * time.Minute
		if timeSinceLastMessage < requiredDelay {
			continue // Not enough time has passed
		}

		// Check if user has replied since last assistant message
		if settings.FollowUp.OnlyIfNoReply {
			hasUserReply, err := w.hasUserReplyAfter(ctx, conv.ID, lastMessage.Timestamp)
			if err != nil {
				logrus.Debugf("⚠️  Follow-up worker: Failed to check user reply: %v", err)
				continue
			}
			if hasUserReply {
				continue // User has replied, no need for follow-up
			}
		}

		// Count existing follow-up messages
		followUpCount, err := w.countFollowUpMessages(ctx, conv.ID, lastMessage.Timestamp)
		if err != nil {
			logrus.Debugf("⚠️  Follow-up worker: Failed to count follow-ups: %v", err)
			continue
		}

		// Check if max follow-ups reached
		if followUpCount >= settings.FollowUp.MaxFollowUps {
			continue // Max follow-ups reached
		}

		// Send follow-up message
		w.sendFollowUpMessage(ctx, agent, conv, settings)
	}
}

// getAgentConversations gets all conversations for an agent
func (w *FollowUpWorker) getAgentConversations(ctx context.Context, agentID string) ([]*agent.Conversation, error) {
	// Get all integrations for agent
	integrations, err := w.agentRepo.GetIntegrationsByAgentID(ctx, agentID)
	if err != nil {
		return nil, err
	}

	var allConversations []*agent.Conversation
	for _, integration := range integrations {
		// Get conversations for this integration
		// Get all conversations and filter by integration
		allConvs, err := w.agentRepo.GetAllConversations(ctx)
		if err == nil {
			for _, conv := range allConvs {
				if conv.IntegrationID == integration.ID && conv.AgentID == agentID {
					allConversations = append(allConversations, conv)
				}
			}
		}
	}

	return allConversations, nil
}

// getLastMessage gets the last message in a conversation
func (w *FollowUpWorker) getLastMessage(ctx context.Context, conversationID string) (*agent.Message, error) {
	messages, err := w.agentRepo.GetRecentMessages(ctx, conversationID, 1)
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		return nil, nil
	}
	return messages[0], nil
}

// hasUserReplyAfter checks if user has replied after a given timestamp
func (w *FollowUpWorker) hasUserReplyAfter(ctx context.Context, conversationID string, after time.Time) (bool, error) {
	messages, err := w.agentRepo.GetRecentMessages(ctx, conversationID, 10)
	if err != nil {
		return false, err
	}

	for _, msg := range messages {
		if msg.Role == "user" && msg.Timestamp.After(after) {
			return true, nil
		}
	}

	return false, nil
}

// countFollowUpMessages counts follow-up messages sent after a given timestamp
func (w *FollowUpWorker) countFollowUpMessages(ctx context.Context, conversationID string, after time.Time) (int, error) {
	messages, err := w.agentRepo.GetRecentMessages(ctx, conversationID, 50)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, msg := range messages {
		if msg.Role == "assistant" && msg.Timestamp.After(after) {
			// Check if this looks like a follow-up message (contains follow-up keywords)
			if w.isFollowUpMessage(msg.Content, []string{
				"checking if you have any questions",
				"anything else I can help",
				"just checking",
				"follow up",
			}) {
				count++
			}
		}
	}

	return count, nil
}

// isFollowUpMessage checks if a message looks like a follow-up
func (w *FollowUpWorker) isFollowUpMessage(content string, keywords []string) bool {
	lowerContent := strings.ToLower(content)
	for _, keyword := range keywords {
		if strings.Contains(lowerContent, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// sendFollowUpMessage sends a follow-up message
func (w *FollowUpWorker) sendFollowUpMessage(ctx context.Context, agent *agent.Agent, conv *agent.Conversation, settings *settings.AgentSettings) {
	// Get follow-up message template
	if len(settings.FollowUp.Messages) == 0 {
		logrus.Warnf("⚠️  Follow-up worker: No follow-up messages configured for agent %s", agent.ID)
		return
	}

	// Select message based on follow-up count
	followUpCount, _ := w.countFollowUpMessages(ctx, conv.ID, time.Now().Add(-24*time.Hour))
	messageIndex := followUpCount % len(settings.FollowUp.Messages)
	followUpMessage := settings.FollowUp.Messages[messageIndex]

	// Get integration
	integration, err := w.agentRepo.GetIntegrationByID(ctx, conv.IntegrationID)
	if err != nil {
		logrus.Errorf("❌ Follow-up worker: Failed to get integration %s: %v", conv.IntegrationID, err)
		return
	}

	if !integration.IsConnected {
		logrus.Debugf("⏭️  Follow-up worker: Integration %s not connected", conv.IntegrationID)
		return
	}

	// Send message based on integration type
	switch integration.Type {
	case "whatsapp":
		w.sendWhatsAppFollowUp(ctx, integration, conv, followUpMessage)
	case "telegram":
		w.sendTelegramFollowUp(ctx, integration, conv, followUpMessage)
	case "instagram":
		w.sendInstagramFollowUp(ctx, integration, conv, followUpMessage)
	default:
		logrus.Warnf("⚠️  Follow-up worker: Unsupported integration type: %s", integration.Type)
	}

	// Store follow-up message in conversation
	followUpMsg := &agent.Message{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        followUpMessage,
		Timestamp:      time.Now(),
	}
	if err := w.agentRepo.AddMessage(ctx, followUpMsg); err != nil {
		logrus.Warnf("⚠️  Follow-up worker: Failed to store follow-up message: %v", err)
	}

	logrus.Infof("✅ Follow-up worker: Sent follow-up message to conversation %s", conv.ID)
}

// sendWhatsAppFollowUp sends follow-up via WhatsApp
func (w *FollowUpWorker) sendWhatsAppFollowUp(ctx context.Context, integration *agent.Integration, conv *agent.Conversation, message string) {
	waConfig, err := agentRepo.ParseWhatsAppConfig(integration.Config)
	if err != nil || waConfig == nil || waConfig.DeviceID == "" {
		logrus.Errorf("❌ Follow-up worker: Invalid WhatsApp config")
		return
	}

	deviceManager := whatsapp.GetDeviceManager()
	if deviceManager == nil {
		logrus.Errorf("❌ Follow-up worker: Device manager not initialized")
		return
	}

	deviceInstance, ok := deviceManager.GetDevice(waConfig.DeviceID)
	if !ok || deviceInstance == nil || !deviceInstance.IsConnected() {
		logrus.Errorf("❌ Follow-up worker: Device %s not found or not connected", waConfig.DeviceID)
		return
	}

	client := deviceInstance.GetClient()
	if client == nil {
		logrus.Errorf("❌ Follow-up worker: WhatsApp client not available")
		return
	}

	recipientJID, err := utils.ParseJID(conv.RemoteJID)
	if err != nil {
		logrus.Errorf("❌ Follow-up worker: Invalid recipient JID: %v", err)
		return
	}

	_, err = client.SendMessage(ctx, recipientJID, &waE2E.Message{Conversation: proto.String(message)})
	if err != nil {
		logrus.Errorf("❌ Follow-up worker: Failed to send WhatsApp message: %v", err)
	}
}

// sendTelegramFollowUp sends follow-up via Telegram
func (w *FollowUpWorker) sendTelegramFollowUp(ctx context.Context, integration *agent.Integration, conv *agent.Conversation, message string) {
	tgConfig, err := agentRepo.ParseTelegramConfig(integration.Config)
	if err != nil || tgConfig == nil || tgConfig.BotToken == "" {
		logrus.Errorf("❌ Follow-up worker: Invalid Telegram config")
		return
	}

	chatID, err := strconv.ParseInt(conv.RemoteJID, 10, 64)
	if err != nil {
		logrus.Errorf("❌ Follow-up worker: Invalid Telegram chat ID: %v", err)
		return
	}

	if err := telegramBot.SendMessageDirect(tgConfig.BotToken, chatID, message); err != nil {
		logrus.Errorf("❌ Follow-up worker: Failed to send Telegram message: %v", err)
	}
}

// sendInstagramFollowUp sends follow-up via Instagram
func (w *FollowUpWorker) sendInstagramFollowUp(ctx context.Context, integration *agent.Integration, conv *agent.Conversation, message string) {
	igConfig, err := agentRepo.ParseInstagramConfig(integration.Config)
	if err != nil || igConfig == nil || igConfig.AccessToken == "" || igConfig.PageID == "" {
		logrus.Errorf("❌ Follow-up worker: Invalid Instagram config")
		return
	}

	// Extract sender ID from remoteJID (format: ig_<sender_id>)
	senderID := conv.RemoteJID
	if len(senderID) > 3 && senderID[:3] == "ig_" {
		senderID = senderID[3:]
	}

	if err := instagramPkg.SendInstagramMessage(igConfig.AccessToken, igConfig.PageID, senderID, message); err != nil {
		logrus.Errorf("❌ Follow-up worker: Failed to send Instagram message: %v", err)
	}
}

