package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"things/internal/models"
)

type taskUpdateJSON struct {
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	Priority   int     `json:"priority"`
	Notes      *string `json:"notes"`
	ProjectID  *int    `json:"project_id"`
	AreaID     *int    `json:"area_id"`
}

type apiJSONResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (a *App) apiTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiJSONResponse{OK: false, Error: "method not allowed"})
		return
	}

	userID := userIDFromRequest(r)

	rest := strings.TrimPrefix(r.URL.Path, "/api/task/")
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

	t, err := a.Data.GetTask(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, apiJSONResponse{OK: false, Error: "task not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
		return
	}
	if t.UserID != userID {
		writeJSON(w, http.StatusNotFound, apiJSONResponse{OK: false, Error: "task not found"})
		return
	}

	switch action {
	case "update":
		var body taskUpdateJSON
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, apiJSONResponse{OK: false, Error: "invalid JSON"})
			return
		}
		name := strings.TrimSpace(body.Name)
		if name == "" {
			writeJSON(w, http.StatusBadRequest, apiJSONResponse{OK: false, Error: "name is required"})
			return
		}
		t.Name = name
		switch models.Status(body.Status) {
		case models.StatusActive, models.StatusPending, models.StatusDone, models.StatusCanceled:
			t.Status = models.Status(body.Status)
		default:
			t.Status = models.StatusActive
		}
		p := models.Priority(body.Priority)
		switch p {
		case models.PriorityLow, models.PriorityMed, models.PriorityHigh:
			t.Priority = p
		default:
			t.Priority = models.PriorityLow
		}
		if body.Notes != nil {
			t.Notes = strings.TrimSpace(*body.Notes)
		}
		t.ProjectID = 0
		if body.ProjectID != nil && *body.ProjectID > 0 {
			t.ProjectID = *body.ProjectID
		}
		t.AreaID = 0
		if body.AreaID != nil && *body.AreaID > 0 {
			t.AreaID = *body.AreaID
		}
		if err := a.Data.UpdateTask(t); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiJSONResponse{OK: true})

	case "done":
		t.Status = models.StatusDone
		if err := a.Data.UpdateTask(t); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiJSONResponse{OK: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiJSONResponse{OK: true})

	default:
		http.NotFound(w, r)
	}
}
