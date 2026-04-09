# graphdo — Project Instructions

## What This Is

A Go CLI (`graphdo`) that interacts with Microsoft Graph to send emails to yourself and manage Microsoft To Do items. Designed for AI agent automation (Claude Cowork, GitHub Copilot) with JSON stdout / human stderr separation. The target audience is non-technical users.

Repository: `github.com/co-native-ab/graphdo`

## Architecture

```
main.go                  Entry point, go:embed SKILL.md, arg parsing, logger setup
internal/
  cmd/
    args.go              CLI structs via go-arg (subcommands as pointer fields)
    handlers.go          Dependencies struct, Dispatch(), all command handlers
  auth/
    auth.go              Authenticator interface + Browser/DeviceCode/Static impls
    tokencache.go        MSAL file-based cache (ExportReplace) + account persistence
  config/
    config.go            Config struct, Load/Save (atomic via temp+rename), Dir(), Validate()
  graph/
    client.go            Lightweight HTTP client (no Graph SDK), APIError, do() helper
    mail.go              User, GetMe, SendMail
    todo.go              TodoItem, TodoList, CRUD + pagination ($top/$skip)
  testutil/
    mockgraph.go         In-memory mock Graph API server (httptest.Server)
main_test.go             Integration tests via testEnv helper
SKILL.md                 Agent skill file (embedded in binary via go:embed)
```

## Key Design Decisions

### No Graph SDK
We use `net/http` directly instead of the Microsoft Graph SDK. The SDK is bloated. Our `graph.Client` wraps `http.Client` with a `TokenSource` interface for lazy token acquisition, JSON encode/decode, and structured error handling. All Graph API interaction goes through `client.do()`, which calls `TokenSource.Token(ctx)` per-request. The client is created once in `main.go` and passed via `Dependencies.GraphClient`.

### MSAL Direct (Not azidentity)
`azidentity.Cache` is an internal alias and can't be used externally. We use MSAL's `public.Client` directly with `cache.ExportReplace` for file-based token persistence. Two files in the config dir: `msal_cache.json` (token cache) and `account.json` (saved account for silent refresh).

### CLI Framework: go-arg
Chosen for being slim and idiomatic. Subcommands are pointer fields — non-nil means selected. `go-arg` handles parsing; `Dispatch()` routes to handlers. We use `Version()` and `Description()` methods on `Args`.

### Interactive UI: charmbracelet/huh
Used for `config` (todo list picker) and `skill install` (target picker). Cannot be tested headless — tests use flags to bypass interactive prompts.

### Client ID Resolution Chain
`--client-id` flag (login only) → config file → `DefaultClientID` constant. Login saves the resolved ID to config. Other commands read from config. `runConfigSelect` preserves existing client ID when saving.

### Output Convention
- **stdout**: Machine-readable JSON only (for AI agents to parse)
- **stderr**: Human-friendly messages (✓ success, errors, debug logs)
- All JSON output uses `json.NewEncoder` with 2-space indent

### Scopes
`Mail.Send`, `Tasks.ReadWrite`, `User.Read`, `offline_access`

## Go Style Rules

- **Early returns** — check errors and return immediately, don't nest
- **Error context** — always wrap: `fmt.Errorf("doing X: %w", err)`
- **Logging** — `log/slog` (Debug for verbose, Info for important events). `--debug` enables Debug level; default is Warn
- **File cleanup** — `defer func() { _ = f.Close() }()` everywhere; normalize paths with `filepath.Clean`/`filepath.Join`
- **Minimal dependencies** — prefer stdlib. Only third-party libs: go-arg, huh, MSAL
- **Comments** — only where clarification is needed, not on every function

## Testing

### Test Architecture
Tests live in `main_test.go` as integration tests. Each test creates a `testEnv` with:
- `testutil.MockState` — thread-safe in-memory Graph API state
- `testutil.NewMockGraphServer` — httptest.Server with Go 1.22+ method routing
- `t.TempDir()` for config directory
- `bytes.Buffer` for captured stdout/stderr

Standard test flags injected by `testEnv.runCLI()`:
```
--access-token test-token --graph-url <mock-url> --config-dir <temp-dir>
```

The `runWithIO()` function mirrors `run()` from main.go but with injected IO.

### Running Tests
```bash
go test ./... -race -count=1          # All tests with race detector
go test -run TestTodoCreate -v        # Single test verbose
```

### Linting
```bash
golangci-lint run ./...               # Uses .golangci.yml (v2 format)
```

The golangci-lint config uses **version: "2"** syntax. Key linters: bodyclose, errcheck, errorlint, gocritic, govet, revive, staticcheck. Formatters: gofmt, goimports (with local prefix `github.com/co-native-ab/graphdo`).

### Adding New Tests
1. Use `newTestEnv(t)` — pre-seeds a user, one todo list ("list-1"), and one task
2. Call `env.seedConfig(t)` if the command needs a configured todo list
3. Run command with `env.runCLI(t, "subcommand", "--flag", "value")`
4. Assert against `env.stdout.Bytes()` (JSON) and `env.stderr.String()` (messages)
5. Check mock state via `env.state.GetSentMails()` or `env.state.GetTodos("list-1")`

### Adding New Mock Endpoints
Add handlers to `NewMockGraphServer()` in `internal/testutil/mockgraph.go`. Use Go 1.22+ method routing: `mux.HandleFunc("METHOD /path/{param}", ...)`. Always call `checkAuth()` first and use `writeJSON()` for responses.

## Adding New Commands

1. **Define structs** in `internal/cmd/args.go` — add `*NewCmd` pointer field to parent struct
2. **Add dispatch case** in `handlers.go` `Dispatch()` switch
3. **Implement handler** in `handlers.go` — follow the pattern: get config → get client → call Graph → encode JSON to stdout
4. **Add tests** in `main_test.go`
5. **Update SKILL.md** (embedded in binary) and `README.md`
6. Run `go test ./... -race -count=1 && golangci-lint run ./...`

## CI/CD

### CI (`ci.yml`)
Runs on push/PR to main: lint (golangci-lint) → check (vet + gofmt) → test → build

### Release (`release.yml`)
Triggered by `v*` tags. Cross-compiles 6 targets (linux/darwin/windows × amd64/arm64). Bare binaries named `graphdo-{os}-{arch}`. Version injected via ldflags: `-X main.version=${{ github.ref_name }}`.

## Config & Auth Files

All stored in the config directory (`~/.config/graphdo` on Linux, OS-appropriate elsewhere):
- `config.json` — client ID, todo list ID/name
- `msal_cache.json` — MSAL token cache
- `account.json` — saved MSAL account for silent token refresh

The `--config-dir` flag overrides the directory (used in tests with `t.TempDir()`).

## Graph API Patterns

- Collections wrapped in `{"value": [...]}` — decoded with `graphListResponse[T]`
- Pagination via `$top` and `$skip` query params
- `POST /me/sendMail` returns HTTP 202 with empty body
- `PATCH` supports partial updates (use `omitempty` tags)
- Errors in `{"error": {"code": "...", "message": "..."}}` → parsed into `APIError`

## SKILL.md

Embedded in the binary via `//go:embed SKILL.md` in main.go. Follows the skill-creator format: YAML frontmatter with `name` + `description`, imperative style, under 500 lines. The `graphdo skill install` command writes it to various agent-specific locations.
