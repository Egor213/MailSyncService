CREATE SCHEMA IF NOT EXISTS mail_sync;

CREATE TABLE mail_sync.providers (
    id          SERIAL,
    name        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT providers_pkey PRIMARY KEY (id)
);

CREATE TABLE mail_sync.protocols (
    id          SERIAL,
    name        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT protocols_pkey PRIMARY KEY (id)
);

CREATE TABLE mail_sync.auth_types (
    id          SERIAL,
    name        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT auth_types_pkey PRIMARY KEY (id)
);

CREATE TABLE mail_sync.sync_statuses (
    id          SERIAL,
    name        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT sync_statuses_pkey PRIMARY KEY (id)
);

CREATE TABLE mail_sync.mailboxes (
    id              UUID,
    email           TEXT NOT NULL UNIQUE,
    provider_id     INT NOT NULL,
    protocol_id     INT NOT NULL,
    server          TEXT NOT NULL,
    port            INT NOT NULL,
    use_tls         BOOLEAN DEFAULT TRUE,
    auth_type_id    INT NOT NULL,
    access_token    TEXT NOT NULL,
    refresh_token   TEXT,
    token_expiry    TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    last_sync_at    TIMESTAMP,
    is_active       BOOLEAN DEFAULT TRUE,
    CONSTRAINT mailboxes_pkey PRIMARY KEY (id),
    CONSTRAINT mailboxes_provider_id_fkey FOREIGN KEY (provider_id)
        REFERENCES mail_sync.providers(id) ON DELETE RESTRICT,
    CONSTRAINT mailboxes_protocol_id_fkey FOREIGN KEY (protocol_id)
        REFERENCES mail_sync.protocols(id) ON DELETE RESTRICT,
    CONSTRAINT mailboxes_auth_type_id_fkey FOREIGN KEY (auth_type_id)
        REFERENCES mail_sync.auth_types(id) ON DELETE RESTRICT
);

CREATE TABLE mail_sync.messages (
    id              UUID,
    mailbox_id      UUID NOT NULL,
    uid             TEXT NOT NULL,
    folder          TEXT NOT NULL DEFAULT 'INBOX',
    subject         TEXT,
    from_addr       TEXT,
    to_addr         TEXT,
    date            TIMESTAMP,
    body_preview    TEXT,
    has_attachments BOOLEAN DEFAULT FALSE,
    seen            BOOLEAN DEFAULT FALSE,
    flags           JSONB DEFAULT '[]'::jsonb,
    synced_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    hash            TEXT NOT NULL,
    CONSTRAINT messages_pkey PRIMARY KEY (id),
    CONSTRAINT messages_mailbox_id_fkey FOREIGN KEY (mailbox_id)
        REFERENCES mail_sync.mailboxes(id) ON DELETE CASCADE,
    CONSTRAINT messages_mailbox_folder_uid_unique UNIQUE (mailbox_id, folder, uid)
);

CREATE TABLE mail_sync.message_bodies (
    message_id UUID PRIMARY KEY REFERENCES mail_sync.messages(id) ON DELETE CASCADE,
    body TEXT,
    body_html TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE mail_sync.sync_jobs (
    id              UUID,
    mailbox_id      UUID NOT NULL,
    status_id       INT NOT NULL,
    started_at      TIMESTAMP,
    finished_at     TIMESTAMP,
    error_msg       TEXT,
    messages_count  INT DEFAULT 0,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT sync_jobs_pkey PRIMARY KEY (id),
    CONSTRAINT sync_jobs_mailbox_id_fkey FOREIGN KEY (mailbox_id)
        REFERENCES mail_sync.mailboxes(id) ON DELETE CASCADE,
    CONSTRAINT sync_jobs_status_id_fkey FOREIGN KEY (status_id)
        REFERENCES mail_sync.sync_statuses(id) ON DELETE RESTRICT
);

CREATE INDEX idx_mailboxes_active ON mail_sync.mailboxes(is_active);
CREATE INDEX idx_mailboxes_provider_id ON mail_sync.mailboxes(provider_id);
CREATE INDEX idx_mailboxes_protocol_id ON mail_sync.mailboxes(protocol_id);
CREATE INDEX idx_mailboxes_auth_type_id ON mail_sync.mailboxes(auth_type_id);
CREATE INDEX idx_messages_mailbox_folder ON mail_sync.messages(mailbox_id, folder);
CREATE INDEX idx_sync_jobs_mailbox_id ON mail_sync.sync_jobs(mailbox_id);
CREATE INDEX idx_message_bodies_created_at ON mail_sync.message_bodies(created_at);

INSERT INTO mail_sync.providers (id, name) VALUES
    (1, 'gmail'),
    (2, 'outlook'),
    (3, 'yandex'),
    (4, 'mailru'),
    (5, 'custom')
ON CONFLICT (id) DO NOTHING;

INSERT INTO mail_sync.protocols (id, name) VALUES
    (1, 'imap'),
    (2, 'pop3')
ON CONFLICT (id) DO NOTHING;

INSERT INTO mail_sync.auth_types (id, name) VALUES
    (1, 'plain'),
    (2, 'oauth2')
ON CONFLICT (id) DO NOTHING;

INSERT INTO mail_sync.sync_statuses (id, name) VALUES
    (1, 'pending'),
    (2, 'running'),
    (3, 'success'),
    (4, 'failed')
ON CONFLICT (id) DO NOTHING;