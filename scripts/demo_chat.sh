#!/usr/bin/env bash
# Levanta el devserver en memoria, ejecuta los escenarios del chatbot
# y valida las respuestas automáticamente. No requiere curl manual.
#
# Uso: ./scripts/demo_chat.sh        (o: make demo)
# Variable opcional: PORT=8080

set -u

PORT="${PORT:-8080}"
BASE_URL="http://localhost:${PORT}"
SERVER_PID=""
PASS=0
FAIL=0

cleanup() {
  if [[ -n "$SERVER_PID" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null
    wait "$SERVER_PID" 2>/dev/null
  fi
}
trap cleanup EXIT

echo "▶ Iniciando devserver en :${PORT}..."
go run ./cmd/devserver >/tmp/trotamundos_devserver.log 2>&1 &
SERVER_PID=$!

# Esperar hasta 15 s a que el servidor responda
for _ in $(seq 1 30); do
  if curl -s -o /dev/null -X POST "${BASE_URL}/api/chat" \
      -H "Content-Type: application/json" -d '{"message":"ping","lang":"es"}'; then
    break
  fi
  sleep 0.5
done

if ! kill -0 "$SERVER_PID" 2>/dev/null; then
  echo "✗ El servidor no pudo iniciar. Log:"
  cat /tmp/trotamundos_devserver.log
  exit 1
fi

# run_case "<nombre>" "<json>" "<regex esperado en la respuesta>"
run_case() {
  local name="$1" payload="$2" expected="$3"
  local body status

  body=$(curl -s -w '\n%{http_code}' -X POST "${BASE_URL}/api/chat" \
    -H "Content-Type: application/json" -d "$payload")
  status=$(tail -n1 <<<"$body")
  body=$(sed '$d' <<<"$body")

  echo
  echo "── ${name}"
  echo "   → ${payload}"
  echo "   ← [${status}] ${body}"

  if [[ "$status" == "200" ]] && grep -Eq "$expected" <<<"$body"; then
    echo "   ✓ OK"
    PASS=$((PASS + 1))
  else
    echo "   ✗ FALLÓ (se esperaba status 200 y match con: ${expected})"
    FAIL=$((FAIL + 1))
  fi
}

run_case "Saludo / consulta general (i18n es)" \
  '{"message": "Hola, necesito asistencia", "lang": "es"}' \
  '"escalated"[[:space:]]*:[[:space:]]*false'

run_case "Estado de viaje (dominio)" \
  '{"message": "estado de mi viaje", "lang": "es"}' \
  '"escalated"[[:space:]]*:[[:space:]]*false'

run_case "Caso complejo: derivación a la agencia" \
  '{"message": "Tengo un problema grave y quiero cancelar mi pasaje urgente", "lang": "es"}' \
  '"escalated"[[:space:]]*:[[:space:]]*true'

run_case "Soporte i18n (en)" \
  '{"message": "Hello, I need help", "lang": "en"}' \
  '"escalated"[[:space:]]*:[[:space:]]*false'

echo
echo "════════════════════════════════"
echo "Resultado: ${PASS} OK, ${FAIL} fallidos"
echo "════════════════════════════════"

[[ "$FAIL" -eq 0 ]]