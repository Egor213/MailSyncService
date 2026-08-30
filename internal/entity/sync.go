package entity

import "time"

type SyncJob struct {
	ID            string     `db:"id"`
	MailboxID     string     `db:"mailbox_id"`
	StatusID      int        `db:"status_id"`
	StartedAt     *time.Time `db:"started_at"`
	FinishedAt    *time.Time `db:"finished_at"`
	ErrorMsg      string     `db:"error_msg"`
	MessagesCount int        `db:"messages_count"`
	CreatedAt     time.Time  `db:"created_at"`

	StatusName string `db:"-"`
}
