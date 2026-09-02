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

type MailboxResponse struct {
	ID string `json:"id"`
}
