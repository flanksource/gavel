package lifecycle

import (
	"fmt"
	"os/exec"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/todos/types"
)

func (p *prExecution) Name() string          { return "gavel-pr-turn" }
func (p *prExecution) Phases() []agent.Phase { return []agent.Phase{agent.PhaseTurn} }

func (p *prExecution) PreRun(hc *agent.HookContext) error {
	snapshot, err := p.client.Status(hc.Context, p.number, model.StatusOptions{Threads: true})
	if err != nil {
		return err
	}
	if snapshot == nil || snapshot.PR == nil {
		return fmt.Errorf("linked PR provider returned no status before dispatch")
	}
	if snapshot.PR.State != "OPEN" || snapshot.PR.HeadRefOID != p.target.HeadSHA || snapshot.PR.HeadRefName != p.target.Branch || snapshot.PR.HeadRepositoryURL != p.target.RepositoryURL {
		return fmt.Errorf("linked PR source changed since the run was prepared; refresh before dispatch")
	}
	return p.target.FetchHead(hc.Context, p.workDir)
}

func prTurnEnvelope(hc *agent.HookContext) (*types.PRRunEnvelope, error) {
	if hc.Response == nil {
		return nil, fmt.Errorf("PR turn has no response")
	}
	if hc.Response.TerminalOutcome != nil && hc.Response.TerminalOutcome.Kind == api.TerminalOutcomeQuestions {
		return nil, nil
	}
	raw := hc.Response.Text
	if hc.Response.StructuredData != nil {
		var err error
		raw, err = structuredDataText(hc.Response.StructuredData)
		if err != nil {
			return nil, err
		}
	}
	return captainai.ParseStructured(raw, (*types.PRRunEnvelope).Validate)
}

func completedPRTurn(hc *agent.HookContext) (bool, error) {
	if hc.Failed || hc.PostErr != nil || hc.Phase != agent.PhaseTurn {
		return false, nil
	}
	env, err := prTurnEnvelope(hc)
	return env != nil && env.EndStatus == types.EndCompleted, err
}

func (p *prExecution) Post(hc *agent.HookContext, phase agent.Phase) error {
	if !p.write || phase != agent.PhaseTurn {
		return nil
	}
	completed, err := completedPRTurn(hc)
	if err != nil || !completed {
		return err
	}
	if hc.Request == nil || hc.Request.Setup == nil || hc.Request.Setup.Cwd == "" {
		return fmt.Errorf("PR turn has no worktree")
	}
	dir := hc.Request.Setup.Cwd
	cmd := exec.CommandContext(hc.Context, "git", "status", "--porcelain", "--untracked-files=normal")
	cmd.Dir = dir
	status, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("inspect PR turn worktree: %w: %s", err, status)
	}
	if strings.TrimSpace(string(status)) != "" {
		return fmt.Errorf("PR turn has uncommitted changes; refusing to push or resolve threads: %s", status)
	}
	head, err := p.target.PushHead(hc.Context, dir)
	if err != nil {
		return err
	}
	env, err := prTurnEnvelope(hc)
	if err != nil {
		return err
	}
	snapshot, err := p.client.Status(hc.Context, p.number, model.StatusOptions{Threads: true})
	if err != nil {
		return err
	}
	if err := validateAddressedThreads(snapshot, head, env.AddressedThreadIDs); err != nil {
		return err
	}
	for _, id := range env.AddressedThreadIDs {
		if err := p.client.ResolveThread(hc.Context, p.number, id); err != nil {
			return fmt.Errorf("resolve addressed PR thread %s: %w", id, err)
		}
	}
	snapshot, err = p.client.Status(hc.Context, p.number, model.StatusOptions{Threads: true, Logs: true, TailLogs: 40})
	if err != nil {
		return err
	}
	if err := validateAddressedThreads(snapshot, head, nil); err != nil {
		return err
	}
	p.snapshot = snapshot
	if err := setPRStatusContext(&hc.Request.Prompt, snapshot); err != nil {
		return err
	}
	hc.Notify("PR #%d pushed revision %s; addressed threads: %s. The fixture will verify the entire PR at this revision.", p.number, head, strings.Join(env.AddressedThreadIDs, ", "))
	return nil
}

func validateAddressedThreads(snapshot *model.StatusSnapshot, head string, ids []string) error {
	if snapshot == nil || snapshot.PR == nil || snapshot.PR.State != "OPEN" || snapshot.PR.HeadRefOID != head {
		return fmt.Errorf("PR status does not match the pushed open PR head %s", head)
	}
	threads := map[string]model.PRComment{}
	for _, thread := range snapshot.PR.ReviewThreads {
		threads[thread.ThreadID] = thread
	}
	for _, id := range ids {
		thread, ok := threads[id]
		if !ok {
			return fmt.Errorf("addressed thread %q does not belong to this PR", id)
		}
		if !thread.IsResolved && !thread.CanResolve {
			return fmt.Errorf("addressed thread %q cannot be resolved by the current user", id)
		}
	}
	return nil
}
