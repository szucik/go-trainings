#!/usr/bin/env bash
# Generuje zestaw certyfikatów do testowania mTLS:
#   certs/ca.{key,crt}        — lokalne CA (podpisuje ceryfikaty klientów)
#   certs/server.{key,crt}    — certyfikat serwera (nginx)
#   certs/ca-bundle.pem       — CA bundle (kopia ca.crt) do ssl_client_certificate
#   certs/client.{key,crt}    — certyfikat klienta podpisany przez CA
#   certs/bad-client.{key,crt}— certyfikat klienta podpisany przez inne CA (odrzucony)

set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)/certs"
mkdir -p "$DIR"

echo "=== Generowanie CA ==="
openssl genrsa -out "$DIR/ca.key" 4096 2>/dev/null
openssl req -new -x509 -key "$DIR/ca.key" -out "$DIR/ca.crt" -days 3650 \
  -subj "/CN=Test CA/O=Test Org/C=PL" 2>/dev/null

echo "=== Generowanie certyfikatu serwera ==="
openssl genrsa -out "$DIR/server.key" 2048 2>/dev/null
openssl req -new -key "$DIR/server.key" \
  -out "$DIR/server.csr" \
  -subj "/CN=localhost/O=Test Server/C=PL" 2>/dev/null
openssl x509 -req -in "$DIR/server.csr" -CA "$DIR/ca.crt" -CAkey "$DIR/ca.key" \
  -CAcreateserial -out "$DIR/server.crt" -days 365 \
  -extfile <(printf "subjectAltName=IP:127.0.0.1,DNS:localhost") 2>/dev/null

echo "=== Generowanie CA bundle (ssl_client_certificate) ==="
cp "$DIR/ca.crt" "$DIR/ca-bundle.pem"

echo "=== Generowanie certyfikatu klienta (prawidłowy) ==="
openssl genrsa -out "$DIR/client.key" 2048 2>/dev/null
openssl req -new -key "$DIR/client.key" \
  -out "$DIR/client.csr" \
  -subj "/CN=test-client/O=Test Client/C=PL" 2>/dev/null
openssl x509 -req -in "$DIR/client.csr" -CA "$DIR/ca.crt" -CAkey "$DIR/ca.key" \
  -CAcreateserial -out "$DIR/client.crt" -days 365 2>/dev/null

echo "=== Generowanie certyfikatu klienta (niezaufany — inne CA) ==="
openssl genrsa -out "$DIR/bad-ca.key" 2048 2>/dev/null
openssl req -new -x509 -key "$DIR/bad-ca.key" -out "$DIR/bad-ca.crt" -days 365 \
  -subj "/CN=Bad CA/O=Untrusted/C=PL" 2>/dev/null
openssl genrsa -out "$DIR/bad-client.key" 2048 2>/dev/null
openssl req -new -key "$DIR/bad-client.key" \
  -out "$DIR/bad-client.csr" \
  -subj "/CN=bad-client/O=Untrusted/C=PL" 2>/dev/null
openssl x509 -req -in "$DIR/bad-client.csr" \
  -CA "$DIR/bad-ca.crt" -CAkey "$DIR/bad-ca.key" \
  -CAcreateserial -out "$DIR/bad-client.crt" -days 365 2>/dev/null

rm -f "$DIR"/*.csr "$DIR"/*.srl

echo ""
echo "Certyfikaty zapisane w $DIR/"
ls -1 "$DIR/"
