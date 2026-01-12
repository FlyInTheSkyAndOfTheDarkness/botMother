package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/agent"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/sirupsen/logrus"
)

// TelegramUpdate represents incoming update from Telegram
type TelegramUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *TelegramMessage `json:"message,omitempty"`
}

// TelegramMessage represents a Telegram message
type TelegramMessage struct {
	MessageID int           `json:"message_id"`
	From      *TelegramUser `json:"from,omitempty"`
	Chat      *TelegramChat `json:"chat"`
	Date      int           `json:"date"`
	Text      string        `json:"text,omitempty"`
}

// TelegramUser represents a Telegram user
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// TelegramChat represents a Telegram chat
type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// TelegramBot manages a single Telegram bot connection
type TelegramBot struct {
	Token         string
	AgentID       string
	IntegrationID string
	agentService  *usecase.AgentService
	stopChan      chan struct{}
	running       bool
	mu            sync.Mutex
}

// BotManager manages multiple Telegram bots
type BotManager struct {
	bots         map[string]*TelegramBot // key: integration_id
	agentService *usecase.AgentService
	mu           sync.RWMutex
}

var (
	botManager *BotManager
	once       sync.Once
)

// GetBotManager returns the singleton bot manager
func GetBotManager() *BotManager {
	return botManager
}

// InitBotManager initializes the bot manager
func InitBotManager(agentService *usecase.AgentService) *BotManager {
	once.Do(func() {
		botManager = &BotManager{
			bots:         make(map[string]*TelegramBot),
			agentService: agentService,
		}
	})
	return botManager
}

// StartBot starts a Telegram bot for an integration
func (m *BotManager) StartBot(integrationID, agentID, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop existing bot if any
	if bot, exists := m.bots[integrationID]; exists {
		bot.Stop()
	}

	bot := &TelegramBot{
		Token:         token,
		AgentID:       agentID,
		IntegrationID: integrationID,
		agentService:  m.agentService,
		stopChan:      make(chan struct{}),
	}

	m.bots[integrationID] = bot
	go bot.Start()

	logrus.Infof("Started Telegram bot for integration %s", integrationID)
	return nil
}

// StopBot stops a Telegram bot
func (m *BotManager) StopBot(integrationID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if bot, exists := m.bots[integrationID]; exists {
		bot.Stop()
		delete(m.bots, integrationID)
		logrus.Infof("Stopped Telegram bot for integration %s", integrationID)
	}
}

// Start begins polling for updates
func (b *TelegramBot) Start() {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return
	}
	b.running = true
	b.mu.Unlock()

	tokenPreview := b.Token
	if len(tokenPreview) > 15 {
		tokenPreview = tokenPreview[:15] + "..."
	}
	logrus.Infof("🤖 Telegram bot STARTING for agent %s (token: %s)", b.AgentID, tokenPreview)

	// Test connection first
	testURL := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", b.Token)
	resp, err := http.Get(testURL)
	if err != nil {
		logrus.Errorf("❌ Telegram bot connection test failed: %v", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	var meResult struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
		Description string `json:"description"`
	}
	json.Unmarshal(body, &meResult)
	
	if !meResult.OK {
		logrus.Errorf("❌ Telegram bot auth failed: %s", meResult.Description)
		return
	}
	
	logrus.Infof("✅ Telegram bot @%s connected successfully!", meResult.Result.Username)

	offset := 0
	for {
		select {
		case <-b.stopChan:
			logrus.Info("🛑 Telegram bot stopped")
			return
		default:
			updates, err := b.getUpdates(offset)
			if err != nil {
				logrus.Errorf("Failed to get Telegram updates: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for _, update := range updates {
				offset = update.UpdateID + 1
				b.handleUpdate(update)
			}

			time.Sleep(1 * time.Second)
		}
	}
}

// Stop stops the bot
func (b *TelegramBot) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		close(b.stopChan)
		b.running = false
	}
}

func (b *TelegramBot) getUpdates(offset int) ([]TelegramUpdate, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30", b.Token, offset)
	
	client := &http.Client{Timeout: 35 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		OK          bool             `json:"ok"`
		Result      []TelegramUpdate `json:"result"`
		ErrorCode   int              `json:"error_code,omitempty"`
		Description string           `json:"description,omitempty"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	if !result.OK {
		return nil, fmt.Errorf("telegram API error %d: %s", result.ErrorCode, result.Description)
	}

	return result.Result, nil
}

func (b *TelegramBot) handleUpdate(update TelegramUpdate) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	msg := update.Message
	
	// Safety checks
	if msg.Chat == nil {
		logrus.Warn("Telegram message has no chat info")
		return
	}
	if msg.From == nil {
		logrus.Warn("Telegram message has no sender info")
		return
	}
	
	chatID := msg.Chat.ID
	userMessage := msg.Text
	userID := fmt.Sprintf("tg_%d", msg.From.ID)

	logrus.Infof("Telegram message from %s (chat %d): %s", userID, chatID, userMessage)

	// Check if agent service is available
	if b.agentService == nil {
		logrus.Error("Agent service is nil, cannot process message")
		return
	}

	// Get AI response from agent service
	ctx := context.Background()
	response, err := b.agentService.HandleIncomingMessage(ctx, b.AgentID, b.IntegrationID, userID, userMessage)
	if err != nil {
		logrus.Errorf("Failed to get AI response for agent %s: %v", b.AgentID, err)
		// Send error message to user
		b.sendMessage(chatID, "Sorry, I encountered an error processing your message.")
		return
	}

	if response == "" {
		logrus.Warn("AI returned empty response")
		return
	}

	logrus.Infof("Sending response to chat %d: %s", chatID, response[:min(50, len(response))]+"...")

	// Send response
	if err := b.sendMessage(chatID, response); err != nil {
		logrus.Errorf("Failed to send Telegram message: %v", err)
	} else {
		logrus.Infof("Successfully sent response to chat %d", chatID)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (b *TelegramBot) sendMessage(chatID int64, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.Token)
	
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error: %s", string(body))
	}

	return nil
}

// LoadBotsFromDB loads and starts all active Telegram integrations
func (m *BotManager) LoadBotsFromDB(ctx context.Context, agentRepo agent.IAgentRepository) error {
	agents, err := agentRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, a := range agents {
		if !a.IsActive {
			continue
		}

		integrations, err := agentRepo.GetIntegrationsByAgentID(ctx, a.ID)
		if err != nil {
			continue
		}

		for _, integration := range integrations {
			if integration.Type == agent.IntegrationTypeTelegram && integration.IsConnected {
				var config agent.TelegramConfig
				if err := json.Unmarshal([]byte(integration.Config), &config); err != nil {
					logrus.Errorf("Failed to parse Telegram config: %v", err)
					continue
				}

				if config.BotToken != "" {
					m.StartBot(integration.ID, a.ID, config.BotToken)
				}
			}
		}
	}

	return nil
}

