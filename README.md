# todo-api-go

A minimal HTTP API for todo management, built in Go with no external web framework.

## Quick Start

### Prerequisites
- Go 1.23+
- MySQL 8.0+

### Setup

1. Prepare configuration files:
   - Create a config file (template: `docs/config/config.json`)
   - Create a password file with your DB password:
     ```json
     {"password": "your-db-password"}
     ```

2. Set environment variables:
   ```bash
   export CONFIG_FILE=/path/to/config.json
   export PASSWORD_FILE=/path/to/password.json
   ```

3. Build and run:
   ```bash
   go build -o todo-api-go ./start.go
   ./todo-api-go
   ```

## Development

**Build:**
```bash
go build -o todo-api-go ./start.go
```

**Run tests:**
```bash
go test ./...
```

**Run a specific test:**
```bash
go test ./internal/servr/... -run TestName
```

## Configuration

The server requires two configuration files:

**config.json** — Application configuration
- `authServerUrl` — External identity provider endpoint for credential validation
- `right` — Required access right for users (e.g., `attach_connect`)
- `storageFolder` — Local directory for uploaded attachment files (stored as `<uuid>.dat`)
- `context` — URL prefix for all routes (e.g., `/todoapi`)

**password.json** — Database credentials
```json
{"password": "your-database-password"}
```

## Architecture

The application is organized in three layers:

**Controllers** (`internal/controller/`) — HTTP request handlers
- Each endpoint is a struct implementing `http.Handler`
- All handlers are wired in `StartRouter` (`router.go`)
- CORS and logging middlewares wrap the mux
- Session auth via `Authorization: Bearer <token>` header or `token12` cookie

**Services** (`internal/servr/`) — Business logic
- `Servr` interface is the single entry point for all operations
- Controllers obtain an instance via `servr.GetServr(db, config)`
- Each method wraps work in a database transaction
- External authentication via HTTP POST to configured IDP

**Data Access** (`internal/servr/dao/`) — SQL queries
- Raw `database/sql` queries against MySQL
- All DAO functions take a `*sql.Tx` (transactions managed at service layer)
- Dynamic IN-clause queries use `_here_` placeholder replacement

**Supporting packages:**
- `internal/data/` — Request/response structs and configuration types
- `internal/cst/` — Shared constants (cookie name, content-type, priority bounds, etc.)
- `internal/start/` — Startup sequence (config loading, DB connection, router startup)

## Database

MySQL schema is in `docs/db/db.sql`. Incremental schema changes are tracked in `docs/db/changes.sql`.

Key tables: `person`, `session`, `session_person`, `todo_item`, `item`, `attachment`.

All IDs are stored as UUID strings in `varchar(64)` columns.

## Deployment

The CI pipeline (`build.yml`) produces a static Linux binary in an Alpine container and pushes it to AWS ECR. The binary has no external runtime dependencies.

Build for Linux:
```bash
GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o todo-api-go ./start.go
```

## License

See the [LICENSE](LICENSE) file for details.
