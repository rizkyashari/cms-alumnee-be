-- +goose Up
CREATE TABLE accounts (
    id BINARY(16) NOT NULL PRIMARY KEY,
    email VARCHAR(254) NULL,
    name VARCHAR(160) NOT NULL,
    account_type INT NOT NULL DEFAULT 0,
    status ENUM('active', 'disabled') NOT NULL DEFAULT 'active',
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    UNIQUE KEY uq_accounts_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE account_contacts (
    id BINARY(16) NOT NULL PRIMARY KEY,
    account_id BINARY(16) NOT NULL,
    contact_type ENUM('email', 'phone') NOT NULL,
    value VARCHAR(254) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMP(6) NULL,
    show_in_directory BOOLEAN NOT NULL DEFAULT FALSE,
    primary_contact_type VARCHAR(10) GENERATED ALWAYS AS (IF(is_primary, contact_type, NULL)) STORED,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    UNIQUE KEY uq_account_contacts_value (value),
    UNIQUE KEY uq_account_contacts_primary (account_id, primary_contact_type),
    KEY idx_account_contacts_account (account_id),
    CONSTRAINT fk_account_contacts_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    CONSTRAINT chk_account_contacts_phone_visibility CHECK (
        show_in_directory = FALSE OR (contact_type = 'phone' AND verified_at IS NOT NULL)
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE external_identities (
    id BINARY(16) NOT NULL PRIMARY KEY,
    account_id BINARY(16) NOT NULL,
    provider ENUM('google', 'facebook') NOT NULL,
    provider_subject VARCHAR(255) NOT NULL,
    linked_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    UNIQUE KEY uq_external_identities_provider_subject (provider, provider_subject),
    KEY idx_external_identities_account (account_id),
    CONSTRAINT fk_external_identities_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE alumni_profiles (
    id BINARY(16) NOT NULL PRIMARY KEY,
    account_id BINARY(16) NOT NULL,
    full_name VARCHAR(160) NOT NULL,
    batch_year SMALLINT UNSIGNED NOT NULL,
    city VARCHAR(120) NOT NULL,
    occupation VARCHAR(160) NULL,
    bio TEXT NULL,
    verification_status ENUM('non_verified', 'verified') NOT NULL DEFAULT 'non_verified',
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    UNIQUE KEY uq_alumni_profiles_account (account_id),
    CONSTRAINT fk_alumni_profiles_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE roles (
    code VARCHAR(24) NOT NULL PRIMARY KEY,
    description VARCHAR(160) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO roles (code, description) VALUES
    ('member', 'Alumni member; access is refined by profile verification status'),
    ('officer', 'Pengurus with explicitly authorized management capabilities'),
    ('admin', 'Administrator with full management capabilities');

CREATE TABLE account_roles (
    account_id BINARY(16) NOT NULL,
    role_code VARCHAR(24) NOT NULL,
    granted_by BINARY(16) NULL,
    granted_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (account_id, role_code),
    KEY idx_account_roles_granted_by (granted_by),
    CONSTRAINT fk_account_roles_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    CONSTRAINT fk_account_roles_role FOREIGN KEY (role_code) REFERENCES roles(code),
    CONSTRAINT fk_account_roles_granted_by FOREIGN KEY (granted_by) REFERENCES accounts(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE sessions (
    id BINARY(16) NOT NULL PRIMARY KEY,
    account_id BINARY(16) NOT NULL,
    token_hash BINARY(32) NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    last_seen_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    expires_at TIMESTAMP(6) NOT NULL,
    revoked_at TIMESTAMP(6) NULL,
    UNIQUE KEY uq_sessions_token_hash (token_hash),
    KEY idx_sessions_account_expiry (account_id, expires_at),
    CONSTRAINT fk_sessions_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- +goose Down
DROP TABLE sessions;
DROP TABLE account_roles;
DROP TABLE roles;
DROP TABLE alumni_profiles;
DROP TABLE external_identities;
DROP TABLE account_contacts;
DROP TABLE accounts;
