package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/egorkaBurkenya/things3-api/config"
	"github.com/egorkaBurkenya/things3-api/handlers"
	"github.com/egorkaBurkenya/things3-api/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.LogLevel == "debug" {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}

	mux := http.NewServeMux()

	// Health (no auth required, handled by middleware exemption)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		handlers.HealthCheck(w, r)
	})

	// Tasks
	mux.HandleFunc("/tasks/", handlers.TasksRouter)
	mux.HandleFunc("/tasks", handlers.TasksRouter)

	// Projects
	mux.HandleFunc("/projects/", handlers.ProjectsRouter)
	mux.HandleFunc("/projects", handlers.ProjectsRouter)

	// Areas
	mux.HandleFunc("/areas/", handlers.AreasRouter)
	mux.HandleFunc("/areas", handlers.AreasRouter)

	handler := middleware.Chain(mux,
		middleware.Recovery(),
		middleware.Logger(),
		middleware.MaxBody(1<<20), // 1MB
		middleware.Auth(cfg.Token),
		middleware.Things3Check(),
	)

	slog.Info("starting things3-api", "addr", cfg.Addr())
	if err := http.ListenAndServe(cfg.Addr(), handler); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
