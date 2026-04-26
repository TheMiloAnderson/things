package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"things/internal/models"
)

type managementProjectsListView struct {
	Section  string
	Projects []models.Project
}

type managementProjectNewView struct {
	Section string
	Areas   []models.Area
}

type managementProjectEditView struct {
	Section   string
	Project   models.Project
	Areas     []models.Area
	Contexts  []models.Context
	TaskForms TasksListViewModel
}

type managementAreasListView struct {
	Section string
	Areas   []models.Area
}

type managementAreaNewView struct {
	Section string
}

type managementContextsListView struct {
	Section  string
	Contexts []models.Context
}

type managementContextNewView struct {
	Section string
}

func (a *App) managementRedirect(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/management" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/management/projects/", http.StatusSeeOther)
}

func (a *App) managementRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/management")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	userID := userIDFromRequest(r)
	u, err := a.Data.GetUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page := PageData{
		IsAuthenticated: true,
		Username:        u.Name,
	}

	switch {
	case len(parts) == 0 || (len(parts) == 1 && parts[0] == ""):
		http.Redirect(w, r, "/management/projects/", http.StatusSeeOther)

	case len(parts) >= 1 && parts[0] == "projects":
		a.handleManagementProjects(w, r, &page, userID, parts)

	case len(parts) >= 1 && parts[0] == "areas":
		a.handleManagementAreas(w, r, &page, userID, parts)

	case len(parts) >= 1 && parts[0] == "contexts":
		a.handleManagementContexts(w, r, &page, userID, parts)

	default:
		http.NotFound(w, r)
	}
}

func (a *App) handleManagementProjects(w http.ResponseWriter, r *http.Request, page *PageData, userID int, parts []string) {
	switch {
	case len(parts) == 2 && parts[1] == "new" && r.Method == http.MethodGet:
		areas, err := a.Data.AllAreas(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		page.Data = managementProjectNewView{Section: "projects", Areas: areas}
		a.render(w, r, "management_projects_new.html", *page)

	case len(parts) == 2 && parts[1] != "new":
		id, err := strconv.Atoi(parts[1])
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			p, err := a.Data.GetProject(id)
			if err != nil {
				if err == sql.ErrNoRows {
					http.NotFound(w, r)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if p.UserID != userID {
				http.NotFound(w, r)
				return
			}
			areas, err := a.Data.AllAreas(userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			contexts, err := a.Data.AllContextsForUser(userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			projects, err := a.Data.AllActiveProjects(userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			projectTasks, err := a.Data.AllTasksForProject(userID, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			areasByID := make(map[int]string, len(areas))
			for _, a := range areas {
				areasByID[a.ID] = a.Name
			}
			projectsByID := make(map[int]string, len(projects))
			for _, prj := range projects {
				projectsByID[prj.ID] = prj.Name
			}
			page.Data = managementProjectEditView{
				Section:  "projects",
				Project:  p,
				Areas:    areas,
				Contexts: contexts,
				TaskForms: TasksListViewModel{
					Tasks:          projectTasks,
					Projects:       projects,
					Areas:          areas,
					AreasByID:      areasByID,
					ProjectsByID:   projectsByID,
					ProjectPageID:  id,
					TaskFormsTitle: "Tasks in this project",
				},
			}
			a.render(w, r, "management_projects_edit.html", *page)
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Could not parse form", http.StatusBadRequest)
				return
			}
			p, err := a.Data.GetProject(id)
			if err != nil {
				if err == sql.ErrNoRows {
					http.NotFound(w, r)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if p.UserID != userID {
				http.NotFound(w, r)
				return
			}
			name := strings.TrimSpace(r.FormValue("name"))
			if name == "" {
				http.Error(w, "Name is required", http.StatusBadRequest)
				return
			}
			p.Name = name
			status := models.Status(strings.TrimSpace(r.FormValue("status")))
			switch status {
			case models.StatusActive, models.StatusDone, models.StatusCanceled:
				p.Status = status
			default:
				p.Status = models.StatusActive
			}
			p.Notes = strings.TrimSpace(r.FormValue("notes"))
			p.AreaID = 0
			if s := strings.TrimSpace(r.FormValue("area_id")); s != "" {
				if v, err := strconv.Atoi(s); err == nil {
					p.AreaID = v
				}
			}
			if err := a.updateProjectWithCascade(p); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/management/projects/"+strconv.Itoa(id), http.StatusSeeOther)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	case len(parts) == 1 && r.Method == http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Could not parse form", http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		status := models.Status(strings.TrimSpace(r.FormValue("status")))
		switch status {
		case models.StatusActive, models.StatusDone, models.StatusCanceled:
		default:
			status = models.StatusActive
		}
		notes := strings.TrimSpace(r.FormValue("notes"))
		areaID := 0
		if s := strings.TrimSpace(r.FormValue("area_id")); s != "" {
			if v, err := strconv.Atoi(s); err == nil {
				areaID = v
			}
		}
		p := models.Project{
			Name:   name,
			Status: status,
			Notes:  notes,
			AreaID: areaID,
			UserID: userID,
		}
		newID, err := a.Data.SaveProject(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/management/projects/"+strconv.FormatInt(newID, 10), http.StatusSeeOther)

	case len(parts) == 1 && r.Method == http.MethodGet:
		projects, err := a.Data.AllProjectsForUser(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		page.Data = managementProjectsListView{Section: "projects", Projects: projects}
		a.render(w, r, "management_projects_list.html", *page)

	default:
		http.NotFound(w, r)
	}
}

func (a *App) handleManagementAreas(w http.ResponseWriter, r *http.Request, page *PageData, userID int, parts []string) {
	switch {
	case len(parts) == 2 && parts[1] == "new" && r.Method == http.MethodGet:
		page.Data = managementAreaNewView{Section: "areas"}
		a.render(w, r, "management_areas_new.html", *page)

	case len(parts) == 1 && r.Method == http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Could not parse form", http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		ar := models.Area{Name: name, UserID: userID}
		if _, err := a.Data.SaveArea(ar); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/management/areas/", http.StatusSeeOther)

	case len(parts) == 1 && r.Method == http.MethodGet:
		areas, err := a.Data.AllAreas(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		page.Data = managementAreasListView{Section: "areas", Areas: areas}
		a.render(w, r, "management_areas_list.html", *page)

	default:
		http.NotFound(w, r)
	}
}

func (a *App) handleManagementContexts(w http.ResponseWriter, r *http.Request, page *PageData, userID int, parts []string) {
	switch {
	case len(parts) == 2 && parts[1] == "new" && r.Method == http.MethodGet:
		page.Data = managementContextNewView{Section: "contexts"}
		a.render(w, r, "management_contexts_new.html", *page)

	case len(parts) == 1 && r.Method == http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Could not parse form", http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		c := models.Context{Name: name, UserID: userID}
		if _, err := a.Data.SaveContext(c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/management/contexts/", http.StatusSeeOther)

	case len(parts) == 1 && r.Method == http.MethodGet:
		contexts, err := a.Data.AllContextsForUser(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		page.Data = managementContextsListView{Section: "contexts", Contexts: contexts}
		a.render(w, r, "management_contexts_list.html", *page)

	default:
		http.NotFound(w, r)
	}
}
