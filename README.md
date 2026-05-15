# OpenResty AWS ALB-style Connection Log

Moduł Lua dla OpenResty/nginx implementujący format i semantykę AWS ALB Connection Log.
Służy do zapisu pól związanych z połączeniem TLS (w tym przybliżonego czasu handshake) w formacie zgodnym z ALB.

Wymagania
- OpenResty (nginx) >= 1.21
- Brak zewnętrznych zależności Lua — używa wbudowanych bibliotek `ngx` i `cjson` (opcjonalnie)

Instalacja (krok po kroku)
1. Skopiuj moduł Lua do katalogu `lua/connection_log.lua` w Twojej konfiguracji nginx.
2. Zaktualizuj `nginx.conf`, dodając `lua_shared_dict tls_timing 10m;` i `lua_package_path` tak, aby zawierał katalog z modułem.
3. Zadeklaruj zmienne `set $conn_* "-";` w `server {}` przed użyciem `log_format`.
4. Dodaj `ssl_certificate_by_lua_block` oraz `log_by_lua_block` wywołujące funkcje modułu.

Przykładowe lokalne pliki w repozytorium
- `nginx/connection_log.lua` — moduł Lua (implementacja)
- `nginx/nginx.conf` — przykładowa konfiguracja

Przykładowy wpis logu (space-separated):
2023-09-21T22:43:21Z 203.0.113.5 52344 443 TLSv1.3 TLS_AES_128_GCM_SHA256 0.043 
"CN=client,O=Org" NotBefore=2023-09-21T22:43:21Z;NotAfter=2026-06-17T22:43:21Z 123456 Success TID_42

Przykładowy wpis logu (JSON):
{
  "timestamp": "2023-09-21T22:43:21Z",
  "client_ip": "203.0.113.5",
  "client_port": 52344,
  "listener_port": 443,
  "tls_protocol": "TLSv1.3",
  "tls_cipher": "TLS_AES_128_GCM_SHA256",
  "tls_handshake_latency": "0.043",
  "leaf_client_cert_subject": "CN=client,O=Org",
  "leaf_client_cert_validity": "NotBefore=2023-09-21T22:43:21Z;NotAfter=2026-06-17T22:43:21Z",
  "leaf_client_cert_serial": "123456",
  "tls_verify_status": "Success",
  "conn_trace_id": "TID_42"
}

Porównanie pól z AWS ALB
- `tls_handshake_latency` — przybliżone, mierzone jako delta między wywołaniem w `ssl_certificate_by_lua_block` i `log_by_lua_block`.
- `conn_trace_id` — używa `$connection` z prefiksem `TID_`, więc unikalność ograniczona do worker-a.

Znane ograniczenia
1. `tls_handshake_latency` nie jest absolutnie precyzyjne — `ssl_certificate_by_lua_block` uruchamia się w trakcie handshake.
2. Brak dostępu do niektórych pól AWS (np. `tls_keyexchange`) w nginx/OpenResty.
3. `conn_trace_id` nie jest UUID — jest lokalnym identyfikatorem połączenia per worker.

Testowanie
1. W repozytorium znajduje się `test_connection_log.lua` — prosty zestaw testów jednostkowych (nie wymaga OpenResty).
2. Uruchom testy poleceniem (w katalogu projektu):

```sh
lua test_connection_log.lua
```

Testy zwracają kod wyjścia `0` gdy wszystkie testy przejdą, `1` gdy są niepowodzenia.

Jeśli chcesz, mogę dodać skrypt uruchamiający testy z `busted` lub CI.
