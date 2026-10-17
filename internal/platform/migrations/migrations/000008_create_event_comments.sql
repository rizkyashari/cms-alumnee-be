-- +goose Up
CREATE TABLE event_comments (
    id BINARY(16) NOT NULL PRIMARY KEY,
    event_id BINARY(16) NOT NULL,
    author_id BINARY(16) NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    edited_at TIMESTAMP(6) NULL,
    KEY idx_event_comments_event (event_id, created_at),
    KEY idx_event_comments_author (author_id),
    CONSTRAINT fk_event_comments_event FOREIGN KEY (event_id) REFERENCES alumni_events(id) ON DELETE CASCADE,
    CONSTRAINT fk_event_comments_author FOREIGN KEY (author_id) REFERENCES accounts(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- +goose Down
DROP TABLE event_comments;
