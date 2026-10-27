#!/usr/bin/env bash
# Prueba end-to-end de la API. Requiere el servidor corriendo en BASE_URL.
#   Terminal 1: go run ./cmd/devserver
#   Terminal 2: bash scripts/test_api.sh
BASE_URL="${BASE_URL:-http://localhost:8080}"
PASS=0; FAIL=0

# check "descripción" CODIGO_ESPERADO método ruta [cuerpo]
# Deja el cuerpo de la respuesta en $BODY
check() {
  local desc="$1" expected="$2" method="$3" path="$4" data="${5:-}"
  local out
  if [ -n "$data" ]; then
    out=$(curl -s -w $'\n%{http_code}' -X "$method" "$BASE_URL$path" -d "$data")
  else
    out=$(curl -s -w $'\n%{http_code}' -X "$method" "$BASE_URL$path")
  fi
  CODE="${out##*$'\n'}"
  BODY="${out%$'\n'*}"
  if [ "$CODE" = "$expected" ]; then
    echo "  OK   $desc ($CODE)"; PASS=$((PASS+1))
  else
    echo "  FAIL $desc: esperado $expected, obtenido $CODE -> $BODY"; FAIL=$((FAIL+1))
  fi
}

if ! curl -s -o /dev/null "$BASE_URL/api/trips/x/summary"; then
  echo "No se pudo conectar a $BASE_URL. ¿Está corriendo el servidor?"; exit 1
fi

echo "== Viajes"
check "GET resumen existente"        200 GET  /api/trips/trip-123/summary
check "GET resumen inexistente"      404 GET  /api/trips/no-existe/summary
check "PUT resumen"                  204 PUT  /api/trips/trip-nuevo/summary '{"destination":"Mendoza"}'
check "GET resumen recién guardado"  200 GET  /api/trips/trip-nuevo/summary
check "PUT con JSON inválido"        400 PUT  /api/trips/trip-nuevo/summary 'no-es-json'
check "POST sobre /summary (405)"    405 POST /api/trips/trip-123/summary

echo "== Correo (SEC1)"
check "Vincular correo"              204 POST   /api/users/user-1/email '{"access_token":"abc"}'
check "Vincular sin token"           400 POST   /api/users/user-1/email '{"access_token":""}'
check "Revocar correo"               204 DELETE /api/users/user-1/email

echo "== Enlaces compartidos (SEC2)"
check "Crear enlace"                 201 POST /api/trips/trip-123/share '{"duration_hours":24}'
TOKEN=$(echo "$BODY" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
check "Crear enlace con duración 0"  400 POST /api/trips/trip-123/share '{"duration_hours":0}'
check "Acceder con token válido"     200 GET  "/api/shared/$TOKEN"
check "Revocar enlace"               204 DELETE "/api/shared/$TOKEN"
check "Acceder tras revocar"         404 GET  "/api/shared/$TOKEN"
check "Acceder con token falso"      404 GET  /api/shared/token-falso

echo
echo "Resultado: $PASS OK, $FAIL FAIL"
[ "$FAIL" -eq 0 ]
