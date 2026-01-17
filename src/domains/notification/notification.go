package notification

import (
	"context"
	"time"
)

// Notification represents an alert/notification
type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id,omitempty"`
	Type      string    `json:"type"` // info, warning, error, success
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	AgentID   string    `json:"agent_id,omitempty"`
	AgentName string    `json:"agent_name,omitempty"`
	Data      string    `json:"data,omitempty"` // JSON metadata
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// NotificationSettings represents notification preferences
type NotificationSettings struct {
	ID                 string `json:"id"`
	UserID             string `json:"user_id"`
	EmailEnabled       bool   `json:"email_enabled"`
	Email              string `json:"email,omitempty"`
	NewMessageAlert    bool   `json:"new_message_alert"`
	ErrorAlert         bool   `json:"error_alert"`
	DailyDigest        bool   `json:"daily_digest"`
	WeeklyReport       bool   `json:"weekly_report"`
}

// INotificationRepository defines database operations for notifications
type INotificationRepository interface {
	Create(ctx context.Context, notif *Notification) error
	GetByUserID(ctx context.Context, userID string, limit int, unreadOnly bool) ([]*Notification, error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}

// INotificationService defines business logic for notifications
type INotificationService interface {
	Create(ctx context.Context, notif *Notification) error
	GetNotifications(ctx context.Context, userID string, limit int, unreadOnly bool) ([]*Notification, error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}




