#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if ! command -v newman >/dev/null 2>&1; then
  echo "Newman no está instalado. Instalándolo..."
  npm install -g newman
fi

newman run "real-state.happy_path.postman_collection.json" \
  -e "real-state.postman_environment.json" \
  --reporters cli,json \
  --reporter-json-export "newman-report.json"

echo "Reporte generado: $ROOT_DIR/newman-report.json"
