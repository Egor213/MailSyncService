package repo

import (
	"context"
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/repo/pgdb"
	"mail-sync-service/pkg/postgres"
)

type Mailbox interface {
	Create(ctx context.Context, mb *entity.Mailbox) error
	GetByID(ctx context.Context, id string) (*entity.Mailbox, error)
	GetByEmail(ctx context.Context, email string) (*entity.Mailbox, error)
	Update(ctx context.Context, mb *entity.Mailbox) error
	Delete(ctx context.Context, id string) error
	ListActive(ctx context.Context) ([]*entity.Mailbox, error)
}

type Message interface {
	Upsert(ctx context.Context, msg *entity.Message) error
	GetByUID(ctx context.Context, mailboxID, uid, folder string) (*entity.Message, error)
	GetLastUID(ctx context.Context, mailboxID, folder string) (string, error)
	MarkSeen(ctx context.Context, id string) error

	SaveBody(ctx context.Context, body *entity.MessageBody) error
	GetBody(ctx context.Context, messageID string) (*entity.MessageBody, error)
}

type SyncJob interface {
	Create(ctx context.Context, job *entity.SyncJob) error
	Update(ctx context.Context, job *entity.SyncJob) error
	GetLastByMailboxID(ctx context.Context, mailboxID string) (*entity.SyncJob, error)
}

type Reference interface {
	GetProviderID(ctx context.Context, name string) (int, error)
	GetProtocolID(ctx context.Context, name string) (int, error)
	GetAuthTypeID(ctx context.Context, name string) (int, error)
	GetSyncStatusID(ctx context.Context, name string) (int, error)
}

type Repositories struct {
	Msg  Message
	Sj   SyncJob
	Mb   Mailbox
	Refs Reference
}

func NewRepositories(pg *postgres.Postgres) *Repositories {
	return &Repositories{
		Msg:  pgdb.NewMessageRepo(pg),
		Sj:   pgdb.NewSyncJobRepo(pg),
		Mb:   pgdb.NewMailboxRepo(pg),
		Refs: pgdb.NewReferenceRepo(pg),
	}
}
