-- +goose Up
CREATE TABLE verification_requests (
    profile_id BINARY(16) NOT NULL PRIMARY KEY,
    status ENUM('pending', 'approved', 'rejected') NOT NULL DEFAULT 'pending',
    decision_kind ENUM('peer', 'bootstrap') NULL,
    requested_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    decided_by BINARY(16) NULL,
    decided_at TIMESTAMP(6) NULL,
    decision_note VARCHAR(500) NULL,
    KEY idx_verification_requests_status (status, requested_at),
    KEY idx_verification_requests_decided_by (decided_by),
    CONSTRAINT fk_verification_requests_profile FOREIGN KEY (profile_id) REFERENCES alumni_profiles(id) ON DELETE CASCADE,
    CONSTRAINT fk_verification_requests_decided_by FOREIGN KEY (decided_by) REFERENCES accounts(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE alumni_attestations (
    id BINARY(16) NOT NULL PRIMARY KEY,
    profile_id BINARY(16) NOT NULL,
    attester_id BINARY(16) NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    UNIQUE KEY uq_alumni_attestations_profile_attester (profile_id, attester_id),
    KEY idx_alumni_attestations_attester (attester_id),
    CONSTRAINT fk_alumni_attestations_profile FOREIGN KEY (profile_id) REFERENCES alumni_profiles(id) ON DELETE CASCADE,
    CONSTRAINT fk_alumni_attestations_attester FOREIGN KEY (attester_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- +goose Down
DROP TABLE alumni_attestations;
DROP TABLE verification_requests;
