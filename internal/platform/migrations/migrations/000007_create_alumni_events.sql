-- +goose Up
CREATE TABLE alumni_events (
    id BINARY(16) NOT NULL PRIMARY KEY,
    author_id BINARY(16) NULL,
    title VARCHAR(200) NOT NULL,
    category VARCHAR(40) NOT NULL,
    venue VARCHAR(200) NOT NULL DEFAULT '',
    description TEXT NOT NULL,
    starts_at TIMESTAMP(6) NOT NULL,
    ends_at TIMESTAMP(6) NULL,
    status ENUM('draft', 'pending_review', 'published', 'rejected', 'archived') NOT NULL DEFAULT 'draft',
    review_note VARCHAR(500) NULL,
    reviewed_by BINARY(16) NULL,
    reviewed_at TIMESTAMP(6) NULL,
    published_at TIMESTAMP(6) NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    KEY idx_alumni_events_status_starts (status, starts_at),
    KEY idx_alumni_events_author (author_id, created_at),
    CONSTRAINT fk_alumni_events_author FOREIGN KEY (author_id) REFERENCES accounts(id) ON DELETE SET NULL,
    CONSTRAINT fk_alumni_events_reviewed_by FOREIGN KEY (reviewed_by) REFERENCES accounts(id) ON DELETE SET NULL,
    CONSTRAINT chk_alumni_events_window CHECK (ends_at IS NULL OR ends_at > starts_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- +goose Down
DROP TABLE alumni_events;
