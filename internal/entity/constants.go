package entity

type Provider string

const (
	ProviderGmail   Provider = "gmail"
	ProviderOutlook Provider = "outlook"
	ProviderYandex  Provider = "yandex"
	ProviderMailru  Provider = "mailru"
	ProviderCustom  Provider = "custom"
)

func (p Provider) String() string { return string(p) }

type Protocol string

const (
	ProtocolIMAP Protocol = "imap"
	ProtocolPOP3 Protocol = "pop3"
)

func (p Protocol) String() string { return string(p) }

type AuthType string

const (
	AuthTypePlain  AuthType = "plain"
	AuthTypeOAuth2 AuthType = "oauth2"
)

func (a AuthType) String() string { return string(a) }

type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending"
	SyncStatusRunning SyncStatus = "running"
	SyncStatusSuccess SyncStatus = "success"
	SyncStatusFailed  SyncStatus = "failed"
)

func (s SyncStatus) String() string { return string(s) }
