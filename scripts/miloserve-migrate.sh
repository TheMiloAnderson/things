#!/usr/bin/env bash
set -euo pipefail

# One-time migration of the things app from systemd + host MySQL to Docker on
# miloserve. Run once from the repo root: bash scripts/miloserve-migrate.sh
# For every deploy after that, use scripts/miloserve-deploy.sh.
#
# Requires sudo for: Docker install, stopping systemd services.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# sudo on macOS runs as root with root's private temp directory. Docker
# Desktop cannot read that directory, so a build started that way fails with
# "can't access os.tempDir".
if [[ -n "${TMPDIR:-}" && ! -w "$TMPDIR" ]]; then
  echo "TMPDIR ($TMPDIR) is not writable; unsetting it." >&2
  unset TMPDIR
fi

SENTINEL="${ROOT}/.miloserve-migrated"
if [[ -f "$SENTINEL" ]]; then
  echo "Migration already completed at $(cat "$SENTINEL")." >&2
  echo "Re-running would overwrite the live database with tasks_dump.sql." >&2
  echo "To deploy, run scripts/miloserve-deploy.sh instead." >&2
  exit 1
fi

if [[ ! -f .env ]]; then
  echo "Missing .env — copy .env.example and fill in values first." >&2
  exit 1
fi

env_val() {
  local key="$1"
  local line
  line="$(grep -E "^${key}=" .env | tail -n 1 || true)"
  line="${line#${key}=}"
  line="${line%\"}"
  line="${line#\"}"
  line="${line%\'}"
  line="${line#\'}"
  printf '%s' "$line"
}

DBUSER="$(env_val DBUSER)"
DBPASS="$(env_val DBPASS)"
MYSQL_ROOT_PASSWORD="$(env_val MYSQL_ROOT_PASSWORD)"

if [[ -z "$DBUSER" || -z "$DBPASS" ]]; then
  echo "DBUSER and DBPASS must be set in .env" >&2
  exit 1
fi

if [[ -z "$MYSQL_ROOT_PASSWORD" ]]; then
  echo "MYSQL_ROOT_PASSWORD missing from .env" >&2
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

install_docker() {
  if command -v docker >/dev/null 2>&1; then
    echo "Docker already installed: $(docker --version)"
    return
  fi

  echo "Installing Docker Engine..."
  sudo apt-get update
  sudo apt-get install -y ca-certificates curl
  sudo install -m 0755 -d /etc/apt/keyrings
  sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
  sudo chmod a+r /etc/apt/keyrings/docker.asc
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
  sudo apt-get update
  sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  sudo usermod -aG docker "$USER"
  echo "Docker installed. You may need to log out/in for group membership; using sudo for compose below."
}

dump_is_complete() {
  [[ -s "$1" ]] && grep -q '^-- Dump completed' "$1"
}

dump_host_mysql() {
  local dump_file="${ROOT}/tasks_dump.sql"
  if dump_is_complete "$dump_file"; then
    echo "Using existing dump: $dump_file"
    return
  fi
  if [[ -f "$dump_file" ]]; then
    echo "Removing incomplete dump: $dump_file"
    rm -f "$dump_file"
  fi

  # --single-transaction issues FLUSH TABLES, which requires RELOAD. The app
  # user does not have it, and a failed mysqldump still leaves its header in
  # the output file. --skip-lock-tables avoids that. Writes during the dump
  # can be torn, so stop the app before migrating a database people are using.
  echo "Dumping host MySQL tasks database..."
  local partial="${dump_file}.partial"
  rm -f "$partial"
  if ! MYSQL_PWD="$DBPASS" mysqldump --no-tablespaces --skip-lock-tables \
    --routines --triggers -u"$DBUSER" tasks > "$partial"; then
    rm -f "$partial"
    echo "mysqldump failed." >&2
    exit 1
  fi
  if ! dump_is_complete "$partial"; then
    rm -f "$partial"
    echo "mysqldump did not finish; no dump was saved." >&2
    exit 1
  fi
  mv "$partial" "$dump_file"
  echo "Dump saved to $dump_file"
}

create_volume() {
  echo "Creating MySQL data volume..."
  docker_cmd volume create things_mysql_data >/dev/null
}

start_db() {
  echo "Starting MySQL container..."
  compose_cmd up -d db

  # The image starts a temporary server on a local socket, loads the
  # password, then shuts it down and starts the real server on port 3306.
  # A socket login succeeds against the temporary server, so an import
  # started then dies with "Lost connection" when that server exits.
  echo "Waiting for MySQL to listen on port 3306..."
  for _ in $(seq 1 60); do
    if compose_cmd exec -T db sh -c \
      'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" exec mysql -uroot -h 127.0.0.1 --protocol=TCP -N -e "SELECT 1"' >/dev/null 2>&1; then
      echo "MySQL is ready."
      return
    fi
    sleep 2
  done
  echo "MySQL did not become healthy in time." >&2
  compose_cmd logs db --tail 50
  exit 1
}

import_dump() {
  local dump_file="${ROOT}/tasks_dump.sql"
  if ! dump_is_complete "$dump_file"; then
    echo "Refusing to import $dump_file; it is missing or incomplete." >&2
    exit 1
  fi

  echo "Importing dump into containerized MySQL..."
  compose_cmd exec -T db sh -c \
    'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" exec mysql -uroot -h 127.0.0.1 --protocol=TCP tasks' < "$dump_file"
}

start_app() {
  echo "Building and starting app container..."
  compose_cmd up -d --build app
}

verify_app() {
  echo "Checking app health..."
  curl -sf -o /dev/null http://127.0.0.1:8888/login || {
    echo "App not responding on 127.0.0.1:8888" >&2
    compose_cmd logs app --tail 50
    exit 1
  }
  echo "App is responding on 127.0.0.1:8888"
}

stop_unit() {
  local unit="$1"
  if ! command -v systemctl >/dev/null 2>&1; then
    echo "systemctl is not available; skipping $unit."
    return
  fi
  sudo systemctl disable --now "$unit" || true
}

cutover() {
  echo "Stopping legacy systemd service..."
  stop_unit things.service
}

stop_host_mysql() {
  echo "Stopping host MySQL..."
  stop_unit mysql.service
}

install_docker
dump_host_mysql
create_volume
start_db
import_dump
cutover
start_app
verify_app
stop_host_mysql

date -u +%Y-%m-%dT%H:%M:%SZ > "$SENTINEL"

echo "Migration finished. Site should be live via nginx at https://miloanderson.org"
echo "Next steps:"
echo "  - Use scripts/miloserve-deploy.sh for all future deploys."
echo "  - Update ~/scripts/backup.sh to dump through the db container."
echo "  - Once the site has run cleanly for a few days, delete tasks_dump.sql."
