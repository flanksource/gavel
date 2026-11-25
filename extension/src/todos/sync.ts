import { serverOrigin } from "../settings";
import { TodoApi, TodoHttpError } from "./api";
import { validateCapture, type Capture, type Catalog, type QueuedTodo, type Snapshot } from "./model";
import { cacheKey, TodoStore } from "./store";
import type { RunContext, RunResult, RunStep, TodoRun } from "./runtime";

export class TodoSync {
  private readonly store: TodoStore;
  private readonly api: TodoApi;
  private running: Promise<void> | undefined;

  constructor(options: { store: TodoStore; fetcher?: typeof fetch }) {
    this.store = options.store;
    this.api = new TodoApi(options.fetcher);
  }

  async capture(server: string, input: Capture): Promise<QueuedTodo> {
    validateCapture(input);
    const todo: QueuedTodo = { ...input, dir: input.dir.trim(), id: crypto.randomUUID(), server: serverOrigin(server), createdAt: Date.now(), state: "pending" };
    await this.store.remember(cacheKey(todo.server, "recent", input.sourceUrl ? new URL(input.sourceUrl).host : ""), todo.dir);
    await this.store.put(todo);
    return todo;
  }

  async snapshot(server: string, dir: string, sourceUrl: string, status = "draft"): Promise<Snapshot> {
    const [projects, labels, todos, recentDir, queue, runContext, steps, runChoice] = await Promise.all([
      this.store.cache<Catalog["projects"]>(cacheKey(server, "projects")),
      this.store.cache<Catalog["labels"]>(cacheKey(server, "labels", dir)),
      this.store.cache<Catalog["todos"]>(cacheKey(server, "todos", dir)),
      this.store.cache<string>(cacheKey(server, "recent", sourceUrl ? new URL(sourceUrl).host : "")),
      this.store.queue(server),
      this.store.cache<RunContext>(cacheKey(server, "run-context", dir)),
      this.store.cache<RunStep[]>(cacheKey(server, `steps:${status}`, dir)),
      this.store.cache<TodoRun | null>(cacheKey(server, "run-choice", dir)),
    ]);
    return { ...(projects ? { projects } : {}), ...(labels ? { labels } : {}), ...(todos ? { todos } : {}), ...(recentDir ? { recentDir } : {}), ...(runContext ? { runContext } : {}), ...(steps ? { steps } : {}), ...(runChoice !== undefined ? { runChoice } : {}), queue };
  }

  async refreshRun(server: string, dir: string, status: string): Promise<void> {
    if (!dir.trim()) throw new Error("Select a project before loading its runtime");
    const results = await Promise.allSettled([
      this.api.runContext(server, dir).then(context => this.store.remember(cacheKey(server, "run-context", dir), context)),
      this.api.steps(server, dir, status).then(steps => this.store.remember(cacheKey(server, `steps:${status}`, dir), steps)),
    ]);
    const errors = results.flatMap(result => result.status === "rejected" ? [result.reason] : []);
    if (errors.length) throw new AggregateError(errors, errors.map(String).join("; "));
  }

  async rememberRun(server: string, dir: string, run: TodoRun | null): Promise<void> {
    await this.store.remember(cacheKey(server, "run-choice", dir), run);
  }

  async refresh(server: string, dir: string): Promise<void> {
    const projects = await this.api.projects(server);
    await this.store.remember(cacheKey(server, "projects"), projects);
    const dirs = dir ? [dir] : projects.map(project => project.dir).filter(Boolean);
    const results = await Promise.allSettled(dirs.map(async workspace => {
      const labels = await this.api.labels(server, workspace);
      await this.store.remember(cacheKey(server, "labels", workspace), labels);
      if (dir) await this.store.remember(cacheKey(server, "todos", workspace), await this.api.todos(server, workspace));
    }));
    const failures = results.flatMap(result => result.status === "rejected" ? [result.reason] : []);
    if (failures.length) throw new AggregateError(failures, failures.map(String).join("; "));
  }

  sync(): Promise<void> {
    if (this.running) return this.running;
    this.running = this.flush().finally(() => { this.running = undefined; });
    return this.running;
  }

  private async flush(): Promise<void> {
    for (const todo of await this.store.queue()) {
      if (todo.state === "sending") {
        await this.store.put({ ...todo, state: "uncertain", error: "The extension stopped during delivery. Check Gavel before retrying to avoid a duplicate." });
        continue;
      }
      if (todo.state !== "pending") continue;
      try {
        await this.api.projects(todo.server);
      } catch (cause) {
        await this.store.put({ ...todo, error: String(cause) });
        continue;
      }
      await this.store.put({ ...todo, state: "sending" });
      let result: { ref: string; run?: RunResult };
      try {
        result = todo.remoteRef && todo.run
          ? { ref: todo.remoteRef, run: await this.api.run(todo.server, todo.dir, todo.remoteRef, todo.run) }
          : await this.api.deliver(todo.server, todo);
      } catch (cause) {
        const rejected = cause instanceof TodoHttpError && cause.status >= 400 && cause.status < 500;
        await this.store.put({ ...todo, state: rejected ? todo.remoteRef ? "run_error" : "error" : "uncertain", error: String(cause) });
        continue;
      }
      const failed = result.run?.status === "failed" || result.run?.status === "skipped";
      const unknown = result.run && !["started", "failed", "skipped"].includes(result.run.status);
      await this.store.put({ ...todo, state: unknown ? "uncertain" : failed ? "run_error" : "synced", remoteRef: result.ref, attachments: [], ...(result.run ? { runResult: result.run } : {}), error: failed || unknown ? result.run?.error ?? `Run ${result.run?.status}` : undefined });
    }
  }

  async retry(server: string, id: string): Promise<void> {
    if (this.running) await this.running;
    const todo = (await this.store.queue(server)).find(entry => entry.id === id);
    if (!todo) throw new Error(`Unknown queued todo ${id}`);
    if (todo.state === "synced" || todo.state === "sending") throw new Error(`Cannot retry a ${todo.state} todo`);
    await this.store.put({ ...todo, state: "pending", error: undefined });
    await this.sync();
  }
}
