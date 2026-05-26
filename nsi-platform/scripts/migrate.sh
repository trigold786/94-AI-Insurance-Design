#!/bin/sh
set -e

until pg_isready -h postgres -U postgres; do sleep 1; done

run_migrations() {
  db="$1"
  dir="$2"
  psql -h postgres -U postgres -d "$db" -c "CREATE TABLE IF NOT EXISTS _migrations (file TEXT PRIMARY KEY, applied_at TIMESTAMPTZ DEFAULT NOW())" 2>/dev/null
  for f in "$dir"/*.sql; do
    name=$(basename "$f")
    already=$(psql -h postgres -U postgres -d "$db" -At -c "SELECT 1 FROM _migrations WHERE file = '$(echo "$name" | sed "s/'/''/g")'")
    if [ "$already" = '1' ]; then
      echo "  $name (already applied, skipping)"
      continue
    fi
    echo "  $name"
    psql -h postgres -U postgres -d "$db" -f "$f" 2>&1 | grep -v 'NOTICE'
    psql -h postgres -U postgres -d "$db" -c "INSERT INTO _migrations (file) VALUES ('$(echo "$name" | sed "s/'/''/g")')" 2>/dev/null
  done
}

echo 'Running api-server migrations...'
run_migrations nsi_api /migrations/api

echo 'Running crawler migrations...'
run_migrations nsi_crawler /migrations/crawler

echo 'Seeding default user...'
psql -h postgres -U postgres -d nsi_api -c "INSERT INTO users (user_id) VALUES ('default-user') ON CONFLICT DO NOTHING" 2>/dev/null

echo 'Migrations complete.'