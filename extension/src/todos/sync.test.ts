import { IDBFactory } from "fake-indexeddb";
import { describe, expect, it, vi } from "vitest";
import { TodoStore, cacheKey } from "./store";
import type { Capture } from "./model";
import { TodoSync } from "./sync";
import { context, steps } from "./testing/runtime";

const server = "http://localhost:19092";
const dir = "/work/acme";
const capture: Capture = { mode: "new", dir, title: "Repair card", body: "Captured context", priority: "medium", status: "draft", labels: ["ui"], ref: "", pr: "", sourceUrl: "https://acme.example", attachments: [] };
const projects = [{ name: "acme", dir }];
const json = (value: unknown, status = 200) => new Response(JSON.stringify(value), { status });

describe("extension todo storage and sync", () => {
  it("persists capture and caches across worker restarts, then sends once after recovery", async () => {
    const factory = new IDBFactory();
    const store = new TodoStore(factory);
    const offline = vi.fn().mockRejectedValue(new TypeError("Connection refused"));
    const sync = new TodoSync({ store, fetcher: offline });
    const saved = await sync.capture(server, capture);
    await store.remember(cacheKey(server, "projects"), projects);
    await store.remember(cacheKey(server, "labels", dir), { definitions: [{ name: "ui" }] });
    await sync.sync();
    const restarted = new TodoStore(factory);
    expect(await restarted.queue()).toEqual([expect.objectContaining({ id: saved.id, state: "pending", body: capture.body, error: expect.stringContaining("Connection refused") })]);
    expect(await restarted.cache(cacheKey(server, "projects"))).toEqual(projects);
    expect(await restarted.cache(cacheKey(server, "labels", "/other"))).toBeUndefined();
    const fetcher = vi.fn(async (_url: unknown, init?: RequestInit) => init?.method === "POST" ? json({ todo: { ref: "t-created" } }) : json(projects));
    const recovered = new TodoSync({ store: restarted, fetcher });
    await Promise.all([recovered.sync(), recovered.sync()]);
    await recovered.sync();
    expect(fetcher.mock.calls.filter(([, init]) => init?.method === "POST")).toHaveLength(1);
    expect(await restarted.queue()).toEqual([expect.objectContaining({ id: saved.id, state: "synced", remoteRef: "t-created", error: undefined })]);
  });

  it("keeps screenshot bytes and comment targets in multipart delivery", async () => {
    const store = new TodoStore(new IDBFactory());
    const fetcher = vi.fn(async (_url: unknown, init?: RequestInit) => init?.method === "PATCH" ? json({ ref: "t-existing" }) : json(projects));
    const sync = new TodoSync({ store, fetcher });
    await sync.capture(server, { ...capture, mode: "existing", ref: "t-existing", attachments: [{ name: "card.png", dataUrl: "data:image/png;base64,aW1hZ2U=" }] });
    await sync.sync();
    const mutation = fetcher.mock.calls.find(([, init]) => init?.method === "PATCH");
    expect(String(mutation?.[0])).toBe(`${server}/api/todos/item?dir=%2Fwork%2Facme`);
    const form = mutation?.[1]?.body as FormData;
    expect([form.get("ref"), form.get("comment"), (form.get("attachment") as File).name, (form.get("attachment") as File).size]).toEqual(["t-existing", capture.body, "card.png", 5]);
  });

  it("retains rejected submissions and does not repeat ambiguous mutations", async () => {
    const store = new TodoStore(new IDBFactory());
    const fetcher = vi.fn(async (_url: unknown, init?: RequestInit) => {
      if (init?.method === "POST") throw new TypeError("Response lost");
      return json(projects);
    });
    const sync = new TodoSync({ store, fetcher });
    await sync.capture(server, capture);
    await sync.sync();
    await sync.sync();
    expect((await store.queue())[0]).toMatchObject({ state: "uncertain", error: expect.stringContaining("Response lost") });
    expect(fetcher.mock.calls.filter(([, init]) => init?.method === "POST")).toHaveLength(1);
  });

  it("rejects invalid capture before persisting it", async () => {
    const store = new TodoStore(new IDBFactory());
    const sync = new TodoSync({ store });
    await expect(sync.capture(server, { ...capture, dir: "" })).rejects.toThrow("workspace");
    await expect(sync.capture(server, { ...capture, title: " " })).rejects.toThrow("title");
    expect(await store.queue()).toEqual([]);
  });

  it("retains project and workspace label caches after refresh fails and isolates servers", async () => {
    const store = new TodoStore(new IDBFactory());
    const labels = { definitions: [{ name: "ui", description: "Interface" }], counts: { ui: 3 } };
    const fetcher = vi.fn(async (url: unknown) => String(url).includes("/labels") ? json(labels) : json(projects));
    const sync = new TodoSync({ store, fetcher });
    await sync.refresh(server, "");
    fetcher.mockRejectedValue(new TypeError("Offline"));
    await expect(sync.refresh(server, dir)).rejects.toThrow("Offline");
    expect(await sync.snapshot(server, dir, capture.sourceUrl)).toEqual({ projects, labels, queue: [] });
    expect(await sync.snapshot("http://localhost:19093", dir, capture.sourceUrl)).toEqual({ queue: [] });
  });

  it("surfaces HTTP validation errors without losing the submitted data", async () => {
    const store = new TodoStore(new IDBFactory());
    const fetcher = vi.fn(async (_url: unknown, init?: RequestInit) => init?.method === "POST" ? json({ error: "Workspace missing" }, 400) : json(projects));
    const sync = new TodoSync({ store, fetcher });
    await sync.capture(server, capture);
    await sync.sync();
    expect((await store.queue())[0]).toMatchObject({ ...capture, state: "error", error: expect.stringContaining("Workspace missing") });
  });

  it("holds an in-flight submission for review after a worker stops", async () => {
    const store = new TodoStore(new IDBFactory());
    const sync = new TodoSync({ store });
    const todo = await sync.capture(server, capture);
    await store.put({ ...todo, state: "sending" });
    await sync.sync();
    expect((await store.queue())[0]).toMatchObject({ state: "uncertain", error: expect.stringContaining("extension stopped") });
  });

  it.each([false, true])("delivers a queued lifecycle run with its complete runtime (screenshot=%s)", async screenshot => {
    const factory = new IDBFactory();
    const store = new TodoStore(factory);
    const run = { step: "plan", presets: ["careful"], spec: { model: "acme-model", mode: "agent", effort: "high", budget: { cost: 3 }, permissions: { mode: "auto" as const } } };
    const sync = new TodoSync({ store });
    await sync.capture(server, { ...capture, run, attachments: screenshot ? [{ name: "card.png", dataUrl: "data:image/png;base64,aW1hZ2U=" }] : [] });
    const fetcher = vi.fn(async (_url: unknown, init?: RequestInit) => init?.method === "POST" ? json({ todo: { ref: "t-run" }, run: { status: "started", sessionId: "s-run" } }) : json(projects));
    await new TodoSync({ store: new TodoStore(factory), fetcher }).sync();
    const body = fetcher.mock.calls.find(([, init]) => init?.method === "POST")?.[1]?.body;
    if (screenshot) {
      const form = body as FormData;
      expect([form.get("step"), JSON.parse(String(form.get("spec"))), JSON.parse(String(form.get("presets")))]).toEqual([run.step, run.spec, run.presets]);
    } else expect(JSON.parse(String(body))).toMatchObject(run);
    expect((await store.queue())[0]).toMatchObject({ run, remoteRef: "t-run", runResult: { status: "started", sessionId: "s-run" } });
  });

  it("retries a failed run on the created todo without creating it again", async () => {
    const store = new TodoStore(new IDBFactory());
    const run = { step: "run", spec: { model: "acme-model", mode: "agent" } };
    const fetcher = vi.fn(async (url: unknown, init?: RequestInit) => {
      if (String(url).includes("/todos/run?")) return json({ status: "started", promptRunId: "p-run" });
      return init?.method === "POST" ? json({ todo: { ref: "t-created" }, run: { status: "failed", error: "Agent unavailable" } }) : json(projects);
    });
    const sync = new TodoSync({ store, fetcher });
    const saved = await sync.capture(server, { ...capture, run });
    await sync.sync();
    expect((await store.queue())[0]).toMatchObject({ state: "run_error", remoteRef: "t-created", error: "Agent unavailable" });
    await sync.retry(server, saved.id);
    const mutations = fetcher.mock.calls.filter(([, init]) => init?.method === "POST");
    expect(mutations.map(([url]) => new URL(String(url)).pathname)).toEqual(["/api/todos/new", "/api/todos/run"]);
    expect(JSON.parse(String(mutations[1]?.[1]?.body))).toEqual({ ref: "t-created", ...run });
    expect((await store.queue())[0]).toMatchObject({ state: "synced", remoteRef: "t-created", runResult: { status: "started", promptRunId: "p-run" } });
  });

  it("keeps runtime catalogs, lifecycle steps and choices offline per workspace and status", async () => {
    const factory = new IDBFactory();
    const store = new TodoStore(factory);
    const fetcher = vi.fn(async (url: unknown) => String(url).includes("/run/context") ? json(context) : json({ steps }));
    const sync = new TodoSync({ store, fetcher });
    const choice = { step: "plan", spec: { model: "acme-model", mode: "agent", effort: "high" } };
    await sync.refreshRun(server, dir, "draft");
    await sync.rememberRun(server, dir, choice);
    fetcher.mockRejectedValue(new TypeError("Offline"));
    await expect(sync.refreshRun(server, dir, "draft")).rejects.toThrow("Offline");
    const restarted = new TodoSync({ store: new TodoStore(factory) });
    expect(await restarted.snapshot(server, dir, capture.sourceUrl, "draft")).toMatchObject({ runContext: context, steps, runChoice: choice });
    expect((await restarted.snapshot(server, dir, capture.sourceUrl, "pending")).steps).toBeUndefined();
    expect(await restarted.snapshot(server, "/work/other", capture.sourceUrl, "draft")).toEqual({ queue: [] });
  });

  it("holds missing run outcomes for review while preserving the created todo reference", async () => {
    const store = new TodoStore(new IDBFactory());
    const fetcher = vi.fn(async (_url: unknown, init?: RequestInit) => init?.method === "POST" ? json({ todo: { ref: "t-created" } }) : json(projects));
    const sync = new TodoSync({ store, fetcher });
    await sync.capture(server, { ...capture, run: { step: "run", spec: {} } });
    await sync.sync();
    await sync.sync();
    expect((await store.queue())[0]).toMatchObject({ state: "uncertain", remoteRef: "t-created", runResult: { status: "unknown" } });
    expect(fetcher.mock.calls.filter(([, init]) => init?.method === "POST")).toHaveLength(1);
  });
});
