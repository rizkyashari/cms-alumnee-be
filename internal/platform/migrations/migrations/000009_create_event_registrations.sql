-- +goose Up
CREATE TABLE event_registrations (
    id BINARY(16) NOT NULL PRIMARY KEY,
    event_id BINARY(16) NOT NULL,
    profile_id BINARY(16) NOT NULL,
    status ENUM('registered', 'cancelled', 'attended') NOT NULL DEFAULT 'registered',
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    UNIQUE KEY uq_event_registrations_event_profile (event_id, profile_id),
    KEY idx_event_registrations_profile (profile_id, status),
    CONSTRAINT fk_event_registrations_event FOREIGN KEY (event_id) REFERENCES alumni_events(id) ON DELETE CASCADE,
    CONSTRAINT fk_event_registrations_profile FOREIGN KEY (profile_id) REFERENCES alumni_profiles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- +goose Down
DROP TABLE event_registrations;
