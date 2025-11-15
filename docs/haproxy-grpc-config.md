# HAProxy gRPC Configuration (TCP Mode)

Best practices HAProxy configuration for nanabush gRPC service using TCP mode (pure socket, no L7 termination).

## Frontend Configuration

```haproxy
frontend grpc-nanabush
    # Bind to all interfaces on port 50051
    bind *:50051
    
    # TCP mode - no L7 processing, pure socket forwarding
    mode tcp
    
    # Enable TCP log format for detailed connection logging
    option tcplog
    
    # Enable logging to syslog or file
    # Log format: tcplog with additional fields
    log global
    log-format "%ci:%cp [%t] %ft %b/%s %Tw/%Tc/%Tt %B %ts %ac/%fc/%bc/%sc/%rc %sq/%bq %hr %hs %{+Q}r"
    
    # Connection timeout - how long to wait for a connection attempt
    timeout connect 10s
    
    # Client timeout - how long to wait for client data
    timeout client 300s
    
    # Client fin timeout - time to wait for FIN after client closes
    timeout client-fin 10s
    
    # Queue timeout - max time a client can wait in queue
    timeout queue 30s
    
    # Maximum connection rate from a single IP (connections per second)
    stick-table type ip size 100k expire 30s store conn_rate(10s)
    tcp-request connection track-sc0 src
    tcp-request connection reject if { sc0_conn_rate gt 50 }
    
    # Maximum connections per IP
    stick-table type ip size 100k expire 30s store conn_cur
    tcp-request connection track-sc1 src
    tcp-request connection reject if { sc1_conn_cur gt 100 }
    
    # Enable connection tracking for statistics
    option log-health-checks
    
    # Capture request for logging (first 128 bytes of TCP payload)
    capture request header len 128
    
    # Capture response for logging (first 128 bytes of TCP payload)
    capture response header len 128
    
    # Default backend
    default_backend nanabush-grpc
```

## Backend Configuration

```haproxy
backend nanabush-grpc
    # TCP mode
    mode tcp
    
    # Enable TCP log format
    option tcplog
    
    # Connection balancing algorithm
    # 'first' - use first available server (good for single server)
    # 'leastconn' - use server with least connections (better for multiple servers)
    balance first
    
    # Enable health checks
    option tcp-check
    
    # TCP health check - connect to port and expect immediate close or connection
    tcp-check connect
    tcp-check expect string "" # Any response is fine for gRPC
    
    # Health check interval
    default-server check inter 10s fall 3 rise 2
    
    # Timeout settings
    timeout connect 10s
    timeout server 300s
    timeout server-fin 10s
    timeout queue 30s
    
    # Retry connections to backend
    retries 3
    
    # Server definition
    # Replace 10.20.2.0 with your MetalLB-assigned IP
    server nanabush-1 10.20.2.0:50051 check resolvers dns resolve-prefer ipv4
    
    # For multiple servers (if you scale later):
    # server nanabush-1 10.20.2.0:50051 check
    # server nanabush-2 10.20.2.1:50051 check backup
```

## Additional Global Settings (if needed)

Add these to your global section for better TCP monitoring:

```haproxy
global
    # Enable stats socket for monitoring
    stats socket /var/run/haproxy/admin.sock mode 660 level admin
    
    # Logging
    log stdout format raw local0
    # Or log to file:
    # log /var/log/haproxy/haproxy.log local0
    
    # Enable detailed TCP stats
    tune.bufsize 16384
    tune.maxrewrite 1024
    
    # TCP settings optimized for gRPC/HTTP/2
    tune.ssl.default-dh-param 2048
    tune.ssl.cachesize 10000
```

## Log Format Explanation

The custom log format provides:
- `%ci:%cp` - Client IP and port
- `[%t]` - Timestamp
- `%ft` - Frontend name
- `%b/%s` - Backend name / Server name
- `%Tw/%Tc/%Tt` - Time to wait / Time to connect / Total time
- `%B` - Bytes read from server
- `%ts` - Termination state
- `%ac/%fc/%bc/%sc/%rc` - Connection counts (active/frontend/backend/server/retries)
- `%sq/%bq` - Queue positions
- `%hr/%hs` - Request/response headers
- `%{+Q}r` - Query string (if any)

## Monitoring and Statistics

```haproxy
# Stats page (if you want HTTP stats - optional, separate from gRPC)
listen stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 10s
    stats admin if TRUE
```

## Best Practices for gRPC/TCP

1. **Connection Limits**: Prevents DDoS and resource exhaustion
   - Rate limiting per IP
   - Maximum connections per IP
   - Queue timeouts

2. **Timeouts**: Important for long-lived gRPC connections
   - Client timeout: 300s (5 min) - gRPC connections can be long-lived
   - Server timeout: 300s - match client timeout
   - Connect timeout: 10s - fail fast if backend is down

3. **Health Checks**: Monitor backend availability
   - TCP check (simpler, faster)
   - Or gRPC health check (more accurate but requires L7)

4. **Logging**: Capture connection details for debugging
   - TCP log format for connection-level details
   - Capture headers (first 128 bytes) for basic inspection
   - Log health check failures

5. **Connection Tracking**: Track per-IP statistics
   - Connection rate limiting
   - Connection count limits
   - Useful for monitoring and security

6. **Retries**: Automatic failover
   - Retry failed connections
   - Multiple backend servers for HA

## DNS Resolution (for dynamic IPs)

If you need DNS resolution (useful if MetalLB IP changes):

```haproxy
resolvers dns
    nameserver dns1 8.8.8.8:53
    nameserver dns2 8.8.4.4:53
    resolve_retries 3
    timeout resolve 1s
    timeout retry 1s
    hold other 30s
    hold refused 30s
    hold nx 30s
    hold timeout 30s
    hold valid 10s
    hold obsolete 30s
```

Then in backend:
```haproxy
server nanabush-1 nanabush-grpc-server.nanabush.svc.cluster.local:50051 check resolvers dns resolve-prefer ipv4
```

## Example Complete Configuration

```haproxy
global
    log stdout format raw local0
    stats socket /var/run/haproxy/admin.sock mode 660 level admin
    
    # TCP optimizations
    tune.bufsize 16384
    tune.maxrewrite 1024

defaults
    mode tcp
    option tcplog
    timeout connect 10s
    timeout client 300s
    timeout server 300s
    retries 3

frontend grpc-nanabush
    bind *:50051
    mode tcp
    option tcplog
    log global
    log-format "%ci:%cp [%t] %ft %b/%s %Tw/%Tc/%Tt %B %ts %ac/%fc/%bc/%sc/%rc %sq/%bq %hr %hs %{+Q}r"
    
    # Connection rate limiting
    stick-table type ip size 100k expire 30s store conn_rate(10s)
    tcp-request connection track-sc0 src
    tcp-request connection reject if { sc0_conn_rate gt 50 }
    
    # Connection count limiting
    stick-table type ip size 100k expire 30s store conn_cur
    tcp-request connection track-sc1 src
    tcp-request connection reject if { sc1_conn_cur gt 100 }
    
    # Logging
    option log-health-checks
    capture request header len 128
    capture response header len 128
    
    default_backend nanabush-grpc

backend nanabush-grpc
    mode tcp
    option tcplog
    balance first
    option tcp-check
    tcp-check connect
    
    timeout connect 10s
    timeout server 300s
    timeout server-fin 10s
    
    default-server check inter 10s fall 3 rise 2
    
    server nanabush-1 10.20.2.0:50051 check
```

