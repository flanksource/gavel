package github

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
)

const prSearchQuery = `query($query: String!, $first: Int!) {
  search(query: $query, type: ISSUE, first: $first) {
    issueCount
    nodes {
      ... on PullRequest {
        number
        title
        author { login }
        headRefName
        baseRefName
        state
        isDraft
        reviewDecision
        url
        updatedAt
        repository {
          nameWithOwner
        }
      }
    }
  }
}`

const prSearchQueryWithStatus = `query($query: String!, $first: Int!) {
  search(query: $query, type: ISSUE, first: $first) {
    issueCount
    nodes {
      ... on PullRequest {
        number
        title
        author { login }
        headRefName
        baseRefName
        state
        isDraft
        reviewDecision
        url
        updatedAt
        repository {
          nameWithOwner
        }
        commits(last: 1) {
          nodes {
            commit {
              statusCheckRollup {
                contexts(first: 100) {
                  nodes {
                    __typename
                    ... on CheckRun {
                      name
                      status
                      conclusion
                      detailsUrl
                    }
                    ... on StatusContext {
                      context
                      state
                      targetUrl
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`

type FailedCheck struct {
	Name       string   `json:"name"`
	DetailsURL string   `json:"detailsUrl,omitempty"`
	FailedSteps []string `json:"failedSteps,omitempty"`
	LogTail    string   `json:"logTail,omitempty"`
}

type CheckSummary struct {
	Passed  int           `json:"passed"`
	Failed  int           `json:"failed"`
	Running int           `json:"running"`
	Pending int           `json:"pending"`
	Failures []FailedCheck `json:"failures,omitempty"`
}

func (cs CheckSummary) PrettySummary() api.Text {
	text := clicky.Text("", "")
	if cs.Passed > 0 {
		text = text.Append(fmt.Sprintf("✓%d", cs.Passed), "text-green-600")
	}
	if cs.Failed > 0 {
		text = text.Append(fmt.Sprintf(" ✗%d", cs.Failed), "text-red-600")
	}
	if cs.Running > 0 {
		text = text.Append(fmt.Sprintf(" ●%d", cs.Running), "text-yellow-600")
	}
	if cs.Pending > 0 {
		text = text.Append(fmt.Sprintf(" ○%d", cs.Pending), "text-gray-500")
	}
	return text
}

func (f FailedCheck) Pretty(indent string) api.Text {
	text := clicky.Text(indent+"✗ ", "text-red-600").Append(f.Name, "text-red-600")
	for _, step := range f.FailedSteps {
		text = text.NewLine().Append(indent+"  ✗ "+step, "text-red-600")
	}
	if f.LogTail != "" {
		for _, line := range strings.Split(strings.TrimSpace(f.LogTail), "\n") {
			if isActionsNoise(line) {
				continue
			}
			text = text.NewLine().Append(indent+"    "+line, "text-gray-500")
		}
	}
	return text
}

func isActionsNoise(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "[command]/usr/bin/git") ||
		strings.HasPrefix(line, "Post job cleanup") ||
		strings.HasPrefix(line, "Node.js 20 actions are deprecated") ||
		strings.HasPrefix(line, "Node.js 16") ||
		strings.HasPrefix(line, "Temporarily overriding HOME") ||
		strings.HasPrefix(line, "Adding repository directory") ||
		strings.HasPrefix(line, "Cleaning up orphan processes") ||
		strings.HasPrefix(line, "http.https://github.com/") ||
		strings.HasPrefix(line, "git version")
}

type PRSearchOptions struct {
	Author  string
	Since   time.Time
	State   string
	All     bool
	Org     string
	Repos   []string // explicit list of owner/repo to search
	Limit   int
	Status  bool // include GitHub Actions check status counts
	Verbose bool // with --status, fetch and show failed step logs
	ShowURL bool // show PR URL instead of #number
}

type searchResponse struct {
	Data   searchData     `json:"data"`
	Errors []graphQLError `json:"errors"`
}

type searchData struct {
	Search searchResult `json:"search"`
}

type searchResult struct {
	IssueCount int            `json:"issueCount"`
	Nodes      []searchPRNode `json:"nodes"`
}

type searchPRNode struct {
	Number         int           `json:"number"`
	Title          string        `json:"title"`
	Author         graphQLAuthor `json:"author"`
	HeadRefName    string        `json:"headRefName"`
	BaseRefName    string        `json:"baseRefName"`
	State          string        `json:"state"`
	IsDraft        bool          `json:"isDraft"`
	ReviewDecision string        `json:"reviewDecision"`
	URL            string        `json:"url"`
	UpdatedAt      time.Time     `json:"updatedAt"`
	Repository     struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Commits graphQLCommits `json:"commits"`
}

type PRListItem struct {
	Number         int           `json:"number"`
	Title          string        `json:"title"`
	Author         string        `json:"author"`
	Repo           string        `json:"repo"`
	Source         string        `json:"source"`
	Target         string        `json:"target"`
	State          string        `json:"state"`
	IsDraft        bool          `json:"isDraft"`
	ReviewDecision string        `json:"reviewDecision,omitempty"`
	URL            string        `json:"url"`
	UpdatedAt      time.Time     `json:"updatedAt"`
	IsCurrent      bool          `json:"isCurrent,omitempty"`
	Ahead          int           `json:"ahead,omitempty"`
	Behind         int           `json:"behind,omitempty"`
	CheckStatus    *CheckSummary `json:"checkStatus,omitempty"`
	ShowURL        bool          `json:"-"`
}

type PRSearchResults []PRListItem

func buildSearchQueryBase(searchOpts PRSearchOptions) []string {
	parts := []string{"is:pr"}

	if searchOpts.Author != "" {
		parts = append(parts, "author:"+searchOpts.Author)
	}

	switch searchOpts.State {
	case "closed":
		parts = append(parts, "is:closed")
	case "merged":
		parts = append(parts, "is:merged")
	case "all":
		// no state filter
	default:
		parts = append(parts, "is:open")
	}

	if !searchOpts.Since.IsZero() {
		parts = append(parts, fmt.Sprintf("updated:>%s", searchOpts.Since.Format("2006-01-02")))
	}

	return parts
}

func buildSearchQuery(opts Options, searchOpts PRSearchOptions) (string, error) {
	parts := buildSearchQueryBase(searchOpts)

	if searchOpts.All {
		org := searchOpts.Org
		if org == "" {
			repo, err := opts.resolveRepo()
			if err != nil {
				return "", fmt.Errorf("cannot determine org (use --org): %w", err)
			}
			org = strings.SplitN(repo, "/", 2)[0]
		}
		parts = append(parts, "org:"+org)
	} else {
		repo, err := opts.resolveRepo()
		if err != nil {
			return "", err
		}
		parts = append(parts, "repo:"+repo)
	}

	return strings.Join(parts, " "), nil
}

func buildSearchQueryForRepo(repo string, searchOpts PRSearchOptions) string {
	parts := buildSearchQueryBase(searchOpts)
	parts = append(parts, "repo:"+repo)
	return strings.Join(parts, " ")
}

func SearchPRs(opts Options, searchOpts PRSearchOptions) (PRSearchResults, error) {
	token, err := opts.token()
	if err != nil {
		return nil, err
	}

	if len(searchOpts.Repos) > 0 {
		return searchMultipleRepos(token, searchOpts)
	}

	queryString, err := buildSearchQuery(opts, searchOpts)
	if err != nil {
		return nil, err
	}

	items, err := executeSearch(token, queryString, searchOpts)
	if err != nil {
		return nil, err
	}

	if searchOpts.Verbose && searchOpts.Status {
		enrichFailedChecks(opts, items)
	}

	if !searchOpts.All {
		markCurrentBranch(opts, items)
	}

	return items, nil
}

func searchMultipleRepos(token string, searchOpts PRSearchOptions) (PRSearchResults, error) {
	var all PRSearchResults
	for _, repo := range searchOpts.Repos {
		queryString := buildSearchQueryForRepo(repo, searchOpts)
		items, err := executeSearch(token, queryString, searchOpts)
		if err != nil {
			logger.Warnf("failed to search %s: %v", repo, err)
			continue
		}
		all = append(all, items...)
	}
	return all, nil
}

func executeSearch(token, queryString string, searchOpts PRSearchOptions) (PRSearchResults, error) {
	limit := searchOpts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := prSearchQuery
	if searchOpts.Status {
		query = prSearchQueryWithStatus
	}

	body := map[string]any{
		"query": query,
		"variables": map[string]any{
			"query": queryString,
			"first": limit,
		},
	}

	ctx := context.Background()
	client := newClient(token)

	logger.Tracef("searching PRs via GraphQL: %s", queryString)
	resp, err := client.R(ctx).
		Header("Content-Type", "application/json").
		Post("https://api.github.com/graphql", body)
	if err != nil {
		return nil, fmt.Errorf("GraphQL request: %w", err)
	}
	if !resp.IsOK() {
		respBody, _ := resp.AsString()
		return nil, fmt.Errorf("GraphQL request: status %d: %s", resp.StatusCode, respBody)
	}

	var result searchResponse
	if err := resp.Into(&result); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}
	if len(result.Errors) > 0 {
		msgs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			msgs[i] = e.Message
		}
		return nil, fmt.Errorf("GraphQL errors: %s", strings.Join(msgs, "; "))
	}

	var items PRSearchResults
	for _, node := range result.Data.Search.Nodes {
		item := PRListItem{
			Number:         node.Number,
			Title:          node.Title,
			Author:         node.Author.Login,
			Repo:           node.Repository.NameWithOwner,
			Source:         node.HeadRefName,
			Target:         node.BaseRefName,
			State:          node.State,
			IsDraft:        node.IsDraft,
			ReviewDecision: node.ReviewDecision,
			URL:            node.URL,
			UpdatedAt:      node.UpdatedAt,
		}
		if searchOpts.Status {
			item.CheckStatus = computeCheckSummary(node)
		}
		item.ShowURL = searchOpts.ShowURL
		items = append(items, item)
	}

	logger.Debugf("found %d PRs (total: %d)", len(items), result.Data.Search.IssueCount)
	return items, nil
}

func computeCheckSummary(node searchPRNode) *CheckSummary {
	if len(node.Commits.Nodes) == 0 {
		return nil
	}
	rollup := node.Commits.Nodes[0].Commit.StatusCheckRollup
	if rollup == nil {
		return nil
	}

	var cs CheckSummary
	for _, check := range rollup.Contexts.Nodes {
		sc := check.toStatusCheck()
		switch {
		case sc.Status == "COMPLETED" && (sc.Conclusion == "SUCCESS" || sc.Conclusion == "NEUTRAL" || sc.Conclusion == "SKIPPED"):
			cs.Passed++
		case sc.Status == "COMPLETED" && (sc.Conclusion == "FAILURE" || sc.Conclusion == "TIMED_OUT" || sc.Conclusion == "STARTUP_FAILURE"):
			cs.Failed++
			cs.Failures = append(cs.Failures, FailedCheck{
				Name:       sc.Name,
				DetailsURL: sc.DetailsURL,
			})
		case sc.Status == "IN_PROGRESS":
			cs.Running++
		default:
			cs.Pending++
		}
	}
	return &cs
}

func enrichFailedChecks(opts Options, items PRSearchResults) {
	for i := range items {
		if items[i].CheckStatus == nil {
			continue
		}
		for j := range items[i].CheckStatus.Failures {
			f := &items[i].CheckStatus.Failures[j]
			if f.DetailsURL == "" {
				continue
			}
			runID, err := ExtractRunID(f.DetailsURL)
			if err != nil {
				continue
			}
			run, err := FetchRunJobs(opts, runID)
			if err != nil {
				logger.Warnf("failed to fetch run %d: %v", runID, err)
				continue
			}
			FetchAndAttachLogs(opts, run, 20)
			for _, job := range run.Jobs {
				if !strings.EqualFold(job.Conclusion, "failure") {
					continue
				}
				for _, step := range job.Steps {
					if strings.EqualFold(step.Conclusion, "failure") {
						f.FailedSteps = append(f.FailedSteps, step.Name)
						if step.Logs != "" {
							f.LogTail = step.Logs
						}
					}
				}
				if f.LogTail == "" && job.Logs != "" {
					f.LogTail = job.Logs
				}
			}
		}
	}
}

func markCurrentBranch(opts Options, items PRSearchResults) {
	branch, err := opts.currentBranch()
	if err != nil || branch == "" {
		return
	}

	for i := range items {
		if items[i].Source == branch {
			items[i].IsCurrent = true
			items[i].Ahead, items[i].Behind = getAheadBehind(opts, items[i].Target)
			break
		}
	}
}

func getAheadBehind(opts Options, target string) (ahead, behind int) {
	cmd := exec.Command("git", "rev-list", "--left-right", "--count",
		fmt.Sprintf("origin/%s...HEAD", target))
	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}
	out, err := cmd.Output()
	if err != nil {
		return 0, 0
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		return 0, 0
	}
	behind, _ = strconv.Atoi(parts[0])
	ahead, _ = strconv.Atoi(parts[1])
	return ahead, behind
}

func StateIcon(state string, isDraft bool) api.Text {
	if isDraft {
		return clicky.Text("○", "text-gray-500")
	}
	switch state {
	case "OPEN":
		return clicky.Text("●", "text-green-600")
	case "MERGED":
		return clicky.Text("●", "text-purple-600")
	case "CLOSED":
		return clicky.Text("●", "text-red-600")
	default:
		return clicky.Text("?", "text-gray-500")
	}
}

func (item PRListItem) Pretty() api.Text {
	return item.prettyWithIndent("  ", true)
}

func (item PRListItem) prettyWithIndent(indent string, showRepo bool) api.Text {
	text := clicky.Text(indent, "")

	if item.IsCurrent {
		text = text.Append("→ ", "text-yellow-600")
	}

	text = text.Add(StateIcon(item.State, item.IsDraft))

	if showRepo && item.Repo != "" {
		parts := strings.SplitN(item.Repo, "/", 2)
		repoName := item.Repo
		if len(parts) == 2 {
			repoName = parts[1]
		}
		text = text.Append(" "+repoName, "text-cyan-600")
	}

	if item.ShowURL {
		text = text.Append(" "+item.URL, "text-gray-500")
	} else {
		text = text.Append(fmt.Sprintf(" #%d", item.Number), "text-gray-500")
	}
	text = text.Append(" "+item.Title, "")

	if item.IsDraft {
		text = text.Append(" DRAFT", "text-gray-500")
	}
	if item.ReviewDecision != "" {
		text = text.Append(" "+item.ReviewDecision, ReviewStyle(item.ReviewDecision))
	}

	text = text.Append(" "+item.Source, "text-cyan-600").
		Append(" → ", "text-gray-500").
		Append(item.Target, "text-cyan-600")

	if item.IsCurrent && (item.Ahead > 0 || item.Behind > 0) {
		text = text.Append(fmt.Sprintf(" ↑%d↓%d", item.Ahead, item.Behind), "text-yellow-600")
	}

	if item.CheckStatus != nil {
		text = text.Append(" ", "").Add(item.CheckStatus.PrettySummary())
		for _, f := range item.CheckStatus.Failures {
			text = text.NewLine().Add(f.Pretty(indent + "    "))
		}
	}

	return text
}

func (r PRSearchResults) Pretty() api.Text {
	if len(r) == 0 {
		return clicky.Text("No pull requests found", "text-gray-500")
	}

	if !r.hasMultipleRepos() {
		text := clicky.Text(fmt.Sprintf("Pull Requests (%d)", len(r)), "font-bold")
		for _, item := range r {
			text = text.NewLine().Add(item.Pretty())
		}
		return text
	}

	text := clicky.Text(fmt.Sprintf("Pull Requests (%d)", len(r)), "font-bold")
	groups := r.groupByRepo()
	for _, g := range groups {
		repoName := g.repo
		if parts := strings.SplitN(g.repo, "/", 2); len(parts) == 2 {
			repoName = parts[1]
		}
		text = text.NewLine().Append(fmt.Sprintf("\n  %s (%d)", repoName, len(g.items)), "font-bold text-cyan-600")
		for _, item := range g.items {
			text = text.NewLine().Add(item.prettyWithIndent("      ", false))
		}
	}
	return text
}

type repoGroup struct {
	repo  string
	items []PRListItem
}

func (r PRSearchResults) hasMultipleRepos() bool {
	if len(r) <= 1 {
		return false
	}
	first := r[0].Repo
	for _, item := range r[1:] {
		if item.Repo != first {
			return true
		}
	}
	return false
}

func (r PRSearchResults) groupByRepo() []repoGroup {
	order := make([]string, 0)
	groups := make(map[string]*repoGroup)
	for _, item := range r {
		if _, ok := groups[item.Repo]; !ok {
			order = append(order, item.Repo)
			groups[item.Repo] = &repoGroup{repo: item.Repo}
		}
		groups[item.Repo].items = append(groups[item.Repo].items, item)
	}
	result := make([]repoGroup, len(order))
	for i, repo := range order {
		result[i] = *groups[repo]
	}
	return result
}
