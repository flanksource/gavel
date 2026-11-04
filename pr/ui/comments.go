package ui

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/flanksource/gavel/comments"
	commentblob "github.com/flanksource/gavel/comments/blob"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	commentlint "github.com/flanksource/gavel/comments/lint"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/comments/providers"
	"github.com/flanksource/gavel/internal/database"
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
	access := commentAccess{}
	return comments.New(comments.Options{Store: store, Access: access, Targets: access, Workflows: []comments.Workflow{comments.HumanReview(), comments.NativePR()}, Blobs: blobs, AttachmentLimit: 10 << 20, Providers: []comments.Provider{providers.GitHub{Resolve: commentGitHub}, providers.Azure{Resolve: commentAzure}}})
}

type commentAccess struct{}

func (commentAccess) Principal(ctx context.Context) (comments.Principal, error) {
	scope := tenant.FromContext(ctx)
	if scope.UserID != nil {
		return comments.Principal{TenantID: scope.OrgID.String(), Actor: comments.Actor{Kind: "user", ID: scope.UserID.String()}}, nil
	}
	if !scope.IsLocal() {
		return comments.Principal{}, comments.ErrForbidden
	}
	return comments.Principal{TenantID: scope.OrgID.String(), Actor: comments.Actor{Kind: "user", ID: fmt.Sprintf("local:%d", os.Getuid())}}, nil
}
func (commentAccess) Authorize(ctx context.Context, p comments.Principal, _ string, t comments.Target) error {
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
func (a commentAccess) Resolve(ctx context.Context, p comments.Principal, t comments.Target) (comments.Target, error) {
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
	var projects []projectInfo
	if s := entityServer(ctx); s != nil && s.hosted {
		projects, err = s.projectList(ctx)
	} else {
		var local []Project
		local, err = LoadProjects(LoadProjectsOptions{IncludeHidden: true})
		for _, p := range local {
			projects = append(projects, projectInfo{Name: p.Name, Dir: p.ResolvedDir(), Repos: p.Repos})
		}
	}
	if err != nil {
		return projectInfo{}, err
	}
	for _, p := range projects {
		if p.Name == name {
			return p, nil
		}
	}
	return projectInfo{}, comments.ErrForbidden
}
