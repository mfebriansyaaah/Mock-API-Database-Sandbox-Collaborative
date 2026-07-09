# Mock API & Database Sandbox Collaborative

A Go-based SaaS sandbox that **dynamically generates mock REST APIs** backed by
Google Cloud Firestore. Frontend developers can design, test, and persist data
against realistic endpoints without waiting on a backend team to ship them.
Built with Go 1.22 and designed for high-concurrency workloads.

## Why this exists

Distributed teams frequently hit a bottleneck: frontend engineers wait on
backend engineers to finish an API before they can integrate. This service
solves that by giving any team the ability to **spin up a sandboxed,
stateful mock API on demand** and share a public URL with the frontend.

Data submitted to the sandbox is actually persisted in Firestore, so a `POST`
followed by a `GET` returns the data you just sent — mimicking real backend
behaviour rather than serving static fixtures.

## Features

- **Dynamic sandbox endpoints.** CRUD routes are generated on the fly per
  project: `/sandbox/{projectId}/{table}` and `/sandbox/{projectId}/{table}/{id}`.
- **Stateful mock data.** Writes (`POST`, `PUT`, `PATCH`) are persisted to
  Firestore under project-scoped collection paths (`sandbox/{projectId}/{table}`).
- **Auto-generated IDs.** `POST` without an `{id}` generates a UUID.
- **Auto-stamped metadata.** Created/updated timestamps via
  `firestore.ServerTimestamp`; `PUT`/`PATCH` use `MergeAll` for partial updates.
- **Non-blocking access logging.** HTTP handlers hand off log entries to a
  bounded channel; a worker pool writes them to Firestore in the background.
- **Bounded resources.** Channel buffer, worker count, and write timeouts are
  all configurable; channel-full drops are counted and never block the client
  response.
- **FIFO log retention.** A background routine prunes each project to the
  configured cap (default 100) on a ticker.
- **Graceful shutdown.** `Logger.Close()` is idempotent, signal-safe, and
  cancels the cleanup ticker cleanly.
- **Cloud-Run ready.** Falls back to Application Default Credentials when
  `GOOGLE_APPLICATION_CREDENTIALS` is not set.

## Project layout

```
.
├── main.go                    # Entry point: loads config, wires services, starts HTTP
├── go.mod / go.sum
├── Makefile                   # Common dev tasks
├── .env.example               # Copy to .env (gitignored) and fill in real values
├── .gitignore
│
├── config/                    # Env loading & defaults
│   ├── config.go
│   └── config_test.go
│
├── database/                  # Firebase Admin SDK init (App, Firestore, Auth)
│   └── firebase.go
│
├── logger/                    # Async access-logger package
│   ├── entry.go               # LogEntry struct + Submitter interface
│   ├── config.go              # LoggerConfig + defaults
│   ├── logger.go              # Logger core, NewLogger, Submit, Close
│   ├── worker.go              # Worker pool & write path
│   ├── cleanup.go             # FIFO cleanup ticker
│   ├── batch.go               # Batched Firestore deletes
│   ├── config_test.go
│   └── logger_test.go
│
├── middleware/                # HTTP middlewares
│   ├── logging.go             # Logging(Submitter) -> http.Handler middleware
│   └── logging_test.go        # Unit tests (uses fakeSubmitter)
│
└── sandbox/                   # Dynamic mock-API handler
    └── handler.go             # CRUD handler for /sandbox/{projectId}/{table}/{id}
```

## Architecture

```
┌──────────┐   request   ┌──────────────────┐  Submit(entry)  ┌──────────────┐
│  client  │ ──────────► │ middleware.      │ ──────────────► │ logger.Logger│
└──────────┘             │ Logging(...)     │   (non-blocking)│  (interface) │
                         └──────────────────┘                 └──────┬───────┘
                                                                     │
                                                              workers (N)
                                                                     │
                                                                     ▼
                                                              Cloud Firestore
                                                                     ▲
                                                                     │
                         ┌──────────────────┐   direct read/write    │
                         │ sandbox.Handler  │ ───────────────────────┘
                         │ (CRUD on         │
                         │  sandbox/{pid}/  │
                         │  {table}/{id})   │
                         └──────────────────┘
```

Key design points:

- **`logger.Submitter` interface.** `middleware.Logging` depends on a
  `Submitter` (single-method: `Submit(*LogEntry) bool`), not the concrete
  `*logger.Logger`. This keeps the middleware trivially unit-testable with a
  fake, and lets future implementations (in-memory ring buffer, Kafka,
  OpenTelemetry) drop in without touching middleware code.
- **Defer-based latency measurement.** Latency is measured in a `defer`
  block so it captures panics in the downstream handler as well.
- **Drop-on-full semantics.** When the in-process channel is saturated, the
  entry is dropped and a warning is logged. The client response is **never**
  delayed by logging.
- **Project-scoped collections.** Sandbox data is namespaced under
  `sandbox/{projectId}/{table}/{id}`, so multiple projects can share a single
  Firestore instance without collision.
- **Compile-time check.** `var _ logger.Submitter = (*logger.Logger)(nil)`
  in the test file guarantees the concrete type satisfies the interface.

## HTTP routes

| Method | Path | Description |
|---|---|---|
| `GET` | `/hello` | Health probe — returns `Hello, Firebase World!` |
| `GET` | `/sandbox/{projectId}/{table}` | List all documents in a project table |
| `GET` | `/sandbox/{projectId}/{table}/{id}` | Fetch a single document |
| `POST` | `/sandbox/{projectId}/{table}` | Create a new document (auto-generated UUID) |
| `POST` | `/sandbox/{projectId}/{table}/{id}` | Create or overwrite a document with a known id |
| `PUT` | `/sandbox/{projectId}/{table}/{id}` | Partial update (Firestore `MergeAll`) |
| `PATCH` | `/sandbox/{projectId}/{table}/{id}` | Partial update (Firestore `MergeAll`) |
| `DELETE` | `/sandbox/{projectId}/{table}/{id}` | Delete a document |

All routes are wrapped with the access-logging middleware.

### Data shape

`POST`/`PUT`/`PATCH` accept any JSON object. The handler automatically
adds server-side metadata:

```json
{
  "_createdAt": "<Firestore ServerTimestamp>",
  "_createdBy": "anonymous",
  "_updatedAt": "<Firestore ServerTimestamp>",
  "_updatedBy": "anonymous",
  "...your fields": "..."
}
```

## Configuration

Copy the example env file and edit:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `GOOGLE_CLOUD_PROJECT` | `mockapi-sandbox-dev` | Firestore project ID |
| `GOOGLE_APPLICATION_CREDENTIALS` | _(unset)_ | Path to service-account JSON. If unset, ADC is used (Cloud Run / GCE). |
| `LOG_CHANNEL_BUFFER` | `1000` | Size of the in-process log channel |
| `LOG_NUM_WORKERS` | `10` | Number of worker goroutines |
| `LOG_CLEANUP_INTERVAL` | `5m` | How often the FIFO cleanup runs |
| `LOG_MAX_LOGS_PER_PROJECT` | `100` | Cap on logs retained per project |

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

Sandbox data appears at:

```
/projects/{GOOGLE_CLOUD_PROJECT}/sandbox/{projectId}/{table}/{id}
```

## Testing

The project ships with unit tests for all packages that have meaningful
in-process logic:

| Package | Coverage | What's tested |
|---|---|---|
| `config` | `config_test.go` | Default values, env override, invalid-fallback semantics |
| `logger` | `config_test.go`, `logger_test.go` | Config defaults, non-blocking submit, nil-drop, project tracking (thread-safe), close idempotency |
| `middleware` | `logging_test.go` | Method/path/projectID recording, status propagation, latency/timestamp, drop on full channel, entry on panic, one-entry-per-request |

`database/firebase.go` and `sandbox/handler.go` are intentionally not
unit-tested — both are thin wrappers around external services (Firebase Admin
SDK and live Firestore) and are exercised end-to-end against a real project
or the Firebase emulator.

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

- **Never commit `private-key.json`.** It is in `.gitignore` for this reason.
  Rotate any key that has been exposed.
- In production, prefer **Workload Identity** or ADC over service-account
  key files.
- The sandbox writes `_createdBy` / `_updatedBy` as the literal string
  `"anonymous"` — auth integration is a planned follow-up (the Firebase Auth
  client is already initialised in `main.go` but currently reserved for
  future use).
- See `.env.example` for the env contract.

### `.gitignore` coverage

The repository ships with a hardened `.gitignore` that protects against the
common leaks when running `go test` or `go build` locally:

| Pattern | Protects against |
|---|---|
| `private-key.json`, `*service-account*.json` | Firebase / GCP service-account keys |
| `/.env`, `.env.local`, `.env.*.local` | Local secrets |
| `*.pem`, `*.key`, `*.crt`, `*.p12` | Generic key material |
| `*.exe`, `*.exe~`, `/server.exe` | Compiled binaries & editor backups |
| `*.test`, `*.test.exe` | Test binaries |
| `coverage.out`, `coverage.html`, `coverage.txt`, `coverage.xml` | Coverage reports |
| `*.prof`, `*.trace`, `gocover/` | Profiling & native coverage (Go 1.20+) |
| `bin/`, `dist/`, `tmp/`, `scratch/` | Build output & scratch folders |
| `debug-logs/` | VS Code Copilot / agent debug logs |

Verify with `git check-ignore -v <file>` after `git init`.

## Make targets

| Target | Action |
|---|---|
| `make tidy` | `go mod tidy` |
| `make build` | Build `server.exe` |
| `make run` | Run from source (loads `.env` automatically) |
| `make test` | `go test -race ./...` |
| `make vet` | `go vet ./...` |
| `make fmt` | `gofmt -w .` |
| `make smoke` | Build, start, hit `/hello`, stop |
| `make clean` | Remove build artifacts |

## License

See [LICENSE](LICENSE) (add one for your project).
