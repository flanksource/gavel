// Hosted landing: a hosted run's commits are already on a branch of the organization's GitHub
// repository, so landing opens a pull request from it through the organization's GitHub App.
package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
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
// commits on, through the organization's GitHub App, titled as the user asks (the todo's title by
// default) and described from the run's commits, so no model runs on serve. Files the user leaves
// out are put back by a commit on the branch, which is taken off again if the PR isn't opened. A
// hosted run lands no other way.
func (s *Server) landHostedRun(ctx context.Context, provider land.Provider, source todoSource, todo *types.TODO, opts land.Options) (*native.RunLanding, error) {
	if opts.Via != native.LandingPR {
		return nil, hostedLandingRefusal("a hosted run lands as a pull request; land it with via pr")
	}
	repo := source.Hosted.Repo
	openPR, err := s.hostedPROpener(ctx, source, todo)
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
	base := strings.TrimSpace(opts.Base)
	if base == "" {
		cfg, err := verify.LoadRepositoryGavelConfig(source.Dir)
		if err != nil {
			return nil, fmt.Errorf("load the PR configuration of %s: %w", repo.Name, err)
		}
		base = strings.TrimSpace(cfg.PR.Base)
	}
	tip := wt.Head
	if len(opts.LeaveOut) > 0 {
		if tip, err = sources.LeaveOut(ctx, repo, branch, wt.Setup, wt.Head, opts.LeaveOut); err != nil {
			return nil, err
		}
	}
	subjects, err := sources.CommitSubjects(ctx, repo, wt.Setup, tip)
	if err == nil && len(subjects) == 0 {
		err = fmt.Errorf("%w: run %s has nothing past its setup", land.ErrNoCommits, run.PromptRunID)
	}
	var pr *model.CreatePRResult
	if err == nil {
		title := strings.TrimSpace(opts.Title)
		if title == "" {
			title = todo.Title
		}
		pr, err = openPR(ctx, model.CreatePRInput{
			Title: title, Body: hostedLandingBody(todo, subjects, opts.LeaveOut), Head: branch, Base: base, Draft: opts.Draft,
		})
	}
	if err != nil {
		if tip != wt.Head {
			err = errors.Join(err, sources.ResetRunBranch(context.WithoutCancel(ctx), repo, branch, tip, wt.Head))
		}
		return nil, err
	}
	number := pr.Number
	recorded, err := provider.Repository().RecordLanding(ctx, issueID, native.RunLanding{
		PromptRunID: run.PromptRunID, Via: native.LandingPR, TargetBranch: pr.Base, LandedSHA: tip,
		PRNumber: &number, PRURL: pr.URL, CommitCount: len(subjects),
	}, "gavel")
	if err != nil {
		return nil, fmt.Errorf("run %s opened %s but recording it failed: %w", run.PromptRunID, pr.URL, err)
	}
	return recorded, nil
}

// hostedPROpener opens a hosted run's pull request on its repository: a GitHub one through the
// organization's GitHub App, one a git connection brings through that connection, onto its default
// branch when the landing names no base. A repository pushed to gavel has no pull requests.
func (s *Server) hostedPROpener(ctx context.Context, source todoSource, todo *types.TODO) (func(context.Context, model.CreatePRInput) (*model.CreatePRResult, error), error) {
	repo := source.Hosted.Repo
	if repo.InstallationID != 0 || (repo.Source == "github" && repo.ConnectionID == nil) {
		account, err := s.todoGitHub(ctx, source, todo)
		if err != nil {
			return nil, err
		}
		return func(_ context.Context, input model.CreatePRInput) (*model.CreatePRResult, error) {
			return github.CreatePR(account, input)
		}, nil
	}
	if repo.ConnectionID == nil || s.workspaces.Sources == nil || s.workspaces.Sources.Connection == nil {
		return nil, hostedLandingRefusal(fmt.Sprintf("repository %s was pushed to gavel, not linked from GitHub or a git connection, so its runs can't land as pull requests", repo.Name))
	}
	if err := requireTodoChangeable(ctx, todo); err != nil {
		return nil, err
	}
	remote, err := s.workspaces.Sources.Connection(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("reach repository %s through its connection: %w", repo.Name, err)
	}
	return provider.NewConnectionClient(remote, repo.DefaultBranch).CreatePR, nil
}

// hostedLandingBody describes a landing's pull request: the todo it carries, its commits and
// the files the user left out of it.
func hostedLandingBody(todo *types.TODO, subjects, leftOut []string) string {
	var body strings.Builder
	fmt.Fprintf(&body, "Lands todo %s: %s\n\n## Commits\n\n", todo.DisplayID(), todo.Title)
	for _, subject := range subjects {
		fmt.Fprintf(&body, "- %s\n", subject)
	}
	if len(leftOut) > 0 {
		body.WriteString("\n## Left out\n\n")
		for _, path := range leftOut {
			fmt.Fprintf(&body, "- %s\n", path)
		}
	}
	return body.String()
}
