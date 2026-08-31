package service

import (
	"context"
	"net/http"

	"mail-sync-service/internal/config"
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/infrastruct/kafka"
	"mail-sync-service/internal/repo"
	"mail-sync-service/internal/repo/redis"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
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
	Repos  *repo.Repositories
	Locker redis.RedisLocker

	KafkaProducer kafka.KafkaProducer

	Config     *config.Config
	HTTPClient *http.Client
	TrManager  TRManager
}

func NewServices(deps ServicesDependencies) *Services {
	syncService := NewSyncService(
		deps.Repos.Mb,
		deps.Repos.Msg,
		deps.Repos.Sj,
		deps.Repos.Refs,
		deps.Locker,
	)

	mailboxService := NewMailboxService(
		deps.Repos.Mb,
		deps.Repos.Refs,
		deps.KafkaProducer,
	)

	oauthService := NewOAuthService(
		&deps.Config.OAuth,
		deps.HTTPClient,
		mailboxService,
	)

	return &Services{
		Mailbox: mailboxService,
		Sync:    syncService,
		OAuth:   oauthService,
	}
}
