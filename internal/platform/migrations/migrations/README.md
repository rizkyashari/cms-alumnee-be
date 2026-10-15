Add versioned SQL migrations here using the name format `000001_short_description.sql`.
Use one MySQL statement per file, make created tables explicitly use InnoDB and utf8mb4,
and never edit a migration after it has been applied. The migration runner records each
version in `schema_migrations`; domain schema is added with its owning module.
