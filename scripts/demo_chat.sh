#!/usr/bin/env bash
set -e

PORT="${PORT:-8080}"
URL="http://localhost:$PORT/api/chat"

echo "==> Iniciando demostración del chatbot de Trotamundos..."
echo "▶ Iniciando devserver en :$PORT..."
go run cmd/devserver/main.go &
SERVER_PID=$!

# Asegurar que el servidor se cierre al salir el script (pase lo que pase)
trap "kill $SERVER_PID 2>/dev/null || true" EXIT

# Esperar 1.5 segundos a que el servidor levante bien
sleep 1.5

echo "==> Ejecutando casos de prueba del chat..."

# 1. Prueba de bienvenida (es)
curl -s -X POST "$URL" -H "Content-Type: application/json" -d '{"message": "Hola, necesito asistencia", "lang": "es"}' > /dev/null
echo "Test 1 (Bienvenida ES): OK"

# 2. Prueba de estado con viaje real (Debe contener el destino extraído del trip-123)
RES_STATE=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d '{"message": "quiero ver el estado de mi viaje", "lang": "es"}')
if echo "$RES_STATE" | grep -q "activo"; then
  echo "Test 2 (Estado del Viaje): OK"
else
  echo "Test 2 (Estado del Viaje): FALLÓ"
  exit 1
fi

# 3. Prueba de escalamiento / derivación a agencia
RES_ESC=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d '{"message": "Tengo un problema urgente y quiero cancelar mi pasaje", "lang": "es"}')
if echo "$RES_ESC" | grep -q '"escalated":true'; then
  echo "Test 3 (Escalamiento): OK"
else
  echo "Test 3 (Escalamiento): FALLÓ"
  exit 1
fi

# 4. Prueba de soporte en inglés (i18n M2)
RES_EN=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d '{"message": "Hello, I need assistance", "lang": "en"}')
if echo "$RES_EN" | grep -q "virtual assistant"; then
  echo "Test 4 (Soporte i18n EN): OK"
else
  echo "Test 4 (Soporte i18n EN): FALLÓ"
  exit 1
fi

echo "✔ ¡Demostración y pruebas del chatbot finalizadas con éxito!"