package ui

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/formatters"
	_ "github.com/flanksource/clicky/formatters/html"
	"github.com/flanksource/gavel/github"
)

const viewTabPRs = "prs"

type routeRequest struct {
	Tab       string
	NodePath  []string
	Format    string
	IsExport  bool
	PRFilters prRouteFilters
}

type prRouteFilters struct {
	State   []string `json:"state,omitempty"`
	Checks  []string `json:"checks,omitempty"`
	Repos   []string `json:"repos,omitempty"`
	Authors []string `json:"authors,omitempty"`
}

type exportReport struct {
	Tab      string        `json:"tab"`
	Path     string        `json:"path,omitempty"`
	Filters  any           `json:"filters,omitempty"`
	Selected *PRViewNode   `json:"selected,omitempty"`
	PRs      []*PRViewNode `json:"prs,omitempty"`
	Done     bool          `json:"done"`

	roots []*PRViewNode `json:"-"`
}

type PRViewNode struct {
	Repo           string               `json:"repo"`
	Number         int                  `json:"number"`
	Title          string               `json:"title"`
	Author         string               `json:"author,omitempty"`
	State          string               `json:"state,omitempty"`
	IsDraft        bool                 `json:"isDraft,omitempty"`
	ReviewDecision string               `json:"reviewDecision,omitempty"`
	Mergeable      string               `json:"mergeable,omitempty"`
	URL            string               `json:"url,omitempty"`
	UpdatedAt      time.Time            `json:"updatedAt,omitempty"`
	Ahead          int                  `json:"ahead,omitempty"`
	Behind         int                  `json:"behind,omitempty"`
	CheckStatus    *github.CheckSummary `json:"checkStatus,omitempty"`

	RoutePath string `json:"route_path,omitempty"`

	// Populated only when Selected (single-PR export).
	PR       *github.PRInfo                `json:"pr,omitempty"`
	Runs     map[int64]*github.WorkflowRun `json:"runs,omitempty"`
	Comments []github.PRComment            `json:"comments,omitempty"`
	Detail   string                        `json:"detailError,omitempty"`
}

func (n *PRViewNode) Pretty() api.Text {
	text := clicky.Text("")
	icon, style := prStateIconStyle(n.State, n.IsDraft)
	text = text.Append(icon, style)
	text = text.Space().Append(fmt.Sprintf("%s#%d", n.Repo, n.Number), "text-muted")
	text = text.Space().Append(n.Title, "bold")
	if n.Author != "" {
		text = text.Space().Append("@"+n.Author, "text-muted")
	}
	if cs := n.CheckStatus; cs != nil {
		text = text.Space().Add(cs.PrettySummary())
	}
	if n.ReviewDecision != "" {
		text = text.Space().Append(n.ReviewDecision, reviewDecisionStyle(n.ReviewDecision))
	}

	if n.Detail != "" {
		text = text.NewLine().Append("detail error: "+n.Detail, "text-red-500")
	}
	if len(n.Comments) > 0 {
		var body strings.Builder
		for _, c := range n.Comments {
			fmt.Fprintf(&body, "@%s", c.Author)
			if c.Path != "" {
				fmt.Fprintf(&body, " %s", c.Path)
				if c.Line > 0 {
					fmt.Fprintf(&body, ":%d", c.Line)
				}
			}
			body.WriteString("\n")
			body.WriteString(strings.TrimSpace(c.Body))
			body.WriteString("\n\n")
		}
		text = text.NewLine().Add(api.Collapsed{
			Label:   fmt.Sprintf("comments (%d)", len(n.Comments)),
			Content: clicky.Text(body.String(), "whitespace-pre-wrap"),
		})
	}
	if cs := n.CheckStatus; cs != nil && len(cs.Failures) > 0 {
		var body strings.Builder
		for _, f := range cs.Failures {
			fmt.Fprintf(&body, "✗ %s\n", f.Name)
			if f.DetailsURL != "" {
				fmt.Fprintf(&body, "  %s\n", f.DetailsURL)
			}
			for _, step := range f.FailedSteps {
				fmt.Fprintf(&body, "  - %s\n", step)
			}
			if f.LogTail != "" {
				body.WriteString("  log tail:\n")
				for _, line := range strings.Split(strings.TrimRight(f.LogTail, "\n"), "\n") {
					fmt.Fprintf(&body, "    %s\n", line)
				}
			}
		}
		text = text.NewLine().Add(api.Collapsed{
			Label:   fmt.Sprintf("failed checks (%d)", len(cs.Failures)),
			Content: clicky.Text(body.String(), "font-mono text-xs whitespace-pre-wrap"),
		})
	}
	return text
}

func (n *PRViewNode) GetChildren() []api.TreeNode { return nil }

func prStateIconStyle(state string, draft bool) (string, string) {
	if draft {
		return "○", "text-gray-400"
	}
	switch strings.ToUpper(state) {
	case "OPEN":
		return "●", "text-green-600"
	case "MERGED":
		return "●", "text-purple-600"
	case "CLOSED":
		return "●", "text-red-600"
	}
	return "?", "text-gray-400"
}

func reviewDecisionStyle(decision string) string {
	switch strings.ToUpper(decision) {
	case "APPROVED":
		return "text-green-600"
	case "CHANGES_REQUESTED":
		return "text-red-600"
	case "REVIEW_REQUIRED":
		return "text-yellow-600"
	}
	return "text-muted"
}

func (r exportReport) Pretty() api.Text {
	text := clicky.Text("PR export", "bold")
	if r.Path != "" {
		text = text.Space().Append("("+r.Path+")", "text-muted")
	}
	if f, ok := r.Filters.(prRouteFilters); ok {
		if filterText := renderPRFilters(f); filterText != "" {
			text = text.Space().Append(filterText, "text-muted")
		}
	}
	text = text.Space().Append(fmt.Sprintf("%d PR(s)", len(r.roots)), "text-muted")
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

func renderPRFilters(f prRouteFilters) string {
	parts := make([]string, 0, 4)
	if len(f.State) > 0 {
		parts = append(parts, "state="+strings.Join(f.State, ","))
	}
	if len(f.Checks) > 0 {
		parts = append(parts, "checks="+strings.Join(f.Checks, ","))
	}
	if len(f.Repos) > 0 {
		parts = append(parts, "repos="+strings.Join(f.Repos, ","))
	}
	if len(f.Authors) > 0 {
		parts = append(parts, "authors="+strings.Join(f.Authors, ","))
	}
	return strings.Join(parts, " ")
}

func parseRouteRequest(r *http.Request) (routeRequest, bool) {
	req := routeRequest{Tab: viewTabPRs}

	path := strings.Trim(r.URL.Path, "/")
	if path == "" {
		req.Format = routeRequestedFormat(r, "")
		req.IsExport = req.Format != ""
		req.PRFilters = parsePRFilters(r)
		return req, true
	}

	segments := strings.Split(path, "/")
	tabSeg := segments[0]
	pathFormat := ""
	if base, format := stripKnownFormat(tabSeg); format != "" {
		tabSeg = base
		pathFormat = format
	}

	if tabSeg != viewTabPRs {
		return routeRequest{}, false
	}
	req.Tab = tabSeg

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
	req.PRFilters = parsePRFilters(r)
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
		}
	}
	if pathFormat != "" {
		return pathFormat
	}
	// Only honor Accept header when it's explicitly set to a supported
	// machine format. We deliberately do NOT fall through to clicky's
	// default "json" when no Accept header is present — browser GETs
	// to /prs must render the SPA, not trigger a JSON export.
	accept := strings.ToLower(r.Header.Get("Accept"))
	if strings.Contains(accept, "application/json") {
		return "json"
	}
	if strings.Contains(accept, "text/markdown") {
		return "markdown"
	}
	return ""
}

func parsePRFilters(r *http.Request) prRouteFilters {
	return prRouteFilters{
		State:   splitList(r.URL.Query().Get("state")),
		Checks:  splitList(r.URL.Query().Get("checks")),
		Repos:   splitList(r.URL.Query().Get("repos")),
		Authors: splitList(r.URL.Query().Get("authors")),
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

var errRouteNodeNotFound = fmt.Errorf("route node not found")

func (s *Server) buildExportReport(req routeRequest) (*exportReport, error) {
	s.mu.RLock()
	prs := s.prs
	done := !s.paused
	s.mu.RUnlock()

	report := &exportReport{
		Tab:     req.Tab,
		Done:    done,
		Path:    strings.Join(req.NodePath, "/"),
		Filters: req.PRFilters,
	}

	roots := make([]*PRViewNode, 0, len(prs))
	for i := range prs {
		roots = append(roots, prListItemToViewNode(prs[i]))
	}
	roots = filterPRNodes(roots, req.PRFilters)
	annotatePRRoutePaths(roots)
	report.PRs = roots
	report.roots = roots

	if len(req.NodePath) > 0 {
		selected := findPRNode(roots, req.NodePath)
		if selected == nil {
			return nil, errRouteNodeNotFound
		}
		if err := s.populatePRDetail(selected); err != nil {
			selected.Detail = err.Error()
		}
		report.Selected = selected
		report.PRs = []*PRViewNode{selected}
		report.roots = []*PRViewNode{selected}
	}

	return report, nil
}

func prListItemToViewNode(pr github.PRListItem) *PRViewNode {
	return &PRViewNode{
		Repo:           pr.Repo,
		Number:         pr.Number,
		Title:          pr.Title,
		Author:         pr.Author,
		State:          pr.State,
		IsDraft:        pr.IsDraft,
		ReviewDecision: pr.ReviewDecision,
		Mergeable:      pr.Mergeable,
		URL:            pr.URL,
		UpdatedAt:      pr.UpdatedAt,
		Ahead:          pr.Ahead,
		Behind:         pr.Behind,
		CheckStatus:    pr.CheckStatus,
	}
}

func annotatePRRoutePaths(nodes []*PRViewNode) {
	for _, node := range nodes {
		node.RoutePath = fmt.Sprintf("%s/%d", node.Repo, node.Number)
	}
}

// findPRNode resolves a hierarchical NodePath ([owner, repo, number] or
// [repo-slug, number]) to a PRViewNode by matching its RoutePath.
func findPRNode(nodes []*PRViewNode, path []string) *PRViewNode {
	target := strings.Join(path, "/")
	for _, node := range nodes {
		if node.RoutePath == target {
			return node
		}
	}
	return nil
}

func filterPRNodes(nodes []*PRViewNode, f prRouteFilters) []*PRViewNode {
	if len(f.State) == 0 && len(f.Checks) == 0 && len(f.Repos) == 0 && len(f.Authors) == 0 {
		return nodes
	}
	stateSet := toSet(f.State)
	checkSet := toSet(f.Checks)
	repoSet := toSet(f.Repos)
	authorSet := toSet(f.Authors)

	out := nodes[:0:0]
	for _, node := range nodes {
		if len(stateSet) > 0 && !matchesState(node, stateSet) {
			continue
		}
		if len(checkSet) > 0 && !matchesChecks(node, checkSet) {
			continue
		}
		if len(repoSet) > 0 && !repoSet[node.Repo] {
			continue
		}
		if len(authorSet) > 0 && !authorSet[node.Author] {
			continue
		}
		out = append(out, node)
	}
	return out
}

func matchesState(node *PRViewNode, set map[string]bool) bool {
	state := strings.ToUpper(node.State)
	if set["draft"] && node.IsDraft {
		return true
	}
	if set["open"] && state == "OPEN" && !node.IsDraft {
		return true
	}
	if set["merged"] && state == "MERGED" {
		return true
	}
	if set["closed"] && state == "CLOSED" {
		return true
	}
	return false
}

func matchesChecks(node *PRViewNode, set map[string]bool) bool {
	cs := node.CheckStatus
	if cs == nil {
		return false
	}
	if set["failing"] && cs.Failed > 0 {
		return true
	}
	if set["running"] && cs.Running > 0 {
		return true
	}
	if set["passing"] && cs.Failed == 0 && cs.Running == 0 {
		return true
	}
	return false
}

func toSet(values []string) map[string]bool {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]bool, len(values))
	for _, v := range values {
		set[v] = true
	}
	return set
}

func writeExportResponse(w http.ResponseWriter, _ *http.Request, report *exportReport, format string) {
	manager := formatters.NewFormatManager()
	output, err := manager.FormatWithOptions(formatters.FormatOptions{Format: format}, report)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to format export: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", clicky.FormatToContentType(format))
	_, _ = w.Write([]byte(output))
}
