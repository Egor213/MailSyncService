package entity

import "time"

type Mailbox struct {
	ID           string     `db:"id"`
	Email        string     `db:"email"`
	ProviderID   int        `db:"provider_id"`
	ProtocolID   int        `db:"protocol_id"`
	Server       string     `db:"server"`
	Port         int        `db:"port"`
	UseTLS       bool       `db:"use_tls"`
	AuthTypeID   int        `db:"auth_type_id"`
	AccessToken  string     `db:"access_token"`
	RefreshToken string     `db:"refresh_token"`
	TokenExpiry  *time.Time `db:"token_expiry"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	LastSyncAt   *time.Time `db:"last_sync_at"`
	IsActive     bool       `db:"is_active"`

	ProviderName string `db:"-"`
	ProtocolName string `db:"-"`
	AuthTypeName string `db:"-"`
}
