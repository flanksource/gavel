import type { AISpecRuntimeValue } from '@flanksource/clicky-ui/ai';

export type TodoRunAgent = 'claude' | 'codex';
export type TodoRunEffort = 'low' | 'medium' | 'high' | 'xhigh' | 'max' | 'ultra';
// TodoRunDriver is the canonical runtime mode. Provider identity comes from model.
export type TodoRunDriver = 'api' | 'agent' | 'cli' | 'cmux';

// TodoRunOptions is the run POST body's options: the api.Spec under its own
// `spec` key plus the run-orchestration extras below. Dirty worktree,
// auto-commit, dry-run, and checks all live in the spec now
// (setup.checkout.dirty, workflow.commits[].on/dryRun, workflow.verify); only the
// prompt/driver selection and the resume decision sit alongside it.
//
// The spec is nested rather than inlined because the Go payload stopped
// embedding api.Spec: api.Spec declares a value-receiver MarshalJSON, so an
// embedding struct inherited it and marshalled to a bare spec, silently dropping
// driver/runMode/resume. Nesting also dissolves the wire collision between this
// payload's `mode`/`resume` and api.Model's own.
export interface TodoRunOptions {
  // Spec is the captain api.Spec: model/mode/effort flat inside it;
  // prompt/budget/permissions/setup/workflow/sessionId nested, mirroring
  // clicky's AISpecRuntimeValue.
  spec?: AISpecRuntimeValue;
  runtimeProfile?: string;
  // step is the lifecycle step name the request dispatches — the sole
  // behaviour-selecting field POST /api/todos/run accepts now (the endpoint
  // decodes strictly and rejects runMode/driver/prompt at the top level).
  // driver/runMode/plan/prompt below stay for the client's own bookkeeping
  // (storage, dropdown labels, the advanced dialog's editors) and are folded
  // into `step` — or into spec.mode for the runtime mechanism — before a
  // request is built; see run.tsx's requestStepFor/buildTodoRunPayload.
  step?: string;
  // Driver is the authoritative canonical runtime mode.
  driver?: TodoRunDriver;
  // runMode is the behaviour class the run executes as (run/plan). Supersedes
  // the `plan` bool; the server accepts both (plan:true is treated as
  // runMode:plan).
  runMode?: 'run' | 'plan';
  // prompt names the template to run — run, plan, triage, or a name declared in
  // .gavel.yaml todos.prompts. It is a separate axis from runMode: each prompt
  // declares the class it runs as, so several prompts share one. Send one or the
  // other; asserting a class that disagrees with the prompt is rejected.
  prompt?: string;
  // Plan-only run: the agent proposes an implementation plan without changing
  // code. Requires cmux mode. Legacy — prefer runMode.
  plan?: boolean;
  // Resume the todo's prior agent session (claude --resume) instead of starting a
  // fresh one. Stays a sibling flag rather than a spec field: a fresh run also
  // carries a (minted) sessionId, so resume can't be inferred from spec.sessionId.
  resume?: boolean;
  // Dispatch even though the todo already has a live run owned by a running
  // process: the two runs proceed in parallel. Set only in answer to the
  // server's 409 run_owned_elsewhere, never as a default.
  force?: boolean;
}

export interface TodoRunResponse {
  status: 'started' | 'skipped' | 'dry_run';
  ref: string;
  dir: string;
  provider?: string;
  driver?: TodoRunDriver;
  // runtimeMode is the mechanism the run dispatched on (api | agent | cli | cmux).
  runtimeMode?: string;
  runtimeProfile?: string;
  model?: string;
  effort?: TodoRunEffort;
  plan?: boolean;
  resume?: boolean;
  // Session id the run uses; lets the UI follow the session log immediately.
  sessionId?: string;
  timeout: string;
  maxBudget?: number;
  maxTurns?: number;
  // Whether the run will auto-commit the agent's changes when it finishes.
  commit?: boolean;
  message: string;
}

// Preview of the exact prompt a run would dispatch, shown in the advanced run
// dialog before the user starts the run.
export interface TodoRunPreviewResponse {
  prompt: string;
  specYaml: string;
  model?: string;
  provider?: string;
  runtimeMode?: string;
  effort?: TodoRunEffort;
  plan?: boolean;
  count: number;
}
