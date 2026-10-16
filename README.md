# AlumniHub API

Go modular monolith for AlumniHub. The production topology is one Go API process, MySQL 8/InnoDB, and Caddy serving the frontend static export and proxying `/api/*`. Redis, PostgreSQL, Strapi, and separately deployed workers are not part of this target.

## Runtime and API

- Go HTTP server using `net/http` and Chi.
- Versioned API prefix: `/api/v1`; the initial implemented endpoints are:
  - `GET /api/v1/health/live` — process liveness; does not require MySQL.
  - `GET /api/v1/health/ready` — readiness; checks MySQL and does not expose internal errors.
- Authentication: `POST /api/v1/auth/demo-login` (loopback-only local testing),
  `POST /api/v1/auth/password-login` (optional four-account staging login),
  `POST /api/v1/auth/email-otp/start`, `POST /api/v1/auth/email-otp/verify`,
  and `POST /api/v1/auth/register` for passwordless email registration/login,
  `GET /api/v1/auth/me`,
  `POST /api/v1/auth/logout`, and OAuth authorization-code routes for Google
  and Facebook.
- `POST /api/v1/auth/profile/complete` completes a new OAuth member profile.
- `GET /api/v1/admin/roles` and `POST /api/v1/admin/roles/{grant|revoke}` let
  Admins grant/revoke Admin and Pengurus roles for existing accounts with a
  verified email.
- Verification workflow: `GET /api/v1/verification/me`,
  `POST /api/v1/verification/request`, `GET /api/v1/verification/candidates`,
  `POST /api/v1/verification/{profileId}/attest` (any same-batch member with a profile),
  `POST /api/v1/verification/{profileId}/bootstrap` (one-time first-member
  exception for Pengurus/Admin), and Admin `GET /api/v1/admin/verifications`,
  `GET /api/v1/admin/verifications/audit`, plus `POST /api/v1/admin/verifications/{profileId}/{approve|reject}`.
  Approval requires five valid same-batch attestations.
  Step-by-step staging walkthrough with ready-made accounts:
  [`docs/staging-walkthrough.md`](docs/staging-walkthrough.md).
- Profile photos: `GET /api/v1/auth/profile/photo` (own) and
  `GET /api/v1/verification/{profileId}/photo` (reviewers of a pending request).
- `GET /api/v1/alumni` — authenticated directory with server-side
  non-verified/verified/admin response projection.
- JSON responses retain the `{ "message": ..., "content": ... }` envelope.
- Structured JSON request logs omit query strings. Recovery responds with a generic 500.
- Caddy access logs omit request URIs and redact Cookie/Authorization headers so
  OAuth callback codes, state, and session credentials are not logged.
- `docs/openapi.yaml` is the initial API contract. Add domain routes and schemas there as modules are implemented.

Target registration, OTP-channel tradeoffs, Kabupaten/Kota handling, photo
storage options, and the Admin bootstrap/management model are documented in
[Authentication, registration, and member lifecycle](docs/auth-registration.md).
That document distinguishes the confirmed first-member verification exception
from features that still need implementation.

Most content remains mock data. The local demo login uses backend-issued,
revocable `HttpOnly` sessions and server-side role capabilities; it is not a
production credential system. Email OTP uses a configured Resend sender and
the member profile photo is stored as sanitized JPEG/PNG bytes in MySQL. The
first Admin bootstrap is restricted to the name/email configured by the
operator and completes only after that mailbox verifies an OTP. OAuth login
can be enabled with Google/Facebook credentials and server-verified Cloudflare
Turnstile CAPTCHA. Password recovery, profile editing after registration,
comments, listings, documents, and phone (SMS/WhatsApp) OTP are not implemented and must not be treated as production-ready.

Authentication middleware resolves the active account, role assignments, and
member verification status from the server-side session on every protected
request. Capabilities are derived from that effective role on the server;
caller-provided capability values are never used for authorization. Missing or
invalid sessions receive `401`; authenticated roles without the required
capability receive `403`; session-store failures receive `503`. Public-content
capability middleware permits anonymous requests.

The alumni directory is the first business endpoint with enforced role-based
data filtering: anonymous users receive `401`, officers without directory
capability receive `403`, non-verified members receive only name and batch
year, and verified members/admins receive the permitted profile fields. Email
and account identifiers are not returned; phone is returned only when a
verified primary phone contact has directory visibility enabled. Unit tests
exercise every role/capability authorization pair, including anonymous,
forged-capability, and unknown-role denials; `make verify` checks the exact
capability sets on database-backed demo sessions and directory denials.
Profile-onboarding routes and Admin/Pengurus role management are implemented.
Business routes for profile editing after onboarding, listing submission,
comments, and documents are not implemented yet, so those capabilities do not
make those features available.

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

### Local role testing

With a local MySQL 8 database available, copy `.env.example` to `.env`, set the
MySQL credentials, set `DEMO_AUTH_ENABLED=true`, and choose a private
`DEMO_AUTH_PASSWORD` of at least 12 characters. Then run the API and frontend:

```sh
# Backend terminal
make migrate
make dev

# Frontend terminal
cd ../cms-alumnee-fe
NEXT_PUBLIC_DEMO_AUTH_ENABLED=true yarn dev --port 3001
```

Open `http://localhost:3001/login`. Next.js proxies `/api/*` to the backend on
port 3000. This demo mode requires the API to bind only to loopback
(`127.0.0.1`, `::1`, or `localhost`) and must not be enabled on a shared or
public server. The endpoint also rejects non-loopback request hosts and
origins, so it cannot be used through a public Caddy hostname. All four test
accounts use the same configured password:

| Effective access | Email |
| --- | --- |
| Non-verified | `nonverified@demo.local` |
| Verified | `verified@demo.local` |
| Pengurus | `officer@demo.local` |
| Admin | `admin@demo.local` |

Public access is anonymous. Demo accounts are created on first use; the
verified demo/staging account is seeded as verified only to test its
capabilities and does not simulate five attestations or admin approval.

The optional staging username/password login provisions four separate test
accounts (non-verified member, verified member, officer, admin) from
`STAGING_AUTH_{NON_VERIFIED,VERIFIED,OFFICER,ADMIN}_{USERNAME,PASSWORD}`.
Configure all eight variables together; usernames must be unique, and each
password must have at least 12 characters. Passwords are bcrypt-hashed in the
database. On API startup, configured accounts are provisioned with
server-assigned roles, and staging credentials no longer present in `.env` are
disabled with their existing sessions revoked; changing a configured password
or role also revokes that account's existing sessions. This is for controlled
staging access, not public self-registration or production account management.
Keep both `DEMO_AUTH_ENABLED` and `NEXT_PUBLIC_DEMO_AUTH_ENABLED` set to
`false` on a public staging server.
Apply migrations before starting the updated API. Keep `.env` private and share
the four credentials with testers out of band; the login UI does not display
example credentials.

The frontend login uses these backend endpoints and does not store roles in
`localStorage`. Provider availability and the public Turnstile site key are
discovered from
`GET /api/v1/auth/oauth/providers`; the frontend does not need provider flags or
a duplicated public key at build time. Put provider credentials and the
Turnstile secret only in backend `.env`.
`PUBLIC_BASE_URL` must match provider callback URLs:
`{PUBLIC_BASE_URL}/api/v1/auth/oauth/{provider}/callback`.

OAuth account linking across providers is not implemented. Facebook email is
not treated as verified unless a verified claim is available; a new OAuth
account must complete the alumni profile before receiving member capabilities.

### Email OTP and first Admin setup

Configure a Resend account and verify the sender address/domain. Add
`RESEND_API_KEY` and `RESEND_FROM_EMAIL` to the private backend `.env`. Set
`PUBLIC_BASE_URL` to the frontend's exact origin (for example, the localhost
frontend origin during local development); same-origin checks protect OTP and
role-changing requests. Set `BOOTSTRAP_ADMIN_NAME` and
`BOOTSTRAP_ADMIN_EMAIL` to the project-owner-supplied first Admin identity.
These bootstrap variables must be configured together with the email sender.
Do not paste API keys into source files, chat, or `.env.example`.

Apply database migrations with `make migrate`, start the API, then open
`/registrasi` and submit the configured Admin email address. The first Admin
account is created only after entering the emailed OTP. This initial account
does not receive a member profile. Once signed in, the Admin can add other
Admin/Pengurus accounts in **Panel Admin → Pengaturan**; each target must first
register and verify their own email. The API prevents removal of the last
active Admin. No developer/DB operation is needed for routine staff additions.

New alumni register at `/registrasi`. Email OTP is currently the only
configured self-service verification channel; the API rejects delivery if
Resend is not configured. Registration creates a Non-verified profile, stores
sanitized photos in MySQL, and reports 90% completeness. The remaining 10%
requires a second verified contact; phone OTP is not available until a
provider is selected.

Domisili codes are format-checked by default. To also check them against the
kabupaten/kota master list, run `python3 scripts/fetch_regencies.py
data/regencies.json` on a machine with internet access and set
`REGENCY_MASTER_PATH=data/regencies.json`; the API then rejects unknown codes
and fails to start if the file is invalid.

The first verified alumnus is confirmed by a Pengurus/Admin through the
one-time bootstrap endpoint (only while no alumnus is verified; seeded
staging/demo verified accounts count). All later alumni follow the
five-attestation plus Admin-approval path. Run the DB-backed workflow tests
against a disposable MySQL with
`TEST_MYSQL_DSN='user:pass@tcp(127.0.0.1:3399)/db?parseTime=true' go test ./internal/modules/auth/`
(they delete all accounts; never use a shared database).

To enable social login, configure the Google OAuth web client and/or Facebook
Login callback allowlist to exactly
`{PUBLIC_BASE_URL}/api/v1/auth/oauth/google/callback` or
`{PUBLIC_BASE_URL}/api/v1/auth/oauth/facebook/callback`. Configure a Cloudflare
Turnstile widget for the same frontend hostname, then set its site key and
secret together with the selected provider's client ID and client secret in
backend `.env`. Restart the API; `GET /api/v1/auth/oauth/providers` will expose
only fully configured providers and the public CAPTCHA site key. The static
frontend reads this configuration at login time, so no OAuth secrets or
provider-specific frontend build flags are needed. Real provider login cannot
be completed until these project-owned credentials and redirect allowlists are
configured.

Google callback verifies the ID token signature and validates issuer, audience,
expiry, and a per-login nonce in addition to the state cookie. Facebook uses a
state-bound authorization-code exchange and its authenticated user profile;
Facebook email is not stored as verified. Both OAuth starts require successful
server-side Turnstile verification before redirecting to the provider.

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

Migrations use Goose and live in `internal/platform/migrations/migrations/`.
Do not edit migrations that have been applied. Migration `000001` creates the
audit-event/outbox foundations; `000002` creates account, contact, role, alumni
profile, external identity, and session tables; `000003` adds optional
username/password credentials for the four configured staging accounts;
`000004` adds email OTP, registration tokens, profile photos and domisili
codes; `000005` adds verification requests and attestations. The
current migration assumes
this API database does not already contain an `accounts` table. Reconcile the
legacy column and ID types before applying it to a shared LMS database. The
outbox processing loop and business-domain persistence are not implemented yet.
sqlc configuration is in `sqlc.yaml`; add query files as domain modules are
introduced.

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

The verifier runs unit tests, starts a fresh temporary MySQL container, applies migrations, checks InnoDB, auth schema, health responses, and structured logs, then removes the temporary database container and API process. It does not affect other containers or database data. Podman must be installed, and its VM must be running.

Frontend static export and lint:

```sh
cd ../cms-alumnee-fe
yarn lint
STATIC_EXPORT=true yarn build
```

See the AlumniHub entity relationships and allowed status values in [`docs/ERD.md`](docs/ERD.md), and the detailed deployment/API topology in [`topology-and-tech-stack.md`](../cms-alumnee-fe/docs/architecture/topology-and-tech-stack.md).
