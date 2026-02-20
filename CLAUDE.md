# things3-api

Local REST API server bridging Things 3 (macOS) with HTTP clients via AppleScript.

## Build & Run

```bash
go build -o things3-api .          # build binary
THINGS_API_TOKEN=xxx ./things3-api # run with token
make build                         # same via Makefile
make install                       # install as launchd service
```

## Project Structure

```
main.go              — entry point, router setup, middleware chain
config/config.go     — env/dotenv config loading
models/types.go      — data structs (Task, Project, Area) + request validation
applescript/runner.go — osascript execution + string escaping (security-critical)
applescript/tasks.go  — task CRUD via AppleScript
applescript/projects.go — project/area CRUD via AppleScript
middleware/auth.go   — Bearer auth, logging, recovery, body limit, Things3 check
handlers/tasks.go    — task HTTP handlers + routing
handlers/projects.go — project HTTP handlers + routing
handlers/areas.go    — area HTTP handlers + routing
handlers/health.go   — health check endpoint
handlers/response.go — JSON response helpers + path extraction
```

## Key Conventions

- **Go 1.21** — no PathValue; routing is manual in handler routers
- **Zero dependencies** — only Go stdlib
- **AppleScript security** — ALWAYS use `applescript.EscapeString()` for user input before embedding in scripts
- **ID validation** — use `models.ValidateThingsID()` before any AppleScript call with user-supplied IDs
- **Auth** — `crypto/subtle.ConstantTimeCompare` for token comparison
- **Host** — defaults to 127.0.0.1 (localhost-only)
- **Port** — 7420 default

## API Routing Pattern

Each resource has a Router function (e.g., `TasksRouter`) registered on `/tasks` and `/tasks/`.
The router parses `r.URL.Path` and `r.Method` to dispatch to the correct handler.
Path parameters are extracted with `extractID()` and `pathSuffix()` from `handlers/response.go`.

## Security Rules

1. Never interpolate user strings into AppleScript without `EscapeString()`
2. Always validate Things IDs with `ValidateThingsID()` before use
3. Never expose internal errors to clients in production — use generic messages
4. Keep `.env` out of git (in .gitignore)
5. Request body limit: 1MB via `MaxBody` middleware
