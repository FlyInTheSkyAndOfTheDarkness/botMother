package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/agent"
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
		`CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			api_key TEXT NOT NULL,
			serp_api_key TEXT DEFAULT '',
			model TEXT NOT NULL DEFAULT 'gpt-4o-mini',
			system_prompt TEXT NOT NULL,
			welcome_message TEXT DEFAULT '',
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Migration for existing tables
		`ALTER TABLE agents ADD COLUMN serp_api_key TEXT DEFAULT ''`,
		`CREATE TABLE IF NOT EXISTS integrations (
			id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL,
			type TEXT NOT NULL,
			is_connected INTEGER DEFAULT 0,
			config TEXT DEFAULT '{}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL,
			integration_id TEXT NOT NULL,
			remote_jid TEXT NOT NULL,
			is_first_reply INTEGER DEFAULT 0,
			is_manual_mode INTEGER DEFAULT 0,
			notes TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE,
			FOREIGN KEY (integration_id) REFERENCES integrations(id) ON DELETE CASCADE,
			UNIQUE(agent_id, integration_id, remote_jid)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_integrations_agent_id ON integrations(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conversations_lookup ON conversations(agent_id, integration_id, remote_jid)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id, timestamp DESC)`,
	}

	for _, query := range queries {
		if _, err := r.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	// Safe migrations for existing tables (ignore errors if columns already exist)
	safeMigrations := []string{
		`ALTER TABLE conversations ADD COLUMN is_manual_mode INTEGER DEFAULT 0`,
		`ALTER TABLE conversations ADD COLUMN notes TEXT DEFAULT ''`,
	}
	for _, query := range safeMigrations {
		r.db.Exec(query) // Ignore errors (column may already exist)
	}

	return nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// Agent CRUD

func (r *SQLiteRepository) Create(ctx context.Context, a *agent.Agent) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO agents (id, name, description, api_key, serp_api_key, model, system_prompt, welcome_message, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.Name, a.Description, a.APIKey, a.SerpAPIKey, a.Model, a.SystemPrompt, a.WelcomeMessage, a.IsActive, a.CreatedAt, a.UpdatedAt,
	)
	return err
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id string) (*agent.Agent, error) {
	a := &agent.Agent{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, api_key, COALESCE(serp_api_key, ''), model, system_prompt, welcome_message, is_active, created_at, updated_at
		FROM agents WHERE id = ?`, id,
	).Scan(&a.ID, &a.Name, &a.Description, &a.APIKey, &a.SerpAPIKey, &a.Model, &a.SystemPrompt, &a.WelcomeMessage, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent not found")
	}
	return a, err
}

func (r *SQLiteRepository) GetAll(ctx context.Context) ([]*agent.Agent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, description, api_key, COALESCE(serp_api_key, ''), model, system_prompt, welcome_message, is_active, created_at, updated_at
		FROM agents ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*agent.Agent
	for rows.Next() {
		a := &agent.Agent{}
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.APIKey, &a.SerpAPIKey, &a.Model, &a.SystemPrompt, &a.WelcomeMessage, &a.IsActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		agents = append(agents, a)
	}
	return agents, rows.Err()
}

func (r *SQLiteRepository) Update(ctx context.Context, a *agent.Agent) error {
	a.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE agents SET name=?, description=?, api_key=?, serp_api_key=?, model=?, system_prompt=?, welcome_message=?, is_active=?, updated_at=?
		WHERE id=?`,
		a.Name, a.Description, a.APIKey, a.SerpAPIKey, a.Model, a.SystemPrompt, a.WelcomeMessage, a.IsActive, a.UpdatedAt, a.ID,
	)
	return err
}

func (r *SQLiteRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM agents WHERE id = ?`, id)
	return err
}

// Integration CRUD

func (r *SQLiteRepository) CreateIntegration(ctx context.Context, i *agent.Integration) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	now := time.Now()
	i.CreatedAt = now
	i.UpdatedAt = now

	if i.Config == "" {
		i.Config = "{}"
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO integrations (id, agent_id, type, is_connected, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		i.ID, i.AgentID, i.Type, i.IsConnected, i.Config, i.CreatedAt, i.UpdatedAt,
	)
	return err
}

func (r *SQLiteRepository) GetIntegrationsByAgentID(ctx context.Context, agentID string) ([]*agent.Integration, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, agent_id, type, is_connected, config, created_at, updated_at
		FROM integrations WHERE agent_id = ?`, agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []*agent.Integration
	for rows.Next() {
		i := &agent.Integration{}
		if err := rows.Scan(&i.ID, &i.AgentID, &i.Type, &i.IsConnected, &i.Config, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		integrations = append(integrations, i)
	}
	return integrations, rows.Err()
}

func (r *SQLiteRepository) GetIntegrationByID(ctx context.Context, id string) (*agent.Integration, error) {
	i := &agent.Integration{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, type, is_connected, config, created_at, updated_at
		FROM integrations WHERE id = ?`, id,
	).Scan(&i.ID, &i.AgentID, &i.Type, &i.IsConnected, &i.Config, &i.CreatedAt, &i.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("integration not found")
	}
	return i, err
}

func (r *SQLiteRepository) GetIntegrationByAgentAndType(ctx context.Context, agentID, integrationType string) (*agent.Integration, error) {
	i := &agent.Integration{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, type, is_connected, config, created_at, updated_at
		FROM integrations WHERE agent_id = ? AND type = ?`, agentID, integrationType,
	).Scan(&i.ID, &i.AgentID, &i.Type, &i.IsConnected, &i.Config, &i.CreatedAt, &i.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error here
	}
	return i, err
}

func (r *SQLiteRepository) UpdateIntegration(ctx context.Context, i *agent.Integration) error {
	i.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE integrations SET is_connected=?, config=?, updated_at=? WHERE id=?`,
		i.IsConnected, i.Config, i.UpdatedAt, i.ID,
	)
	return err
}

func (r *SQLiteRepository) DeleteIntegration(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM integrations WHERE id = ?`, id)
	return err
}

// Conversation management

func (r *SQLiteRepository) GetOrCreateConversation(ctx context.Context, agentID, integrationID, remoteJID string) (*agent.Conversation, error) {
	c := &agent.Conversation{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, integration_id, remote_jid, is_first_reply, COALESCE(is_manual_mode, 0), COALESCE(notes, ''), created_at, updated_at
		FROM conversations WHERE agent_id = ? AND integration_id = ? AND remote_jid = ?`,
		agentID, integrationID, remoteJID,
	).Scan(&c.ID, &c.AgentID, &c.IntegrationID, &c.RemoteJID, &c.IsFirstReply, &c.IsManualMode, &c.Notes, &c.CreatedAt, &c.UpdatedAt)

	if err == sql.ErrNoRows {
		// Create new conversation
		c = &agent.Conversation{
			ID:            uuid.New().String(),
			AgentID:       agentID,
			IntegrationID: integrationID,
			RemoteJID:     remoteJID,
			IsFirstReply:  false,
			IsManualMode:  false,
			Notes:         "",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		_, err = r.db.ExecContext(ctx,
			`INSERT INTO conversations (id, agent_id, integration_id, remote_jid, is_first_reply, is_manual_mode, notes, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			c.ID, c.AgentID, c.IntegrationID, c.RemoteJID, c.IsFirstReply, c.IsManualMode, c.Notes, c.CreatedAt, c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return c, nil
}

func (r *SQLiteRepository) UpdateConversation(ctx context.Context, c *agent.Conversation) error {
	c.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET is_first_reply=?, is_manual_mode=?, notes=?, updated_at=? WHERE id=?`,
		c.IsFirstReply, c.IsManualMode, c.Notes, c.UpdatedAt, c.ID,
	)
	return err
}

// Message history

func (r *SQLiteRepository) AddMessage(ctx context.Context, m *agent.Message) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now()
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO messages (id, conversation_id, role, content, timestamp)
		VALUES (?, ?, ?, ?, ?)`,
		m.ID, m.ConversationID, m.Role, m.Content, m.Timestamp,
	)
	return err
}

func (r *SQLiteRepository) GetRecentMessages(ctx context.Context, conversationID string, limit int) ([]*agent.Message, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, conversation_id, role, content, timestamp
		FROM messages WHERE conversation_id = ?
		ORDER BY timestamp DESC LIMIT ?`, conversationID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*agent.Message
	for rows.Next() {
		m := &agent.Message{}
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, rows.Err()
}

// GetAllConversations returns all conversations with optional filtering
func (r *SQLiteRepository) GetAllConversations(ctx context.Context) ([]*agent.Conversation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, agent_id, integration_id, remote_jid, is_first_reply, COALESCE(is_manual_mode, 0), COALESCE(notes, ''), created_at, updated_at
		FROM conversations ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []*agent.Conversation
	for rows.Next() {
		c := &agent.Conversation{}
		if err := rows.Scan(&c.ID, &c.AgentID, &c.IntegrationID, &c.RemoteJID, &c.IsFirstReply, &c.IsManualMode, &c.Notes, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}
	return conversations, rows.Err()
}

// GetConversationByID returns a single conversation
func (r *SQLiteRepository) GetConversationByID(ctx context.Context, id string) (*agent.Conversation, error) {
	c := &agent.Conversation{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, integration_id, remote_jid, is_first_reply, COALESCE(is_manual_mode, 0), COALESCE(notes, ''), created_at, updated_at
		FROM conversations WHERE id = ?`, id,
	).Scan(&c.ID, &c.AgentID, &c.IntegrationID, &c.RemoteJID, &c.IsFirstReply, &c.IsManualMode, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// SetConversationManualMode sets the manual mode for a conversation
func (r *SQLiteRepository) SetConversationManualMode(ctx context.Context, conversationID string, isManual bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET is_manual_mode = ?, updated_at = ? WHERE id = ?`,
		isManual, time.Now(), conversationID,
	)
	return err
}

// UpdateConversationNotes updates the notes for a conversation
func (r *SQLiteRepository) UpdateConversationNotes(ctx context.Context, conversationID, notes string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET notes = ?, updated_at = ? WHERE id = ?`,
		notes, time.Now(), conversationID,
	)
	return err
}

// GetMessagesForConversation returns all messages for a conversation
func (r *SQLiteRepository) GetMessagesForConversation(ctx context.Context, conversationID string) ([]*agent.Message, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, conversation_id, role, content, timestamp
		FROM messages WHERE conversation_id = ?
		ORDER BY timestamp ASC`, conversationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*agent.Message
	for rows.Next() {
		m := &agent.Message{}
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

// GetLastMessageForConversation returns the last message
func (r *SQLiteRepository) GetLastMessageForConversation(ctx context.Context, conversationID string) (*agent.Message, error) {
	m := &agent.Message{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, conversation_id, role, content, timestamp
		FROM messages WHERE conversation_id = ?
		ORDER BY timestamp DESC LIMIT 1`, conversationID,
	).Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.Timestamp)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Helper to parse integration config
func ParseWhatsAppConfig(configJSON string) (*agent.WhatsAppConfig, error) {
	var config agent.WhatsAppConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func ParseTelegramConfig(configJSON string) (*agent.TelegramConfig, error) {
	var config agent.TelegramConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

