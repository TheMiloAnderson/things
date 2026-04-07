package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"things/internal/models"
)

type projectUpdateJSON struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Notes  string `json:"notes"`
	AreaID *int   `json:"area_id"`
}

type nameUpdateJSON struct {
	Name string `json:"name"`
}

func (a *App) apiProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiJSONResponse{OK: false, Error: "method not allowed"})
		return
	}
	userID := userIDFromRequest(r)
	rest := strings.TrimPrefix(r.URL.Path, "/api/project/")
	rest = strings.Trim(rest, "/")
	parts := strings.Split(rest, "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}
	action := parts[1]

	p, err := a.Data.GetProject(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, apiJSONResponse{OK: false, Error: "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
		return
	}
	if p.UserID != userID {
		writeJSON(w, http.StatusNotFound, apiJSONResponse{OK: false, Error: "not found"})
		return
	}

	switch action {
	case "update":
		var body projectUpdateJSON
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, apiJSONResponse{OK: false, Error: "invalid JSON"})
			return
		}
		name := strings.TrimSpace(body.Name)
		if name == "" {
			writeJSON(w, http.StatusBadRequest, apiJSONResponse{OK: false, Error: "name is required"})
			return
		}
		p.Name = name
		switch models.Status(body.Status) {
		case models.StatusActive, models.StatusDone, models.StatusCanceled:
			p.Status = models.Status(body.Status)
		default:
			p.Status = models.StatusActive
		}
		p.Notes = body.Notes
		p.AreaID = 0
		if body.AreaID != nil && *body.AreaID > 0 {
			p.AreaID = *body.AreaID
		}
		if err := a.updateProjectWithCascade(p); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiJSONResponse{OK: true})

	case "done":
		p.Status = models.StatusDone
		if err := a.updateProjectWithCascade(p); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiJSONResponse{OK: true})

	case "canceled":
		p.Status = models.StatusCanceled
		if err := a.updateProjectWithCascade(p); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiJSONResponse{OK: true})

	case "reactivate":
		if p.Status != models.StatusCanceled {
			writeJSON(w, http.StatusBadRequest, apiJSONResponse{OK: false, Error: "only canceled projects can be re-activated"})
			return
		}
		p.Status = models.StatusActive
		if err := a.Data.UpdateProject(p); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiJSONResponse{OK: true})

	case "delete":
		if err := a.Data.DeleteProjectForUser(id, userID); err != nil {
			if err == sql.ErrNoRows {
				writeJSON(w, http.StatusNotFound, apiJSONResponse{OK: false, Error: "not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiJSONResponse{OK: true})

	default:
		http.NotFound(w, r)
	}
}

func (a *App) apiAreaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiJSONResponse{OK: false, Error: "method not allowed"})
		return
	}
	userID := userIDFromRequest(r)
	rest := strings.TrimPrefix(r.URL.Path, "/api/area/")
	rest = strings.Trim(rest, "/")
	parts := strings.Split(rest, "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}
	action := parts[1]

	ar, err := a.Data.GetArea(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, apiJSONResponse{OK: false, Error: "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
		return
	}
	if ar.UserID != userID {
		writeJSON(w, http.StatusNotFound, apiJSONResponse{OK: false, Error: "not found"})
		return
	}

	switch action {
	case "update":
		var body nameUpdateJSON
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, apiJSONResponse{OK: false, Error: "invalid JSON"})
			return
		}
		name := strings.TrimSpace(body.Name)
		if name == "" {
			writeJSON(w, http.StatusBadRequest, apiJSONResponse{OK: false, Error: "name is required"})
			return
		}
		ar.Name = name
		if err := a.Data.UpdateArea(ar); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiJSONResponse{OK: true})

	case "delete":
		if err := a.Data.DeleteAreaForUser(id, userID); err != nil {
			if err == sql.ErrNoRows {
				writeJSON(w, http.StatusNotFound, apiJSONResponse{OK: false, Error: "not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiJSONResponse{OK: true})

	default:
		http.NotFound(w, r)
	}
}

func (a *App) apiContextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiJSONResponse{OK: false, Error: "method not allowed"})
		return
	}
	userID := userIDFromRequest(r)
	rest := strings.TrimPrefix(r.URL.Path, "/api/context/")
	rest = strings.Trim(rest, "/")
	parts := strings.Split(rest, "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}
	action := parts[1]

	c, err := a.Data.GetContext(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, apiJSONResponse{OK: false, Error: "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
		return
	}
	if c.UserID != userID {
		writeJSON(w, http.StatusNotFound, apiJSONResponse{OK: false, Error: "not found"})
		return
	}

	switch action {
	case "update":
		var body nameUpdateJSON
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, apiJSONResponse{OK: false, Error: "invalid JSON"})
			return
		}
		name := strings.TrimSpace(body.Name)
		if name == "" {
			writeJSON(w, http.StatusBadRequest, apiJSONResponse{OK: false, Error: "name is required"})
			return
		}
		c.Name = name
		if err := a.Data.UpdateContext(c); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiJSONResponse{OK: true})

	case "delete":
		if err := a.Data.DeleteContextForUser(id, userID); err != nil {
			if err == sql.ErrNoRows {
				writeJSON(w, http.StatusNotFound, apiJSONResponse{OK: false, Error: "not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiJSONResponse{OK: true})

	default:
		http.NotFound(w, r)
	}
}
