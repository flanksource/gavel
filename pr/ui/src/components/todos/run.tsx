import { lazy, Suspense, useCallback, useEffect, useRef, useState, type ComponentType } from "react";
import { Button, Combobox, DropdownMenu, Field, Modal, SegmentedControl } from "@flanksource/clicky-ui/components";
import { ModelSelector, ProviderSelector } from "@flanksource/clicky-ui/chat";
import { PromptRunEditor, promptRuntimeValueToPayload, type AISpecRuntimeValue } from "@flanksource/clicky-ui/ai";
import type { StaticIconComponent } from "@flanksource/clicky-ui/data";
import { UiChevronDown, UiChevronRight, UiCloud, UiCog, UiColumns, UiListDashes, UiPlay, UiRobotAi, UiSparkles, UiTerminal, type IconProps } from "@flanksource/clicky-ui/icons";
import type { TodoRunAgent, TodoRunEffort, TodoRunOptions, TodoRunPreviewResponse, TodoRunResponse } from "../../types";
import { Spinner } from "../../icons/Spinner";
import { inputClass, todoQuery } from "./format";
import {
  PROVIDERS,
  agentForBackend,
  backendCatalog as findBackendCatalog,
  backendsForAgent,
  buildRunFamilies,
  defaultBackendForAgent,
  defaultModelForSelection,
  driverForSelection,
  isCmuxBackend,
  modelsForSelection,
  runContextWithFallback,
  type RunBackendCatalog,
  type RunContext,
} from "./providers";

// RunMode is the prompt the dialog runs: Run (implement), Plan (propose only),
// or Verify (score the committed work against acceptance criteria).
export type RunMode = "run" | "plan" | "verify";
export type TodoRunAction = "run" | "plan";
export type TodoRunRuntimeMode = "cmux" | "agent" | "cli" | "api";

// The spec editor exposes exactly what gavel's dispatch reads: model/effort/budget,
// the prompt override, tool/permission posture, plus the run's Workspace (dirty
// worktree), Verify (checks), and Commit (auto-commit / dry-run). The last three
// replace the old loose checkboxes now that those options live on the api.Spec.
const RUN_SPEC_SECTIONS = ["model", "prompt", "permissions", "workspace", "verify", "commit"] as const;

// Runs auto-commit by default — the old commit=true default, now expressed as the
// spec's Workflow.PostRun.commit. The advanced dialog's Commit section can turn it
// off; a plan-only run never commits (the server suppresses commit in plan mode).
const AUTO_COMMIT: Pick<TodoRunOptions, "workflow"> = { workflow: { postRun: { commit: true } } };

// MdxEditorField is the same markdown editor field JsonSchemaForm uses for its
// markdown fields. It lazily pulls in the heavy @mdxeditor/editor, so it is
// code-split and rendered under Suspense with a plain-textarea fallback.
const MdxEditorField = lazy(() => import("@flanksource/clicky-ui/mdx-editor").then((m) => ({ default: m.MdxEditorField })));

export const defaultRunOptions: TodoRunOptions = { driver: "claude-cmux", model: "claude-sonnet-5", effort: "medium", ...AUTO_COMMIT };

type RunPreset = { label: string; icon: ComponentType<IconProps>; options: TodoRunOptions };
type RunChoiceState = {
  last: Partial<Record<TodoRunAction, TodoRunOptions>>;
  recentAdvanced: Partial<Record<TodoRunAction, TodoRunOptions[]>>;
};

// The split-button menu offers two actions — Run (implement) and Plan (propose
// a plan without changing code) — each with a Claude and a Codex option, plus
// Advanced for the full dialog.
export const runActionGroups: Array<{ action: "Run" | "Plan"; detail: string; presets: RunPreset[] }> = [
  {
    action: "Run",
    detail: "implement",
    presets: [
      { label: "Claude", icon: UiSparkles, options: { driver: "claude-cmux", model: "claude-sonnet-5", effort: "medium", ...AUTO_COMMIT } },
      { label: "Codex", icon: UiTerminal, options: { driver: "codex-cmux", backend: "codex-cmux", model: "gpt-5.5", effort: "medium", ...AUTO_COMMIT } },
    ],
  },
  {
    action: "Plan",
    detail: "plan only · no changes",
    presets: [
      { label: "Claude", icon: UiSparkles, options: { driver: "claude-cmux", backend: "claude-cmux", model: "claude-sonnet-5", effort: "medium", runMode: "plan", plan: true } },
      { label: "Codex", icon: UiTerminal, options: { driver: "codex-cmux", backend: "codex-cmux", model: "gpt-5.5", effort: "medium", runMode: "plan", plan: true } },
    ],
  },
];

const RUN_CHOICE_STORAGE_KEY = "gavel.pr-ui.todoRunChoices.v1";

const RUN_ACTION_CONFIG: Record<TodoRunAction, { label: string; detail: string; icon: ComponentType<IconProps>; title: string }> = {
  run: { label: "Run", detail: "implement", icon: UiPlay, title: "Run todo" },
  plan: { label: "Plan", detail: "plan only", icon: UiListDashes, title: "Plan todo" },
};

const RUNTIME_MODE_ORDER: TodoRunRuntimeMode[] = ["cmux", "agent", "cli", "api"];
const RUNTIME_MODE_CONFIG: Record<TodoRunRuntimeMode, { label: string; icon: StaticIconComponent }> = {
  cmux: { label: "cmux", icon: UiColumns },
  agent: { label: "agent", icon: UiRobotAi },
  cli: { label: "cli", icon: UiTerminal },
  api: { label: "API", icon: UiCloud },
};

export type TodoRunModelChoice = {
  key: string;
  action: TodoRunAction;
  backend: RunBackendCatalog;
  modelID: string;
  modelLabel: string;
  modelShort: string;
  mechanism: string;
  icon: ComponentType<IconProps>;
  options: TodoRunOptions;
};

const RUN_PRESETS: Record<TodoRunAction, RunPreset[]> = {
  run: runActionGroups[0]?.presets ?? [],
  plan: runActionGroups[1]?.presets ?? [],
};

function emptyRunChoiceState(): RunChoiceState {
  return { last: {}, recentAdvanced: {} };
}

function cloneRunOptions(options: TodoRunOptions): TodoRunOptions {
  return JSON.parse(JSON.stringify(options)) as TodoRunOptions;
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

function useTodoRunContext(enabled = true): RunContext {
  const [runContext, setRunContext] = useState<RunContext | null>(null);

  useEffect(() => {
    if (!enabled) return;
    let cancelled = false;
    fetch("/api/todos/run/context")
      .then(async (res) => {
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || "Failed to load run context");
        if (!cancelled) setRunContext(data as RunContext);
      })
      .catch(() => {
        if (!cancelled) setRunContext(null);
      });
    return () => {
      cancelled = true;
    };
  }, [enabled]);

  return runContextWithFallback(runContext);
}

function primaryBackendForAction(context: RunContext): RunBackendCatalog {
  return context.backends.find(backend => backend.id === context.defaultBackend) ?? context.backends[0];
}

function backendForOptions(context: RunContext, options: TodoRunOptions): RunBackendCatalog {
  const requested = options.backend || options.driver || options.mode || "";
  return (
    context.backends.find(backend => backend.id === requested || backend.driver === requested) ??
    (context.defaultBackend ? context.backends.find(backend => backend.id === context.defaultBackend) : undefined) ??
    context.backends[0]
  );
}

function mechanismForBackend(backend: RunBackendCatalog): string {
  const value = backend.mechanisms[0]?.value;
  if (value) return value;
  const parts = backend.driver.split("-");
  return parts.length > 1 ? parts.slice(1).join("-") : backend.driver;
}

function iconForRunBackend(backend: RunBackendCatalog): ComponentType<IconProps> {
  return backend.agent === "claude" ? UiSparkles : UiTerminal;
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

function runtimeModeForOptions(context: RunContext, options: TodoRunOptions): TodoRunRuntimeMode {
  return runtimeModeForBackend(backendForOptions(context, options));
}

function runtimeModeLabel(mode: TodoRunRuntimeMode): string {
  return RUNTIME_MODE_CONFIG[mode].label;
}

function hasRuntimeMode(context: RunContext, mode: TodoRunRuntimeMode): boolean {
  return context.backends.some((backend) => runtimeModeForBackend(backend) === mode);
}

function firstAvailableRuntimeMode(context: RunContext, preferred: TodoRunRuntimeMode): TodoRunRuntimeMode {
  if (hasRuntimeMode(context, preferred)) return preferred;
  return RUNTIME_MODE_ORDER.find((mode) => hasRuntimeMode(context, mode)) ?? "cmux";
}

function fallbackModelForBackend(backend: RunBackendCatalog): RunBackendCatalog["models"][number] {
  return {
    id: backend.defaultModel,
    provider: backend.provider,
    label: backend.defaultModel,
    reasoning: true,
    configured: backend.configured ?? false,
  };
}

function modelsForRunBackend(backend: RunBackendCatalog): RunBackendCatalog["models"] {
  return backend.models.length > 0 ? backend.models : [fallbackModelForBackend(backend)];
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
  const options: TodoRunOptions = {
    driver: backend.driver,
    backend: backend.id,
    model: modelID || backend.defaultModel,
    effort,
    ...(action === "run" ? AUTO_COMMIT : {}),
  };
  return normalizeRunOptions(action, options);
}

export function runChoicesForAction(context: RunContext, action: TodoRunAction, effort: TodoRunEffort = "medium"): TodoRunModelChoice[] {
  const config = RUN_ACTION_CONFIG[action];
  return context.backends.flatMap((backend) => {
    const mechanism = mechanismForBackend(backend);
    const Icon = iconForRunBackend(backend);
    return modelsForRunBackend(backend).map((model) => {
      const modelID = model.id || backend.defaultModel;
      const modelShort = shortTodoRunModelName(modelID);
      return {
        key: `${action}:${backend.id}:${modelID}`,
        action,
        backend,
        modelID,
        modelLabel: model.label || modelShort,
        modelShort,
        mechanism,
        icon: Icon,
        options: runOptionsForBackendModel(action, backend, modelID, effort),
      } satisfies TodoRunModelChoice;
    });
  }).filter(choice => !!choice.modelID && !!config);
}

export function runChoicesForRuntimeMode(context: RunContext, action: TodoRunAction, runtimeMode: TodoRunRuntimeMode, effort: TodoRunEffort = "medium"): TodoRunModelChoice[] {
  return runChoicesForAction(context, action, effort).filter((choice) => runtimeModeForBackend(choice.backend) === runtimeMode);
}

export function runButtonLabelForOptions(action: TodoRunAction, options: TodoRunOptions, context: RunContext): string {
  const backend = backendForOptions(context, options);
  const model = options.model || backend.defaultModel;
  return `${RUN_ACTION_CONFIG[action].label} (${runtimeModeLabel(runtimeModeForBackend(backend))}:${shortTodoRunModelName(labelForRunModel(backend, model))})`;
}

export function defaultRunOptionsForAction(action: TodoRunAction, context?: RunContext | null): TodoRunOptions {
  const resolved = runContextWithFallback(context);
  const backend = primaryBackendForAction(resolved);
  if (backend) {
    return runOptionsForBackendModel(action, backend, backend.defaultModel);
  }
  return normalizeRunOptions(action, RUN_PRESETS[action][0]?.options ?? defaultRunOptions);
}

export function loadLastTodoRunOptions(action: TodoRunAction, context?: RunContext | null): TodoRunOptions {
  const state = readRunChoiceState();
  return normalizeRunOptions(action, state.last[action] ?? defaultRunOptionsForAction(action, context));
}

export function loadRecentAdvancedTodoRunOptions(action: TodoRunAction): TodoRunOptions[] {
  const state = readRunChoiceState();
  return (state.recentAdvanced[action] ?? []).map(item => normalizeRunOptions(action, item));
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

// useTodoRun POSTs a run for one or more todo refs in a workspace. A single ref
// runs on its own; multiple refs run together in one agent session (the server
// dispatches them as a combined group). Both the single-todo detail pane and the
// list's multi-select bar drive runs through this one hook.
export function useTodoRun(dir: string, provider: string) {
  const [runBusy, setRunBusy] = useState(false);
  const [runMessage, setRunMessage] = useState("");
  const [runError, setRunError] = useState("");

  const reset = useCallback(() => {
    setRunMessage("");
    setRunError("");
  }, []);

  const run = useCallback(
    async (refs: string[], options: TodoRunOptions = defaultRunOptions): Promise<TodoRunResponse | null> => {
      const cleaned = refs.map((r) => r.trim()).filter(Boolean);
      if (cleaned.length === 0 || runBusy) return null;
      setRunBusy(true);
      setRunError("");
      setRunMessage("");
      try {
        // Send `ref` for a single todo (matching the original payload) and `refs`
        // for a multi-select group run.
        const body = cleaned.length === 1 ? { ref: cleaned[0], ...options } : { refs: cleaned, ...options };
        const response = await fetch(`/api/todos/run?${todoQuery(dir, provider)}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body),
        });
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "Run failed");
        const result = data as TodoRunResponse;
        setRunMessage(result.message || (result.status === "dry_run" ? "Todo run validated" : "Todo run started"));
        return result;
      } catch (err: any) {
        setRunError(err?.message || "Run failed");
        return null;
      } finally {
        setRunBusy(false);
      }
    },
    [dir, provider, runBusy],
  );

  return { runBusy, runMessage, runError, reset, run };
}

type TodoRunDropdownSelect = (action: TodoRunAction, options: TodoRunOptions, advanced?: boolean) => void;

function TodoRunDropdownContent({
  context,
  initialAction,
  closeParent,
  onSelect,
  onAdvanced,
}: {
  context: RunContext;
  initialAction: TodoRunAction;
  closeParent: () => void;
  onSelect: TodoRunDropdownSelect;
  onAdvanced: (action: TodoRunAction) => void;
}) {
  const initialOptions = loadLastTodoRunOptions(initialAction, context);
  const [runtimeMode, setRuntimeMode] = useState<TodoRunRuntimeMode>(() => firstAvailableRuntimeMode(context, runtimeModeForOptions(context, initialOptions)));
  const [effort, setEffort] = useState<TodoRunEffort>((initialOptions.effort as TodoRunEffort | undefined) ?? "medium");
  const activeRuntimeMode = firstAvailableRuntimeMode(context, runtimeMode);
  const choices = runChoicesForRuntimeMode(context, initialAction, activeRuntimeMode, effort);
  const recentAdvanced = loadRecentAdvancedTodoRunOptions(initialAction);
  const runtimeModeOptions = RUNTIME_MODE_ORDER.map((item) => ({
    id: item,
    label: RUNTIME_MODE_CONFIG[item].label,
    icon: RUNTIME_MODE_CONFIG[item].icon,
    disabled: !hasRuntimeMode(context, item),
  }));
  const effortValues: TodoRunEffort[] = context.efforts.length > 0 ? context.efforts : ["medium"];
  const effortOptions = effortValues.map((item) => ({
    id: item,
    label: item,
  }));

  useEffect(() => {
    const nextOptions = loadLastTodoRunOptions(initialAction, context);
    setRuntimeMode(firstAvailableRuntimeMode(context, runtimeModeForOptions(context, nextOptions)));
    setEffort((nextOptions.effort as TodoRunEffort | undefined) ?? "medium");
  }, [initialAction, context.backends, context.defaultBackend]);

  return (
    <div className="p-1 text-xs">
      <div className="space-y-1.5 border-b border-border px-1 pb-2 pt-1">
        <SegmentedControl<TodoRunRuntimeMode>
          aria-label="Runtime mode"
          size="sm"
          value={activeRuntimeMode}
          onChange={setRuntimeMode}
          options={runtimeModeOptions}
          className="w-full"
        />
        <SegmentedControl<TodoRunEffort>
          aria-label="Effort"
          size="sm"
          value={effort}
          onChange={setEffort}
          options={effortOptions}
          wrap
          className="w-full"
        />
      </div>

      <div className="pt-1">
        {context.backends.map((backend) => {
          const backendChoices = choices.filter(choice => choice.backend.id === backend.id);
          if (backendChoices.length === 0) return null;
          return (
            <TodoRunProviderSubmenu
              key={backend.id}
              backend={backend}
              choices={backendChoices}
              onSelect={(choice) => onSelect(initialAction, choice.options)}
              onCloseParent={closeParent}
            />
          );
        })}
      </div>

      {recentAdvanced.length > 0 && (
        <>
          <div className="my-1 border-t border-border" />
          <div className="px-2 pb-0.5 pt-1.5 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">Recent advanced</div>
          {recentAdvanced.map((options, index) => (
            <Button
              key={`${initialAction}:advanced:${runOptionsKey(options)}`}
              variant="ghost"
              type="button"
              onClick={() => {
                closeParent();
                onSelect(initialAction, { ...options, effort }, true);
              }}
              className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
            >
              <UiCog className="shrink-0 text-sm text-muted-foreground" />
              <span className="min-w-0 flex-1">
                <span className="block truncate font-medium text-foreground">Advanced {index + 1}</span>
                <span className="block truncate text-[11px] text-muted-foreground">{runChoiceDetail({ ...options, effort }, RUN_ACTION_CONFIG[initialAction].detail, context)}</span>
              </span>
            </Button>
          ))}
        </>
      )}

      <div className="my-1 border-t border-border" />
      <Button
        variant="ghost"
        type="button"
        onClick={() => {
          closeParent();
          onAdvanced(initialAction);
        }}
        className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
      >
        <UiCog className="shrink-0 text-sm text-muted-foreground" />
        <span className="min-w-0 flex-1">
          <span className="block truncate font-medium text-foreground">Advanced</span>
          <span className="block truncate text-[11px] text-muted-foreground">model, effort, timeout, limits</span>
        </span>
      </Button>
    </div>
  );
}

function TodoRunProviderSubmenu({
  backend,
  choices,
  onSelect,
  onCloseParent,
}: {
  backend: RunBackendCatalog;
  choices: TodoRunModelChoice[];
  onSelect: (choice: TodoRunModelChoice) => void;
  onCloseParent: () => void;
}) {
  const HeaderIcon = iconForRunBackend(backend);
  return (
    <DropdownMenu
      align="right"
      menuLabel={`${backend.label} models`}
      menuClassName="w-[220px] max-w-[calc(100vw-24px)]"
      trigger={
        <Button
          variant="ghost"
          type="button"
          className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
          title={backend.label}
          aria-label={backend.label}
        >
          <HeaderIcon className="shrink-0 text-sm text-muted-foreground" />
          <span className="min-w-0 flex-1">
            <span className="block truncate font-medium text-foreground">{backend.label}</span>
          </span>
          <UiChevronRight className="shrink-0 text-[11px] text-muted-foreground" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          {choices.map((choice) => (
            <Button
              key={choice.key}
              variant="ghost"
              type="button"
              onClick={() => {
                close();
                onCloseParent();
                onSelect(choice);
              }}
              className="flex h-auto w-full items-center justify-start rounded px-2 py-1.5 text-left hover:bg-muted"
            >
              <span className="min-w-0 flex-1">
                <span className="block truncate font-medium text-foreground">{choice.modelShort}</span>
              </span>
            </Button>
          ))}
        </div>
      )}
    </DropdownMenu>
  );
}

export function TodoRunSplitButton({
  disabled,
  loading,
  label = "Run",
  icon = UiPlay,
  tone = "default",
  title = "Run todo",
  onRun,
  onAdvanced,
}: {
  disabled?: boolean;
  loading?: boolean;
  label?: string;
  icon?: ComponentType<IconProps>;
  tone?: "default" | "danger";
  title?: string;
  onRun: (options?: TodoRunOptions) => void;
  onAdvanced: () => void;
}) {
  const context = useTodoRunContext(!disabled);
  const primaryOptions = loadLastTodoRunOptions("run", context);
  const primaryTone = tone === "danger" ? "text-red-600 hover:bg-red-500/10 hover:text-red-700" : "text-foreground hover:bg-muted";
  const PrimaryIcon = loading ? Spinner : icon;
  return (
    <div className="inline-flex h-8 shrink-0 items-stretch rounded-md border border-border bg-background">
      <Button
        variant="ghost"
        type="button"
        onClick={() => onRun(primaryOptions)}
        disabled={disabled}
        title={title}
        className={`inline-flex h-8 items-center gap-1 rounded-none border-r border-border px-2 text-xs font-medium disabled:opacity-50 ${primaryTone}`}
      >
        <PrimaryIcon className="text-xs" />
        <span>{label}</span>
      </Button>
      <DropdownMenu
        align="right"
        menuLabel="Run todo"
        menuClassName="max-h-[70vh] w-[320px] max-w-[calc(100vw-24px)] overflow-y-auto"
        trigger={
          <Button variant="ghost" size="icon" type="button" disabled={disabled} title="Run options" aria-label="Run options" className="h-8 w-7 rounded-none text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50">
            <UiChevronDown className="text-xs" />
          </Button>
        }
      >
        {close => (
          <TodoRunDropdownContent
            context={context}
            initialAction="run"
            closeParent={close}
            onSelect={(selectedAction, options) => onRun(rememberTodoRunOptions(selectedAction, options))}
            onAdvanced={() => onAdvanced()}
          />
        )}
      </DropdownMenu>
    </div>
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
  onRun: (options?: TodoRunOptions) => void;
  onAdvanced: (action: TodoRunAction) => void;
}) {
  const config = RUN_ACTION_CONFIG[action];
  const context = useTodoRunContext(!disabled);
  const [selectedOptions, setSelectedOptions] = useState<TodoRunOptions | null>(null);
  useEffect(() => {
    setSelectedOptions(null);
  }, [action]);
  const lastOptions = selectedOptions ?? loadLastTodoRunOptions(action, context);
  const primaryTone = tone === "danger" ? "text-red-600 hover:bg-red-500/10 hover:text-red-700" : "text-foreground hover:bg-muted";
  const PrimaryIcon = loading ? Spinner : icon ?? config.icon;

  function runWith(selectedAction: TodoRunAction, options: TodoRunOptions, advanced = false) {
    const remembered = rememberTodoRunOptions(selectedAction, options, advanced);
    if (selectedAction === action) setSelectedOptions(remembered);
    onRun(remembered);
  }

  return (
    <div className="inline-flex h-8 shrink-0 items-stretch rounded-md border border-border bg-background">
      <Button
        variant="ghost"
        type="button"
        onClick={() => runWith(action, lastOptions)}
        disabled={disabled}
        title={title ?? config.title}
        className={`inline-flex h-8 items-center gap-1 rounded-none border-r border-border px-2 text-xs font-medium disabled:opacity-50 ${primaryTone}`}
      >
        <PrimaryIcon className="text-xs" />
        <span>{label ?? runButtonLabelForOptions(action, lastOptions, context)}</span>
      </Button>
      <DropdownMenu
        align="right"
        menuLabel={`${config.label} todo options`}
        menuClassName="max-h-[70vh] w-[320px] max-w-[calc(100vw-24px)] overflow-y-auto"
        trigger={
          <Button variant="ghost" size="icon" type="button" disabled={disabled} title={`${config.label} options`} aria-label={`${config.label} options`} className="h-8 w-7 rounded-none text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50">
            <UiChevronDown className="text-xs" />
          </Button>
        }
      >
        {close => (
          <TodoRunDropdownContent
            context={context}
            initialAction={action}
            closeParent={close}
            onSelect={runWith}
            onAdvanced={onAdvanced}
          />
        )}
      </DropdownMenu>
    </div>
  );
}

function runChoiceDetail(options: TodoRunOptions, fallback: string, context?: RunContext | null): string {
  const resolved = runContextWithFallback(context);
  const backend = backendForOptions(resolved, options);
  const mode = backend ? runtimeModeLabel(runtimeModeForBackend(backend)) : options.backend || options.driver || options.mode || fallback;
  const modelID = options.model || backend?.defaultModel || "default model";
  const model = shortTodoRunModelName(backend ? labelForRunModel(backend, modelID) : modelID);
  const effort = options.effort ? ` · ${options.effort}` : "";
  return `${mode} · ${model}${effort}`;
}

const INITIAL_RUNTIME_VALUE: AISpecRuntimeValue = { backend: "claude-cmux", model: "claude-sonnet-5", effort: "medium", ...AUTO_COMMIT };
const INITIAL_VERIFY_BACKEND = "claude-agent";
const INITIAL_VERIFY_MODEL = "claude-sonnet-5";

export function TodoRunAdvancedDialog({
  open,
  onClose,
  onRun,
  loading,
  title = "Run todo",
  initialMode = "run",
  dir,
  provider,
  refs,
}: {
  open: boolean;
  onClose: () => void;
  onRun: (options: TodoRunOptions) => void;
  loading?: boolean;
  title?: string;
  initialMode?: RunMode;
  // dir/provider/refs identify the todo(s) this dialog will run, so it can fetch
  // a live preview of the prompt that will be sent as the options change.
  dir: string;
  provider: string;
  refs: string[];
}) {
  // Run/Plan share one AISpecRuntimeValue (model/backend/effort/budget/
  // permissions/prompt), edited via clicky's PromptRunEditor. Verify only ever
  // needs a captain backend + model (its wire body has no effort/budget/
  // permissions), so it keeps its own small, independent selection instead of
  // sharing — and instead of showing Effort/Budget/Permissions controls that
  // verify's request would silently ignore.
  const [runtimeValue, setRuntimeValue] = useState<AISpecRuntimeValue>(INITIAL_RUNTIME_VALUE);
  const [mode, setMode] = useState<RunMode>("run");
  // Resume stays a discrete toggle (session-identity decision, cmux only); dirty,
  // auto-commit, dry-run, and checks now live on runtimeValue's spec (Workspace/
  // Commit/Verify sections), not as parallel booleans.
  const [resume, setResume] = useState(false);
  const [verifyAgent, setVerifyAgent] = useState<TodoRunAgent>("claude");
  const [verifyBackend, setVerifyBackend] = useState(INITIAL_VERIFY_BACKEND);
  const [verifyModel, setVerifyModel] = useState(INITIAL_VERIFY_MODEL);
  // promptDraft is the editable prompt body sent as the verbatim override;
  // promptDirty stops the live preview from clobbering the user's edits.
  const [promptDraft, setPromptDraft] = useState("");
  const [promptDirty, setPromptDirty] = useState(false);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState("");
  const [verifyBusy, setVerifyBusy] = useState(false);
  const [verifyError, setVerifyError] = useState("");
  const [regenNonce, setRegenNonce] = useState(0);
  const [runContext, setRunContext] = useState<RunContext | null>(null);
  // Ref mirror of promptDirty so the preview effect can read it without
  // refetching on every keystroke.
  const promptDirtyRef = useRef(false);

  const context = runContextWithFallback(runContext);
  const families = buildRunFamilies(context);
  const agent = agentForBackend(context, runtimeValue.backend);
  const isCmux = isCmuxBackend(agent, runtimeValue.backend);
  const activeModels = modelsForSelection(context, agent, runtimeValue.backend);
  const modelFallback = defaultModelForSelection(context, agent, runtimeValue.backend);
  const { driver, runBackend } = driverForSelection(context, agent, runtimeValue.backend);
  const plan = mode === "plan";
  const isVerify = mode === "verify";
  const canVerify = refs.length === 1; // verify scores one issue's commits
  const verifyBackends = backendsForAgent(context, verifyAgent).filter((item) => !isCmuxBackend(verifyAgent, item.id));
  const verifyBackendCatalog =
    verifyBackends.find((item) => item.id === verifyBackend) ??
    verifyBackends[0] ??
    findBackendCatalog(context, verifyBackend, verifyAgent);
  const verifyModelFallback = verifyBackendCatalog.defaultModel;

  // A backend switch (cmux <-> a captain backend, or a family switch) can leave
  // a model id that no longer belongs to the new mode (for example after
  // switching providers) — reset
  // to the new mode's default in that case, mirroring the old changeMechanism/
  // changeProvider/changeBackend resets.
  function changeRuntime(next: AISpecRuntimeValue) {
    const prevBackend = runtimeValue.backend ?? "";
    const nextBackend = next.backend ?? "";
    if (nextBackend === prevBackend) {
      setRuntimeValue(next);
      return;
    }
    const nextAgent = agentForBackend(context, nextBackend);
    const candidates = modelsForSelection(context, nextAgent, nextBackend);
    const modelStillValid = !!next.model && candidates.some((m) => m.id === next.model);
    setRuntimeValue({
      ...next,
      model: modelStillValid ? next.model : defaultModelForSelection(context, nextAgent, nextBackend),
    });
  }

  function changeMode(next: RunMode) {
    setMode(next);
    setPromptDirty(false);
    promptDirtyRef.current = false;
    setVerifyError("");
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

  function changeVerifyAgent(next: TodoRunAgent) {
    const nextBackend =
      backendsForAgent(context, next).find((item) => !isCmuxBackend(next, item.id)) ??
      defaultBackendForAgent(context, next);
    setVerifyAgent(next);
    setVerifyBackend(nextBackend.id);
    setVerifyModel(nextBackend.defaultModel);
  }

  function changeVerifyBackend(next: string) {
    const nextBackend = findBackendCatalog(context, next, verifyAgent);
    setVerifyBackend(nextBackend.id);
    setVerifyModel(nextBackend.defaultModel);
  }

  useEffect(() => {
    if (!open) return;
    setRuntimeValue(INITIAL_RUNTIME_VALUE);
    setMode(initialMode);
    setResume(false);
    setVerifyAgent("claude");
    setVerifyBackend(INITIAL_VERIFY_BACKEND);
    setVerifyModel(INITIAL_VERIFY_MODEL);
    setPromptDraft("");
    setPromptDirty(false);
    promptDirtyRef.current = false;
    setVerifyError("");
  }, [open, initialMode]);

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    fetch("/api/todos/run/context")
      .then(async (res) => {
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || "Failed to load run context");
        if (!cancelled) setRunContext(data as RunContext);
      })
      .catch(() => {
        if (!cancelled) setRunContext(null);
      });
    return () => {
      cancelled = true;
    };
  }, [open]);

  // refs is a fresh array each render at the call sites, so key the preview fetch
  // on its contents rather than its identity to avoid an endless refetch loop.
  const refsKey = refs.join("\n");
  const previewModel = isVerify ? verifyModel.trim() || verifyModelFallback : runtimeValue.model?.trim() || modelFallback;
  const previewBackend = isVerify ? verifyBackend : runBackend;

  // Fetch the prompt that will be sent whenever the dialog is open and a
  // prompt-affecting option changes (driver/model/effort/plan/resume). The
  // server builds it from the same code path the run uses, so it matches exactly.
  // Fetch the generated prompt body (Run/Plan) or verify prompt (Verify) and seed
  // the editor with it unless the user has edited it. The server builds it from
  // the same code path the run/verify uses, so it matches what would be sent.
  useEffect(() => {
    if (!open) {
      setPreviewError("");
      return;
    }
    const list = refsKey.split("\n").filter(Boolean);
    if (list.length === 0) {
      setPreviewError("");
      return;
    }
    const url = isVerify ? `/api/todos/verify/preview?${todoQuery(dir, provider)}` : `/api/todos/run/preview?${todoQuery(dir, provider)}`;
    const body = isVerify
      ? { provider, dir, ref: list[0], backend: previewBackend, model: previewModel }
      : {
          refs: list,
          driver,
          backend: previewBackend,
          model: previewModel,
          effort: runtimeValue.effort,
          runMode: plan ? "plan" : "run",
          resume: isCmux ? resume : undefined,
        };

    let cancelled = false;
    const controller = new AbortController();
    setPreviewLoading(true);
    setPreviewError("");
    fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      signal: controller.signal,
    })
      .then(async (res) => {
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || "Preview failed");
        if (!cancelled && !promptDirtyRef.current) setPromptDraft((data as TodoRunPreviewResponse).prompt ?? "");
      })
      .catch((err: any) => {
        if (!cancelled && err?.name !== "AbortError") setPreviewError(err?.message || "Preview failed");
      })
      .finally(() => {
        if (!cancelled) setPreviewLoading(false);
      });
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [open, dir, provider, refsKey, driver, previewBackend, previewModel, runtimeValue.effort, plan, resume, isCmux, isVerify, regenNonce]);

  if (!open) return null;

  // runVerify POSTs the (edited) verify prompt to the verification endpoint and
  // closes on success; the parent's todo polling reflects the new status.
  async function runVerify() {
    const list = refsKey.split("\n").filter(Boolean);
    if (list.length === 0) return;
    setVerifyBusy(true);
    setVerifyError("");
    try {
      const res = await fetch(`/api/todos/verify?${todoQuery(dir, provider)}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ provider, dir, ref: list[0], backend: verifyBackend, model: previewModel, prompt: promptDraft }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "Verify failed");
      onClose();
    } catch (err: any) {
      setVerifyError(err?.message || "Verify failed");
    } finally {
      setVerifyBusy(false);
    }
  }

  function submit() {
    if (isVerify) {
      void runVerify();
      return;
    }
    // spec carries the model/effort/permissions plus the run's setup (dirty),
    // workflow.verify (checks) and workflow.postRun (auto-commit / dry-run) — all
    // edited via the spec editor's Workspace/Verify/Commit sections. Plan-only runs
    // never commit or verify; the server suppresses both for run mode plan.
    const { spec } = promptRuntimeValueToPayload(runtimeValue);
    onRun({
      ...spec,
      driver,
      backend: runBackend,
      runMode: plan ? "plan" : "run",
      plan: plan ? true : undefined,
      resume: isCmux ? resume : undefined,
      prompt: {
        // The edited prompt body is sent verbatim as the override.
        user: promptDraft.trim() ? promptDraft : undefined,
      },
    });
  }

  const modeOptions: { id: RunMode; label: string }[] = [{ id: "run", label: "Run" }];
  modeOptions.push({ id: "plan", label: "Plan" });
  if (canVerify) modeOptions.push({ id: "verify", label: "Verify" });
  const verifyBackendOptions = verifyBackends.map((item) => ({
    value: item.id,
    label: item.configured === false ? `${item.label} (not ready)` : item.label,
  }));

  // Shared regenerate/error/editor/dirty-note block: passed to PromptRunEditor's
  // `promptEditor` slot for Run/Plan (it supplies the "Prompt" block title), and
  // rendered under a matching manual title for Verify (which doesn't use
  // PromptRunEditor, since its wire body has no effort/budget/permissions to edit).
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
          <Button onClick={submit} loading={isVerify ? verifyBusy : loading}>
            {isVerify ? "Verify" : plan ? "Plan" : "Run"}
          </Button>
        </div>
      }
    >
      <div className="space-y-3">
        <Field label="Mode">
          <SegmentedControl aria-label="Mode" value={mode} onChange={(v) => changeMode(v as RunMode)} options={modeOptions} />
        </Field>
        {isVerify ? (
          <>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
              <Field label="Agent">
                <ProviderSelector ariaLabel="Agent" value={verifyAgent} onChange={changeVerifyAgent} providers={PROVIDERS.map((p) => ({ id: p.id, label: p.label, icon: p.icon }))} />
              </Field>
              <Field label="Backend">
                <Combobox ariaLabel="Captain backend" value={verifyBackendCatalog.id} onChange={changeVerifyBackend} options={verifyBackendOptions} allowCustomValue={false} required />
              </Field>
              <Field label="Model">
                <ModelSelector models={verifyBackendCatalog.models} value={verifyModel} onChange={setVerifyModel} className="w-full" />
              </Field>
            </div>
            <div className="space-y-1">
              <div className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Verify prompt</div>
              {promptEditorNode}
              {verifyError && <div className="text-xs text-red-600">{verifyError}</div>}
            </div>
          </>
        ) : (
          <PromptRunEditor value={runtimeValue} onChange={changeRuntime} models={activeModels} families={families} tools={context.tools} specSections={RUN_SPEC_SECTIONS} promptEditor={promptEditorNode} promptLabel="Prompt">
            {isCmux && (
              // Resume is the one run-orchestration toggle without a spec home: it
              // continues the todo's prior claude session (--resume) rather than
              // minting a fresh one, and only applies to cmux runs.
              <label className="inline-flex items-center gap-2 text-xs">
                <input type="checkbox" checked={resume} onChange={(e) => setResume(e.currentTarget.checked)} />
                <span>Resume session</span>
              </label>
            )}
          </PromptRunEditor>
        )}
      </div>
    </Modal>
  );
}
