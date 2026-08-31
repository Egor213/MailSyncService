package pgdb

import (
	"context"
	"mail-sync-service/pkg/postgres"
)

type ReferenceRepo struct {
	*postgres.Postgres
}

func NewReferenceRepo(pg *postgres.Postgres) *ReferenceRepo {
	return &ReferenceRepo{pg}
}

func (r *ReferenceRepo) GetProviderID(ctx context.Context, name string) (int, error) {
	var id int
	sql, args, _ := r.Builder.Select("id").From("mail_sync.providers").Where("name = ?", name).ToSql()
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(&id)
	return id, err
}

func (r *ReferenceRepo) GetProtocolID(ctx context.Context, name string) (int, error) {
	var id int
	sql, args, _ := r.Builder.Select("id").From("mail_sync.protocols").Where("name = ?", name).ToSql()
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(&id)
	return id, err
}

func (r *ReferenceRepo) GetAuthTypeID(ctx context.Context, name string) (int, error) {
	var id int
	sql, args, _ := r.Builder.Select("id").From("mail_sync.auth_types").Where("name = ?", name).ToSql()
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(&id)
	return id, err
}

func (r *ReferenceRepo) GetSyncStatusID(ctx context.Context, name string) (int, error) {
	var id int
	sql, args, _ := r.Builder.Select("id").From("mail_sync.sync_statuses").Where("name = ?", name).ToSql()
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(&id)
	return id, err
}
