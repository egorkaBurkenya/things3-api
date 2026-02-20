package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// internalError logs the real error server-side and returns a generic message
// to the client to avoid leaking internal details (file paths, AppleScript errors, etc.).
func internalError(w http.ResponseWriter, err error) {
	slog.Error("internal error", "error", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

// extractID extracts an ID from a URL path segment.
// For example, given path "/tasks/ABC-123" and prefix "/tasks/", returns "ABC-123".
func extractID(path, prefix string) string {
	s := strings.TrimPrefix(path, prefix)
	s = strings.TrimSuffix(s, "/")
	if idx := strings.Index(s, "/"); idx != -1 {
		return s[:idx]
	}
	return s
}

// pathSuffix returns the remaining path after the ID.
// For example, "/tasks/ABC-123/complete" with prefix "/tasks/" returns "/complete".
func pathSuffix(path, prefix string) string {
	s := strings.TrimPrefix(path, prefix)
	if idx := strings.Index(s, "/"); idx != -1 {
		return s[idx:]
	}
	return ""
}
