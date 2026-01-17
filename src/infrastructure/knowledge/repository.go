package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/knowledge"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := &SQLiteRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *SQLiteRepository) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS documents (
		id TEXT PRIMARY KEY,
		agent_id TEXT NOT NULL,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		content TEXT,
		size INTEGER DEFAULT 0,
		status TEXT DEFAULT 'processing',
		error TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS chunks (
		id TEXT PRIMARY KEY,
		document_id TEXT NOT NULL,
		content TEXT NOT NULL,
		embedding TEXT,
		metadata TEXT,
		token_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_documents_agent ON documents(agent_id);
	CREATE INDEX IF NOT EXISTS idx_chunks_document ON chunks(document_id);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) CreateDocument(ctx context.Context, doc *knowledge.Document) error {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now()
	}
	doc.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO documents (id, agent_id, name, type, content, size, status, error, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		doc.ID, doc.AgentID, doc.Name, doc.Type, doc.Content, doc.Size, doc.Status, doc.Error, doc.CreatedAt, doc.UpdatedAt)
	return err
}

func (r *SQLiteRepository) GetDocument(ctx context.Context, id string) (*knowledge.Document, error) {
	doc := &knowledge.Document{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, name, type, content, size, status, error, created_at, updated_at 
		 FROM documents WHERE id = ?`, id).
		Scan(&doc.ID, &doc.AgentID, &doc.Name, &doc.Type, &doc.Content, &doc.Size, &doc.Status, &doc.Error, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *SQLiteRepository) GetDocumentsByAgentID(ctx context.Context, agentID string) ([]*knowledge.Document, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, agent_id, name, type, size, status, error, created_at, updated_at 
		 FROM documents WHERE agent_id = ? ORDER BY created_at DESC`, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*knowledge.Document
	for rows.Next() {
		doc := &knowledge.Document{}
		err := rows.Scan(&doc.ID, &doc.AgentID, &doc.Name, &doc.Type, &doc.Size, &doc.Status, &doc.Error, &doc.CreatedAt, &doc.UpdatedAt)
		if err != nil {
			continue
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func (r *SQLiteRepository) UpdateDocument(ctx context.Context, doc *knowledge.Document) error {
	doc.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE documents SET name = ?, content = ?, size = ?, status = ?, error = ?, updated_at = ? WHERE id = ?`,
		doc.Name, doc.Content, doc.Size, doc.Status, doc.Error, doc.UpdatedAt, doc.ID)
	return err
}

func (r *SQLiteRepository) DeleteDocument(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM documents WHERE id = ?`, id)
	return err
}

func (r *SQLiteRepository) CreateChunk(ctx context.Context, chunk *knowledge.Chunk) error {
	if chunk.ID == "" {
		chunk.ID = uuid.New().String()
	}
	if chunk.CreatedAt.IsZero() {
		chunk.CreatedAt = time.Now()
	}

	embeddingJSON := ""
	if chunk.Embedding != nil {
		data, _ := json.Marshal(chunk.Embedding)
		embeddingJSON = string(data)
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO chunks (id, document_id, content, embedding, metadata, token_count, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		chunk.ID, chunk.DocumentID, chunk.Content, embeddingJSON, chunk.Metadata, chunk.TokenCount, chunk.CreatedAt)
	return err
}

func (r *SQLiteRepository) GetChunksByDocumentID(ctx context.Context, docID string) ([]*knowledge.Chunk, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, document_id, content, embedding, metadata, token_count, created_at 
		 FROM chunks WHERE document_id = ?`, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []*knowledge.Chunk
	for rows.Next() {
		chunk := &knowledge.Chunk{}
		var embeddingJSON string
		err := rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.Content, &embeddingJSON, &chunk.Metadata, &chunk.TokenCount, &chunk.CreatedAt)
		if err != nil {
			continue
		}
		if embeddingJSON != "" {
			json.Unmarshal([]byte(embeddingJSON), &chunk.Embedding)
		}
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

func (r *SQLiteRepository) DeleteChunksByDocumentID(ctx context.Context, docID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM chunks WHERE document_id = ?`, docID)
	return err
}

func (r *SQLiteRepository) SearchChunks(ctx context.Context, agentID string, queryEmbedding []float64, topK int) ([]knowledge.SearchResult, error) {
	// Get all chunks for agent's documents
	rows, err := r.db.QueryContext(ctx,
		`SELECT c.id, c.document_id, c.content, c.embedding, d.name
		 FROM chunks c
		 JOIN documents d ON c.document_id = d.id
		 WHERE d.agent_id = ? AND d.status = 'ready'`, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type scoredResult struct {
		result knowledge.SearchResult
		score  float64
	}
	var scored []scoredResult

	for rows.Next() {
		var chunkID, docID, content, embeddingJSON, docName string
		err := rows.Scan(&chunkID, &docID, &content, &embeddingJSON, &docName)
		if err != nil {
			continue
		}

		var embedding []float64
		if embeddingJSON != "" {
			json.Unmarshal([]byte(embeddingJSON), &embedding)
		}

		// Calculate cosine similarity
		score := cosineSimilarity(queryEmbedding, embedding)
		scored = append(scored, scoredResult{
			result: knowledge.SearchResult{
				ChunkID:    chunkID,
				DocumentID: docID,
				DocName:    docName,
				Content:    content,
				Score:      score,
			},
			score: score,
		})
	}

	// Sort by score descending
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Take top K
	var results []knowledge.SearchResult
	for i := 0; i < len(scored) && i < topK; i++ {
		results = append(results, scored[i].result)
	}

	return results, nil
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}




