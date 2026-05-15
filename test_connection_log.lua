-- Prosty zestaw testów dla connection_log.lua
-- Testuje funkcje eksportowane dla testów: _map_verify_status, _nginx_date_to_iso8601, _format_validity

-- Prosty stub ngx używany przez moduł podczas testów
ngx = ngx or {}
ngx.WARN = ngx.WARN or 4
ngx.ERR = ngx.ERR or 3
ngx.log = function(...) end

-- Upewnij się, że moduł z nginx/ jest w ścieżce modułów
package.path = "./nginx/?.lua;" .. package.path

local ok, conn = pcall(require, "connection_log")
if not ok then
    print("FATAL: nie można załadować modułu connection_log: " .. tostring(conn))
    os.exit(1)
end

local map = conn._map_verify_status
local date_to_iso = conn._nginx_date_to_iso8601
local format_validity = conn._format_validity

local total = 0
local failures = 0

local function assert_eq(actual, expected, name)
    total = total + 1
    if actual == expected then
        print("PASS: " .. name)
    else
        failures = failures + 1
        print("FAIL: " .. name .. " — oczekiwano: '" .. tostring(expected) .. "', otrzymano: '" .. tostring(actual) .. "'")
    end
end

-- Testy map_verify_status
assert_eq(map("SUCCESS"), "Success", "map: SUCCESS -> Success")
assert_eq(map("NONE"), "-", "map: NONE -> -")
assert_eq(map(nil), "-", "map: nil -> -")

assert_eq(map("FAILED:certificate has expired"), "Failed:ClientCertExpired", "map: expired -> ClientCertExpired")
assert_eq(map("FAILED:cert is expired"), "Failed:ClientCertExpired", "map: cert.*expired -> ClientCertExpired")
assert_eq(map("FAILED:certificate is not yet valid"), "Failed:ClientCertNotYetValid", "map: not yet valid -> ClientCertNotYetValid")
assert_eq(map("FAILED:certificate revoked"), "Failed:ClientCertRevoked", "map: revoked -> ClientCertRevoked")
assert_eq(map("FAILED:unable to get local issuer"), "Failed:ClientCertUntrusted", "map: unable to get local issuer -> ClientCertUntrusted")
assert_eq(map("FAILED:chain too long"), "Failed:ClientCertMaxChainDepthExceeded", "map: chain too long -> MaxChainDepthExceeded")
assert_eq(map("FAILED:too long"), "Failed:ClientCertMaxSizeExceeded", "map: too long -> MaxSizeExceeded")
assert_eq(map("FAILED:some weird error"), "Failed:UnmappedConnectionError", "map: unmapped -> UnmappedConnectionError")

-- Testy nginx_date_to_iso8601
local good = "Sep 21 22:43:21 2023 GMT"
assert_eq(date_to_iso(good), "2023-09-21T22:43:21Z", "date_to_iso: poprawna data")
assert_eq(date_to_iso("-"), nil, "date_to_iso: '-' -> nil")
assert_eq(date_to_iso("") , nil, "date_to_iso: '' -> nil")
assert_eq(date_to_iso(nil), nil, "date_to_iso: nil -> nil")
assert_eq(date_to_iso("invalid date"), nil, "date_to_iso: invalid -> nil")

-- Testy format_validity
local v = format_validity(good, good)
assert_eq(v, "NotBefore=2023-09-21T22:43:21Z;NotAfter=2023-09-21T22:43:21Z", "format_validity: obie daty")

local v2 = format_validity(nil, good)
assert_eq(v2, "NotBefore=-;NotAfter=2023-09-21T22:43:21Z", "format_validity: start nil")

local v3 = format_validity(nil, nil)
assert_eq(v3, "-", "format_validity: obie nil -> -")

-- Podsumowanie
print(string.rep("-", 60))
print(string.format("Ran %d tests: %d passed, %d failed", total, total - failures, failures))
if failures > 0 then
    os.exit(1)
else
    os.exit(0)
end
