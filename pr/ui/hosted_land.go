// Hosted landing: a hosted run's commits are already on a branch of the organization's GitHub
// repository, so landing opens a pull request from it through the organization's GitHub App.
package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/todos/land"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
)

// hostedLandingRefusal is a landing a hosted run can't make.
type hostedLandingRefusal string

func (r hostedLandingRefusal) Error() string { return string(r) }

// landHostedRun lands todo's newest run step as a pull request from the branch serve put its
// commits on, through the organization's GitHub App, titled from the todo and described from the
// run's commits, so no model runs on serve. A hosted run lands no other way.
func (s *Server) landHostedRun(ctx context.Context, provider land.Provider, source todoSource, todo *types.TODO, opts land.Options) (*native.RunLanding, error) {
	if opts.Via != native.LandingPR {
		return nil, hostedLandingRefusal("a hosted run lands as a pull request; land it with via pr")
	}
	repo := source.Hosted.Repo
	if repo.Source != "github" {
		return nil, hostedLandingRefusal(fmt.Sprintf("repository %s was pushed to gavel, not linked from GitHub, so its runs can't land as pull requests", repo.Name))
	}
	account, err := s.todoGitHub(ctx, source, todo)
	if err != nil {
		return nil, err
	}
	issueID, err := uuid.Parse(todo.ID)
	if err != nil {
		return nil, fmt.Errorf("native TODO has invalid ID %q: %w", todo.ID, err)
	}
	run, err := provider.Repository().LatestRunWorktree(ctx, provider.Captain(), issueID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", land.ErrNoRunWorkspace, err)
	}
	if landed := run.Landing; landed != nil {
		return nil, fmt.Errorf("%w: run %s landed via %s on %s at %s", native.ErrAlreadyLanded,
			run.PromptRunID, landed.Via, landed.TargetBranch, landed.LandedSHA)
	}
	wt := run.Worktree
	sources := s.workspaces.Sources
	branch, err := sources.RunBranchAt(ctx, repo, issueID, wt.Head)
	if err != nil {
		return nil, fmt.Errorf("land run %s: %w", run.PromptRunID, err)
	}
	subjects, err := sources.CommitSubjects(ctx, repo, wt.Setup, wt.Head)
	if err != nil {
		return nil, err
	}
	if len(subjects) == 0 {
		return nil, fmt.Errorf("%w: run %s has nothing past its setup", land.ErrNoCommits, run.PromptRunID)
	}
	base := strings.TrimSpace(opts.Base)
	if base == "" {
		cfg, err := verify.LoadRepositoryGavelConfig(source.Dir)
		if err != nil {
			return nil, fmt.Errorf("load the PR configuration of %s: %w", repo.Name, err)
		}
		base = strings.TrimSpace(cfg.PR.Base)
	}
	pr, err := github.CreatePR(account, model.CreatePRInput{
		Title: todo.Title, Body: hostedLandingBody(todo, subjects), Head: branch, Base: base, Draft: opts.Draft,
	})
	if err != nil {
		return nil, err
	}
	number := pr.Number
	recorded, err := provider.Repository().RecordLanding(ctx, issueID, native.RunLanding{
		PromptRunID: run.PromptRunID, Via: native.LandingPR, TargetBranch: pr.Base, LandedSHA: wt.Head,
		PRNumber: &number, PRURL: pr.URL, CommitCount: len(subjects),
	}, "gavel")
	if err != nil {
		return nil, fmt.Errorf("run %s opened %s but recording it failed: %w", run.PromptRunID, pr.URL, err)
	}
	return recorded, nil
}

// hostedLandingBody describes a landing's pull request: the todo it carries and its commits.
func hostedLandingBody(todo *types.TODO, subjects []string) string {
	var body strings.Builder
	fmt.Fprintf(&body, "Lands todo %s: %s\n\n## Commits\n\n", todo.DisplayID(), todo.Title)
	for _, subject := range subjects {
		fmt.Fprintf(&body, "- %s\n", subject)
	}
	return body.String()
}
