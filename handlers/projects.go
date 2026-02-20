package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/egorkaBurkenya/things3-api/applescript"
	"github.com/egorkaBurkenya/things3-api/models"
)

// ProjectsRouter handles all /projects routes.
func ProjectsRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/projects" || path == "/projects/":
		switch r.Method {
		case http.MethodGet:
			getAllProjects(w, r)
		case http.MethodPost:
			createProject(w, r)
		default:
			methodNotAllowed(w)
		}
	default:
		id := extractID(path, "/projects/")
		suffix := pathSuffix(path, "/projects/")

		switch {
		case suffix == "/complete" && r.Method == http.MethodPost:
			completeProject(w, r, id)
		case suffix == "" && r.Method == http.MethodGet:
			getProjectByID(w, r, id)
		case suffix == "" && r.Method == http.MethodPatch:
			updateProject(w, r, id)
		default:
			writeError(w, http.StatusNotFound, "not found")
		}
	}
}

func getAllProjects(w http.ResponseWriter, _ *http.Request) {
	projects, err := applescript.GetAllProjects()
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func getProjectByID(w http.ResponseWriter, _ *http.Request, id string) {
	if err := models.ValidateThingsID(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	project, err := applescript.GetProjectByID(id)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func createProject(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	project, err := applescript.CreateProject(req)
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func updateProject(w http.ResponseWriter, r *http.Request, id string) {
	if err := models.ValidateThingsID(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req models.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	project, err := applescript.UpdateProject(id, req)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func completeProject(w http.ResponseWriter, _ *http.Request, id string) {
	if err := models.ValidateThingsID(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	if err := applescript.CompleteProject(id); err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
