CREATE DATABASE IF NOT EXISTS mail_sync;

CREATE TABLE IF NOT EXISTS mail_sync.sync_metrics (
    mailbox_id     String,
    provider       String,
    protocol       String,
    status         String,
    messages_count UInt32,
    duration_ms    UInt32,
    error_msg      String,
    started_at     DateTime,
    finished_at    DateTime,
    created_at     DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(started_at)
ORDER BY (mailbox_id, started_at);
