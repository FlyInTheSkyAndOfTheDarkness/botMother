package calendar

import (
	"context"
	"database/sql"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/calendar"
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
	CREATE TABLE IF NOT EXISTS calendar_credentials (
		id TEXT PRIMARY KEY,
		agent_id TEXT UNIQUE NOT NULL,
		name TEXT,
		client_id TEXT,
		client_secret TEXT,
		access_token TEXT,
		refresh_token TEXT,
		token_expiry DATETIME,
		calendar_id TEXT DEFAULT 'primary',
		is_connected INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_calendar_creds_agent ON calendar_credentials(agent_id);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) CreateCredential(ctx context.Context, cred *calendar.CalendarCredential) error {
	if cred.ID == "" {
		cred.ID = uuid.New().String()
	}
	if cred.CreatedAt.IsZero() {
		cred.CreatedAt = time.Now()
	}
	cred.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO calendar_credentials (id, agent_id, name, client_id, client_secret, 
		                                   access_token, refresh_token, token_expiry, 
		                                   calendar_id, is_connected, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cred.ID, cred.AgentID, cred.Name, cred.ClientID, cred.ClientSecret,
		cred.AccessToken, cred.RefreshToken, cred.TokenExpiry,
		cred.CalendarID, cred.IsConnected, cred.CreatedAt, cred.UpdatedAt)
	return err
}

func (r *SQLiteRepository) GetCredential(ctx context.Context, id string) (*calendar.CalendarCredential, error) {
	cred := &calendar.CalendarCredential{}
	var tokenExpiry sql.NullTime
	
	err := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, name, client_id, client_secret, access_token, 
		        refresh_token, token_expiry, calendar_id, is_connected, 
		        created_at, updated_at 
		 FROM calendar_credentials WHERE id = ?`, id).
		Scan(&cred.ID, &cred.AgentID, &cred.Name, &cred.ClientID, &cred.ClientSecret,
			&cred.AccessToken, &cred.RefreshToken, &tokenExpiry,
			&cred.CalendarID, &cred.IsConnected, &cred.CreatedAt, &cred.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	if tokenExpiry.Valid {
		cred.TokenExpiry = tokenExpiry.Time
	}
	return cred, nil
}

func (r *SQLiteRepository) GetCredentialByAgentID(ctx context.Context, agentID string) (*calendar.CalendarCredential, error) {
	cred := &calendar.CalendarCredential{}
	var tokenExpiry sql.NullTime
	
	err := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, name, client_id, client_secret, access_token, 
		        refresh_token, token_expiry, calendar_id, is_connected, 
		        created_at, updated_at 
		 FROM calendar_credentials WHERE agent_id = ?`, agentID).
		Scan(&cred.ID, &cred.AgentID, &cred.Name, &cred.ClientID, &cred.ClientSecret,
			&cred.AccessToken, &cred.RefreshToken, &tokenExpiry,
			&cred.CalendarID, &cred.IsConnected, &cred.CreatedAt, &cred.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	if tokenExpiry.Valid {
		cred.TokenExpiry = tokenExpiry.Time
	}
	return cred, nil
}

func (r *SQLiteRepository) UpdateCredential(ctx context.Context, cred *calendar.CalendarCredential) error {
	cred.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx,
		`UPDATE calendar_credentials SET 
		 name = ?, client_id = ?, client_secret = ?, access_token = ?,
		 refresh_token = ?, token_expiry = ?, calendar_id = ?, 
		 is_connected = ?, updated_at = ?
		 WHERE id = ?`,
		cred.Name, cred.ClientID, cred.ClientSecret, cred.AccessToken,
		cred.RefreshToken, cred.TokenExpiry, cred.CalendarID,
		cred.IsConnected, cred.UpdatedAt, cred.ID)
	return err
}

func (r *SQLiteRepository) DeleteCredential(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM calendar_credentials WHERE id = ?`, id)
	return err
}

