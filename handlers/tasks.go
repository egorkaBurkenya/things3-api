package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/egorkaBurkenya/things3-api/applescript"
	"github.com/egorkaBurkenya/things3-api/models"
)

// TasksRouter handles all /tasks routes.
func TasksRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/tasks" || path == "/tasks/":
		switch r.Method {
		case http.MethodGet:
			getFilteredTasks(w, r)
		case http.MethodPost:
			createTask(w, r)
		default:
			methodNotAllowed(w)
		}
	case path == "/tasks/inbox":
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		getInboxTasks(w, r)
	case path == "/tasks/today":
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		getTodayTasks(w, r)
	case path == "/tasks/upcoming":
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		getUpcomingTasks(w, r)
	case path == "/tasks/anytime":
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		getAnytimeTasks(w, r)
	case path == "/tasks/someday":
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		getSomedayTasks(w, r)
	default:
		// /tasks/{id} or /tasks/{id}/complete or /tasks/{id}/cancel
		id := extractID(path, "/tasks/")
		suffix := pathSuffix(path, "/tasks/")

		switch {
		case suffix == "/complete" && r.Method == http.MethodPost:
			completeTask(w, r, id)
		case suffix == "/cancel" && r.Method == http.MethodPost:
			cancelTask(w, r, id)
		case suffix == "" && r.Method == http.MethodGet:
			getTaskByID(w, r, id)
		case suffix == "" && r.Method == http.MethodPatch:
			updateTask(w, r, id)
		case suffix == "" && r.Method == http.MethodDelete:
			deleteTask(w, r, id)
		default:
			writeError(w, http.StatusNotFound, "not found")
		}
	}
}

func getInboxTasks(w http.ResponseWriter, _ *http.Request) {
	tasks, err := applescript.GetInboxTasks()
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func getTodayTasks(w http.ResponseWriter, _ *http.Request) {
	tasks, err := applescript.GetTodayTasks()
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func getUpcomingTasks(w http.ResponseWriter, _ *http.Request) {
	tasks, err := applescript.GetUpcomingTasks()
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func getAnytimeTasks(w http.ResponseWriter, _ *http.Request) {
	tasks, err := applescript.GetAnytimeTasks()
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func getSomedayTasks(w http.ResponseWriter, _ *http.Request) {
	tasks, err := applescript.GetSomedayTasks()
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func getFilteredTasks(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	area := r.URL.Query().Get("area")
	tag := r.URL.Query().Get("tag")

	if project == "" && area == "" && tag == "" {
		writeError(w, http.StatusBadRequest, "at least one filter parameter is required: project, area, or tag")
		return
	}

	// Limit query parameter lengths to prevent abuse.
	if len(project) > 500 || len(area) > 500 || len(tag) > 500 {
		writeError(w, http.StatusBadRequest, "filter parameter values must be under 500 characters")
		return
	}

	tasks, err := applescript.GetFilteredTasks(project, area, tag)
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func getTaskByID(w http.ResponseWriter, _ *http.Request, id string) {
	if err := models.ValidateThingsID(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := applescript.GetTaskByID(id)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := applescript.CreateTask(req)
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func updateTask(w http.ResponseWriter, r *http.Request, id string) {
	if err := models.ValidateThingsID(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := applescript.UpdateTask(id, req)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func completeTask(w http.ResponseWriter, _ *http.Request, id string) {
	if err := models.ValidateThingsID(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	if err := applescript.CompleteTask(id); err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func cancelTask(w http.ResponseWriter, _ *http.Request, id string) {
	if err := models.ValidateThingsID(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	if err := applescript.CancelTask(id); err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func deleteTask(w http.ResponseWriter, _ *http.Request, id string) {
	if err := models.ValidateThingsID(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	if err := applescript.DeleteTask(id); err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func isNotFound(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") ||
		strings.Contains(msg, "can't get") ||
		strings.Contains(msg, "couldn't find")
}
