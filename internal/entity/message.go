package entity

import "time"

type Message struct {
	ID             string    `db:"id"`
	MailboxID      string    `db:"mailbox_id"`
	UID            string    `db:"uid"`    // UID from IMAP or UIDL from POP3
	Folder         string    `db:"folder"` // INBOX, etc.
	Subject        string    `db:"subject"`
	From           string    `db:"from_addr"`
	To             string    `db:"to_addr"`
	Date           time.Time `db:"date"`
	BodyPreview    string    `db:"body_preview"`
	HasAttachments bool      `db:"has_attachments"`
	Seen           bool      `db:"seen"`
	Flags          []string  `db:"flags"`
	SyncedAt       time.Time `db:"synced_at"`
	Hash           string    `db:"hash"` // для дедупликации
}

type MessageBody struct {
	MessageID string    `db:"message_id"`
	Body      string    `db:"body"`
	BodyHTML  string    `db:"body_html"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
