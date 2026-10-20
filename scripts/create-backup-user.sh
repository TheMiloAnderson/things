#!/usr/bin/env bash
# Create or update the read-only backup user inside the running db container.
# Run once from the repo root, and again if BACKUP_DBPASS changes:
#   bash scripts/create-backup-user.sh
#
# The account can read the tasks database and run FLUSH TABLES, which
# mysqldump --single-transaction requires. It cannot insert, update, or delete.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ ! -f .env ]]; then
  echo "Missing .env" >&2
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

sql_quote() {
  local escaped
  escaped="$(printf '%s' "$1" | sed -e 's/\\/\\\\/g' -e "s/'/''/g")"
  printf "'%s'" "$escaped"
}

BACKUP_DBUSER="$(env_val BACKUP_DBUSER)"
BACKUP_DBPASS="$(env_val BACKUP_DBPASS)"

if [[ -z "$BACKUP_DBUSER" || -z "$BACKUP_DBPASS" ]]; then
  echo "BACKUP_DBUSER and BACKUP_DBPASS must be set in .env" >&2
  exit 1
fi
if [[ ! "$BACKUP_DBUSER" =~ ^[A-Za-z0-9_]+$ ]]; then
  echo "BACKUP_DBUSER may contain only letters, digits, and underscores" >&2
  exit 1
fi

USER_SQL="$(sql_quote "$BACKUP_DBUSER")"
PASS_SQL="$(sql_quote "$BACKUP_DBPASS")"

echo "Creating backup user ${BACKUP_DBUSER} in the db container..."
docker compose exec -T db sh -c 'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" exec mysql -uroot' <<SQL
CREATE USER IF NOT EXISTS ${USER_SQL}@'%' IDENTIFIED BY ${PASS_SQL};
ALTER USER ${USER_SQL}@'%' IDENTIFIED BY ${PASS_SQL};
GRANT SELECT, SHOW VIEW, TRIGGER, EVENT ON tasks.* TO ${USER_SQL}@'%';
GRANT FLUSH_TABLES ON *.* TO ${USER_SQL}@'%';
SQL
echo "Backup user ${BACKUP_DBUSER} is ready."
