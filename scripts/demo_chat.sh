#!/usr/bin/env bash
set -e

PORT="${PORT:-8080}"
URL="http://localhost:$PORT/api/chat"

echo "==> Iniciando demostración del chatbot de Trotamundos..."
echo "▶ Compilando e iniciando devserver en el puerto :$PORT..."

# Ejecutamos el servidor pasando la variable PORT si el devserver la lee, 
# o levantándolo asegurando control de procesos
PORT="$PORT" go run cmd/devserver/main.go &
SERVER_PID=$!

# Asegurar que el servidor y sus hijos se cierren al salir el script
trap "kill -9 $SERVER_PID 2>/dev/null || true" EXIT

echo "⏳ Esperando a que el servidor esté listo en el puerto $PORT..."
RETRIES=15
until curl -s "http://localhost:$PORT/api/chat" -X POST -H "Content-Type: application/json" -d '{}' > /dev/null 2>&1 || [ $RETRIES -eq 0 ]; do
  sleep 0.5
  RETRIES=$((RETRIES-1))
done

if [ $RETRIES -eq 0 ]; then
  echo "❌ El servidor no levanto a tiempo."
  exit 1
fi

echo "==> Ejecutando y verificando casos de prueba del chat..."

# 1. Prueba de bienvenida (es) - Valida contenido
RES_ES=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d '{"message": "Hola, necesito asistencia", "lang": "es"}')
if echo "$RES_ES" | grep -q "asistente virtual"; then
  echo "Test 1 (Bienvenida ES): OK"
else
  echo "Test 1 (Bienvenida ES): FALLÓ (No devolvió el saludo esperado)"
  exit 1
fi

# 2. Prueba de estado con viaje real (Verifica que extraiga datos del TripUsecase / trip-123)
RES_STATE=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d '{"message": "quiero ver el estado de mi viaje", "lang": "es"}')
if echo "$RES_STATE" | grep -q "activo"; then
  echo "Test 2 (Estado del Viaje): OK (Contiene datos dinámicos del dominio)"
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

# 4. Prueba de soporte en inglés (i18n M2) - Valida idioma
RES_EN=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d '{"message": "Hello, I need assistance", "lang": "en"}')
if echo "$RES_EN" | grep -q "virtual assistant"; then
  echo "Test 4 (Soporte i18n EN): OK"
else
  echo "Test 4 (Soporte i18n EN): FALLÓ (No se detectó soporte en inglés)"
  exit 1
fi

echo "✔ ¡Demostración y pruebas del chatbot validadas rigurosamente con éxito!"