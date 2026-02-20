# things3-api

A local REST API server that brings HTTP access to [Things 3](https://culturedcode.com/things/) on macOS. It communicates with the Things 3 desktop app through AppleScript, exposing endpoints to manage tasks, projects, and areas. All requests (except the health check) require Bearer token authentication. The server compiles to a single Go binary with zero external dependencies.

## Prerequisites

- **macOS** -- AppleScript is used to communicate with Things 3, so this only runs on a Mac.
- **Things 3** -- must be installed and running on the same machine.
- **Go 1.21+** -- required to build the binary from source.

## Quick Start

```bash
# Clone the repository
git clone https://github.com/egorkaBurkenya/things3-api.git
cd things3-api

# Build the binary
make build

# Generate an API token
make token

# Create a .env file
cat > .env <<EOF
THINGS_API_TOKEN=<paste-your-generated-token>
THINGS_API_PORT=7420
THINGS_API_HOST=127.0.0.1
LOG_LEVEL=info
EOF

# Run the server
./things3-api
```

The API is now available at `http://localhost:7420`.

## Installation

### Build from source

```bash
make build
```

This produces a `things3-api` binary in the project directory.

### Install as a launchd service

The included launchd plist keeps the server running in the background and restarts it automatically on login.

```bash
make install
```

This will:

1. Build the binary.
2. Copy it to `/usr/local/bin/things3-api`.
3. Install the launchd plist to `~/Library/LaunchAgents/com.things3api.plist`.
4. Load the service immediately.

Before running `make install`, edit `launchd/com.things3api.plist` and replace `your-secret-token-here` with your actual token:

```xml
<key>THINGS_API_TOKEN</key>
<string>your-actual-token</string>
```

### Other Makefile targets

| Command          | Description                          |
|------------------|--------------------------------------|
| `make run`       | Run the server with `go run`         |
| `make build`     | Compile the binary                   |
| `make install`   | Build, install binary, load service  |
| `make uninstall` | Stop service, remove binary and plist|
| `make restart`   | Restart the launchd service          |
| `make logs`      | Tail the service log file            |
| `make clean`     | Remove the compiled binary           |
| `make token`     | Generate a random 32-byte hex token  |

## Configuration

The server reads configuration from environment variables. A `.env` file in the working directory is also supported.

| Variable           | Default     | Description                              |
|--------------------|-------------|------------------------------------------|
| `THINGS_API_TOKEN` | *(required)* | Bearer token for authenticating requests |
| `THINGS_API_PORT`  | `7420`      | Port the server listens on               |
| `THINGS_API_HOST`  | `127.0.0.1` | Host/IP the server binds to              |
| `LOG_LEVEL`        | `info`      | Log level (`info` or `debug`)            |

### Generating a token

```bash
openssl rand -hex 32
```

Or use the Makefile shortcut:

```bash
make token
```

### Example .env file

```
THINGS_API_TOKEN=a1b2c3d4e5f6...
THINGS_API_PORT=7420
THINGS_API_HOST=127.0.0.1
LOG_LEVEL=info
```

Environment variables set in the shell take precedence over values in `.env`.

## Tailscale Setup

By default the server binds to `127.0.0.1`, which means it only accepts connections from the local machine. To expose the API to other devices on your [Tailscale](https://tailscale.com) tailnet, change the bind address.

### Option 1: Bind to all interfaces

```
THINGS_API_HOST=0.0.0.0
```

This allows connections from any network interface, including your Tailscale IP. Make sure your macOS firewall or other network policies restrict access to trusted networks.

### Option 2: Bind to your Tailscale IP

Find your Tailscale IP:

```bash
tailscale ip -4
```

Then set the host to that address:

```
THINGS_API_HOST=100.x.y.z
```

This restricts the server to only accept connections through the Tailscale interface.

### Accessing from other machines

From any other device on your tailnet:

```bash
export TOKEN="your-token"
curl -H "Authorization: Bearer $TOKEN" http://100.x.y.z:7420/health
```

Tailscale provides encrypted, authenticated connections between your devices, so the API traffic is protected in transit across the tailnet. You still need the Bearer token for API authentication. No port forwarding or firewall changes are required.

## API Reference

All endpoints return JSON. All endpoints except `/health` require a Bearer token in the `Authorization` header.

The examples below assume:

```bash
export TOKEN="your-token"
```

---

### Health

#### GET /health

Check server status and whether Things 3 is running. No authentication required.

```bash
curl http://localhost:7420/health
```

Response:

```json
{
  "status": "ok",
  "things3": "running"
}
```

If Things 3 is not open, `things3` will be `"not_running"`.

---

### Tasks

#### GET /tasks/inbox

List all tasks in the Inbox.

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:7420/tasks/inbox
```

#### GET /tasks/today

List all tasks scheduled for Today.

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:7420/tasks/today
```

#### GET /tasks/upcoming

List all tasks in the Upcoming list.

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:7420/tasks/upcoming
```

#### GET /tasks/anytime

List all tasks in the Anytime list.

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:7420/tasks/anytime
```

#### GET /tasks/someday

List all tasks in the Someday list.

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:7420/tasks/someday
```

#### GET /tasks?project=X&area=Y&tag=Z

List tasks filtered by project name, area name, or tag. At least one filter parameter is required.

```bash
# Filter by project
curl -H "Authorization: Bearer $TOKEN" "http://localhost:7420/tasks?project=Website%20Redesign"

# Filter by area
curl -H "Authorization: Bearer $TOKEN" "http://localhost:7420/tasks?area=Work"

# Filter by tag
curl -H "Authorization: Bearer $TOKEN" "http://localhost:7420/tasks?tag=urgent"
```

#### GET /tasks/:id

Get a single task by its Things 3 ID.

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:7420/tasks/ABC-123-DEF
```

Response:

```json
{
  "id": "ABC-123-DEF",
  "title": "Review pull request",
  "notes": "Check the API changes",
  "status": "open",
  "project": "Website Redesign",
  "area": "Work",
  "tags": ["urgent", "dev"],
  "due": "2026-03-01",
  "created_at": "2026-02-20"
}
```

#### POST /tasks

Create a new task.

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Buy groceries",
    "notes": "Milk, eggs, bread",
    "project": "Errands",
    "due": "2026-02-25",
    "when": "today",
    "tags": ["shopping"]
  }' \
  http://localhost:7420/tasks
```

| Field     | Type       | Required | Description                                                                 |
|-----------|------------|----------|-----------------------------------------------------------------------------|
| `title`   | string     | Yes      | Task title (max 1000 characters)                                            |
| `notes`   | string     | No       | Task notes (max 10000 characters)                                           |
| `project` | string     | No       | Project name to assign the task to                                          |
| `area`    | string     | No       | Area name (ignored if `project` is set)                                     |
| `due`     | string     | No       | Due date in `YYYY-MM-DD` format                                             |
| `when`    | string     | No       | Schedule: `today`, `evening`, `tomorrow`, `someday`, `anytime`, or `YYYY-MM-DD` |
| `tags`    | string[]   | No       | List of tag names                                                           |

Returns the created task with status `201 Created`.

#### PATCH /tasks/:id

Update an existing task. Only include the fields you want to change.

```bash
curl -X PATCH -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Buy groceries and snacks",
    "due": "2026-02-28",
    "tags": ["shopping", "priority"]
  }' \
  http://localhost:7420/tasks/ABC-123-DEF
```

All fields are optional. Set `due` or `when` to an empty string to clear them. Set `project` to an empty string to move the task to the Inbox.

#### POST /tasks/:id/complete

Mark a task as completed.

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:7420/tasks/ABC-123-DEF/complete
```

Response:

```json
{
  "ok": true
}
```

#### POST /tasks/:id/cancel

Mark a task as canceled.

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:7420/tasks/ABC-123-DEF/cancel
```

#### DELETE /tasks/:id

Move a task to the Trash.

```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://localhost:7420/tasks/ABC-123-DEF
```

---

### Projects

#### GET /projects

List all projects.

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:7420/projects
```

Response:

```json
[
  {
    "id": "PRJ-456",
    "name": "Website Redesign",
    "notes": "Q2 initiative",
    "area": "Work",
    "task_count": 12
  }
]
```

#### GET /projects/:id

Get a single project by ID.

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:7420/projects/PRJ-456
```

#### POST /projects

Create a new project.

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Q3 Planning",
    "notes": "Strategic planning for next quarter",
    "area": "Work",
    "when": "today"
  }' \
  http://localhost:7420/projects
```

| Field   | Type   | Required | Description                                                         |
|---------|--------|----------|---------------------------------------------------------------------|
| `name`  | string | Yes      | Project name (max 500 characters)                                   |
| `notes` | string | No       | Project notes                                                       |
| `area`  | string | No       | Area name to assign the project to                                  |
| `when`  | string | No       | Schedule: `today`, `someday`, `anytime`, or `YYYY-MM-DD`            |

Returns the created project with status `201 Created`.

#### PATCH /projects/:id

Update an existing project. Only include the fields you want to change.

```bash
curl -X PATCH -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Q3 Planning (Updated)",
    "area": "Strategy"
  }' \
  http://localhost:7420/projects/PRJ-456
```

| Field   | Type   | Description                                          |
|---------|--------|------------------------------------------------------|
| `name`  | string | New project name (max 500 characters)                |
| `notes` | string | New project notes                                    |
| `area`  | string | Area name (empty string to remove area assignment)   |

#### POST /projects/:id/complete

Mark a project as completed.

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:7420/projects/PRJ-456/complete
```

---

### Areas

#### GET /areas

List all areas.

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:7420/areas
```

Response:

```json
[
  {
    "id": "AREA-789",
    "name": "Work"
  }
]
```

#### GET /areas/:id

Get a single area by ID, including its projects.

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:7420/areas/AREA-789
```

Response:

```json
{
  "id": "AREA-789",
  "name": "Work",
  "projects": [
    {
      "id": "PRJ-456",
      "name": "Website Redesign",
      "notes": "Q2 initiative",
      "task_count": 12
    }
  ]
}
```

#### POST /areas

Create a new area.

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Personal"
  }' \
  http://localhost:7420/areas
```

| Field  | Type   | Required | Description                      |
|--------|--------|----------|----------------------------------|
| `name` | string | Yes      | Area name (max 500 characters)   |

Returns the created area with status `201 Created`.

#### PATCH /areas/:id

Update an existing area.

```bash
curl -X PATCH -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Personal Life"
  }' \
  http://localhost:7420/areas/AREA-789
```

| Field  | Type   | Description                       |
|--------|--------|-----------------------------------|
| `name` | string | New area name (max 500 characters)|

---

## Error Codes

All errors are returned as JSON with an `error` field.

```json
{
  "error": "description of the problem"
}
```

| Status Code | Meaning                | When it occurs                                        |
|-------------|------------------------|-------------------------------------------------------|
| 400         | Bad Request            | Invalid request body, missing required fields, or validation failure |
| 401         | Unauthorized           | Missing or invalid Bearer token                       |
| 404         | Not Found              | Resource does not exist or unknown endpoint            |
| 405         | Method Not Allowed     | HTTP method not supported for the endpoint            |
| 500         | Internal Server Error  | Unexpected server error or AppleScript failure        |
| 503         | Service Unavailable    | Things 3 is not running on this Mac                   |

The `503` response includes an additional `message` field:

```json
{
  "error": "Things 3 is not running",
  "message": "Please open Things 3 on this Mac"
}
```

## Security

- **Localhost-only by default.** The server binds to `127.0.0.1`, rejecting connections from external networks unless explicitly configured otherwise.
- **Bearer token authentication.** Every request (except `/health`) must include a valid token in the `Authorization` header. Token comparison uses constant-time comparison to prevent timing attacks.
- **AppleScript injection prevention.** All user-supplied strings are escaped before being embedded in AppleScript commands, preventing script injection.
- **Request body size limits.** Request bodies are capped at 1 MB to prevent abuse. Field-level limits are also enforced (e.g., 1000 characters for task titles, 10000 for notes).
- **Input validation.** IDs are validated against a strict alphanumeric pattern. Dates must be valid ISO 8601 format. Enum values (like `when`) are checked against an allowed list.

## Contributing

Contributions are welcome. To get started:

1. Fork the repository.
2. Create a feature branch: `git checkout -b my-feature`.
3. Make your changes and ensure the code compiles: `make build`.
4. Commit your changes with clear, descriptive messages.
5. Push to your fork and open a pull request.

Please keep pull requests focused on a single change. If you are fixing a bug, include a description of the issue and how to reproduce it. If you are adding a feature, explain the use case.

For bug reports and feature requests, open an issue on GitHub.

## License

This project is licensed under the [MIT License](LICENSE).
