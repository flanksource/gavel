package main

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/spf13/cobra"
)

// projectsCmd groups the CRUD subcommands for the workspace projects the PR
// dashboard tracks (the same store at ~/.config/gavel/projects.json that the
// /api/projects entity serves).
var projectsCmd = &cobra.Command{
	Use:     "projects",
	Aliases: []string{"project"},
	Short:   "Manage workspace projects (local directories the PR dashboard tracks)",
}

type ProjectsListOptions struct{}

// projectListRow is the display shape for `projects list`: the stored config
// plus the resolved todo backend (suffixed " (auto)" when auto-detected).
type projectListRow struct {
	Name        string   `json:"name"`
	Dir         string   `json:"dir"`
	Repos       []string `json:"repos"`
	TodoBackend string   `json:"todoBackend"`
}

func runProjectsList(_ ProjectsListOptions) ([]projectListRow, error) {
	ps := ui.LoadProjects()
	rows := make([]projectListRow, 0, len(ps))
	for _, p := range ps {
		rows = append(rows, projectListRow{
			Name:        p.Name,
			Dir:         p.Dir,
			Repos:       p.Repos,
			TodoBackend: ui.TodoBackendLabel(p.ResolvedDir(), p.TodoProvider),
		})
	}
	return rows, nil
}

type ProjectsGetOptions struct {
	Args []string `args:"true"`
}

func runProjectsGet(opts ProjectsGetOptions) (ui.Project, error) {
	if len(opts.Args) < 1 {
		return ui.Project{}, fmt.Errorf("usage: gavel projects get <name>")
	}
	p, ok := ui.GetProject(opts.Args[0])
	if !ok {
		return ui.Project{}, fmt.Errorf("%w: %q", ui.ErrProjectNotFound, opts.Args[0])
	}
	return p, nil
}

type ProjectsAddOptions struct {
	Args  []string `args:"true"`
	Repos []string `flag:"repo" help:"GitHub repo this directory backs (owner/repo); repeatable"`
	Todos string   `flag:"todos" help:"Todo backend: grite or todos (empty = auto-detect)"`
}

func runProjectsAdd(opts ProjectsAddOptions) (ui.Project, error) {
	if len(opts.Args) < 2 {
		return ui.Project{}, fmt.Errorf("usage: gavel projects add <name> <dir> [--repo owner/repo] [--todos grite|todos]")
	}
	p := ui.Project{Name: opts.Args[0], Dir: opts.Args[1], Repos: opts.Repos, TodoProvider: opts.Todos}
	if err := ui.CreateProject(p); err != nil {
		return ui.Project{}, err
	}
	return p, nil
}

type ProjectsUpdateOptions struct {
	Args  []string `args:"true"`
	Repos []string `flag:"repo" help:"Replace the repo list (owner/repo); repeatable"`
	Todos string   `flag:"todos" help:"Todo backend: grite or todos"`
}

func runProjectsUpdate(opts ProjectsUpdateOptions) (ui.Project, error) {
	if len(opts.Args) < 1 {
		return ui.Project{}, fmt.Errorf("usage: gavel projects update <name> [dir] [--repo owner/repo] [--todos grite|todos]")
	}
	name := opts.Args[0]
	p, ok := ui.GetProject(name)
	if !ok {
		return ui.Project{}, fmt.Errorf("%w: %q", ui.ErrProjectNotFound, name)
	}
	// Only the fields the caller supplied are changed; everything else keeps its
	// stored value (dir is kept unless a second positional is given).
	if len(opts.Args) >= 2 {
		p.Dir = opts.Args[1]
	}
	if len(opts.Repos) > 0 {
		p.Repos = opts.Repos
	}
	if opts.Todos != "" {
		p.TodoProvider = opts.Todos
	}
	if err := ui.UpdateProject(name, p); err != nil {
		return ui.Project{}, err
	}
	return p, nil
}

type ProjectsDeleteOptions struct {
	Args []string `args:"true"`
}

func runProjectsDelete(opts ProjectsDeleteOptions) (string, error) {
	if len(opts.Args) < 1 {
		return "", fmt.Errorf("usage: gavel projects delete <name>")
	}
	if err := ui.DeleteProject(opts.Args[0]); err != nil {
		return "", err
	}
	return fmt.Sprintf("deleted project %q", opts.Args[0]), nil
}

func init() {
	rootCmd.AddCommand(projectsCmd)
	clicky.AddNamedCommand("list", projectsCmd, ProjectsListOptions{}, runProjectsList).Aliases = []string{"ls"}
	clicky.AddNamedCommand("get", projectsCmd, ProjectsGetOptions{}, runProjectsGet)
	clicky.AddNamedCommand("add", projectsCmd, ProjectsAddOptions{}, runProjectsAdd).Aliases = []string{"create"}
	clicky.AddNamedCommand("update", projectsCmd, ProjectsUpdateOptions{}, runProjectsUpdate)
	clicky.AddNamedCommand("delete", projectsCmd, ProjectsDeleteOptions{}, runProjectsDelete).Aliases = []string{"rm"}
}
