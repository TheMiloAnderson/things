#!/usr/bin/env bash
set -euo pipefail

# Deploy the things app with Docker on miloserve.
# Run from the repo root on the server: bash scripts/miloserve-deploy.sh
#
# Requires sudo for: Docker install (first run only), stopping systemd services.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ ! -f .env ]]; then
  echo "Missing .env — copy .env.example and fill in values first." >&2
  exit 1
fi

# shellcheck disable=SC1091
source .env

if [[ -z "${MYSQL_ROOT_PASSWORD:-}" ]]; then
  echo "MYSQL_ROOT_PASSWORD missing from .env" >&2
  exit 1
fi

docker_cmd() {
  if docker info >/dev/null 2>&1; then
    docker "$@"
  else
    sudo docker "$@"
  fi
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

dump_host_mysql() {
  local dump_file="${ROOT}/tasks_dump.sql"
  if [[ -f "$dump_file" ]]; then
    echo "Using existing dump: $dump_file"
    return
  fi

  echo "Dumping host MySQL tasks database..."
  mysqldump --single-transaction --routines --triggers \
    -u"$DBUSER" -p"$DBPASS" tasks > "$dump_file"
  echo "Dump saved to $dump_file"
}

start_db() {
  echo "Starting MySQL container..."
  compose_cmd up -d db

  echo "Waiting for MySQL container to become healthy..."
  for _ in $(seq 1 60); do
    if compose_cmd exec -T db mysqladmin ping -h localhost >/dev/null 2>&1; then
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
  if [[ ! -f "$dump_file" ]]; then
    echo "No dump file to import." >&2
    exit 1
  fi

  echo "Importing dump into containerized MySQL..."
  compose_cmd exec -T db mysql -uroot -p"$MYSQL_ROOT_PASSWORD" tasks < "$dump_file"
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

cutover() {
  echo "Stopping legacy systemd service..."
  sudo systemctl disable --now things.service || true
}

stop_host_mysql() {
  echo "Stopping host MySQL..."
  sudo systemctl disable --now mysql.service || true
}

install_docker
dump_host_mysql
start_db
import_dump
cutover
start_app
verify_app
stop_host_mysql

echo "Deployment finished. Site should be live via nginx at https://miloanderson.org"
