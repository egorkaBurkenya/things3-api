# Technical Context

## Language & Runtime
- **Go 1.21** (stdlib only, zero external dependencies)
- Single binary compilation: `go build -o things3-api .`

## Architecture
```
HTTP Request → Middleware Chain → Handler Router → AppleScript Wrapper → osascript → Things 3
```

### Middleware Chain (order matters)
1. Recovery (panic handler)
2. Logger (request logging via slog)
3. MaxBody (1MB limit)
4. Auth (Bearer token, skips /health)
5. Things3Check (503 if app not running, skips /health)

### Routing
Go 1.21 doesn't support `PathValue` or method-based mux patterns.
Each resource has a Router function (e.g., `TasksRouter`) that parses `r.URL.Path` and `r.Method` manually.

### AppleScript Integration
- Scripts are written to temp files, executed via `osascript`, then cleaned up
- Output is tab-delimited text, parsed into Go structs
- All user input is escaped via `EscapeString()` to prevent injection

## Configuration
- Environment variables + optional .env file
- `THINGS_API_TOKEN` (required) — Bearer auth token
- `THINGS_API_PORT` (default: 7420)
- `THINGS_API_HOST` (default: 127.0.0.1)
- `LOG_LEVEL` (default: info)

## Deployment
- macOS launchd service (`com.things3api.plist`)
- Binary installed to `/usr/local/bin/things3-api`
- Logs: `/tmp/things3-api.log`, `/tmp/things3-api.err`
