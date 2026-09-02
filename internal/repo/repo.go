package repo

import (
	"context"
	"time"

	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/repo/clickhouse"
	"mail-sync-service/internal/repo/elasticsearch"
	"mail-sync-service/internal/repo/pgdb"
	clicksvc "mail-sync-service/pkg/clickhouse"
	elaspkg "mail-sync-service/pkg/elastic"
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

type Search interface {
	Search(ctx context.Context, input elasticsearch.SearchInput) (*elasticsearch.SearchResult, error)
	IndexMessage(ctx context.Context, msg *entity.Message) error
	IndexMessageBody(ctx context.Context, messageID, body, bodyHTML string) error
	GetBody(ctx context.Context, messageID string) (string, string, error)
}

type Metrics interface {
	SaveSyncMetric(ctx context.Context, metric *entity.SyncMetric) error
	GetOverviewStats(ctx context.Context, since time.Time) (*clickhouse.AggregatedStats, error)
}

type Repositories struct {
	Msg  Message
	Sj   SyncJob
	Mb   Mailbox
	Refs Reference
	Search Search
	Metrics Metrics
}

func NewRepositories(pg *postgres.Postgres, es *elaspkg.Client, esIndex string, ch *clicksvc.Client) *Repositories {
	return &Repositories{
		Msg:     pgdb.NewMessageRepo(pg),
		Sj:      pgdb.NewSyncJobRepo(pg),
		Mb:      pgdb.NewMailboxRepo(pg),
		Refs:    pgdb.NewReferenceRepo(pg),
		Search:  elasticsearch.NewSearchRepo(es, esIndex),
		Metrics: clickhouse.NewMetricsRepo(ch),
	}
}
