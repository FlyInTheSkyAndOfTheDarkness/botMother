package health

// IntegrationStatus represents the status of an integration
type IntegrationStatus struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // whatsapp, telegram, instagram
	IsConnected bool   `json:"is_connected"`
	IsActive    bool   `json:"is_active"`
	Status      string `json:"status"`      // "connected", "disconnected", "error"
	Message     string `json:"message,omitempty"`
	DeviceID    string `json:"device_id,omitempty"` // For WhatsApp
	BotToken    string `json:"bot_token,omitempty"` // For Telegram (masked)
}

// AgentHealth represents health status for an agent
type AgentHealth struct {
	AgentID      string              `json:"agent_id"`
	AgentName    string              `json:"agent_name"`
	IsActive     bool                `json:"is_active"`
	Integrations []IntegrationStatus  `json:"integrations"`
}

// SystemHealth represents overall system health
type SystemHealth struct {
	Status      string         `json:"status"` // "healthy", "degraded", "unhealthy"
	WhatsApp    WhatsAppHealth `json:"whatsapp"`
	Telegram    TelegramHealth `json:"telegram"`
	Agents      []AgentHealth  `json:"agents"`
	TotalAgents int            `json:"total_agents"`
	ActiveAgents int            `json:"active_agents"`
}

// WhatsAppHealth represents WhatsApp system health
type WhatsAppHealth struct {
	TotalDevices    int `json:"total_devices"`
	ConnectedDevices int `json:"connected_devices"`
	LoggedInDevices  int `json:"logged_in_devices"`
}

// TelegramHealth represents Telegram system health
type TelegramHealth struct {
	TotalBots    int `json:"total_bots"`
	RunningBots  int `json:"running_bots"`
}

