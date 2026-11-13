package editor

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/comments/postgres"
	ui "github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
)

// routeNamespace makes a configured route's ID a function of where it applies,
// so the same .gavel.yaml entry keeps one ID across every projection of it.
var routeNamespace = uuid.MustParse("6f1c4a52-0f43-4c56-9a7e-0b9a3c1d2e7f")

// configRoutes is the editor's comment route store, a project settings adapter:
// a page's agent is configured in the .gavel.yaml chain above it
// (comments.agents), which is canonical, and comment_routes holds the JSONB
// projection of that configuration, rewritten whenever routes are read or changed.
type configRoutes struct {
	db    *postgres.Store
	files *Store
}

var _ comments.RouteStore = configRoutes{}

func (c configRoutes) Routes(ctx context.Context, tenant, scope string) ([]comments.Route, error) {
	root, err := projectRoot(scope)
	if err != nil {
		return nil, err
	}
	routes, err := configuredRoutes(tenant, scope, root, c.files.Paths())
	if err != nil {
		return nil, err
	}
	if err := c.db.ReplaceRoutes(ctx, tenant, scope, routes); err != nil {
		return nil, fmt.Errorf("store the configured routes of %s: %w", scope, err)
	}
	return routes, nil
}

func (c configRoutes) Route(ctx context.Context, tenant, id string) (comments.Route, error) {
	return c.db.Route(ctx, tenant, id)
}

// SaveRoute writes r's agent into the .gavel.yaml of the directory its prefix
// names (its default) or of the file it names (its page entry), then re-projects.
func (c configRoutes) SaveRoute(ctx context.Context, r comments.Route) (comments.Route, error) {
	agent, err := commentAgentOf(r)
	if err != nil {
		return comments.Route{}, err
	}
	place, err := placeOf(r.ScopeURI, r.TargetPrefix)
	if err != nil {
		return comments.Route{}, err
	}
	if err := verify.SaveCommentAgent(place.Dir, place.Page, &agent); err != nil {
		return comments.Route{}, err
	}
	return c.reprojected(ctx, r.TenantID, r.ScopeURI, routeID(r.ScopeURI, r.TargetPrefix))
}

func (c configRoutes) DeleteRoute(ctx context.Context, tenant, id string) error {
	r, err := c.db.Route(ctx, tenant, id)
	if err != nil {
		return err
	}
	place, err := placeOf(r.ScopeURI, r.TargetPrefix)
	if err != nil {
		return err
	}
	if err := verify.SaveCommentAgent(place.Dir, place.Page, nil); err != nil {
		return err
	}
	_, err = c.Routes(ctx, tenant, r.ScopeURI)
	return err
}

func (c configRoutes) reprojected(ctx context.Context, tenant, scope, id string) (comments.Route, error) {
	routes, err := c.Routes(ctx, tenant, scope)
	if err != nil {
		return comments.Route{}, err
	}
	for _, r := range routes {
		if r.ID == id {
			return r, nil
		}
	}
	return comments.Route{}, fmt.Errorf("route %s was saved but is not among the routes configured for %s", id, scope)
}

func routeID(scope, prefix string) string {
	return uuid.NewSHA1(routeNamespace, []byte(scope+"\n"+prefix)).String()
}

// filesPrefix is the target prefix covering rel ("." for the whole project) in scope.
func filesPrefix(scope, rel string) string {
	if rel == "." || rel == "" {
		return scope + "/files"
	}
	return scope + "/files/" + filepath.ToSlash(rel)
}

// configuredRoutes reads the .gavel.yaml chain of every edited file under root
// (verify.GavelConfigDirs): a directory's default applies beneath it, a page
// entry to its file. A directory above the project applies to the whole
// project, below the project's own configuration.
func configuredRoutes(tenant, scope, root string, files map[string]string) ([]comments.Route, error) {
	seen := map[string]bool{}
	var dirs []string
	for _, path := range files {
		if _, inside := within(root, path); !inside {
			continue
		}
		for _, dir := range verify.GavelConfigDirs(filepath.Dir(path)) {
			if !seen[dir.Dir] {
				seen[dir.Dir] = true
				dirs = append(dirs, dir.Dir)
			}
		}
	}
	// Ancestors first: a nearer .gavel.yaml replaces a farther one's route on the same prefix.
	sort.SliceStable(dirs, func(i, j int) bool { return depth(dirs[i]) < depth(dirs[j]) })
	byPrefix := map[string]comments.Route{}
	for _, dir := range dirs {
		cfg, err := verify.LoadSingleGavelConfig(filepath.Join(dir, verify.GavelConfigFileName))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		configured := cfg.Comments.Agents
		if configured.Default != nil {
			rel, inside := within(root, dir)
			if !inside {
				rel = "."
			}
			if err := addRoute(byPrefix, tenant, scope, filesPrefix(scope, rel), *configured.Default); err != nil {
				return nil, err
			}
		}
		for key, agent := range configured.Pages {
			rel, inside := within(root, filepath.Join(dir, filepath.FromSlash(key)))
			if !inside {
				continue
			}
			if err := addRoute(byPrefix, tenant, scope, filesPrefix(scope, rel), agent); err != nil {
				return nil, err
			}
		}
	}
	routes := make([]comments.Route, 0, len(byPrefix))
	for _, r := range byPrefix {
		routes = append(routes, r)
	}
	sort.Slice(routes, func(i, j int) bool { return routes[i].TargetPrefix < routes[j].TargetPrefix })
	return routes, nil
}

func addRoute(byPrefix map[string]comments.Route, tenant, scope, prefix string, agent verify.CommentAgent) error {
	options, err := json.Marshal(agents.RunOptions{Presets: agent.Presets, Spec: agent.Spec, Variables: agent.Variables, Instructions: agent.Instructions})
	if err != nil {
		return fmt.Errorf("encode the agent run options for %s: %w", prefix, err)
	}
	target := comments.DeliverTarget{Kind: "run", Prompt: agent.PromptOrDefault()}
	if string(options) != "{}" {
		target.RunOptions = options
	}
	byPrefix[prefix] = comments.Route{
		ID: routeID(scope, prefix), TenantID: tenant, ScopeURI: scope, TargetPrefix: prefix,
		WorkflowID: comments.HumanReview().ID, Target: target, Enabled: agent.Enabled, Assign: agent.Assign,
	}
	return nil
}

// commentAgentOf is the .gavel.yaml entry a route is saved as. Only a captain
// run is configured: a live session is a viewer's own choice, not project settings.
func commentAgentOf(r comments.Route) (verify.CommentAgent, error) {
	if r.WorkflowID != comments.HumanReview().ID {
		return verify.CommentAgent{}, fmt.Errorf("%w: a configured comment agent works %s threads, not %s", comments.ErrInvalid, comments.HumanReview().ID, r.WorkflowID)
	}
	if r.Target.Kind != "run" {
		return verify.CommentAgent{}, fmt.Errorf("%w: a configured comment agent is a captain run; choose a live session per viewer instead", comments.ErrInvalid)
	}
	var options agents.RunOptions
	if len(r.Target.RunOptions) > 0 {
		decoder := json.NewDecoder(strings.NewReader(string(r.Target.RunOptions)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&options); err != nil {
			return verify.CommentAgent{}, fmt.Errorf("%w: runOptions: %v", comments.ErrInvalid, err)
		}
	}
	return verify.CommentAgent{
		Enabled: r.Enabled, Assign: r.Assign, Prompt: r.Target.Prompt,
		Presets: options.Presets, Spec: options.Spec, Variables: options.Variables, Instructions: options.Instructions,
	}, nil
}

// agentPlace is the .gavel.yaml entry a target prefix is configured by: the
// default of directory Dir, or its Page entry.
type agentPlace struct {
	Dir  string
	Page string
}

// placeOf maps a prefix under scope to the project directory or file it names.
func placeOf(scope, prefix string) (agentPlace, error) {
	root, err := projectRoot(scope)
	if err != nil {
		return agentPlace{}, err
	}
	rel, ok := strings.CutPrefix(prefix, scope+"/files")
	if !ok || (rel != "" && !strings.HasPrefix(rel, "/")) {
		return agentPlace{}, fmt.Errorf("%w: %s is not a path under %s/files", comments.ErrInvalid, prefix, scope)
	}
	rel = strings.TrimPrefix(rel, "/")
	if rel != "" && !verify.IsCommentPageKey(rel) {
		return agentPlace{}, fmt.Errorf("%w: %s is not a clean path under the project", comments.ErrInvalid, prefix)
	}
	path := filepath.Join(root, filepath.FromSlash(rel))
	projectDir, err := os.OpenRoot(root)
	if err != nil {
		return agentPlace{}, fmt.Errorf("open project root %s: %w", root, err)
	}
	defer projectDir.Close()
	info, err := projectDir.Stat(cmp.Or(filepath.FromSlash(rel), "."))
	if err != nil {
		return agentPlace{}, fmt.Errorf("%w: %s: %v", comments.ErrInvalid, prefix, err)
	}
	if info.IsDir() {
		return agentPlace{Dir: path}, nil
	}
	return agentPlace{Dir: filepath.Dir(path), Page: filepath.Base(path)}, nil
}

// projectRoot is the canonical root of the project a gavel://project scope names.
func projectRoot(scope string) (string, error) {
	name, err := url.PathUnescape(strings.TrimPrefix(scope, "gavel://project/"))
	if err != nil || name == scope {
		return "", fmt.Errorf("%w: scope %s", errInvalidTarget, scope)
	}
	project, err := ui.GetProject(name)
	if err != nil {
		return "", err
	}
	root, err := ui.CanonicalPath(project.ResolvedDir())
	if err != nil {
		return "", fmt.Errorf("resolve project %q directory: %w", project.Name, err)
	}
	return root, nil
}

// within is path relative to root, and whether it is root or beneath it.
func within(root, path string) (string, bool) {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return rel, true
}

func depth(dir string) int { return strings.Count(filepath.Clean(dir), string(filepath.Separator)) }

// pageAgent is GET /api/file/{id}/agent: the routes configured for the file's
// page and its folder, the one that applies to it, and where saving them writes.
type pageAgent struct {
	Scope        string          `json:"scope"`
	PagePrefix   string          `json:"pagePrefix"`
	FolderPrefix string          `json:"folderPrefix"`
	ConfigPath   string          `json:"configPath"`
	Page         *comments.Route `json:"page"`
	Folder       *comments.Route `json:"folder"`
	Effective    *comments.Route `json:"effective"`
}

func (s *Server) pageAgentOf(w http.ResponseWriter, r *http.Request) {
	doc, err := s.Store.Read(r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	target, err := targetOf(r.Context(), doc.Path)
	if err != nil {
		respond(w, nil, err)
		return
	}
	service, err := s.OpenComments(r.Context())
	if err != nil {
		respond(w, nil, err)
		return
	}
	routes, err := service.Routes(r.Context(), target.Scope)
	if err != nil {
		respond(w, nil, err)
		return
	}
	rel, _ := strings.CutPrefix(target.URI, target.Scope+"/files/")
	out := pageAgent{
		Scope: target.Scope, PagePrefix: target.URI, FolderPrefix: filesPrefix(target.Scope, filepath.Dir(filepath.FromSlash(rel))),
		ConfigPath: filepath.Join(filepath.Dir(doc.Path), verify.GavelConfigFileName),
	}
	for i := range routes {
		switch routes[i].TargetPrefix {
		case out.PagePrefix:
			out.Page = &routes[i]
		case out.FolderPrefix:
			out.Folder = &routes[i]
		}
	}
	if best, _ := comments.MatchRoute(routes, target.URI, comments.HumanReview().ID); best.ID != "" {
		out.Effective = &best
	}
	respond(w, out, nil)
}
