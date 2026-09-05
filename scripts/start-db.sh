#!/usr/bin/env bash
set -euo pipefail

if command -v pg_ctlcluster >/dev/null 2>&1; then
  sudo pg_ctlcluster 16 main start || true
else
  docker compose up -d db
fi

sudo -u postgres psql -tc "SELECT 1 FROM pg_roles WHERE rolname='hola'" | grep -q 1 || \
  sudo -u postgres psql -c "CREATE USER hola WITH PASSWORD 'hola';"
sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname='hola'" | grep -q 1 || \
  sudo -u postgres psql -c "CREATE DATABASE hola OWNER hola;"
sudo -u postgres psql -d hola -c "GRANT ALL ON SCHEMA public TO hola;" >/dev/null

echo "postgres listo: postgres://hola:hola@localhost:5432/hola"
