package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/knowledge"
	knowledgeRepo "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/knowledge"
	openai "github.com/sashabaranov/go-openai"
)

type KnowledgeService struct {
	repo        *knowledgeRepo.SQLiteRepository
	agentService *AgentService
}

func NewKnowledgeService(repo *knowledgeRepo.SQLiteRepository, agentService *AgentService) *KnowledgeService {
	return &KnowledgeService{
		repo:        repo,
		agentService: agentService,
	}
}

func (s *KnowledgeService) UploadDocument(ctx context.Context, req knowledge.CreateDocumentRequest) (*knowledge.Document, error) {
	doc := &knowledge.Document{
		AgentID: req.AgentID,
		Name:    req.Name,
		Type:    req.Type,
		Content: req.Content,
		Size:    int64(len(req.Content)),
		Status:  "processing",
	}

	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, err
	}

	// Process document in background
	go s.processDocument(context.Background(), doc)

	return doc, nil
}

func (s *KnowledgeService) processDocument(ctx context.Context, doc *knowledge.Document) {
	// Get agent to get API key
	agent, err := s.agentService.GetAgent(ctx, doc.AgentID)
	if err != nil {
		doc.Status = "error"
		doc.Error = "Failed to get agent: " + err.Error()
		s.repo.UpdateDocument(ctx, doc)
		return
	}

	// Split content into chunks (simple chunking by paragraphs)
	chunks := s.splitIntoChunks(doc.Content, 500) // ~500 tokens per chunk

	// Create OpenAI client for embeddings
	client := openai.NewClient(agent.APIKey)

	for _, chunkContent := range chunks {
		// Get embedding from OpenAI
		embResp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Model: openai.AdaEmbeddingV2,
			Input: []string{chunkContent},
		})

		var embedding []float64
		if err == nil && len(embResp.Data) > 0 {
			embedding = make([]float64, len(embResp.Data[0].Embedding))
			for i, v := range embResp.Data[0].Embedding {
				embedding[i] = float64(v)
			}
		}

		chunk := &knowledge.Chunk{
			DocumentID: doc.ID,
			Content:    chunkContent,
			Embedding:  embedding,
			TokenCount: len(strings.Fields(chunkContent)), // Rough token estimate
		}

		if err := s.repo.CreateChunk(ctx, chunk); err != nil {
			continue // Skip failed chunks
		}
	}

	doc.Status = "ready"
	s.repo.UpdateDocument(ctx, doc)
}

func (s *KnowledgeService) splitIntoChunks(content string, maxTokens int) []string {
	var chunks []string
	
	// Split by double newlines (paragraphs)
	paragraphs := strings.Split(content, "\n\n")
	
	var currentChunk strings.Builder
	currentTokens := 0
	
	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		
		paraTokens := len(strings.Fields(para))
		
		if currentTokens+paraTokens > maxTokens && currentChunk.Len() > 0 {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
			currentTokens = 0
		}
		
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(para)
		currentTokens += paraTokens
	}
	
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}
	
	// If no chunks created (single large text), split by sentences
	if len(chunks) == 0 && len(content) > 0 {
		chunks = append(chunks, content)
	}
	
	return chunks
}

func (s *KnowledgeService) GetDocuments(ctx context.Context, agentID string) ([]*knowledge.Document, error) {
	return s.repo.GetDocumentsByAgentID(ctx, agentID)
}

func (s *KnowledgeService) DeleteDocument(ctx context.Context, id string) error {
	// Delete chunks first
	if err := s.repo.DeleteChunksByDocumentID(ctx, id); err != nil {
		return err
	}
	return s.repo.DeleteDocument(ctx, id)
}

func (s *KnowledgeService) Search(ctx context.Context, req knowledge.SearchRequest) ([]knowledge.SearchResult, error) {
	// Get agent for API key
	agent, err := s.agentService.GetAgent(ctx, req.AgentID)
	if err != nil {
		return nil, err
	}

	// Get query embedding
	client := openai.NewClient(agent.APIKey)
	embResp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.AdaEmbeddingV2,
		Input: []string{req.Query},
	})
	if err != nil {
		return nil, err
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	queryEmbedding := make([]float64, len(embResp.Data[0].Embedding))
	for i, v := range embResp.Data[0].Embedding {
		queryEmbedding[i] = float64(v)
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 5
	}

	return s.repo.SearchChunks(ctx, req.AgentID, queryEmbedding, topK)
}

func (s *KnowledgeService) GetRelevantContext(ctx context.Context, agentID string, query string) (string, error) {
	results, err := s.Search(ctx, knowledge.SearchRequest{
		AgentID: agentID,
		Query:   query,
		TopK:    3,
	})
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", nil
	}

	var contextBuilder strings.Builder
	contextBuilder.WriteString("Relevant information from knowledge base:\n\n")
	
	for i, result := range results {
		if result.Score < 0.7 { // Only include relevant results
			continue
		}
		contextBuilder.WriteString(fmt.Sprintf("--- From %s ---\n%s\n\n", result.DocName, result.Content))
		if i >= 2 { // Max 3 chunks
			break
		}
	}

	return contextBuilder.String(), nil
}

