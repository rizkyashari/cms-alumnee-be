-- +goose Up
CREATE TABLE email_otp_challenges (
    id BINARY(16) NOT NULL PRIMARY KEY,
    email VARCHAR(254) NOT NULL,
    code_hash BINARY(32) NOT NULL,
    purpose ENUM('registration', 'login', 'admin_bootstrap') NOT NULL,
    expires_at TIMESTAMP(6) NOT NULL,
    attempt_count TINYINT UNSIGNED NOT NULL DEFAULT 0,
    consumed_at TIMESTAMP(6) NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    KEY idx_email_otp_challenges_email (email, created_at),
    KEY idx_email_otp_challenges_expiry (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE registration_tokens (
    token_hash BINARY(32) NOT NULL PRIMARY KEY,
    email VARCHAR(254) NOT NULL,
    expires_at TIMESTAMP(6) NOT NULL,
    consumed_at TIMESTAMP(6) NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    KEY idx_registration_tokens_expiry (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE alumni_profile_photos (
    account_id BINARY(16) NOT NULL PRIMARY KEY,
    mime_type VARCHAR(32) NOT NULL,
    image_data MEDIUMBLOB NOT NULL,
    updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    CONSTRAINT fk_alumni_profile_photos_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

ALTER TABLE alumni_profiles
    ADD COLUMN country_code CHAR(2) NOT NULL DEFAULT 'ID' AFTER city,
    ADD COLUMN city_code VARCHAR(16) NULL AFTER country_code;

-- +goose Down
ALTER TABLE alumni_profiles
    DROP COLUMN city_code,
    DROP COLUMN country_code;
DROP TABLE alumni_profile_photos;
DROP TABLE registration_tokens;
DROP TABLE email_otp_challenges;
