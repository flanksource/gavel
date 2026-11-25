package lifecycle

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	git "github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/ai/prfix"
	commitpkg "github.com/flanksource/gavel/commit"
	workflow "github.com/flanksource/gavel/lifecycle"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
)

type Host struct {
	DB         *captaindb.DB
	Dir        string
	Provider   git.Provider
	Repo       git.RepoURL
	Number     int
	Definition workflow.Lifecycle
}

type Preview struct {
	HeadSHA    string                `json:"headSha"`
	Definition string                `json:"definition"`
	Actions    []Action              `json:"actions"`
	Runs       []captaindb.PromptRun `json:"runs"`
}

func SessionID(url string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("gavel.pr.lifecycle:"+url))
}

func (h Host) Snapshot(ctx context.Context) (Snapshot, error) {
	pr, err := h.Provider.GetPullRequest(ctx, h.Repo, h.Number)
	if err != nil {
		return Snapshot{}, err
	}
	metadata, err := git.ReadPullRequestMetadata(ctx, h.Provider, h.Repo, git.PullRequestMetadataOptions{Number: h.Number})
	if err != nil {
		return Snapshot{}, err
	}
	activity, err := h.Provider.ListPullRequestActivity(ctx, h.Repo, h.Number)
	if err != nil {
		return Snapshot{}, err
	}
	if metadata.HeadSHA != pr.HeadSHA {
		return Snapshot{}, fmt.Errorf("PR head changed while loading context; refresh")
	}
	return Snapshot{PR: pr, Metadata: *metadata, Activity: activity}, nil
}

func (h Host) Preview(ctx context.Context, snapshot Snapshot) (Preview, error) {
	encoded, err := json.Marshal(h.Definition)
	if err != nil {
		return Preview{}, err
	}
	preview := Preview{HeadSHA: snapshot.PR.HeadSHA, Definition: fmt.Sprintf("%x", sha256.Sum256(encoded))}
	session := SessionID(snapshot.PR.URL)
	preview.Runs, err = h.DB.ListPromptRuns(ctx, captaindb.PromptRunFilter{SessionID: &session})
	if err != nil {
		return preview, err
	}
	engine, err := workflow.New(h.Definition)
	if err != nil {
		return preview, err
	}
	preview.Actions, err = Actions(engine, snapshot, stepRuns(preview.Runs))
	return preview, err
}

func stepRuns(runs []captaindb.PromptRun) []workflow.StepRun {
	result := make([]workflow.StepRun, 0, len(runs))
	for i := len(runs) - 1; i >= 0; i-- {
		run := runs[i]
		result = append(result, workflow.StepRun{Step: run.SpecProfile, State: string(run.State), PromptRunID: run.ID.String(), StartedAt: run.StartedAt, FinishedAt: run.FinishedAt})
	}
	return result
}

func (h Host) Prepare(ctx context.Context, snapshot Snapshot, selection Selection) (promptrun.Input, error) {
	preview, err := h.Preview(ctx, snapshot)
	if err != nil {
		return promptrun.Input{}, err
	}
	action, err := SelectAction(snapshot, preview.Actions, preview.Definition, selection)
	if err != nil {
		return promptrun.Input{}, err
	}
	input, err := h.input(ctx, snapshot, action, preview.Runs)
	if err != nil {
		return input, err
	}
	input.Record = h.record(snapshot, action, preview.Runs)
	_, err = promptrun.Preflight(input)
	return input, err
}

func (h Host) input(ctx context.Context, snapshot Snapshot, action Action, runs []captaindb.PromptRun) (promptrun.Input, error) {
	cfg, err := verify.LoadGavelConfig(h.Dir)
	if err != nil {
		return promptrun.Input{}, err
	}
	spec := cfg.PR.Fix
	if action.Step.Prompt == "pr.review" {
		spec = cfg.PR.Review
	}
	if override, ok := cfg.PR.Steps[action.Step.Name]; ok {
		spec = verify.Merge(spec, override)
	}
	engine, err := workflow.New(h.Definition)
	if err != nil {
		return promptrun.Input{}, err
	}
	inputs, err := engine.Inputs(action.Step, workflow.Context{Subject: snapshot.Subject(), Event: action.Event.Variables(), Runs: stepRuns(runs)})
	if err != nil {
		return promptrun.Input{}, err
	}
	inputs["lifecycle"] = true
	source, err := json.MarshalIndent(action.Event, "", "  ")
	if err != nil {
		return promptrun.Input{}, err
	}
	layers, err := prfix.Layers(prfix.ResolveOptions{Base: cfg.AI, Context: action.Step.Spec, Prompt: spec, Reference: action.Step.Prompt, Dir: h.Dir, Inputs: inputs, PR: prfix.PRContext{Number: snapshot.PR.Number, URL: snapshot.PR.URL, Title: snapshot.PR.Title, Branch: snapshot.PR.Head, BaseBranch: snapshot.PR.Base, StatusText: string(source)}})
	if err != nil {
		return promptrun.Input{}, err
	}
	sourceRepo, err := h.Repo.SourceRepository(snapshot.PR.HeadRepository)
	if err != nil {
		return promptrun.Input{}, err
	}
	target := &provider.RunTarget{RepositoryURL: sourceRepo.Clone, Branch: snapshot.PR.Head, HeadSHA: snapshot.PR.HeadSHA, Rebase: action.Event.Type == "git.conflict"}
	setup := api.Spec{Setup: &shell.Setup{Checkout: target.Checkout()}}
	setup.Setup.Checkout.Path = h.Dir
	layers = append(layers, api.RequestSpecLayer("PR lifecycle isolation", setup))
	saved, _, err := captainconfig.Load()
	if err != nil {
		return promptrun.Input{}, err
	}
	resolved, err := (captaincli.AIRuntimeOptions{}).Resolve(captaincli.AIRuntimeResolveOptions{Context: ctx, Layers: layers, Saved: saved, Cwd: h.Dir, RequireModel: true})
	if err != nil {
		return promptrun.Input{}, err
	}
	schema, err := json.Marshal(map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"success": map[string]any{"type": "boolean"}, "event": map[string]any{"type": "string", "enum": action.Step.Events}, "comment": map[string]any{"type": "string"}}, "required": []string{"success", "event", "comment"}})
	if err != nil {
		return promptrun.Input{}, err
	}
	resolved.Resolution.Spec.Prompt.SchemaJSON = schema
	resolved.Resolution.Spec.Prompt.Schema = nil
	var commits []any
	if resolved.Resolution.Spec.Workflow != nil {
		commits = commitpkg.AgentHooks(commitpkg.AgentHooksOptions{Commits: resolved.Resolution.Spec.Workflow.Commits, PRTarget: target})
	}
	return promptrun.Input{Resolved: resolved.Resolution, RuntimePresets: resolved.Presets, Config: resolved.Config, Hooks: commits, CallerOwnsCommits: len(commits) > 0, Approvals: &promptrun.ApprovalOptions{RequestedBy: "gavel-dashboard"}, Timeout: 45 * time.Minute, Repo: h.Dir}, nil
}
