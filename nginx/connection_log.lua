-- =============================================================================
-- connection_log.lua
-- AWS ALB-style Connection Log — moduł Lua dla OpenResty
-- =============================================================================
-- Odwzorowuje pola AWS ALB Connection Log:
--   timestamp, client_ip, client_port, listener_port,
--   tls_protocol, tls_cipher, tls_handshake_latency,
--   leaf_client_cert_subject, leaf_client_cert_validity,
--   leaf_client_cert_serial_number, tls_verify_status, conn_trace_id
--
-- Użycie w nginx.conf:
--
--   lua_shared_dict tls_timing 10m;
--   lua_package_path "/etc/nginx/lua/?.lua;;";
--
--   server {
--       set $conn_tls_handshake_latency "-";
--       set $conn_leaf_cert_subject     "-";
--       set $conn_leaf_cert_validity    "-";
--       set $conn_leaf_cert_serial      "-";
--       set $conn_tls_verify_status     "-";
--
--       ssl_certificate_by_lua_block {
--           require("connection_log").ssl_certificate_phase()
--       }
--
--       log_by_lua_block {
--           require("connection_log").log_phase()
--       }
--   }
-- =============================================================================

local _M = {}
_M._VERSION = "1.0.0"

-- Nazwa shared dict zadeklarowanego w nginx.conf
local SHARED_DICT = "tls_timing"

-- TTL wpisu w shared dict (sekundy)
-- Musi być dłuższy niż maksymalny czas TLS handshake
local ENTRY_TTL = 30

-- -----------------------------------------------------------------------------
-- Prywatne: bezpieczne pobranie zmiennej nginx
-- Zwraca "-" gdy nil lub pusty string — zgodnie z konwencją AWS ALB
-- -----------------------------------------------------------------------------
local function get_var(name)
    local ok, val = pcall(function() return ngx.var[name] end)
    if not ok or not val or val == "" then
        return "-"
    end
    return val
end

-- -----------------------------------------------------------------------------
-- Prywatne: unikalny klucz dla połączenia w shared dict
-- Kombinacja ip:port jest unikalna per aktywne połączenie TCP
-- -----------------------------------------------------------------------------
local function conn_key()
    -- Bezpieczne pobieranie — w niektórych kontekstach (ssl_certificate_by_lua)
    -- dostęp do niektórych zmiennych może być zablokowany. Najpierw spróbuj
    -- odczytać `remote_addr:remote_port`. Jeśli to nie zadziała, spróbuj
    -- `connection` (dostępne w obu kontekstach). W ostateczności użyj
    -- unikalnego fallbacku.
    local ok, addr = pcall(function() return ngx.var.remote_addr end)
    local ok2, port = pcall(function() return ngx.var.remote_port end)
    if ok and ok2 and addr and addr ~= "" and port and port ~= "" then
        return addr .. ":" .. port
    end

    local ok3, conn = pcall(function() return ngx.var.connection end)
    if ok3 and conn and conn ~= "" then
        return "conn:" .. tostring(conn)
    end

    local ok4, sess = pcall(function() return ngx.var.ssl_session_id end)
    if ok4 and sess and sess ~= "" then
        return "sess:" .. tostring(sess)
    end

    return "anon:" .. tostring(ngx.now()) .. ":" .. tostring(math.random(1000000))
end

-- -----------------------------------------------------------------------------
-- Prywatne: parsowanie daty z formatu nginx do ISO 8601
--
-- nginx $ssl_client_v_start / $ssl_client_v_end:
--   "Sep 21 22:43:21 2023 GMT"
--
-- AWS ALB format:
--   "2023-09-21T22:43:21Z"
-- -----------------------------------------------------------------------------
local MONTHS = {
    Jan="01", Feb="02", Mar="03", Apr="04",
    May="05", Jun="06", Jul="07", Aug="08",
    Sep="09", Oct="10", Nov="11", Dec="12"
}

local function nginx_date_to_iso8601(date_str)
    if not date_str or date_str == "" or date_str == "-" then
        return nil
    end

    -- Pattern: "Sep 21 22:43:21 2023 GMT"
    local mon, day, time_part, year =
        date_str:match("^(%a%a%a)%s+(%d+)%s+([%d:]+)%s+(%d%d%d%d)")

    if not mon or not MONTHS[mon] then
        ngx.log(ngx.WARN, "connection_log: cannot parse date: ", date_str)
        return nil
    end

    return string.format("%s-%s-%02dT%sZ",
        year,
        MONTHS[mon],
        tonumber(day),
        time_part
    )
end

-- -----------------------------------------------------------------------------
-- Prywatne: formatowanie pola leaf_client_cert_validity
--
-- AWS ALB format:
--   "NotBefore=2023-09-21T22:43:21Z;NotAfter=2026-06-17T22:43:21Z"
-- -----------------------------------------------------------------------------
local function format_validity(v_start_raw, v_end_raw)
    local v_start = nginx_date_to_iso8601(v_start_raw)
    local v_end   = nginx_date_to_iso8601(v_end_raw)

    if not v_start and not v_end then
        return "-"
    end

    return string.format("NotBefore=%s;NotAfter=%s",
        v_start or "-",
        v_end   or "-"
    )
end

-- -----------------------------------------------------------------------------
-- Prywatne: mapowanie $ssl_client_verify na kody AWS ALB tls_verify_status
--
-- nginx $ssl_client_verify zwraca:
--   "SUCCESS"
--   "FAILED:<openssl reason string>"
--   "NONE"  (klient nie podał certyfikatu)
--
-- AWS ALB tls_verify_status zwraca:
--   "Success"
--   "Failed:ClientCertUntrusted"
--   "Failed:ClientCertExpired"
--   "Failed:ClientCertNotYetValid"
--   "Failed:ClientCertRevoked"
--   "Failed:ClientCertMaxChainDepthExceeded"
--   "Failed:ClientCertMaxSizeExceeded"
--   "Failed:UnmappedConnectionError"
--   "-" gdy mTLS wyłączone lub klient nie podał certu
-- -----------------------------------------------------------------------------
local function map_verify_status(verify_raw)
    if not verify_raw or verify_raw == "" then
        return "-"
    end

    if verify_raw == "SUCCESS" then
        return "Success"
    end

    -- Klient nie podał certyfikatu (ssl_verify_client optional)
    if verify_raw == "NONE" then
        return "-"
    end

    -- nginx zwraca "FAILED:<reason>" — mapujemy reason na kody ALB
    local reason = verify_raw:match("[Ff][Aa][Ii][Ll][Ee][Dd]:?%s*(.*)")
    if not reason then
        return "Failed:UnmappedConnectionError"
    end

    reason = reason:lower()

    if reason:find("certificate has expired")
    or reason:find("cert.*expired")
    or reason:find("expired") then
        return "Failed:ClientCertExpired"
    end

    if reason:find("certificate is not yet valid")
    or reason:find("not yet valid") then
        return "Failed:ClientCertNotYetValid"
    end

    if reason:find("certificate revoked")
    or reason:find("revoked") then
        return "Failed:ClientCertRevoked"
    end

    if reason:find("self.signed")
    or reason:find("unable to get local issuer")
    or reason:find("unable to verify")
    or reason:find("certificate verify failed") then
        return "Failed:ClientCertUntrusted"
    end

    if reason:find("chain too long")
    or reason:find("chain depth") then
        return "Failed:ClientCertMaxChainDepthExceeded"
    end

    if reason:find("too long")
    or reason:find("size") then
        return "Failed:ClientCertMaxSizeExceeded"
    end

    return "Failed:UnmappedConnectionError"
end

-- =============================================================================
-- Publiczne API
-- =============================================================================

-- -----------------------------------------------------------------------------
-- ssl_certificate_phase()
-- Wywoływana z ssl_certificate_by_lua_block{}
--
-- Zapisuje czas początku TLS handshake do shared dict.
-- Ten callback jest wywoływany podczas negocjacji TLS, przed HTTP requestem.
-- Dzięki temu możemy obliczyć dokładne tls_handshake_latency w log_phase().
-- -----------------------------------------------------------------------------
function _M.ssl_certificate_phase()
    local shared = ngx.shared[SHARED_DICT]
    if not shared then
        ngx.log(ngx.ERR,
            "connection_log: shared dict '", SHARED_DICT,
            "' nie istnieje — sprawdź lua_shared_dict w nginx.conf"
        )
        return
    end

    local key = conn_key()
    local ok, err, forcible = shared:set(key, ngx.now(), ENTRY_TTL)
    -- Logowanie debugowe (będzie widoczne w error.log)
    local ok_log, _ = pcall(function()
        ngx.log(ngx.ERR, "connection_log: ssl_certificate_phase set key=", tostring(key), " ok=", tostring(ok), " err=", tostring(err))
    end)

    if not ok then
        ngx.log(ngx.WARN,
            "connection_log: nie można zapisać TLS start time: ", err
        )
    elseif forcible then
        -- shared dict jest pełny — usunięto stary wpis by zmieścić nowy
        ngx.log(ngx.WARN,
            "connection_log: shared dict pełny, usunięto stary wpis"
        )
    end
end

-- -----------------------------------------------------------------------------
-- log_phase()
-- Wywoływana z log_by_lua_block{}
--
-- Oblicza tls_handshake_latency i ustawia zmienne nginx które są używane
-- przez log_format w nginx.conf:
--
--   $conn_tls_handshake_latency   — czas handshake w sekundach (np. "0.043")
--   $conn_leaf_cert_subject       — DN certyfikatu klienta
--   $conn_leaf_cert_validity      — ważność certyfikatu w formacie AWS
--   $conn_leaf_cert_serial        — numer seryjny certyfikatu
--   $conn_tls_verify_status       — status weryfikacji w formacie AWS
-- -----------------------------------------------------------------------------
function _M.log_phase()
    -- ------------------------------------------------------------------
    -- 1. Oblicz tls_handshake_latency
    -- ------------------------------------------------------------------
    local handshake_latency = "-"
    local shared = ngx.shared[SHARED_DICT]

    if shared then
        -- Przy próbie dopasowania klucza sprawdzamy kilka kandydatów,
        -- ponieważ różne konteksty (ssl_certificate_by_lua vs log_by_lua)
        -- mogą mieć różne ograniczenia co do dostępnych zmiennych.
        local candidates = {}
        -- adres:port
        local ok, addr = pcall(function() return ngx.var.remote_addr end)
        local ok2, port = pcall(function() return ngx.var.remote_port end)
        if ok and ok2 and addr and addr ~= "" and port and port ~= "" then
            table.insert(candidates, addr .. ":" .. port)
        end
        -- connection id
        local ok3, conn = pcall(function() return ngx.var.connection end)
        if ok3 and conn and conn ~= "" then
            table.insert(candidates, "conn:" .. tostring(conn))
        end
        -- ssl session id
        local ok4, sess = pcall(function() return ngx.var.ssl_session_id end)
        if ok4 and sess and sess ~= "" then
            table.insert(candidates, "sess:" .. tostring(sess))
        end

        local start_time
        local matched_key
        for _, k in ipairs(candidates) do
            start_time = shared:get(k)
            if start_time then
                matched_key = k
                break
            end
        end

        if start_time then
            local latency = ngx.now() - start_time
            handshake_latency = string.format("%.3f", latency)
            if matched_key then
                shared:delete(matched_key)
            end
        else
            handshake_latency = "-"
        end
    end

    -- ------------------------------------------------------------------
    -- 2. Pobierz surowe dane certyfikatu klienta
    -- ------------------------------------------------------------------
    local v_start_raw = get_var("ssl_client_v_start")
    local v_end_raw   = get_var("ssl_client_v_end")

    -- ------------------------------------------------------------------
    -- 3. Zmapuj i sformatuj pola
    -- ------------------------------------------------------------------
    local verify_status = map_verify_status(get_var("ssl_client_verify"))
    local validity      = format_validity(
        v_start_raw ~= "-" and v_start_raw or nil,
        v_end_raw   ~= "-" and v_end_raw   or nil
    )

    -- ------------------------------------------------------------------
    -- 4. Ustaw zmienne nginx dla log_format
    -- Zapis przez ngx.var — zmienne muszą być zadeklarowane przez "set"
    -- w bloku server{} w nginx.conf
    -- ------------------------------------------------------------------
    local function safe_set(var_name, value)
        local ok, err = pcall(function()
            ngx.var[var_name] = value
        end)
        if not ok then
            ngx.log(ngx.ERR,
                "connection_log: nie można ustawić ", var_name, ": ", err
            )
        end
    end

    safe_set("conn_tls_handshake_latency", handshake_latency)
    safe_set("conn_leaf_cert_subject",     get_var("ssl_client_s_dn"))
    safe_set("conn_leaf_cert_validity",    validity)
    safe_set("conn_leaf_cert_serial",      get_var("ssl_client_serial"))
    safe_set("conn_tls_verify_status",     verify_status)
end

-- =============================================================================
-- Eksport funkcji pomocniczych dla testów jednostkowych
-- Przykład:
--   local m = require("connection_log")
--   assert(m._map_verify_status("FAILED:certificate has expired")
--          == "Failed:ClientCertExpired")
-- =============================================================================
_M._map_verify_status     = map_verify_status
_M._format_validity       = format_validity
_M._nginx_date_to_iso8601 = nginx_date_to_iso8601

return _M
