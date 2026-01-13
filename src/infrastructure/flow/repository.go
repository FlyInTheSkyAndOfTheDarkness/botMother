package flow

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/flow"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	repo := &SQLiteRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return repo, nil
}

func (r *SQLiteRepository) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS flows (
			id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			is_active INTEGER DEFAULT 1,
			nodes TEXT DEFAULT '[]',
			edges TEXT DEFAULT '[]',
			variables TEXT DEFAULT '[]',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS credentials (
			id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			config TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_flows_agent_id ON flows(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_credentials_agent_id ON credentials(agent_id)`,
	}

	for _, query := range queries {
		if _, err := r.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// === Flow CRUD ===

func (r *SQLiteRepository) CreateFlow(ctx context.Context, f *flow.Flow) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	now := time.Now()
	f.CreatedAt = now
	f.UpdatedAt = now

	nodesJSON, err := json.Marshal(f.Nodes)
	if err != nil {
		return fmt.Errorf("failed to marshal nodes: %w", err)
	}

	edgesJSON, err := json.Marshal(f.Edges)
	if err != nil {
		return fmt.Errorf("failed to marshal edges: %w", err)
	}

	variablesJSON, err := json.Marshal(f.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO flows (id, agent_id, name, description, is_active, nodes, edges, variables, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		f.ID, f.AgentID, f.Name, f.Description, f.IsActive, string(nodesJSON), string(edgesJSON), string(variablesJSON), f.CreatedAt, f.UpdatedAt,
	)
	return err
}

func (r *SQLiteRepository) GetFlowByID(ctx context.Context, id string) (*flow.Flow, error) {
	f := &flow.Flow{}
	var nodesJSON, edgesJSON, variablesJSON string

	err := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, name, description, is_active, nodes, edges, variables, created_at, updated_at
		FROM flows WHERE id = ?`, id,
	).Scan(&f.ID, &f.AgentID, &f.Name, &f.Description, &f.IsActive, &nodesJSON, &edgesJSON, &variablesJSON, &f.CreatedAt, &f.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flow not found")
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(nodesJSON), &f.Nodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}
	if err := json.Unmarshal([]byte(edgesJSON), &f.Edges); err != nil {
		return nil, fmt.Errorf("failed to unmarshal edges: %w", err)
	}
	if err := json.Unmarshal([]byte(variablesJSON), &f.Variables); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}

	return f, nil
}

func (r *SQLiteRepository) GetFlowsByAgentID(ctx context.Context, agentID string) ([]*flow.Flow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, agent_id, name, description, is_active, nodes, edges, variables, created_at, updated_at
		FROM flows WHERE agent_id = ? ORDER BY created_at DESC`, agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flows []*flow.Flow
	for rows.Next() {
		f := &flow.Flow{}
		var nodesJSON, edgesJSON, variablesJSON string

		if err := rows.Scan(&f.ID, &f.AgentID, &f.Name, &f.Description, &f.IsActive, &nodesJSON, &edgesJSON, &variablesJSON, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(nodesJSON), &f.Nodes); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(edgesJSON), &f.Edges); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(variablesJSON), &f.Variables); err != nil {
			return nil, err
		}

		flows = append(flows, f)
	}
	return flows, rows.Err()
}

func (r *SQLiteRepository) UpdateFlow(ctx context.Context, f *flow.Flow) error {
	f.UpdatedAt = time.Now()

	nodesJSON, err := json.Marshal(f.Nodes)
	if err != nil {
		return fmt.Errorf("failed to marshal nodes: %w", err)
	}

	edgesJSON, err := json.Marshal(f.Edges)
	if err != nil {
		return fmt.Errorf("failed to marshal edges: %w", err)
	}

	variablesJSON, err := json.Marshal(f.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`UPDATE flows SET name=?, description=?, is_active=?, nodes=?, edges=?, variables=?, updated_at=?
		WHERE id=?`,
		f.Name, f.Description, f.IsActive, string(nodesJSON), string(edgesJSON), string(variablesJSON), f.UpdatedAt, f.ID,
	)
	return err
}

func (r *SQLiteRepository) DeleteFlow(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM flows WHERE id = ?`, id)
	return err
}

// === Credential CRUD ===

func (r *SQLiteRepository) CreateCredential(ctx context.Context, c *flow.Credential) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO credentials (id, agent_id, name, type, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.AgentID, c.Name, c.Type, c.Config, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *SQLiteRepository) GetCredentialByID(ctx context.Context, id string) (*flow.Credential, error) {
	c := &flow.Credential{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, name, type, config, created_at, updated_at
		FROM credentials WHERE id = ?`, id,
	).Scan(&c.ID, &c.AgentID, &c.Name, &c.Type, &c.Config, &c.CreatedAt, &c.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("credential not found")
	}
	return c, err
}

func (r *SQLiteRepository) GetCredentialsByAgentID(ctx context.Context, agentID string) ([]*flow.Credential, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, agent_id, name, type, config, created_at, updated_at
		FROM credentials WHERE agent_id = ? ORDER BY created_at DESC`, agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []*flow.Credential
	for rows.Next() {
		c := &flow.Credential{}
		if err := rows.Scan(&c.ID, &c.AgentID, &c.Name, &c.Type, &c.Config, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		credentials = append(credentials, c)
	}
	return credentials, rows.Err()
}

func (r *SQLiteRepository) UpdateCredential(ctx context.Context, c *flow.Credential) error {
	c.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE credentials SET name=?, config=?, updated_at=? WHERE id=?`,
		c.Name, c.Config, c.UpdatedAt, c.ID,
	)
	return err
}

func (r *SQLiteRepository) DeleteCredential(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM credentials WHERE id = ?`, id)
	return err
}


