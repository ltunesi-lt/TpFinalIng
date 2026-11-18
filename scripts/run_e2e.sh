#!/usr/bin/env bash
set -e

echo "==> Limpiando puerto 8080..."
lsof -t -i:8080 | xargs kill -9 2>/dev/null || true

echo "==> Compilando Dev Server..."
go build -o devserver ./cmd/devserver/main.go

echo "==> Iniciando Dev Server en segundo plano..."
./devserver > devserver.log 2>&1 &
SERVER_PID=$!

cleanup() {
    echo "==> Limpiando recursos..."
    kill $SERVER_PID 2>/dev/null || true
    rm -f devserver devserver.log
}
trap cleanup EXIT INT TERM

echo "==> Esperando que el servidor responda..."
RETRIES=30
until curl -s http://127.0.0.1:8080/api/trips/trip-123/summary >/dev/null 2>&1 || [ $RETRIES -eq 0 ]; do
    sleep 0.3
    RETRIES=$((RETRIES-1))
done

if [ $RETRIES -eq 0 ]; then
    echo "❌ El servidor no respondió a tiempo. Logs del servidor:"
    cat devserver.log
    exit 1
fi

echo "==> 1. Ejecutando script de pruebas de la API tradicional (test_api.sh)..."
bash scripts/test_api.sh

echo "==> 2. Ejecutando demostración y validación del Chatbot (demo_chat.sh)..."
bash scripts/demo_chat.sh

echo "==> ¡Todas las pruebas E2E finalizaron con éxito!"