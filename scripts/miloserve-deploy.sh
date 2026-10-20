#!/usr/bin/env bash
set -euo pipefail

# Deploy the things app on miloserve. Run from the repo root after git pull:
#   bash scripts/miloserve-deploy.sh
#
# Dumps the database, then rebuilds and restarts only the app container.
# The db container and its volume are never modified.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ -n "${TMPDIR:-}" && ! -w "$TMPDIR" ]]; then
  echo "TMPDIR ($TMPDIR) is not writable; unsetting it." >&2
  unset TMPDIR
fi

BACKUP_DIR="${ROOT}/backups"
KEEP_BACKUPS=10

if [[ ! -f .miloserve-migrated ]]; then
  echo "This server hasn't been migrated yet. Run scripts/miloserve-migrate.sh first." >&2
  exit 1
fi

docker_cmd() {
  if docker info >/dev/null 2>&1; then
    docker "$@"
    return
  fi
  if [[ "$(uname -s)" == "Darwin" ]]; then
    echo "Docker is not responding. Start Docker Desktop and run this again." >&2
    exit 1
  fi
  sudo docker "$@"
}

compose_cmd() {
  docker_cmd compose "$@"
}

dump_db() {
  if [[ -z "$(compose_cmd ps --status running -q db)" ]]; then
    echo "db container is not running; refusing to deploy without a fresh dump." >&2
    exit 1
  fi

  umask 077
  mkdir -p "$BACKUP_DIR"
  local dump_file
  dump_file="${BACKUP_DIR}/tasks-$(date -u +%Y%m%dT%H%M%SZ).sql"

  echo "Dumping database to $dump_file..."
  if ! compose_cmd exec -T db sh -c \
    'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" exec mysqldump -uroot --single-transaction --routines --triggers tasks' \
    > "${dump_file}.partial"; then
    echo "mysqldump failed; aborting deploy. Partial output is at ${dump_file}.partial" >&2
    exit 1
  fi
  if ! tail -n 1 "${dump_file}.partial" | grep -q '^-- Dump completed'; then
    echo "Dump looks truncated; aborting deploy. Partial output is at ${dump_file}.partial" >&2
    exit 1
  fi
  mv "${dump_file}.partial" "$dump_file"
}

prune_backups() {
  ls -1t "$BACKUP_DIR"/tasks-*.sql | tail -n +$((KEEP_BACKUPS + 1)) | xargs -r rm --
}

start_app() {
  echo "Building and restarting app and backup containers..."
  compose_cmd up -d --build app backup
}

verify_app() {
  echo "Checking app health..."
  for _ in $(seq 1 15); do
    if curl -sf -o /dev/null http://127.0.0.1:8888/login; then
      echo "App is responding on 127.0.0.1:8888"
      return
    fi
    sleep 2
  done
  echo "App not responding on 127.0.0.1:8888" >&2
  compose_cmd logs app --tail 50
  exit 1
}

dump_db
prune_backups
start_app
verify_app

echo "Deploy finished."
