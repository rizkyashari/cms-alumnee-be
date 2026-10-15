-- +goose Up
CREATE TABLE audit_events (
    id BINARY(16) NOT NULL PRIMARY KEY,
    actor_id BINARY(16) NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id BINARY(16) NULL,
    metadata JSON NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    KEY idx_audit_events_created_at (created_at),
    KEY idx_audit_events_actor_id_created_at (actor_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE outbox_events (
    id BINARY(16) NOT NULL PRIMARY KEY,
    event_type VARCHAR(150) NOT NULL,
    payload JSON NOT NULL,
    occurred_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    processed_at TIMESTAMP(6) NULL,
    attempts INT UNSIGNED NOT NULL DEFAULT 0,
    KEY idx_outbox_events_unprocessed (processed_at, occurred_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- +goose Down
DROP TABLE outbox_events;
DROP TABLE audit_events;
