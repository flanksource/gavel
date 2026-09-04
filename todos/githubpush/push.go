package githubpush

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// AliasKind marks the aliases this package writes, so a pushed TODO can be
// told apart from an imported one.
const AliasKind = "github"

// ErrAlreadyLinked means the TODO already carries a GitHub reference. Pushing
// again would open a duplicate issue, so callers must opt in with Update
// (rewrite that issue) or Force (open a second one).
var ErrAlreadyLinked = errors.New("todo is already linked to a GitHub issue")

type Options struct {
	// GitHub carries the token and the target repo. An empty Repo resolves from
	// the workspace's origin remote.
	GitHub github.Options
	// BaseURL is the already-resolved absolute origin used to rewrite
	// server-relative attachment links. See ResolveBaseURL.
	BaseURL string
	// Force opens a second issue for a TODO that is already linked.
	Force bool
	// Update rewrites the linked issue's title and body instead of refusing.
	Update bool
	// Issue targets one specific issue — `123`, `owner/repo#123`, or its URL —
	// rewriting it and linking it to the TODO. It implies Update.
	Issue string
	// Labels copies the TODO's labels onto the issue.
	Labels bool
	// Plan includes the TODO's current plan in the issue body.
	Plan bool
}

type Result struct {
	TODO   *types.TODO `json:"todo,omitempty"`
	Repo   string      `json:"repo"`
	Number int         `json:"number"`
	URL    string      `json:"url"`
	Alias  string      `json:"alias"`
	// Updated reports that an existing issue was rewritten rather than opened.
	Updated bool `json:"updated,omitempty"`
}

// deps isolates the GitHub write so tests drive the orchestration without a
// network, mirroring commit.pushDeps.
type deps struct {
	saveIssue func(github.Options, github.IssueInput) (*github.IssueResult, error)
}

func defaultDeps() deps {
	return deps{saveIssue: github.SaveIssue}
}

// Push opens a GitHub issue for one TODO — or rewrites the issue it is already
// linked to — and records the reference on it.
func Push(ctx context.Context, provider todos.Provider, ref string, opts Options) (*Result, error) {
	return pushWithDeps(ctx, provider, ref, opts, defaultDeps())
}

func pushWithDeps(ctx context.Context, provider todos.Provider, ref string, opts Options, d deps) (*Result, error) {
	if provider == nil {
		return nil, fmt.Errorf("todo provider is required")
	}
	todo, err := provider.Get(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("load todo %q: %w", ref, err)
	}
	linker, ok := provider.(todos.AliasProvider)
	if !ok {
		return nil, fmt.Errorf("todo provider does not support external issue links; native PostgreSQL storage is required")
	}
	aliases, err := linker.Aliases(ctx, todo)
	if err != nil {
		return nil, err
	}
	tgt, err := resolveTarget(aliases, opts)
	if err != nil {
		return nil, err
	}

	plan, err := planMarkdown(ctx, provider, todo, opts)
	if err != nil {
		return nil, err
	}
	body, err := renderBody(composeIssueBody(todo, plan), opts.BaseURL)
	if err != nil {
		return nil, err
	}
	in := github.IssueInput{Number: tgt.Number, Title: todo.Title, Body: body}
	if opts.Labels {
		in.Labels = todo.Labels
	}

	ghOpts := opts.GitHub
	if tgt.Repo != "" {
		ghOpts.Repo = tgt.Repo
	}
	issue, err := d.saveIssue(ghOpts, in)
	if err != nil {
		return nil, err
	}

	alias := fmt.Sprintf("%s#%d", issue.Repo, issue.Number)
	if !hasAlias(aliases, alias) {
		if err := linker.AddAlias(ctx, todo, todos.TodoAlias{Alias: alias, Kind: AliasKind}); err != nil {
			return nil, fmt.Errorf("wrote %s but failed to record the link locally%s: %w",
				issue.URL, duplicateWarning(issue.Updated), err)
		}
	}
	if err := provider.Comment(ctx, todo, issueHistoryNote(issue)); err != nil {
		return nil, fmt.Errorf("wrote and linked %s but failed to record it in the todo's history: %w", issue.URL, err)
	}

	return &Result{
		TODO: todo, Repo: issue.Repo, Number: issue.Number,
		URL: issue.URL, Alias: alias, Updated: issue.Updated,
	}, nil
}

// resolveTarget decides whether this push opens a new issue or rewrites one,
// and refuses the combinations that would silently do the wrong thing.
func resolveTarget(aliases []todos.TodoAlias, opts Options) (Ref, error) {
	if opts.Force && (opts.Update || strings.TrimSpace(opts.Issue) != "") {
		return Ref{}, fmt.Errorf("--force opens a new issue and --update/--issue rewrite an existing one; pick one")
	}
	if ref := strings.TrimSpace(opts.Issue); ref != "" {
		return ParseIssueRef(ref)
	}

	linked := linkedIssues(aliases)
	if opts.Force || len(linked) == 0 {
		if opts.Update {
			return Ref{}, fmt.Errorf("todo is not linked to a GitHub issue yet, so there is nothing to update; " +
				"push it without --update, or name the issue with --issue owner/repo#123")
		}
		return Ref{}, nil
	}
	if !opts.Update {
		return Ref{}, fmt.Errorf("%w: %s (re-run with --update to rewrite it, or --force to open a second issue)",
			ErrAlreadyLinked, strings.Join(aliasNames(linked), ", "))
	}
	if len(linked) > 1 {
		return Ref{}, fmt.Errorf("todo is linked to %d GitHub issues (%s); name the one to rewrite with --issue",
			len(linked), strings.Join(aliasNames(linked), ", "))
	}
	return ParseIssueRef(linked[0].Alias)
}

func linkedIssues(aliases []todos.TodoAlias) []todos.TodoAlias {
	var linked []todos.TodoAlias
	for _, alias := range aliases {
		if strings.EqualFold(alias.Kind, AliasKind) {
			linked = append(linked, alias)
		}
	}
	return linked
}

func aliasNames(aliases []todos.TodoAlias) []string {
	names := make([]string, 0, len(aliases))
	for _, alias := range aliases {
		names = append(names, alias.Alias)
	}
	return names
}

func hasAlias(aliases []todos.TodoAlias, want string) bool {
	for _, alias := range aliases {
		if strings.EqualFold(alias.Alias, want) && strings.EqualFold(alias.Kind, AliasKind) {
			return true
		}
	}
	return false
}

// planMarkdown reads the TODO's current plan — the latest selected revision,
// approved or not — so the issue carries the plan reviewers are looking at
// rather than only the plan an implementation run is allowed to follow.
func planMarkdown(ctx context.Context, provider todos.Provider, todo *types.TODO, opts Options) (string, error) {
	if !opts.Plan {
		return "", nil
	}
	content, ok := provider.(todos.PlanContentProvider)
	if !ok {
		return "", fmt.Errorf("todo provider does not expose durable plan content; " +
			"native PostgreSQL storage is required (pass --plan=false to push the body alone)")
	}
	return content.PlanMarkdown(ctx, todo, types.ModePlan)
}

// composeIssueBody lays the issue out the way a reader works through it: what
// the work is, how it will be done, and how it is proven done.
func composeIssueBody(todo *types.TODO, plan string) string {
	body := strings.TrimSpace(todo.MarkdownBody)
	if plan = strings.TrimSpace(plan); plan != "" {
		if body != "" {
			body += "\n\n"
		}
		body += "# Plan\n\n" + plan
	}
	return todos.ComposeIssueMarkdown(body, todo.VerificationMarkdown)
}

func issueHistoryNote(issue *github.IssueResult) string {
	if issue.Updated {
		return "Updated GitHub issue " + issue.URL
	}
	return "Pushed to GitHub issue " + issue.URL
}

// duplicateWarning fires only for a created issue: a retry would open another
// one, while a retried update just rewrites the same issue.
func duplicateWarning(updated bool) string {
	if updated {
		return ""
	}
	return " (retrying will open a DUPLICATE issue)"
}

// renderBody makes the body fetchable from outside the dashboard. Attachments
// are stored as server-relative links that GitHub cannot resolve, so a body
// carrying any requires an absolute origin rather than silently shipping links
// that render as broken images.
func renderBody(body, baseURL string) (string, error) {
	if !todos.HasAttachmentURLs(body) {
		return body, nil
	}
	if baseURL == "" {
		return "", fmt.Errorf("todo body links attachments under %s but no base URL is configured; "+
			"GitHub cannot fetch server-relative links — pass --base-url https://gavel.example.com "+
			"or set todos.baseUrl in .gavel.yaml", todos.AttachmentURLPrefix)
	}
	return todos.AbsolutizeAttachmentURLs(body, baseURL), nil
}
