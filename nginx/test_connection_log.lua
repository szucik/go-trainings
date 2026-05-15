-- =============================================================================
-- test_connection_log.lua
-- Testy jednostkowe dla modułu connection_log.lua
-- Uruchomienie: lua test_connection_log.lua  (z katalogu nginx/)
-- Nie wymaga OpenResty — używa minimalnego stuba ngx
-- =============================================================================

-- Stub ngx: tylko to co używa moduł w funkcjach prywatnych
ngx = {
    WARN  = 4,
    ERR   = 3,
    log   = function() end,
    now   = function() return os.time() end,
    var   = {},
    shared = {},
}

local m = require("connection_log")

local pass_count = 0
local fail_count = 0

local function test(name, got, expected)
    if got == expected then
        print("PASS  " .. name)
        pass_count = pass_count + 1
    else
        print("FAIL  " .. name)
        print("      got:      " .. tostring(got))
        print("      expected: " .. tostring(expected))
        fail_count = fail_count + 1
    end
end

-- =============================================================================
-- _map_verify_status
-- =============================================================================

test("map_verify_status: SUCCESS",
    m._map_verify_status("SUCCESS"),
    "Success")

test("map_verify_status: NONE (klient bez certu)",
    m._map_verify_status("NONE"),
    "-")

test("map_verify_status: nil",
    m._map_verify_status(nil),
    "-")

test("map_verify_status: pusty string",
    m._map_verify_status(""),
    "-")

test("map_verify_status: FAILED certificate has expired",
    m._map_verify_status("FAILED:certificate has expired"),
    "Failed:ClientCertExpired")

test("map_verify_status: FAILED cert expired (skrócona forma)",
    m._map_verify_status("FAILED:cert expired"),
    "Failed:ClientCertExpired")

test("map_verify_status: FAILED certificate is not yet valid",
    m._map_verify_status("FAILED:certificate is not yet valid"),
    "Failed:ClientCertNotYetValid")

test("map_verify_status: FAILED not yet valid (skrócona forma)",
    m._map_verify_status("FAILED:not yet valid"),
    "Failed:ClientCertNotYetValid")

test("map_verify_status: FAILED certificate revoked",
    m._map_verify_status("FAILED:certificate revoked"),
    "Failed:ClientCertRevoked")

test("map_verify_status: FAILED revoked (skrócona forma)",
    m._map_verify_status("FAILED:revoked"),
    "Failed:ClientCertRevoked")

test("map_verify_status: FAILED self signed",
    m._map_verify_status("FAILED:self signed certificate"),
    "Failed:ClientCertUntrusted")

test("map_verify_status: FAILED unable to get local issuer",
    m._map_verify_status("FAILED:unable to get local issuer certificate"),
    "Failed:ClientCertUntrusted")

test("map_verify_status: FAILED unable to verify",
    m._map_verify_status("FAILED:unable to verify the first certificate"),
    "Failed:ClientCertUntrusted")

test("map_verify_status: FAILED certificate verify failed",
    m._map_verify_status("FAILED:certificate verify failed"),
    "Failed:ClientCertUntrusted")

test("map_verify_status: FAILED chain too long",
    m._map_verify_status("FAILED:chain too long"),
    "Failed:ClientCertMaxChainDepthExceeded")

test("map_verify_status: FAILED chain depth",
    m._map_verify_status("FAILED:exceeded chain depth"),
    "Failed:ClientCertMaxChainDepthExceeded")

test("map_verify_status: FAILED too long (rozmiar certu)",
    m._map_verify_status("FAILED:certificate too long"),
    "Failed:ClientCertMaxSizeExceeded")

test("map_verify_status: FAILED nieznany błąd",
    m._map_verify_status("FAILED:some unknown openssl error"),
    "Failed:UnmappedConnectionError")

test("map_verify_status: FAILED bez powodu",
    m._map_verify_status("FAILED:"),
    "Failed:UnmappedConnectionError")

-- =============================================================================
-- _nginx_date_to_iso8601
-- =============================================================================

test("nginx_date_to_iso8601: poprawna data",
    m._nginx_date_to_iso8601("Sep 21 22:43:21 2023 GMT"),
    "2023-09-21T22:43:21Z")

test("nginx_date_to_iso8601: styczeń z jednocyfrowym dniem",
    m._nginx_date_to_iso8601("Jan  5 09:00:00 2024 GMT"),
    "2024-01-05T09:00:00Z")

test("nginx_date_to_iso8601: grudzień",
    m._nginx_date_to_iso8601("Dec 31 23:59:59 2025 GMT"),
    "2025-12-31T23:59:59Z")

test("nginx_date_to_iso8601: luty",
    m._nginx_date_to_iso8601("Feb  1 00:00:00 2024 GMT"),
    "2024-02-01T00:00:00Z")

test("nginx_date_to_iso8601: nil",
    m._nginx_date_to_iso8601(nil),
    nil)

test("nginx_date_to_iso8601: pusty string",
    m._nginx_date_to_iso8601(""),
    nil)

test("nginx_date_to_iso8601: myślnik",
    m._nginx_date_to_iso8601("-"),
    nil)

test("nginx_date_to_iso8601: błędny format (ISO)",
    m._nginx_date_to_iso8601("2023-09-21"),
    nil)

test("nginx_date_to_iso8601: błędny format (epoch)",
    m._nginx_date_to_iso8601("1695336201"),
    nil)

-- =============================================================================
-- _format_validity
-- =============================================================================

test("format_validity: obie daty poprawne",
    m._format_validity("Sep 21 22:43:21 2023 GMT", "Jun 17 22:43:21 2026 GMT"),
    "NotBefore=2023-09-21T22:43:21Z;NotAfter=2026-06-17T22:43:21Z")

test("format_validity: v_start nil",
    m._format_validity(nil, "Jun 17 22:43:21 2026 GMT"),
    "NotBefore=-;NotAfter=2026-06-17T22:43:21Z")

test("format_validity: v_end nil",
    m._format_validity("Sep 21 22:43:21 2023 GMT", nil),
    "NotBefore=2023-09-21T22:43:21Z;NotAfter=-")

test("format_validity: obie nil",
    m._format_validity(nil, nil),
    "-")

test("format_validity: v_start pusty string",
    m._format_validity("", "Jun 17 22:43:21 2026 GMT"),
    "NotBefore=-;NotAfter=2026-06-17T22:43:21Z")

test("format_validity: obie puste stringi",
    m._format_validity("", ""),
    "-")

-- =============================================================================
-- Podsumowanie
-- =============================================================================

local total = pass_count + fail_count
print(string.rep("=", 60))
print(string.format("Wynik: %d PASS  %d FAIL  (%d łącznie)", pass_count, fail_count, total))

if fail_count > 0 then
    os.exit(1)
end
