#!/bin/sh
set -e

cd /app

[ -z "$MIGRATE_TO" ] && MIGRATE_TO="last"

if [ -z "$NO_MIGRATE" ]; then
    printf 'Running migration to %s\n' "$MIGRATE_TO"
    tern migrate --host "$DATABASE_HOST" --user "$DATABASE_USER" \
    --password "$DATABASE_PASSWORD" --database "$DATABASE_NAME" \
    -m migrations -d "$MIGRATE_TO" >/dev/null
else
    printf 'Skipping migration\n'
fi

exec "$@"