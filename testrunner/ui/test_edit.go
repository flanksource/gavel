package testui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	goparser "go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/utils"
)

const (
	testEditActionSkip   = "skip"
	testEditActionDelete = "delete"
	testEditScopeTest    = "test"
	testEditScopeFile    = "file"
)

var errEditAmbiguous = errors.New("test edit target is ambiguous")

type TestEditRequest struct {
	Action      string   `json:"action"`
	Scope       string   `json:"scope"`
	Framework   string   `json:"framework,omitempty"`
	WorkDir     string   `json:"work_dir,omitempty"`
	PackagePath string   `json:"package_path,omitempty"`
	File        string   `json:"file,omitempty"`
	Line        int      `json:"line,omitempty"`
	TestName    string   `json:"test_name,omitempty"`
	Suite       []string `json:"suite,omitempty"`
}

type TestEditResponse struct {
	File    string `json:"file"`
	Action  string `json:"action"`
	Scope   string `json:"scope"`
	Changed bool   `json:"changed"`
	Edited  int    `json:"edited,omitempty"`
	Removed int    `json:"removed,omitempty"`
	Message string `json:"message,omitempty"`
}

func (s *Server) handleTestEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TestEditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := req.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	root := s.resolveTestEditRootLocked(req.WorkDir)
	if root == "" {
		s.mu.Unlock()
		http.Error(w, "test edit not supported", http.StatusNotImplemented)
		return
	}
	target, err := resolveTestEditFile(root, req)
	if err != nil {
		s.mu.Unlock()
		writeTestEditError(w, err)
		return
	}
	resp, err := applyTestEdit(target, req)
	if err == nil {
		resp.File = displayEditFile(root, target)
		s.tests = applyEditToTests(s.tests, req)
	}
	s.mu.Unlock()
	if err != nil {
		writeTestEditError(w, err)
		return
	}

	s.notify()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (r TestEditRequest) validate() error {
	switch r.Action {
	case testEditActionSkip, testEditActionDelete:
	default:
		return fmt.Errorf("action must be %q or %q", testEditActionSkip, testEditActionDelete)
	}
	switch r.Scope {
	case testEditScopeTest, testEditScopeFile:
	default:
		return fmt.Errorf("scope must be %q or %q", testEditScopeTest, testEditScopeFile)
	}
	switch parsers.Framework(r.Framework) {
	case parsers.GoTest, parsers.Ginkgo, parsers.Vitest:
	default:
		return fmt.Errorf("framework %q is not editable", r.Framework)
	}
	if r.File == "" {
		return fmt.Errorf("file is required")
	}
	if r.Scope == testEditScopeTest && r.TestName == "" && r.Line <= 0 {
		return fmt.Errorf("test_name or line is required for test scope")
	}
	return nil
}

func (s *Server) testEditSupportedLocked() bool {
	if s.replayed {
		return false
	}
	if s.gitRoot != "" || (s.git != nil && s.git.Root != "") {
		return true
	}
	return testsHaveWorkDir(s.tests)
}

func (s *Server) resolveTestEditRootLocked(workDir string) string {
	if s.replayed {
		return ""
	}
	if workDir != "" {
		if root := cleanExistingRoot(workDir); root != "" {
			return root
		}
		return ""
	}
	if s.gitRoot != "" {
		if root := cleanExistingRoot(s.gitRoot); root != "" {
			return root
		}
	}
	if s.git != nil && s.git.Root != "" {
		if root := cleanExistingRoot(s.git.Root); root != "" {
			s.gitRoot = root
			return root
		}
	}
	return ""
}

func cleanExistingRoot(path string) string {
	root := utils.FindGitRoot(path)
	if root == "" {
		root = path
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return ""
	}
	return filepath.Clean(abs)
}

func testsHaveWorkDir(tests []parsers.Test) bool {
	for _, test := range tests {
		if test.WorkDir != "" {
			return true
		}
		if testsHaveWorkDir(test.Children) {
			return true
		}
	}
	return false
}

func resolveTestEditFile(root string, req TestEditRequest) (string, error) {
	base := root
	if req.WorkDir != "" {
		base = req.WorkDir
	}
	if abs, err := filepath.Abs(base); err == nil {
		base = filepath.Clean(abs)
	}

	file := filepath.Clean(req.File)
	if filepath.IsAbs(file) {
		file = filepath.Clean(file)
	} else {
		file = filepath.Join(base, file)
	}
	if !pathInside(root, file) {
		if req.WorkDir != "" && !filepath.IsAbs(req.File) {
			alt := filepath.Join(root, filepath.Clean(req.File))
			if pathInside(root, alt) {
				file = alt
			}
		}
	}
	if !pathInside(root, file) {
		return "", fmt.Errorf("file %q is outside edit root", req.File)
	}
	info, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return "", os.ErrNotExist
		}
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("file %q is a directory", req.File)
	}
	return file, nil
}

func pathInside(root, path string) bool {
	root, _ = filepath.Abs(root)
	path, _ = filepath.Abs(path)
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..")
}

func displayEditFile(root, target string) string {
	if rel, err := filepath.Rel(root, target); err == nil {
		return filepath.ToSlash(rel)
	}
	return target
}

func writeTestEditError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, os.ErrNotExist):
		http.Error(w, "test file not found", http.StatusNotFound)
	case errors.Is(err, errEditAmbiguous):
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func applyTestEdit(path string, req TestEditRequest) (TestEditResponse, error) {
	resp := TestEditResponse{Action: req.Action, Scope: req.Scope}
	if req.Scope == testEditScopeFile && req.Action == testEditActionDelete {
		if err := os.Remove(path); err != nil {
			return resp, err
		}
		resp.Changed = true
		resp.Removed = 1
		resp.Message = "file deleted"
		return resp, nil
	}

	switch parsers.Framework(req.Framework) {
	case parsers.GoTest:
		return editGoTestFile(path, req)
	case parsers.Ginkgo:
		return editGinkgoFile(path, req)
	case parsers.Vitest:
		return editVitestFile(path, req)
	default:
		return resp, fmt.Errorf("framework %q is not editable", req.Framework)
	}
}

func editGoTestFile(path string, req TestEditRequest) (TestEditResponse, error) {
	resp := TestEditResponse{Action: req.Action, Scope: req.Scope}
	src, err := os.ReadFile(path)
	if err != nil {
		return resp, err
	}
	fset := token.NewFileSet()
	file, err := goparser.ParseFile(fset, path, src, goparser.ParseComments)
	if err != nil {
		return resp, fmt.Errorf("parse go test file: %w", err)
	}

	if req.Scope == testEditScopeFile && req.Action == testEditActionSkip {
		edited := 0
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !isGoTestFunc(fn.Name.Name) || fn.Body == nil {
				continue
			}
			if prependSkipToBlock(fn.Body, testParamName(fn)) {
				edited++
			}
		}
		if edited == 0 {
			return resp, fmt.Errorf("%w: no tests found in %s", errEditAmbiguous, filepath.Base(path))
		}
		resp.Edited = edited
		resp.Changed = true
		return writeFormattedGo(path, fset, file, resp)
	}

	parent, subtests := splitGoTestName(req.TestName)
	fn := findGoTestFunc(fset, file, parent, req.Line)
	if fn == nil {
		return resp, os.ErrNotExist
	}
	if len(subtests) == 0 {
		switch req.Action {
		case testEditActionSkip:
			if fn.Body == nil || !prependSkipToBlock(fn.Body, testParamName(fn)) {
				return resp, fmt.Errorf("%w: unable to insert skip", errEditAmbiguous)
			}
			resp.Edited = 1
			resp.Changed = true
			return writeFormattedGo(path, fset, file, resp)
		case testEditActionDelete:
			file.Decls = removeDecl(file.Decls, fn)
			resp.Removed = 1
			resp.Changed = true
			return writeFormattedGo(path, fset, file, resp)
		}
	}

	body := fn.Body
	tName := testParamName(fn)
	for i, part := range subtests {
		idx, call, lit := findSubtestCall(body, tName, part)
		if idx < 0 || call == nil || lit == nil || lit.Body == nil {
			return resp, fmt.Errorf("%w: subtest %q not found", errEditAmbiguous, strings.Join(subtests[:i+1], "/"))
		}
		if i == len(subtests)-1 {
			switch req.Action {
			case testEditActionSkip:
				if !prependSkipToBlock(lit.Body, testParamNameFromFuncLit(lit, tName)) {
					return resp, fmt.Errorf("%w: unable to insert subtest skip", errEditAmbiguous)
				}
				resp.Edited = 1
			case testEditActionDelete:
				body.List = append(body.List[:idx], body.List[idx+1:]...)
				resp.Removed = 1
			}
			resp.Changed = true
			return writeFormattedGo(path, fset, file, resp)
		}
		body = lit.Body
		tName = testParamNameFromFuncLit(lit, tName)
	}
	return resp, fmt.Errorf("%w: subtest target not found", errEditAmbiguous)
}

func isGoTestFunc(name string) bool {
	return name != "TestMain" && (strings.HasPrefix(name, "Test") || strings.HasPrefix(name, "Benchmark"))
}

func splitGoTestName(name string) (string, []string) {
	parts := strings.Split(name, "/")
	if len(parts) == 0 {
		return name, nil
	}
	return parts[0], parts[1:]
}

func findGoTestFunc(fset *token.FileSet, file *ast.File, name string, line int) *ast.FuncDecl {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !isGoTestFunc(fn.Name.Name) {
			continue
		}
		if name != "" && fn.Name.Name == name {
			return fn
		}
		if name == "" && line > 0 {
			start := fset.Position(fn.Pos()).Line
			end := fset.Position(fn.End()).Line
			if line >= start && line <= end {
				return fn
			}
		}
	}
	return nil
}

func testParamName(fn *ast.FuncDecl) string {
	if fn.Type != nil && fn.Type.Params != nil && len(fn.Type.Params.List) > 0 && len(fn.Type.Params.List[0].Names) > 0 {
		return fn.Type.Params.List[0].Names[0].Name
	}
	return "t"
}

func testParamNameFromFuncLit(lit *ast.FuncLit, fallback string) string {
	if lit.Type != nil && lit.Type.Params != nil && len(lit.Type.Params.List) > 0 && len(lit.Type.Params.List[0].Names) > 0 {
		return lit.Type.Params.List[0].Names[0].Name
	}
	return fallback
}

func prependSkipToBlock(block *ast.BlockStmt, param string) bool {
	if block == nil || param == "" {
		return false
	}
	if len(block.List) > 0 && isSkipStmt(block.List[0], param) {
		return true
	}
	stmt := &ast.ExprStmt{X: &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(param), Sel: ast.NewIdent("Skip")},
		Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: strconvQuote("skipped by gavel")}},
	}}
	block.List = append([]ast.Stmt{stmt}, block.List...)
	return true
}

func isSkipStmt(stmt ast.Stmt, param string) bool {
	expr, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := expr.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Skip" {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == param
}

func strconvQuote(s string) string {
	return strconv.Quote(s)
}

func findSubtestCall(block *ast.BlockStmt, receiver, title string) (int, *ast.CallExpr, *ast.FuncLit) {
	if block == nil {
		return -1, nil, nil
	}
	for i, stmt := range block.List {
		expr, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}
		call, ok := expr.X.(*ast.CallExpr)
		if !ok || !isSelectorCall(call.Fun, receiver, "Run") || len(call.Args) < 2 {
			continue
		}
		if stringLiteral(call.Args[0]) != title {
			continue
		}
		lit, ok := call.Args[1].(*ast.FuncLit)
		if !ok {
			return i, call, nil
		}
		return i, call, lit
	}
	return -1, nil, nil
}

func isSelectorCall(expr ast.Expr, receiver, method string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != method {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == receiver
}

func stringLiteral(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	value := strings.Trim(lit.Value, "`\"")
	return value
}

func removeDecl(decls []ast.Decl, target ast.Decl) []ast.Decl {
	for i, decl := range decls {
		if decl == target {
			return append(decls[:i], decls[i+1:]...)
		}
	}
	return decls
}

func writeFormattedGo(path string, fset *token.FileSet, file *ast.File, resp TestEditResponse) (TestEditResponse, error) {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return resp, fmt.Errorf("format go source: %w", err)
	}
	if err := writeFilePreserveMode(path, buf.Bytes()); err != nil {
		return resp, err
	}
	return resp, nil
}

func editGinkgoFile(path string, req TestEditRequest) (TestEditResponse, error) {
	resp := TestEditResponse{Action: req.Action, Scope: req.Scope}
	src, err := os.ReadFile(path)
	if err != nil {
		return resp, err
	}
	fset := token.NewFileSet()
	file, err := goparser.ParseFile(fset, path, src, goparser.ParseComments)
	if err != nil {
		return resp, fmt.Errorf("parse ginkgo file: %w", err)
	}

	changed := 0
	if req.Scope == testEditScopeFile {
		if req.Action != testEditActionSkip {
			return resp, fmt.Errorf("%w: ginkgo file delete should remove the file directly", errEditAmbiguous)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if makeGinkgoPending(call) {
				changed++
			}
			return true
		})
		if changed == 0 {
			return resp, fmt.Errorf("%w: no ginkgo specs found", errEditAmbiguous)
		}
		resp.Edited = changed
		resp.Changed = true
		return writeFormattedGo(path, fset, file, resp)
	}

	var target *ast.CallExpr
	matches := 0
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isGinkgoLeafCall(call) {
			return true
		}
		pos := fset.Position(call.Pos())
		end := fset.Position(call.End())
		name := ""
		if len(call.Args) > 0 {
			name = stringLiteral(call.Args[0])
		}
		if req.TestName != "" && name != "" && name != req.TestName {
			return true
		}
		if req.Line > 0 && (req.Line < pos.Line || req.Line > end.Line) {
			return true
		}
		matches++
		target = call
		return true
	})
	if target == nil || matches == 0 {
		return resp, fmt.Errorf("%w: ginkgo spec not found", errEditAmbiguous)
	}
	if matches > 1 {
		return resp, fmt.Errorf("%w: multiple ginkgo specs matched", errEditAmbiguous)
	}
	switch req.Action {
	case testEditActionSkip:
		if !makeGinkgoPending(target) {
			return resp, fmt.Errorf("%w: unsupported ginkgo call", errEditAmbiguous)
		}
		resp.Edited = 1
	case testEditActionDelete:
		if !deleteGinkgoStmt(file, target) {
			return resp, fmt.Errorf("%w: unable to delete ginkgo call", errEditAmbiguous)
		}
		resp.Removed = 1
	}
	resp.Changed = true
	return writeFormattedGo(path, fset, file, resp)
}

func isGinkgoLeafCall(call *ast.CallExpr) bool {
	name, _ := ginkgoCallName(call.Fun)
	switch strings.TrimPrefix(name, "F") {
	case "It", "Specify", "Entry":
		return true
	}
	return name == "PIt" || name == "PSpecify" || name == "PEntry"
}

func makeGinkgoPending(call *ast.CallExpr) bool {
	name, prefix := ginkgoCallName(call.Fun)
	pending, ok := pendingGinkgoName(name)
	if !ok {
		return false
	}
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		fun.Name = pending
		return true
	case *ast.SelectorExpr:
		if prefix != "" {
			fun.Sel.Name = pending
			return true
		}
	}
	return false
}

func pendingGinkgoName(name string) (string, bool) {
	switch name {
	case "It", "FIt", "PIt":
		return "PIt", true
	case "Specify", "FSpecify", "PSpecify":
		return "PSpecify", true
	case "Entry", "FEntry", "PEntry":
		return "PEntry", true
	}
	return "", false
}

func ginkgoCallName(expr ast.Expr) (name string, prefix string) {
	switch fun := expr.(type) {
	case *ast.Ident:
		return fun.Name, ""
	case *ast.SelectorExpr:
		id, _ := fun.X.(*ast.Ident)
		if id == nil {
			return fun.Sel.Name, ""
		}
		return fun.Sel.Name, id.Name
	}
	return "", ""
}

func deleteGinkgoStmt(file *ast.File, target *ast.CallExpr) bool {
	deleted := false
	ast.Inspect(file, func(n ast.Node) bool {
		if deleted {
			return false
		}
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for i, stmt := range block.List {
			if expr, ok := stmt.(*ast.ExprStmt); ok && expr.X == target {
				block.List = append(block.List[:i], block.List[i+1:]...)
				deleted = true
				return false
			}
		}
		return true
	})
	return deleted
}

func editVitestFile(path string, req TestEditRequest) (TestEditResponse, error) {
	resp := TestEditResponse{Action: req.Action, Scope: req.Scope}
	srcBytes, err := os.ReadFile(path)
	if err != nil {
		return resp, err
	}
	src := string(srcBytes)
	var ranges []vitestCallRange
	if req.Scope == testEditScopeFile {
		ranges = findVitestCalls(src, 0, "")
	} else {
		ranges = findVitestCalls(src, req.Line, req.TestName)
	}
	if len(ranges) == 0 {
		return resp, fmt.Errorf("%w: vitest target not found", errEditAmbiguous)
	}
	if req.Scope == testEditScopeTest && len(ranges) > 1 {
		return resp, fmt.Errorf("%w: multiple vitest targets matched", errEditAmbiguous)
	}

	switch req.Action {
	case testEditActionSkip:
		for i := len(ranges) - 1; i >= 0; i-- {
			r := ranges[i]
			if r.HasSkip {
				continue
			}
			src = src[:r.NameStart] + vitestSkipCallee(src[r.NameStart:r.NameEnd]) + src[r.NameEnd:]
			resp.Edited++
		}
	case testEditActionDelete:
		for i := len(ranges) - 1; i >= 0; i-- {
			r := ranges[i]
			src = src[:r.DeleteStart] + src[r.DeleteEnd:]
			resp.Removed++
		}
	}
	if err := writeFilePreserveMode(path, []byte(src)); err != nil {
		return resp, err
	}
	resp.Changed = true
	return resp, nil
}

func writeFilePreserveMode(path string, data []byte) error {
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	return os.WriteFile(path, data, mode)
}

type vitestCallRange struct {
	DeleteStart int
	DeleteEnd   int
	NameStart   int
	NameEnd     int
	LineStart   int
	LineEnd     int
	Title       string
	HasSkip     bool
}

var vitestCallRe = regexp.MustCompile(`(?m)(^|[^\w$])((?:describe|test|it)(?:\.(?:only|skip|concurrent|each|todo|fails))*)(\s*\()`)

func findVitestCalls(src string, line int, title string) []vitestCallRange {
	var out []vitestCallRange
	matches := vitestCallRe.FindAllStringSubmatchIndex(src, -1)
	for _, m := range matches {
		nameStart := m[4]
		nameEnd := m[5]
		open := m[6] + strings.Index(src[m[6]:m[7]], "(")
		end := findBalancedParenEnd(src, open)
		if end < 0 {
			continue
		}
		callText := src[nameStart:nameEnd]
		if unsupportedVitestCallee(callText) {
			continue
		}
		firstTitle := firstVitestTitle(src[open+1 : end-1])
		if title != "" && firstTitle != "" && firstTitle != title {
			continue
		}
		lineStart := 1 + strings.Count(src[:nameStart], "\n")
		lineEnd := 1 + strings.Count(src[:end], "\n")
		if line > 0 && (line < lineStart || line > lineEnd) {
			continue
		}
		delStart := lineStartOffset(src, nameStart)
		delEnd := consumeTrailingWhitespaceLine(src, end)
		out = append(out, vitestCallRange{
			DeleteStart: delStart,
			DeleteEnd:   delEnd,
			NameStart:   nameStart,
			NameEnd:     nameEnd,
			LineStart:   lineStart,
			LineEnd:     lineEnd,
			Title:       firstTitle,
			HasSkip:     strings.Contains(callText, ".skip"),
		})
	}
	return out
}

func unsupportedVitestCallee(callee string) bool {
	for _, part := range strings.Split(callee, ".")[1:] {
		switch part {
		case "only", "skip", "concurrent":
			continue
		default:
			return true
		}
	}
	return false
}

func vitestSkipCallee(callee string) string {
	parts := strings.Split(callee, ".")
	if len(parts) == 0 {
		return callee + ".skip"
	}
	base := parts[0]
	var modifiers []string
	hasSkip := false
	for _, part := range parts[1:] {
		switch part {
		case "skip":
			hasSkip = true
		case "only":
			continue
		default:
			modifiers = append(modifiers, part)
		}
	}
	if !hasSkip {
		modifiers = append(modifiers, "skip")
	}
	return base + "." + strings.Join(modifiers, ".")
}

func firstVitestTitle(args string) string {
	args = strings.TrimSpace(args)
	if args == "" {
		return ""
	}
	quote := args[0]
	if quote != '\'' && quote != '"' && quote != '`' {
		return ""
	}
	var b strings.Builder
	escaped := false
	for i := 1; i < len(args); i++ {
		ch := args[i]
		if escaped {
			b.WriteByte(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == quote {
			return b.String()
		}
		b.WriteByte(ch)
	}
	return ""
}

func findBalancedParenEnd(src string, open int) int {
	depth := 0
	var quote byte
	escaped := false
	for i := open; i < len(src); i++ {
		ch := src[i]
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == quote {
				quote = 0
			}
			continue
		}
		switch ch {
		case '\'', '"', '`':
			quote = ch
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				end := i + 1
				for end < len(src) && unicode.IsSpace(rune(src[end])) && src[end] != '\n' {
					end++
				}
				if end < len(src) && src[end] == ';' {
					end++
				}
				return end
			}
		}
	}
	return -1
}

func lineStartOffset(src string, offset int) int {
	if offset <= 0 {
		return 0
	}
	if idx := strings.LastIndex(src[:offset], "\n"); idx >= 0 {
		return idx + 1
	}
	return 0
}

func consumeTrailingWhitespaceLine(src string, offset int) int {
	for offset < len(src) && (src[offset] == ' ' || src[offset] == '\t' || src[offset] == '\r') {
		offset++
	}
	if offset < len(src) && src[offset] == '\n' {
		offset++
	}
	return offset
}

func applyEditToTests(tests []parsers.Test, req TestEditRequest) []parsers.Test {
	out := make([]parsers.Test, 0, len(tests))
	for _, t := range tests {
		if req.Action == testEditActionDelete && testMatchesEdit(t, req) {
			continue
		}
		if len(t.Children) > 0 {
			t.Children = applyEditToTests(t.Children, req)
			if req.Action == testEditActionDelete && req.Scope == testEditScopeFile && len(t.Children) == 0 && t.File == "" {
				continue
			}
		}
		if req.Action == testEditActionSkip && testMatchesEdit(t, req) {
			t.Passed = false
			t.Failed = false
			t.Pending = false
			t.Running = false
			t.Skipped = true
			t.Message = "skipped by gavel"
		}
		out = append(out, t)
	}
	return out
}

func testMatchesEdit(t parsers.Test, req TestEditRequest) bool {
	if req.Framework != "" && t.Framework.String() != req.Framework {
		return false
	}
	if req.File != "" && !sameCleanPath(t.File, req.File) {
		return false
	}
	if req.Scope == testEditScopeFile {
		return t.File != ""
	}
	if req.TestName != "" && t.Name != req.TestName {
		return false
	}
	if req.Line > 0 && t.Line > 0 && t.Line != req.Line {
		return false
	}
	return true
}

func sameCleanPath(a, b string) bool {
	return filepath.ToSlash(filepath.Clean(a)) == filepath.ToSlash(filepath.Clean(b))
}
