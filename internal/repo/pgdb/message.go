package pgdb

import (
	"context"
	"errors"
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/repo"

	"github.com/jackc/pgx/v5"
)

func (r *MessageRepo) SaveBody(ctx context.Context, body *entity.MessageBody) error {
	sql, args, _ := r.Builder.
		Insert("mail_sync.message_bodies").
		Columns("message_id", "body", "body_html").
		Values(body.MessageID, body.Body, body.BodyHTML).
		Suffix("ON CONFLICT (message_id) DO UPDATE SET body = EXCLUDED.body, body_html = EXCLUDED.body_html, updated_at = NOW()").
		ToSql()
	_, err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Exec(ctx, sql, args...)
	return err
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
		return nil, repo.ErrNotFound
	}
	return &body, err
}
