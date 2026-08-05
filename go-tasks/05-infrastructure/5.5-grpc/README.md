# Token Service - gRPC

## Architecture

```
┌─────────────┐      gRPC      ┌─────────────┐
│  HTTP API   │ ─────────────► │ TokenService│
│  (Auth API) │  :50051        │   (gRPC)    │
│  :8080      │                │             │
└─────────────┘                └─────────────┘
```

## Proto Contract

See `proto/token.proto`:

```protobuf
service TokenService {
  rpc IssueToken(IssueTokenRequest) returns (IssueTokenResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc RevokeToken(RevokeTokenRequest) returns (RevokeTokenResponse);
}
```

## Quick Start

```bash
# Generate proto code
make proto

# Run the service
go run main.go
```

## API Endpoints

### HTTP Gateway (:8080)

**POST /api/v1/tokens** - Issue a new token
```bash
curl -X POST http://localhost:8080/api/v1/tokens \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user-123","scope":"read:write"}'
```

**GET /api/v1/tokens/validate?token=...** - Validate a token
```bash
curl "http://localhost:8080/api/v1/tokens/validate?token=tok_user-123_123456"
```

**POST /api/v1/tokens/revoke** - Revoke a token
```bash
curl -X POST http://localhost:8080/api/v1/tokens/revoke \
  -H "Content-Type: application/json" \
  -d '{"token":"tok_user-123_123456","reason":"logout"}'
```

**GET /health** - Health check
```bash
curl http://localhost:8080/health
```

## Interceptors

### 5.5.5: Server Interceptors

1. **LoggingInterceptor** - Logs all gRPC requests with duration
2. **AuthInterceptor** - Service-to-service auth via `Authorization: Bearer <token>` header
3. **MetricsInterceptor** - Counts requests per method

### Client Configuration

The gRPC client automatically adds the service token for authentication:
```go
client, err := NewTokenClient("localhost:50051")
```

## Integration Test

```bash
# Start the service
go run main.go &

# Issue a token (success)
curl -X POST http://localhost:8080/api/v1/tokens \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user-123"}'
# Response: {"token":"tok_user-123_...","expiresAt":1234567890}

# Validate the token
curl "http://localhost:8080/api/v1/tokens/validate?token=tok_user-123_..."
# Response: {"valid":true,"userId":"user-123","scope":""}

# Revoke the token
curl -X POST http://localhost:8080/api/v1/tokens/revoke \
  -H "Content-Type: application/json" \
  -d '{"token":"tok_user-123_...","reason":"test"}'
# Response: {"revoked":true}

# Validate again (should be invalid)
curl "http://localhost:8080/api/v1/tokens/validate?token=tok_user-123_..."
# Response: {"valid":false}
```

## Error Codes

| gRPC Code | HTTP Status | Description |
|-----------|-------------|-------------|
| `INVALID_ARGUMENT` | 400 | Missing required fields |
| `UNAUTHENTICATED` | 401 | Invalid/missing service token |
| `NOT_FOUND` | 404 | Token not found |
| `INTERNAL` | 500 | Internal server error |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVICE_TOKEN` | `service-secret-token-123` | Internal service auth token |
