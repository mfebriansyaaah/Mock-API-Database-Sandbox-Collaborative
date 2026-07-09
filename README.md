# Mock API & Database Sandbox Collaborative

Asynchronous, non-blocking HTTP request logger for a Google Cloud Firestore
backend. Built with Go 1.22, designed for high-concurrency SaaS workloads.

## Features

- **Non-blocking logging.** HTTP handlers hand off log entries to a bounded
  channel; a worker pool writes them to Firestore in the background.
- **Bounded resources.** Channel buffer, worker count, and write timeouts
  are all configurable; channel-full drops are counted, never block the
  client response.
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
├── database/                  # Firebase Admin SDK init
│   └── firebase.go
│
├── logger/                    # Async logger package
│   ├── entry.go               # LogEntry struct + Submitter interface
│   ├── config.go              # LoggerConfig + defaults
│   ├── logger.go              # Logger core, NewLogger, Submit, Close
│   ├── worker.go              # Worker pool & write path
│   ├── cleanup.go             # FIFO cleanup ticker
│   ├── batch.go               # Batched Firestore deletes
│   ├── config_test.go
│   └── logger_test.go
│
└── middleware/                # HTTP middlewares
    ├── logging.go             # Logging(Submitter) -> http.Handler middleware
    └── logging_test.go        # Unit tests (uses fakeSubmitter)
```

## Architecture

The HTTP path is intentionally one-directional and decoupled:

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
```

Key design points:

- **`logger.Submitter` interface.** `middleware.Logging` depends on a
  `Submitter` (single-method: `Submit(*LogEntry) bool`), not the concrete
  `*logger.Logger`. This keeps the middleware trivially unit-testable
  with a fake, and lets future implementations (in-memory ring buffer,
  Kafka, OpenTelemetry) drop in without touching middleware code.
- **Defer-based latency measurement.** Latency is measured in a `defer`
  block so it captures panics in the downstream handler as well.
- **Drop-on-full semantics.** When the in-process channel is saturated,
  the entry is dropped and a warning is logged. The client response is
  **never** delayed by logging.
- **Compile-time check.** `var _ logger.Submitter = (*logger.Logger)(nil)`
  in the test file guarantees the concrete type satisfies the interface.

## Configuration

Copy the example env file and edit:

````bash
cp .env.example .env
````

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

````bash
# Run tests
make test

# Build and run
make run

# Or build a binary
make build
./server.exe
````

Smoke test the running server:

````bash
curl http://localhost:8080/hello
# -> Hello, Firebase World!
````

Logs appear in Firestore at:

```
/projects/{GOOGLE_CLOUD_PROJECT}/logs/{logId}
```

## Testing

The project ships with unit tests for all packages that have meaningful
in-process logic:

| Package | Coverage | What's tested |
|---|---|---|
| `config` | `config_test.go` | Default values, env override, invalid-fallback semantics |
| `logger` | `config_test.go`, `logger_test.go` | Config defaults, non-blocking submit, nil-drop, project tracking (thread-safe), close idempotency |
| `middleware` | `logging_test.go` | Method/path/projectID recording, status propagation, latency/timestamp, drop on full channel, entry on panic, one-entry-per-request |

`database/firebase.go` is intentionally not unit-tested — it is a thin
wrapper around the Firebase Admin SDK and is exercised in integration via
a real Firestore project (or the Firebase emulator).

````bash
# Run all tests
make test

# Verbose output (per-test)
go test ./... -v

# With coverage
go test ./... -cover
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
````

## Security

- **Never commit `private-key.json`.** It is in `.gitignore` for this reason.
  Rotate any key that has been exposed.
- In production, prefer **Workload Identity** or ADC over service-account
  key files.
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
| `make run` | Run from source |
| `make test` | `go test -race ./...` |
| `make vet` | `go vet ./...` |
| `make fmt` | `gofmt -w .` |
| `make smoke` | Build, start, hit `/hello`, stop |
| `make clean` | Remove build artifacts |

## License

See [LICENSE](LICENSE) (add one for your project).
