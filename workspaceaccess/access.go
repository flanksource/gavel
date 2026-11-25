package workspaceaccess

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Root struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	Source    string `json:"source"`
	Available bool   `json:"available" gorm:"-"`
}
type Repository interface {
	List(context.Context) ([]Root, error)
	Grant(context.Context, string) (Root, error)
	Revoke(context.Context, string) error
}
type Service struct {
	Pinned     []string
	Repository Repository
}
type Snapshot struct {
	Roots     []Root `json:"roots"`
	CanManage bool   `json:"canManage"`
}
type PreviewRequest struct {
	Path string `json:"path"`
}
type GrantRequest struct {
	Path          string `json:"path"`
	CanonicalPath string `json:"canonicalPath"`
}
type Denied struct {
	Root         string `json:"root"`
	ErrorMessage string `json:"error"`
	Code         string `json:"code"`
}

func (e *Denied) Error() string { return e.ErrorMessage }
func New(pinned []string) *Service {
	return &Service{Pinned: append([]string{}, pinned...), Repository: durable{}}
}
func Canonical(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", errors.New("folder path must be absolute")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("select a directory")
	}
	return filepath.Clean(resolved), nil
}
func inside(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
func (s *Service) Roots(ctx context.Context) ([]Root, error) {
	roots, err := s.Repository.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, path := range s.Pinned {
		roots = append(roots, Root{Path: path, Source: "command-line"})
	}
	for i := range roots {
		_, err := Canonical(roots[i].Path)
		roots[i].Available = err == nil
	}
	return roots, nil
}
func (s *Service) Authorize(ctx context.Context, path string) (string, error) {
	resolved, err := Canonical(path)
	if err != nil {
		return "", err
	}
	for _, pinned := range s.Pinned {
		canonical, err := Canonical(pinned)
		if err != nil {
			return "", err
		}
		if inside(canonical, resolved) {
			return resolved, nil
		}
	}
	roots, err := s.Repository.List(ctx)
	if err != nil {
		return "", err
	}
	for _, root := range roots {
		if inside(root.Path, resolved) {
			return resolved, nil
		}
	}
	return "", &Denied{Root: path, Code: "workspace_access_denied", ErrorMessage: fmt.Sprintf("Workspace access is not enabled for %s", path)}
}
func (s *Service) Grant(ctx context.Context, path, expected string) (Root, error) {
	canonical, err := Canonical(path)
	if err != nil {
		return Root{}, err
	}
	if canonical != expected {
		return Root{}, errors.New("folder changed since preview; review access again")
	}
	for _, pinned := range s.Pinned {
		root, err := Canonical(pinned)
		if err != nil {
			return Root{}, err
		}
		if inside(root, canonical) {
			return Root{}, errors.New("folder is already enabled by the command line")
		}
	}
	return s.Repository.Grant(ctx, canonical)
}

type grant struct {
	ID        string `gorm:"primaryKey"`
	MachineID string
	UserID    string
	Path      string
	CreatedAt time.Time
}

func (grant) TableName() string { return "workspace_access_grants" }
