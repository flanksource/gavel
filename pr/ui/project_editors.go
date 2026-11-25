package ui

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/flanksource/commons/logger"
)

type projectEditor struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type projectOpenRequest struct {
	Path   string `json:"path"`
	Editor string `json:"editor"`
	Line   int    `json:"line,omitempty"`
}

// editorCandidate is an editor gavel knows how to launch: through its CLI,
// which can jump to a line, or on darwin through its application bundle.
type editorCandidate struct {
	projectEditor
	cli  string
	app  string
	args func(file string, line int) []string
}

func gotoArgs(file string, line int) []string  { return []string{"-g", fileAtLine(file, line)} }
func colonArgs(file string, line int) []string { return []string{fileAtLine(file, line)} }
func jetbrainsArgs(file string, line int) []string {
	if line == 0 {
		return []string{file}
	}
	return []string{"--line", strconv.Itoa(line), file}
}

func fileAtLine(file string, line int) string {
	if line == 0 {
		return file
	}
	return file + ":" + strconv.Itoa(line)
}

var editorCandidates = []editorCandidate{
	{projectEditor: projectEditor{ID: "vscode", Label: "VS Code"}, cli: "code", app: "Visual Studio Code", args: gotoArgs},
	{projectEditor: projectEditor{ID: "cursor", Label: "Cursor"}, cli: "cursor", app: "Cursor", args: gotoArgs},
	{projectEditor: projectEditor{ID: "windsurf", Label: "Windsurf"}, cli: "windsurf", app: "Windsurf", args: gotoArgs},
	{projectEditor: projectEditor{ID: "zed", Label: "Zed"}, cli: "zed", app: "Zed", args: colonArgs},
	{projectEditor: projectEditor{ID: "sublime", Label: "Sublime Text"}, cli: "subl", app: "Sublime Text", args: colonArgs},
	{projectEditor: projectEditor{ID: "goland", Label: "GoLand"}, cli: "goland", app: "GoLand", args: jetbrainsArgs},
	{projectEditor: projectEditor{ID: "idea", Label: "IntelliJ IDEA"}, cli: "idea", app: "IntelliJ IDEA", args: jetbrainsArgs},
}

var (
	editorLookPath = exec.LookPath
	editorGOOS     = runtime.GOOS
	editorAppDirs  = func() []string {
		dirs := []string{"/Applications"}
		if home, err := os.UserHomeDir(); err == nil {
			dirs = append(dirs, filepath.Join(home, "Applications"))
		}
		return dirs
	}
	launchEditor = startDetached
)

// editorLaunch is how a detected editor starts: its CLI path, or the darwin
// application bundle name opened with `open -a`.
type editorLaunch struct {
	cli string
	app string
}

func detectEditor(candidate editorCandidate) (editorLaunch, bool) {
	if path, err := editorLookPath(candidate.cli); err == nil {
		return editorLaunch{cli: path}, true
	}
	if editorGOOS != "darwin" {
		return editorLaunch{}, false
	}
	for _, dir := range editorAppDirs() {
		if info, err := os.Stat(filepath.Join(dir, candidate.app+".app")); err == nil && info.IsDir() {
			return editorLaunch{app: candidate.app}, true
		}
	}
	return editorLaunch{}, false
}

func (s *Server) handleEditors(w http.ResponseWriter, _ *http.Request) {
	editors := []projectEditor{}
	for _, candidate := range editorCandidates {
		if _, ok := detectEditor(candidate); ok {
			editors = append(editors, candidate.projectEditor)
		}
	}
	respondJSON(w, http.StatusOK, editors)
}

// handleProjectOpen opens a project file in an editor detected on this host.
func (s *Server) handleProjectOpen(w http.ResponseWriter, r *http.Request) {
	project, err := getProject(r.Context(), r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	workDir, ok := s.requestWorkDir(w, r, project)
	if !ok {
		return
	}
	var request projectOpenRequest
	if err := decodeStrict(r, &request); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	name, args, err := editorCommand(workDir, request)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := launchEditor(name, args); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("launch %s: %v", request.Editor, err))
		return
	}
	respondJSON(w, http.StatusOK, projectOKResponse{OK: true})
}

func editorCommand(workDir string, request projectOpenRequest) (string, []string, error) {
	if request.Line < 0 {
		return "", nil, fmt.Errorf("line %d must not be negative", request.Line)
	}
	var candidate *editorCandidate
	for i := range editorCandidates {
		if editorCandidates[i].ID == request.Editor {
			candidate = &editorCandidates[i]
		}
	}
	if candidate == nil {
		return "", nil, fmt.Errorf("unknown editor %q", request.Editor)
	}
	launch, ok := detectEditor(*candidate)
	if !ok {
		return "", nil, fmt.Errorf("%q is not installed on this host", request.Editor)
	}
	file, err := projectFileAbsolutePath(workDir, request.Path)
	if err != nil {
		return "", nil, err
	}
	if launch.cli != "" {
		return launch.cli, candidate.args(file, request.Line), nil
	}
	return "open", []string{"-a", launch.app, file}, nil
}

// projectFileAbsolutePath resolves a project-relative path to the existing
// file it names, refusing one whose symlinks lead out of the project.
func projectFileAbsolutePath(workDir, requested string) (string, error) {
	cleaned, err := normalizeProjectPath(requested)
	if err != nil {
		return "", err
	}
	root, err := filepath.EvalSymlinks(workDir)
	if err != nil {
		return "", fmt.Errorf("resolve project directory: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(filepath.Join(root, filepath.FromSlash(cleaned)))
	if err != nil {
		return "", fmt.Errorf("%q: %w", requested, err)
	}
	if resolved != root && !strings.HasPrefix(resolved, root+string(filepath.Separator)) {
		return "", fmt.Errorf("%q resolves outside the project", requested)
	}
	return resolved, nil
}

// startDetached starts an editor in its own session so it outlives the gavel
// server, and reaps it in the background without making the request wait.
func startDetached(name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = detachedSysProcAttr()
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			logger.Warnf("editor %s %s exited: %v", name, strings.Join(args, " "), err)
		}
	}()
	return nil
}
