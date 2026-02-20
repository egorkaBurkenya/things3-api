# Claude Code Prompt: things3-api

Paste this entire prompt into Claude Code on your Mac.

---

Build a production-ready local REST API server that bridges Things 3 (Mac app) with external HTTP clients via AppleScript.

## Project

**Name:** `things3-api`  
**Language:** Go (single binary, no runtime dependencies)  
**Port:** 7420 (configurable via env)  
**Auth:** Bearer token (env var `THINGS_API_TOKEN`)  
**Goal:** Allow external services (AI assistants, scripts) to read and write Things 3 data securely over HTTP.

## Repository Structure

```
things3-api/
├── main.go
├── applescript/
│   ├── tasks.go       — AppleScript wrappers for tasks
│   ├── projects.go    — AppleScript wrappers for projects/areas
│   └── runner.go      — osascript execution helper
├── handlers/
│   ├── tasks.go
│   ├── projects.go
│   ├── areas.go
│   └── health.go
├── middleware/
│   └── auth.go        — Bearer token validation
├── models/
│   └── types.go       — Task, Project, Area structs
├── config/
│   └── config.go      — Config from env/.env file
├── launchd/
│   └── com.things3api.plist  — macOS launchd service
├── .env.example
├── Makefile
└── README.md
```

## API Endpoints

### Health
```
GET /health
→ 200 {"status":"ok","things3":"running"|"not_running"}
```

### Tasks — Read
```
GET /tasks/inbox
→ all tasks in Inbox

GET /tasks/today
→ all tasks in Today view (scheduled for today + evening)

GET /tasks/upcoming
→ tasks in Upcoming view

GET /tasks/anytime
→ tasks in Anytime view

GET /tasks/someday
→ tasks in Someday view

GET /tasks?project=<name>&area=<name>&tag=<tag>
→ tasks filtered by project name, area name, or tag (all optional)

GET /tasks/:id
→ single task by Things ID
```

### Tasks — Write
```
POST /tasks
Body: {
  "title": "string (required)",
  "notes": "string",
  "project": "project name",
  "area": "area name",
  "due": "2026-02-20",          // deadline date ISO 8601
  "when": "today|evening|tomorrow|someday|anytime|2026-02-21",
  "tags": ["tag1", "tag2"]
}
→ 201 {"id":"things-id","title":"..."}

PATCH /tasks/:id
Body: any subset of POST fields (only provided fields are updated)
→ 200 updated task

POST /tasks/:id/complete
→ 200 {"ok":true}

POST /tasks/:id/cancel
→ 200 {"ok":true}

DELETE /tasks/:id
→ moves task to Trash
→ 200 {"ok":true}
```

### Projects
```
GET /projects
→ all projects with their area, task count

GET /projects/:id
→ single project

POST /projects
Body: {
  "name": "string (required)",
  "area": "area name",
  "notes": "string",
  "when": "today|someday|anytime|2026-02-21"
}
→ 201 {"id":"...","name":"..."}

PATCH /projects/:id
Body: {"name":"...","notes":"...","area":"..."}
→ 200 updated project

POST /projects/:id/complete
→ 200 {"ok":true}
```

### Areas
```
GET /areas
→ all areas with project list

GET /areas/:id
→ single area

POST /areas
Body: {"name":"string (required)"}
→ 201 {"id":"...","name":"..."}

PATCH /areas/:id
Body: {"name":"..."}
→ 200 updated area
```

## Data Models

```go
type Task struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Notes     string    `json:"notes,omitempty"`
    Status    string    `json:"status"` // open, completed, cancelled
    Project   string    `json:"project,omitempty"`
    Area      string    `json:"area,omitempty"`
    Tags      []string  `json:"tags,omitempty"`
    Due       string    `json:"due,omitempty"`       // ISO date
    When      string    `json:"when,omitempty"`
    CreatedAt string    `json:"created_at,omitempty"`
}

type Project struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Area  string `json:"area,omitempty"`
    Notes string `json:"notes,omitempty"`
}

type Area struct {
    ID       string    `json:"id"`
    Name     string    `json:"name"`
    Projects []Project `json:"projects,omitempty"`
}
```

## AppleScript Examples

Use these as reference. Execute via `osascript -e '...'` or temp `.scpt` files for multi-line scripts.

```applescript
-- Get all tasks in Today
tell application "Things3"
    set todayTasks to to dos of list "Today"
    set result to {}
    repeat with t in todayTasks
        set end of result to {id of t, name of t, notes of t}
    end repeat
    return result
end tell

-- Get tasks in Inbox
tell application "Things3"
    to dos of list "Inbox"
end tell

-- Create a task
tell application "Things3"
    make new to do with properties {name:"Task title", notes:"Notes here"}
end tell

-- Create task in a project
tell application "Things3"
    set proj to first project whose name is "MyProject"
    make new to do in proj with properties {name:"Task title"}
end tell

-- Complete a task by ID
tell application "Things3"
    set t to first to do whose id is "THINGS-ID-HERE"
    set status of t to completed
end tell

-- Update task title and notes
tell application "Things3"
    set t to first to do whose id is "THINGS-ID-HERE"
    set name of t to "New title"
    set notes of t to "New notes"
end tell

-- Create project
tell application "Things3"
    make new project with properties {name:"New Project"}
end tell

-- Create area
tell application "Things3"
    make new area with properties {name:"New Area"}
end tell

-- Get all projects
tell application "Things3"
    projects
end tell

-- Get all areas
tell application "Things3"
    areas
end tell

-- Set due date
tell application "Things3"
    set t to first to do whose id is "ID"
    set due date of t to date "20/02/2026"
end tell
```

## AppleScript Runner

For multi-line scripts, write to a temp file and execute:

```go
func RunAppleScript(script string) (string, error) {
    f, _ := os.CreateTemp("", "things3-*.scpt")
    f.WriteString(script)
    f.Close()
    defer os.Remove(f.Name())
    out, err := exec.Command("osascript", f.Name()).Output()
    return strings.TrimSpace(string(out)), err
}
```

## Configuration (.env)

```env
THINGS_API_TOKEN=your-secret-token-here
THINGS_API_PORT=7420
THINGS_API_HOST=127.0.0.1   # use 0.0.0.0 if behind Tailscale
LOG_LEVEL=info
```

## Auth Middleware

```go
// Every request must include:
// Authorization: Bearer <THINGS_API_TOKEN>
// Return 401 if missing or wrong
```

## launchd Service (com.things3api.plist)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.things3api</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/things3-api</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>THINGS_API_TOKEN</key>
        <string>REPLACE_WITH_TOKEN</string>
        <key>THINGS_API_PORT</key>
        <string>7420</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/things3-api.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/things3-api.err</string>
</dict>
</plist>
```

Install with:
```bash
cp launchd/com.things3api.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.things3api.plist
```

## Makefile

```makefile
build:
	go build -o things3-api .

install: build
	cp things3-api /usr/local/bin/things3-api
	cp launchd/com.things3api.plist ~/Library/LaunchAgents/
	launchctl load ~/Library/LaunchAgents/com.things3api.plist

uninstall:
	launchctl unload ~/Library/LaunchAgents/com.things3api.plist
	rm -f ~/Library/LaunchAgents/com.things3api.plist
	rm -f /usr/local/bin/things3-api

restart:
	launchctl unload ~/Library/LaunchAgents/com.things3api.plist
	launchctl load ~/Library/LaunchAgents/com.things3api.plist

logs:
	tail -f /tmp/things3-api.log
```

## README Requirements

Include:
1. What it does (one paragraph)
2. Prerequisites: macOS + Things 3 installed
3. Installation steps
4. Configuration (.env setup)
5. Tailscale sharing setup (expose to remote machines securely)
6. API reference (all endpoints with curl examples)
7. Contributing section

## Error Handling

- `401 Unauthorized` — missing/wrong token
- `404 Not Found` — task/project/area not found
- `500 Internal Server Error` — AppleScript failed (include error message)
- `503 Service Unavailable` — Things 3 not running

Check if Things 3 is running before every request. If not, return 503 with helpful message.

## Notes

- Things 3 IDs look like: `5B6E4E12-3A2B-4C1D-8E9F-0A1B2C3D4E5F`
- AppleScript is synchronous — no concurrency issues
- Test every endpoint manually after implementation
- Add request logging (method, path, status, duration)
- The `when` field maps to Things scheduling: "today" = Today list, "evening" = This Evening, "someday" = Someday, specific date = scheduled for that date
