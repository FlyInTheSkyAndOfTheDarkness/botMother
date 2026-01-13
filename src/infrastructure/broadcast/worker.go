package broadcast

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/agent"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/settings"
	agentRepo "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/agent"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/telegram"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// BroadcastWorker handles background sending of broadcast messages
type BroadcastWorker struct {
	settingsService *usecase.SettingsService
	agentRepo       *agentRepo.SQLiteRepository
	rateLimiter     *RateLimiter
	maxRetries      int
	retryDelay      time.Duration
}

// RateLimiter controls message sending rate
type RateLimiter struct {
	mu          sync.Mutex
	lastSent    time.Time
	minInterval time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(minInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		minInterval: minInterval,
	}
}

// Wait blocks until it's safe to send the next message
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	elapsed := time.Since(rl.lastSent)
	if elapsed < rl.minInterval {
		sleepTime := rl.minInterval - elapsed
		time.Sleep(sleepTime)
	}
	rl.lastSent = time.Now()
}

// NewBroadcastWorker creates a new broadcast worker
func NewBroadcastWorker(settingsService *usecase.SettingsService, agentRepo *agentRepo.SQLiteRepository) *BroadcastWorker {
	return &BroadcastWorker{
		settingsService: settingsService,
		agentRepo:       agentRepo,
		rateLimiter:     NewRateLimiter(1 * time.Second), // 1 message per second by default
		maxRetries:      3,
		retryDelay:      5 * time.Second,
	}
}

// ExecuteBroadcast sends broadcast messages in background
func (w *BroadcastWorker) ExecuteBroadcast(ctx context.Context, broadcastID string) error {
	// Get broadcast
	broadcast, err := w.settingsService.GetBroadcast(ctx, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to get broadcast: %w", err)
	}

	if broadcast.Status != "pending" {
		return fmt.Errorf("broadcast %s is not in pending status (current: %s)", broadcastID, broadcast.Status)
	}

	// Update status to sending
	if err := w.settingsService.UpdateBroadcastStatus(ctx, broadcastID, "sending", 0, 0); err != nil {
		return fmt.Errorf("failed to update broadcast status: %w", err)
	}

	logrus.Infof("📢 [Broadcast] Starting broadcast %s: %d recipients via %s", broadcastID, len(broadcast.Recipients), broadcast.IntegrationType)

	// Start sending in background
	go w.sendBroadcast(ctx, broadcast)

	return nil
}

// sendBroadcast sends messages to all recipients
func (w *BroadcastWorker) sendBroadcast(ctx context.Context, broadcast *settings.BroadcastMessage) {
	sentCount := 0
	failedCount := 0

	for i, recipient := range broadcast.Recipients {
		// Rate limiting
		w.rateLimiter.Wait()

		// Try to send with retries
		err := w.sendWithRetry(ctx, broadcast, recipient, i+1, len(broadcast.Recipients))
		if err != nil {
			logrus.Errorf("❌ [Broadcast] Failed to send to %s: %v", recipient, err)
			failedCount++
		} else {
			sentCount++
		}

		// Update progress periodically (every 10 messages or at the end)
		if (i+1)%10 == 0 || i+1 == len(broadcast.Recipients) {
			if err := w.settingsService.UpdateBroadcastStatus(ctx, broadcast.ID, "sending", sentCount, failedCount); err != nil {
				logrus.Warnf("⚠️  [Broadcast] Failed to update progress: %v", err)
			}
			logrus.Infof("📊 [Broadcast] Progress: %d/%d sent, %d failed", sentCount, len(broadcast.Recipients), failedCount)
		}
	}

	// Final status update
	finalStatus := "completed"
	if failedCount == len(broadcast.Recipients) {
		finalStatus = "failed"
	} else if failedCount > 0 {
		finalStatus = "completed" // Partial success
	}

	if err := w.settingsService.UpdateBroadcastStatus(ctx, broadcast.ID, finalStatus, sentCount, failedCount); err != nil {
		logrus.Errorf("❌ [Broadcast] Failed to update final status: %v", err)
	} else {
		logrus.Infof("✅ [Broadcast] Completed broadcast %s: %d sent, %d failed", broadcast.ID, sentCount, failedCount)
	}
}

// sendWithRetry sends a message with retry logic
func (w *BroadcastWorker) sendWithRetry(ctx context.Context, broadcast *settings.BroadcastMessage, recipient string, current, total int) error {
	var lastErr error
	for attempt := 1; attempt <= w.maxRetries; attempt++ {
		err := w.sendMessage(ctx, broadcast, recipient)
		if err == nil {
			if attempt > 1 {
				logrus.Infof("✅ [Broadcast] Successfully sent to %s after %d attempts", recipient, attempt)
			}
			return nil
		}

		lastErr = err
		if attempt < w.maxRetries {
			logrus.Warnf("⚠️  [Broadcast] Attempt %d/%d failed for %s: %v, retrying in %v...", attempt, w.maxRetries, recipient, err, w.retryDelay)
			time.Sleep(w.retryDelay)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", w.maxRetries, lastErr)
}

// sendMessage sends a single message based on integration type
func (w *BroadcastWorker) sendMessage(ctx context.Context, broadcast *settings.BroadcastMessage, recipient string) error {
	switch broadcast.IntegrationType {
	case agent.IntegrationTypeWhatsApp:
		return w.sendWhatsAppMessage(ctx, broadcast, recipient)
	case agent.IntegrationTypeTelegram:
		return w.sendTelegramMessage(ctx, broadcast, recipient)
	default:
		return fmt.Errorf("unsupported integration type: %s", broadcast.IntegrationType)
	}
}

// sendWhatsAppMessage sends a message via WhatsApp
func (w *BroadcastWorker) sendWhatsAppMessage(ctx context.Context, broadcast *settings.BroadcastMessage, recipient string) error {
	// Get agent integrations
	integrations, err := w.agentRepo.GetIntegrationsByAgentID(ctx, broadcast.AgentID)
	if err != nil {
		return fmt.Errorf("failed to get integrations: %w", err)
	}

	// Find WhatsApp integration
	var waIntegration *agent.Integration
	for _, integration := range integrations {
		if integration.Type == agent.IntegrationTypeWhatsApp && integration.IsConnected {
			waIntegration = integration
			break
		}
	}

	if waIntegration == nil {
		return fmt.Errorf("no connected WhatsApp integration found for agent %s", broadcast.AgentID)
	}

	// Parse WhatsApp config
	waConfig, err := agentRepo.ParseWhatsAppConfig(waIntegration.Config)
	if err != nil || waConfig == nil {
		return fmt.Errorf("failed to parse WhatsApp config: %w", err)
	}

	if waConfig.DeviceID == "" {
		return fmt.Errorf("WhatsApp integration has no device ID configured")
	}

	// Get device manager
	deviceManager := whatsapp.GetDeviceManager()
	if deviceManager == nil {
		return fmt.Errorf("device manager not initialized")
	}

	// Get device instance
	deviceInstance, ok := deviceManager.GetDevice(waConfig.DeviceID)
	if !ok || deviceInstance == nil {
		return fmt.Errorf("device %s not found", waConfig.DeviceID)
	}

	// Get WhatsApp client
	client := deviceInstance.GetClient()
	if client == nil || !client.IsConnected() {
		return fmt.Errorf("WhatsApp client for device %s is not connected", waConfig.DeviceID)
	}

	// Parse recipient JID
	recipientJID, err := utils.ParseJID(recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient JID %s: %w", recipient, err)
	}

	// Send message
	logrus.Debugf("📤 [Broadcast] Sending WhatsApp message to %s via device %s", recipientJID, waConfig.DeviceID)
	_, err = client.SendMessage(ctx, recipientJID, &waE2E.Message{
		Conversation: proto.String(broadcast.Message),
	})
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp message: %w", err)
	}

	logrus.Debugf("✅ [Broadcast] WhatsApp message sent to %s", recipientJID)
	return nil
}

// sendTelegramMessage sends a message via Telegram
func (w *BroadcastWorker) sendTelegramMessage(ctx context.Context, broadcast *settings.BroadcastMessage, recipient string) error {
	// Get agent integrations
	integrations, err := w.agentRepo.GetIntegrationsByAgentID(ctx, broadcast.AgentID)
	if err != nil {
		return fmt.Errorf("failed to get integrations: %w", err)
	}

	// Find Telegram integration
	var tgIntegration *agent.Integration
	for _, integration := range integrations {
		if integration.Type == agent.IntegrationTypeTelegram && integration.IsConnected {
			tgIntegration = integration
			break
		}
	}

	if tgIntegration == nil {
		return fmt.Errorf("no connected Telegram integration found for agent %s", broadcast.AgentID)
	}

	// Parse Telegram config
	tgConfig, err := agentRepo.ParseTelegramConfig(tgIntegration.Config)
	if err != nil || tgConfig == nil {
		return fmt.Errorf("failed to parse Telegram config: %w", err)
	}

	// Parse recipient chat ID (should be int64)
	var chatID int64
	if _, err := fmt.Sscanf(recipient, "%d", &chatID); err != nil {
		return fmt.Errorf("invalid Telegram chat ID %s: %w", recipient, err)
	}

	// Send message directly using token
	logrus.Debugf("📤 [Broadcast] Sending Telegram message to chat %d", chatID)
	err = telegram.SendMessageDirect(tgConfig.BotToken, chatID, broadcast.Message)
	if err != nil {
		return fmt.Errorf("failed to send Telegram message: %w", err)
	}

	logrus.Debugf("✅ [Broadcast] Telegram message sent to chat %d", chatID)
	return nil
}

