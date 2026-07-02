# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

**Build:**
```bash
go build -o todo-api-go ./start.go
```

**Build for Linux (cross-compile):**
```bash
GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o todo-api-go ./start.go
```

**Run:**
```bash
CONFIG_FILE=docs/config/config.json PASSWORD_FILE=docs/config/password.json ./todo-api-go
```

**Test:**
```bash
go test ./...
```

**Run single test:**
```bash
go test ./internal/servr/... -run TestName
```

## Configuration

The server requires two environment variables at startup:
- `CONFIG_FILE` — path to JSON config (see `docs/config/config.json` for shape)
- `PASSWORD_FILE` — path to a JSON file containing only `{"password": "..."}` for the DB password

`ServerConfig` fields of note:
- `authServerUrl` — external IDP endpoint (POSTed to for credential validation)
- `right` — the access right string a user must possess (`attach_connect`)
- `storageFolder` — local directory where uploaded attachment files are stored as `<uuid>.dat`
- `context` — URL prefix for all routes (e.g. `/todoapi`)

## Architecture

This is a single-binary Go HTTP API with no framework. Three layers:

### 1. `internal/controller` — HTTP handlers
Each endpoint is a struct implementing `http.Handler`. All handlers are wired in `StartRouter` (`router.go`). The CORS and logging middlewares wrap the mux. All responses go through `writeResponse` (`write.go`), which always returns JSON — errors become `{"message":"..."}` with HTTP 400, successes with HTTP 200.

Session auth is extracted in `getSession` (`session.go`) from either `Authorization: Bearer <token>` header or the `token12` cookie.

### 2. `internal/servr` — business logic
`Servr` interface (`server.go`) is the single entry point for all business operations. Controllers call `servr.GetServr(db, config)` to get an instance — there is no DI container. Each method wraps its work in a DB transaction using the `begin`/`rollback` helpers.

Authentication delegates to an external IDP via HTTP POST (`callIdp` in `login.go`). Sessions support multiple simultaneous logged-in persons (a session holds a list of `person_id` values).

### 3. `internal/servr/dao` — SQL queries
Raw `database/sql` queries against MySQL. All DAO functions take a `*sql.Tx` — transactions are managed in `servr`, not in `dao`. Dynamic IN-clause queries are built by replacing a `_here_` placeholder with the correct number of `?` parameters.

### Supporting packages
- `internal/data` — all request/response structs and config types
- `internal/cst` — shared constants (cookie name, content-type, priority bounds, etc.)
- `internal/start` — startup sequence: load config → connect DB → start router

## Database

MySQL. Schema is in `docs/db/db.sql`. The `docs/db/changes.sql` file tracks incremental schema changes applied after the initial schema.

Key tables: `person`, `session`, `session_person`, `todo_item`, `item`, `attachment`.

All IDs are UUID strings stored as `varchar(64)`.

## Deployment

CI (`build.yml`) builds a static Linux binary inside a `golang:1.23.2-alpine3.20` Docker container, then wraps it in an `amd64/alpine` image pushed to AWS ECR. The binary is the only artifact in the image — no external runtime dependencies.
