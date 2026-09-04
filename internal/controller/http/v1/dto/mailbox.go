package httpdto

import "time"

type CreateMailboxRequest struct {
	Email        string `json:"email" validate:"required,email"`
	Provider     string `json:"provider" validate:"required"`
	Protocol     string `json:"protocol" validate:"required"`
	Server       string `json:"server" validate:"required"`
	Port         int    `json:"port" validate:"required"`
	UseTLS       bool   `json:"use_tls"`
	AuthType     string `json:"auth_type" validate:"required"`
	AccessToken  string `json:"access_token" validate:"required"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenExpiry  *time.Time `json:"token_expiry,omitempty"`
}

type UpdateMailboxRequest struct {
	Email        string     `json:"email,omitempty"`
	Server       string     `json:"server,omitempty"`
	Port         int        `json:"port,omitempty"`
	UseTLS       *bool      `json:"use_tls,omitempty"`
	AccessToken  string     `json:"access_token,omitempty"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	TokenExpiry  *time.Time `json:"token_expiry,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
}

type MailboxResponse struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	ProviderID   int        `json:"provider_id"`
	ProtocolID   int        `json:"protocol_id"`
	Server       string     `json:"server"`
	Port         int        `json:"port"`
	UseTLS       bool       `json:"use_tls"`
	AuthTypeID   int        `json:"auth_type_id"`
	TokenExpiry  *time.Time `json:"token_expiry,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	IsActive     bool       `json:"is_active"`
}
