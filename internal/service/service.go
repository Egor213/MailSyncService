package service

import (
	"context"
	"net/http"

	"mail-sync-service/internal/config"
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/infrastruct/kafka"
	"mail-sync-service/internal/repo"
	"mail-sync-service/internal/repo/redis"

	"github.com/avito-tech/go-transaction-manager/trm"
)

type TRManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
	DoWithSettings(ctx context.Context, s trm.Settings, fn func(ctx context.Context) error) (err error)
}

type Mailbox interface {
	CreateMailbox(ctx context.Context, in CreateMailboxInput) (*entity.Mailbox, error)
	GetMailbox(ctx context.Context, id string) (*entity.Mailbox, error)
	UpdateMailbox(ctx context.Context, mb *entity.Mailbox) error
	DeleteMailbox(ctx context.Context, id string) error
	ListActive(ctx context.Context) ([]*entity.Mailbox, error)
	TriggerSync(ctx context.Context, mailboxID string) error
}

type Sync interface {
	SyncMailbox(ctx context.Context, mailboxID string) error
	SyncAllActive(ctx context.Context) error
	GetLastSyncStatus(ctx context.Context, mailboxID string) (*entity.SyncJob, error)
}

type OAuth interface {
	GetAuthURL(ctx context.Context, provider string) (string, error)
	HandleCallback(ctx context.Context, provider, code, state string) (string, error)
}

type Services struct {
	Mailbox Mailbox
	Sync    Sync
	OAuth   OAuth
}

type ServicesDependencies struct {
	MailboxRepo repo.Mailbox
	MessageRepo repo.Message
	SyncJobRepo repo.SyncJob
	RefRepo     repo.Reference
	Locker      redis.RedisLocker

	KafkaProducer kafka.KafkaProducer

	OAuthConfig *config.OAuth
	HTTPClient  *http.Client
	TrManager   TRManager
}

func NewServices(deps ServicesDependencies) *Services {
	syncService := NewSyncService(
		deps.MailboxRepo,
		deps.MessageRepo,
		deps.SyncJobRepo,
		deps.RefRepo,
		deps.Locker,
	)

	oauthService := NewOAuthService(
		deps.OAuthConfig,
		deps.HTTPClient,
		deps.MailboxRepo,
	)

	mailboxService := NewMailboxService(
		deps.MailboxRepo,
		deps.RefRepo,
		deps.KafkaProducer,
	)

	return &Services{
		Mailbox: mailboxService,
		Sync:    syncService,
		OAuth:   oauthService,
	}
}
