# AlumniHub API

Go modular monolith for AlumniHub. The production topology is one Go API process, MySQL 8/InnoDB, and Caddy serving the frontend static export and proxying `/api/*`. Redis, PostgreSQL, Strapi, and separately deployed workers are not part of this target.

## Runtime and API

- Go HTTP server using `net/http` and Chi.
- Versioned API prefix: `/api/v1`; the initial implemented endpoints are:
  - `GET /api/v1/health/live` — process liveness; does not require MySQL.
  - `GET /api/v1/health/ready` — readiness; checks MySQL and does not expose internal errors.
- JSON responses retain the `{ "message": ..., "content": ... }` envelope.
- Structured JSON request logs omit query strings. Recovery responds with a generic 500.
- `docs/openapi.yaml` is the initial API contract. Add domain routes and schemas there as modules are implemented.

The frontend remains a prototype: most content is mock data and its demo login is local-only. OTP/OAuth, server-side sessions, profile and directory APIs, admin authorization, and domain persistence are not implemented by this foundation and must not be treated as production-ready.

## Local development

Use Go `1.27.1` or a compatible newer version. Docker is not required. Podman can provide a disposable MySQL 8.4 instance for verification:

```sh
podman machine start
```

For direct API development, use a local MySQL 8 instance or a database endpoint approved for this project. The `make verify` target defaults to Podman and starts a temporary official MySQL 8.4 container. It uses the database name and app credentials in `.env`; it does not require a MySQL client installed on the Mac. The container has no persistent volume and is removed when verification finishes. To verify against an already-running local MySQL server instead, set `MYSQL_RUNTIME=local` and install the `mysql` client.

```sh
test -f .env || cp .env.example .env
# Edit .env locally. Do not commit it or share its contents.
make migrate
make dev
```

The Make targets read simple `KEY=value` entries from `.env` without executing the file as shell code. The API listens on `HTTP_ADDR` (default `127.0.0.1:3000`). In another terminal:

```sh
curl -i http://127.0.0.1:3000/api/v1/health/live
curl -i http://127.0.0.1:3000/api/v1/health/ready
```

For deployment, Compose remains available. Local verification uses Podman by default.

## Database and migrations

MySQL connection settings use `MYSQL_HOST`, `MYSQL_PORT`, `MYSQL_USER`, `MYSQL_PASSWORD`, and `MYSQL_DATABASE`. The API does not run schema changes during startup. Apply versioned SQL migrations explicitly:

```sh
make migrate
```

Migrations use Goose and live in `internal/platform/migrations/migrations/`. Do not edit migrations that have been applied. A small initial migration creates the audit-event and transactional-outbox schema foundations; the outbox processing loop and domain tables are not implemented yet. sqlc configuration is in `sqlc.yaml`; add query files as domain modules are introduced.

## Checks

```sh
gofmt -w $(find cmd internal middleware -name '*.go' -type f)
go test ./...
go vet ./...
```

To run the integrated checklist for Git secret handling, MySQL/InnoDB, migrations, health responses, structured request logs, tests, and `go vet` using Podman:

```sh
test -f .env || cp .env.example .env
# Set MYSQL_USER, MYSQL_PASSWORD, and MYSQL_DATABASE in .env; do not share or commit secrets.
make verify
```

The verifier runs unit tests, starts a fresh temporary MySQL container, applies migrations, checks InnoDB and health/log behavior, then removes both the database container and temporary API process. It does not affect other containers or database data. Podman must be installed, and its VM must be running.

Frontend static export and lint:

```sh
cd ../cms-alumnee-fe
yarn lint
STATIC_EXPORT=true yarn build
```

See the AlumniHub entity relationships and allowed status values in [`docs/ERD.md`](docs/ERD.md), and the detailed deployment/API topology in [`topology-and-tech-stack.md`](../cms-alumnee-fe/docs/architecture/topology-and-tech-stack.md).
