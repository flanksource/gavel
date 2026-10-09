// Package prcreate opens a pull request from existing commits: it cherry-picks
// them in order onto a fresh worktree branched from a base ref, pushes the
// resulting topic branch to origin and opens the PR, leaving the caller's
// working tree untouched.
package prcreate

import (
	"context"
	"errors"
	"fmt"

	"github.com/flanksource/captain/pkg/aiflags"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/commons/logger"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
)

// Deps are the external calls Create makes; tests substitute them to avoid
// GitHub and the LLM.
type Deps struct {
	Preflight       func(context.Context, provider.Options) error
	CreatePR        func(provider.Options, model.CreatePRInput) (*model.CreatePRResult, error)
	GenerateContent func(ctx context.Context, in commitpkg.PRContentInput) (commitpkg.PRContent, error)
}

// DefaultDeps wires the PR provider and AI PR content generation.
func DefaultDeps() Deps {
	return Deps{Preflight: provider.PreflightCreate, CreatePR: func(opts provider.Options, in model.CreatePRInput) (*model.CreatePRResult, error) {
		return provider.CreatePR(opts.Context, opts, in)
	}, GenerateContent: commitpkg.GeneratePRContent}
}

func (d Deps) validate() error {
	if d.CreatePR == nil {
		return errors.New("prcreate: Input.Deps.CreatePR is nil; pass prcreate.DefaultDeps()")
	}
	if d.GenerateContent == nil {
		return errors.New("prcreate: Input.Deps.GenerateContent is nil; pass prcreate.DefaultDeps()")
	}
	if d.Preflight == nil {
		return errors.New("prcreate: Input.Deps.Preflight is nil; pass prcreate.DefaultDeps()")
	}
	return nil
}

type Input struct {
	// SHAs are cherry-picked onto Base in the given order.
	SHAs []string
	// Base is the git ref to branch from (e.g. origin/main); a leading
	// "origin/" is stripped for the PR's base branch.
	Base  string
	Draft bool
	// Mainline selects the parent (1-based) when cherry-picking merge commits.
	Mainline int
	// Repo is a GitHub owner/repo or Azure repository URL; empty resolves origin.
	Repo  string
	Flags aiflags.ModelFlags
	// Saved is the Captain configuration snapshot; nil loads it from disk.
	Saved *captainconfig.Config
	Deps  Deps
}

type Result struct {
	PR          *model.CreatePRResult
	Content     commitpkg.PRContent
	TopicBranch string
	// TopicHead is the topic branch tip after all cherry-picks.
	TopicHead string
}

// ConflictError reports a cherry-pick conflict. The worktree is kept at
// Worktree on Branch so the conflict can be resolved by hand.
type ConflictError struct {
	SHA, Worktree, Branch string
	Err                   error
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("git cherry-pick %s: %v", e.SHA[:8], e.Err)
}

func (e *ConflictError) Unwrap() error { return e.Err }

// Create cherry-picks in.SHAs onto a fresh worktree off in.Base, pushes the
// topic branch named by AI-generated PR content and opens the PR. The
// worktree is removed afterwards unless a cherry-pick conflicts.
func Create(ctx context.Context, repoRoot string, in Input) (*Result, error) {
	if err := in.Deps.validate(); err != nil {
		return nil, err
	}
	if err := in.Deps.Preflight(ctx, provider.Options{WorkDir: repoRoot, Repo: in.Repo, Context: ctx}); err != nil {
		return nil, fmt.Errorf("PR provider preflight: %w", err)
	}
	if in.Saved == nil {
		saved, _, err := captainconfig.Load()
		if err != nil {
			return nil, fmt.Errorf("load Captain configuration: %w", err)
		}
		in.Saved = &saved
	}
	shas, err := preflight(repoRoot, in)
	if err != nil {
		return nil, err
	}
	ws, err := openWorktree(repoRoot, in.Base, shas[0])
	if err != nil {
		return nil, err
	}
	keepWorktree := false
	defer func() {
		if keepWorktree {
			return
		}
		if err := removeWorktree(repoRoot, ws.path); err != nil {
			logger.Warnf("worktree cleanup: %v", err)
		}
	}()
	if err := cherryPickAll(ws, shas, in); err != nil {
		var conflict *ConflictError
		keepWorktree = errors.As(err, &conflict)
		return nil, err
	}
	return publish(ctx, repoRoot, ws, len(shas), in)
}

func cherryPickAll(ws worktree, shas []string, in Input) error {
	for _, sha := range shas {
		err := gitCherryPick(ws.path, sha, in.Mainline)
		if err == nil {
			continue
		}
		if isNoOpCherryPick(ws.path) {
			baseLocal, _ := splitBaseRef(in.Base)
			return fmt.Errorf("commit %s is already in %s; nothing to PR", sha[:8], baseLocal)
		}
		return &ConflictError{SHA: sha, Worktree: ws.path, Branch: ws.tmpBranch, Err: err}
	}
	return nil
}

func publish(ctx context.Context, repoRoot string, ws worktree, picked int, in Input) (*Result, error) {
	prIn, err := contentInput(ws.path, picked, in)
	if err != nil {
		return nil, err
	}
	content, err := in.Deps.GenerateContent(ctx, prIn)
	if err != nil {
		return nil, fmt.Errorf("generate PR content: %w", err)
	}
	topic := content.Branch + "-" + ws.suffix
	if err := runGitQuiet(ws.path, "branch", "-m", ws.tmpBranch, topic); err != nil {
		return nil, fmt.Errorf("rename branch %s -> %s: %w", ws.tmpBranch, topic, err)
	}
	head, err := captureGit(ws.path, "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("resolve topic head: %w", err)
	}
	if err := gitPushTopic(ws.path, topic); err != nil {
		return nil, fmt.Errorf("git push %s: %w", topic, err)
	}
	baseLocal, _ := splitBaseRef(in.Base)
	created, err := in.Deps.CreatePR(provider.Options{WorkDir: repoRoot, Repo: in.Repo, Context: ctx}, model.CreatePRInput{
		Title: content.Title, Body: content.Body, Head: topic, Base: baseLocal, Draft: in.Draft,
	})
	if err != nil {
		return nil, fmt.Errorf("create PR: %w", err)
	}
	if created == nil {
		return nil, fmt.Errorf("create PR for %s: no result returned", topic)
	}
	logger.Infof("Opened PR #%d against %s: %s", created.Number, created.Base, created.URL)
	return &Result{PR: created, Content: content, TopicBranch: topic, TopicHead: head}, nil
}
