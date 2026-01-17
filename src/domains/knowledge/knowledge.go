package knowledge

import (
	"context"
	"time"
)

// Document represents a knowledge base document
type Document struct {
	ID          string    `json:"id"`
	AgentID     string    `json:"agent_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // pdf, txt, md, docx, url
	Content     string    `json:"content,omitempty"`
	Chunks      []Chunk   `json:"chunks,omitempty"`
	Size        int64     `json:"size"`
	Status      string    `json:"status"` // processing, ready, error
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Chunk represents a document chunk for embeddings
type Chunk struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	Content    string    `json:"content"`
	Embedding  []float64 `json:"embedding,omitempty"`
	Metadata   string    `json:"metadata,omitempty"` // JSON: page number, section, etc
	TokenCount int       `json:"token_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// SearchResult represents a RAG search result
type SearchResult struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	DocName    string  `json:"document_name"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
}

// CreateDocumentRequest represents request to create a document
type CreateDocumentRequest struct {
	AgentID string `json:"agent_id" validate:"required"`
	Name    string `json:"name" validate:"required"`
	Type    string `json:"type" validate:"required"`
	Content string `json:"content,omitempty"` // For text/markdown
	URL     string `json:"url,omitempty"`     // For URL type
}

// SearchRequest represents a RAG search request
type SearchRequest struct {
	AgentID string `json:"agent_id" validate:"required"`
	Query   string `json:"query" validate:"required"`
	TopK    int    `json:"top_k,omitempty"` // Default 5
}

// IKnowledgeRepository defines database operations for knowledge base
type IKnowledgeRepository interface {
	CreateDocument(ctx context.Context, doc *Document) error
	GetDocument(ctx context.Context, id string) (*Document, error)
	GetDocumentsByAgentID(ctx context.Context, agentID string) ([]*Document, error)
	UpdateDocument(ctx context.Context, doc *Document) error
	DeleteDocument(ctx context.Context, id string) error
	
	CreateChunk(ctx context.Context, chunk *Chunk) error
	GetChunksByDocumentID(ctx context.Context, docID string) ([]*Chunk, error)
	DeleteChunksByDocumentID(ctx context.Context, docID string) error
	SearchChunks(ctx context.Context, agentID string, embedding []float64, topK int) ([]SearchResult, error)
}

// IKnowledgeService defines business logic for knowledge base
type IKnowledgeService interface {
	UploadDocument(ctx context.Context, req CreateDocumentRequest) (*Document, error)
	GetDocuments(ctx context.Context, agentID string) ([]*Document, error)
	DeleteDocument(ctx context.Context, id string) error
	Search(ctx context.Context, req SearchRequest) ([]SearchResult, error)
	GetRelevantContext(ctx context.Context, agentID string, query string) (string, error)
}




