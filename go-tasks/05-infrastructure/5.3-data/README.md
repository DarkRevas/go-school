# Token Service - Data Layer

## Schema

### Tables

**users**
- `id` - BIGSERIAL PRIMARY KEY
- `email` - TEXT NOT NULL UNIQUE
- `created_at` - TIMESTAMPTZ DEFAULT NOW()

**accounts**
- `id` - BIGSERIAL PRIMARY KEY
- `user_id` - BIGINT REFERENCES users(id)
- `balance` - BIGINT DEFAULT 0
- `created_at` - TIMESTAMPTZ DEFAULT NOW()

**refresh_tokens**
- `id` - BIGSERIAL PRIMARY KEY
- `user_id` - BIGINT REFERENCES users(id)
- `token` - TEXT NOT NULL UNIQUE
- `expires_at` - TIMESTAMPTZ NOT NULL
- `revoked` - BOOLEAN DEFAULT FALSE
- `created_at` - TIMESTAMPTZ DEFAULT NOW()

## Tools

- **PostgreSQL** - relational database
- **pgx/v5** - PostgreSQL driver with connection pooling
- **sqlc** - type-safe SQL code generation

## Setup

```bash
# Start PostgreSQL
docker-compose up -d

# Run migrations (manual or use a migration tool)
psql -f migrations/001_users.up.sql
psql -f migrations/002_accounts.up.sql
psql -f migrations/003_refresh_tokens.up.sql

# Generate sqlc code
sqlc generate

# Run the service
go run main.go
```

## Architecture

- **Repositories** - direct data access with pgxpool
- **Services** - business logic with transactions
- **Domain errors** - typed errors for application layer
