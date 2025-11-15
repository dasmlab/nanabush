# HAProxy Configuration Errors - Fixed

## Errors Fixed

### 1. Stick-Table Name Conflict
**Error**: `stick-table name 'nanabush-grpc' conflicts with table declared in frontend 'nanabush-grpc'`

**Cause**: HAProxy was inferring table names from frontend/backend names, causing conflicts.

**Fix**: Added explicit table names:
- `name grpc_conn_rate` for connection rate tracking
- `name grpc_conn_cur` for connection count tracking

**Correct syntax**:
```haproxy
stick-table type ip size 100k expire 30s store conn_rate(10s) name grpc_conn_rate
tcp-request connection track-sc0 src
tcp-request connection reject if { sc0_conn_rate gt 50 }
```

### 2. Capture Syntax Error
**Error**: `'capture request' expects 'header' <header_name> 'len' <len>.`

**Cause**: In TCP mode, there are no HTTP headers. You must capture payload, not headers.

**Fix**: Changed from `capture request header` to `capture request payload`:
```haproxy
# Wrong (HTTP mode):
capture request header len 128

# Correct (TCP mode):
capture request payload len 128
```

### 3. Frontend-Only Settings in Frontend
**Warnings**:
- `'option log-health-checks' ignored because frontend has no backend capability`
- `'timeout connect' will be ignored because frontend has no backend capability`
- `'timeout queue' will be ignored because frontend has no backend capability`

**Fix**: 
- Removed `option log-health-checks` from frontend (only works in backend)
- Removed `timeout connect` and `timeout queue` from frontend (backend-only)
- Kept only frontend-valid timeouts: `timeout client` and `timeout client-fin`

### 4. Duplicate `option tcplog`
**Warning**: `'option tcplog' overrides previous 'option httplog' in 'defaults' section`

**Fix**: Removed duplicate `option tcplog` (was declared twice in frontend)

### 5. Log Format - Query String in TCP Mode
**Issue**: `%{+Q}r` (query string) doesn't make sense in TCP mode

**Fix**: Removed `%{+Q}r` from log format

## Complete Corrected Frontend Configuration

```haproxy
frontend grpc-nanabush
    bind *:50051
    mode tcp
    
    # Logging
    option tcplog
    log global
    log-format "%ci:%cp [%t] %ft %b/%s %Tw/%Tc/%Tt %B %ts %ac/%fc/%bc/%sc/%rc %sq/%bq %hr %hs"
    
    # Capture payload (not headers) in TCP mode
    capture request payload len 128
    capture response payload len 128
    
    # Rate limiting with explicit table names
    stick-table type ip size 100k expire 30s store conn_rate(10s) name grpc_conn_rate
    tcp-request connection track-sc0 src
    tcp-request connection reject if { sc0_conn_rate gt 50 }
    
    stick-table type ip size 100k expire 30s store conn_cur name grpc_conn_cur
    tcp-request connection track-sc1 src
    tcp-request connection reject if { sc1_conn_cur gt 100 }
    
    # Frontend timeouts only
    timeout client 300s
    timeout client-fin 10s
    
    default_backend nanabush-grpc
```

## Key Differences TCP vs HTTP Mode

| Feature | HTTP Mode | TCP Mode |
|---------|-----------|----------|
| Headers | `capture request header` | Not available |
| Payload | `capture request payload` | `capture request payload` |
| Health Checks | `option httpchk` | `option tcp-check` |
| Timeouts | All timeouts available | Frontend: `client`, `client-fin`<br>Backend: `connect`, `server`, `queue` |
| Logging | `option httplog` | `option tcplog` |

