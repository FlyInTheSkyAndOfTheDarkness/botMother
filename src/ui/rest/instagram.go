package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/instagram"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type InstagramHandler struct{}

func InitRestInstagram(app fiber.Router) InstagramHandler {
	handler := InstagramHandler{}

	// Instagram webhook endpoint
	app.Post("/instagram/webhook", handler.Webhook)
	app.Get("/instagram/webhook", handler.VerifyWebhook)

	return handler
}

// VerifyWebhook handles Instagram webhook verification (GET request)
func (h *InstagramHandler) VerifyWebhook(c *fiber.Ctx) error {
	// Instagram webhook verification
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	// TODO: Store verify_token in config
	expectedToken := "your_verify_token_here" // Should be configurable

	if mode == "subscribe" && token == expectedToken {
		logrus.Infof("✅ [Instagram] Webhook verified successfully")
		return c.SendString(challenge)
	}

	logrus.Warnf("⚠️  [Instagram] Webhook verification failed: mode=%s, token=%s", mode, token)
	return fiber.NewError(fiber.StatusForbidden, "Verification failed")
}

// Webhook handles incoming Instagram webhook events (POST request)
func (h *InstagramHandler) Webhook(c *fiber.Ctx) error {
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		logrus.Errorf("❌ [Instagram] Failed to parse webhook payload: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid payload")
	}

	logrus.Debugf("📷 [Instagram] Received webhook: %+v", payload)

	// Get handler
	handler := instagram.GetInstagramHandler()
	if handler == nil {
		logrus.Error("❌ [Instagram] Handler not initialized")
		return fiber.NewError(fiber.StatusInternalServerError, "Handler not initialized")
	}

	// Process webhook
	ctx := c.UserContext()
	entries, ok := payload["entry"].([]interface{})
	if !ok {
		logrus.Warnf("⚠️  [Instagram] No entries in webhook payload")
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Webhook received",
		})
	}

	for _, entry := range entries {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}

		// Get page ID
		pageID, _ := entryMap["id"].(string)
		if pageID == "" {
			// Try messaging field
			messaging, ok := entryMap["messaging"].([]interface{})
			if !ok {
				continue
			}

			for _, msg := range messaging {
				msgMap, ok := msg.(map[string]interface{})
				if !ok {
					continue
				}

				// Extract sender and message
				sender, ok := msgMap["sender"].(map[string]interface{})
				if !ok {
					continue
				}
				senderID, _ := sender["id"].(string)

				message, ok := msgMap["message"].(map[string]interface{})
				if !ok {
					continue
				}
				messageText, _ := message["text"].(string)

				if senderID != "" && messageText != "" {
					// Try to get page ID from entry
					pageID, _ = entryMap["id"].(string)
					if pageID == "" {
						// Try to get from recipient
						recipient, ok := msgMap["recipient"].(map[string]interface{})
						if ok {
							pageID, _ = recipient["id"].(string)
						}
					}

					logrus.Infof("📷 [Instagram] Processing message from %s to page %s: %s", senderID, pageID, messageText)
					if err := handler.HandleIncomingMessage(ctx, senderID, messageText, pageID); err != nil {
						logrus.Errorf("❌ [Instagram] Failed to handle message: %v", err)
					}
				}
			}
		}
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Webhook processed",
	})
}

