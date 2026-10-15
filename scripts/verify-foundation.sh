✓ Compiled successfully
✓ Generating static pages using 9 workers (31/31)
● /berita/[slug]
  ├ ● /berita/1
  ├ ● /berita/2
  ├ ● /berita/3#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"
source "$ROOT_DIR/scripts/env.sh"

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

pass() {
  printf 'PASS: %s\n' "$*"
}

for command_name in go git curl python3; do
  command -v "$command_name" >/dev/null 2>&1 || fail "$command_name is required"
done

[[ -f .env ]] || fail "Missing .env. Copy .env.example and set MYSQL_PASSWORD locally."
git check-ignore --quiet .env || fail ".env is not ignored by Git"
if git ls-files --error-unmatch -- .env >/dev/null 2>&1; then
  fail ".env is tracked by Git; remove it from tracking before continuing"
fi
load_dotenv .env || fail "Could not load .env"
pass "local .env exists, is ignored by Git, and was loaded without executing shell code"

[[ -n "${MYSQL_USER:-}" ]] || fail "MYSQL_USER must be set in .env"
[[ -n "${MYSQL_PASSWORD:-}" ]] || fail "MYSQL_PASSWORD must be set in .env"
[[ -n "${MYSQL_DATABASE:-}" ]] || fail "MYSQL_DATABASE must be set in .env"
MYSQL_RUNTIME="${MYSQL_RUNTIME:-podman}"
MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
HTTP_PORT="$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1]); s.close()')" ||
  fail "Could not allocate a temporary local API port"
HEALTH_BASE_URL="http://127.0.0.1:${HTTP_PORT}"

printf '\n== Go unit tests and static analysis ==\n'
go test ./... || fail "go test ./... failed"
go vet ./... || fail "go vet ./... failed"
pass "Go tests and vet"

printf '\n== MySQL/InnoDB and migrations ==\n'
temp_dir="$(mktemp -d "${TMPDIR:-/tmp}/cms-alumnee-verify.XXXXXX")"
api_pid=""
mysql_container=""
cleanup() {
  if [[ -n "$api_pid" ]]; then
    kill "$api_pid" 2>/dev/null || true
    wait "$api_pid" 2>/dev/null || true
  fi
  if [[ -n "$mysql_container" ]]; then
    podman rm --force "$mysql_container" >/dev/null 2>&1 || true
  fi
  if [[ -n "${temp_dir:-}" && -d "$temp_dir" ]]; then
    rm -f "$temp_dir/api" "$temp_dir/api.log"
    rmdir "$temp_dir"
  fi
}
trap cleanup EXIT

mysql_query() {
  case "$MYSQL_RUNTIME" in
    podman)
      podman exec "$mysql_container" sh -c \
        'MYSQL_PWD="$MYSQL_PASSWORD" mysql --user="$MYSQL_USER" --database="$MYSQL_DATABASE" --batch --skip-column-names --execute="$1"' \
        sh "$1"
      ;;
    local)
      command -v mysql >/dev/null 2>&1 || fail "mysql client is required with MYSQL_RUNTIME=local"
      MYSQL_PWD="$MYSQL_PASSWORD" mysql \
        --protocol=tcp \
        --host="$MYSQL_HOST" \
        --port="$MYSQL_PORT" \
        --user="$MYSQL_USER" \
        --database="$MYSQL_DATABASE" \
        --batch --skip-column-names \
        --execute="$1"
      ;;
    *)
      fail "MYSQL_RUNTIME must be podman or local"
      ;;
  esac
}

if [[ "$MYSQL_RUNTIME" == "podman" ]]; then
  command -v podman >/dev/null 2>&1 || fail "Podman is required (or set MYSQL_RUNTIME=local with a local MySQL server)"
  podman info >/dev/null 2>&1 || fail "Podman is not ready; start it with 'podman machine start'"
  mysql_container="cms-alumnee-verify-$$"
  podman run --detach --rm \
    --name "$mysql_container" \
    --env MYSQL_ROOT_PASSWORD="$MYSQL_PASSWORD" \
    --env MYSQL_DATABASE \
    --env MYSQL_USER \
    --env MYSQL_PASSWORD \
    --publish 127.0.0.1::3306 \
    docker.io/library/mysql:8.4 \
    --default-storage-engine=InnoDB \
    --character-set-server=utf8mb4 \
    --collation-server=utf8mb4_0900_ai_ci >/dev/null ||
    fail "Could not start temporary MySQL 8 container"
  mapped_port="$(podman port "$mysql_container" 3306/tcp | head -n 1)"
  MYSQL_PORT="${mapped_port##*:}"
  [[ "$MYSQL_PORT" =~ ^[0-9]+$ ]] || fail "Could not determine the temporary MySQL port"
  MYSQL_HOST=127.0.0.1
  pass "started a disposable MySQL 8.4 Podman container on 127.0.0.1:$MYSQL_PORT"
elif [[ "$MYSQL_RUNTIME" == "local" ]]; then
  command -v mysql >/dev/null 2>&1 || fail "mysql client is required with MYSQL_RUNTIME=local"
else
  fail "MYSQL_RUNTIME must be podman or local"
fi

mysql_ready=0
for _ in $(seq 1 90); do
  if mysql_query 'SELECT 1;' >/dev/null 2>&1; then
    mysql_ready=1
    break
  fi
  sleep 2
done
[[ "$mysql_ready" == "1" ]] || fail "MySQL did not become ready; inspect Podman with 'podman logs $mysql_container'"

mysql_version="$(mysql_query 'SELECT VERSION();')" || fail "Could not connect to MySQL; check the local server and credentials"
case "$mysql_version" in
  8.*) pass "MySQL 8 is reachable ($mysql_version)" ;;
  *) fail "Expected MySQL 8, got version $mysql_version" ;;
esac

MYSQL_HOST="$MYSQL_HOST" MYSQL_PORT="$MYSQL_PORT" go run ./cmd/migrate || fail "Database migration failed"

innodb_tables="$(mysql_query "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name IN ('audit_events', 'outbox_events') AND engine = 'InnoDB';")" || fail "Could not verify InnoDB tables"
[[ "$innodb_tables" == "2" ]] || fail "Expected audit_events and outbox_events to use InnoDB"
pass "audit_events and outbox_events use InnoDB"

applied_migration="$(mysql_query 'SELECT MAX(version_id) FROM goose_db_version WHERE is_applied = 1;')" || fail "Could not verify Goose migration metadata"
[[ "$applied_migration" == "1" ]] || fail "Expected initial Goose migration version 1; got ${applied_migration:-none}"
pass "Goose migration version $applied_migration is applied"

printf '\n== Health endpoints and structured request logs ==\n'
go build -o "$temp_dir/api" ./cmd/api || fail "Could not build API"
HTTP_ADDR="127.0.0.1:$HTTP_PORT" MYSQL_HOST="$MYSQL_HOST" MYSQL_PORT="$MYSQL_PORT" \
  "$temp_dir/api" >"$temp_dir/api.log" 2>&1 &
api_pid=$!

live_response=""
for _ in $(seq 1 30); do
  if live_response="$(curl --silent --show-error --max-time 2 "$HEALTH_BASE_URL/api/v1/health/live" 2>/dev/null)"; then
    break
  fi
  if ! kill -0 "$api_pid" 2>/dev/null; then
    fail "API process exited during startup; confirm HTTP_ADDR is available"
  fi
  sleep 1
done
grep -Fq '"status":"ok"' <<<"$live_response" || fail "Liveness endpoint did not return ok"
pass "GET /api/v1/health/live returns ok"

ready_response="$(curl --silent --show-error --fail --max-time 3 "$HEALTH_BASE_URL/api/v1/health/ready")" || fail "Readiness endpoint failed; check MySQL credentials and access"
grep -Fq '"status":"ready"' <<<"$ready_response" || fail "Readiness response was unexpected"
pass "GET /api/v1/health/ready returns ready"

marker="verify-query-secret-${$}-$(date +%s)"
curl --silent --show-error --fail --max-time 3 \
  "$HEALTH_BASE_URL/api/v1/health/live?token=${marker}" >/dev/null ||
  fail "Could not issue request for query-string logging check"

for _ in $(seq 1 10); do
  grep -Fq '"msg":"http request"' "$temp_dir/api.log" && break
  sleep 1
done
grep -Fq '"msg":"http request"' "$temp_dir/api.log" || fail "Structured request log was not found"
grep -Fq '"route":"/api/v1/health/live"' "$temp_dir/api.log" || fail "Request log did not contain the route pattern"
grep -Fq '"status":200' "$temp_dir/api.log" || fail "Request log did not contain the HTTP status"
if grep -Fq "$marker" "$temp_dir/api.log"; then
  fail "Query parameter unexpectedly appeared in API logs"
fi
pass "structured JSON logs contain request fields and omit query parameters"

printf '\nAll foundation checks passed. The temporary MySQL container and test API process will stop automatically.\n'
