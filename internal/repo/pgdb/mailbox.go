package pgdb

import (
	"context"
	"errors"
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/repo"
	"mail-sync-service/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type MailboxRepo struct {
	*postgres.Postgres
}

func NewMailboxRepo(pg *postgres.Postgres) *MailboxRepo {
	return &MailboxRepo{pg}
}

func (r *MailboxRepo) Create(ctx context.Context, mb *entity.Mailbox) error {
	sql, args, _ := r.Builder.
		Insert("mailboxes").
		Columns("id", "email", "provider_id", "protocol_id", "server", "port",
			"use_tls", "auth_type_id", "access_token", "refresh_token", "token_expiry", "is_active").
		Values(mb.ID, mb.Email, mb.ProviderID, mb.ProtocolID, mb.Server, mb.Port,
			mb.UseTLS, mb.AuthTypeID, mb.AccessToken, mb.RefreshToken, mb.TokenExpiry, mb.IsActive).
		ToSql()
	_, err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Exec(ctx, sql, args...)
	return err
}

func (r *MailboxRepo) GetByID(ctx context.Context, id string) (*entity.Mailbox, error) {
	sql, args, _ := r.Builder.
		Select("m.*", "p.name as provider_name", "pr.name as protocol_name", "a.name as auth_type_name").
		From("mailboxes m").
		LeftJoin("providers p ON m.provider_id = p.id").
		LeftJoin("protocols pr ON m.protocol_id = pr.id").
		LeftJoin("auth_types a ON m.auth_type_id = a.id").
		Where("m.id = ?", id).
		ToSql()
	var mb entity.Mailbox
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(
		&mb.ID, &mb.Email, &mb.ProviderID, &mb.ProtocolID, &mb.Server, &mb.Port,
		&mb.UseTLS, &mb.AuthTypeID, &mb.AccessToken, &mb.RefreshToken, &mb.TokenExpiry,
		&mb.CreatedAt, &mb.UpdatedAt, &mb.LastSyncAt, &mb.IsActive,
		&mb.ProviderName, &mb.ProtocolName, &mb.AuthTypeName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	return &mb, err
}

// Остальные методы (ListActive, Update, Delete) аналогично обновить.
