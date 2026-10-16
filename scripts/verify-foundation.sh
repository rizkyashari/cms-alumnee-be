#!/usr/bin/env bash
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
    rm -f "$temp_dir/api" "$temp_dir/api.log" "$temp_dir/auth.cookies" \
      "$temp_dir/non_verified.cookies" "$temp_dir/verified.cookies" \
      "$temp_dir/officer.cookies" "$temp_dir/admin.cookies" \
      "$temp_dir/staging-non_verified.cookies" "$temp_dir/staging-verified.cookies" \
      "$temp_dir/staging-officer.cookies" "$temp_dir/staging-admin.cookies"
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

innodb_tables="$(mysql_query "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name IN ('audit_events', 'outbox_events', 'accounts', 'account_contacts', 'external_identities', 'alumni_profiles', 'roles', 'account_roles', 'sessions', 'password_credentials') AND engine = 'InnoDB';")" || fail "Could not verify InnoDB tables"
[[ "$innodb_tables" == "10" ]] || fail "Expected the foundation and auth tables to use InnoDB"
pass "foundation and auth tables use InnoDB"

applied_migration="$(mysql_query 'SELECT MAX(version_id) FROM goose_db_version WHERE is_applied = 1;')" || fail "Could not verify Goose migration metadata"
[[ "$applied_migration" == "3" ]] || fail "Expected latest Goose migration version 3; got ${applied_migration:-none}"
pass "Goose migration version $applied_migration is applied"

printf '\n== Health endpoints and structured request logs ==\n'
go build -o "$temp_dir/api" ./cmd/api || fail "Could not build API"
demo_password="verify-only-local-password"
HTTP_ADDR="127.0.0.1:$HTTP_PORT" MYSQL_HOST="$MYSQL_HOST" MYSQL_PORT="$MYSQL_PORT" \
  DEMO_AUTH_ENABLED=true DEMO_AUTH_PASSWORD="$demo_password" \
  STAGING_AUTH_NON_VERIFIED_USERNAME=verify-nonverified STAGING_AUTH_NON_VERIFIED_PASSWORD=verify-nonverified-password \
  STAGING_AUTH_VERIFIED_USERNAME=verify-verified STAGING_AUTH_VERIFIED_PASSWORD=verify-verified-password \
  STAGING_AUTH_OFFICER_USERNAME=verify-officer STAGING_AUTH_OFFICER_PASSWORD=verify-officer-password \
  STAGING_AUTH_ADMIN_USERNAME=verify-admin STAGING_AUTH_ADMIN_PASSWORD=verify-admin-password \
  PUBLIC_BASE_URL=http://localhost:3001 TURNSTILE_SITE_KEY= TURNSTILE_SECRET_KEY= \
  GOOGLE_OAUTH_CLIENT_ID= GOOGLE_OAUTH_CLIENT_SECRET= \
  FACEBOOK_OAUTH_CLIENT_ID= FACEBOOK_OAUTH_CLIENT_SECRET= \
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

oauth_config="$(curl --silent --show-error --fail --max-time 3 "$HEALTH_BASE_URL/api/v1/auth/oauth/providers")" ||
  fail "Could not read public OAuth provider configuration"
python3 -c '
import json,sys
content=json.loads(sys.argv[1])["content"]
assert content["providers"] == [], content
assert content["captcha_site_key"] == "", content
assert content["password_login_enabled"] is True, content
' "$oauth_config" || fail "Unconfigured OAuth providers or CAPTCHA key were exposed"
pass "public auth configuration exposes no disabled provider credentials"

printf '\n== Role login and session authorization ==\n'
anonymous_status="$(curl --silent --show-error --output /dev/null --write-out '%{http_code}' \
  --max-time 3 "$HEALTH_BASE_URL/api/v1/auth/me")" || fail "Could not test anonymous auth request"
[[ "$anonymous_status" == "401" ]] || fail "Anonymous /auth/me returned $anonymous_status, expected 401"
pass "anonymous session request is rejected"
anonymous_directory_status="$(curl --silent --show-error --output /dev/null --write-out '%{http_code}' \
  --max-time 3 "$HEALTH_BASE_URL/api/v1/alumni")" || fail "Could not test anonymous directory request"
[[ "$anonymous_directory_status" == "401" ]] || fail "Anonymous directory returned $anonymous_directory_status, expected 401"
pass "anonymous directory access is rejected"

for role_email in \
  "non_verified:nonverified@demo.local:directory:read_limited" \
  "verified:verified@demo.local:directory:read_full" \
  "officer:officer@demo.local:documents:manage" \
  "admin:admin@demo.local:alumni:verify"; do
  IFS=: read -r expected_role email expected_capability <<<"$role_email"
  login_payload="$(python3 -c 'import json,sys; print(json.dumps({"email":sys.argv[1],"password":sys.argv[2]}))' "$email" "$demo_password")"
  cookie_file="$temp_dir/$expected_role.cookies"
  login_response="$(curl --silent --show-error --fail --max-time 3 \
    --cookie-jar "$cookie_file" \
    --header 'Content-Type: application/json' \
    --data "$login_payload" "$HEALTH_BASE_URL/api/v1/auth/demo-login")" ||
    fail "Demo login failed for the $expected_role role"
  python3 -c '
import json,sys
response=json.loads(sys.argv[1])
user=response["content"]["user"]
expected={
    "non_verified": {"content:read_public","directory:read_limited","profile:edit_self","business:submit","alumni:attest"},
    "verified": {"content:read_public","directory:read_full","profile:edit_self","business:submit","content:comment","alumni:attest"},
    "officer": {"content:read_public","documents:manage","alumni:attest"},
    "admin": {"content:read_public","directory:read_full","profile:edit_self","business:submit","content:comment","documents:manage","alumni:attest","admin:manage","alumni:verify"},
}
assert user["role"] == sys.argv[2], (user["role"], sys.argv[2])
assert set(user["capabilities"]) == expected[sys.argv[2]], (user["capabilities"], expected[sys.argv[2]])
assert sys.argv[3] in user["capabilities"], (sys.argv[3], user["capabilities"])
' "$login_response" "$expected_role" "$expected_capability" ||
    fail "Role capability set mismatch for $expected_role"
  current_user="$(curl --silent --show-error --fail --max-time 3 \
    --cookie "$cookie_file" "$HEALTH_BASE_URL/api/v1/auth/me")" ||
    fail "Session lookup failed for the $expected_role role"
  python3 -c '
import json,sys
response=json.loads(sys.argv[1])["content"]["user"]
login=json.loads(sys.argv[3])["content"]["user"]
assert response["role"] == sys.argv[2], (response["role"], sys.argv[2])
assert response["capabilities"] == login["capabilities"], (response["capabilities"], login["capabilities"])
' \
    "$current_user" "$expected_role" "$login_response" || fail "Session role or capabilities mismatch for $expected_role"
  pass "$expected_role login, session cookie, and exact capability set"
done

kill "$api_pid" 2>/dev/null || true
wait "$api_pid" 2>/dev/null || true
api_pid=""
HTTP_ADDR="127.0.0.1:$HTTP_PORT" MYSQL_HOST="$MYSQL_HOST" MYSQL_PORT="$MYSQL_PORT" \
  DEMO_AUTH_ENABLED=true DEMO_AUTH_PASSWORD="$demo_password" \
  STAGING_AUTH_NON_VERIFIED_USERNAME=verify-nonverified STAGING_AUTH_NON_VERIFIED_PASSWORD=verify-nonverified-password \
  STAGING_AUTH_VERIFIED_USERNAME=verify-verified STAGING_AUTH_VERIFIED_PASSWORD=verify-verified-password \
  STAGING_AUTH_OFFICER_USERNAME=verify-officer STAGING_AUTH_OFFICER_PASSWORD=verify-officer-password \
  STAGING_AUTH_ADMIN_USERNAME=verify-admin STAGING_AUTH_ADMIN_PASSWORD=verify-admin-password \
  PUBLIC_BASE_URL=http://localhost:3001 TURNSTILE_SITE_KEY= TURNSTILE_SECRET_KEY= \
  GOOGLE_OAUTH_CLIENT_ID= GOOGLE_OAUTH_CLIENT_SECRET= \
  FACEBOOK_OAUTH_CLIENT_ID= FACEBOOK_OAUTH_CLIENT_SECRET= \
  "$temp_dir/api" >>"$temp_dir/api.log" 2>&1 &
api_pid=$!
staging_api_ready=0
for _ in $(seq 1 30); do
  if curl --silent --show-error --max-time 2 "$HEALTH_BASE_URL/api/v1/health/live" >/dev/null 2>&1; then
    staging_api_ready=1
    break
  fi
  if ! kill -0 "$api_pid" 2>/dev/null; then
    fail "API process exited before staging login checks"
  fi
  sleep 1
done
[[ "$staging_api_ready" == "1" ]] || fail "API did not restart for staging login checks"

for role_login in \
  "non_verified:verify-nonverified:verify-nonverified-password:directory:read_limited" \
  "verified:verify-verified:verify-verified-password:directory:read_full" \
  "officer:verify-officer:verify-officer-password:documents:manage" \
  "admin:verify-admin:verify-admin-password:alumni:verify"; do
  IFS=: read -r expected_role username password expected_capability <<<"$role_login"
  login_payload="$(python3 -c 'import json,sys; print(json.dumps({"username":sys.argv[1],"password":sys.argv[2]}))' "$username" "$password")"
  cookie_file="$temp_dir/staging-$expected_role.cookies"
  login_response="$(curl --silent --show-error --fail --max-time 3 \
    --cookie-jar "$cookie_file" \
    --header 'Content-Type: application/json' \
    --header 'Origin: http://localhost:3001' \
    --data "$login_payload" "$HEALTH_BASE_URL/api/v1/auth/password-login")" ||
    fail "Staging password login failed for the $expected_role role"
  python3 -c '
import json,sys
user=json.loads(sys.argv[1])["content"]["user"]
assert user["role"] == sys.argv[2], (user["role"], sys.argv[2])
assert sys.argv[3] in user["capabilities"], (sys.argv[3], user["capabilities"])
' "$login_response" "$expected_role" "$expected_capability" ||
    fail "Staging account role or capability mismatch for $expected_role"
  current_user="$(curl --silent --show-error --fail --max-time 3 \
    --cookie "$cookie_file" "$HEALTH_BASE_URL/api/v1/auth/me")" ||
    fail "Staging session lookup failed for the $expected_role role"
  python3 -c '
import json,sys
user=json.loads(sys.argv[1])["content"]["user"]
assert user["role"] == sys.argv[2], (user["role"], sys.argv[2])
' "$current_user" "$expected_role" || fail "Staging session role mismatch for $expected_role"
  pass "staging username/password login and server-assigned $expected_role role"
done

bad_login_status="$(curl --silent --show-error --output /dev/null --write-out '%{http_code}' \
  --max-time 3 --header 'Content-Type: application/json' \
  --header 'Origin: http://localhost:3001' \
  --data '{"username":"verify-admin","password":"incorrect-staging-password"}' \
  "$HEALTH_BASE_URL/api/v1/auth/password-login")" || fail "Could not test invalid staging credentials"
[[ "$bad_login_status" == "401" ]] || fail "Invalid staging credentials returned $bad_login_status, expected 401"
pass "invalid staging credentials are rejected"

directory_limited="$(curl --silent --show-error --fail --max-time 3 \
  --cookie "$temp_dir/non_verified.cookies" "$HEALTH_BASE_URL/api/v1/alumni")" ||
  fail "Non-verified directory request failed"
python3 -c '
import json,sys
items=json.loads(sys.argv[1])["content"]["items"]
assert items, "expected seeded alumni profiles"
for item in items:
    assert set(item) == {"name", "batch_year"}, item
' "$directory_limited" || fail "Non-verified directory exposed fields beyond name and batch year"
pass "non-verified directory returns only name and batch year"

directory_full="$(curl --silent --show-error --fail --max-time 3 \
  --cookie "$temp_dir/verified.cookies" "$HEALTH_BASE_URL/api/v1/alumni")" ||
  fail "Verified directory request failed"
python3 -c '
import json,sys
items=json.loads(sys.argv[1])["content"]["items"]
assert items, "expected seeded alumni profiles"
for item in items:
    assert {"name", "batch_year", "city", "occupation", "bio"} <= set(item), item
    assert "email" not in item and "account_id" not in item and "phone" not in item, item
' "$directory_full" || fail "Verified directory fields or phone privacy were incorrect"
pass "verified directory returns full profile fields and hides private phone numbers"

officer_directory_status="$(curl --silent --show-error --output /dev/null --write-out '%{http_code}' \
  --cookie "$temp_dir/officer.cookies" --max-time 3 "$HEALTH_BASE_URL/api/v1/alumni")" ||
  fail "Could not test officer directory restriction"
[[ "$officer_directory_status" == "403" ]] || fail "Officer directory returned $officer_directory_status, expected 403"
pass "officer without a directory capability is denied"

curl --silent --show-error --fail --max-time 3 \
  --request POST --header 'Origin: http://localhost:3001' --cookie "$temp_dir/admin.cookies" \
  "$HEALTH_BASE_URL/api/v1/auth/logout" >/dev/null ||
  fail "Session logout failed"
revoked_status="$(curl --silent --show-error --output /dev/null --write-out '%{http_code}' \
  --cookie "$temp_dir/admin.cookies" --max-time 3 "$HEALTH_BASE_URL/api/v1/auth/me")" ||
  fail "Could not verify revoked session"
[[ "$revoked_status" == "401" ]] || fail "Revoked session returned $revoked_status, expected 401"
pass "logout revokes the server-side session"

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
