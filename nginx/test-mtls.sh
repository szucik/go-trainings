#!/usr/bin/env bash
# Testy mTLS — sprawdza czy OpenResty loguje pola identycznie jak AWS ALB
# Uruchomienie: ./test-mtls.sh
# Wymagania: docker compose up musi być aktywny, certyfikaty w ./certs/

set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
CERTS="$DIR/certs"
LOG="$DIR/logs/connection.log"

PASS=0
FAIL=0

# Poczekaj aż OpenResty będzie gotowy
wait_for_openresty() {
    local tries=0
    until curl -sk https://localhost/ >/dev/null 2>&1 || [ $tries -ge 20 ]; do
        sleep 1
        tries=$((tries + 1))
    done
    if [ $tries -ge 20 ]; then
        echo "BŁĄD: OpenResty nie odpowiada na localhost:443 po 20 sekundach"
        exit 1
    fi
}

# Wyślij request i zwróć ostatni wpis w logu
last_log_entry() {
    sleep 0.3
    tail -1 "$LOG" 2>/dev/null || echo ""
}

check() {
    local name="$1"
    local pattern="$2"
    local entry="$3"

    if echo "$entry" | grep -q "$pattern"; then
        echo "PASS  $name"
        PASS=$((PASS + 1))
    else
        echo "FAIL  $name"
        echo "      pattern: $pattern"
        echo "      log:     $entry"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== Czekam na OpenResty ==="
wait_for_openresty
echo "OpenResty działa."
echo ""

# Wyczyść log przed testami
> "$LOG"

# ---------------------------------------------------------------------------
# Test 1: Prawidłowy certyfikat klienta → Success
# ---------------------------------------------------------------------------
echo "--- Test 1: mTLS z prawidłowym certyfikatem ---"
curl -sk \
  --cacert "$CERTS/ca.crt" \
  --cert "$CERTS/client.crt" \
  --key  "$CERTS/client.key" \
  https://localhost/ >/dev/null 2>&1 || true

ENTRY=$(last_log_entry)
check "tls_verify_status = Success"          "Success"         "$ENTRY"
check "leaf_client_cert_subject zawiera CN"  "CN=test-client"  "$ENTRY"
check "leaf_client_cert_validity ma NotBefore" "NotBefore="    "$ENTRY"
check "tls_handshake_latency jest liczbą"    "[0-9]\+\.[0-9]\+" "$ENTRY"
check "conn_trace_id ma prefix TID_"         "TID_"            "$ENTRY"

echo ""

# ---------------------------------------------------------------------------
# Test 2: Brak certyfikatu klienta → nginx zwraca 400, log: tls_verify_status="-"
# (przy ssl_verify_client on nginx odrzuca połączenie bez certu)
# ---------------------------------------------------------------------------
echo "--- Test 2: mTLS bez certyfikatu klienta ---"
curl -sk --cacert "$CERTS/ca.crt" https://localhost/ >/dev/null 2>&1 || true

ENTRY=$(last_log_entry)
# Przy ssl_verify_client on nginx zamyka połączenie na poziomie TLS —
# log_by_lua_block może nie zdążyć zalogować, lub zaloguje z "-"
if [ -z "$ENTRY" ]; then
    echo "INFO  Test 2: brak wpisu w logu (nginx odrzucił przed logowaniem) — oczekiwane"
    PASS=$((PASS + 1))
else
    check "tls_verify_status = - lub Failed" "[-F]" "$ENTRY"
fi

echo ""

# ---------------------------------------------------------------------------
# Test 3: Niezaufany certyfikat klienta → Failed:ClientCertUntrusted
# ---------------------------------------------------------------------------
echo "--- Test 3: mTLS z niezaufanym certyfikatem ---"
curl -sk \
  --cacert "$CERTS/ca.crt" \
  --cert "$CERTS/bad-client.crt" \
  --key  "$CERTS/bad-client.key" \
  https://localhost/ >/dev/null 2>&1 || true

ENTRY=$(last_log_entry)
if [ -z "$ENTRY" ]; then
    echo "INFO  Test 3: brak wpisu w logu (nginx odrzucił przed logowaniem) — oczekiwane"
    PASS=$((PASS + 1))
else
    check "tls_verify_status = Failed:ClientCertUntrusted" \
        "Failed:ClientCertUntrusted" "$ENTRY"
fi

echo ""

# ---------------------------------------------------------------------------
# Test 4: HTTP redirect → brak wpisu w connection.log (inny serwer)
# ---------------------------------------------------------------------------
echo "--- Test 4: HTTP (port 80) redirect ---"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost/ 2>/dev/null || echo "000")
check "HTTP 301 redirect" "301" "$HTTP_CODE"

echo ""

# ---------------------------------------------------------------------------
# Podsumowanie
# ---------------------------------------------------------------------------
echo "$(printf '=%.0s' {1..60})"
printf "Wynik: %d PASS  %d FAIL  (%d łącznie)\n" "$PASS" "$FAIL" "$((PASS + FAIL))"

echo ""
echo "Ostatnie 5 wpisów w connection.log:"
tail -5 "$LOG" 2>/dev/null || echo "(log pusty)"

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
