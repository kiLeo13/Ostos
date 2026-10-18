CREATE TABLE families (
    id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    owner_id BIGINT NOT NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    deleted_at DATETIME(6) NULL,
    PRIMARY KEY (id),
    KEY idx_families_owner (owner_id, deleted_at),
    CONSTRAINT fk_families_owner FOREIGN KEY (owner_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT chk_families_id_positive CHECK (id > 0),
    CONSTRAINT chk_families_name CHECK (CHAR_LENGTH(name) BETWEEN 1 AND 100)
) ENGINE=InnoDB;

CREATE TABLE family_members (
    id BIGINT NOT NULL,
    family_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role ENUM('manager', 'member') NOT NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_family_members_family_user (family_id, user_id),
    KEY idx_family_members_user (user_id, family_id),
    CONSTRAINT fk_family_members_family FOREIGN KEY (family_id) REFERENCES families (id) ON DELETE CASCADE,
    CONSTRAINT fk_family_members_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT chk_family_members_id_positive CHECK (id > 0)
) ENGINE=InnoDB;

CREATE TABLE family_invitations (
    id BIGINT NOT NULL,
    family_id BIGINT NOT NULL,
    invited_user_id BIGINT NOT NULL,
    invited_by_user_id BIGINT NOT NULL,
    expires_at DATETIME(6) NOT NULL,
    created_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_family_invitations_family_user (family_id, invited_user_id),
    KEY idx_family_invitations_recipient (invited_user_id, expires_at),
    KEY idx_family_invitations_inviter (invited_by_user_id),
    CONSTRAINT fk_family_invitations_family FOREIGN KEY (family_id) REFERENCES families (id) ON DELETE CASCADE,
    CONSTRAINT fk_family_invitations_recipient FOREIGN KEY (invited_user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_family_invitations_inviter FOREIGN KEY (invited_by_user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT chk_family_invitations_id_positive CHECK (id > 0),
    CONSTRAINT chk_family_invitations_expiry CHECK (expires_at > created_at)
) ENGINE=InnoDB;

CREATE TABLE audit_events (
    id BIGINT NOT NULL,
    actor_user_id BIGINT NULL,
    family_id BIGINT NULL,
    action_code VARCHAR(80) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    target_type VARCHAR(40) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    target_id BIGINT NULL,
    metadata JSON NOT NULL,
    request_id VARCHAR(100) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    created_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    KEY idx_audit_events_family_created (family_id, created_at, id),
    KEY idx_audit_events_actor_created (actor_user_id, created_at, id),
    CONSTRAINT fk_audit_events_actor FOREIGN KEY (actor_user_id) REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT fk_audit_events_family FOREIGN KEY (family_id) REFERENCES families (id) ON DELETE SET NULL,
    CONSTRAINT chk_audit_events_id_positive CHECK (id > 0)
) ENGINE=InnoDB;
