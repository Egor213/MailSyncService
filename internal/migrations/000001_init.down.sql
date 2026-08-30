DROP INDEX IF EXISTS mail_sync.idx_sync_jobs_mailbox_id;
DROP INDEX IF EXISTS mail_sync.idx_messages_mailbox_folder;
DROP INDEX IF EXISTS mail_sync.idx_mailboxes_auth_type_id;
DROP INDEX IF EXISTS mail_sync.idx_mailboxes_protocol_id;
DROP INDEX IF EXISTS mail_sync.idx_mailboxes_provider_id;
DROP INDEX IF EXISTS mail_sync.idx_mailboxes_active;

DROP TABLE IF EXISTS mail_sync.sync_jobs;
DROP TABLE IF EXISTS mail_sync.messages;
DROP TABLE IF EXISTS mail_sync.mailboxes;
DROP TABLE IF EXISTS mail_sync.sync_statuses;
DROP TABLE IF EXISTS mail_sync.auth_types;
DROP TABLE IF EXISTS mail_sync.protocols;
DROP TABLE IF EXISTS mail_sync.providers;

DROP SCHEMA IF EXISTS mail_sync;