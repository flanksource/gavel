package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/flanksource/gavel/pr/blame"
)

func RecommendReviewers(opts Options, input blame.Request) (*blame.Recommendations, error) {
	if input.Limit <= 0 || input.HistoryLimit <= 0 {
		return nil, fmt.Errorf("reviewer limit and history limit must be positive")
	}
	pr, err := FetchPR(opts, input.Number)
	if err != nil {
		return nil, err
	}
	if pr.BaseRefOID == "" || pr.HeadRefOID == "" || pr.Author.Login == "" {
		return nil, fmt.Errorf("PR #%d requires base and head commits and an author", pr.Number)
	}
	remote, err := RepoRemote(opts)
	if err != nil {
		return nil, err
	}
	if err := ensureCommits(opts.WorkDir, remote, pr.Number, pr.BaseRefName, pr.BaseRefOID, pr.HeadRefOID); err != nil {
		return nil, err
	}
	ctx := input.Context
	if ctx == nil {
		ctx = context.Background()
	}
	result, err := blame.Analyze(ctx, blame.Options{
		Dir: opts.WorkDir, Base: pr.BaseRefOID, Head: pr.HeadRefOID, HistoryLimit: input.HistoryLimit,
		Exclude:       append(append([]string{}, input.Exclude...), pr.Author.Login),
		ResolveAuthor: func(ctx context.Context, sha string) (string, error) { return commitAuthor(ctx, opts, sha) },
		Since:         input.Since})
	if err != nil {
		return nil, err
	}
	total := len(result.Reviewers)
	if total > input.Limit {
		result.Reviewers = result.Reviewers[:input.Limit]
	}
	return &blame.Recommendations{Result: result, Provider: "github", Number: pr.Number, URL: pr.URL, Title: pr.Title, Author: pr.Author.Login, TotalReviewers: total}, nil
}

func commitAuthor(ctx context.Context, opts Options, sha string) (string, error) {
	token, err := opts.token()
	if err != nil {
		return "", err
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return "", err
	}
	repoPath, err := repoAPIPath(repo)
	if err != nil {
		return "", err
	}
	response, err := cachedGet(ctx, opts.cacheStore(), token, "/repos/"+repoPath+"/commits/"+url.PathEscape(sha), nil)
	if err != nil {
		return "", err
	}
	var commit struct {
		SHA    string `json:"sha"`
		Author *struct {
			Login string `json:"login"`
		} `json:"author"`
	}
	if err := json.Unmarshal(response.Body, &commit); err != nil {
		return "", fmt.Errorf("parse commit %s: %w", sha, err)
	}
	if commit.SHA != sha {
		return "", fmt.Errorf("expected commit %s, received %q", sha, commit.SHA)
	}
	if commit.Author == nil {
		return "", nil // GitHub has no linked account; retain the Git name and email.
	}
	if commit.Author.Login == "" {
		return "", fmt.Errorf("GitHub returned an author without a login for %s", sha)
	}
	return commit.Author.Login, nil
}
