// Package settings resolves a target's gavel configuration from its layers,
// the way captain composes spec layers: every layer applies in order, nearest
// last, field by field, with each configured field traced to its layer.
//
// A layer position is a directory in the target's .gavel.yaml chain
// (verify.GavelConfigDirs): the user's home, the git root's parent and other
// directories outside the project, the project root, and each sub-folder of the
// project down to the target. At a position, the .gavel.yaml and its
// .gavel.<name>.yaml snippets are the layer; in DB mode a position with no file
// may instead hold a layer authored in the database.
package settings

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
)

// Kind is what a layer position is to the target's project.
type Kind string

const (
	// KindUser is the user's home directory: the tenant's own layer.
	KindUser Kind = "user"
	// KindParent is a directory above the project other than home (the git
	// root's parent, or the git root when the project is beneath it). It is
	// read from disk only: it has no identity in the project.
	KindParent Kind = "parent"
	// KindProject is the project's root directory.
	KindProject Kind = "project"
	// KindPath is a sub-folder of the project.
	KindPath Kind = "path"
)

// Ref names one layer: a position, and the file at it ("" for the layer
// authored in the database).
type Ref struct {
	Kind Kind `json:"kind"`
	// Path is the slash-separated folder relative to the project for KindPath,
	// and the absolute directory for KindParent.
	Path string `json:"path,omitempty"`
	// File is the .gavel.yaml or .gavel.<name>.yaml at the position, or "" for
	// the database layer.
	File string `json:"file,omitempty"`
}

// Layer is one configuration layer of a target, in the order it applies.
type Layer struct {
	Ref
	// Location is the file's absolute path, or "db" for a database layer.
	Location string `json:"location"`
	// GitSHA is the repository HEAD the file was synced at in DB mode ("none"
	// outside git); empty for a database layer and in file-only mode.
	GitSHA string             `json:"gitSha,omitempty"`
	Config verify.GavelConfig `json:"config"`
}

// Label is how provenance names the layer.
func (l Layer) Label() string {
	if l.File != "" {
		return l.Location
	}
	return "db:" + string(l.Kind) + ":" + l.Path
}

// Position is one directory of the target's chain and what can be saved there.
type Position struct {
	Kind Kind   `json:"kind"`
	Path string `json:"path,omitempty"`
	Dir  string `json:"dir"`
	// Files are the config files at the position, .gavel.yaml first.
	Files []string `json:"files"`
	// DB is whether the position takes a database layer: DB mode, no file at
	// the position, and a position with an identity (not KindParent).
	DB bool `json:"db"`
}

// Project is the project a target lies in. A web project (no git, nothing on
// disk) has no Dir: its layers below the user's are all database layers.
type Project struct {
	ID  uuid.UUID
	Dir string
}

// Target is what a configuration is resolved for, in the project holding it
// (nil outside every project): Path, a file or directory on disk, for a
// project with a Dir; Key, the target's slash-separated path in the project
// (e.g. pages/team), for a web project.
type Target struct {
	Path    string
	Key     string
	Project *Project
}

func (t Target) web() bool { return t.Project != nil && t.Project.Dir == "" }

// Resolved is a target's effective configuration and how it was composed.
type Resolved struct {
	Positions []Position         `json:"positions"`
	Layers    []Layer            `json:"layers"`
	Merged    verify.GavelConfig `json:"merged"`
	// Provenance maps each configured leaf (a JSON pointer) to the Label of the layer that set it.
	Provenance map[string]string `json:"provenance"`
}

// Service resolves and saves a target's configuration layers.
type Service interface {
	Resolve(ctx context.Context, target Target) (Resolved, error)
	// Save edits the layer ref names, which must be a position of target's
	// chain: a file there (created when absent), or its database layer.
	Save(ctx context.Context, target Target, ref Ref, edit func(*verify.GavelConfig) error) error
}

// diskChain is target's positions, lowest precedence first, and the file
// layers at them, read as verify reads them (.gavel.yaml then snippets by
// name, through its os.Root guard).
func diskChain(target Target) ([]Position, []Layer, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, fmt.Errorf("find the home directory: %w", err)
	}
	if target.web() {
		return webChain(target, home)
	}
	dir, err := targetDir(target.Path)
	if err != nil {
		return nil, nil, err
	}
	if target.Project != nil {
		if _, inside := within(target.Project.Dir, dir); !inside {
			return nil, nil, fmt.Errorf("%s is not inside its project's directory %s", target.Path, target.Project.Dir)
		}
	}
	var positions []Position
	var layers []Layer
	for _, layerDir := range chainDirs(dir, home, target.Project) {
		position := positionOf(layerDir, home, target.Project)
		position.Files = []string{}
		sources, err := verify.LoadGavelConfigDir(layerDir)
		if err != nil {
			return nil, nil, err
		}
		for _, source := range sources {
			name := filepath.Base(source.Path)
			position.Files = append(position.Files, name)
			layers = append(layers, Layer{Ref: Ref{Kind: position.Kind, Path: position.Path, File: name}, Location: source.Path, Config: source.Config})
		}
		positions = append(positions, position)
	}
	return positions, layers, nil
}

// chainDirs is dir's chain, lowest precedence first: verify's directories
// outside the project (home, the git root's parent, ...), then the project root
// and every folder down to dir, whether or not the project is a git checkout.
func chainDirs(dir, home string, project *Project) []string {
	var dirs []string
	for _, layerDir := range verify.GavelConfigDirs(dir) {
		if project != nil {
			if _, inside := within(project.Dir, layerDir.Dir); inside && layerDir.Dir != home {
				continue
			}
		}
		dirs = append(dirs, layerDir.Dir)
	}
	if project == nil {
		return dirs
	}
	var nested []string
	for current := dir; ; current = filepath.Dir(current) {
		nested = append(nested, current)
		if current == project.Dir {
			break
		}
	}
	for i := len(nested) - 1; i >= 0; i-- {
		if !slices.Contains(dirs, nested[i]) {
			dirs = append(dirs, nested[i])
		}
	}
	return dirs
}

// webChain is a web project target's positions: the user's home (with its
// files), the project, then each ancestor of the target's key, the last two
// kinds holding database layers only.
func webChain(target Target, home string) ([]Position, []Layer, error) {
	if !verify.IsCommentPageKey(target.Key) {
		return nil, nil, fmt.Errorf("web target key %q must be a clean relative slash-separated path", target.Key)
	}
	homeConfig, err := verify.LoadGavelConfigTrace(home)
	if err != nil {
		return nil, nil, err
	}
	user := Position{Kind: KindUser, Dir: home, Files: []string{}}
	var layers []Layer
	for _, source := range homeConfig.Sources {
		if filepath.Dir(source.Path) == home {
			user.Files = append(user.Files, filepath.Base(source.Path))
			layers = append(layers, Layer{Ref: Ref{Kind: KindUser, File: filepath.Base(source.Path)}, Location: source.Path, Config: source.Config})
		}
	}
	positions := []Position{user, {Kind: KindProject, Dir: "web:", Files: []string{}}}
	segments := strings.Split(target.Key, "/")
	for i := 1; i < len(segments); i++ {
		path := strings.Join(segments[:i], "/")
		positions = append(positions, Position{Kind: KindPath, Path: path, Dir: "web:" + path, Files: []string{}})
	}
	return positions, layers, nil
}

func positionOf(dir, home string, project *Project) Position {
	if project != nil {
		if rel, ok := within(project.Dir, dir); ok {
			if rel == "." {
				return Position{Kind: KindProject, Dir: dir}
			}
			return Position{Kind: KindPath, Path: filepath.ToSlash(rel), Dir: dir}
		}
	}
	if dir == home {
		return Position{Kind: KindUser, Dir: dir}
	}
	return Position{Kind: KindParent, Path: dir, Dir: dir}
}

// positionFor is the position of target's chain ref names, or an error when
// ref names a directory outside the chain: a save can never write elsewhere.
func positionFor(positions []Position, ref Ref) (Position, error) {
	for _, position := range positions {
		if position.Kind == ref.Kind && position.Path == ref.Path {
			return position, nil
		}
	}
	return Position{}, fmt.Errorf("%s layer %q is not in the target's configuration chain", ref.Kind, ref.Path)
}

func targetDir(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", path, err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return absolute, nil
	}
	return filepath.Dir(absolute), nil
}

func within(root, path string) (string, bool) {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return rel, true
}

// compose merges layers in order with verify's structural merge and traces each field to its layer.
func compose(positions []Position, layers []Layer) (Resolved, error) {
	sources := make([]verify.GavelConfigSource, len(layers))
	for i, layer := range layers {
		sources[i] = verify.GavelConfigSource{Origin: string(layer.Kind), Path: layer.Label(), Config: layer.Config}
	}
	merged, provenance, err := verify.ComposeGavelConfig(sources)
	if err != nil {
		return Resolved{}, err
	}
	return Resolved{Positions: positions, Layers: layers, Merged: merged, Provenance: provenance}, nil
}
