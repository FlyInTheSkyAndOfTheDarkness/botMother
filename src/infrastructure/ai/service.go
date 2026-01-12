package ai

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	serp "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/serp"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Service struct {
	client     *openai.Client
	serpService *serp.SerpService
}

func NewService(apiToken string, serpAPIKey string) *Service {
	if apiToken == "" {
		return nil
	}
	s := &Service{
		client: openai.NewClient(apiToken),
	}
	if serpAPIKey != "" {
		s.serpService = serp.NewService(serpAPIKey)
		if s.serpService != nil {
			logrus.Infof("SerpAPI service initialized (key length: %d)", len(serpAPIKey))
		} else {
			logrus.Warnf("Failed to initialize SerpAPI service (key length: %d)", len(serpAPIKey))
		}
	} else {
		logrus.Debug("SerpAPI service not initialized (no API key provided)")
	}
	return s
}

// GenerateResponse generates an AI response for the given user message
func (s *Service) GenerateResponse(ctx context.Context, userMessage string, systemPrompt string, model string) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("AI service not initialized")
	}

	// Use default model if not specified
	if model == "" {
		model = config.AIChatbotModel
	}
	if model == "" {
		model = "gpt-4o-mini"
	}

	// Use default system prompt if not specified
	if systemPrompt == "" {
		systemPrompt = config.AIChatbotSystemPrompt
	}
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant. Respond concisely and helpfully to user messages."
	}

	// Check if user is asking for current/real-time information
	needsSearch := s.needsInternetSearch(userMessage)
	var searchResults string
	if needsSearch {
		if s.serpService == nil {
			logrus.Debugf("SerpAPI service not available (no API key configured)")
		} else {
			// Extract search query from user message
			searchQuery := s.extractSearchQuery(userMessage)
			if searchQuery != "" {
				logrus.Infof("Searching internet for: %s", searchQuery)
				results, err := s.serpService.Search(searchQuery)
				if err == nil {
					searchResults = results
					logrus.Infof("SerpAPI search successful, results length: %d", len(searchResults))
					// Add search results to user message
					userMessage = fmt.Sprintf("User question: %s\n\nSearch results from internet:\n%s\n\nPlease answer based on the search results above.", userMessage, searchResults)
				} else {
					logrus.Warnf("SerpAPI search failed: %v", err)
				}
			}
		}
	}

	req := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userMessage,
			},
		},
		MaxTokens: 500, // Limit response length for WhatsApp
		Temperature: 0.7,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		logrus.Errorf("Failed to generate AI response: %v", err)
		return "", fmt.Errorf("failed to generate AI response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// TranscribeAudio transcribes an audio file using OpenAI Whisper API
func (s *Service) TranscribeAudio(ctx context.Context, audioPath string) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("AI service not initialized")
	}

	// Open the audio file
	audioFile, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer audioFile.Close()

	// Create the transcription request
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: audioPath,
		Reader:   audioFile,
		Format:   openai.AudioResponseFormatText,
	}

	transcription, err := s.client.CreateTranscription(ctx, req)
	if err != nil {
		logrus.Errorf("Failed to transcribe audio: %v", err)
		return "", fmt.Errorf("failed to transcribe audio: %w", err)
	}

	return strings.TrimSpace(transcription.Text), nil
}

// needsInternetSearch checks if the user message requires internet search
func (s *Service) needsInternetSearch(message string) bool {
	// Keywords that indicate need for current information
	keywords := []string{
		"сегодня", "today", "сейчас", "now", "текущ", "current",
		"погода", "weather", "курс", "rate", "exchange",
		"новости", "news", "актуальн", "latest", "recent",
		"дата", "date", "время", "time", "какое число", "what date", "какая дата", "какой день",
		"сколько", "how much", "цена", "price", "стоимость", "cost",
		"какой сегодня", "what is today", "какое сегодня", "what today",
	}
	
	lowerMsg := strings.ToLower(message)
	for _, keyword := range keywords {
		if strings.Contains(lowerMsg, keyword) {
			logrus.Debugf("Message requires internet search (keyword: %s): %s", keyword, message)
			return true
		}
	}
	return false
}

// extractSearchQuery extracts a search query from the user message
func (s *Service) extractSearchQuery(message string) string {
	// Remove common question words and clean up
	query := strings.TrimSpace(message)
	
	// Remove question marks and common prefixes
	query = strings.TrimRight(query, "?")
	
	// Remove common question words at the start
	questionWords := []string{"какая", "какой", "какое", "какие", "как", "what", "when", "where", "who", "how"}
	for _, word := range questionWords {
		if strings.HasPrefix(strings.ToLower(query), word+" ") {
			query = strings.TrimPrefix(query, word+" ")
			break
		}
	}
	
	return strings.TrimSpace(query)
}

// TranscribeAudioFromBytes transcribes audio from bytes using OpenAI Whisper API
func (s *Service) TranscribeAudioFromBytes(ctx context.Context, audioBytes []byte, filename string) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("AI service not initialized")
	}

	// Create a reader from bytes
	reader := bytes.NewReader(audioBytes)

	// Create the transcription request
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: filename,
		Reader:   reader,
		Format:   openai.AudioResponseFormatText,
	}

	transcription, err := s.client.CreateTranscription(ctx, req)
	if err != nil {
		logrus.Errorf("Failed to transcribe audio from bytes: %v", err)
		return "", fmt.Errorf("failed to transcribe audio: %w", err)
	}

	return strings.TrimSpace(transcription.Text), nil
}

