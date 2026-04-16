package testui

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/clicky/formatters"
	_ "github.com/flanksource/clicky/formatters/html"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/bench"
	"github.com/flanksource/gavel/testrunner/parsers"
)

const (
	viewTabTests       = "tests"
	viewTabLint        = "lint"
	viewTabBench       = "bench"
	viewTabDiagnostics = "diagnostics"
)

type routeRequest struct {
	Tab         string
	NodePath    []string
	Format      string
	IsExport    bool
	TestFilters testRouteFilters
	LintFilters lintRouteFilters
}

type testRouteFilters struct {
	Status    []string `json:"status,omitempty"`
	Framework []string `json:"framework,omitempty"`
}

type lintRouteFilters struct {
	Grouping string   `json:"grouping,omitempty"`
	Severity []string `json:"severity,omitempty"`
	Linter   []string `json:"linter,omitempty"`
}

type exportReport struct {
	Tab      string      `json:"tab"`
	Path     string      `json:"path,omitempty"`
	Filters  any         `json:"filters,omitempty"`
	Selected *ViewNode   `json:"selected,omitempty"`
	Tests    []*ViewNode `json:"tests,omitempty"`
	Lint     []*ViewNode `json:"lint,omitempty"`
	Bench    []*ViewNode `json:"bench,omitempty"`
	Done     bool        `json:"done"`

	roots []*ViewNode `json:"-"`
}

type ViewNode struct {
	Name        string        `json:"name"`
	Path        string        `json:"path,omitempty"`
	Framework   string        `json:"framework,omitempty"`
	Kind        string        `json:"kind,omitempty"`
	Status      string        `json:"status,omitempty"`
	Package     string        `json:"package,omitempty"`
	PackagePath string        `json:"package_path,omitempty"`
	File        string        `json:"file,omitempty"`
	Line        int           `json:"line,omitempty"`
	Command     string        `json:"command,omitempty"`
	Message     string        `json:"message,omitempty"`
	Stdout      string        `json:"stdout,omitempty"`
	Stderr      string        `json:"stderr,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
	Children    []*ViewNode   `json:"children,omitempty"`
}

func (n *ViewNode) Pretty() api.Text {
	text := clicky.Text("")
	var style string

	switch n.Status {
	case "failed":
		text = text.Append(icons.Fail, "text-red-500")
		style = "text-red-500"
	case "skipped":
		text = text.Append(icons.Skip, "text-orange-500")
		style = "text-yellow-500"
	case "pending":
		text = text.Append(icons.Pending, "text-blue-500")
		style = "text-blue-500"
	default:
		text = text.Append(icons.Pass, "text-green-600")
		style = "text-green-600"
	}

	if n.File != "" && n.Line > 0 {
		text = text.Space().Append(fmt.Sprintf("%s:%d", n.File, n.Line), "text-muted")
	} else if n.File != "" {
		text = text.Space().Append(n.File, "text-muted")
	}

	text = text.Space().Append(n.Name, style)

	if n.Duration > 0 {
		text = text.Space().Append(n.Duration.String(), "text-muted")
	}

	if n.Message != "" {
		text = text.Space().Append(n.Message, style)
	}
	if n.Command != "" {
		text = text.NewLine().Add(api.Collapsed{
			Label:   "command",
			Content: clicky.Text(n.Command, "font-mono text-xs"),
		})
	}
	if n.Stdout != "" {
		text = text.NewLine().Add(api.Collapsed{
			Label:   "stdout",
			Content: clicky.Text(n.Stdout, "font-mono text-xs whitespace-pre-wrap"),
		})
	}
	if n.Stderr != "" {
		text = text.NewLine().Add(api.Collapsed{
			Label:   "stderr",
			Content: clicky.Text(n.Stderr, "font-mono text-xs whitespace-pre-wrap"),
		})
	}
	return text
}

func (n *ViewNode) GetChildren() []api.TreeNode {
	children := make([]api.TreeNode, 0, len(n.Children))
	for _, child := range n.Children {
		children = append(children, child)
	}
	return children
}

func (r exportReport) Pretty() api.Text {
	label := capitalize(r.Tab)
	text := clicky.Text(label+" export", "bold")
	if r.Path != "" {
		text = text.Space().Append("("+r.Path+")", "text-muted")
	}
	if r.Filters != nil {
		switch f := r.Filters.(type) {
		case testRouteFilters:
			if len(f.Status) > 0 || len(f.Framework) > 0 {
				text = text.Space().Append(renderTestFilters(f), "text-muted")
			}
		case lintRouteFilters:
			if filterText := renderLintFilters(f); filterText != "" {
				text = text.Space().Append(filterText, "text-muted")
			}
		}
	}
	return text
}

func (r exportReport) GetChildren() []api.TreeNode {
	if r.Selected != nil {
		return []api.TreeNode{r.Selected}
	}
	children := make([]api.TreeNode, 0, len(r.roots))
	for _, root := range r.roots {
		children = append(children, root)
	}
	return children
}

func renderTestFilters(f testRouteFilters) string {
	parts := make([]string, 0, 2)
	if len(f.Status) > 0 {
		parts = append(parts, "status="+strings.Join(f.Status, ","))
	}
	if len(f.Framework) > 0 {
		parts = append(parts, "framework="+strings.Join(f.Framework, ","))
	}
	return strings.Join(parts, " ")
}

func renderLintFilters(f lintRouteFilters) string {
	parts := make([]string, 0, 3)
	if f.Grouping != "" {
		parts = append(parts, "grouping="+f.Grouping)
	}
	if len(f.Severity) > 0 {
		parts = append(parts, "severity="+strings.Join(f.Severity, ","))
	}
	if len(f.Linter) > 0 {
		parts = append(parts, "linter="+strings.Join(f.Linter, ","))
	}
	return strings.Join(parts, " ")
}

func parseRouteRequest(r *http.Request) (routeRequest, bool) {
	req := routeRequest{
		Tab: viewTabTests,
		LintFilters: lintRouteFilters{
			Grouping: "linter-file",
		},
	}

	path := strings.Trim(r.URL.Path, "/")
	if path == "" {
		req.Format = routeRequestedFormat(r, "")
		req.IsExport = req.Format != ""
		req.TestFilters = parseTestFilters(r)
		req.LintFilters = parseLintFilters(r)
		return req, true
	}

	segments := strings.Split(path, "/")
	tabSeg := segments[0]
	pathFormat := ""
	if base, format := stripKnownFormat(tabSeg); format != "" {
		tabSeg = base
		pathFormat = format
	}

	switch tabSeg {
	case viewTabTests, viewTabLint, viewTabBench, viewTabDiagnostics:
		req.Tab = tabSeg
	default:
		return routeRequest{}, false
	}

	if len(segments) > 1 {
		req.NodePath = append(req.NodePath, segments[1:]...)
	}
	if len(req.NodePath) > 0 {
		last := req.NodePath[len(req.NodePath)-1]
		if base, format := stripKnownFormat(last); format != "" {
			req.NodePath[len(req.NodePath)-1] = base
			pathFormat = format
		}
	}

	req.Format = routeRequestedFormat(r, pathFormat)
	req.IsExport = req.Format != ""
	req.TestFilters = parseTestFilters(r)
	req.LintFilters = parseLintFilters(r)
	return req, true
}

func stripKnownFormat(segment string) (string, string) {
	ext := strings.ToLower(filepath.Ext(segment))
	if ext == "" {
		return segment, ""
	}
	format := map[string]string{
		".json": "json",
		".md":   "markdown",
		".pdf":  "pdf",
	}[ext]
	if format == "" {
		return segment, ""
	}
	return strings.TrimSuffix(segment, ext), format
}

func routeRequestedFormat(r *http.Request, pathFormat string) string {
	if format := r.URL.Query().Get("format"); format != "" {
		switch strings.ToLower(format) {
		case "json":
			return "json"
		case "markdown", "md":
			return "markdown"
		case "pdf":
			return "pdf"
		}
	}
	if pathFormat != "" {
		return pathFormat
	}
	switch clicky.WithHttpRequest(r).Format {
	case "json", "markdown", "pdf":
		return clicky.WithHttpRequest(r).Format
	default:
		return ""
	}
}

func parseTestFilters(r *http.Request) testRouteFilters {
	return testRouteFilters{
		Status:    splitList(r.URL.Query().Get("status")),
		Framework: splitList(r.URL.Query().Get("framework")),
	}
}

func parseLintFilters(r *http.Request) lintRouteFilters {
	grouping := r.URL.Query().Get("grouping")
	if grouping == "" {
		grouping = "linter-file"
	}
	return lintRouteFilters{
		Grouping: grouping,
		Severity: splitList(r.URL.Query().Get("severity")),
		Linter:   splitList(r.URL.Query().Get("linter")),
	}
}

func splitList(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// capitalize uppercases the first ASCII letter of s. Used for tab labels
// ("tests" → "Tests") where strings.Title would be deprecated overkill.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] -= 'a' - 'A'
	}
	return string(r)
}

func (s *Server) buildExportReport(req routeRequest) (*exportReport, error) {
	snap := s.snapshot()
	report := &exportReport{
		Tab:  req.Tab,
		Done: snap.Done,
		Path: strings.Join(req.NodePath, "/"),
	}

	switch req.Tab {
	case viewTabTests:
		report.Filters = req.TestFilters
		roots := make([]*ViewNode, 0, len(snap.Tests))
		for _, test := range snap.Tests {
			roots = append(roots, testToViewNode(test))
		}
		roots = filterTestNodes(roots, req.TestFilters)
		annotateViewPaths(roots, nil)
		report.Tests = roots
		report.roots = roots
	case viewTabLint:
		report.Filters = req.LintFilters
		roots := buildLintViewNodes(snap.Lint, req.LintFilters)
		annotateViewPaths(roots, nil)
		report.Lint = roots
		report.roots = roots
	case viewTabBench:
		report.Bench = buildBenchViewNodes(snap.Bench)
		annotateViewPaths(report.Bench, nil)
		report.roots = report.Bench
	default:
		return nil, fmt.Errorf("unsupported tab: %s", req.Tab)
	}

	if len(req.NodePath) > 0 {
		selected := findViewNode(report.roots, req.NodePath)
		if selected == nil {
			return nil, errRouteNodeNotFound
		}
		report.Selected = selected
		switch req.Tab {
		case viewTabTests:
			report.Tests = []*ViewNode{selected}
		case viewTabLint:
			report.Lint = []*ViewNode{selected}
		case viewTabBench:
			report.Bench = []*ViewNode{selected}
		}
		report.roots = []*ViewNode{selected}
	}

	return report, nil
}

var errRouteNodeNotFound = fmt.Errorf("route node not found")

func writeExportResponse(w http.ResponseWriter, r *http.Request, report *exportReport, format string) {
	manager := formatters.NewFormatManager()
	output, err := manager.FormatWithOptions(formatters.FormatOptions{Format: format}, report)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to format export: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", clicky.FormatToContentType(format))
	_, _ = w.Write([]byte(output))
}

func testToViewNode(test parsers.Test) *ViewNode {
	node := &ViewNode{
		Name:        displayTestName(test.Name, test.Framework.String()),
		Framework:   test.Framework.String(),
		Kind:        "test",
		Package:     test.Package,
		PackagePath: test.PackagePath,
		File:        test.File,
		Line:        test.Line,
		Command:     test.Command,
		Message:     test.Message,
		Stdout:      test.Stdout,
		Stderr:      test.Stderr,
		Duration:    test.Duration,
	}
	switch {
	case test.Failed:
		node.Status = "failed"
	case test.Skipped:
		node.Status = "skipped"
	case test.Pending:
		node.Status = "pending"
	case test.Passed:
		node.Status = "passed"
	}
	for _, child := range test.Children {
		node.Children = append(node.Children, testToViewNode(child))
	}
	if test.Framework == parsers.Framework("task") {
		node.Kind = "task"
	}
	return node
}

func displayTestName(name, framework string) string {
	if framework != parsers.GoTest.String() {
		return name
	}
	if strings.HasSuffix(name, "/") {
		return name
	}
	parts := strings.Split(name, "/")
	for i, part := range parts {
		if i == 0 {
			part = strings.TrimPrefix(part, "Test")
			part = insertSpaces(part)
		} else {
			part = strings.ReplaceAll(part, "_", " ")
		}
		parts[i] = part
	}
	return strings.Join(parts, " / ")
}

func insertSpaces(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			prev := rune(s[i-1])
			if unicode.IsLower(prev) || unicode.IsDigit(prev) {
				b.WriteRune(' ')
			}
		}
		b.WriteRune(r)
	}
	return b.String()
}

func filterTestNodes(nodes []*ViewNode, filters testRouteFilters) []*ViewNode {
	if len(filters.Status) == 0 && len(filters.Framework) == 0 {
		return nodes
	}
	statusSet := toSet(filters.Status)
	frameworkSet := toSet(filters.Framework)
	var out []*ViewNode
	for _, node := range nodes {
		if filtered := filterTestNode(node, statusSet, frameworkSet); filtered != nil {
			out = append(out, filtered)
		}
	}
	return out
}

func filterTestNode(node *ViewNode, statusSet, frameworkSet map[string]bool) *ViewNode {
	if len(node.Children) > 0 {
		cloned := *node
		cloned.Children = nil
		for _, child := range node.Children {
			if filtered := filterTestNode(child, statusSet, frameworkSet); filtered != nil {
				cloned.Children = append(cloned.Children, filtered)
			}
		}
		if len(cloned.Children) == 0 {
			return nil
		}
		return &cloned
	}
	if len(statusSet) > 0 && !statusSet[node.Status] {
		return nil
	}
	if len(frameworkSet) > 0 && !frameworkSet[node.Framework] {
		return nil
	}
	cloned := *node
	return &cloned
}

func buildLintViewNodes(results []*linters.LinterResult, filters lintRouteFilters) []*ViewNode {
	if filters.Grouping == "file-linter-rule" {
		return buildLintByFileLinterRule(results, filters)
	}
	return buildLintByLinterFile(results, filters)
}

type lintRuleBucket struct {
	File       string
	Violations []models.Violation
}

type lintLinterBucket struct {
	Files       map[string][]models.Violation
	NoFileRules map[string]*lintRuleBucket
}

type lintFileTreeNode struct {
	Name     string
	Path     string
	Children map[string]*lintFileTreeNode
	Files    map[string]map[string]map[string][]models.Violation
}

func buildLintByLinterFile(results []*linters.LinterResult, filters lintRouteFilters) []*ViewNode {
	byLinter := map[string]*lintLinterBucket{}
	meta := map[string]*linters.LinterResult{}
	for _, result := range results {
		if len(filters.Linter) > 0 && !contains(filters.Linter, result.Linter) {
			continue
		}
		for _, violation := range result.Violations {
			if len(filters.Severity) > 0 && !contains(filters.Severity, string(violation.Severity)) {
				continue
			}
			bucket, ok := byLinter[result.Linter]
			if !ok {
				bucket = &lintLinterBucket{
					Files:       map[string][]models.Violation{},
					NoFileRules: map[string]*lintRuleBucket{},
				}
				byLinter[result.Linter] = bucket
			}
			file := relLintPath(violation.File, result.WorkDir)
			if file == "" {
				rule := lintRuleName(violation)
				if _, ok := bucket.NoFileRules[rule]; !ok {
					bucket.NoFileRules[rule] = &lintRuleBucket{}
				}
				bucket.NoFileRules[rule].Violations = append(bucket.NoFileRules[rule].Violations, violation)
				meta[result.Linter] = result
				continue
			}
			bucket.Files[file] = append(bucket.Files[file], violation)
			meta[result.Linter] = result
		}
	}

	lintersList := sortedKeys(byLinter)
	nodes := make([]*ViewNode, 0, len(lintersList))
	for _, linterName := range lintersList {
		bucket := byLinter[linterName]
		fileNames := sortedKeys(bucket.Files)
		var children []*ViewNode
		total := 0
		for _, file := range fileNames {
			violations := bucket.Files[file]
			total += len(violations)
			children = append(children, lintFileNode(file, file, linterName, violations))
		}
		for _, rule := range sortedKeys(bucket.NoFileRules) {
			violations := bucket.NoFileRules[rule].Violations
			total += len(violations)
		}
		status := "passed"
		if total > 0 {
			status = "failed"
		}
		nodes = append(nodes, &ViewNode{
			Name:      fmt.Sprintf("%s (%d)", linterName, total),
			Kind:      "linter",
			Status:    status,
			Framework: "lint",
			Message:   lintMetaMessage(meta[linterName]),
			Children:  children,
		})
	}
	return nodes
}

func buildLintByFileLinterRule(results []*linters.LinterResult, filters lintRouteFilters) []*ViewNode {
	root := newLintFileTreeNode("", "")
	byNoFileLinter := map[string]map[string]*lintRuleBucket{}
	meta := map[string]*linters.LinterResult{}
	for _, result := range results {
		if len(filters.Linter) > 0 && !contains(filters.Linter, result.Linter) {
			continue
		}
		for _, violation := range result.Violations {
			if len(filters.Severity) > 0 && !contains(filters.Severity, string(violation.Severity)) {
				continue
			}
			file := relLintPath(violation.File, result.WorkDir)
			meta[result.Linter] = result
			if file == "" {
				if _, ok := byNoFileLinter[result.Linter]; !ok {
					byNoFileLinter[result.Linter] = map[string]*lintRuleBucket{}
				}
				rule := lintRuleName(violation)
				if _, ok := byNoFileLinter[result.Linter][rule]; !ok {
					byNoFileLinter[result.Linter][rule] = &lintRuleBucket{}
				}
				byNoFileLinter[result.Linter][rule].Violations = append(byNoFileLinter[result.Linter][rule].Violations, violation)
				continue
			}
			insertLintFileNode(root, file, result.Linter, violation)
		}
	}

	nodes := buildLintFolderNodes(root, meta)
	for _, linterName := range sortedKeys(byNoFileLinter) {
		rules := byNoFileLinter[linterName]
		ruleNames := sortedKeys(rules)
		var ruleNodes []*ViewNode
		total := 0
		for _, ruleName := range ruleNames {
			violations := rules[ruleName].Violations
			total += len(violations)
			ruleNodes = append(ruleNodes, lintRuleNode(ruleName, linterName, "", violations))
		}
		nodes = append(nodes, &ViewNode{
			Name:      fmt.Sprintf("%s (%d)", linterName, total),
			Kind:      "linter",
			Status:    statusFromChildNodes(ruleNodes),
			Framework: "lint",
			Message:   lintMetaMessage(meta[linterName]),
			Children:  ruleNodes,
		})
	}
	return nodes
}

func lintMetaMessage(result *linters.LinterResult) string {
	if result == nil {
		return ""
	}
	if result.Error != "" {
		return result.Error
	}
	return ""
}

func lintFileNode(label, file, linterName string, violations []models.Violation) *ViewNode {
	sort.Slice(violations, func(i, j int) bool {
		return violations[i].Line < violations[j].Line
	})
	children := make([]*ViewNode, 0, len(violations))
	for _, violation := range violations {
		children = append(children, lintViolationNode(linterName, file, violation))
	}
	return &ViewNode{
		Name:      fmt.Sprintf("%s (%d)", label, len(violations)),
		Kind:      "lint-file",
		Status:    statusFromViolations(violations),
		Framework: "lint",
		File:      file,
		Children:  children,
	}
}

func lintRuleNode(ruleName, linterName, file string, violations []models.Violation) *ViewNode {
	sort.Slice(violations, func(i, j int) bool {
		return violations[i].Line < violations[j].Line
	})
	children := make([]*ViewNode, 0, len(violations))
	for _, violation := range violations {
		children = append(children, lintViolationNode(linterName, file, violation))
	}
	return &ViewNode{
		Name:      fmt.Sprintf("%s (%d)", ruleName, len(violations)),
		Kind:      "lint-rule",
		Status:    statusFromViolations(violations),
		Framework: "lint",
		File:      file,
		Children:  children,
	}
}

func lintViolationNode(linterName, file string, violation models.Violation) *ViewNode {
	msg := ""
	if violation.Message != nil {
		msg = *violation.Message
	}
	name := ""
	switch {
	case file != "" && violation.Line > 0:
		name = fmt.Sprintf("%s:%d", file, violation.Line)
	case file != "":
		name = file
	case violation.Line > 0:
		name = fmt.Sprintf("line %d", violation.Line)
	default:
		name = "(violation)"
	}
	if msg != "" {
		name = name + " " + msg
	}
	return &ViewNode{
		Name:      name,
		Kind:      "violation",
		Status:    statusFromSeverity(violation.Severity),
		Framework: "lint",
		File:      file,
		Line:      violation.Line,
		Message:   msg,
		Command:   linterName,
	}
}

func newLintFileTreeNode(name, path string) *lintFileTreeNode {
	return &lintFileTreeNode{
		Name:     name,
		Path:     path,
		Children: map[string]*lintFileTreeNode{},
		Files:    map[string]map[string]map[string][]models.Violation{},
	}
}

func insertLintFileNode(root *lintFileTreeNode, file, linterName string, violation models.Violation) {
	segments := collapsedLintSegments(file)
	if len(segments) == 0 {
		return
	}
	current := root
	for _, segment := range segments[:len(segments)-1] {
		child, ok := current.Children[segment]
		if !ok {
			path := segment
			if current.Path != "" {
				path = current.Path + "/" + segment
			}
			child = newLintFileTreeNode(segment, path)
			current.Children[segment] = child
		}
		current = child
	}

	base := segments[len(segments)-1]
	if _, ok := current.Files[base]; !ok {
		current.Files[base] = map[string]map[string][]models.Violation{}
	}
	if _, ok := current.Files[base][linterName]; !ok {
		current.Files[base][linterName] = map[string][]models.Violation{}
	}
	rule := lintRuleName(violation)
	current.Files[base][linterName][rule] = append(current.Files[base][linterName][rule], violation)
}

func buildLintFolderNodes(root *lintFileTreeNode, meta map[string]*linters.LinterResult) []*ViewNode {
	nodes := make([]*ViewNode, 0, len(root.Children)+len(root.Files))
	for _, folderName := range sortedKeys(root.Children) {
		child := root.Children[folderName]
		children := buildLintFolderNodes(child, meta)
		total := lintNodeViolationCount(children)
		nodes = append(nodes, &ViewNode{
			Name:      fmt.Sprintf("%s (%d)", child.Name, total),
			Kind:      "lint-folder",
			Status:    statusFromChildNodes(children),
			Framework: "lint",
			File:      child.Path,
			Children:  children,
		})
	}
	for _, fileName := range sortedKeys(root.Files) {
		fullPath := fileName
		if root.Path != "" {
			fullPath = root.Path + "/" + fileName
		}
		lintersMap := root.Files[fileName]
		linterNames := sortedKeys(lintersMap)
		var linterNodes []*ViewNode
		total := 0
		for _, linterName := range linterNames {
			rules := lintersMap[linterName]
			ruleNames := sortedKeys(rules)
			var ruleNodes []*ViewNode
			linterTotal := 0
			for _, ruleName := range ruleNames {
				violations := rules[ruleName]
				linterTotal += len(violations)
				ruleNodes = append(ruleNodes, lintRuleNode(ruleName, linterName, fullPath, violations))
			}
			total += linterTotal
			linterNodes = append(linterNodes, &ViewNode{
				Name:      fmt.Sprintf("%s (%d)", linterName, linterTotal),
				Kind:      "linter",
				Status:    statusFromChildNodes(ruleNodes),
				Framework: "lint",
				File:      fullPath,
				Message:   lintMetaMessage(meta[linterName]),
				Children:  ruleNodes,
			})
		}
		nodes = append(nodes, &ViewNode{
			Name:      fmt.Sprintf("%s (%d)", fileName, total),
			Kind:      "lint-file",
			Status:    statusFromChildNodes(linterNodes),
			Framework: "lint",
			File:      fullPath,
			Children:  linterNodes,
		})
	}
	return nodes
}

func lintNodeViolationCount(nodes []*ViewNode) int {
	total := 0
	for _, node := range nodes {
		if len(node.Children) == 0 {
			total++
			continue
		}
		total += lintNodeViolationCount(node.Children)
	}
	return total
}

func statusFromChildNodes(children []*ViewNode) string {
	status := "passed"
	for _, child := range children {
		if child.Status == "failed" {
			return "failed"
		}
		if child.Status == "pending" {
			status = "pending"
		}
	}
	return status
}

func lintRuleName(violation models.Violation) string {
	if violation.Rule != nil && violation.Rule.Method != "" {
		return violation.Rule.Method
	}
	return "(no rule)"
}

func collapsedLintSegments(path string) []string {
	parts := strings.Split(normalizeLintPath(path), "/")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	if len(filtered) >= 3 && filtered[0] == ".shell" && filtered[1] == "checkout" {
		return append([]string{strings.Join(filtered[:3], "/")}, filtered[3:]...)
	}
	return filtered
}

func normalizeLintPath(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return strings.TrimSuffix(path, "/")
}

func buildBenchViewNodes(cmp *bench.BenchComparison) []*ViewNode {
	if cmp == nil {
		return nil
	}
	nodes := make([]*ViewNode, 0, len(cmp.Deltas))
	for _, delta := range cmp.Deltas {
		status := "passed"
		switch {
		case delta.OnlyIn != "":
			status = "skipped"
		case delta.Significant && delta.DeltaPct > cmp.Threshold:
			status = "failed"
		case delta.Significant && delta.DeltaPct < -cmp.Threshold:
			status = "passed"
		}
		message := fmt.Sprintf("%s -> %s (%+.2f%%)", formatBenchNs(delta.BaseMean), formatBenchNs(delta.HeadMean), delta.DeltaPct)
		if delta.OnlyIn != "" {
			message = "only in " + delta.OnlyIn
		}
		nodes = append(nodes, &ViewNode{
			Name:      delta.Name,
			Kind:      "bench",
			Status:    status,
			Framework: "bench",
			Package:   delta.Package,
			Message:   message,
		})
	}
	return nodes
}

func formatBenchNs(ns float64) string {
	switch {
	case ns >= 1e9:
		return fmt.Sprintf("%.2fs", ns/1e9)
	case ns >= 1e6:
		return fmt.Sprintf("%.2fms", ns/1e6)
	case ns >= 1e3:
		return fmt.Sprintf("%.2fµs", ns/1e3)
	default:
		return fmt.Sprintf("%.2fns", ns)
	}
}

func annotateViewPaths(nodes []*ViewNode, parent []string) {
	counts := map[string]int{}
	slugs := make([]string, len(nodes))
	for i, node := range nodes {
		slug := slugify(node.Name)
		slugs[i] = slug
		counts[slug]++
	}
	seen := map[string]int{}
	for i, node := range nodes {
		slug := slugs[i]
		seen[slug]++
		if counts[slug] > 1 {
			slug = fmt.Sprintf("%s~%d", slug, seen[slug])
		}
		path := append(append([]string{}, parent...), slug)
		node.Path = strings.Join(path, "/")
		annotateViewPaths(node.Children, path)
	}
}

func findViewNode(nodes []*ViewNode, path []string) *ViewNode {
	target := strings.Join(path, "/")
	for _, node := range nodes {
		if node.Path == target {
			return node
		}
		if child := findViewNode(node.Children, path); child != nil {
			return child
		}
	}
	return nil
}

func slugify(input string) string {
	input = strings.ToLower(input)
	var b strings.Builder
	lastDash := false
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "node"
	}
	return out
}

func relLintPath(file, workDir string) string {
	if file == "" {
		return ""
	}
	file = normalizeLintPath(file)
	if workDir == "" {
		return file
	}
	prefix := normalizeLintPath(workDir)
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	if strings.HasPrefix(file, prefix) {
		return strings.TrimPrefix(file, prefix)
	}
	return file
}

func statusFromViolations(violations []models.Violation) string {
	status := "passed"
	for _, violation := range violations {
		switch violation.Severity {
		case models.SeverityError:
			return "failed"
		case models.SeverityWarning:
			status = "failed"
		case models.SeverityInfo:
			if status == "passed" {
				status = "passed"
			}
		}
	}
	return status
}

func statusFromSeverity(severity models.ViolationSeverity) string {
	switch severity {
	case models.SeverityError, models.SeverityWarning:
		return "failed"
	default:
		return "passed"
	}
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func toSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}
