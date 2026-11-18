#!/usr/bin/env bash
set -e

PORT="${PORT:-8080}"
URL="http://localhost:$PORT/api/chat"
BINARY_PATH="/tmp/devserver_test_bin"

echo "==> Iniciando demostración del chatbot de Trotamundos..."

# Asegurar que el puerto esté libre antes de empezar
if fuser -s "$PORT/tcp" 2>/dev/null; then
  echo "⚠️ El puerto $PORT está ocupado. Liberándolo..."
  fuser -k "$PORT/tcp" 2>/dev/null || true
  sleep 1
fi

echo "▶ Compilando devserver..."
go build -o "$BINARY_PATH" ./cmd/devserver/main.go

echo "▶ Iniciando binario en el puerto :$PORT..."
PORT="$PORT" "$BINARY_PATH" &
SERVER_PID=$!

# Asegurar que el binario hijo y el proceso mueran al salir
trap "kill -9 $SERVER_PID 2>/dev/null || true; rm -f $BINARY_PATH" EXIT

echo "⏳ Esperando a que el servidor esté listo..."
RETRIES=15
until curl -s "$URL" -X POST -H "Content-Type: application/json" -d '{}' > /dev/null 2>&1 || [ $RETRIES -eq 0 ]; do
  sleep 0.5
  RETRIES=$((RETRIES-1))
done

if [ $RETRIES -eq 0 ]; then
  echo "❌ El servidor no levantó a tiempo."
  exit 1
fi

echo "==> Ejecutando y verificando casos de prueba del chat..."

# 1. Prueba de bienvenida (es)
RES_ES=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d '{"message": "Hola, necesito asistencia", "lang": "es"}')
if echo "$RES_ES" | grep -q "asistente virtual"; then
  echo "Test 1 (Bienvenida ES): OK"
else
  echo "Test 1 (Bienvenida ES): FALLÓ"
  exit 1
fi

# 2. Prueba de estado con viaje real (Verifica extracción dinámica de "Buenos Aires")
RES_STATE=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d '{"message": "quiero ver el estado de mi viaje", "lang": "es"}')
if echo "$RES_STATE" | grep -q "Buenos Aires"; then
  echo "Test 2 (Estado del Viaje con destino real): OK"
else
  echo "Test 2 (Estado del Viaje): FALLÓ"
  exit 1
fi

# 3. Prueba de escalamiento / derivación a agencia (Verifica escalated:true)
RES_ESC=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d '{"message": "Tengo un problema urgente y quiero cancelar mi pasaje", "lang": "es"}')
if echo "$RES_ESC" | grep -q '"escalated":true'; then
  echo "Test 3 (Escalamiento): OK"
else
  echo "Test 3 (Escalamiento): FALLÓ"
  exit 1
fi

# 4. Prueba de soporte bilingüe en inglés (M2) con palabras clave en inglés
RES_EN=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d '{"message": "What is my trip status?", "lang": "en"}')
if echo "$RES_EN" | grep -q "Buenos Aires" && echo "$RES_EN" | grep -q "active"; then
  echo "Test 4 (Soporte i18n EN con keywords en inglés): OK"
else
  echo "Test 4 (Soporte i18n EN): FALLÓ"
  exit 1
fi

echo "✔ ¡Demostración y pruebas del chatbot validadas de forma impecable!"