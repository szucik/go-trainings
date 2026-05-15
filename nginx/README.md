# AWS ALB Connection Log — OpenResty Drop-in

Implementacja AWS ALB Connection Log w OpenResty/nginx. Odwzorowuje format i semantykę logów opisanych w [dokumentacji AWS](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-connection-logs.html), umożliwiając lokalną replikację zachowania ALB — łącznie z mTLS i metadanymi certyfikatu klienta.

---

## Wymagania

- **OpenResty** >= 1.21 (zawiera LuaJIT, ngx_lua, ngx_stream_lua)
- **OpenSSL** >= 1.1 (do mTLS)
- Certyfikaty serwera i opcjonalnie CA bundle do weryfikacji klientów
- Lua 5.1 (tylko do uruchomienia testów lokalnie bez OpenResty)

---

## Instalacja

### 1. Skopiuj pliki

```bash
sudo cp connection_log.lua /etc/nginx/lua/connection_log.lua
sudo cp nginx.conf         /etc/nginx/nginx.conf
```

### 2. Wygeneruj certyfikaty (dev/test)

```bash
# Certyfikat serwera
openssl req -x509 -newkey rsa:4096 -keyout /etc/nginx/certs/server.key \
  -out /etc/nginx/certs/server.crt -days 365 -nodes \
  -subj "/CN=localhost"

# CA bundle do mTLS (używane do weryfikacji certów klientów)
# W produkcji użyj własnego CA
cp /etc/nginx/certs/server.crt /etc/nginx/certs/ca-bundle.pem
```

### 3. Utwórz katalogi logów

```bash
sudo mkdir -p /var/log/nginx
sudo touch /var/log/nginx/connection.log /var/log/nginx/error.log
```

### 4. Uruchom OpenResty

```bash
sudo openresty -t          # test konfiguracji
sudo openresty             # start
sudo openresty -s reload   # przeładowanie po zmianach
```

---

## Przykładowe wpisy w logu

### Format space-separated (domyślny, identyczny z AWS ALB)

```
2024-03-15T12:34:56+00:00 192.168.1.100 54321 443 TLSv1.3 TLS_AES_256_GCM_SHA384 0.043 "CN=client.example.com,O=Example Corp,C=PL" NotBefore=2023-09-21T22:43:21Z;NotAfter=2026-06-17T22:43:21Z A1B2C3D4E5F6 Success TID_42
```

### Format JSON (`alb_connection_log_json`)

```json
{
  "timestamp": "2024-03-15T12:34:56+00:00",
  "client_ip": "192.168.1.100",
  "client_port": "54321",
  "listener_port": "443",
  "tls_protocol": "TLSv1.3",
  "tls_cipher": "TLS_AES_256_GCM_SHA384",
  "tls_handshake_latency": "0.043",
  "leaf_client_cert_subject": "CN=client.example.com,O=Example Corp,C=PL",
  "leaf_client_cert_validity": "NotBefore=2023-09-21T22:43:21Z;NotAfter=2026-06-17T22:43:21Z",
  "leaf_client_cert_serial": "A1B2C3D4E5F6",
  "tls_verify_status": "Success",
  "conn_trace_id": "TID_42"
}
```

---

## Porównanie pól z AWS ALB Connection Log

| # | Pole AWS ALB                   | Źródło OpenResty              | Uwagi                              |
|---|--------------------------------|-------------------------------|-------------------------------------|
| 1 | `timestamp`                    | `$time_iso8601`               | identyczny format                   |
| 2 | `client_ip`                    | `$remote_addr`                |                                     |
| 3 | `client_port`                  | `$remote_port`                |                                     |
| 4 | `listener_port`                | `$server_port`                |                                     |
| 5 | `tls_protocol`                 | `$ssl_protocol`               |                                     |
| 6 | `tls_cipher`                   | `$ssl_cipher`                 |                                     |
| 7 | `tls_handshake_latency`        | `ngx.now()` delta             | przybliżenie — patrz Ograniczenia  |
| 8 | `leaf_client_cert_subject`     | `$ssl_client_s_dn`            |                                     |
| 9 | `leaf_client_cert_validity`    | `$ssl_client_v_start/v_end`   | konwersja formatu daty              |
|10 | `leaf_client_cert_serial`      | `$ssl_client_serial`          |                                     |
|11 | `tls_verify_status`            | `$ssl_client_verify` (mapped) | mapowanie kodów błędów              |
|12 | `conn_trace_id`                | `$connection` z prefiksem `TID_` | nie jest UUID jak w AWS          |
| — | `tls_keyexchange`              | niedostępne w nginx           | pole pomijane                       |

---

## Znane ograniczenia

1. **`tls_handshake_latency` jest przybliżeniem** — `ssl_certificate_by_lua_block` odpala się w środku handshake (po weryfikacji certyfikatu serwera), nie na jego początku. Rzeczywista latency jest nieznacznie wyższa. Różnica typowo kilka ms, akceptowalna do celów diagnostycznych.

2. **`tls_keyexchange` niedostępne** — nginx nie eksponuje metody wymiany klucza (np. `ECDHE`) jako osobnej zmiennej. Pole jest pomijane w logu.

3. **`conn_trace_id` nie jest globalnie unikalny** — używamy licznika `$connection` z prefiksem `TID_` zamiast UUID jak w ALB. Wartość jest unikalna per worker process, nie globalnie przez cały czas życia loadbalancera.

4. **mTLS wymagane domyślnie** — konfiguracja ma `ssl_verify_client on`. Aby zezwolić na połączenia bez certyfikatu klienta zmień na `ssl_verify_client optional`.

---

## Testowanie

### Testy jednostkowe (bez OpenResty)

```bash
cd nginx/
lua test_connection_log.lua
```

Testy sprawdzają funkcje prywatne modułu bez uruchamiania serwera:
- `_map_verify_status` — wszystkie kody błędów mTLS
- `_nginx_date_to_iso8601` — parsowanie daty nginx do ISO 8601
- `_format_validity` — formatowanie okresu ważności certyfikatu

### Test połączenia z OpenResty

```bash
# Połączenie TLS bez certyfikatu klienta (oczekiwany błąd przy ssl_verify_client on)
curl -k https://localhost/

# Połączenie mTLS z certyfikatem klienta
curl -k --cert client.crt --key client.key https://localhost/

# Podgląd logów na żywo
tail -f /var/log/nginx/connection.log
```
