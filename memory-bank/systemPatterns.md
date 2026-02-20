# System Patterns

## Code Organization
```
main.go           — entry point, wiring
config/           — configuration loading
models/           — data types + validation
applescript/      — Things 3 interaction layer
middleware/       — HTTP middleware chain
handlers/         — HTTP request handlers
```

## Handler Pattern
Each resource has a single Router function registered on both `/resource` and `/resource/`:
```go
func ResourceRouter(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    switch {
    case path == "/resource":
        // list or create based on r.Method
    default:
        id := extractID(path, "/resource/")
        suffix := pathSuffix(path, "/resource/")
        // dispatch to handler based on suffix and method
    }
}
```

## AppleScript Pattern
1. Build script string with escaped user inputs
2. Write to temp file
3. Execute via `osascript`
4. Parse tab-delimited output
5. Return Go structs

## Security Pattern
- Input validation at model level (Validate() methods)
- ID validation via regex (`[A-Za-z0-9\-]+`)
- String escaping at AppleScript level
- Constant-time token comparison
- Request body size limits

## Error Handling
- Handlers check `isNotFound(err)` for 404 responses
- AppleScript errors bubble up as 500
- Things 3 not running → 503 via middleware
- Validation errors → 400
