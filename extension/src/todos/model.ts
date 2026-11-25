export type Project = { name: string; dir: string; repos?: string[] };
export type Labels = { definitions: { name: string; description?: string }[]; counts?: Record<string, number> };
export type Todo = { ref: string; title: string; status: string; shortId?: string };
export type Attachment = { name: string; dataUrl: string };
export const priorities = ["high", "medium", "low"] as const;
export const statuses = ["draft", "pending", "verified", "completed", "skipped"] as const;
export type Capture = {
  mode: "new" | "existing";
  dir: string;
  title: string;
  body: string;
  priority: typeof priorities[number];
  status: typeof statuses[number];
  labels: string[];
  ref: string;
  pr: string;
  sourceUrl: string;
  attachments: Attachment[];
  run?: TodoRun | undefined;
};
export type QueuedTodo = Capture & {
  id: string;
  server: string;
  createdAt: number;
  state: "pending" | "sending" | "error" | "uncertain" | "synced" | "run_error";
  error?: string | undefined;
  remoteRef?: string;
  runResult?: RunResult;
};
export type Catalog = { projects?: Project[]; labels?: Labels; todos?: Todo[]; recentDir?: string; runContext?: RunContext; steps?: RunStep[]; runChoice?: TodoRun | null };
export type Snapshot = Catalog & { queue: QueuedTodo[] };

export function capturePR(value: string): { number: number; url: string } | undefined {
  if (!value.trim()) return undefined;
  const url = new URL(value.trim());
  const number = Number(url.pathname.split("/").filter(Boolean).at(-1));
  if (!["http:", "https:"].includes(url.protocol) || url.username || url.password || !Number.isSafeInteger(number) || number <= 0) throw new Error("Enter a pull request URL with a positive PR number");
  return { number, url: url.href.replace(/\/$/, "") };
}

export function captureFromParams(params: URLSearchParams): Capture {
  const first = (...keys: string[]) => keys.map(key => params.get(key)).find(value => value?.trim()) ?? "";
  const ref = first("ref", "todo", "issue");
  const autoSave = /^(1|true|yes|on)$/i.test(first("autoSave", "autosave", "auto_save"));
  return {
    mode: ref || first("mode", "target") === "existing" ? "existing" : "new",
    dir: first("dir", "workspace"), title: first("title", "name"), body: first("body", "description", "text"),
    priority: paramChoice(first("priority", "severity"), priorities, "medium"),
    status: paramChoice(first("status"), statuses, autoSave ? "pending" : "draft"),
    labels: [...new Set(["labels", "label"].flatMap(key => params.getAll(key)).flatMap(value => value.split(",")).map(label => label.trim()).filter(Boolean))],
    ref, pr: first("pr"), sourceUrl: first("sourceUrl", "sourceURL", "url"), attachments: [],
  };
}

function paramChoice<T extends string>(value: string, choices: readonly T[], defaultValue: T): T {
  if (!value) return defaultValue;
  const choice = choices.find(choice => choice === value);
  if (!choice) throw new Error(`Invalid todo choice ${value}; expected one of ${choices.join(", ")}`);
  return choice;
}

export function validateCapture(value: Capture): void {
  if (!value || !["new", "existing"].includes(value.mode)) throw new Error("Invalid todo capture mode");
  if (!value.dir?.trim()) throw new Error("Choose a workspace or enter its directory");
  if (value.mode === "new" && !value.title?.trim()) throw new Error("A todo title is required");
  if (value.mode === "existing" && (!value.ref?.trim() || (!value.body?.trim() && !value.attachments.length))) throw new Error("Choose an issue and add a comment or screenshot");
  if (!priorities.includes(value.priority) || !statuses.includes(value.status)) throw new Error("Invalid todo priority or status");
  if (!Array.isArray(value.labels) || !value.labels.every(label => typeof label === "string" && !!label.trim())) throw new Error("Invalid todo labels");
  for (const attachment of value.attachments) {
    if (!attachment.name || !/^data:image\/[\w.+-]+;base64,[A-Za-z0-9+/]*={0,2}$/.test(attachment.dataUrl)) throw new Error("Screenshots must be image data URLs with a filename");
  }
  if (value.mode === "new") capturePR(value.pr);
  if (value.run) {
    if (value.mode !== "new") throw new Error("Run after creation is only available for a new todo");
    if (!value.run.step?.trim() || !value.run.spec || typeof value.run.spec !== "object" || Array.isArray(value.run.spec)) throw new Error("A run requires a lifecycle step and runtime spec");
    if (value.run.presets && !value.run.presets.every(preset => typeof preset === "string" && !!preset.trim())) throw new Error("Invalid run presets");
  }
}
import type { RunContext, RunResult, RunStep, TodoRun } from "./runtime";

