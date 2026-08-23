#!/usr/bin/env bash

set -Eeuo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$REPO_DIR/backend"
SERVICE_NAME="${SERVICE_NAME:-mangahost.service}"
LOCAL_HEALTH_URL="${LOCAL_HEALTH_URL:-http://127.0.0.1:4690/health}"
PUBLIC_HEALTH_URL="${PUBLIC_HEALTH_URL:-https://api-manga.chiraitori.dev/health}"
SMOKE_PORT="${SMOKE_PORT:-4691}"
NEXT_BINARY="$BACKEND_DIR/server.next"
LIVE_BINARY="$BACKEND_DIR/server"
PREVIOUS_BINARY="$BACKEND_DIR/server.previous"
LOCK_FILE="/tmp/r2apimangahost-deploy.lock"
SMOKE_LOG=""
SMOKE_PID=""

log() {
  printf '\n\033[1;36m==> %s\033[0m\n' "$*"
}

die() {
  printf '\n\033[1;31mDeploy failed: %s\033[0m\n' "$*" >&2
  exit 1
}

cleanup() {
  if [[ -n "$SMOKE_PID" ]] && kill -0 "$SMOKE_PID" 2>/dev/null; then
    kill "$SMOKE_PID" 2>/dev/null || true
    wait "$SMOKE_PID" 2>/dev/null || true
  fi
  [[ -n "$SMOKE_LOG" ]] && rm -f -- "$SMOKE_LOG"
  rm -f -- "$NEXT_BINARY"
}
trap cleanup EXIT INT TERM

wait_for_health() {
  local url="$1"
  local attempts="${2:-20}"
  local delay="${3:-1}"
  local i

  for ((i = 1; i <= attempts; i++)); do
    if curl --fail --silent --max-time 3 "$url" >/dev/null; then
      return 0
    fi
    sleep "$delay"
  done
  return 1
}

for command in git go curl flock sudo systemctl; do
  command -v "$command" >/dev/null 2>&1 || die "missing required command: $command"
done

[[ -d "$REPO_DIR/.git" ]] || die "$REPO_DIR is not a Git repository"
[[ -f "$BACKEND_DIR/.env" ]] || die "missing $BACKEND_DIR/.env"

exec 9>"$LOCK_FILE"
flock -n 9 || die "another deployment is already running"

if [[ -n "$(git -C "$REPO_DIR" status --porcelain --untracked-files=normal)" ]]; then
  git -C "$REPO_DIR" status --short
  die "repository has local changes; commit or discard them before deploying"
fi

log "Backing up .env"
ENV_BACKUP_DIR="$HOME/.mangahost-env-backups"
ENV_BACKUP="$ENV_BACKUP_DIR/.env-$(date -u +%Y%m%d-%H%M%S)"
mkdir -p "$ENV_BACKUP_DIR"
cp "$BACKEND_DIR/.env" "$ENV_BACKUP"
chmod 600 "$ENV_BACKUP" "$BACKEND_DIR/.env"

log "Pulling origin/main"
git -C "$REPO_DIR" fetch origin main
git -C "$REPO_DIR" checkout main
git -C "$REPO_DIR" pull --ff-only origin main
git -C "$REPO_DIR" submodule sync --recursive
git -C "$REPO_DIR" submodule update --init --recursive

log "Building backend"
cd "$BACKEND_DIR"
if [[ "${SKIP_TESTS:-0}" != "1" ]]; then
  GOTOOLCHAIN=auto go test ./...
fi
GOTOOLCHAIN=auto go build -trimpath -ldflags="-s -w" -o "$NEXT_BINARY" ./cmd/server
chmod 755 "$NEXT_BINARY"

log "Smoke-testing on port $SMOKE_PORT"
SMOKE_LOG="$(mktemp /tmp/mangahost-smoke.XXXXXX.log)"
PORT="$SMOKE_PORT" "$NEXT_BINARY" >"$SMOKE_LOG" 2>&1 &
SMOKE_PID="$!"
if ! wait_for_health "http://127.0.0.1:$SMOKE_PORT/health" 20 1; then
  tail -n 50 "$SMOKE_LOG" >&2 || true
  die "smoke test failed"
fi
kill "$SMOKE_PID" 2>/dev/null || true
wait "$SMOKE_PID" 2>/dev/null || true
SMOKE_PID=""

log "Installing and restarting $SERVICE_NAME"
if [[ -f "$LIVE_BINARY" ]]; then
  cp -p "$LIVE_BINARY" "$PREVIOUS_BINARY"
fi
mv -f "$NEXT_BINARY" "$LIVE_BINARY"

if ! sudo systemctl restart "$SERVICE_NAME" || ! wait_for_health "$LOCAL_HEALTH_URL" 25 1; then
  printf '\nNew service failed health check; rolling back binary...\n' >&2
  if [[ -f "$PREVIOUS_BINARY" ]]; then
    cp -p "$PREVIOUS_BINARY" "$LIVE_BINARY"
    sudo systemctl restart "$SERVICE_NAME"
    wait_for_health "$LOCAL_HEALTH_URL" 25 1 || true
  fi
  sudo systemctl status "$SERVICE_NAME" --no-pager -l >&2 || true
  die "service rollout failed and rollback was attempted"
fi

if ! curl --fail --silent --show-error --max-time 10 "$PUBLIC_HEALTH_URL" >/dev/null; then
  printf '\nWarning: local health passed but public health failed: %s\n' "$PUBLIC_HEALTH_URL" >&2
fi

log "Deployment successful"
printf 'Commit: %s\n' "$(git -C "$REPO_DIR" rev-parse --short HEAD)"
printf 'Service: %s\n' "$(systemctl is-active "$SERVICE_NAME")"
printf 'Local health: %s\n' "$LOCAL_HEALTH_URL"
printf 'Environment backup: %s\n' "$ENV_BACKUP"
