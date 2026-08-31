package mail

import (
	"context"
	"mail-sync-service/internal/entity"
)

type MessageMeta struct {
	UID            string
	Folder         string
	Subject        string
	From           string
	To             string
	Date           string
	BodyPreview    string
	HasAttachments bool
	Seen           bool
	Flags          []string
	BodyText       string
	BodyHTML       string
}

type MailClient interface {
	Connect(ctx context.Context, server string, port int, useTLS bool) error
	Authenticate(ctx context.Context, mailbox *entity.Mailbox) error
	ListMessages(ctx context.Context, folder string, lastUID string) ([]*MessageMeta, error)
	FetchFullMessage(ctx context.Context, folder, uid string) (*MessageMeta, error)
	Close() error
}
