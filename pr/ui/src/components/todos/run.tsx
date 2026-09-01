import { useCallback, type ComponentType } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ChatModel } from "@flanksource/clicky-ui/chat";
import { effortOptionsForModel, promptRuntimeValueToPayload, reconcileModelCapabilities, type AISpecRuntimeValue, type RuntimeBarValue } from "@flanksource/clicky-ui/ai";
import { UiListChecks, UiListDashes, UiPlay, type IconProps } from "@flanksource/clicky-ui/icons";
import type { TodoRunDriver, TodoRunEffort, TodoRunOptions, TodoRunPreviewResponse, TodoRunResponse } from "../../types";
import { todoQuery } from "./format";
import { settingsRunContextQuery } from "../settings/queries";
import { invalidateTodoCaches, todoMutationJSON, TodoMutationError } from "./todoMutations";
export { TodoRunEffortBadge, todoRunEffortPresentation } from "./TodoRunEffortBadge";
import {
  agentForRuntime,
  PROVIDERS,
  type RunModeCatalog,
  type RunContext,
} from "./providers";

// RunMode is the behaviour class a run executes as: run (implement and commit)
// or plan (neither). Verification is a fixture-backed issue lifecycle action in
// the Verification tab, not an agent run mode.
export type RunMode = "run" | "plan";
// TodoRunAction is the prompt the dialog runs. Triage shares plan's behaviour
// class but is a different prompt, which is why action and mode are separate
// types: several prompts map onto one class.
export type TodoRunAction = "run" | "plan" | "triage";
export const TODO_RUN_ACTIONS: readonly TodoRunAction[] = ["run", "plan", "triage"] as const;
export type TodoRunRuntimeMode = "cmux" | "agent" | "cli" | "api";

// Runs auto-commit by default — the old commit=true default, now expressed as a
// single commit policy on the spec's Workflow.Commits. `on: "run"` keeps the
// dashboard's existing shape (one commit once the run finishes) rather than the
// per-turn fixup chain, which stays opt-in while a todo run executes in the
// user's live working tree. The advanced dialog's Commit section can turn it off;
// a plan-only run never commits because the plan action omits this workflow.
const AUTO_COMMIT: Pick<AISpecRuntimeValue, "workflow"> = { workflow: { commits: [{ on: "run", gates: "full" }] } };

export const defaultRunOptions: TodoRunOptions = { driver: "cli", spec: { effort: "medium", ...AUTO_COMMIT } };

type RunChoiceState = {
  last: Partial<Record<TodoRunAction, TodoRunOptions>>;
  recentAdvanced: Partial<Record<TodoRunAction, TodoRunOptions[]>>;
};

// v2: TodoRunOptions moved model/mode/effort/budget under a nested `spec`.
// A v1 entry would still parse, but every spec field would read as unset and the
// remembered model would silently revert to the mode default, so the version
// is bumped to discard it instead.
const RUN_CHOICE_STORAGE_KEY = "gavel.pr-ui.todoRunChoices.v2";

export const runActionConfig: Record<TodoRunAction, { label: string; detail: string; icon: ComponentType<IconProps>; title: string }> = {
  run: { label: "Run", detail: "implement", icon: UiPlay, title: "Run todo" },
  plan: { label: "Plan", detail: "plan only", icon: UiListDashes, title: "Plan todo" },
  triage: {
    label: "Triage",
    detail: "compact + review fixture",
    icon: UiListChecks,
    title: "Compact the description and review the verification fixture",
  },
};

const RUNTIME_MODE_CONFIG: Record<TodoRunRuntimeMode, { label: string }> = {
  cmux: { label: "cmux" },
  agent: { label: "Agent" },
  cli: { label: "cli" },
  api: { label: "API" },
};

function emptyRunChoiceState(): RunChoiceState {
  return { last: {}, recentAdvanced: {} };
}

function cloneRunOptions(options: TodoRunOptions): TodoRunOptions {
  return JSON.parse(JSON.stringify(options)) as TodoRunOptions;
}

// runSpec is the api.Spec half of a run's options. The spec is nested under its
// own key rather than inlined (see TodoRunOptions), so the many helpers that only
// care about model/mode/effort read it through here.
export function runSpec(options: TodoRunOptions): AISpecRuntimeValue {
  return options.spec ?? {};
}

function normalizeRunOptions(action: TodoRunAction, options: TodoRunOptions): TodoRunOptions {
  const next = cloneRunOptions(options);
  if (action === "triage") {
    // The prompt name is the whole request: triage declares its own behaviour
    // class, so asserting a mode alongside it would be rejected as contradictory.
    next.prompt = "triage";
    delete next.runMode;
    delete next.plan;
    return next;
  }
  delete next.prompt;
  if (action === "plan") {
    next.runMode = "plan";
    next.plan = true;
  } else {
    next.runMode = "run";
    delete next.plan;
  }
  return next;
}

function coerceRunOptions(value: unknown): TodoRunOptions | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) return null;
  return value as TodoRunOptions;
}

function readRunChoiceState(): RunChoiceState {
  try {
    if (typeof localStorage === "undefined") return emptyRunChoiceState();
    const raw = localStorage.getItem(RUN_CHOICE_STORAGE_KEY);
    if (!raw) return emptyRunChoiceState();
    const parsed = JSON.parse(raw) as Partial<RunChoiceState>;
    const state = emptyRunChoiceState();
    const last = parsed.last ?? {};
    const recent = parsed.recentAdvanced ?? {};
    for (const action of TODO_RUN_ACTIONS) {
      const lastOptions = coerceRunOptions(last[action]);
      if (lastOptions) state.last[action] = normalizeRunOptions(action, lastOptions);
      const recentOptions = Array.isArray(recent[action]) ? recent[action] : [];
      state.recentAdvanced[action] = recentOptions
        .map(coerceRunOptions)
        .filter((item): item is TodoRunOptions => !!item)
        .map(item => normalizeRunOptions(action, item))
        .slice(0, 3);
    }
    return state;
  } catch {
    return emptyRunChoiceState();
  }
}

function writeRunChoiceState(state: RunChoiceState): void {
  try {
    if (typeof localStorage === "undefined") return;
    localStorage.setItem(RUN_CHOICE_STORAGE_KEY, JSON.stringify(state));
  } catch {
    // Persistence is best-effort; unavailable storage should not block a run.
  }
}

function sortForKey(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(sortForKey);
  if (!value || typeof value !== "object") return value;
  return Object.fromEntries(
    Object.entries(value as Record<string, unknown>)
      .filter(([, entry]) => entry !== undefined)
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([key, entry]) => [key, sortForKey(entry)]),
  );
}

export function runOptionsKey(options: TodoRunOptions): string {
  return JSON.stringify(sortForKey(options));
}

function actionFromRunOptions(options: TodoRunOptions): TodoRunAction {
  if (options.prompt === "triage") return "triage";
  return options.plan || options.runMode === "plan" ? "plan" : "run";
}

export interface TodoRunContextState {
  context: RunContext | null;
  loading: boolean;
  error: string;
}

function unavailableRunContextError(context: RunContext): string {
  if (context.runtimes.length === 0) return "Captain returned no runtime catalog";
  if (context.modes.some(runtime => runtime.models.length > 0)) return "";
  const details = context.modes.map(runtime => runtime.modelError?.trim()).filter(Boolean);
  return details[0] || "Captain returned no run models";
}

export function useTodoRunContext(enabled = true): TodoRunContextState {
  const query = useQuery({ ...settingsRunContextQuery(), enabled });
  if (!enabled) return { context: null, loading: false, error: "" };
  if (query.error) {
    return { context: null, loading: query.isFetching, error: query.error instanceof Error ? query.error.message : "Failed to load run context" };
  }
  const context = query.data ?? null;
  if (context && (
    !Array.isArray(context.modes) ||
    !Array.isArray(context.runtimes) ||
    !Array.isArray(context.models) ||
    !Array.isArray(context.efforts) ||
    !Array.isArray(context.tools)
  )) {
    return { context: null, loading: false, error: "Captain returned an invalid run context" };
  }
  return { context, loading: query.isFetching, error: context ? unavailableRunContextError(context) : "" };
}

export function TodoRunContextError({ error }: { error: string }) {
  if (!error) return null;
  return <div role="alert" className="max-w-sm text-xs text-red-600">{error}</div>;
}

// promptDefaultFor is the runtime the server resolved for one action's prompt.
// It outranks defaultMode because it already accounts for the prompt's own
// frontmatter — todos-triage.prompt and todos-plan.prompt pin `model: claude`
// and declare a per-tool policy only the Claude transports carry, so seeding
// them from a codex account default produced a run Captain refuses.
function promptDefaultFor(context: RunContext, action: TodoRunAction): { mode?: string; model?: string } {
  return context.promptDefaults?.[action] ?? {};
}

function modeById(context: RunContext, id: string | undefined, model?: string): RunModeCatalog | undefined {
  if (!id) return undefined;
  const agent = agentForRuntime(context, id, model);
  return context.modes.find(runtime => runtime.id === id && runtime.agent === agent && runtime.models.length > 0);
}

function primaryModeForAction(context: RunContext, action: TodoRunAction): RunModeCatalog {
  const promptDefault = promptDefaultFor(context, action);
  return modeById(context, promptDefault.mode, promptDefault.model)
    ?? modeById(context, context.defaultMode)
    ?? context.modes.find(runtime => runtime.models.length > 0)
    ?? (() => { throw new Error("Captain returned no run models"); })();
}

function modeForOptions(context: RunContext, options: TodoRunOptions): RunModeCatalog {
  const spec = runSpec(options);
  const requested = spec.mode || options.driver || "";
  const actionDefault = promptDefaultFor(context, actionFromRunOptions(options));
  return (
    modeById(context, requested, spec.model) ??
    modeById(context, actionDefault.mode, actionDefault.model) ??
    modeById(context, context.defaultMode) ??
    context.modes.find(runtime => runtime.models.length > 0) ??
    (() => { throw new Error("Captain returned no run models"); })()
  );
}

function runtimeModeForCatalog(runtime: RunModeCatalog): TodoRunRuntimeMode {
  if (runtime.id in RUNTIME_MODE_CONFIG) return runtime.id as TodoRunRuntimeMode;
  throw new Error(`Invalid run mode ${JSON.stringify(runtime.id)}`);
}

function runtimeModeLabel(mode: TodoRunRuntimeMode): string {
  return RUNTIME_MODE_CONFIG[mode].label;
}

function modelsForRunMode(runtime: RunModeCatalog): RunModeCatalog["models"] {
  return runtime.models;
}

const ALL_EFFORTS: TodoRunEffort[] = ["low", "medium", "high", "xhigh", "max", "ultra"];

function modelForRunMode(runtime: RunModeCatalog, modelID: string | undefined): ChatModel {
  const model = modelsForRunMode(runtime).find(item => item.id === modelID)
    ?? modelsForRunMode(runtime).find(item => item.id === runtime.defaultModel)
    ?? modelsForRunMode(runtime)[0];
  if (!model) throw new Error(runtime.modelError || `Captain returned no models for ${runtime.label}`);
  return model;
}

function contextEfforts(context: RunContext): TodoRunEffort[] {
  return context.efforts.length > 0 ? context.efforts : ALL_EFFORTS;
}

export function shortTodoRunModelName(id: string | undefined): string {
  let label = (id || "").trim();
  if (!label) return "default";
  const slash = label.lastIndexOf("/");
  if (slash >= 0) label = label.slice(slash + 1);
  for (const prefix of ["claude-agent-", "claude-code-", "claude-", "codex-"]) {
    if (label.startsWith(prefix)) {
      label = label.slice(prefix.length);
      break;
    }
  }
  const humanClaude = label.toLowerCase().trim().match(/^(?:claude\s+)?(fable|opus|sonnet|haiku)(?:\s+(\d+(?:\.\d+)*))?(?:\s+(.+))?$/);
  if (humanClaude) {
    const [, tier, version, rest] = humanClaude;
    const head = version ? `${tier}-${version}` : tier;
    const tail = (rest ?? "").trim().replace(/\s+/g, "-");
    return tail ? `${head}-${tail}` : head;
  }
  if (label.toLowerCase().startsWith("gpt-")) return label.toLowerCase();

  const parts = label.toLowerCase().split("-").filter(Boolean);
  if (["fable", "opus", "sonnet", "haiku"].includes(parts[0] ?? "")) {
    const version: string[] = [];
    let index = 1;
    while (index < parts.length && /^\d+$/.test(parts[index]!)) {
      version.push(parts[index]!);
      index += 1;
    }
    const head = version.length > 0 ? `${parts[0]}-${version.join(".")}` : parts[0];
    const rest = parts.slice(index).join("-");
    return rest ? `${head}-${rest}` : head;
  }
  return label;
}

function labelForRunModel(runtime: RunModeCatalog, modelID: string): string {
  const model = modelsForRunMode(runtime).find(item => item.id === modelID);
  return model?.label || modelID;
}

function runOptionsForModeModel(action: TodoRunAction, runtime: RunModeCatalog, modelID: string, effort: TodoRunEffort = "medium"): TodoRunOptions {
  const spec = reconcileModelCapabilities({
    mode: runtime.id,
    model: modelID || runtime.defaultModel,
    effort,
    ...(action === "run" ? AUTO_COMMIT : {}),
  } satisfies AISpecRuntimeValue, modelForRunMode(runtime, modelID), ALL_EFFORTS);
  return normalizeRunOptions(action, { driver: runtime.driver, spec });
}

export function runButtonQualifierForOptions(options: TodoRunOptions, context: RunContext): string {
  const runtime = modeForOptions(context, options);
  const model = runSpec(options).model || runtime.defaultModel;
  return `(${runtimeModeLabel(runtimeModeForCatalog(runtime))}:${shortTodoRunModelName(labelForRunModel(runtime, model))})`;
}

// todoRunModeLabel is the runtime mechanism a run would use (Agent/cmux/cli/API),
// resolved from the run options against the runtime catalog — the same derivation
// the run buttons use, exposed for the start-of-session hero's "Runtime" chip.
export function todoRunModeLabel(options: TodoRunOptions, context: RunContext): string {
  return runtimeModeLabel(runtimeModeForCatalog(modeForOptions(context, options)));
}

export function runButtonLabelForOptions(action: TodoRunAction, options: TodoRunOptions, context: RunContext): string {
  return `${runActionConfig[action].label} ${runButtonQualifierForOptions(options, context)}`;
}

export function todoRunButtonPresentation(options: TodoRunOptions, context: RunContext) {
  const runtime = modeForOptions(context, options);
  const spec = runSpec(options);
  const modelID = spec.model || runtime.defaultModel;
  const model = modelForRunMode(runtime, modelID);
  const provider = PROVIDERS.find(item => item.id === runtime.agent);
  const supportedEfforts = effortOptionsForModel(model, contextEfforts(context));
  const effort = spec.effort && supportedEfforts.includes(spec.effort)
    ? spec.effort as TodoRunEffort
    : undefined;

  return {
    provider,
    model: shortTodoRunModelName(labelForRunModel(runtime, modelID)),
    effort,
  };
}

export function defaultRunOptionsForAction(action: TodoRunAction, context?: RunContext | null): TodoRunOptions {
  if (context) {
    const runtime = primaryModeForAction(context, action);
    return runOptionsForModeModel(action, runtime, promptDefaultFor(context, action).model || runtime.defaultModel);
  }
  return normalizeRunOptions(action, defaultRunOptions);
}

export function reconcileTodoRunOptions(action: TodoRunAction, options: TodoRunOptions, context: RunContext): TodoRunOptions {
  const normalized = normalizeRunOptions(action, options);
  const runtime = modeForOptions(context, normalized);
  const spec = runSpec(normalized);
  const modelIsCurrent = !!spec.model && runtime.models.some(model => model.id === spec.model);
  const model = modelIsCurrent ? spec.model! : runtime.defaultModel;
  return normalizeRunOptions(action, {
    ...normalized,
    driver: runtime.driver,
    spec: reconcileModelCapabilities({ ...spec, mode: runtime.id, model }, modelForRunMode(runtime, model), contextEfforts(context)),
  });
}

export function loadLastTodoRunOptions(action: TodoRunAction, context?: RunContext | null): TodoRunOptions {
	const state = readRunChoiceState();
	const options = normalizeRunOptions(action, state.last[action] ?? defaultRunOptionsForAction(action, context));
	return context ? reconcileTodoRunOptions(action, options, context) : options;
}

export function loadRecentAdvancedTodoRunOptions(action: TodoRunAction, context?: RunContext | null): TodoRunOptions[] {
	const state = readRunChoiceState();
	const seen = new Set<string>();
	return (state.recentAdvanced[action] ?? [])
		.map(item => context ? reconcileTodoRunOptions(action, item, context) : normalizeRunOptions(action, item))
		.filter(item => {
			const key = runOptionsKey(item);
			if (seen.has(key)) return false;
			seen.add(key);
			return true;
		});
}

export function rememberTodoRunOptions(action: TodoRunAction, options: TodoRunOptions, advanced = false): TodoRunOptions {
  const nextOptions = normalizeRunOptions(action, options);
  const state = readRunChoiceState();
  state.last[action] = nextOptions;
  if (advanced) {
    const nextKey = runOptionsKey(nextOptions);
    const recent = (state.recentAdvanced[action] ?? []).filter(item => runOptionsKey(item) !== nextKey);
    state.recentAdvanced[action] = [nextOptions, ...recent].slice(0, 3);
  }
  writeRunChoiceState(state);
  return nextOptions;
}

export function rememberTodoRunOptionsForMode(options: TodoRunOptions, advanced = false): TodoRunOptions {
  return rememberTodoRunOptions(actionFromRunOptions(options), options, advanced);
}

// useTodoRun POSTs a run for one native TODO in a workspace.
export function useTodoRun(dir: string) {
  const client = useQueryClient();
  const mutation = useMutation({
    mutationKey: ["todos", "run", { dir: dir.trim() }],
    mutationFn: ({ ref, options }: { ref: string; options: TodoRunOptions }) => todoMutationJSON<TodoRunResponse>(
      `/api/todos/run?${todoQuery(dir)}`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ref, ...options }),
      },
      `Failed to run todo ${ref}`,
    ),
    onSuccess: (_result, { ref }) => invalidateTodoCaches(client, dir, ref),
  });

  const run = useCallback(
    async (ref: string, options: TodoRunOptions = defaultRunOptions): Promise<TodoRunResponse | null> => {
      const cleaned = ref.trim();
      if (!cleaned || mutation.isPending) return null;
      try {
        return await mutation.mutateAsync({ ref: cleaned, options });
      } catch (err) {
        // The todo already has a live run on a process that is still going. That
        // is a decision, not a failure: running both is allowed once confirmed.
        if (!options.force && err instanceof TodoMutationError && err.status === 409) {
          if (!window.confirm(`${err.message}\n\nStart a second run in parallel?`)) return null;
          try {
            return await mutation.mutateAsync({ ref: cleaned, options: { ...options, force: true } });
          } catch {
            return null;
          }
        }
        return null;
      }
    },
    [mutation],
  );

  const result = mutation.data;
  return {
    runBusy: mutation.isPending,
    runMessage: result?.message || (result?.status === "dry_run" ? "Todo run validated" : result ? "Todo run started" : ""),
    runError: mutation.error instanceof Error ? mutation.error.message : "",
    reset: mutation.reset,
    run,
  };
}

export function useTodoRunPreview(dir: string) {
  return useMutation({
    mutationKey: ["todos", "run", "preview", { dir: dir.trim() }],
    mutationFn: ({ body, signal }: { body: TodoRunOptions & { ref: string }; signal?: AbortSignal }) => todoMutationJSON<TodoRunPreviewResponse>(
      `/api/todos/run/preview?${todoQuery(dir)}`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
        signal,
      },
      "Failed to preview todo run",
    ),
  });
}

export function todoRunOptionsForRuntimeChange({
  action,
  context,
  options,
  runtime,
}: {
  action: TodoRunAction;
  context: RunContext;
  options: TodoRunOptions;
  runtime: RuntimeBarValue;
}): TodoRunOptions {
  return reconcileTodoRunOptions(action, {
    ...options,
    spec: { ...runSpec(options), ...runtime },
  }, context);
}

export function runChoiceDetail(options: TodoRunOptions, fallback: string, context?: RunContext | null): string {
  if (!context) return fallback;
  const runtime = modeForOptions(context, options);
  const spec = runSpec(options);
  const mode = runtimeModeLabel(runtimeModeForCatalog(runtime));
  const modelID = spec.model || runtime.defaultModel;
  const model = shortTodoRunModelName(labelForRunModel(runtime, modelID));
  const effort = spec.effort ? ` · ${spec.effort}` : "";
  return `${mode} · ${model}${effort}`;
}

export function buildTodoRunPayload({
  ref,
  driver,
  runMode,
  runtime,
  mode,
  resume,
  promptDraft,
  promptDirty,
}: {
  ref: string;
  driver: TodoRunDriver;
  runMode?: string;
  runtime: AISpecRuntimeValue;
  mode: TodoRunAction;
  resume: boolean;
  promptDraft: string;
  promptDirty: boolean;
}): TodoRunOptions & { ref: string } {
  const { spec } = promptRuntimeValueToPayload(runtime);
  const prompt = promptDirty ? { ...spec.prompt, user: promptDraft } : spec.prompt;
  // normalizeRunOptions owns which of runMode/plan/prompt a given action sends,
  // so the dialog cannot drift from what the phase buttons send.
  return {
    ref,
    ...normalizeRunOptions(mode, {
      driver,
      resume: resume || undefined,
      spec: {
        ...spec,
        mode: runMode,
        prompt,
      },
    }),
  };
}
