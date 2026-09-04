package pgdb

import (
	"context"
	"errors"
	"mail-sync-service/internal/entity"
	repoerrs "mail-sync-service/internal/repo/errors"
	"mail-sync-service/pkg/crypto"
	"mail-sync-service/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type MailboxRepo struct {
	*postgres.Postgres
	encKey []byte
}

func NewMailboxRepo(pg *postgres.Postgres, encKey []byte) *MailboxRepo {
	return &MailboxRepo{Postgres: pg, encKey: encKey}
}

func (r *MailboxRepo) Create(ctx context.Context, mb *entity.Mailbox) error {
	at, err := r.maybeEncrypt(mb.AccessToken)
	if err != nil {
		return err
	}
	rt, err := r.maybeEncrypt(mb.RefreshToken)
	if err != nil {
		return err
	}
	sql, args, _ := r.Builder.
		Insert("mail_sync.mailboxes").
		Columns("id", "email", "provider_id", "protocol_id", "server", "port",
			"use_tls", "auth_type_id", "access_token", "refresh_token", "token_expiry", "is_active").
		Values(mb.ID, mb.Email, mb.ProviderID, mb.ProtocolID, mb.Server, mb.Port,
			mb.UseTLS, mb.AuthTypeID, at, rt, mb.TokenExpiry, mb.IsActive).
		ToSql()
	_, err = r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Exec(ctx, sql, args...)
	return err
}

func (r *MailboxRepo) GetByID(ctx context.Context, id string) (*entity.Mailbox, error) {
	sql, args, _ := r.Builder.
		Select("m.*", "p.name as provider_name", "pr.name as protocol_name", "a.name as auth_type_name").
		From("mail_sync.mailboxes m").
		LeftJoin("mail_sync.providers p ON m.provider_id = p.id").
		LeftJoin("mail_sync.protocols pr ON m.protocol_id = pr.id").
		LeftJoin("mail_sync.auth_types a ON m.auth_type_id = a.id").
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
		return nil, repoerrs.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	mb.AccessToken, _ = r.maybeDecrypt(mb.AccessToken)
	mb.RefreshToken, _ = r.maybeDecrypt(mb.RefreshToken)
	return &mb, nil
}

func (r *MailboxRepo) GetByEmail(ctx context.Context, email string) (*entity.Mailbox, error) {
	sql, args, _ := r.Builder.
		Select("m.*", "p.name as provider_name", "pr.name as protocol_name", "a.name as auth_type_name").
		From("mail_sync.mailboxes m").
		LeftJoin("mail_sync.providers p ON m.provider_id = p.id").
		LeftJoin("mail_sync.protocols pr ON m.protocol_id = pr.id").
		LeftJoin("mail_sync.auth_types a ON m.auth_type_id = a.id").
		Where("m.email = ?", email).
		ToSql()

	var mb entity.Mailbox
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(
		&mb.ID, &mb.Email, &mb.ProviderID, &mb.ProtocolID, &mb.Server, &mb.Port,
		&mb.UseTLS, &mb.AuthTypeID, &mb.AccessToken, &mb.RefreshToken, &mb.TokenExpiry,
		&mb.CreatedAt, &mb.UpdatedAt, &mb.LastSyncAt, &mb.IsActive,
		&mb.ProviderName, &mb.ProtocolName, &mb.AuthTypeName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repoerrs.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	mb.AccessToken, _ = r.maybeDecrypt(mb.AccessToken)
	mb.RefreshToken, _ = r.maybeDecrypt(mb.RefreshToken)
	return &mb, nil
}

func (r *MailboxRepo) Update(ctx context.Context, mb *entity.Mailbox) error {
	at, err := r.maybeEncrypt(mb.AccessToken)
	if err != nil {
		return err
	}
	rt, err := r.maybeEncrypt(mb.RefreshToken)
	if err != nil {
		return err
	}
	sql, args, _ := r.Builder.
		Update("mail_sync.mailboxes").
		Set("email", mb.Email).
		Set("provider_id", mb.ProviderID).
		Set("protocol_id", mb.ProtocolID).
		Set("server", mb.Server).
		Set("port", mb.Port).
		Set("use_tls", mb.UseTLS).
		Set("auth_type_id", mb.AuthTypeID).
		Set("access_token", at).
		Set("refresh_token", rt).
		Set("token_expiry", mb.TokenExpiry).
		Set("is_active", mb.IsActive).
		Set("updated_at", "NOW()").
		Where("id = ?", mb.ID).
		ToSql()

	_, err = r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Exec(ctx, sql, args...)
	return err
}

func (r *MailboxRepo) Delete(ctx context.Context, id string) error {
	sql, args, _ := r.Builder.
		Delete("mail_sync.mailboxes").
		Where("id = ?", id).
		ToSql()
	_, err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Exec(ctx, sql, args...)
	return err
}

func (r *MailboxRepo) ListActive(ctx context.Context) ([]*entity.Mailbox, error) {
	sql, args, _ := r.Builder.
		Select("m.*", "p.name as provider_name", "pr.name as protocol_name", "a.name as auth_type_name").
		From("mail_sync.mailboxes m").
		LeftJoin("mail_sync.providers p ON m.provider_id = p.id").
		LeftJoin("mail_sync.protocols pr ON m.protocol_id = pr.id").
		LeftJoin("mail_sync.auth_types a ON m.auth_type_id = a.id").
		Where("m.is_active = ?", true).
		ToSql()

	rows, err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mailboxes []*entity.Mailbox
	for rows.Next() {
		var mb entity.Mailbox
		err := rows.Scan(
			&mb.ID, &mb.Email, &mb.ProviderID, &mb.ProtocolID, &mb.Server, &mb.Port,
			&mb.UseTLS, &mb.AuthTypeID, &mb.AccessToken, &mb.RefreshToken, &mb.TokenExpiry,
			&mb.CreatedAt, &mb.UpdatedAt, &mb.LastSyncAt, &mb.IsActive,
			&mb.ProviderName, &mb.ProtocolName, &mb.AuthTypeName,
		)
		if err != nil {
			return nil, err
		}
		mb.AccessToken, _ = r.maybeDecrypt(mb.AccessToken)
		mb.RefreshToken, _ = r.maybeDecrypt(mb.RefreshToken)
		mailboxes = append(mailboxes, &mb)
	}
	return mailboxes, rows.Err()
}

func (r *MailboxRepo) maybeEncrypt(val string) (string, error) {
	if len(r.encKey) == 0 || val == "" {
		return val, nil
	}
	return crypto.Encrypt(val, r.encKey)
}

func (r *MailboxRepo) maybeDecrypt(val string) (string, error) {
	if len(r.encKey) == 0 || val == "" {
		return val, nil
	}
	return crypto.Decrypt(val, r.encKey)
}
