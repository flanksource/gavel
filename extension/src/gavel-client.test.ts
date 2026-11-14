import { expect, it } from "vitest";
import type { NativeComment } from "../comments/gavel-projection";
import { freshTarget } from "./agent-retry";
import {
  GavelClient,
  commentAgentFor,
  GavelError,
  dataUrlBytes,
  domSelector,
  fetchRuntimeCatalog,
  listProjects,
  pageKey,
  pageTarget,
  parseSse,
  projectScope,
} from "./gavel-client";

const server = "http://localhost:9092";
const page = { url: "https://example.com/docs?b=2&a=1#intro", title: "Docs" };
const scope = projectScope("my site");
const targetUri = `${scope}/pages/${encodeURIComponent("https://example.com/docs#intro")}`;
const codex = { spec: { model: "codex-example", effort: "high" }, presets: ["fast"] };

type Call = { url: URL; method: string; headers: Record<string, string>; body: unknown };

/** A fake dashboard: answers each request with the next queued response and records what was asked. */
function fakeDashboard(...responses: (Response | ((call: Call) => Response))[]) {
  const calls: Call[] = [];
  const fetcher: typeof fetch = async (input, init) => {
    const headers = Object.fromEntries(new Headers(init?.headers).entries());
    const body = init?.body instanceof Uint8Array || init?.body instanceof Blob ? init.body : typeof init?.body === "string" ? JSON.parse(init.body) : undefined;
    const call = { url: new URL(String(input)), method: init?.method ?? "GET", headers, body };
    calls.push(call);
    const next = responses.shift();
    if (!next) throw new Error(`Unexpected request ${call.method} ${call.url.href}`);
    return typeof next === "function" ? next(call) : next;
  };
  return { calls, client: new GavelClient({ server, project: "my site", fetch: fetcher }) };
}
const json = (value: unknown, status = 200) => new Response(JSON.stringify(value), { status, headers: { "Content-Type": "application/json" } });

const native = (overrides: Partial<NativeComment> = {}): NativeComment => ({
  id: "c-1",
  bodyMarkdown: "Too small",
  author: { kind: "user", id: "u-1", label: "Ada" },
  target: { uri: targetUri, scopeUri: scope, kind: "web-page" },
  status: "open",
  version: 1,
  createdAt: "2026-11-01T10:00:00Z",
  updatedAt: "2026-11-01T10:00:00Z",
  content: { playground: { page: pageKey(page.url), anchor: "#hero", body: "Too small" }, page },
  ...overrides,
});

it("names the project scope the way gavel's path escaping does", () => {
  expect(projectScope("my site")).toBe("gavel://project/my%20site");
  expect(projectScope("a/b;c,d?e")).toBe("gavel://project/a%2Fb%3Bc%2Cd%3Fe");
  expect(projectScope("it's (x)*!")).toBe("gavel://project/it%27s%20%28x%29%2A%21");
  expect(projectScope("a+b=c:d@e&f$g")).toBe("gavel://project/a+b=c:d@e&f$g");
  expect(projectScope("héllo")).toBe("gavel://project/h%C3%A9llo");
});

it("targets a page by its URL without the query, and describes the commented element as a dom selector", () => {
  expect(pageKey(page.url)).toBe("https://example.com/docs#intro");
  const locators = [{ kind: "css" as const, selector: "#hero" }];
  const selector = domSelector({ anchor: "#hero", label: "Hero", element: { source: "https://example.com/docs", html: "<h1>Hi</h1>", locators, pageUrl: page.url } });
  expect(selector).toEqual({ kind: "dom", anchor: "#hero", locators, html: "<h1>Hi</h1>" });
  expect(pageTarget({ scope, page, selector })).toEqual({ uri: targetUri, scopeUri: scope, kind: "web-page", label: "Docs", selector });
  expect(pageTarget({ scope, page })).toEqual({ uri: targetUri, scopeUri: scope, kind: "web-page", label: "Docs" });
  expect(targetUri.startsWith(`${scope}/`)).toBe(true);
});

it("lists a page's comments across pages of results and projects them with direct attachment URLs", async () => {
  const row = (id: string) => ({ ...native({ id, attachments: [{ id: "a-1", name: "screenshot.png" }], content: { playground: { page: pageKey(page.url), anchor: "#hero", element: { source: "s", html: "<h1/>" } }, page } }), readVersion: 0, unread: true });
  const first = Array.from({ length: 200 }, (_, index) => row(`c-${index}`));
  const { calls, client } = fakeDashboard(json({ data: first, page: { total: 201, limit: 200, offset: 0 } }), json({ data: [row("c-200")], page: { total: 201, limit: 200, offset: 200 } }));
  const comments = await client.list(page);
  expect(comments).toHaveLength(201);
  expect(comments[0]?.element?.screenshot).toEqual({ status: "captured", filename: "screenshot.png", url: `${server}/api/v1/comments/c-0/attachments/a-1` });
  expect(calls.map(call => Object.fromEntries(call.url.searchParams))).toEqual([
    { scope, target: targetUri, limit: "200", offset: "0" },
    { scope, target: targetUri, limit: "200", offset: "200" },
  ]);
  expect(calls[0]?.url.pathname).toBe("/api/v1/comments");
  expect(calls[0]?.headers).not.toHaveProperty("x-gavel-comment-actor");
});

it("creates a human-review comment on the page target, carrying page, selector and element capture", async () => {
  const { calls, client } = fakeDashboard(json(native({ version: 1 })));
  const element = { source: "https://example.com/docs", html: "<h1>Hi</h1>", locators: [{ kind: "css" as const, selector: "#hero" }], pageUrl: page.url };
  const created = await client.create({ page, comment: { id: "req-1", body: "Too small", createdAt: "2026-11-01T10:00:00Z", author: { name: "You", kind: "user" }, anchor: "#hero" }, selection: { anchor: "#hero", label: "Hero", element } });
  expect(created.id).toBe("c-1");
  expect(calls).toHaveLength(1);
  expect(calls[0]).toMatchObject({ method: "POST", body: {
    target: { uri: targetUri, scopeUri: scope, kind: "web-page", label: "Docs", selector: { kind: "dom", anchor: "#hero", html: "<h1>Hi</h1>" } },
    bodyMarkdown: "Too small",
    workflowId: "human-review",
    requestId: "req-1",
    content: { page, playground: { page: pageKey(page.url), anchor: "#hero", body: "Too small", element: { html: "<h1>Hi</h1>" } } },
  } });
});

it("uploads a captured screenshot as an attachment at the created version and strips it from the stored element", async () => {
  const png = "data:image/png;base64,iVBORw0KGgo=";
  const { calls, client } = fakeDashboard(json(native({ version: 1 })), json(native({ version: 2, attachments: [{ id: "a-9", name: "screenshot.png" }] })));
  const element = { source: "s", html: "<h1/>", pageUrl: page.url, screenshot: { status: "captured" as const, dataUrl: png } };
  await client.create({ page, comment: { id: "req-2", body: "Look", createdAt: "t", author: null, anchor: "#hero" }, selection: { anchor: "#hero", label: "Hero", element } });
  expect(JSON.stringify(calls[0]?.body)).not.toContain("iVBOR");
  expect(calls[1]?.method).toBe("POST");
  expect(calls[1]?.url.pathname).toBe("/api/v1/comments/c-1/attachments");
  expect(Object.fromEntries(calls[1]?.url.searchParams ?? [])).toEqual({ name: "screenshot.png", expectedVersion: "1" });
  expect(calls[1]?.headers["content-type"]).toBe("image/png");
  const upload = calls[1]?.body as Uint8Array;
  expect([...upload]).toEqual([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);
});

it("removes the incomplete comment when the screenshot upload fails and reports the upload error", async () => {
  const { calls, client } = fakeDashboard(json(native({ version: 1 })), json({ error: "attachment too large" }, 413), json([native({ version: 2 })]));
  const element = { source: "s", html: "<h1/>", pageUrl: page.url, screenshot: { status: "captured" as const, dataUrl: "data:image/png;base64,iVBORw0KGgo=" } };
  await expect(client.create({ page, comment: { id: "req-3", body: "x", createdAt: "t", author: null, anchor: "#hero" }, selection: { anchor: "#hero", label: "Hero", element } })).rejects.toThrow(/attachment too large/);
  expect(calls[2]?.url.pathname).toBe("/api/v1/comments/batch");
  expect(calls[2]?.body).toEqual([{ id: "c-1", expectedVersion: 1, action: "delete" }]);
});

it("sends lifecycle actions with the expected version and reports gavel's refusal with the endpoint", async () => {
  const { calls, client } = fakeDashboard(json(native({ version: 3, status: "resolved" })), json({ error: "comment version or request conflict" }, 409));
  const resolved = await client.act({ id: "c-1", version: 2, action: "resolve" });
  expect(resolved.status).toBe("resolved");
  expect(calls[0]).toMatchObject({ method: "POST", body: { expectedVersion: 2 } });
  expect(calls[0]?.url.pathname).toBe("/api/v1/comments/c-1/resolve");
  const failure = await client.act({ id: "c-1", version: 2, action: "close" }).catch((cause: unknown) => cause);
  expect(failure).toBeInstanceOf(GavelError);
  expect(failure).toMatchObject({ status: 409, message: expect.stringContaining("version or request conflict") });
  expect((failure as GavelError).message).toContain("/api/v1/comments/c-1/close");
});

it.each([
  { name: "applying", extra: { action: "accept-proposal", messageId: "m-2", mode: "apply" as const }, body: { expectedVersion: 2, messageId: "m-2", mode: "apply" } },
  { name: "rejecting with a reason", extra: { action: "reject-proposal", messageId: "m-2", bodyMarkdown: "Too broad" }, body: { expectedVersion: 2, messageId: "m-2", bodyMarkdown: "Too broad" } },
  { name: "requesting a revision", extra: { action: "request-revision", messageId: "m-2", bodyMarkdown: "Keep the API" }, body: { expectedVersion: 2, messageId: "m-2", bodyMarkdown: "Keep the API" } },
])("posts a proposal decision to the thread root when $name", async ({ extra, body }) => {
  const { calls, client } = fakeDashboard(json(native({ version: 3 })));
  await client.act({ id: "c-1", version: 2, ...extra });
  expect(calls[0]?.method).toBe("POST");
  expect(calls[0]?.url.pathname).toBe(`/api/v1/comments/c-1/${extra.action}`);
  expect(calls[0]?.body).toEqual(body);
});

it("fetches a proposal's diff for a reply on a thread and reports gavel's refusal", async () => {
  const diff = { diff: "diff --git a/x b/x\n", commit: "def456" };
  const { calls, client } = fakeDashboard(json(diff), json({ error: "proposal not found" }, 404));
  expect(await client.proposalDiff("c-1", "m 2")).toEqual(diff);
  expect(calls[0]?.method).toBe("GET");
  expect(calls[0]?.url.pathname).toBe("/api/v1/comments/c-1/proposals/m%202/diff");
  await expect(client.proposalDiff("c-1", "m-9")).rejects.toMatchObject({ status: 404, message: expect.stringContaining("proposal not found") });
});

it("replies on the thread root with the reply stored as the playground shape", async () => {
  const { calls, client } = fakeDashboard(json(native({ version: 2, messages: [{ id: "m-1", bodyMarkdown: "Fixed", author: { kind: "user", id: "u-1" }, createdAt: "2026-11-01T10:01:00Z", content: { playground: { page: pageKey(page.url) } } }] })));
  const thread = await client.reply({ page, root: "c-1", version: 1, body: "Fixed" });
  expect(thread.map(comment => comment.id)).toEqual(["c-1", "m-1"]);
  expect(calls[0]?.url.pathname).toBe("/api/v1/comments/c-1/reply");
  expect(calls[0]?.body).toMatchObject({ expectedVersion: 1, bodyMarkdown: "Fixed", content: { playground: { page: pageKey(page.url), body: "Fixed", author: { name: "You", kind: "user" } } } });
});

it("delivers a thread to a session or as a new run grouped by its page, and records the hand-off", async () => {
  const sent = native({ version: 2, content: { playground: { page: pageKey(page.url) }, page, agentRun: { sessionId: "s-1", requestedAt: "2026-11-01T10:02:00Z", state: "sent", target: { label: "claude: site", channel: "cmux" } } } });
  const { calls, client } = fakeDashboard(json(sent), json(sent));
  const result = await client.deliver({ id: "c-1", version: 1, requestId: "r-1", target: { kind: "session", sessionId: "s-1" } });
  expect(result.agentRun).toEqual({ sessionId: "s-1", requestedAt: "2026-11-01T10:02:00Z", state: "sent", target: { label: "claude: site", channel: "cmux" } });
  await client.deliver({ id: "c-1", version: 1, requestId: "r-2", target: client.runTarget({ page, runtime: codex, prompt: "web-page-comments" }) });
  expect(calls[0]?.url.pathname).toBe("/api/v1/comments/c-1/deliver");
  expect(calls[0]?.body).toEqual({ expectedVersion: 1, requestId: "r-1", target: { kind: "session", sessionId: "s-1" } });
  expect(calls[1]?.body).toEqual({ expectedVersion: 1, requestId: "r-2", target: { kind: "run", prompt: "web-page-comments", groupKey: targetUri, runOptions: { spec: { model: "codex-example", effort: "high" }, presets: ["fast"] } } });
  expect(client.runTarget({ page, prompt: "p" })).toEqual({ kind: "run", prompt: "p", groupKey: targetUri });
});

it("groups a run by the page, or by the whole project's shared session when grouping is shared", () => {
  const { client } = fakeDashboard();
  expect(client.runTarget({ page, grouping: "page" })).toMatchObject({ groupKey: targetUri });
  expect(client.runTarget({ page, grouping: "shared" })).toMatchObject({ groupKey: `extension:shared:${scope}` });
  expect(client.runTarget({ page: { ...page, url: "https://example.com/other" }, grouping: "shared" })).toMatchObject({ groupKey: `extension:shared:${scope}` });
});

it("posts a fresh run target with its group key so gavel starts a new session for the group", async () => {
  const { calls, client } = fakeDashboard(json(native({ version: 3 })));
  await client.deliver({ id: "c-1", version: 2, requestId: "r-3", target: freshTarget(client.runTarget({ page, runtime: { spec: { model: "codex-example" } } })) });
  expect(calls[0]?.body).toEqual({ expectedVersion: 2, requestId: "r-3", target: { kind: "run", prompt: "editor-comments", groupKey: targetUri, runOptions: { spec: { model: "codex-example" } }, fresh: true } });
});

it("lists the scope's agent sessions, live and ended managed ones", async () => {
  const ended = { id: "s-2", source: "claude", title: "old run", inCmux: false, status: "ended", managed: true, resumable: true };
  const { calls, client } = fakeDashboard(json([{ id: "s-1", source: "claude", title: "site work", inCmux: true }, ended]));
  expect(await client.agentSessions()).toEqual([{ id: "s-1", source: "claude", title: "site work", inCmux: true }, ended]);
  expect(calls[0]?.url.pathname).toBe("/api/v1/comments/agent-sessions");
  expect(calls[0]?.url.searchParams.get("scope")).toBe(scope);
});

it("delivers several threads to one target in one request and projects the updated threads", async () => {
  const sent = (id: string) => native({ id, version: 2, content: { playground: { page: pageKey(page.url) }, page, agentRun: { sessionId: "s-1", requestedAt: "2026-11-01T10:02:00Z", state: "sent" } } });
  const { calls, client } = fakeDashboard(json([sent("c-1"), sent("c-2")]));
  const threads = [{ id: "c-1", expectedVersion: 1 }, { id: "c-2", expectedVersion: 3 }];
  const updated = await client.deliverBatch({ requestId: "r-9", target: { kind: "session", sessionId: "s-1" }, threads });
  expect(updated.map(comment => [comment.id, comment.agentRun?.sessionId])).toEqual([["c-1", "s-1"], ["c-2", "s-1"]]);
  expect(calls[0]?.method).toBe("POST");
  expect(calls[0]?.url.pathname).toBe("/api/v1/comments/deliver");
  expect(calls[0]?.body).toEqual({ requestId: "r-9", target: { kind: "session", sessionId: "s-1" }, threads });
});

it("surfaces gavel's refusal of a batch with a stale thread", async () => {
  const { client } = fakeDashboard(json({ error: "comment c-2 changed (version 4, expected 3)" }, 409));
  const failed = client.deliverBatch({ requestId: "r-9", target: { kind: "session", sessionId: "s-1" }, threads: [{ id: "c-2", expectedVersion: 3 }] });
  await expect(failed).rejects.toMatchObject({ status: 409, message: expect.stringContaining("comment c-2 changed") });
  await expect(failed).rejects.toBeInstanceOf(GavelError);
});

it("rejects a batch response that is not a list of threads", async () => {
  const { client } = fakeDashboard(json({ ok: true }));
  await expect(client.deliverBatch({ requestId: "r-9", target: { kind: "session", sessionId: "s-1" }, threads: [] })).rejects.toThrow(/invalid batch delivery response/);
});

it("reads and saves the page's comment agent through gavel's settings, never naming a session", async () => {
  const resolved = { settings: { positions: [], layers: [], provenance: {} }, agent: { agent: { enabled: true, prompt: "editor-comments" }, provenance: {}, found: true } };
  const { calls, client } = fakeDashboard(json(resolved), json(resolved), json(resolved));
  expect(await client.commentAgent(page)).toEqual(resolved);
  const agent = commentAgentFor(client.runTarget({ page, runtime: codex, prompt: "web-page-comments" }));
  expect(await client.saveCommentAgent({ page, layer: { kind: "project" }, agent })).toEqual(resolved);
  await client.saveCommentAgent({ page, layer: { kind: "project" }, agent: null });
  expect(`${calls[0]?.url.pathname}?${calls[0]?.url.searchParams.toString()}`).toBe(`/api/v1/comments/agent?${new URLSearchParams({ scope, uri: targetUri }).toString()}`);
  expect(calls[1]).toMatchObject({
    method: "PUT",
    body: {
      target: { scopeUri: scope, uri: targetUri },
      layer: { kind: "project" },
      page: true,
      agent: { enabled: true, prompt: "web-page-comments", presets: ["fast"], spec: { model: "codex-example", effort: "high" }, group: targetUri },
    },
  });
  expect(calls[1]?.url.pathname).toBe("/api/v1/comments/agent");
  expect(JSON.stringify(calls[1]?.body)).not.toContain("sessionId");
  expect(calls[2]?.body).toEqual({ target: { scopeUri: scope, uri: targetUri }, layer: { kind: "project" }, page: true, agent: null });
});

it("derives the settings agent from a run target, carrying only what the run sets", () => {
  expect(commentAgentFor({ kind: "run", prompt: "p" })).toEqual({ enabled: true, prompt: "p" });
  expect(commentAgentFor({ kind: "run", prompt: "p", groupKey: "g", runOptions: { instructions: "be brief", variables: { a: 1 } } })).toEqual({ enabled: true, prompt: "p", group: "g", instructions: "be brief", variables: { a: 1 } });
});

it("reports a gavel host without comment agent settings with the endpoint", async () => {
  const { client } = fakeDashboard(json({ error: "comment agent settings are not available" }, 501));
  const failure = await client.commentAgent(page).catch((cause: unknown) => cause);
  expect(failure).toMatchObject({ status: 501, message: expect.stringContaining("/api/v1/comments/agent") });
});

it("parses server-sent events incrementally across chunk boundaries", () => {
  const events: { id: string; event: string; data: string }[] = [];
  const parse = parseSse(event => events.push(event));
  parse('id: 1\nevent: comment\ndata: {"sequence"');
  expect(events).toEqual([]);
  parse(':1}\n\nid: 2\nevent: comment\r\ndata: a\ndata: b\r\n\r\n: keep-alive\n\n');
  expect(events).toEqual([
    { id: "1", event: "comment", data: '{"sequence":1}' },
    { id: "2", event: "comment", data: "a\nb" },
  ]);
});

it("streams events for the project scope from a cursor", async () => {
  const body = new Response('id: 7\nevent: comment\ndata: {"sequence":7,"commentId":"c-1","kind":"create"}\n\n');
  const { calls, client } = fakeDashboard(body);
  const seen: number[] = [];
  await client.streamEvents({ after: 6, onEvent: event => seen.push(event.sequence) });
  expect(seen).toEqual([7]);
  expect(Object.fromEntries(calls[0]?.url.searchParams ?? [])).toEqual({ scope, after: "6" });
  expect(calls[0]?.url.pathname).toBe("/api/v1/comments/events");
});

it("decodes PNG data URLs and refuses anything else", () => {
  expect([...dataUrlBytes("data:image/png;base64,iVBORw0KGgo=")]).toEqual([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);
  expect(() => dataUrlBytes("data:image/jpeg;base64,AAAA")).toThrow(/PNG/);
});

it("lists the dashboard's projects by name", async () => {
  const fetcher: typeof fetch = async input => {
    expect(String(input)).toBe(`${server}/api/projects`);
    return json([{ name: "alpha", short: "a", dir: "/x", repos: [] }, { name: "my site", short: "ms", dir: "/y", repos: [] }]);
  };
  expect(await listProjects({ server, fetch: fetcher })).toEqual(["alpha", "my site"]);
  await expect(listProjects({ server, fetch: async () => json({ projects: [] }) })).rejects.toThrow(/project list/);
});

const models = [{ id: "claude-example", name: "Claude Example", provider: "anthropic" }];
const runtimes = [
  { family: "claude", provider: "anthropic", catalogPrefix: "anthropic", modes: [{ mode: "api", defaultModel: "claude-api", availability: { state: "available" } }, { mode: "agent", defaultModel: "claude-example", availability: { state: "available" } }] },
  { family: "codex", provider: "openai", catalogPrefix: "openai", modes: [{ mode: "agent", defaultModel: "codex-example", availability: { state: "missing_credentials" } }] },
];

it("loads the model and agent runtime catalogs from the dashboard's chat API", async () => {
  const paths: string[] = [];
  const fetcher: typeof fetch = async input => {
    const { pathname } = new URL(String(input));
    paths.push(pathname);
    return json(pathname === "/api/chat/models" ? models : runtimes);
  };
  const catalog = await fetchRuntimeCatalog({ server, fetch: fetcher });
  expect(paths.sort()).toEqual(["/api/chat/models", "/api/chat/runtimes"]);
  expect(catalog.models).toEqual(models);
  expect(catalog.families.map(family => [family.id, family.modes.map(mode => mode.id)])).toEqual([["claude", ["agent"]]]);
});

it.each([
  { name: "an empty model catalog", models: [], runtimes, error: /model catalog is empty or invalid/ },
  { name: "a runtime catalog that is not a list", models, runtimes: { claude: {} }, error: /runtime catalog is empty or invalid/ },
  { name: "a runtime family without modes", models, runtimes: [{ family: "claude", provider: "anthropic", catalogPrefix: "anthropic" }], error: /runtime catalog is empty or invalid/ },
  { name: "no available agent runtime", models, runtimes: [runtimes[1]], error: /no available agent runtime/ },
])("refuses $name instead of offering built-in runtimes", async ({ models: modelBody, runtimes: runtimeBody, error }) => {
  const fetcher: typeof fetch = async input => json(new URL(String(input)).pathname === "/api/chat/models" ? modelBody : runtimeBody);
  await expect(fetchRuntimeCatalog({ server, fetch: fetcher })).rejects.toThrow(error);
});

it("reports a failing catalog request with its status", async () => {
  await expect(fetchRuntimeCatalog({ server, fetch: async () => json({ error: "chat disabled" }, 503) })).rejects.toThrow(/chat disabled.*HTTP 503/);
});

