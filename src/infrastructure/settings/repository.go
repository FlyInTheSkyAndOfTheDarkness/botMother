package settings

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/settings"
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
	CREATE TABLE IF NOT EXISTS agent_settings (
		id TEXT PRIMARY KEY,
		agent_id TEXT UNIQUE NOT NULL,
		working_hours TEXT,
		translation TEXT,
		follow_up TEXT,
		sentiment TEXT,
		max_tokens_per_msg INTEGER DEFAULT 500,
		temperature REAL DEFAULT 0.7,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS broadcasts (
		id TEXT PRIMARY KEY,
		agent_id TEXT NOT NULL,
		integration_type TEXT NOT NULL,
		message TEXT NOT NULL,
		media_url TEXT,
		recipients TEXT,
		status TEXT DEFAULT 'pending',
		total_recipients INTEGER DEFAULT 0,
		sent_count INTEGER DEFAULT 0,
		failed_count INTEGER DEFAULT 0,
		scheduled_at DATETIME,
		started_at DATETIME,
		completed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_agent_settings_agent ON agent_settings(agent_id);
	CREATE INDEX IF NOT EXISTS idx_broadcasts_agent ON broadcasts(agent_id);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) GetAgentSettings(ctx context.Context, agentID string) (*settings.AgentSettings, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, working_hours, translation, follow_up, sentiment, 
		        max_tokens_per_msg, temperature, created_at, updated_at 
		 FROM agent_settings WHERE agent_id = ?`, agentID)

	s := &settings.AgentSettings{}
	var workingHoursJSON, translationJSON, followUpJSON, sentimentJSON sql.NullString

	err := row.Scan(&s.ID, &s.AgentID, &workingHoursJSON, &translationJSON, 
		&followUpJSON, &sentimentJSON, &s.MaxTokensPerMsg, &s.Temperature, 
		&s.CreatedAt, &s.UpdatedAt)
	
	if err == sql.ErrNoRows {
		// Return default settings if none exist
		return settings.DefaultAgentSettings(agentID), nil
	}
	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if workingHoursJSON.Valid {
		json.Unmarshal([]byte(workingHoursJSON.String), &s.WorkingHours)
	}
	if translationJSON.Valid {
		json.Unmarshal([]byte(translationJSON.String), &s.Translation)
	}
	if followUpJSON.Valid {
		json.Unmarshal([]byte(followUpJSON.String), &s.FollowUp)
	}
	if sentimentJSON.Valid {
		json.Unmarshal([]byte(sentimentJSON.String), &s.Sentiment)
	}

	return s, nil
}

func (r *SQLiteRepository) SaveAgentSettings(ctx context.Context, s *settings.AgentSettings) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	s.UpdatedAt = time.Now()

	workingHoursJSON, _ := json.Marshal(s.WorkingHours)
	translationJSON, _ := json.Marshal(s.Translation)
	followUpJSON, _ := json.Marshal(s.FollowUp)
	sentimentJSON, _ := json.Marshal(s.Sentiment)

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO agent_settings (id, agent_id, working_hours, translation, follow_up, sentiment, 
		                             max_tokens_per_msg, temperature, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(agent_id) DO UPDATE SET
		 	working_hours = excluded.working_hours,
		 	translation = excluded.translation,
		 	follow_up = excluded.follow_up,
		 	sentiment = excluded.sentiment,
		 	max_tokens_per_msg = excluded.max_tokens_per_msg,
		 	temperature = excluded.temperature,
		 	updated_at = excluded.updated_at`,
		s.ID, s.AgentID, string(workingHoursJSON), string(translationJSON),
		string(followUpJSON), string(sentimentJSON), s.MaxTokensPerMsg, 
		s.Temperature, s.CreatedAt, s.UpdatedAt)

	return err
}

func (r *SQLiteRepository) CreateBroadcast(ctx context.Context, b *settings.BroadcastMessage) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	b.CreatedAt = time.Now()

	recipientsJSON, _ := json.Marshal(b.Recipients)

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO broadcasts (id, agent_id, integration_type, message, media_url, recipients,
		                         status, total_recipients, scheduled_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.AgentID, b.IntegrationType, b.Message, b.MediaURL, 
		string(recipientsJSON), b.Status, b.TotalRecipients, b.ScheduledAt, b.CreatedAt)

	return err
}

func (r *SQLiteRepository) GetBroadcast(ctx context.Context, id string) (*settings.BroadcastMessage, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, integration_type, message, media_url, recipients,
		        status, total_recipients, sent_count, failed_count,
		        scheduled_at, started_at, completed_at, created_at
		 FROM broadcasts WHERE id = ?`, id)

	b := &settings.BroadcastMessage{}
	var recipientsJSON string
	var scheduledAt, startedAt, completedAt sql.NullTime

	err := row.Scan(&b.ID, &b.AgentID, &b.IntegrationType, &b.Message, &b.MediaURL,
		&recipientsJSON, &b.Status, &b.TotalRecipients, &b.SentCount, &b.FailedCount,
		&scheduledAt, &startedAt, &completedAt, &b.CreatedAt)
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(recipientsJSON), &b.Recipients)
	if scheduledAt.Valid {
		b.ScheduledAt = &scheduledAt.Time
	}
	if startedAt.Valid {
		b.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		b.CompletedAt = &completedAt.Time
	}

	return b, nil
}

func (r *SQLiteRepository) GetBroadcastsByAgentID(ctx context.Context, agentID string) ([]*settings.BroadcastMessage, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, agent_id, integration_type, message, status, 
		        total_recipients, sent_count, failed_count, scheduled_at, created_at
		 FROM broadcasts WHERE agent_id = ? ORDER BY created_at DESC`, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var broadcasts []*settings.BroadcastMessage
	for rows.Next() {
		b := &settings.BroadcastMessage{}
		var scheduledAt sql.NullTime
		
		err := rows.Scan(&b.ID, &b.AgentID, &b.IntegrationType, &b.Message, &b.Status,
			&b.TotalRecipients, &b.SentCount, &b.FailedCount, &scheduledAt, &b.CreatedAt)
		if err != nil {
			continue
		}
		if scheduledAt.Valid {
			b.ScheduledAt = &scheduledAt.Time
		}
		broadcasts = append(broadcasts, b)
	}

	return broadcasts, nil
}

func (r *SQLiteRepository) UpdateBroadcast(ctx context.Context, b *settings.BroadcastMessage) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE broadcasts SET status = ?, sent_count = ?, failed_count = ?,
		                       started_at = ?, completed_at = ?
		 WHERE id = ?`,
		b.Status, b.SentCount, b.FailedCount, b.StartedAt, b.CompletedAt, b.ID)
	return err
}


