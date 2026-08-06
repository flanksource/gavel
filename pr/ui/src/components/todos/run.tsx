import { lazy, Suspense, useCallback, useEffect, useRef, useState, type ComponentType } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button, Field, Modal, SegmentedControl } from "@flanksource/clicky-ui/components";
import type { ChatModel } from "@flanksource/clicky-ui/chat";
import { effortOptionsForModel, PromptRunEditor, promptRuntimeValueToPayload, reconcileModelCapabilities, RuntimeBar, type AIPromptRunValue, type AISpecRuntimeValue, type RuntimeBarValue } from "@flanksource/clicky-ui/ai";
import { UiCog, UiListDashes, UiPlay, type IconProps } from "@flanksource/clicky-ui/icons";
import type { TodoRunEffort, TodoRunOptions, TodoRunPreviewResponse, TodoRunResponse } from "../../types";
import { Spinner } from "../../icons/Spinner";
import { inputClass, todoQuery } from "./format";
import { settingsRunContextQuery } from "../settings/queries";
import { invalidateTodoCaches, todoMutationJSON } from "./todoMutations";
export { TodoRunEffortBadge, todoRunEffortPresentation } from "./TodoRunEffortBadge";
import {
  PROVIDERS,
  agentForBackend,
  buildRunFamilies,
  defaultModelForSelection,
  driverForSelection,
  isCmuxBackend,
  modelsForSelection,
  type RunBackendCatalog,
  type RunContext,
} from "./providers";

// RunMode is the public agent prompt the dialog runs: Run (implement) or Plan
// (propose only). Verification is a fixture-backed issue lifecycle action in
// the Verification tab, not an agent run mode.
export type RunMode = "run" | "plan";
export type TodoRunAction = "run" | "plan";
export type TodoRunRuntimeMode = "cmux" | "agent" | "cli" | "api";

// The spec editor exposes exactly what gavel's dispatch reads: model/effort/budget,
// the prompt override, tool/permission posture, plus the run's Workspace (dirty
// worktree), Verify (definition-of-done checks), and Commit (auto-commit / dry-run). The last three
// replace the old loose checkboxes now that those options live on the api.Spec.
const RUN_SPEC_SECTIONS = ["model", "prompt", "permissions", "workspace", "verify", "commit"] as const;

// Runs auto-commit by default — the old commit=true default, now expressed as a
// single commit policy on the spec's Workflow.Commits. `on: "run"` keeps the
// dashboard's existing shape (one commit once the run finishes) rather than the
// per-turn fixup chain, which stays opt-in while a todo run executes in the
// user's live working tree. The advanced dialog's Commit section can turn it off;
// a plan-only run never commits because the plan action omits this workflow.
const AUTO_COMMIT: Pick<AISpecRuntimeValue, "workflow"> = { workflow: { commits: [{ on: "run", gates: "full" }] } };

// MdxEditorField is the same markdown editor field JsonSchemaForm uses for its
// markdown fields. It lazily pulls in the heavy @mdxeditor/editor, so it is
// code-split and rendered under Suspense with a plain-textarea fallback.
const MdxEditorField = lazy(() => import("@flanksource/clicky-ui/mdx-editor").then((m) => ({ default: m.MdxEditorField })));

export const defaultRunOptions: TodoRunOptions = { driver: "cli", spec: { effort: "medium", ...AUTO_COMMIT } };

type RunChoiceState = {
  last: Partial<Record<TodoRunAction, TodoRunOptions>>;
  recentAdvanced: Partial<Record<TodoRunAction, TodoRunOptions[]>>;
};

// v2: TodoRunOptions moved model/backend/effort/budget under a nested `spec`.
// A v1 entry would still parse, but every spec field would read as unset and the
// remembered model would silently revert to the backend default, so the version
// is bumped to discard it instead.
const RUN_CHOICE_STORAGE_KEY = "gavel.pr-ui.todoRunChoices.v2";

const RUN_ACTION_CONFIG: Record<TodoRunAction, { label: string; detail: string; icon: ComponentType<IconProps>; title: string }> = {
  run: { label: "Run", detail: "implement", icon: UiPlay, title: "Run todo" },
  plan: { label: "Plan", detail: "plan only", icon: UiListDashes, title: "Plan todo" },
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
// care about model/backend/effort read it through here.
export function runSpec(options: TodoRunOptions): AISpecRuntimeValue {
  return options.spec ?? {};
}

function normalizeRunOptions(action: TodoRunAction, options: TodoRunOptions): TodoRunOptions {
  const next = cloneRunOptions(options);
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
    for (const action of ["run", "plan"] as const) {
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

function runOptionsKey(options: TodoRunOptions): string {
  return JSON.stringify(sortForKey(options));
}

function actionFromRunOptions(options: TodoRunOptions): TodoRunAction {
  return options.plan || options.runMode === "plan" ? "plan" : "run";
}

export interface TodoRunContextState {
  context: RunContext | null;
  loading: boolean;
  error: string;
}

function unavailableRunContextError(context: RunContext): string {
  if (context.backends.some(backend => backend.models.length > 0)) return "";
  const details = context.backends.map(backend => backend.modelError?.trim()).filter(Boolean);
  return details[0] || "Captain returned no run models";
}

export function useTodoRunContext(enabled = true): TodoRunContextState {
  const query = useQuery({ ...settingsRunContextQuery(), enabled });
  if (!enabled) return { context: null, loading: false, error: "" };
  if (query.error) {
    return { context: null, loading: query.isFetching, error: query.error instanceof Error ? query.error.message : "Failed to load run context" };
  }
  const context = query.data ?? null;
  if (context && (!Array.isArray(context.backends) || !Array.isArray(context.efforts) || !Array.isArray(context.tools))) {
    return { context: null, loading: false, error: "Captain returned an invalid run context" };
  }
  return { context, loading: query.isFetching, error: context ? unavailableRunContextError(context) : "" };
}

export function TodoRunContextError({ error }: { error: string }) {
  if (!error) return null;
  return <div role="alert" className="max-w-sm text-xs text-red-600">{error}</div>;
}

function primaryBackendForAction(context: RunContext): RunBackendCatalog {
  return context.backends.find(backend => backend.id === context.defaultBackend && backend.models.length > 0)
    ?? context.backends.find(backend => backend.models.length > 0)
    ?? (() => { throw new Error("Captain returned no run models"); })();
}

function backendForOptions(context: RunContext, options: TodoRunOptions): RunBackendCatalog {
  const requested = runSpec(options).backend || options.driver || options.mode || "";
  return (
    context.backends.find(backend => backend.models.length > 0 && (backend.id === requested || backend.driver === requested)) ??
    (context.defaultBackend ? context.backends.find(backend => backend.models.length > 0 && backend.id === context.defaultBackend) : undefined) ??
    context.backends.find(backend => backend.models.length > 0) ??
    (() => { throw new Error("Captain returned no run models"); })()
  );
}

function mechanismForBackend(backend: RunBackendCatalog): string {
  const value = backend.mechanisms[0]?.value;
  if (value) return value;
  const parts = backend.driver.split("-");
  return parts.length > 1 ? parts.slice(1).join("-") : backend.driver;
}

function runtimeModeForBackend(backend: RunBackendCatalog): TodoRunRuntimeMode {
  const id = backend.id.toLowerCase();
  if (id.endsWith("-cmux")) return "cmux";
  if (id.endsWith("-agent")) return "agent";
  if (id.endsWith("-cli")) return "cli";
  if (backend.type === "api" || ["anthropic", "openai", "gemini", "deepseek"].includes(id)) return "api";

  switch (mechanismForBackend(backend).toLowerCase()) {
    case "cmux":
      return "cmux";
    case "api":
      return "api";
    case "cli":
      return "cli";
    default:
      return "agent";
  }
}

function runtimeModeLabel(mode: TodoRunRuntimeMode): string {
  return RUNTIME_MODE_CONFIG[mode].label;
}

function modelsForRunBackend(backend: RunBackendCatalog): RunBackendCatalog["models"] {
  return backend.models;
}

const ALL_EFFORTS: TodoRunEffort[] = ["low", "medium", "high", "xhigh", "max", "ultra"];

function modelForRunBackend(backend: RunBackendCatalog, modelID: string | undefined): ChatModel {
  const model = modelsForRunBackend(backend).find(item => item.id === modelID)
    ?? modelsForRunBackend(backend).find(item => item.id === backend.defaultModel)
    ?? modelsForRunBackend(backend)[0];
  if (!model) throw new Error(backend.modelError || `Captain returned no models for ${backend.label}`);
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

function labelForRunModel(backend: RunBackendCatalog, modelID: string): string {
  const model = modelsForRunBackend(backend).find(item => item.id === modelID);
  return model?.label || modelID;
}

function runOptionsForBackendModel(action: TodoRunAction, backend: RunBackendCatalog, modelID: string, effort: TodoRunEffort = "medium"): TodoRunOptions {
  const spec = reconcileModelCapabilities({
    backend: backend.id,
    model: modelID || backend.defaultModel,
    effort,
    ...(action === "run" ? AUTO_COMMIT : {}),
  } satisfies AISpecRuntimeValue, modelForRunBackend(backend, modelID), ALL_EFFORTS);
  return normalizeRunOptions(action, { driver: backend.driver, spec });
}

export function runButtonQualifierForOptions(options: TodoRunOptions, context: RunContext): string {
  const backend = backendForOptions(context, options);
  const model = runSpec(options).model || backend.defaultModel;
  return `(${runtimeModeLabel(runtimeModeForBackend(backend))}:${shortTodoRunModelName(labelForRunModel(backend, model))})`;
}

// todoRunModeLabel is the runtime mechanism a run would use (Agent/cmux/cli/API),
// resolved from the run options against the backend catalog — the same derivation
// the run buttons use, exposed for the start-of-session hero's "Runtime" chip.
export function todoRunModeLabel(options: TodoRunOptions, context: RunContext): string {
  return runtimeModeLabel(runtimeModeForBackend(backendForOptions(context, options)));
}

export function runButtonLabelForOptions(action: TodoRunAction, options: TodoRunOptions, context: RunContext): string {
  return `${RUN_ACTION_CONFIG[action].label} ${runButtonQualifierForOptions(options, context)}`;
}

export function todoRunButtonPresentation(options: TodoRunOptions, context: RunContext) {
  const backend = backendForOptions(context, options);
  const spec = runSpec(options);
  const modelID = spec.model || backend.defaultModel;
  const model = modelForRunBackend(backend, modelID);
  const provider = PROVIDERS.find(item => item.id === backend.agent);
  const supportedEfforts = effortOptionsForModel(model, contextEfforts(context));
  const effort = spec.effort && supportedEfforts.includes(spec.effort)
    ? spec.effort as TodoRunEffort
    : undefined;

  return {
    provider,
    model: shortTodoRunModelName(labelForRunModel(backend, modelID)),
    effort,
  };
}

export function defaultRunOptionsForAction(action: TodoRunAction, context?: RunContext | null): TodoRunOptions {
  if (context) {
    const backend = primaryBackendForAction(context);
    return runOptionsForBackendModel(action, backend, backend.defaultModel);
  }
  return normalizeRunOptions(action, defaultRunOptions);
}

export function reconcileTodoRunOptions(action: TodoRunAction, options: TodoRunOptions, context: RunContext): TodoRunOptions {
  const normalized = normalizeRunOptions(action, options);
  const backend = backendForOptions(context, normalized);
  const spec = runSpec(normalized);
  const modelIsCurrent = !!spec.model && backend.models.some(model => model.id === spec.model);
  const model = modelIsCurrent ? spec.model! : backend.defaultModel;
  return normalizeRunOptions(action, {
    ...normalized,
    driver: backend.driver,
    spec: reconcileModelCapabilities({ ...spec, backend: backend.id, model }, modelForRunBackend(backend, model), contextEfforts(context)),
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
      } catch {
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
    mutationFn: ({ body, signal }: { body: Record<string, unknown>; signal?: AbortSignal }) => todoMutationJSON<TodoRunPreviewResponse>(
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

export function TodoRunRuntimeBar({
  action,
  context,
  options,
  disabled,
  onChange,
}: {
  action: TodoRunAction;
  context: RunContext;
  options: TodoRunOptions;
  disabled?: boolean;
  onChange: (options: TodoRunOptions) => void;
}) {
  const spec = runSpec(options);
  const agent = agentForBackend(context, spec.backend);
  return (
    <fieldset disabled={disabled} className="min-w-0 border-0 p-0 disabled:opacity-50">
      <RuntimeBar<AISpecRuntimeValue>
        value={spec}
        variant="combo"
        families={buildRunFamilies(context)}
        models={modelsForSelection(context, agent, spec.backend)}
        reasoningEfforts={context.efforts}
        ariaLabel={`${RUN_ACTION_CONFIG[action].label} runtime`}
        className="min-w-0"
        onChange={runtime => onChange(todoRunOptionsForRuntimeChange({ action, context, options, runtime }))}
      />
    </fieldset>
  );
}

export function TodoRunActionButton({
  action,
  disabled,
  loading,
  label,
  icon,
  tone = "default",
  title,
  options: controlledOptions,
  onOptionsChange,
  onRun,
  onAdvanced,
}: {
  action: TodoRunAction;
  disabled?: boolean;
  loading?: boolean;
  label?: string;
  icon?: ComponentType<IconProps>;
  tone?: "default" | "danger";
  title?: string;
  options?: TodoRunOptions;
  onOptionsChange?: (options: TodoRunOptions) => void;
  onRun: (options?: TodoRunOptions) => void;
  onAdvanced: (action: TodoRunAction) => void;
}) {
  const config = RUN_ACTION_CONFIG[action];
  const { context, loading: contextLoading, error: contextError } = useTodoRunContext();
  const [selectedOptions, setSelectedOptions] = useState<TodoRunOptions | null>(null);
  useEffect(() => {
    setSelectedOptions(null);
  }, [action, context]);
  const unavailable = contextLoading || !context || !!contextError;
  const candidateOptions = controlledOptions ?? selectedOptions ?? loadLastTodoRunOptions(action, context);
  const currentOptions = context ? reconcileTodoRunOptions(action, candidateOptions, context) : candidateOptions;
  const primaryTone = tone === "danger" ? "text-red-600 hover:bg-red-500/10 hover:text-red-700" : "text-foreground hover:bg-muted";
  const PrimaryIcon = loading ? Spinner : icon ?? config.icon;

  function selectOptions(options: TodoRunOptions) {
    const remembered = rememberTodoRunOptions(action, options);
    setSelectedOptions(remembered);
    onOptionsChange?.(remembered);
  }

  function runWith(options: TodoRunOptions) {
    const remembered = rememberTodoRunOptions(action, options);
    onRun(remembered);
  }

  return (
    <div className="flex flex-col items-start gap-1">
      <div className="inline-flex min-h-8 shrink-0 items-stretch gap-1">
        <Button
          variant="ghost"
          type="button"
          onClick={() => runWith(currentOptions)}
          disabled={disabled || unavailable}
          title={title ?? config.title}
          className={`inline-flex h-8 items-center gap-1 rounded-md border border-border px-2 text-xs font-medium disabled:opacity-50 ${primaryTone}`}
        >
          <PrimaryIcon className="text-xs" />
          <span>{label ?? config.label}</span>
        </Button>
        {context && !contextError ? (
          <TodoRunRuntimeBar
            action={action}
            context={context}
            options={currentOptions}
            disabled={disabled || unavailable}
            onChange={selectOptions}
          />
        ) : (
          <Button variant="outline" size="sm" type="button" disabled title={`${config.label} runtime unavailable`} aria-label={`${config.label} runtime unavailable`} className="h-8 px-2 text-xs">
            Runtime
          </Button>
        )}
        <Button
          variant="outline"
          size="icon"
          type="button"
          disabled={disabled || unavailable}
          title={`Advanced ${config.label.toLowerCase()} options`}
          aria-label={`Advanced ${config.label.toLowerCase()} options`}
          onClick={() => onAdvanced(action)}
          className="h-8 w-8 shrink-0 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
        >
          <UiCog className="text-sm" />
        </Button>
      </div>
      <TodoRunContextError error={contextError} />
    </div>
  );
}

function runChoiceDetail(options: TodoRunOptions, fallback: string, context?: RunContext | null): string {
  if (!context) return fallback;
  const backend = backendForOptions(context, options);
  const spec = runSpec(options);
  const mode = runtimeModeLabel(runtimeModeForBackend(backend));
  const modelID = spec.model || backend.defaultModel;
  const model = shortTodoRunModelName(labelForRunModel(backend, modelID));
  const effort = spec.effort ? ` · ${spec.effort}` : "";
  return `${mode} · ${model}${effort}`;
}

const INITIAL_RUNTIME_VALUE: AISpecRuntimeValue = { effort: "medium", ...AUTO_COMMIT };

export function TodoRunAdvancedDialog({
  open,
  onClose,
  onRun,
  loading,
  title = "Run todo",
  initialMode = "run",
  dir,
  refID,
}: {
  open: boolean;
  onClose: () => void;
  onRun: (options: TodoRunOptions) => void;
  loading?: boolean;
  title?: string;
  initialMode?: RunMode;
  // dir/refID identify the TODO this dialog will run, so it can fetch
  // a live preview of the prompt that will be sent as the options change.
  dir: string;
  refID: string;
}) {
  const [runRequest, setRunRequest] = useState<AIPromptRunValue>({
    spec: INITIAL_RUNTIME_VALUE,
  });
  const runtimeValue = runRequest.spec ?? {};
  const [mode, setMode] = useState<RunMode>("run");
  // Resume stays a discrete toggle (session-identity decision, cmux only); dirty,
  // auto-commit, dry-run, and checks now live on runtimeValue's spec (Workspace/
  // Commit/Verify sections), not as parallel booleans.
  const [resume, setResume] = useState(false);
  // promptDraft is the editable prompt body sent as the verbatim override;
  // promptDirty stops the live preview from clobbering the user's edits.
  const [promptDraft, setPromptDraft] = useState("");
  const [promptDirty, setPromptDirty] = useState(false);
  const [previewError, setPreviewError] = useState("");
  const [regenNonce, setRegenNonce] = useState(0);
  const { context, loading: contextLoading, error: contextError } = useTodoRunContext(open);
  const previewMutation = useTodoRunPreview(dir);
  const previewLoading = previewMutation.isPending;
  // Ref mirror of promptDirty so the preview effect can read it without
  // refetching on every keystroke.
  const promptDirtyRef = useRef(false);

  const families = context ? buildRunFamilies(context) : [];
  const agent = context ? agentForBackend(context, runtimeValue.backend) : undefined;
  const isCmux = !!agent && isCmuxBackend(agent, runtimeValue.backend);
  const activeModels = context && agent ? modelsForSelection(context, agent, runtimeValue.backend) : [];
  const modelFallback = context && agent ? defaultModelForSelection(context, agent, runtimeValue.backend) : "";
  const selection = context && agent ? driverForSelection(context, agent, runtimeValue.backend) : null;
  const driver = selection?.driver;
  const runBackend = selection?.runBackend;
  const plan = mode === "plan";
  const advancedAction: TodoRunAction = plan ? "plan" : "run";
  const recentAdvanced = context
    ? loadRecentAdvancedTodoRunOptions(advancedAction, context)
    : [];

  function changeMode(next: RunMode) {
    setMode(next);
    setPromptDirty(false);
    promptDirtyRef.current = false;
  }

  function editPrompt(v: string) {
    setPromptDraft(v);
    setPromptDirty(true);
    promptDirtyRef.current = true;
  }

  function regeneratePrompt() {
    setPromptDirty(false);
    promptDirtyRef.current = false;
    setRegenNonce((n) => n + 1);
  }

  useEffect(() => {
    if (!open) return;
    setRunRequest({ spec: INITIAL_RUNTIME_VALUE });
    setMode(initialMode);
    setResume(false);
    setPromptDraft("");
    setPromptDirty(false);
    promptDirtyRef.current = false;
  }, [open, initialMode]);

  useEffect(() => {
    if (!open || !context) return;
    const action: TodoRunAction = initialMode === "plan" ? "plan" : "run";
    setRunRequest({
      spec: runSpec(
        reconcileTodoRunOptions(
          action,
          loadLastTodoRunOptions(action, context),
          context,
        ),
      ),
    });
  }, [open, initialMode, context]);

  const previewModel = runtimeValue.model?.trim() || modelFallback;
  const previewBackend = runBackend;

  // Fetch the prompt that will be sent whenever the dialog is open and a
  // prompt-affecting option changes (driver/model/effort/plan/resume). The
  // server builds it from the same code path the run uses, so it matches exactly.
  // Fetch the generated Run/Plan prompt body and seed the editor unless the
  // user has edited it. The server uses the same rendering path as dispatch.
  useEffect(() => {
    if (!open) {
      setPreviewError("");
      return;
    }
    if (!context || contextError || !refID || !driver) {
      setPreviewError("");
      return;
    }
    // Same payload shape as the run POST — the preview handler decodes it with
    // the same todoRunPayload — so the spec half is nested, not inlined.
    const body = {
      ref: refID,
      driver,
      runMode: plan ? "plan" : "run",
      resume: isCmux ? resume : undefined,
      spec: { backend: previewBackend, model: previewModel, effort: runtimeValue.effort },
    };

    let cancelled = false;
    const controller = new AbortController();
    setPreviewError("");
    previewMutation.mutate({ body, signal: controller.signal }, {
      onSuccess: data => {
        if (!cancelled && !promptDirtyRef.current) setPromptDraft(data.prompt ?? "");
      },
      onError: err => {
        if (!cancelled && !(err instanceof DOMException && err.name === "AbortError")) setPreviewError(err.message);
      },
    });
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [open, context, contextError, dir, refID, driver, previewBackend, previewModel, runtimeValue.effort, plan, resume, isCmux, regenNonce, previewMutation.mutate]);

  if (!open) return null;

  function submit() {
    if (!context || !driver) return;
    // spec carries the model/effort/permissions plus the run's setup (dirty),
    // workflow.verify (checks) and workflow.commits (auto-commit / dry-run) — all
    // edited via the spec editor's Workspace/Verify/Commit sections. Plan-only runs
    // never commit or verify; the server suppresses both for run mode plan.
    const { spec } = promptRuntimeValueToPayload(runtimeValue);
    onRun({
      driver,
      runMode: plan ? "plan" : "run",
      plan: plan ? true : undefined,
      resume: isCmux ? resume : undefined,
      spec: {
        ...spec,
        backend: runBackend,
        prompt: {
          // The edited prompt body is sent verbatim as the override.
          user: promptDraft.trim() ? promptDraft : undefined,
        },
      },
    });
  }

  const modeOptions: { id: RunMode; label: string }[] = [{ id: "run", label: "Run" }];
  modeOptions.push({ id: "plan", label: "Plan" });

  // Shared regenerate/error/editor/dirty-note block passed to PromptRunEditor.
  const promptEditorNode = (
    <div className="space-y-1">
      <div className="flex items-center justify-end gap-2">
        {previewLoading && <Spinner className="text-xs text-muted-foreground" />}
        <Button
          variant="ghost"
          type="button"
          onClick={regeneratePrompt}
          disabled={previewLoading}
          title="Discard edits and regenerate from the options above"
          className="h-auto rounded px-1.5 py-0.5 text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          Regenerate
        </Button>
      </div>
      {previewError && <div className="text-xs text-red-600">{previewError}</div>}
      <Suspense fallback={<textarea className={`${inputClass} h-auto min-h-[16rem] resize-y font-mono`} value={promptDraft} onChange={(e) => editPrompt(e.currentTarget.value)} placeholder={previewLoading ? "Loading prompt…" : "Prompt"} />}>
        <MdxEditorField value={promptDraft} onChange={editPrompt} placeholder={previewLoading ? "Loading prompt…" : "Prompt"} className="min-h-[16rem]" />
      </Suspense>
      {promptDirty && <div className="text-[11px] text-muted-foreground">Edited — sent verbatim as the prompt.</div>}
    </div>
  );

  return (
    <Modal
      open
      onClose={onClose}
      title={title}
      size="2xl"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={submit} loading={loading} disabled={contextLoading || !context || !!contextError}>
            {plan ? "Plan" : "Run"}
          </Button>
        </div>
      }
    >
      <div className="space-y-3">
        <TodoRunContextError error={contextError} />
        {contextLoading && <div className="text-xs text-muted-foreground">Loading Captain run providers…</div>}
        {context && !contextError && (
          <>
            <Field label="Mode">
              <SegmentedControl aria-label="Mode" value={mode} onChange={(v) => changeMode(v as RunMode)} options={modeOptions} />
            </Field>
            <PromptRunEditor
              value={runRequest}
              onChange={setRunRequest}
              models={activeModels}
              families={families}
              tools={context.tools}
              specSections={RUN_SPEC_SECTIONS}
              promptEditor={promptEditorNode}
              promptLabel="Prompt"
            >
              <>
                {isCmux && (
                  <label className="inline-flex items-center gap-2 text-xs">
                    <input type="checkbox" checked={resume} onChange={(e) => setResume(e.currentTarget.checked)} />
                    <span>Resume session</span>
                  </label>
                )}
                {recentAdvanced.length > 0 && (
                  <div className="space-y-1 border-t border-border pt-3">
                    <div className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">Recent advanced</div>
                    <div className="flex flex-wrap gap-1.5">
                      {recentAdvanced.map((options, index) => (
                        <Button
                          key={runOptionsKey(options)}
                          variant="outline"
                          size="sm"
                          type="button"
                          onClick={() =>
                            setRunRequest((current) => ({
                              ...current,
                              spec: runSpec(reconcileTodoRunOptions(advancedAction, options, context)),
                            }))
                          }
                        >
                          {index + 1}. {runChoiceDetail(options, "advanced", context)}
                        </Button>
                      ))}
                    </div>
                  </div>
                )}
              </>
            </PromptRunEditor>
          </>
        )}
      </div>
    </Modal>
  );
}
