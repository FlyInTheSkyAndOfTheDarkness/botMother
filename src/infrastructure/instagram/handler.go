package instagram

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
	agentRepo "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/agent"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/sirupsen/logrus"
)

// InstagramMessageHandler handles incoming messages for agents with Instagram integrations
type InstagramMessageHandler struct {
	agentRepo   *agentRepo.SQLiteRepository
	agentService *usecase.AgentService
	mu          sync.RWMutex
}

var (
	instagramHandler *InstagramMessageHandler
	handlerOnce     sync.Once
)

// InitInstagramHandler initializes the Instagram message handler
func InitInstagramHandler(agentRepo *agentRepo.SQLiteRepository, agentService *usecase.AgentService) {
	handlerOnce.Do(func() {
		instagramHandler = &InstagramMessageHandler{
			agentRepo:    agentRepo,
			agentService: agentService,
		}
		logrus.Info("📷 [Instagram] Handler initialized")
	})
}

// GetInstagramHandler returns the singleton Instagram handler
func GetInstagramHandler() *InstagramMessageHandler {
	return instagramHandler
}

// HandleIncomingMessage processes an incoming Instagram message
func (h *InstagramMessageHandler) HandleIncomingMessage(ctx context.Context, senderID, messageText string, pageID string) error {
	if h == nil || h.agentRepo == nil {
		return fmt.Errorf("Instagram handler not initialized")
	}

	logrus.Infof("📷 [Instagram] Processing message from sender %s (page: %s): %s", senderID, pageID, messageText[:min(50, len(messageText))])

	// Find agents with active Instagram integrations matching this page
	agents, err := h.agentRepo.GetAll(ctx)
	if err != nil {
		logrus.Errorf("❌ [Instagram] Failed to get agents: %v", err)
		return fmt.Errorf("failed to get agents: %w", err)
	}

	foundAgent := false
	for _, ag := range agents {
		if !ag.IsActive {
			logrus.Debugf("⏭️  [Instagram] Skipping agent %s (%s) - not active", ag.ID, ag.Name)
			continue
		}

		logrus.Infof("✅ [Instagram] Checking agent %s (%s) - is active", ag.ID, ag.Name)

		// Get integrations for this agent
		integrations, err := h.agentRepo.GetIntegrationsByAgentID(ctx, ag.ID)
		if err != nil {
			logrus.Warnf("⚠️  [Instagram] Failed to get integrations for agent %s: %v", ag.ID, err)
			continue
		}

		// Find Instagram integration that matches this page
		var matchingIntegration *agent.Integration
		for _, integration := range integrations {
			if integration.Type != agent.IntegrationTypeInstagram {
				continue
			}

			if !integration.IsConnected {
				logrus.Debugf("⏭️  [Instagram] Integration %s not connected", integration.ID)
				continue
			}

			// Parse Instagram config
			igConfig, err := ParseInstagramConfig(integration.Config)
			if err != nil || igConfig == nil {
				logrus.Warnf("⚠️  [Instagram] Failed to parse config for integration %s: %v", integration.ID, err)
				continue
			}

			// Match by Page ID
			if igConfig.PageID == pageID {
				matchingIntegration = integration
				logrus.Infof("✅ [Instagram] Found matching integration %s for page %s", integration.ID, pageID)
				break
			}
		}

		if matchingIntegration == nil {
			logrus.Debugf("⏭️  [Instagram] No matching Instagram integration found for agent %s with page %s", ag.ID, pageID)
			continue
		}

		foundAgent = true
		logrus.Infof("🚀 [Instagram] Processing message for agent %s (%s) with integration %s", ag.ID, ag.Name, matchingIntegration.ID)

		// Process message for this agent
		go h.processMessageForAgent(ctx, ag, matchingIntegration, senderID, messageText)
	}

	if !foundAgent {
		logrus.Warnf("⚠️  [Instagram] No active agent with connected Instagram integration found for message from %s (page: %s)", senderID, pageID)
	}

	return nil
}

// processMessageForAgent processes a message for a specific agent
func (h *InstagramMessageHandler) processMessageForAgent(ctx context.Context, ag *agent.Agent, integration *agent.Integration, senderID, userMessage string) {
	// Use senderID as remoteJID for Instagram
	remoteJID := fmt.Sprintf("ig_%s", senderID)

	// Get AI response from agent service
	if h.agentService == nil {
		logrus.Error("❌ [Instagram] Agent service is nil, cannot process message")
		return
	}

	logrus.Infof("🤖 [Instagram] Calling HandleIncomingMessage for agent %s, integration %s, user %s", ag.ID, integration.ID, senderID)
	response, err := h.agentService.HandleIncomingMessage(ctx, ag.ID, integration.ID, remoteJID, userMessage)
	if err != nil {
		logrus.Errorf("❌ [Instagram] Failed to get AI response for agent %s: %v", ag.ID, err)
		return
	}

	if response == "" {
		logrus.Warnf("⚠️  [Instagram] AI returned empty response for agent %s (manual mode or error)", ag.ID)
		return
	}

	logrus.Infof("💡 [Instagram] AI response generated for agent %s: %s", ag.ID, response[:min(50, len(response))])

	// Send response via Instagram API
	igConfig, _ := ParseInstagramConfig(integration.Config)
	if igConfig == nil {
		logrus.Errorf("❌ [Instagram] Failed to parse Instagram config for integration %s", integration.ID)
		return
	}

	if err := SendInstagramMessage(igConfig.AccessToken, igConfig.PageID, senderID, response); err != nil {
		logrus.Errorf("❌ [Instagram] Failed to send message to %s: %v", senderID, err)
	} else {
		logrus.Infof("✅ [Instagram] Response sent successfully to %s", senderID)
	}
}

// ParseInstagramConfig parses Instagram config from JSON
func ParseInstagramConfig(configJSON string) (*agent.InstagramConfig, error) {
	var config agent.InstagramConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SendInstagramMessage sends a message via Instagram Graph API
func SendInstagramMessage(accessToken, pageID, recipientID, message string) error {
	// Instagram Graph API endpoint for sending messages
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", pageID)

	payload := map[string]interface{}{
		"recipient": map[string]string{
			"id": recipientID,
		},
		"message": map[string]string{
			"text": message,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Make HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Instagram API error (status %d): %s", resp.StatusCode, string(body))
	}

	logrus.Debugf("✅ [Instagram] Message sent successfully to %s", recipientID)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

