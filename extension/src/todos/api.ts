import { capturePR, type Capture, type Labels, type Project, type Todo } from "./model";
import type { RunContext, RunResult, RunStep, TodoRun } from "./runtime";

export class TodoHttpError extends Error {
  constructor(readonly status: number, message: string) { super(message); }
}

export class TodoApi {
  constructor(private readonly fetcher: typeof fetch = (input, init) => fetch(input, init)) {}

  private async json(server: string, path: string, init: RequestInit = {}): Promise<unknown> {
    const response = await this.fetcher(new URL(path, server).href, { ...init, signal: AbortSignal.timeout(15_000) });
    if (!response.ok) throw new TodoHttpError(response.status, `${init.method ?? "GET"} ${path}: HTTP ${response.status}: ${await response.text()}`);
    return response.json();
  }

  async projects(server: string): Promise<Project[]> {
    const value = await this.json(server, "/api/projects");
    if (!Array.isArray(value) || !value.every((entry: Partial<Project> | null) => entry && typeof entry.name === "string" && typeof entry.dir === "string")) throw new Error("Gavel returned invalid projects");
    return value as Project[];
  }

  async labels(server: string, dir: string): Promise<Labels> {
    const value = await this.json(server, `/api/todos/labels?${new URLSearchParams({ dir })}`) as Labels | null;
    if (!value || !Array.isArray(value.definitions) || !value.definitions.every(entry => typeof entry.name === "string")) throw new Error("Gavel returned invalid labels");
    return value;
  }

  async todos(server: string, dir: string): Promise<Todo[]> {
    const value = await this.json(server, `/api/todos?${new URLSearchParams({ dir })}`) as { items?: Todo[] } | null;
    if (!Array.isArray(value?.items) || !value.items.every(entry => typeof entry.ref === "string" && typeof entry.title === "string" && typeof entry.status === "string")) throw new Error("Gavel returned invalid todos");
    return value.items;
  }

  async runContext(server: string, dir: string): Promise<RunContext> {
    const value = await this.json(server, `/api/todos/run/context?${new URLSearchParams({ dir })}`) as RunContext | null;
    if (!value || !Array.isArray(value.models) || !value.models.length || !Array.isArray(value.efforts) || !Array.isArray(value.runtimes) || !value.runtimes.length || value.runtimes.some(family => !Array.isArray(family.modes) || !family.modes.length)) throw new Error("Gavel returned an unavailable or invalid runtime catalog");
    return value;
  }

  async steps(server: string, dir: string, status: string): Promise<RunStep[]> {
    const value = await this.json(server, `/api/todos/new/steps?${new URLSearchParams({ dir, status })}`) as { steps?: RunStep[] } | null;
    if (!Array.isArray(value?.steps) || !value.steps.every(step => typeof step.name === "string" && typeof step.label === "string")) throw new Error("Gavel returned invalid lifecycle steps");
    return value.steps;
  }

  async run(server: string, dir: string, ref: string, run: TodoRun): Promise<RunResult> {
    const value = await this.json(server, `/api/todos/run?${new URLSearchParams({ dir })}`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ ref, ...run }) }) as RunResult | null;
    if (!value || typeof value.status !== "string") throw new Error("Gavel did not return the run result");
    return value;
  }

  async deliver(server: string, capture: Capture): Promise<{ ref: string; run?: RunResult }> {
    const existing = capture.mode === "existing";
    const fields = existing ? { ref: capture.ref, comment: capture.body } : {
      title: capture.title, body: capture.body, priority: capture.priority, status: capture.status, labels: capture.labels,
      ...(capture.pr ? { pr: capturePR(capture.pr) } : {}),
      ...(capture.run ? { step: capture.run.step, spec: capture.run.spec, ...(capture.run.presets ? { presets: capture.run.presets } : {}) } : {}),
    };
    let body: BodyInit;
    if (capture.attachments.length) {
      const form = new FormData();
      for (const [name, value] of Object.entries(fields)) {
        if (name === "labels" && Array.isArray(value)) value.forEach(label => form.append(name, label));
        else form.append(name, typeof value === "string" ? value : JSON.stringify(value));
      }
      for (const attachment of capture.attachments) {
        const comma = attachment.dataUrl.indexOf(",");
        const mime = attachment.dataUrl.slice(5, attachment.dataUrl.indexOf(";"));
        const bytes = Uint8Array.from(atob(attachment.dataUrl.slice(comma + 1)), char => char.charCodeAt(0));
        form.append("attachment", new Blob([bytes], { type: mime }), attachment.name);
      }
      body = form;
    } else body = JSON.stringify(fields);
    const value = await this.json(server, `/api/todos/${existing ? "item" : "new"}?${new URLSearchParams({ dir: capture.dir })}`, {
      method: existing ? "PATCH" : "POST", body,
      ...(typeof body === "string" ? { headers: { "Content-Type": "application/json" } } : {}),
    }) as { todo?: { ref?: string }; ref?: string; run?: RunResult } | null;
    const ref = existing ? value?.ref : value?.todo?.ref;
    if (!ref) throw new Error("Gavel did not return the saved todo reference");
    if (capture.run && typeof value?.run?.status !== "string") return { ref, run: { status: "unknown", error: "Todo saved, but Gavel did not return the requested run result. Check its runs before retrying." } };
    return { ref, ...(value?.run ? { run: value.run } : {}) };
  }
}
