#!/bin/sh
# Build do frontend + binário Wails e execução (adequado ao Docker: evita wails dev + Vite, onde o bridge nem sempre injeta no WebKit).
set -e
cd /app
echo "[DualBench] npm run build (frontend)…"
(cd frontend && npm run build)
echo "[DualBench] wails build…"
wails build -tags webkit2_41 -o build/bin/DualBench
echo "[DualBench] iniciando janela…"
exec build/bin/DualBench
