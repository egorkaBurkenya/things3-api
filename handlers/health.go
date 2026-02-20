package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/egorkaBurkenya/things3-api/applescript"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := "running"
	if !applescript.IsThings3Running() {
		status = "not_running"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"things3": status,
	})
}
