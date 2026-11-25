import { IDBFactory } from "fake-indexeddb";
import { expect, it, vi } from "vitest";
import { CommentSync, type CommentSubmission } from "./comment-sync";
import { TodoStore } from "./todos/store";

const target = { server: "http://localhost:19092", project: "acme" };
const page = { url: "https://example.com/docs?view=full", title: "Docs" };
const dataUrl = "data:image/png;base64,iVBORw0KGgo=";
const draft: CommentSubmission = {
  page,
  input: { body: "Check the heading", anchor: "#hero" },
  selection: { anchor: "#hero", label: "Hero", element: { source: "fixture", html: "<h1/>", pageUrl: page.url, screenshot: { status: "captured", dataUrl } } },
  deliver: { kind: "run", prompt: "editor-comments", fresh: true, runOptions: { spec: { model: "acme-model", mode: "agent", effort: "high", budget: { cost: 3 } } } },
};

it("adds comment storage without dropping saved TODO captures or cached choices", async () => {
  const factory = new IDBFactory();
  const existing = { id: "todo-1", title: "Review the sidebar", state: "pending", createdAt: 1 };
  await new Promise<void>((resolve, reject) => {
    const open = factory.open("gavel-todo-capture", 1);
    open.onupgradeneeded = () => {
      open.result.createObjectStore("queue", { keyPath: "id" }).put(existing);
      open.result.createObjectStore("cache").put("acme", "project");
    };
    open.onsuccess = () => { open.result.close(); resolve(); };
    open.onerror = () => reject(open.error);
  });
  const store = new TodoStore(factory);
  const sync = new CommentSync({ store });
  const id = await sync.capture(target, draft);
  expect(await Promise.all([store.queue(), store.cache("project"), sync.list(target, page.url)])).toEqual([
    [existing], "acme", [{ id, state: "pending", body: draft.input.body }],
  ]);
});

it("persists a screenshot and runtime before contacting Gavel and replays the identical create after worker restart", async () => {
  const factory = new IDBFactory();
  const requests: string[] = [];
  const fetcher = vi.fn<typeof fetch>().mockImplementation(async (_url, init) => {
    requests.push(String(init?.body));
    throw new TypeError("Response lost");
  });
  const sync = new CommentSync({ store: new TodoStore(factory), fetcher });
  const id = await sync.capture(target, draft);
  expect(fetcher).not.toHaveBeenCalled();
  await sync.sync();

  const restarted = new CommentSync({ store: new TodoStore(factory), fetcher: async (_url, init) => {
    requests.push(String(init?.body));
    return Response.json({ id: "server-thread", bodyMarkdown: draft.input.body, content: JSON.parse(String(init?.body)).content, version: 2, status: "open", createdAt: "2026-11-01T10:00:00Z", updatedAt: "2026-11-01T10:00:00Z" });
  } });
  expect(await restarted.list(target, page.url)).toEqual([{ id, state: "pending", body: draft.input.body, error: expect.stringContaining("Response lost") }]);
  expect(await restarted.list({ ...target, project: "other" }, page.url)).toEqual([]);
  await Promise.all([restarted.sync(), restarted.sync()]);
  expect(requests).toEqual([requests[0], requests[0]]);
  expect(JSON.parse(requests[0]!)).toMatchObject({ requestId: id, deliver: draft.deliver, attachments: [{ filename: "screenshot.png", data: "iVBORw0KGgo=" }] });
  expect(await restarted.list(target, page.url)).toEqual([{ id, state: "synced", remoteId: "server-thread", adoptSession: false }]);
});

it("keeps rejected drafts visible and stops resubmitting until retried", async () => {
  const fetcher = vi.fn<typeof fetch>().mockResolvedValue(Response.json({ error: "Runtime unavailable" }, { status: 400 }));
  const sync = new CommentSync({ store: new TodoStore(new IDBFactory()), fetcher });
  const id = await sync.capture(target, draft);
  await sync.sync();
  await sync.sync();
  expect(fetcher).toHaveBeenCalledTimes(1);
  expect(await sync.list(target, page.url)).toEqual([{ id, state: "error", body: draft.input.body, error: expect.stringContaining("Runtime unavailable") }]);
});

it("keeps the selected mention and returns Gavel's ID for automatic session adoption", async () => {
  const { deliver, ...capture } = draft;
  const fetcher = vi.fn<typeof fetch>().mockImplementation(async (url, init) => {
    const body = JSON.parse(String(init?.body));
    if (new URL(String(url)).pathname.endsWith("/delivery/preview")) {
      return Response.json({ mention: !!body.delivery.mentionId, route: { target: deliver } });
    }
    return Response.json({ id: `gavel-${body.requestId}`, content: body.content, version: 1, status: "open", createdAt: "2026-11-01T10:00:00Z", updatedAt: "2026-11-01T10:00:00Z" });
  });
  const sync = new CommentSync({ store: new TodoStore(new IDBFactory()), fetcher });
  const automatic = await sync.capture(target, { ...capture, delivery: {} });
  const targeted = await sync.capture(target, { ...capture, delivery: { mentionId: "agent:review" } });
  await sync.sync();
  expect(Object.fromEntries((await sync.list(target, page.url)).map(({ id, ...receipt }) => [id, receipt]))).toEqual({
    [automatic]: { state: "synced", remoteId: `gavel-${automatic}`, adoptSession: true },
    [targeted]: { state: "synced", remoteId: `gavel-${targeted}`, adoptSession: false },
  });
  const targetedCreate = fetcher.mock.calls.map(([, init]) => JSON.parse(String(init?.body))).find(body => body.requestId === targeted);
  expect(targetedCreate?.delivery).toEqual({ mentionId: "agent:review" });
});
