package pgdb

import (
	"context"
	"errors"
	"mail-sync-service/internal/entity"
	repoerrs "mail-sync-service/internal/repo/errors"
	"mail-sync-service/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type MessageRepo struct {
	*postgres.Postgres
}

func NewMessageRepo(pg *postgres.Postgres) *MessageRepo {
	return &MessageRepo{pg}
}

func (r *MessageRepo) Upsert(ctx context.Context, msg *entity.Message) error {
	sql, args, _ := r.Builder.
		Insert("mail_sync.messages").
		Columns("id", "mailbox_id", "uid", "folder", "subject", "from_addr", "to_addr",
			"date", "body_preview", "has_attachments", "seen", "flags", "synced_at", "hash").
		Values(msg.ID, msg.MailboxID, msg.UID, msg.Folder, msg.Subject, msg.From,
			msg.To, msg.Date, msg.BodyPreview, msg.HasAttachments, msg.Seen, msg.Flags,
			msg.SyncedAt, msg.Hash).
		Suffix("ON CONFLICT (mailbox_id, folder, uid) DO UPDATE SET " +
			"subject = EXCLUDED.subject, from_addr = EXCLUDED.from_addr, " +
			"to_addr = EXCLUDED.to_addr, date = EXCLUDED.date, " +
			"body_preview = EXCLUDED.body_preview, has_attachments = EXCLUDED.has_attachments, " +
			"seen = EXCLUDED.seen, flags = EXCLUDED.flags, synced_at = EXCLUDED.synced_at, " +
			"hash = EXCLUDED.hash").
		ToSql()

	_, err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Exec(ctx, sql, args...)
	return err
}

func (r *MessageRepo) GetByUID(ctx context.Context, mailboxID, uid, folder string) (*entity.Message, error) {
	sql, args, _ := r.Builder.
		Select("*").
		From("mail_sync.messages").
		Where("mailbox_id = ? AND folder = ? AND uid = ?", mailboxID, folder, uid).
		ToSql()

	var msg entity.Message
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(
		&msg.ID, &msg.MailboxID, &msg.UID, &msg.Folder, &msg.Subject, &msg.From,
		&msg.To, &msg.Date, &msg.BodyPreview, &msg.HasAttachments, &msg.Seen,
		&msg.Flags, &msg.SyncedAt, &msg.Hash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repoerrs.ErrNotFound
	}
	return &msg, err
}

func (r *MessageRepo) GetLastUID(ctx context.Context, mailboxID, folder string) (string, error) {
	sql, args, _ := r.Builder.
		Select("uid").
		From("mail_sync.messages").
		Where("mailbox_id = ? AND folder = ?", mailboxID, folder).
		OrderBy("uid::bigint DESC").
		Limit(1).
		ToSql()

	var uid string
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return uid, err
}

func (r *MessageRepo) MarkSeen(ctx context.Context, id string) error {
	sql, args, _ := r.Builder.
		Update("mail_sync.messages").
		Set("seen", true).
		Where("id = ?", id).
		ToSql()
	_, err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Exec(ctx, sql, args...)
	return err
}

func (r *MessageRepo) SaveBody(ctx context.Context, body *entity.MessageBody) error {
	sql, args, _ := r.Builder.
		Insert("mail_sync.message_bodies").
		Columns("message_id", "body", "body_html").
		Values(body.MessageID, body.Body, body.BodyHTML).
		Suffix("ON CONFLICT (message_id) DO UPDATE SET " +
			"body = EXCLUDED.body, body_html = EXCLUDED.body_html, updated_at = NOW()").
		ToSql()
	_, err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Exec(ctx, sql, args...)
	return err
}

func (r *MessageRepo) GetByID(ctx context.Context, id string) (*entity.Message, error) {
	sql, args, _ := r.Builder.
		Select("id", "mailbox_id", "uid", "folder", "subject", "from_addr", "to_addr",
			"date", "body_preview", "has_attachments", "seen", "flags", "synced_at", "hash").
		From("mail_sync.messages").
		Where("id = ?", id).
		ToSql()
	var msg entity.Message
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(
		&msg.ID, &msg.MailboxID, &msg.UID, &msg.Folder, &msg.Subject, &msg.From,
		&msg.To, &msg.Date, &msg.BodyPreview, &msg.HasAttachments, &msg.Seen,
		&msg.Flags, &msg.SyncedAt, &msg.Hash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repoerrs.ErrNotFound
	}
	return &msg, err
}

func (r *MessageRepo) GetBody(ctx context.Context, messageID string) (*entity.MessageBody, error) {
	sql, args, _ := r.Builder.
		Select("message_id", "body", "body_html", "created_at", "updated_at").
		From("mail_sync.message_bodies").
		Where("message_id = ?", messageID).
		ToSql()

	var body entity.MessageBody
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(
		&body.MessageID, &body.Body, &body.BodyHTML, &body.CreatedAt, &body.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repoerrs.ErrNotFound
	}
	return &body, err
}
