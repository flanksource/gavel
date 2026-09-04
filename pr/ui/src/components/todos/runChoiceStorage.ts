import type { TodoRunOptions } from "../../types";

// The client-side bookkeeping for the run dialog: which action a request maps
// onto, the localStorage-backed "last used" / "recent advanced" history, and
// the wire-facing step name a set of options resolves to. Split out of run.tsx
// (which pulled the whole run/lifecycle catalog together) purely to keep that
// file under the 500-line limit — every export here still reaches its callers
// through run.tsx's own re-exports, so nothing outside these two files needs
// to know the split exists.

// TodoRunAction is the prompt the dialog runs. Triage shares plan's behaviour
// class but is a different prompt, which is why action and mode are separate
// types: several prompts map onto one class.
export type TodoRunAction = "run" | "plan" | "triage";
export const TODO_RUN_ACTIONS: readonly TodoRunAction[] = ["run", "plan", "triage"] as const;

type RunChoiceState = {
  last: Partial<Record<TodoRunAction, TodoRunOptions>>;
  recentAdvanced: Partial<Record<TodoRunAction, TodoRunOptions[]>>;
};

// v3: POST /api/todos/run now decodes strictly and rejects runMode/driver/
// prompt at the top level — the request body is built fresh from `step` (see
// requestStepFor) rather than spreading a v2 entry's shape onto the wire, so a
// stored v2 entry is discarded rather than migrated, same as v1 -> v2 before it.
const RUN_CHOICE_STORAGE_KEY = "gavel.pr-ui.todoRunChoices.v3";

function emptyRunChoiceState(): RunChoiceState {
  return { last: {}, recentAdvanced: {} };
}

function cloneRunOptions(options: TodoRunOptions): TodoRunOptions {
  return JSON.parse(JSON.stringify(options)) as TodoRunOptions;
}

export function normalizeRunOptions(action: TodoRunAction, options: TodoRunOptions): TodoRunOptions {
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

export function readRunChoiceState(): RunChoiceState {
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

export function writeRunChoiceState(state: RunChoiceState): void {
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

export function actionFromRunOptions(options: TodoRunOptions): TodoRunAction {
  if (options.prompt === "triage") return "triage";
  return options.plan || options.runMode === "plan" ? "plan" : "run";
}

// requestStepFor is the lifecycle step name a request actually dispatches.
// `step` is authoritative when set (buildTodoRunPayload and the verify
// mutation set it directly); otherwise it falls back to whatever prompt/
// runMode/plan the options carry for their own bookkeeping — a custom
// `.gavel.yaml` prompt name survives via `prompt`, run/plan fall out of the
// legacy runMode/plan flags. This is the one place those fields are read for
// the wire; POST /api/todos/run itself only ever sees `step`.
export function requestStepFor(options: TodoRunOptions): string {
  if (options.step) return options.step;
  if (options.prompt) return options.prompt;
  return options.plan || options.runMode === "plan" ? "plan" : "run";
}
