package dashboardsvc

import (
	_ "embed"
	"errors"
	"html/template"
	"net/http"

	"golang.org/x/tools/txtar"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s *svc) initProjects() {
	s.HandleFunc(rootPath+"projects", s.projects)
	s.HandleFunc("GET "+rootPath+"projects/{pid}", s.project)
	s.HandleFunc("DELETE "+rootPath+"projects/{pid}", s.deleteProject)
}

type project struct{ sdktypes.Project }

func (p project) FieldsOrder() []string       { return []string{"name", "project_id"} }
func (p project) HideFields() []string        { return nil }
func (p project) ExtraFields() map[string]any { return nil }

func toProject(sdkP sdktypes.Project) project { return project{sdkP} }

func (s *svc) listProjects(w http.ResponseWriter, r *http.Request) (list, error) {
	sdkPs, err := s.Projects().List(r.Context(), sdktypes.InvalidOrgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	return genListData(nil, kittehs.Transform(sdkPs, toProject)), nil
}

func (s *svc) projects(w http.ResponseWriter, r *http.Request) {
	ps, err := s.listProjects(w, r)
	if err != nil {
		return
	}

	renderList(w, r, "projects", ps)
}

func (s *svc) project(w http.ResponseWriter, r *http.Request) {
	pid, err := sdktypes.StrictParseProjectID(r.PathValue("pid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkP, err := s.Projects().GetByID(r.Context(), pid)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sdkerrors.ErrNotFound) {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	p := toProject(sdkP)

	cs, err := s.listConnections(w, r, sdkservices.ListConnectionsFilter{
		ProjectID: pid,
	})
	if err != nil {
		return
	}

	ts, err := s.listTriggers(w, r, sdkservices.ListTriggersFilter{ProjectID: pid})
	if err != nil {
		return
	}

	rscs, err := s.Projects().DownloadResources(r.Context(), pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var a txtar.Archive
	for k, v := range rscs {
		a.Files = append(a.Files, txtar.File{Name: k, Data: v})
	}

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "project.html", struct {
		Message       string
		Title         string
		Name          string
		JSON          template.HTML
		Connections   list
		Triggers      list
		Sessions      list
		ID            string
		ResourcesHash string
		Resources     template.HTML
	}{
		Title:         "Project: " + p.Name().String(),
		Name:          p.Name().String(),
		JSON:          marshalObject(sdkP.ToProto()),
		Connections:   cs,
		Triggers:      ts,
		ID:            p.ID().String(),
		ResourcesHash: kittehs.Must1(kittehs.SHA256HashMap(rscs)),
		Resources:     template.HTML(txtar.Format(&a)),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *svc) deleteProject(w http.ResponseWriter, r *http.Request) {
	pid, err := sdktypes.StrictParseProjectID(r.PathValue("pid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.Projects().Delete(r.Context(), pid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
