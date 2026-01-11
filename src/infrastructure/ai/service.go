package ai

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Service struct {
	client *openai.Client
}

func NewService(apiToken string) *Service {
	if apiToken == "" {
		return nil
	}
	return &Service{
		client: openai.NewClient(apiToken),
	}
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

