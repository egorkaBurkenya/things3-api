package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/egorkaBurkenya/things3-api/applescript"
	"github.com/egorkaBurkenya/things3-api/models"
)

// AreasRouter handles all /areas routes.
func AreasRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/areas" || path == "/areas/":
		switch r.Method {
		case http.MethodGet:
			getAllAreas(w, r)
		case http.MethodPost:
			createArea(w, r)
		default:
			methodNotAllowed(w)
		}
	default:
		id := extractID(path, "/areas/")
		suffix := pathSuffix(path, "/areas/")

		switch {
		case suffix == "" && r.Method == http.MethodGet:
			getAreaByID(w, r, id)
		case suffix == "" && r.Method == http.MethodPatch:
			updateArea(w, r, id)
		default:
			writeError(w, http.StatusNotFound, "not found")
		}
	}
}

func getAllAreas(w http.ResponseWriter, _ *http.Request) {
	areas, err := applescript.GetAllAreas()
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, areas)
}

func getAreaByID(w http.ResponseWriter, _ *http.Request, id string) {
	if err := models.ValidateThingsID(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid area id")
		return
	}

	area, err := applescript.GetAreaByID(id)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "area not found")
			return
		}
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, area)
}

func createArea(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAreaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	area, err := applescript.CreateArea(req)
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, area)
}

func updateArea(w http.ResponseWriter, r *http.Request, id string) {
	if err := models.ValidateThingsID(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid area id")
		return
	}

	var req models.UpdateAreaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	area, err := applescript.UpdateArea(id, req)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "area not found")
			return
		}
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, area)
}
