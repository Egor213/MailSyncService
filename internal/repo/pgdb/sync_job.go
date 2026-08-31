package pgdb

import (
	"context"
	"errors"
	"mail-sync-service/internal/entity"
	repoerrs "mail-sync-service/internal/repo/errors"
	"mail-sync-service/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type SyncJobRepo struct {
	*postgres.Postgres
}

func NewSyncJobRepo(pg *postgres.Postgres) *SyncJobRepo {
	return &SyncJobRepo{pg}
}

func (r *SyncJobRepo) Create(ctx context.Context, job *entity.SyncJob) error {
	sql, args, _ := r.Builder.
		Insert("mail_sync.sync_jobs").
		Columns("id", "mailbox_id", "status_id", "started_at", "finished_at", "error_msg", "messages_count").
		Values(job.ID, job.MailboxID, job.StatusID, job.StartedAt, job.FinishedAt, job.ErrorMsg, job.MessagesCount).
		ToSql()
	_, err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Exec(ctx, sql, args...)
	return err
}

func (r *SyncJobRepo) Update(ctx context.Context, job *entity.SyncJob) error {
	sql, args, _ := r.Builder.
		Update("mail_sync.sync_jobs").
		Set("status_id", job.StatusID).
		Set("started_at", job.StartedAt).
		Set("finished_at", job.FinishedAt).
		Set("error_msg", job.ErrorMsg).
		Set("messages_count", job.MessagesCount).
		Where("id = ?", job.ID).
		ToSql()
	_, err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Exec(ctx, sql, args...)
	return err
}

func (r *SyncJobRepo) GetLastByMailboxID(ctx context.Context, mailboxID string) (*entity.SyncJob, error) {
	sql, args, _ := r.Builder.
		Select("j.*", "s.name as status_name").
		From("mail_sync.sync_jobs j").
		LeftJoin("mail_sync.sync_statuses s ON j.status_id = s.id").
		Where("j.mailbox_id = ?", mailboxID).
		OrderBy("j.created_at DESC").
		Limit(1).
		ToSql()

	var job entity.SyncJob
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(
		&job.ID, &job.MailboxID, &job.StatusID, &job.StartedAt, &job.FinishedAt,
		&job.ErrorMsg, &job.MessagesCount, &job.CreatedAt,
		&job.StatusName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repoerrs.ErrNotFound
	}
	return &job, err
}
