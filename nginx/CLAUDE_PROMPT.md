# Kontekst projektu

Implementuję AWS ALB Connection Log w OpenResty/nginx jako drop-in replacement.
Celem jest odwzorowanie formatu i semantyki logów z dokumentacji AWS:
https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-connection-logs.html

## Co już zostało zaprojektowane

### Architektura

Dwa pliki:
- `nginx.conf` — konfiguracja OpenResty
- `lua/connection_log.lua` — moduł Lua z logiką

### Przepływ danych

```
TCP connect
    → ssl_certificate_by_lua_block   ← zapisuje ngx.now() do lua_shared_dict
    → TLS handshake
    → HTTP request → response
    → log_by_lua_block               ← oblicza latency, mapuje pola, ustawia $conn_*
    → access_log                     ← log_format odczytuje $conn_* i zapisuje wpis
```

### Pola logu (kolejność identyczna jak AWS ALB)

| Pozycja | Pole AWS ALB                  | Źródło nginx                  | Zmienna nginx          |
|---------|-------------------------------|-------------------------------|------------------------|
| 1       | timestamp                     | $time_iso8601                 | natywna                |
| 2       | client_ip                     | $remote_addr                  | natywna                |
| 3       | client_port                   | $remote_port                  | natywna                |
| 4       | listener_port                 | $server_port                  | natywna                |
| 5       | tls_protocol                  | $ssl_protocol                 | natywna                |
| 6       | tls_cipher                    | $ssl_cipher                   | natywna                |
| 7       | tls_handshake_latency         | ngx.now() delta               | $conn_tls_handshake_latency |
| 8       | leaf_client_cert_subject      | $ssl_client_s_dn              | $conn_leaf_cert_subject |
| 9       | leaf_client_cert_validity     | $ssl_client_v_start/v_end     | $conn_leaf_cert_validity |
| 10      | leaf_client_cert_serial       | $ssl_client_serial            | $conn_leaf_cert_serial  |
| 11      | tls_verify_status             | $ssl_client_verify (mapped)   | $conn_tls_verify_status |
| 12      | conn_trace_id                 | $connection                   | natywna (prefix TID_)  |

### Logika Lua (connection_log.lua)

Moduł eksportuje dwie funkcje publiczne:

**`ssl_certificate_phase()`**
- wywoływana z `ssl_certificate_by_lua_block{}`
- zapisuje `ngx.now()` do `lua_shared_dict tls_timing` z kluczem `"ip:port"`
- TTL wpisu: 30 sekund

**`log_phase()`**
- wywoływana z `log_by_lua_block{}`
- pobiera timestamp z shared dict, oblicza delta → `$conn_tls_handshake_latency`
- parsuje datę nginx (`"Sep 21 22:43:21 2023 GMT"`) do ISO 8601 (`"2023-09-21T22:43:21Z"`)
- formatuje validity: `"NotBefore=...;NotAfter=..."` zgodnie z formatem ALB
- mapuje `$ssl_client_verify` (nginx) → kody ALB:
  - `SUCCESS`                     → `"Success"`
  - `NONE`                        → `"-"`
  - `FAILED:certificate expired`  → `"Failed:ClientCertExpired"`
  - `FAILED:not yet valid`        → `"Failed:ClientCertNotYetValid"`
  - `FAILED:revoked`              → `"Failed:ClientCertRevoked"`
  - `FAILED:unable to get issuer` → `"Failed:ClientCertUntrusted"`
  - `FAILED:chain too long`       → `"Failed:ClientCertMaxChainDepthExceeded"`
  - wszystko inne                 → `"Failed:UnmappedConnectionError"`
- ustawia zmienne `$conn_*` przez `ngx.var[name] = value`

Moduł eksportuje też funkcje prywatne dla testów:
- `_map_verify_status(raw)`
- `_format_validity(v_start, v_end)`
- `_nginx_date_to_iso8601(date_str)`

### Konfiguracja nginx.conf

```
lua_shared_dict tls_timing 10m;
lua_package_path "/etc/nginx/lua/?.lua;;";
```

Dwa formaty logu:
- `alb_connection_log` — space-separated identyczny z AWS
- `alb_connection_log_json` — JSON (łatwiejszy do Athena/Splunk)

W bloku `server {}`:
- `set $conn_* "-";` dla każdej zmiennej (wymagane przez nginx zanim log_format użyje)
- `ssl_verify_client on;` + `ssl_client_certificate /etc/nginx/certs/ca-bundle.pem;`
- nagłówki `X-Amzn-Mtls-*` przekazywane do backendu (jak ALB)

### Znane ograniczenia

1. `tls_handshake_latency` jest przybliżeniem — `ssl_certificate_by_lua_block` odpala
   się w środku handshake, nie na jego początku. Zaniżenie o kilka ms, akceptowalne.

2. `tls_keyexchange` (pole 13 w AWS) — niedostępne w nginx, pomijamy.

3. `conn_trace_id` — używamy `$connection` (liczba całkowita) z prefiksem `TID_`
   zamiast UUID jak w ALB. Unikalność per worker process, nie globalnie.

## Struktura plików na WSL

```
/etc/nginx/
├── nginx.conf
├── lua/
│   └── connection_log.lua
└── certs/
    ├── server.crt
    ├── server.key
    └── ca-bundle.pem          # tylko przy mTLS
```

Log output:
```
/var/log/nginx/connection.log
/var/log/nginx/error.log
```

## Zadania do wykonania przez Claude Code

Utwórz poniższe pliki dokładnie w podanych ścieżkach:

### 1. `/etc/nginx/lua/connection_log.lua`

Pełny moduł Lua zgodnie z powyższą specyfikacją. Wymagania:
- kompatybilny z OpenResty >= 1.21
- używa tylko bibliotek wbudowanych w OpenResty (cjson, ngx.shared, ngx.var)
- obsługa błędów przez pcall wszędzie gdzie możliwy jest wyjątek
- komentarze po polsku
- eksport funkcji prywatnych dla testów jako `_M._nazwa`

### 2. `/etc/nginx/nginx.conf`

Pełna konfiguracja zgodnie z powyższą specyfikacją. Wymagania:
- oba formaty logu: space-separated i JSON
- server 443 z mTLS (`ssl_verify_client on`)
- server 80 z redirect do HTTPS
- wszystkie `set $conn_*` zadeklarowane przed użyciem
- komentarze po polsku wyjaśniające każdą sekcję

### 3. `./test_connection_log.lua`

Testy jednostkowe dla funkcji prywatnych modułu. Wymagania:
- nie wymaga OpenResty — mockuje `ngx` przez prosty stub
- testuje wszystkie ścieżki `map_verify_status`:
  SUCCESS, NONE, każdy FAILED:* kod
- testuje `nginx_date_to_iso8601`:
  poprawna data, nil, pusty string, błędny format
- testuje `format_validity`:
  obie daty poprawne, jedna nil, obie nil
- wynik: `PASS` / `FAIL` per test, summary na końcu

### 4. `./README.md`

Dokumentacja projektu po polsku. Zawiera:
- opis projektu (1 akapit)
- wymagania (OpenResty version, system deps)
- instalacja krok po kroku
- przykładowy wpis w logu (space-separated i JSON)
- porównanie pól z oryginalnym AWS ALB connection log
- sekcja "Znane ograniczenia"
- sekcja "Testowanie" z instrukcją uruchomienia test_connection_log.lua

## Dodatkowe wymagania

- Nie używaj zewnętrznych zależności Lua poza tym co jest w OpenResty
- Nie modyfikuj żadnych plików poza wymienionymi powyżej
- Jeśli coś jest niejasne w specyfikacji, zapytaj przed implementacją
