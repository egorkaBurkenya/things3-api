# Project Brief: things3-api

## Purpose
Local REST API server that bridges Things 3 (macOS task manager) with external HTTP clients via AppleScript. Enables AI assistants, scripts, and automation tools to read and write Things 3 data securely over HTTP.

## Core Requirements
- Single Go binary, zero runtime dependencies
- Bearer token authentication
- Full CRUD for tasks, projects, and areas
- AppleScript-based communication with Things 3
- macOS launchd service for auto-start
- Localhost-only by default, with optional Tailscale support

## Target Users
- Developers automating task management
- AI assistants (Claude, GPT, etc.) that need to interact with Things 3
- Scripts and integrations

## Technical Constraints
- macOS only (AppleScript dependency)
- Things 3 must be installed and running
- Go 1.21+ (no external dependencies)
- AppleScript is synchronous — no concurrency issues with Things 3
