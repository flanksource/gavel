package editor

import (
	"errors"
	"net/http"
	"sort"
)

// CommentProjectFile is an edited file inside a project, by the target URI its threads use.
type CommentProjectFile struct {
	ID  string `json:"id"`
	URI string `json:"uri"`
}

// CommentProject is one project holding edited files: the scope every thread
// in it lives under, and which of its target URIs are open in this session.
type CommentProject struct {
	Scope   string               `json:"scope"`
	Project string               `json:"project"`
	Dir     string               `json:"dir"`
	Files   []CommentProjectFile `json:"files"`
}

// commentProjects serves GET /api/comments/projects: the distinct projects of
// the edited files, ordered by name. Files outside every project are skipped,
// as the comment summary skips them.
func (s *Server) commentProjects(w http.ResponseWriter, r *http.Request) {
	byScope := map[string]*CommentProject{}
	for _, file := range s.Store.Files() {
		if file.Kind != KindDocument {
			continue
		}
		target, err := targetOf(r.Context(), file.Path)
		if errors.As(err, new(NoProjectError)) {
			continue
		}
		if err != nil {
			respond(w, nil, err)
			return
		}
		project := byScope[target.Scope]
		if project == nil {
			project = &CommentProject{Scope: target.Scope, Project: target.Project, Dir: target.Dir, Files: []CommentProjectFile{}}
			byScope[target.Scope] = project
		}
		project.Files = append(project.Files, CommentProjectFile{ID: file.ID, URI: target.URI})
	}
	out := make([]CommentProject, 0, len(byScope))
	for _, project := range byScope {
		out = append(out, *project)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Project < out[j].Project })
	respond(w, out, nil)
}
