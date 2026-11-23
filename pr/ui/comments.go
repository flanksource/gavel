package ui

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/flanksource/gavel/projects"

	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	commentblob "github.com/flanksource/gavel/comments/blob"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	commentlint "github.com/flanksource/gavel/comments/lint"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/comments/providers"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/settings"
	"github.com/flanksource/gavel/tenant"
	"gorm.io/gorm"
)

var commentEntityOnce sync.Once
var commentEntityErr error

// RegisterComments declares request-scoped operations without opening or provisioning a database.
func RegisterComments() error {
	commentEntityOnce.Do(func() {
		if commentEntityErr = commentapi.Register(OpenComments); commentEntityErr != nil {
			return
		}
		commentEntityErr = commentlint.Register(commentlint.APIOptions{Service: OpenComments, Directory: commentLintDirectory})
	})
	return commentEntityErr
}
func OpenComments(ctx context.Context) (*comments.Service, error) {
	var db *gorm.DB
	var err error
	if s := entityServer(ctx); s != nil && s.orgDB != nil {
		db = s.orgDB
	} else {
		db, err = database.Require(ctx, "Gavel comments")
	}
	if err != nil {
		return nil, err
	}
	store, err := postgres.New(db)
	if err != nil {
		return nil, err
	}
	var blobs comments.Blobs = store
	if s := entityServer(ctx); s != nil && s.blobs != nil {
		blobs = commentblob.Store{Store: s.blobs}
	}
	access := CommentAccess{}
	options := comments.Options{Store: store, Access: access, Targets: access, Workflows: []comments.Workflow{comments.HumanReview(), comments.NativePR()}, Blobs: blobs, AttachmentLimit: 10 << 20, Providers: []comments.Provider{providers.GitHub{Resolve: commentGitHub}, providers.Azure{Resolve: commentAzure}}}
	// Agent delivery reaches this machine's Captain sessions, so only the local tenant gets it.
	if tenant.FromContext(ctx).IsLocal() {
		if options.Delivery, err = agents.Delivery(agents.Options{Bridge: agents.Captain{}, Groups: store, Workdir: CommentWorkdir, Amp: agents.AmpFromEnvironment(), Repositories: CommentRepositories}); err != nil {
			return nil, err
		}
		catalog, err := projects.Shared(ctx)
		if err != nil {
			return nil, err
		}
		configured, err := settings.Open(ctx, catalog)
		if err != nil {
			return nil, err
		}
		options.Routes = agents.ProjectRoutes(catalog, configured)
		options.Locate = func(ctx context.Context, path string) (comments.Location, error) {
			return LocateCommentPath(ctx, catalog, path)
		}
		options.Proposals = CommentProposals{}
	}
	return comments.New(options)
}

// LocateCommentPath lists a file's threads exactly, a directory's by prefix and
// a project root's whole scope.
func LocateCommentPath(ctx context.Context, catalog projects.Service, path string) (comments.Location, error) {
	target, err := agents.ProjectPathTarget(ctx, catalog, path)
	if errors.Is(err, projects.ErrNotFound) {
		return comments.Location{}, fmt.Errorf("%w: %s is not inside a gavel project", comments.ErrInvalid, path)
	}
	if err != nil {
		return comments.Location{}, err
	}
	switch {
	case target.URI == "":
		return comments.Location{ScopeURI: target.Scope}, nil
	case target.IsDir:
		return comments.Location{ScopeURI: target.Scope, TargetPrefix: target.URI + "/"}, nil
	default:
		return comments.Location{ScopeURI: target.Scope, TargetURI: target.URI}, nil
	}
}

// CommentWorkdir is the directory of the project a gavel://project/<name> target belongs to.
func CommentWorkdir(ctx context.Context, t comments.Target) (string, error) {
	project, err := commentProject(ctx, t.ScopeURI)
	if err != nil {
		return "", err
	}
	if project.Dir == "" {
		return "", fmt.Errorf("%w: project %q has no local directory for an agent session to run in", comments.ErrNoWorkdir, project.Name)
	}
	return projects.CanonicalPath(project.Dir)
}

func CommentRepositories(ctx context.Context, scope string) ([]string, error) {
	project, err := commentProject(ctx, scope)
	return project.Repos, err
}

var _ comments.TargetSources = CommentAccess{}

// TargetSource is the text on disk of a gavel://project file target
// (<scope>/files/<path>), so its threads' patch anchors resolve against the
// file as it is now. A project's other targets (its dashboard pages, a pull
// request) and other scopes keep no file content. A file deleted since its
// threads were written resolves as empty text, so every anchor on it is
// outdated; any other read failure is an error.
func (CommentAccess) TargetSource(ctx context.Context, t comments.Target) (string, bool, error) {
	if !tenant.FromContext(ctx).IsLocal() {
		return "", false, nil
	}
	if !strings.HasPrefix(t.ScopeURI, "gavel://project/") || !strings.HasPrefix(t.URI, t.ScopeURI+"/files/") {
		return "", false, nil
	}
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return "", false, err
	}
	_, _, path, err := agents.ProjectFile(ctx, catalog, t.ScopeURI, t.URI)
	if err != nil {
		return "", false, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", true, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(data), true, nil
}

// CommentAccess is the comment host for gavel://project/<name> scopes: the
// signed-in user or, locally, the asserted actor or the OS user. `gavel edit`
// embeds it so file threads live under their project, where `gavel comments` reaches them.
type CommentAccess struct{}

func (CommentAccess) Principal(ctx context.Context) (comments.Principal, error) {
	scope := tenant.FromContext(ctx)
	if scope.UserID != nil {
		return comments.Principal{TenantID: scope.OrgID.String(), Actor: comments.Actor{Kind: "user", ID: scope.UserID.String()}}, nil
	}
	if !scope.IsLocal() {
		return comments.Principal{}, comments.ErrForbidden
	}
	// Locally the caller may act as an agent it asserted (GAVEL_COMMENT_ACTOR on
	// the CLI, the loopback actor header over HTTP): the same machine-local trust.
	if actor, ok := commentapi.ActorFrom(ctx); ok {
		return comments.Principal{TenantID: scope.OrgID.String(), Actor: actor}, nil
	}
	actor, err := commentapi.LocalUser()
	return comments.Principal{TenantID: scope.OrgID.String(), Actor: actor}, err
}
func (CommentAccess) Authorize(ctx context.Context, p comments.Principal, _ string, t comments.Target) error {
	if p.TenantID != tenant.FromContext(ctx).OrgID.String() {
		return comments.ErrForbidden
	}
	if s := entityServer(ctx); s != nil && s.orgID().String() != p.TenantID {
		return comments.ErrForbidden
	}
	project, err := commentProject(ctx, t.ScopeURI)
	if err != nil {
		return err
	}
	if t.URI == "" {
		return nil
	}
	if strings.HasPrefix(t.URI, t.ScopeURI+"/") {
		return nil
	}
	u, err := url.Parse(t.URI)
	if err != nil || u.Scheme != "https" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return comments.ErrForbidden
	}
	if u.Host == "github.com" {
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) != 4 || parts[2] != "pull" {
			return comments.ErrForbidden
		}
		for _, repo := range project.Repos {
			if repo == parts[0]+"/"+parts[1] {
				return nil
			}
		}
		return comments.ErrForbidden
	}
	if u.Host == "dev.azure.com" {
		_, err := commentAzure(ctx, p, t)
		return err
	}
	return comments.ErrForbidden
}
func (a CommentAccess) Resolve(ctx context.Context, p comments.Principal, t comments.Target) (comments.Target, error) {
	if err := a.Authorize(ctx, p, "resolve-target", t); err != nil {
		return t, err
	}
	if t.Kind == "" || t.URI == "" {
		return t, comments.ErrInvalid
	}
	if t.Kind == "code" {
		selector, err := comments.ParseCodeSelector(t.Selector)
		if err != nil {
			return t, err
		}
		if selector.Kind == "code-range" && selector.Revision != t.Revision {
			return t, comments.ErrConflict
		}
	}
	return t, nil
}
func commentProject(ctx context.Context, scope string) (projectInfo, error) {
	raw, ok := strings.CutPrefix(scope, "gavel://project/")
	if !ok || raw == "" {
		return projectInfo{}, fmt.Errorf("%w: scope must name gavel://project/<escaped-project-name>", comments.ErrInvalid)
	}
	name, err := url.PathUnescape(raw)
	if err != nil || url.PathEscape(name) != raw {
		return projectInfo{}, comments.ErrInvalid
	}
	if s := entityServer(ctx); s != nil && s.hosted {
		infos, err := s.projectList(ctx)
		if err != nil {
			return projectInfo{}, err
		}
		for _, p := range infos {
			if p.Name == name {
				return p, nil
			}
		}
		return projectInfo{}, comments.ErrForbidden
	}
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return projectInfo{}, err
	}
	p, err := catalog.Get(ctx, name)
	if errors.Is(err, projects.ErrNotFound) || (err == nil && p.Kind == projects.KindWeb) {
		return projectInfo{}, comments.ErrForbidden
	}
	if err != nil {
		return projectInfo{}, err
	}
	return projectInfo{Name: p.Name, Dir: p.ResolvedDir(), Repos: p.Repos}, nil
}
