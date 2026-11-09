// Package uirplan inventories ordinary Go tests and calls UIR's producer-neutral affected-test planner.
package uirplan

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/gavel/internal/changegraph"
	"github.com/flanksource/gavel/testrunner/uirexport"
)

const ContractVersion = 1

type Test struct {
	ID        string   `json:"id"`
	Framework string   `json:"framework"`
	Package   string   `json:"package"`
	Suite     []string `json:"suite,omitempty"`
	Name      string   `json:"name"`
	File      string   `json:"file,omitempty"`
	Line      int      `json:"line,omitempty"`
	Target    string   `json:"-"`
}

type InventoryResult struct {
	Tests   []Test
	Targets []string
}

type Request struct {
	Version      int               `json:"version"`
	GitCommit    string            `json:"git_commit"`
	Dirty        bool              `json:"dirty,omitempty"`
	GitTree      string            `json:"git_tree,omitempty"`
	Environment  map[string]string `json:"environment"`
	InputsDigest string            `json:"inputs_digest"`
	InputContext string            `json:"input_context"`
	Inventory    []Test            `json:"inventory"`
}

type Response struct {
	Version         int      `json:"version"`
	PlanID          string   `json:"plan_id"`
	SnapshotID      string   `json:"snapshot_id"`
	InventoryDigest string   `json:"inventory_digest"`
	Mode            string   `json:"mode"`
	Reason          string   `json:"reason,omitempty"`
	BaselineRunID   *string  `json:"baseline_run_id,omitempty"`
	Skippable       []string `json:"skippable"`
}

// Inventory returns every build-selected top-level ordinary Go test in the current module.
func Inventory(workDir string, tags []string) (InventoryResult, error) {
	graph, err := changegraph.Load(workDir, tags)
	if err != nil {
		return InventoryResult{}, err
	}
	moduleRoot, err := moduleRoot(workDir)
	if err != nil {
		return InventoryResult{}, err
	}
	var inventory []Test
	targets := map[string]bool{}
	seen := map[string]bool{}
	for _, pkg := range graph.Packages {
		if len(pkg.TestGoFiles)+len(pkg.XTestGoFiles) == 0 {
			continue
		}
		targetRel, err := filepath.Rel(workDir, pkg.Dir)
		if err != nil || !filepath.IsLocal(targetRel) {
			continue
		}
		target := "./" + filepath.ToSlash(targetRel)
		targets[target] = true
		if pkg.Module == nil || filepath.Clean(pkg.Module.Dir) != filepath.Clean(moduleRoot) {
			continue
		}
		for _, name := range append(append([]string(nil), pkg.TestGoFiles...), pkg.XTestGoFiles...) {
			path := filepath.Join(pkg.Dir, name)
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return InventoryResult{}, fmt.Errorf("parse %s: %w", path, err)
			}
			rel, err := filepath.Rel(moduleRoot, path)
			if err != nil {
				return InventoryResult{}, err
			}
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || !ordinaryTest(fn) || containsRunSpecs(fn.Body) {
					continue
				}
				id := "go test\x00" + pkg.ImportPath + "\x00" + fn.Name.Name
				if seen[id] {
					return InventoryResult{}, fmt.Errorf("ordinary test %s %s is declared more than once", pkg.ImportPath, fn.Name.Name)
				}
				seen[id] = true
				inventory = append(inventory, Test{ID: id, Framework: "go test", Package: pkg.ImportPath, Name: fn.Name.Name,
					File: filepath.ToSlash(rel), Line: fset.Position(fn.Pos()).Line, Target: target})
			}
		}
	}
	sort.Slice(inventory, func(i, j int) bool { return inventory[i].ID < inventory[j].ID })
	result := InventoryResult{Tests: inventory}
	for target := range targets {
		result.Targets = append(result.Targets, target)
	}
	sort.Strings(result.Targets)
	return result, nil
}

func moduleRoot(workDir string) (string, error) {
	directory, err := filepath.Abs(workDir)
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil && !info.IsDir() {
			return directory, nil
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("no go.mod above %s", workDir)
		}
		directory = parent
	}
}

func ordinaryTest(fn *ast.FuncDecl) bool {
	if fn.Recv != nil || fn.Body == nil || fn.Name.Name == "TestMain" || !strings.HasPrefix(fn.Name.Name, "Test") || fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
		return false
	}
	star, ok := fn.Type.Params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := star.X.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "T"
}

func containsRunSpecs(body ast.Node) bool {
	found := false
	ast.Inspect(body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch function := call.Fun.(type) {
		case *ast.Ident:
			found = function.Name == "RunSpecs"
		case *ast.SelectorExpr:
			found = function.Sel.Name == "RunSpecs"
		}
		return !found
	})
	return found
}

func InputsDigest(filesDigest, inputContext string) string {
	digest := sha256.New()
	_, _ = digest.Write([]byte("gavel-uir-inputs-v1\x00"))
	_, _ = digest.Write([]byte(filesDigest))
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write([]byte(inputContext))
	return hex.EncodeToString(digest.Sum(nil))
}

func Execute(ctx context.Context, binary, workDir string, request Request) (Response, error) {
	if binary == "" {
		binary = "uir"
	}
	file, err := os.CreateTemp("", "gavel-uir-plan-*.json")
	if err != nil {
		return Response{}, err
	}
	name := file.Name()
	defer os.Remove(name)
	if err := json.NewEncoder(file).Encode(request); err != nil {
		file.Close()
		return Response{}, err
	}
	if err := file.Close(); err != nil {
		return Response{}, err
	}
	command := exec.CommandContext(ctx, binary, "--format", "json", "tests", "plan", name, "--location", workDir)
	command.Dir = workDir
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		return Response{}, fmt.Errorf("uir tests plan: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var response Response
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return Response{}, fmt.Errorf("decode uir tests plan output: %w: %s", err, strings.TrimSpace(stdout.String()))
	}
	if response.Version != ContractVersion || response.PlanID == "" || response.InventoryDigest == "" || (response.Mode != "run_all" && response.Mode != "selected") {
		return Response{}, fmt.Errorf("uir tests plan returned an invalid response")
	}
	known := map[string]bool{}
	for _, test := range request.Inventory {
		known[test.ID] = true
	}
	for _, id := range response.Skippable {
		if !known[id] {
			return Response{}, fmt.Errorf("uir tests plan returned unknown inventory id %q", id)
		}
	}
	return response, nil
}

func Provenance(request *Request, provenance uirexport.Provenance) {
	request.GitCommit, request.Dirty, request.GitTree = provenance.GitCommit, provenance.Dirty, provenance.GitTree
}
