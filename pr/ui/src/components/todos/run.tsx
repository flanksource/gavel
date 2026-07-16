import { lazy, Suspense, useCallback, useEffect, useRef, useState, type ComponentType } from "react";
import { Button, DropdownMenu, Field, Modal, SegmentedControl } from "@flanksource/clicky-ui/components";
import { ModelSelector, type ChatModel } from "@flanksource/clicky-ui/chat";
import { effortOptionsForModel, PromptRunEditor, promptRuntimeValueToPayload, reconcileModelCapabilities, type AISpecRuntimeValue } from "@flanksource/clicky-ui/ai";
import type { StaticIconComponent } from "@flanksource/clicky-ui/data";
import { UiChevronDown, UiCloud, UiCog, UiColumns, UiListDashes, UiPlay, UiRobotAi, UiSparkles, UiTerminal, type IconProps } from "@flanksource/clicky-ui/icons";
import type { TodoRunAgent, TodoRunEffort, TodoRunOptions, TodoRunPreviewResponse, TodoRunResponse } from "../../types";
import { Spinner } from "../../icons/Spinner";
import { inputClass, todoQuery } from "./format";
import { TodoRunEffortBadge, todoRunEffortPresentation } from "./TodoRunEffortBadge";
export { TodoRunEffortBadge, todoRunEffortPresentation } from "./TodoRunEffortBadge";
import {
  PROVIDERS,
  agentBackendForAgent,
  agentForBackend,
  backendCatalog as findBackendCatalog,
  backendsForAgent,
  buildRunFamilies,
  defaultModelForSelection,
  driverForSelection,
  isCmuxBackend,
  modelsForSelection,
  runContextWithFallback,
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

// Runs auto-commit by default — the old commit=true default, now expressed as the
// spec's Workflow.PostRun.commit. The advanced dialog's Commit section can turn it
// off; a plan-only run never commits (the server suppresses commit in plan mode).
const AUTO_COMMIT: Pick<TodoRunOptions, "workflow"> = { workflow: { postRun: { commit: true } } };

// MdxEditorField is the same markdown editor field JsonSchemaForm uses for its
// markdown fields. It lazily pulls in the heavy @mdxeditor/editor, so it is
// code-split and rendered under Suspense with a plain-textarea fallback.
const MdxEditorField = lazy(() => import("@flanksource/clicky-ui/mdx-editor").then((m) => ({ default: m.MdxEditorField })));

export const defaultRunOptions: TodoRunOptions = { driver: "cli", backend: "claude-agent", model: "claude-sonnet-5", effort: "medium", ...AUTO_COMMIT };

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
      { label: "Claude", icon: UiSparkles, options: { driver: "cli", backend: "claude-agent", model: "claude-sonnet-5", effort: "medium", ...AUTO_COMMIT } },
      { label: "Codex", icon: UiTerminal, options: { driver: "cli", backend: "codex-agent", model: "gpt-5.5", effort: "medium", ...AUTO_COMMIT } },
    ],
  },
  {
    action: "Plan",
    detail: "plan only · no changes",
    presets: [
      { label: "Claude", icon: UiSparkles, options: { driver: "cli", backend: "claude-agent", model: "claude-sonnet-5", effort: "medium", runMode: "plan", plan: true } },
      { label: "Codex", icon: UiTerminal, options: { driver: "cli", backend: "codex-agent", model: "gpt-5.5", effort: "medium", runMode: "plan", plan: true } },
    ],
  },
];

const RUN_CHOICE_STORAGE_KEY = "gavel.pr-ui.todoRunChoices.v1";

const RUN_ACTION_CONFIG: Record<TodoRunAction, { label: string; detail: string; icon: ComponentType<IconProps>; title: string }> = {
  run: { label: "Run", detail: "implement", icon: UiPlay, title: "Run todo" },
  plan: { label: "Plan", detail: "plan only", icon: UiListDashes, title: "Plan todo" },
};

const RUNTIME_MODE_ORDER: TodoRunRuntimeMode[] = ["agent", "cmux", "cli", "api"];
const RUNTIME_MODE_CONFIG: Record<TodoRunRuntimeMode, { label: string; icon: StaticIconComponent }> = {
  cmux: { label: "cmux", icon: UiColumns },
  agent: { label: "Agent", icon: UiRobotAi },
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

export function useTodoRunContext(enabled = true): RunContext {
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
  const providerIcon = PROVIDERS.find(provider => provider.id === backend.agent)?.icon;
  if (providerIcon && typeof providerIcon !== "string") return providerIcon as ComponentType<IconProps>;
  return backend.agent === "claude" ? UiSparkles : UiRobotAi;
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

const ALL_EFFORTS: TodoRunEffort[] = ["low", "medium", "high", "xhigh", "max", "ultra"];

function modelForRunBackend(backend: RunBackendCatalog, modelID: string | undefined): ChatModel {
  return modelsForRunBackend(backend).find(model => model.id === modelID) ?? fallbackModelForBackend(backend);
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

export function todoRunModelFamilyName(value: string | undefined): string {
  const raw = (value || "").trim();
  if (!raw) return "Default";
  const lower = raw.toLowerCase();
  for (const family of ["fable", "opus", "sonnet", "haiku"]) {
    if (lower.includes(family)) return `${family[0]!.toUpperCase()}${family.slice(1)}`;
  }
  for (const variant of ["sol", "luna", "terra", "spark"]) {
    if (new RegExp(`(?:^|[-\\s])${variant}(?:$|[-\\s])`).test(lower)) {
      return `${variant[0]!.toUpperCase()}${variant.slice(1)}`;
    }
  }
  if (lower.includes("gpt") || lower.includes("codex")) return "GPT";
  if (lower.includes("gemini")) return "Gemini";
  if (lower.includes("deepseek")) return "DeepSeek";

  const family = raw
    .split(/[\s/_-]+/)
    .find(part => part && !/^\d+(?:\.\d+)*$/.test(part) && !["claude", "agent", "code", "openai", "anthropic"].includes(part.toLowerCase()));
  return family ? `${family[0]!.toUpperCase()}${family.slice(1)}` : raw;
}

function labelForRunModel(backend: RunBackendCatalog, modelID: string): string {
  const model = modelsForRunBackend(backend).find(item => item.id === modelID);
  return model?.label || modelID;
}

function runOptionsForBackendModel(action: TodoRunAction, backend: RunBackendCatalog, modelID: string, effort: TodoRunEffort = "medium"): TodoRunOptions {
  const options = reconcileModelCapabilities({
    driver: backend.driver,
    backend: backend.id,
    model: modelID || backend.defaultModel,
    effort,
    ...(action === "run" ? AUTO_COMMIT : {}),
  } satisfies TodoRunOptions, modelForRunBackend(backend, modelID), ALL_EFFORTS) as TodoRunOptions;
  return normalizeRunOptions(action, options);
}

export function runChoicesForAction(context: RunContext, action: TodoRunAction, effort: TodoRunEffort = "medium"): TodoRunModelChoice[] {
  const config = RUN_ACTION_CONFIG[action];
  return context.backends.flatMap((backend) => {
    const mechanism = mechanismForBackend(backend);
    const Icon = iconForRunBackend(backend);
    return modelsForRunBackend(backend).map((model) => {
      const modelID = model.id || backend.defaultModel;
      const modelLabel = model.label?.trim() || "";
      const modelShort = shortTodoRunModelName(modelID);
      return {
        key: `${action}:${backend.id}:${modelID}`,
        action,
        backend,
        modelID,
        modelLabel,
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

export function runButtonQualifierForOptions(options: TodoRunOptions, context: RunContext): string {
  const backend = backendForOptions(context, options);
  const model = options.model || backend.defaultModel;
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
  const modelID = options.model || backend.defaultModel;
  const model = modelForRunBackend(backend, modelID);
  const provider = PROVIDERS.find(item => item.id === backend.agent);
  const supportedEfforts = effortOptionsForModel(model, contextEfforts(context));
  const effort = options.effort && supportedEfforts.includes(options.effort)
    ? options.effort as TodoRunEffort
    : undefined;

  return {
    provider,
    model: shortTodoRunModelName(labelForRunModel(backend, modelID)),
    effort,
  };
}

export function defaultRunOptionsForAction(action: TodoRunAction, context?: RunContext | null): TodoRunOptions {
  const resolved = runContextWithFallback(context);
  const backend = primaryBackendForAction(resolved);
  if (backend) {
    return runOptionsForBackendModel(action, backend, backend.defaultModel);
  }
	return normalizeRunOptions(action, RUN_PRESETS[action][0]?.options ?? defaultRunOptions);
}

export function reconcileTodoRunOptions(action: TodoRunAction, options: TodoRunOptions, context: RunContext): TodoRunOptions {
  const normalized = normalizeRunOptions(action, options);
  const backend = backendForOptions(context, normalized);
  const modelIsCurrent = !!normalized.model && backend.models.some(model => model.id === normalized.model);
  const model = modelIsCurrent ? normalized.model! : backend.defaultModel;
  return normalizeRunOptions(action, reconcileModelCapabilities({
    ...normalized,
    driver: backend.driver,
    backend: backend.id,
    model,
  }, modelForRunBackend(backend, model), contextEfforts(context)) as TodoRunOptions);
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
  const [runBusy, setRunBusy] = useState(false);
  const [runMessage, setRunMessage] = useState("");
  const [runError, setRunError] = useState("");

  const reset = useCallback(() => {
    setRunMessage("");
    setRunError("");
  }, []);

  const run = useCallback(
    async (ref: string, options: TodoRunOptions = defaultRunOptions): Promise<TodoRunResponse | null> => {
      const cleaned = ref.trim();
      if (!cleaned || runBusy) return null;
      setRunBusy(true);
      setRunError("");
      setRunMessage("");
      try {
        const body = { ref: cleaned, ...options };
        const response = await fetch(`/api/todos/run?${todoQuery(dir)}`, {
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
    [dir, runBusy],
  );

  return { runBusy, runMessage, runError, reset, run };
}

type TodoRunDropdownSelect = (action: TodoRunAction, options: TodoRunOptions, advanced?: boolean) => void;

export function TodoRunDropdownContent({
  context,
  initialAction,
  closeParent,
  onSelect,
  onAdvanced,
  showAdvanced = true,
}: {
  context: RunContext;
  initialAction: TodoRunAction;
  closeParent: () => void;
  onSelect: TodoRunDropdownSelect;
  onAdvanced?: (action: TodoRunAction) => void;
  showAdvanced?: boolean;
}) {
  const initialOptions = loadLastTodoRunOptions(initialAction, context);
  const initialBackend = backendForOptions(context, initialOptions);
  const [selectedAgent, setSelectedAgent] = useState<TodoRunAgent>(initialBackend.agent);
  const [effort, setEffort] = useState<TodoRunEffort>((initialOptions.effort as TodoRunEffort | undefined) ?? "medium");
  const backend = agentBackendForAgent(context, selectedAgent);
  const rememberedModel = initialBackend.id === backend.id ? initialOptions.model : undefined;
  const effortModel = modelForRunBackend(backend, rememberedModel || backend.defaultModel);
  const choices = runChoicesForAction(context, initialAction, effort).filter(choice => choice.backend.id === backend.id);

  useEffect(() => {
    const nextOptions = loadLastTodoRunOptions(initialAction, context);
    setSelectedAgent(backendForOptions(context, nextOptions).agent);
    setEffort((nextOptions.effort as TodoRunEffort | undefined) ?? "medium");
  }, [initialAction, context.backends, context.defaultBackend]);

  return (
    <div className="p-1 text-xs">
      <div className="space-y-2 border-b border-border px-2 pb-3 pt-2">
        <TodoRunProviderSegments
          value={selectedAgent}
          onChange={(nextAgent) => {
            const nextBackend = agentBackendForAgent(context, nextAgent);
            const nextModel = modelForRunBackend(nextBackend, nextBackend.defaultModel);
            const reconciled = reconcileModelCapabilities({ effort }, nextModel, contextEfforts(context));
            setSelectedAgent(nextAgent);
            setEffort((reconciled.effort as TodoRunEffort | undefined) ?? effort);
          }}
        />
        <TodoRunEffortSlider model={effortModel} fallbackEfforts={contextEfforts(context)} value={effort} onChange={setEffort} />
      </div>

      <div className="space-y-0.5 py-1">
        {choices.map(choice => {
          const provider = PROVIDERS.find(item => item.id === choice.backend.agent);
          const ProviderIcon = provider?.icon ?? choice.icon;
          const fullModelName = choice.modelLabel || choice.modelID;
          return (
            <Button
              key={choice.key}
              variant="ghost"
              type="button"
              title={fullModelName}
              onClick={() => {
                closeParent();
                onSelect(initialAction, runOptionsForBackendModel(initialAction, backend, choice.modelID, effort));
              }}
              className="group flex h-9 w-full items-center justify-start gap-2 rounded-md px-2 text-left hover:bg-muted"
            >
              <span className="inline-flex size-6 shrink-0 items-center justify-center rounded-md bg-muted/70 ring-1 ring-border/60 group-hover:bg-background">
                <ProviderIcon className="size-3.5" style={{ color: provider?.iconColor }} />
              </span>
              <span className="min-w-0 flex-1 truncate text-sm font-medium text-foreground">{fullModelName}</span>
            </Button>
          );
        })}
      </div>

      {showAdvanced && onAdvanced && (
        <>
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
        </>
      )}
    </div>
  );
}

function TodoRunProviderSegments({
  value,
  onChange,
}: {
  value: TodoRunAgent;
  onChange: (value: TodoRunAgent) => void;
}) {
  return (
    <SegmentedControl<TodoRunAgent>
      aria-label="Provider"
      size="sm"
      value={value}
      onChange={onChange}
      className="w-full"
      options={PROVIDERS.map(provider => {
        const ProviderIcon = provider.icon;
        return {
          id: provider.id,
          label: <span className="inline-flex items-center gap-1.5"><ProviderIcon className="size-3.5" style={{ color: provider.iconColor }} />{provider.label}</span>,
        };
      })}
    />
  );
}

function effortIconColor(effort: TodoRunEffort | undefined): string {
  switch (effort) {
    case "low": return "text-sky-500";
    case "medium": return "text-emerald-500";
    case "high": return "text-amber-500";
    case "xhigh": return "text-orange-500";
    case "max": return "text-rose-500";
    case "ultra": return "text-fuchsia-500";
    default: return "text-muted-foreground";
  }
}

function TodoRunEffortSlider({
  model,
  fallbackEfforts,
  value,
  onChange,
}: {
  model: ChatModel | undefined;
  fallbackEfforts: readonly TodoRunEffort[];
  value: TodoRunEffort;
  onChange: (value: TodoRunEffort) => void;
}) {
  const efforts = effortOptionsForModel(model, fallbackEfforts) as TodoRunEffort[];
  const reconciled = reconcileModelCapabilities({ effort: value }, model, fallbackEfforts);
  const selected = (reconciled.effort as TodoRunEffort | undefined) ?? efforts[0];
  const index = Math.max(0, efforts.indexOf(selected!));

  if (efforts.length === 0) {
    return <div className="flex h-8 items-center justify-center text-[11px] text-muted-foreground">Fixed effort for this model</div>;
  }
  if (efforts.length === 1) {
    return <div className="flex h-8 items-center justify-center text-[11px] font-medium capitalize text-muted-foreground">{efforts[0]}</div>;
  }

  const percent = (index / (efforts.length - 1)) * 100;
  const SelectedLowIcon = todoRunEffortPresentation(efforts[0]).icon;
  const SelectedHighIcon = todoRunEffortPresentation(efforts[efforts.length - 1]).icon;
  return (
    <div className="grid grid-cols-[auto_minmax(0,1fr)_auto] items-start gap-1.5" aria-label="Model effort">
      <SelectedLowIcon className={`mt-3 size-3.5 ${effortIconColor(efforts[0])}`} aria-hidden="true" />
      <div className="relative h-11 pt-1">
        <input
          type="range"
          min={0}
          max={efforts.length - 1}
          step={1}
          value={index}
          aria-label="Effort"
          aria-valuetext={selected}
          onChange={event => onChange(efforts[Number(event.currentTarget.value)]!)}
          className="h-1 w-full cursor-pointer appearance-none rounded-full bg-muted accent-foreground [&::-moz-range-thumb]:size-3 [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:border-2 [&::-moz-range-thumb]:border-background [&::-moz-range-thumb]:bg-foreground [&::-webkit-slider-thumb]:size-3 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:border-2 [&::-webkit-slider-thumb]:border-background [&::-webkit-slider-thumb]:bg-foreground [&::-webkit-slider-thumb]:shadow-sm"
        />
        <span aria-hidden="true" className="pointer-events-none absolute top-7 -translate-x-1/2" style={{ left: `${percent}%` }}>
          <TodoRunEffortBadge effort={selected} />
        </span>
      </div>
      <SelectedHighIcon className={`mt-3 size-3.5 ${effortIconColor(efforts[efforts.length - 1])}`} aria-hidden="true" />
    </div>
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
  const presentation = todoRunButtonPresentation(lastOptions, context);
  const ProviderIcon = presentation.provider?.icon ?? iconForRunBackend(backendForOptions(context, lastOptions));
  const effortPresentation = presentation.effort ? todoRunEffortPresentation(presentation.effort) : null;
  const EffortIcon = effortPresentation?.icon;
  const showSelection = !label && !loading;

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
        {showSelection ? (
          <ProviderIcon className="text-xs" style={{ color: presentation.provider?.iconColor }} />
        ) : (
          <PrimaryIcon className="text-xs" />
        )}
        <span>{label ?? `${config.label} (${presentation.model})`}</span>
        {showSelection && EffortIcon && effortPresentation && (
          <span className="inline-flex" title={`Effort: ${effortPresentation.label}`} aria-label={`Effort: ${effortPresentation.label}`}>
            <EffortIcon className={`size-3.5 ${effortIconColor(presentation.effort)}`} aria-hidden="true" />
          </span>
        )}
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

function TodoRunAdvancedRuntimeControls({
  context,
  value,
  onChange,
  recent,
}: {
  context: RunContext;
  value: AISpecRuntimeValue;
  onChange: (value: AISpecRuntimeValue) => void;
  recent: TodoRunOptions[];
}) {
  const selectedAgent = agentForBackend(context, value.backend);
  const selectedBackend = findBackendCatalog(context, value.backend ?? "", selectedAgent);
  const models = modelsForRunBackend(selectedBackend);
  const selectedModel = modelForRunBackend(selectedBackend, value.model || selectedBackend.defaultModel);
  const providerBackends = backendsForAgent(context, selectedAgent)
    .slice()
    .sort((a, b) => RUNTIME_MODE_ORDER.indexOf(runtimeModeForBackend(a)) - RUNTIME_MODE_ORDER.indexOf(runtimeModeForBackend(b)));

  function selectBackend(nextBackend: RunBackendCatalog) {
    const model = nextBackend.models.some(item => item.id === value.model) ? value.model! : nextBackend.defaultModel;
    onChange(reconcileModelCapabilities({ ...value, backend: nextBackend.id, model }, modelForRunBackend(nextBackend, model), contextEfforts(context)));
  }

  return (
    <div className="space-y-3">
      <Field label="Provider">
        <TodoRunProviderSegments
          value={selectedAgent}
          onChange={nextAgent => selectBackend(agentBackendForAgent(context, nextAgent))}
        />
      </Field>

      <Field label="Effort">
        <TodoRunEffortSlider
          model={selectedModel}
          fallbackEfforts={contextEfforts(context)}
          value={(value.effort as TodoRunEffort | undefined) ?? "medium"}
          onChange={effort => onChange({ ...value, effort })}
        />
      </Field>

      <Field label="Mode">
        <SegmentedControl<string>
          aria-label="Runtime mode"
          size="sm"
          value={selectedBackend.id}
          onChange={backendID => selectBackend(findBackendCatalog(context, backendID, selectedAgent))}
          className="w-full"
          options={providerBackends.map(backend => ({
            id: backend.id,
            label: RUNTIME_MODE_CONFIG[runtimeModeForBackend(backend)].label,
            icon: RUNTIME_MODE_CONFIG[runtimeModeForBackend(backend)].icon,
            disabled: backend.configured === false,
          }))}
        />
      </Field>

      <Field label="Model">
        <ModelSelector
          models={models}
          value={value.model || selectedBackend.defaultModel}
          onChange={model => onChange(reconcileModelCapabilities({ ...value, model }, modelForRunBackend(selectedBackend, model), contextEfforts(context)))}
          className="w-full"
        />
      </Field>

      {selectedModel.temperature !== false && (
        <Field label="Temperature">
          <input
            type="number"
            min={0}
            max={2}
            step={0.1}
            value={value.temperature ?? ""}
            aria-label="Temperature"
            placeholder="Model default"
            onChange={event => {
              const next = event.currentTarget.value;
              onChange({ ...value, temperature: next === "" ? undefined : Number(next) });
            }}
            className={inputClass}
          />
        </Field>
      )}

      {recent.length > 0 && (
        <div className="space-y-1 border-t border-border pt-3">
          <div className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">Recent advanced</div>
          <div className="flex flex-wrap gap-1.5">
            {recent.map((options, index) => (
              <Button
                key={runOptionsKey(options)}
                variant="outline"
                size="sm"
                type="button"
                onClick={() => onChange(reconcileTodoRunOptions(actionFromRunOptions(options), options, context))}
              >
                {index + 1}. {runChoiceDetail(options, "advanced", context)}
              </Button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

const INITIAL_RUNTIME_VALUE: AISpecRuntimeValue = { backend: "claude-agent", model: "claude-sonnet-5", effort: "medium", ...AUTO_COMMIT };

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
  // Run/Plan share one AISpecRuntimeValue (model/backend/effort/budget/
  // permissions/prompt), edited via clicky's PromptRunEditor.
  const [runtimeValue, setRuntimeValue] = useState<AISpecRuntimeValue>(INITIAL_RUNTIME_VALUE);
  const [mode, setMode] = useState<RunMode>("run");
  // Resume stays a discrete toggle (session-identity decision, cmux only); dirty,
  // auto-commit, dry-run, and checks now live on runtimeValue's spec (Workspace/
  // Commit/Verify sections), not as parallel booleans.
  const [resume, setResume] = useState(false);
  // promptDraft is the editable prompt body sent as the verbatim override;
  // promptDirty stops the live preview from clobbering the user's edits.
  const [promptDraft, setPromptDraft] = useState("");
  const [promptDirty, setPromptDirty] = useState(false);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState("");
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
  const advancedAction: TodoRunAction = plan ? "plan" : "run";
  const recentAdvanced = loadRecentAdvancedTodoRunOptions(advancedAction, context);

  // A backend switch (cmux <-> a captain backend, or a family switch) can leave
  // a model id that no longer belongs to the new mode (for example after
  // switching providers) — reset
  // to the new mode's default in that case, mirroring the old changeMechanism/
  // changeProvider/changeBackend resets.
  function changeRuntime(next: AISpecRuntimeValue) {
    const nextBackend = next.backend ?? "";
    const nextAgent = agentForBackend(context, nextBackend);
    const candidates = modelsForSelection(context, nextAgent, nextBackend);
    const modelStillValid = !!next.model && candidates.some((m) => m.id === next.model);
    const model = modelStillValid ? next.model! : defaultModelForSelection(context, nextAgent, nextBackend);
    setRuntimeValue(reconcileModelCapabilities({
      ...next,
      model,
    }, modelForRunBackend(findBackendCatalog(context, nextBackend, nextAgent), model), contextEfforts(context)));
  }

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
    setRuntimeValue(INITIAL_RUNTIME_VALUE);
    setMode(initialMode);
    setResume(false);
    setPromptDraft("");
    setPromptDirty(false);
    promptDirtyRef.current = false;
  }, [open, initialMode]);

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    fetch("/api/todos/run/context")
      .then(async (res) => {
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || "Failed to load run context");
        if (!cancelled) {
          const resolved = runContextWithFallback(data as RunContext);
          setRunContext(data as RunContext);
          const action: TodoRunAction = initialMode === "plan" ? "plan" : "run";
          setRuntimeValue(reconcileTodoRunOptions(action, loadLastTodoRunOptions(action, resolved), resolved));
        }
      })
      .catch(() => {
        if (!cancelled) setRunContext(null);
      });
    return () => {
      cancelled = true;
    };
  }, [open, initialMode]);

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
    if (!refID) {
      setPreviewError("");
      return;
    }
    const url = `/api/todos/run/preview?${todoQuery(dir)}`;
    const body = {
      ref: refID,
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
  }, [open, dir, refID, driver, previewBackend, previewModel, runtimeValue.effort, plan, resume, isCmux, regenNonce]);

  if (!open) return null;

  function submit() {
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
          <Button onClick={submit} loading={loading}>
            {plan ? "Plan" : "Run"}
          </Button>
        </div>
      }
    >
      <div className="space-y-3">
        <Field label="Mode">
          <SegmentedControl aria-label="Mode" value={mode} onChange={(v) => changeMode(v as RunMode)} options={modeOptions} />
        </Field>
        <PromptRunEditor
          value={runtimeValue}
          onChange={changeRuntime}
          models={activeModels}
          families={families}
          tools={context.tools}
          specSections={RUN_SPEC_SECTIONS}
          promptEditor={promptEditorNode}
          promptLabel="Prompt"
          runtimeControls={<TodoRunAdvancedRuntimeControls context={context} value={runtimeValue} onChange={changeRuntime} recent={recentAdvanced} />}
        >
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
      </div>
    </Modal>
  );
}
