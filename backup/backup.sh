#!/usr/bin/env bash
# Dump the tasks database from the db container and upload it to Cloudflare R2.
# Started by supercronic from /etc/backup.crontab. Run it by hand with:
#   docker compose exec backup /usr/local/bin/backup.sh
set -euo pipefail

: "${DB_HOST:=db}"
: "${DB_NAME:=tasks}"
: "${R2_BUCKET:=miloserve-backups}"
: "${R2_BASE:=backups}"

if [[ -z "${BACKUP_DBUSER:-}" || -z "${BACKUP_DBPASS:-}" || -z "${R2_ENDPOINT:-}" \
   || -z "${R2_ACCESS_KEY_ID:-}" || -z "${R2_SECRET_ACCESS_KEY:-}" ]]; then
  echo "[backup] skipped: BACKUP_DBUSER, BACKUP_DBPASS, and the R2 variables are not all set" >&2
  exit 0
fi

export RCLONE_CONFIG_R2_TYPE=s3
export RCLONE_CONFIG_R2_PROVIDER=Cloudflare
export RCLONE_CONFIG_R2_REGION=auto
export RCLONE_CONFIG_R2_ENDPOINT="$R2_ENDPOINT"
export RCLONE_CONFIG_R2_ACCESS_KEY_ID="$R2_ACCESS_KEY_ID"
export RCLONE_CONFIG_R2_SECRET_ACCESS_KEY="$R2_SECRET_ACCESS_KEY"
export RCLONE_CONFIG_R2_NO_CHECK_BUCKET=true

exec 9>/tmp/things-backup.lock
flock -n 9 || { echo "[backup] another run in progress, exiting" >&2; exit 0; }

DATE="$(date -u +%Y-%m-%d)"
BACKUP_DIR="$(mktemp -d)"
trap 'rm -rf "$BACKUP_DIR"' EXIT

echo "[backup] starting $DATE"
# --single-transaction runs FLUSH TABLES. The backup user is granted
# FLUSH_TABLES by scripts/create-backup-user.sh.
if ! MYSQL_PWD="$BACKUP_DBPASS" mysqldump \
  --host="$DB_HOST" --protocol=TCP --user="$BACKUP_DBUSER" \
  --single-transaction --no-tablespaces --routines --events \
  "$DB_NAME" > "$BACKUP_DIR/tasks-${DATE}.sql"; then
  echo "[backup] mysqldump failed" >&2
  exit 1
fi
if ! grep -q '^-- Dump completed' "$BACKUP_DIR/tasks-${DATE}.sql"; then
  echo "[backup] dump did not finish; nothing was uploaded" >&2
  exit 1
fi

DEST="r2:${R2_BUCKET}/${R2_BASE}/${DATE}/"
REMOTE_BASE="r2:${R2_BUCKET}/${R2_BASE}/"
rclone copy "$BACKUP_DIR" "$DEST" --transfers=2 --checkers=4
rclone delete "$REMOTE_BASE" --min-age 365d
rclone rmdirs "$REMOTE_BASE" --leave-root || true

echo "[backup] finished $DATE -> $DEST"
