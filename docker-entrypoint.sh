#!/bin/sh
set -e
cd /app
if [ ! -d frontend/node_modules ]; then
  echo "Instalando dependências do frontend..."
  if [ -f frontend/package-lock.json ]; then
    (cd frontend && npm ci)
  else
    (cd frontend && npm install)
  fi
fi
exec "$@"
