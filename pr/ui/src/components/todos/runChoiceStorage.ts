import type { TodoRunOptions } from "../../types";

// The client-side bookkeeping for the run dialog: which action a request maps
// onto, the localStorage-backed "last used" / "recent advanced" history, and
// the wire-facing step name a set of options resolves to. Split out of run.tsx
// (which pulled the whole run/lifecycle catalog together) purely to keep that
// file under the 500-line limit — every export here still reaches its callers
// through run.tsx's own re-exports, so nothing outside these two files needs
// to know the split exists.

// TodoRunAction is the name of the lifecycle step the dialog runs. The set of
// steps is declared by the server (todo.lifecycle.steps and the run context's
// lifecycle.steps), so no step name is enumerated here.
export type TodoRunAction = string;

const CREATE_STEP_STORAGE_KEY = "gavel.pr-ui.todoCreateStep.v1";

// readCreateStepChoice is the lifecycle step last picked in a create dialog's
// "run after creation" control. An empty string is None, which is also what a
// browser with no stored choice (or unavailable storage) starts from.
export function readCreateStepChoice(): string {
  try {
    if (typeof localStorage === "undefined") return "";
    return localStorage.getItem(CREATE_STEP_STORAGE_KEY) ?? "";
  } catch {
    return "";
  }
}

export function writeCreateStepChoice(step: string): void {
  try {
    if (typeof localStorage === "undefined") return;
    localStorage.setItem(CREATE_STEP_STORAGE_KEY, step);
  } catch {
    // Persistence is best-effort; unavailable storage should not block creating a todo.
  }
}

type RunChoiceState = {
  last: Partial<Record<string, TodoRunOptions>>;
  recentAdvanced: Partial<Record<string, TodoRunOptions[]>>;
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

export function normalizeRunOptions(action: string, options: TodoRunOptions): TodoRunOptions {
  if (!action.trim()) throw new Error('A lifecycle step is required');
  const next = cloneRunOptions(options);
  next.step = action;
  delete next.driver;
  delete next.prompt;
  delete next.runMode;
  delete next.plan;
  return next;
}

// withoutPromptContent drops the prompt body and attachments a run was
// dispatched with. Both were rendered from one todo, so a remembered run choice
// carrying them replays that todo's prompt on every later run of the step. An
// emptied prompt section is removed rather than kept as `prompt: {}`, which the
// server reads as an explicit, empty prompt override.
export function withoutPromptContent(options: TodoRunOptions): TodoRunOptions {
  if (!options.spec?.prompt) return options;
  const { prompt, ...spec } = options.spec;
  const { user: _user, attachments: _attachments, ...kept } = prompt;
  const hasKept = Object.values(kept).some(value => value !== undefined);
  return { ...options, spec: hasKept ? { ...spec, prompt: kept } : spec };
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
    for (const action of new Set([...Object.keys(last), ...Object.keys(recent)])) {
      const lastOptions = coerceRunOptions(last[action]);
      if (lastOptions) state.last[action] = withoutPromptContent(normalizeRunOptions(action, lastOptions));
      const recentOptions = recent[action];
      if (!Array.isArray(recentOptions)) continue;
      state.recentAdvanced[action] = recentOptions
        .map(coerceRunOptions)
        .filter((item): item is TodoRunOptions => !!item)
        .map(item => withoutPromptContent(normalizeRunOptions(action, item)))
        .slice(0, 3);
    }
    return state;
  } catch {
    return emptyRunChoiceState();
  }
}

// lastTodoRunOptions returns the raw "last used" entry for a lifecycle step,
// with no synthesized fallback — undefined means nothing has actually been
// run for this step yet. Contrast with run.tsx's loadLastTodoRunOptions,
// which falls back to defaultRunOptionsForAction so the quick-run buttons
// always have something to dispatch; this reader backs the advanced dialogs'
// "Last used" restore option, which must only appear when there is a real
// remembered choice to offer.
export function lastTodoRunOptions(action: string): TodoRunOptions | undefined {
  return readRunChoiceState().last[action];
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

export function actionFromRunOptions(options: TodoRunOptions): string {
  return requestStepFor(options);
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
