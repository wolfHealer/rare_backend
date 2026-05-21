#!/usr/bin/env bash
# 将 Go 风格 MYSQL_DSN 转为 golang-migrate 的 mysql:// URL
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

if [[ -z "${MYSQL_DSN:-}" ]]; then
  echo "MYSQL_DSN is not set (copy .env.example to .env)" >&2
  exit 1
fi

base="${MYSQL_DSN%%\?*}"
echo "mysql://${base}?multiStatements=true"
