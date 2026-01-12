package team

import (
	"context"
	"time"
)

// Role represents user roles in the system
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleManager  Role = "manager"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

// User represents a platform user
type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Password    string    `json:"-"` // Never expose in JSON
	Role        Role      `json:"role"`
	Avatar      string    `json:"avatar,omitempty"`
	IsActive    bool      `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Team represents a team/organization
type Team struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// TeamMember represents a user's membership in a team
type TeamMember struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	UserID    string    `json:"user_id"`
	Role      Role      `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// AgentAccess represents which agents a user can access
type AgentAccess struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	AgentID   string    `json:"agent_id"`
	CanEdit   bool      `json:"can_edit"`
	CanView   bool      `json:"can_view"`
	CanChat   bool      `json:"can_chat"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserRequest represents request to create a user
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
	Role     Role   `json:"role"`
}

// UpdateUserRequest represents request to update a user
type UpdateUserRequest struct {
	Name     string `json:"name,omitempty"`
	Role     Role   `json:"role,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents login response with token
type LoginResponse struct {
	Token     string    `json:"token"`
	User      *User     `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Permissions for each role
var RolePermissions = map[Role][]string{
	RoleAdmin: {
		"agents:create", "agents:read", "agents:update", "agents:delete",
		"users:create", "users:read", "users:update", "users:delete",
		"flows:create", "flows:read", "flows:update", "flows:delete",
		"analytics:read", "settings:update", "chat:takeover",
	},
	RoleManager: {
		"agents:create", "agents:read", "agents:update",
		"users:read",
		"flows:create", "flows:read", "flows:update",
		"analytics:read", "chat:takeover",
	},
	RoleOperator: {
		"agents:read",
		"flows:read",
		"analytics:read", "chat:takeover",
	},
	RoleViewer: {
		"agents:read",
		"analytics:read",
	},
}

// HasPermission checks if a role has a specific permission
func (r Role) HasPermission(permission string) bool {
	perms, ok := RolePermissions[r]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

// IUserRepository defines database operations for users
type IUserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	UpdateLastLogin(ctx context.Context, id string) error
}

// IUserService defines business logic for user management
type IUserService interface {
	CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	GetAllUsers(ctx context.Context) ([]*User, error)
	UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	ValidateToken(ctx context.Context, token string) (*User, error)
}

