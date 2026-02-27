# Go style guide

Conventions for Go code in this project. Follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments); this doc adds project-specific choices.

---

## Layout

- **cmd/** — Entrypoints. `cmd/api/main.go` runs the HTTP API. One directory per binary; keep main small (parse config, wire dependencies, start server).
- **internal/** — Private application code. Not importable by other modules. Packages: `internal/http` (router, middleware), `internal/auth`, `internal/users`, `internal/config`, `internal/infra` (AWS, DB clients). Add `internal/artists`, `internal/...` as features grow.
- **pkg/** — Optional: code that could be reused by another module. Prefer `internal/` unless we explicitly need to share.
- **No `src/`** — Standard Go layout; top-level package directories under repo root.

---

## Naming

- **Packages** — Short, lowercase, single word where possible: `http`, `auth`, `users`, `config`, `infra`. No `util` or `common` unless clearly scoped (e.g. `internal/httputil` for shared HTTP helpers).
- **Files** — Snake_case or lowercase: `handler.go`, `service.go`, `middleware.go`. Group by responsibility; one package per directory.
- **Interfaces** — Prefer small interfaces (one or few methods). Name by behaviour: `Reader`, `UserStore`. Suffix `-er` only when it reads well (e.g. `Handler`).
- **Errors** — Use `errors.New` or `fmt.Errorf`; wrap with `%w` when adding context. Sentinel errors: `var ErrNotFound = errors.New("not found")`. Export only errors that callers need to act on.
- **HTTP** — Handlers as methods on a struct (e.g. `*users.Handler`); `NewHandler(svc)` constructor. Route registration in `internal/http/router.go`; handler logic in domain package (e.g. `internal/users/handler.go`).

---

## Config and env

- **Source** — Environment variables. Use [envconfig](https://github.com/kelseyhightower/envconfig) with struct tags: `envconfig:"VAR_NAME" default:"value"`. No config files in repo for secrets.
- **Validation** — Required vars: `required:"true"`. Fail fast in `main` or in a `Load()`/`Validate()` step before starting the server.
- **Naming** — Env vars UPPER_SNAKE: `ADDR`, `JWT_SECRET`, `DYNAMO_TABLE`, `AWS_REGION`. Prefix shared vars if needed (e.g. `AFTERWAVE_`*) only to avoid clashes; otherwise keep short.

---

## Logging

- **Logger** — Use [slog](https://pkg.go.dev/log/slog). Default: JSON to stdout (e.g. `slog.New(slog.NewJSONHandler(os.Stdout, nil))`). Set default with `slog.SetDefault(logger)` so packages that don’t receive a logger still work.
- **Pass logger** — Prefer passing `*slog.Logger` into constructors (router, handlers, services) for testability and consistency. Avoid global logger in business logic where avoidable.
- **Structured** — Key-value: `logger.Info("message", "key", value, "key2", value2)`. Use consistent keys: `err`, `user_id`, `path`, `duration`. Include `trace_id`/`span_id` when available (e.g. from request context).
- **Levels** — `Info` for normal flow (e.g. request completed); `Error` for failures and panics; `Debug` for verbose dev-only. Avoid logging secrets or full request bodies.

---

## HTTP

- **Router** — stdlib `net/http` + `ServeMux`. Use Go 1.22+ route patterns (e.g. `POST /signup`, `GET /artists/{id}`). Middleware as `func(http.Handler) http.Handler`; chain in router with a small helper (e.g. `chain(h, mw1, mw2)`).
- **Middleware order** — Recoverer → RealIP → RequestLogger → auth (where needed). Apply per-route or per-group; keep global middleware minimal.
- **Handlers** — Return JSON for API: `w.Header().Set("Content-Type", "application/json")`; use `json.Encode` for response body. Error responses: consistent shape (e.g. `{"error": "message"}`) and appropriate status codes. Avoid exposing internal errors to the client.
- **Context** — Pass `r.Context()` into services and DB calls. Use context for timeouts and cancellation; store request-scoped values (e.g. user ID from JWT) in context only when needed downstream.

---

## Errors and status codes

- **Client errors** — 400 Bad Request (validation, malformed body, or generic signup failure), 401 Unauthorized (missing or invalid auth), 403 Forbidden (auth OK but not allowed), 404 Not Found, 409 Conflict (e.g. handle already in use).
- **Server errors** — 500 Internal Server Error on unexpected failures. Log the error with stack or context; do not return internal details in the response.
- **Errors in handlers** — Prefer a small set of sentinel or typed errors from the service; map them to HTTP status in the handler. Avoid `http.Error` with raw `err.Error()` for internal errors.

---

## Testing

We use **unit tests** for fast feedback on logic and handlers, and **API-level tests** that hit the real API with a real database and S3. Both are part of the standard workflow.

### Unit tests

- **Allowed and encouraged** — Unit tests live next to the code and run with `go test`. Use them for handlers, services, and business logic.
- **Table-driven tests** — Prefer for handlers and services: slice of `struct { name, input, want }`; loop and run. Use subtests: `t.Run(tt.name, func(t *testing.T) { ... })`.
- **Mocks** — Interface-based: define small interfaces for DB or external clients; inject real or fake in tests. No codegen for mocks unless we have many interfaces; hand-written fakes are fine.
- **HTTP tests (in-process)** — Use `httptest.ResponseRecorder` and `httptest.NewRequest`; call the handler or router. Test status code and body (or important fields). No real DB or S3; use fakes or mocks.
- **Place tests** — `*_test.go` next to the code. Package `foo` for black-box tests; `foo_test` with `import "foo"` for white-box. Prefer same package for unit tests.

**Run unit tests:** `go test ./...` (or `make test` if we add a target that does that).

### API-level tests (integration)

- **Full-stack tests** — A dedicated test suite that uses the API **like a real user**: real HTTP requests to the running API, talking to a **real database** (e.g. DynamoDB Local or a test table), **real S3** (e.g. LocalStack or a test bucket), and any other real dependencies (SES, etc.). No mocks in this suite; we test the full path from HTTP to persistence.
- **Purpose** — Catch integration bugs, auth flows, DB/S3 behaviour, and regressions that unit tests with mocks might miss. Run before merge or in CI.
- **How we run them** — **Docker Compose** brings up the required services (e.g. DynamoDB Local, LocalStack for S3, optional local API or test runner). A **Makefile** provides targets to start the stack and run the API-level tests (e.g. `make test-api` or `make integration`). Developers run `make test-api` (or equivalent) locally; CI runs the same via Docker Compose.
- **Scope** — Tests live in a dedicated package or directory (e.g. `tests/api/` or `internal/http/api_test` with build tag `integration`). They start (or assume) a running API and hit it over HTTP; they may seed the DB and assert on responses and side effects (e.g. objects in S3). Keep tests independent (clean up or use unique IDs) so they can run in any order.
- **Environment** — Use the same env (or a `.env.test`) that points at the Compose services (e.g. `DYNAMO_ENDPOINT=http://localhost:8000`, `S3_ENDPOINT=http://localhost:4566`). The Makefile can export these or pass them into the test runner.

**Run API-level tests:** `make test-api` (or `make integration`) — Makefile starts Docker Compose (if needed) and runs the integration test suite. Exact target names TBD; document in the repo README and in the Makefile.

---

## Dependencies

- **Stdlib first** — Prefer `net/http`, `encoding/json`, `context`, `slog`. Add dependencies only when they clearly simplify code or are standard in the ecosystem (e.g. AWS SDK, envconfig, JWT library).
- **Versioning** — Use Go modules; `go mod tidy`. Prefer a single major version per dependency; upgrade deliberately. No unused imports or modules.
- **Vendoring** — Optional: `go mod vendor` for reproducible builds in CI. Not required for day one; CI can use `go mod download` with checksum verification.

---

## Formatting and linting

- **Format** — `gofmt` or `go fmt`. No stylistic debates; tool decides.
- **Linting** — Run `go vet` in CI. Consider `staticcheck` or `golangci-lint` with a small rule set (e.g. `govet`, `errcheck`, `staticcheck`). Fix or explicitly ignore with comments; don’t disable globally.
- **Imports** — Group: stdlib, blank line, third-party, blank line, internal. Use `goimports` or your editor’s “organize imports” to maintain order.

---

## OpenTelemetry

- **Use** — We use OpenTelemetry for tracing (otelhttp, trace). Keep spans around HTTP and critical paths; avoid high-cardinality attributes. Export config (OTLP endpoint, sampling) via env or config; see Deployment.
- **Context** — Pass context through so trace IDs propagate. Middleware and handlers should not block on tracing; keep instrumentation low-overhead.

---

## Summary

- Layout: cmd/, internal/, small main, domain packages.
- Config: env only, envconfig, fail fast.
- Logging: slog, JSON, structured, pass logger.
- HTTP: stdlib, middleware chain, consistent errors and status codes.
- **Tests:** Unit tests (table-driven, interfaces, httptest) next to code; **API-level tests** hit the real API with real DB and S3, run via **Docker Compose** and **Makefile** (e.g. `make test-api`).
- Format: gofmt, go vet (and optionally staticcheck/golangci-lint).

