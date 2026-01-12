package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/flow"
	flowRepo "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/flow"
	_ "github.com/lib/pq"
)

type FlowService struct {
	repo *flowRepo.SQLiteRepository
}

func NewFlowService(repo *flowRepo.SQLiteRepository) *FlowService {
	return &FlowService{repo: repo}
}

// === Flow Operations ===

func (s *FlowService) CreateFlow(ctx context.Context, req flow.CreateFlowRequest) (*flow.FlowResponse, error) {
	f := &flow.Flow{
		AgentID:     req.AgentID,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
		Nodes:       req.Nodes,
		Edges:       req.Edges,
		Variables:   []flow.Variable{},
	}

	if f.Nodes == nil {
		f.Nodes = []flow.Node{}
	}
	if f.Edges == nil {
		f.Edges = []flow.Edge{}
	}

	if err := s.repo.CreateFlow(ctx, f); err != nil {
		return nil, fmt.Errorf("failed to create flow: %w", err)
	}

	return s.flowToResponse(f), nil
}

func (s *FlowService) GetFlow(ctx context.Context, id string) (*flow.FlowResponse, error) {
	f, err := s.repo.GetFlowByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.flowToResponse(f), nil
}

func (s *FlowService) GetFlowsByAgent(ctx context.Context, agentID string) ([]*flow.FlowResponse, error) {
	flows, err := s.repo.GetFlowsByAgentID(ctx, agentID)
	if err != nil {
		return nil, err
	}

	var responses []*flow.FlowResponse
	for _, f := range flows {
		responses = append(responses, s.flowToResponse(f))
	}
	return responses, nil
}

func (s *FlowService) UpdateFlow(ctx context.Context, id string, req flow.UpdateFlowRequest) (*flow.FlowResponse, error) {
	f, err := s.repo.GetFlowByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		f.Name = *req.Name
	}
	if req.Description != nil {
		f.Description = *req.Description
	}
	if req.IsActive != nil {
		f.IsActive = *req.IsActive
	}
	if req.Nodes != nil {
		f.Nodes = req.Nodes
	}
	if req.Edges != nil {
		f.Edges = req.Edges
	}
	if req.Variables != nil {
		f.Variables = req.Variables
	}

	if err := s.repo.UpdateFlow(ctx, f); err != nil {
		return nil, fmt.Errorf("failed to update flow: %w", err)
	}

	return s.flowToResponse(f), nil
}

func (s *FlowService) DeleteFlow(ctx context.Context, id string) error {
	return s.repo.DeleteFlow(ctx, id)
}

func (s *FlowService) flowToResponse(f *flow.Flow) *flow.FlowResponse {
	return &flow.FlowResponse{
		ID:          f.ID,
		AgentID:     f.AgentID,
		Name:        f.Name,
		Description: f.Description,
		IsActive:    f.IsActive,
		Nodes:       f.Nodes,
		Edges:       f.Edges,
		Variables:   f.Variables,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}

// === Credential Operations ===

func (s *FlowService) CreateCredential(ctx context.Context, req flow.CreateCredentialRequest) (*flow.CredentialResponse, error) {
	c := &flow.Credential{
		AgentID: req.AgentID,
		Name:    req.Name,
		Type:    req.Type,
		Config:  req.Config,
	}

	if err := s.repo.CreateCredential(ctx, c); err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	return s.credentialToResponse(c), nil
}

func (s *FlowService) GetCredential(ctx context.Context, id string) (*flow.CredentialResponse, error) {
	c, err := s.repo.GetCredentialByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.credentialToResponse(c), nil
}

func (s *FlowService) GetCredentialFull(ctx context.Context, id string) (*flow.Credential, error) {
	return s.repo.GetCredentialByID(ctx, id)
}

func (s *FlowService) GetCredentialsByAgent(ctx context.Context, agentID string) ([]*flow.CredentialResponse, error) {
	credentials, err := s.repo.GetCredentialsByAgentID(ctx, agentID)
	if err != nil {
		return nil, err
	}

	var responses []*flow.CredentialResponse
	for _, c := range credentials {
		responses = append(responses, s.credentialToResponse(c))
	}
	return responses, nil
}

func (s *FlowService) UpdateCredential(ctx context.Context, id string, req flow.UpdateCredentialRequest) (*flow.CredentialResponse, error) {
	c, err := s.repo.GetCredentialByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Config != nil {
		c.Config = *req.Config
	}

	if err := s.repo.UpdateCredential(ctx, c); err != nil {
		return nil, fmt.Errorf("failed to update credential: %w", err)
	}

	return s.credentialToResponse(c), nil
}

func (s *FlowService) DeleteCredential(ctx context.Context, id string) error {
	return s.repo.DeleteCredential(ctx, id)
}

func (s *FlowService) credentialToResponse(c *flow.Credential) *flow.CredentialResponse {
	return &flow.CredentialResponse{
		ID:        c.ID,
		AgentID:   c.AgentID,
		Name:      c.Name,
		Type:      c.Type,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// TestDatabaseConnection tests a PostgreSQL/Supabase connection
func (s *FlowService) TestDatabaseConnection(ctx context.Context, config flow.DatabaseCredential) error {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// GetCredentialConfig returns parsed credential config
func (s *FlowService) GetDatabaseCredential(ctx context.Context, credentialID string) (*flow.DatabaseCredential, error) {
	c, err := s.repo.GetCredentialByID(ctx, credentialID)
	if err != nil {
		return nil, err
	}

	if c.Type != flow.CredentialTypeDatabase {
		return nil, fmt.Errorf("credential is not a database credential")
	}

	var config flow.DatabaseCredential
	if err := json.Unmarshal([]byte(c.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	return &config, nil
}

