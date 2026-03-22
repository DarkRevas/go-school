# Token Service - Observability

## Features

### 5.4.1 Structured Logging (slog)
- JSON format output
- Base fields: `service`, `env`, `version`
- `LogError` helper for consistent error logging

### 5.4.2 HTTP Metrics (Prometheus)
- `http_requests_total{method, route, status}` - request counter
- `http_request_duration_seconds{method, route}` - request latency histogram
- `http_requests_in_flight` - current active requests gauge
- `/metrics` endpoint for Prometheus scraping

### 5.4.3 Route Normalization
Converts dynamic paths to stable labels:
- `/api/v1/users/123` → `/api/v1/users/{id}`
- `/api/v1/tokens/abc123` → `/api/v1/tokens/{token}`

### 5.4.4 OpenTelemetry Tracing
- stdout exporter for development
- Spans for: HTTP requests, service operations, DB calls
- Automatic span context propagation

### 5.4.5 Correlation
Logs include `trace_id` and `span_id` when available:
```json
{"level":"INFO","trace_id":"abc123...","span_id":"def456...","msg":"token issued"}
```

### 5.4.6 Integration
Critical scenario "token issuance" includes:
- 3+ metrics (requests_total, duration, in_flight)
- 2+ spans per request (HTTP + service layer)
- Error logs with trace context

## Quick Start

```bash
# Install dependencies
go mod tidy

# Run the service
go run main.go
```

## Test Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Issue a token (triggers full tracing)
curl "http://localhost:8080/api/v1/tokens?user_id=user-123"

# Validate a token
curl "http://localhost:8080/api/v1/tokens/validate?token=token_user-123_123456"

# Get metrics
curl http://localhost:8080/metrics
```

## Verify Observability

### Metrics
```bash
curl -s http://localhost:8080/metrics | grep http_requests
```

### Traces
Traces are printed to stdout in JSON format. Each request logs `trace_id` which can be used to correlate logs with traces.

### Logs with Trace Context
```json
{"level":"INFO","service":"token-service","env":"development","version":"0.0.0","trace_id":"...","span_id":"...","msg":"token issued"}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVICE_NAME` | `token-service` | Service identifier |
| `ENV` | `development` | Environment name |
| `VERSION` | `0.0.0` | Service version |
