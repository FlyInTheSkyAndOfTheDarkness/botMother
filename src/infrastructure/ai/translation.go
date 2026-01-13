package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

// TranslateText translates text from source language to target language using OpenAI
func (s *Service) TranslateText(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("AI service not initialized")
	}

	if text == "" {
		return "", nil
	}

	// Use OpenAI for translation
	prompt := fmt.Sprintf("Translate the following text from %s to %s. Only return the translation, no explanations:\n\n%s", sourceLang, targetLang, text)

	req := openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a professional translator. Translate accurately and preserve the meaning and tone.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   500,
		Temperature: 0.3, // Lower temperature for more consistent translations
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		logrus.Errorf("Failed to translate text: %v", err)
		return "", fmt.Errorf("failed to translate text: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no translation response")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// DetectLanguage detects the language of the text using OpenAI
func (s *Service) DetectLanguage(ctx context.Context, text string) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("AI service not initialized")
	}

	if text == "" {
		return "en", nil // Default to English
	}

	prompt := fmt.Sprintf("Detect the language of the following text. Respond with only the ISO 639-1 language code (e.g., 'en', 'ru', 'es', 'fr'):\n\n%s", text)

	req := openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a language detection expert. Respond with only the ISO 639-1 language code.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   10,
		Temperature: 0.1,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		logrus.Errorf("Failed to detect language: %v", err)
		return "en", nil // Default to English on error
	}

	if len(resp.Choices) == 0 {
		return "en", nil
	}

	langCode := strings.TrimSpace(strings.ToLower(resp.Choices[0].Message.Content))
	// Validate it's a 2-letter code
	if len(langCode) == 2 {
		return langCode, nil
	}

	return "en", nil // Default to English if invalid
}

// AnalyzeSentiment analyzes the sentiment of the text and returns a score (-1 to 1, where -1 is very negative, 1 is very positive)
func (s *Service) AnalyzeSentiment(ctx context.Context, text string) (float64, string, error) {
	if s == nil || s.client == nil {
		return 0, "neutral", fmt.Errorf("AI service not initialized")
	}

	if text == "" {
		return 0, "neutral", nil
	}

	prompt := fmt.Sprintf("Analyze the sentiment of the following text. Respond with ONLY a JSON object in this exact format: {\"score\": -0.5, \"label\": \"negative\"}\n\nScore should be between -1 (very negative) and 1 (very positive). Label should be one of: \"very_negative\", \"negative\", \"neutral\", \"positive\", \"very_positive\".\n\nText: %s", text)

	req := openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a sentiment analysis expert. Always respond with valid JSON only.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   50,
		Temperature: 0.3,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		logrus.Errorf("Failed to analyze sentiment: %v", err)
		return 0, "neutral", fmt.Errorf("failed to analyze sentiment: %w", err)
	}

	if len(resp.Choices) == 0 {
		return 0, "neutral", nil
	}

	// Parse JSON response
	responseText := strings.TrimSpace(resp.Choices[0].Message.Content)
	
	// Try to extract JSON from response (in case there's extra text)
	jsonStart := strings.Index(responseText, "{")
	jsonEnd := strings.LastIndex(responseText, "}")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		responseText = responseText[jsonStart : jsonEnd+1]
	}

	var result struct {
		Score float64 `json:"score"`
		Label string  `json:"label"`
	}

	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		logrus.Warnf("Failed to parse sentiment response: %v, response: %s", err, responseText)
		return 0, "neutral", nil
	}

	return result.Score, result.Label, nil
}

