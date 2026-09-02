package entity

import "time"

type SyncMetric struct {
	MailboxID     string    `json:"mailbox_id"`
	Provider      string    `json:"provider"`
	Protocol      string    `json:"protocol"`
	Status        string    `json:"status"`
	MessagesCount int       `json:"messages_count"`
	DurationMs    int       `json:"duration_ms"`
	ErrorMsg      string    `json:"error_msg"`
	StartedAt     time.Time `json:"started_at"`
	FinishedAt    time.Time `json:"finished_at"`
}
