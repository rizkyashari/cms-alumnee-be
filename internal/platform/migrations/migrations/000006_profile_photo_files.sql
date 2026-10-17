-- +goose Up
-- A photo now lives either as a file on disk (file_name) or, for rows written
-- before file storage was enabled, as a blob (image_data); never both.
ALTER TABLE alumni_profile_photos
    MODIFY image_data MEDIUMBLOB NULL,
    ADD COLUMN file_name VARCHAR(64) NULL AFTER mime_type,
    ADD CONSTRAINT chk_alumni_profile_photos_source CHECK ((file_name IS NULL) <> (image_data IS NULL));

-- +goose Down
-- Only valid while no photo has been moved to disk (every row still has image_data).
ALTER TABLE alumni_profile_photos
    DROP CHECK chk_alumni_profile_photos_source,
    DROP COLUMN file_name,
    MODIFY image_data MEDIUMBLOB NOT NULL;
