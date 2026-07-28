# Mock API & Database Sandbox Collaborative

A Go-based sandbox that **dynamically generates mock REST APIs** backed by your choice of database: **Firestore**, **PostgreSQL**, or **MySQL**. Frontend developers can design schema and test against realistic endpoints without waiting on a backend team.

Built with Go 1.24 and designed for high-concurrency workloads. Ships with a built-in **Web UI** (Vite + React) for visual sandbox management.

## Why this exists

Distributed teams frequently hit a bottleneck: frontend engineers wait on backend engineers to finish an API before they can integrate. This service solves that by giving any team the ability to **spin up a sandboxed, stateful mock API on demand**.

Data submitted to the sandbox is actually persisted in your chosen database, so a `POST` followed by a `GET` returns the data you just sent, mimicking real backend behaviour rather than serving static fixtures.

## Features

- **Multi-database support.** Choose your backend: Firestore, PostgreSQL, or MySQL.
- **Dynamic sandbox endpoints.** CRUD routes are generated on the fly per project: `/sandbox/{projectId}/{table}` and `/sandbox/{projectId}/{table}/{id}`.
- **Stateful mock data.** Writes (`POST`, `PUT`, `PATCH`) are persisted to your configured database under project-scoped paths.
- **Auto-generated IDs.** `POST` without an `{id}` generates a UUID.
- **Auto-stamped metadata.** Created/updated timestamps; `PUT`/`PATCH` use `MergeAll` for partial updates.
- **Server-side pagination.** `GET /sandbox/{projectId}/{table}` accepts `?limit=N&offset=M` (default 25, max 100) and returns a `{ data, limit, offset, count, nextOffset }` envelope.
- **Quota-aware errors.** Firestore `ResourceExhausted` / `Quota exceeded` is surfaced as HTTP `503` with a retry-friendly message.
- **Non-blocking access logging.** HTTP handlers hand off log entries to a bounded channel; a worker pool writes them to Firestore in the background (separate from sandbox data).
- **Bounded resources.** Channel buffer, worker count, and write timeouts are all configurable; channel-full drops are counted and never block the client response.
- **FIFO log retention.** A background routine prunes each project to the configured cap (default 100) on a ticker.
- **Graceful shutdown.** `Logger.Close()` is idempotent, signal-safe, and cancels the cleanup ticker cleanly.
- **Cloud-Run ready.** Falls back to Application Default Credentials when `GOOGLE_APPLICATION_CREDENTIALS` is not set.

## Architecture Overview

```
┌──────────┐   request   ┌──────────────────┐  Submit(entry)  ┌──────────────┐
│  client  │ ──────────► │ middleware.      │ ──────────────► │ logger.Logger│
└──────────┘             │ Logging(...)     │   (non-blocking)│  (Firestore) │
                         └──────────────────┘                 └──────────────┘
                                                                      │
                         ┌──────────────────┐                           │
                         │ sandbox.Handler  │ ◄───────────────────────┘
                         │ (CRUD on         │   Sandbox data stored in
                         │  sandbox/{pid}/  │   chosen database:
                         │  {table}/{id})   │   Firestore / PostgreSQL / MySQL
                         └──────────────────┘
```

**Key design points:**

- **Database abstraction.** The `database.DatabaseClient` interface allows seamless switching between Firestore, PostgreSQL, and MySQL backends without changing handler code.
- **Separate concerns.** Sandbox data (user data) and access logs are stored in different systems. Logs always go to Firestore; sandbox data goes to your configured database.
- **`logger.Submitter`** **interface.** `middleware.Logging` depends on a `Submitter` (single-method: `Submit(*LogEntry) bool`), not the concrete `*logger.Logger`. This keeps the middleware trivially unit-testable.
- **Defer-based latency measurement.** Latency is measured in a `defer` block so it captures panics in the downstream handler as well.
- **Drop-on-full semantics.** When the in-process channel is saturated, the entry is dropped and a warning is logged. The client response is **never** delayed by logging.

## Project layout

```
.
├── main.go                    # Entry point: loads config, wires services, starts HTTP
├── go.mod / go.sum
├── Makefile                   # Common dev tasks
├── .env.example               # Copy to .env (gitignored) and fill in real values
├── .gitignore
│
├── internal/                  # Private application packages (not importable externally)
│   │
│   ├── config/                # Env loading & defaults
│   │   ├── config.go
│   │   └── config_test.go
│   │
│   ├── database/              # Database abstraction layer
│   │   ├── db_interface.go    # DatabaseClient interface, GetAllOptions, factory
│   │   ├── firestore_adapter.go # Firestore implementation (active adapter)
│   │   ├── postgresql.go      # PostgreSQL implementation
│   │   ├── mysql.go           # MySQL implementation
│   │   ├── utils.go           # Table path parsing utilities
│   │   └── database_test.go
│   │
│   ├── logger/                # Async access-logger package
│   │   ├── entry.go           # LogEntry struct + Submitter interface
│   │   ├── config.go          # LoggerConfig + defaults
│   │   ├── logger.go          # Logger core, NewLogger, Submit, Close
│   │   ├── worker.go          # Worker pool & write path
│   │   ├── cleanup.go         # FIFO cleanup ticker
│   │   ├── batch.go           # Batched Firestore deletes
│   │   ├── config_test.go
│   │   └── logger_test.go
│   │
│   ├── middleware/            # HTTP middlewares
│   │   ├── logging.go         # Logging(Submitter) -> http.Handler middleware
│   │   ├── cors.go            # CORS middleware
│   │   ├── apikey.go          # API key auth middleware
│   │   └── logging_test.go
│   │
│   ├── sandbox/               # Dynamic mock-API handler
│   │   ├── handler.go         # CRUD handler for /sandbox/{projectId}/{table}/{id}
│   │   └── sandbox_test.go
│   │
│   └── apikey/                # API key management
│       └── keys.go
│
└── web-ui/                    # Vite + React 18 admin console (JS, Tailwind)
    ├── README.md              # Frontend-specific docs
    ├── package.json
    ├── vite.config.js         # Dev proxy to :8080
    └── src/                   # App entry, pages, components, stores, api
```

## HTTP routes

| Method   | Path                                | Description                                          |
| -------- | ----------------------------------- | ---------------------------------------------------- |
| `GET`    | `/hello`                            | Health probe — returns `Hello, Firebase World!`      |
| `GET`    | `/sandbox/{projectId}/{table}`      | List documents (supports `?limit&offset`, see below) |
| `GET`    | `/sandbox/{projectId}/{table}/{id}` | Fetch a single document                              |
| `POST`   | `/sandbox/{projectId}/{table}`      | Create a new document (auto-generated UUID)          |
| `POST`   | `/sandbox/{projectId}/{table}/{id}` | Create or overwrite a document with a known id       |
| `PUT`    | `/sandbox/{projectId}/{table}/{id}` | Partial update (MergeAll)                            |
| `PATCH`  | `/sandbox/{projectId}/{table}/{id}` | Partial update (MergeAll)                            |
| `DELETE` | `/sandbox/{projectId}/{table}/{id}` | Delete a document                                    |

All routes are wrapped with the access-logging middleware.

### List endpoint pagination

`GET /sandbox/{projectId}/{table}` accepts two optional query parameters:

| Param    | Default | Max   | Description                                  |
| -------- | ------- | ----- | -------------------------------------------- |
| `limit`  | `25`    | `100` | Maximum documents to return in this page     |
| `offset` | `0`     | —     | Number of documents to skip before this page |

The response is wrapped in an envelope so the UI can render pagination controls:

```json
{
  "data":       [ { "id": "...", "...": "..." } ],
  "limit":      25,
  "offset":     0,
  "count":      17,
  "nextOffset": 17
}
```

If the underlying adapter hits a backend quota (e.g. Firestore's 1 MiB response
limit), the server returns HTTP `503 Service Unavailable` with a message
suggesting a smaller `?limit=` or a retry.

### Data shape

`POST`/`PUT`/`PATCH` accept any JSON object. The handler automatically adds server-side metadata:

```json
{
  "_createdAt": "<ServerTimestamp>",
  "_createdBy": "anonymous",
  "_updatedAt": "<ServerTimestamp>",
  "_updatedBy": "anonymous",
  "...your fields": "..."
}
```

## Configuration

Copy the example env file and edit:

```bash
cp .env.example .env
```

### Core Configuration

| Variable               | Default               | Description                                             |
| ---------------------- | --------------------- | ------------------------------------------------------- |
| `PORT`                 | `8080`                | HTTP listen port                                        |
| `DATABASE_TYPE`        | _(unset)_             | Database backend: `firestore`, `postgresql`, or `mysql` |
| `GOOGLE_CLOUD_PROJECT` | `mockapi-sandbox-dev` | GCP project used for Firestore + log writes             |

### Database-Specific Configuration

**Firestore:**

| Variable                         | Default               | Description                                          |
| -------------------------------- | --------------------- | ---------------------------------------------------- |
| `GOOGLE_CLOUD_PROJECT`           | `mockapi-sandbox-dev` | Firestore project ID                                 |
| `GOOGLE_APPLICATION_CREDENTIALS` | _(unset)_             | Path to service-account JSON. If unset, ADC is used. |

**PostgreSQL:**

| Variable            | Description                                |
| ------------------- | ------------------------------------------ |
| `POSTGRES_HOST`     | Database host                              |
| `POSTGRES_PORT`     | Database port (default: 5432)              |
| `POSTGRES_USER`     | Username                                   |
| `POSTGRES_PASSWORD` | Password                                   |
| `POSTGRES_DB`       | Database name                              |
| `POSTGRES_SSL_MODE` | SSL mode (disable, allow, prefer, require) |

**MySQL:**

| Variable         | Description                   |
| ---------------- | ----------------------------- |
| `MYSQL_HOST`     | Database host                 |
| `MYSQL_PORT`     | Database port (default: 3306) |
| `MYSQL_USER`     | Username                      |
| `MYSQL_PASSWORD` | Password                      |
| `MYSQL_DB`       | Database name                 |

### Logger Configuration

| Variable                   | Default | Description                        |
| -------------------------- | ------- | ---------------------------------- |
| `LOG_CHANNEL_BUFFER`       | `1000`  | Size of the in-process log channel |
| `LOG_NUM_WORKERS`          | `10`    | Number of worker goroutines        |
| `LOG_CLEANUP_INTERVAL`     | `5m`    | How often the FIFO cleanup runs    |
| `LOG_MAX_LOGS_PER_PROJECT` | `100`   | Cap on logs retained per project   |

## Quick start

```bash
# Run tests
make test

# Build and run
make run

# Or build a binary
make build
./server.exe
```

Smoke-test the running server:

```bash
# Health probe
curl http://localhost:8080/hello
# -> Hello, Firebase World!

# Create a document
curl -X POST http://localhost:8080/sandbox/demo/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Ada","role":"engineer"}'

# List documents
curl http://localhost:8080/sandbox/demo/users
```

Logs appear in Firestore at:

```
/projects/{GOOGLE_CLOUD_PROJECT}/logs/{logId}
```

Sandbox data appears at (depending on your `DATABASE_TYPE`):

- **Firestore:** `/projects/{GOOGLE_CLOUD_PROJECT}/sandbox/{projectId}/{table}/{id}`
- **PostgreSQL/MySQL:** Tables created under configured database

## Web UI

A Vite + React 18 admin console lives in [`web-ui/`](web-ui/). It is the
recommended way to explore the sandbox without writing `curl` calls by hand.

```bash
# 1. start the Go backend (terminal A)
make run

# 2. start the web UI (terminal B)
cd web-ui
npm install
npm run dev
```

The Vite dev server runs on <http://localhost:5173> and proxies `/sandbox` and
`/hello` to the Go backend on <http://localhost:8080>, so the UI uses relative
URLs and the browser never sees a CORS preflight.

Highlights:

- **Overview** — Live stats per project, table, and document count.
- **Projects / Tables / Documents** — Visual CRUD with bulk operations and
  server-side pagination (default 25/page, options 10/25/50/100).
- **REST Tester** — Arbitrary `GET/POST/PUT/PATCH/DELETE` against any sandbox
  path, with JSON body editing and a response viewer (`Ctrl+Enter` to send).
- **Access Logs / Settings** — Local-only views; the Go backend owns the
  canonical Firestore log store.
- **Light / Dark theme** with smooth transitions and a responsive bottom-nav
  on mobile.

See [`web-ui/README.md`](web-ui/README.md) for the full frontend documentation.

## Testing

The project ships with unit tests for all packages that have meaningful in-process logic:

| Package      | Coverage                           | What's tested                                                                                                                       |
| ------------ | ---------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `internal/config`     | `config_test.go`                   | Default values, env override, invalid-fallback semantics                                                                            |
| `internal/database`   | `database_test.go`                 | Interface compliance, factory function, table path parsing                                                                          |
| `internal/logger`     | `config_test.go`, `logger_test.go` | Config defaults, non-blocking submit, nil-drop, project tracking (thread-safe), close idempotency                                   |
| `internal/middleware` | `logging_test.go`                  | Method/path/projectID recording, status propagation, latency/timestamp, drop on full channel, entry on panic, one-entry-per-request |
| `internal/sandbox`    | `sandbox_test.go`                  | CRUD operations, error handling, not-found scenarios                                                                                |

`internal/database/firestore_adapter.go` exercises Firebase/Firestore
end-to-end against a real Firestore instance or the emulator.
`internal/sandbox/handler.go` is covered indirectly by
`internal/sandbox/sandbox_test.go` (via the `MockDatabaseClient`); its thin HTTP
plumbing is the only piece not asserted at the unit level.

```bash
# Run all tests
make test

# Verbose output (per-test)
go test ./... -v

# With coverage
go test ./... -cover
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Security

- **Never commit** `private-key.json`. It is in `.gitignore` for this reason. Rotate any key that has been exposed.
- In production, prefer **Workload Identity** or ADC over service-account key files.
- The sandbox writes `_createdBy` / `_updatedBy` as the literal string `"anonymous"`. Real auth integration is a planned follow-up (the Firebase Auth client is already initialised in `main.go` but currently reserved for future use).
- See `.env.example` for the env contract.

### `.gitignore` coverage

The repository ships with a hardened `.gitignore` that protects against the common leaks when running `go test` or `go build` locally:

| Pattern                                                         | Protects against                       |
| --------------------------------------------------------------- | -------------------------------------- |
| `private-key.json`, `*service-account*.json`                    | Firebase / GCP service-account keys    |
| `/.env`, `.env.local`, `.env.*.local`                           | Local secrets                          |
| `*.pem`, `*.key`, `*.crt`, `*.p12`                              | Generic key material                   |
| `*.exe`, `*.exe~`, `/server.exe`                                | Compiled binaries & editor backups     |
| `*.test`, `*.test.exe`                                          | Test binaries                          |
| `coverage.out`, `coverage.html`, `coverage.txt`, `coverage.xml` | Coverage reports                       |
| `*.prof`, `*.trace`, `gocover/`                                 | Profiling & native coverage (Go 1.20+) |
| `bin/`, `dist/`, `tmp/`, `scratch/`                             | Build output & scratch folders         |
| `debug-logs/`                                                   | VS Code Copilot / agent debug logs     |

Verify with `git check-ignore -v <file>` after `git init`.

## Make targets

| Target       | Action                                       |
| ------------ | -------------------------------------------- |
| `make tidy`  | `go mod tidy`                                |
| `make build` | Build `server.exe`                           |
| `make run`   | Run from source (loads `.env` automatically) |
| `make test`  | `go test ./...`                              |
| `make vet`   | `go vet ./...`                               |
| `make fmt`   | `gofmt -w .`                                 |
| `make smoke` | Build, start, hit `/hello`, stop             |
| `make clean` | Remove build artifacts                       |

## License

See [LICENSE](LICENSE)&#x20;
