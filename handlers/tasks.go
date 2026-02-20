package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/egorkaBurkenya/things3-api/applescript"
	"github.com/egorkaBurkenya/things3-api/database"
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
		// /tasks/{id} or /tasks/{id}/complete or /tasks/{id}/cancel or /tasks/{id}/checklist/...
		id := extractID(path, "/tasks/")
		suffix := pathSuffix(path, "/tasks/")

		switch {
		case suffix == "/complete" && r.Method == http.MethodPost:
			completeTask(w, r, id)
		case suffix == "/cancel" && r.Method == http.MethodPost:
			cancelTask(w, r, id)
		case suffix == "/checklist" || suffix == "/checklist/":
			switch r.Method {
			case http.MethodGet:
				getChecklistItems(w, r, id)
			case http.MethodPost:
				addChecklistItem(w, r, id)
			default:
				methodNotAllowed(w)
			}
		case strings.HasPrefix(suffix, "/checklist/"):
			itemID := strings.TrimPrefix(suffix, "/checklist/")
			itemID = strings.TrimSuffix(itemID, "/")
			switch r.Method {
			case http.MethodPatch:
				updateChecklistItem(w, r, id, itemID)
			case http.MethodDelete:
				deleteChecklistItem(w, r, id, itemID)
			default:
				methodNotAllowed(w)
			}
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

	if len(req.ChecklistItems) > 0 {
		// Use URL scheme to create task with checklist items (AppleScript can't do checklists).
		taskID, err := database.CreateTaskWithChecklist(
			req.Title, req.ChecklistItems,
			req.Notes, req.Project, req.Area, req.Due, req.When, req.Tags,
		)
		if err != nil {
			internalError(w, err)
			return
		}
		task, err := applescript.GetTaskByID(taskID)
		if err != nil {
			internalError(w, err)
			return
		}
		items, _ := database.GetChecklistItems(taskID)
		task.ChecklistItems = items
		writeJSON(w, http.StatusCreated, task)
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

func getChecklistItems(w http.ResponseWriter, _ *http.Request, taskID string) {
	if err := models.ValidateThingsID(taskID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	items, err := database.GetChecklistItems(taskID)
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func addChecklistItem(w http.ResponseWriter, r *http.Request, taskID string) {
	if err := models.ValidateThingsID(taskID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req models.CreateChecklistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Try URL scheme (shows in Things UI immediately) if auth token is available.
	authToken := os.Getenv("THINGS_URL_TOKEN")
	if authToken != "" {
		if err := database.AddChecklistItem(taskID, req.Title, authToken); err != nil {
			internalError(w, err)
			return
		}
		// Wait briefly for Things to process, then read back from DB.
		time.Sleep(500 * time.Millisecond)
		items, _ := database.GetChecklistItems(taskID)
		if len(items) > 0 {
			writeJSON(w, http.StatusCreated, items[len(items)-1])
			return
		}
		writeJSON(w, http.StatusCreated, models.ChecklistItem{Title: req.Title})
		return
	}

	// Fallback: direct SQLite insert (readable via API but may not show in Things UI).
	item, err := database.AddChecklistItemDirect(taskID, req)
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func updateChecklistItem(w http.ResponseWriter, r *http.Request, taskID, itemID string) {
	if err := models.ValidateThingsID(taskID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	if err := models.ValidateThingsID(itemID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid checklist item id")
		return
	}

	var req models.UpdateChecklistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := database.UpdateChecklistItem(taskID, itemID, req)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "checklist item not found")
			return
		}
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func deleteChecklistItem(w http.ResponseWriter, _ *http.Request, taskID, itemID string) {
	if err := models.ValidateThingsID(taskID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	if err := models.ValidateThingsID(itemID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid checklist item id")
		return
	}

	if err := database.DeleteChecklistItem(taskID, itemID); err != nil {
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
