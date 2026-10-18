CREATE TABLE users (
    id BIGINT NOT NULL,
    email VARCHAR(254) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    username VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    email_verified_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    deleted_at DATETIME(6) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_users_email (email),
    UNIQUE KEY uq_users_username (username),
    KEY idx_users_deleted_at (deleted_at),
    CONSTRAINT chk_users_id_positive CHECK (id > 0),
    CONSTRAINT chk_users_username_length CHECK (CHAR_LENGTH(username) BETWEEN 2 AND 32),
    CONSTRAINT chk_users_display_name CHECK (CHAR_LENGTH(display_name) BETWEEN 1 AND 100)
) ENGINE=InnoDB;

CREATE TABLE user_credentials (
    id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    password_hash VARCHAR(255) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    password_changed_at DATETIME(6) NOT NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_user_credentials_user (user_id),
    CONSTRAINT fk_user_credentials_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT chk_user_credentials_id_positive CHECK (id > 0)
) ENGINE=InnoDB;

CREATE TABLE auth_sessions (
    id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    refresh_token_hash BINARY(32) NOT NULL,
    device_name VARCHAR(100) NULL,
    created_at DATETIME(6) NOT NULL,
    last_used_at DATETIME(6) NOT NULL,
    expires_at DATETIME(6) NOT NULL,
    revoked_at DATETIME(6) NULL,
    replaced_by_session_id BIGINT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_auth_sessions_refresh_hash (refresh_token_hash),
    KEY idx_auth_sessions_user_active (user_id, revoked_at, expires_at),
    KEY idx_auth_sessions_replaced_by (replaced_by_session_id),
    CONSTRAINT fk_auth_sessions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_sessions_replaced_by FOREIGN KEY (replaced_by_session_id) REFERENCES auth_sessions (id) ON DELETE SET NULL,
    CONSTRAINT chk_auth_sessions_id_positive CHECK (id > 0),
    CONSTRAINT chk_auth_sessions_expiry CHECK (expires_at > created_at)
) ENGINE=InnoDB;

CREATE TABLE account_tokens (
    id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    purpose ENUM('verify_email', 'reset_password') NOT NULL,
    bearer_token_hash BINARY(32) NOT NULL,
    expires_at DATETIME(6) NOT NULL,
    used_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_account_tokens_hash (bearer_token_hash),
    KEY idx_account_tokens_user_purpose (user_id, purpose, used_at, expires_at),
    CONSTRAINT fk_account_tokens_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT chk_account_tokens_id_positive CHECK (id > 0),
    CONSTRAINT chk_account_tokens_expiry CHECK (expires_at > created_at)
) ENGINE=InnoDB;

CREATE TABLE email_outbox (
    id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    template ENUM('verify_email', 'reset_password') NOT NULL,
    recipient VARCHAR(254) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    template_data JSON NOT NULL,
    encrypted_token_payload BLOB NULL,
    delivery_state ENUM('pending', 'processing', 'sent', 'dead') NOT NULL DEFAULT 'pending',
    attempt_count SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    next_attempt_at DATETIME(6) NOT NULL,
    sent_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    KEY idx_email_outbox_delivery (delivery_state, next_attempt_at, id),
    KEY idx_email_outbox_user (user_id, created_at),
    CONSTRAINT fk_email_outbox_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT chk_email_outbox_id_positive CHECK (id > 0)
) ENGINE=InnoDB;
