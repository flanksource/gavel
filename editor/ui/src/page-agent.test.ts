import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { agentLabel, loadViewerChoice, originOf, routeBodyOf, runOf, saveViewerChoice, type CommentRoute, type PageAgent } from "./page-agent";

const scope = "gavel://project/acme";
const folderPrefix = `${scope}/files/slides`;
const pagePrefix = `${folderPrefix}/19-security.mdx`;
const route = (targetPrefix: string, overrides: Partial<CommentRoute> = {}): CommentRoute => ({
  id: `route-${targetPrefix.length}`,
  scopeUri: scope,
  targetPrefix,
  workflowId: "human-review",
  target: { kind: "run", prompt: "slides-review" },
  enabled: true,
  ...overrides,
});
const agentWith = (effective: CommentRoute | null): PageAgent => ({
  scope,
  pagePrefix,
  folderPrefix,
  configPath: "/repo/slides/.gavel.yaml",
  page: null,
  folder: null,
  effective,
});

describe("routeBodyOf", () => {
  it("targets the page or its folder, keeping the prompt apart from the other run choices", () => {
    const run = { prompt: "slides-review", presets: ["fast"], instructions: "Keep the tables" };
    expect([
      routeBodyOf({ agent: agentWith(null), scope: "page", enabled: true, assign: false, run }),
      routeBodyOf({ agent: agentWith(null), scope: "folder", enabled: false, assign: true, run: { prompt: "editor-comments" } }),
    ]).toEqual([
      { scopeUri: scope, targetPrefix: pagePrefix, workflowId: "human-review", target: { kind: "run", prompt: "slides-review", runOptions: { presets: ["fast"], instructions: "Keep the tables" } }, enabled: true, assign: false },
      { scopeUri: scope, targetPrefix: folderPrefix, workflowId: "human-review", target: { kind: "run", prompt: "editor-comments" }, enabled: false, assign: true },
    ]);
  });

  it("refuses a run without a prompt", () => {
    expect(() => routeBodyOf({ agent: agentWith(null), scope: "page", enabled: true, assign: false, run: {} })).toThrow(/needs a prompt/);
  });

  it("round-trips a saved route back into the run it starts", () => {
    const saved = route(pagePrefix, { target: { kind: "run", prompt: "slides-review", runOptions: { presets: ["fast"] } } });
    expect(runOf(saved)).toEqual({ prompt: "slides-review", presets: ["fast"] });
  });
});

describe("agentLabel", () => {
  it.each([
    ["no route", null, "none", "No page agent"],
    ["the page's own route", route(pagePrefix, { assign: true }), "page", "slides-review · auto + assign (page)"],
    ["the folder's route", route(folderPrefix), "folder", "slides-review · auto (folder)"],
    ["a disabled project-wide route", route(`${scope}/files`, { enabled: false }), "inherited", "Page agent off (inherited)"],
  ])("labels %s by its origin", (_case, effective, origin, label) => {
    expect([originOf(agentWith(effective)), agentLabel(agentWith(effective))]).toEqual([origin, label]);
  });
});

describe("viewer choice", () => {
  beforeEach(() => {
    const stored = new Map<string, string>();
    vi.stubGlobal("window", {
      localStorage: { getItem: (key: string) => stored.get(key) ?? null, setItem: (key: string, value: string) => void stored.set(key, value) },
    });
  });
  afterEach(() => vi.unstubAllGlobals());

  it("defaults to the configured run, and remembers a live session per page", () => {
    const before = loadViewerChoice(pagePrefix);
    saveViewerChoice(pagePrefix, { kind: "session", sessionId: "s-1" });
    expect([before, loadViewerChoice(pagePrefix), loadViewerChoice(folderPrefix)]).toEqual([{ kind: "run" }, { kind: "session", sessionId: "s-1" }, { kind: "run" }]);
  });

  it("ignores a stored value that is not a choice", () => {
    window.localStorage.setItem(`gavel-edit:page-agent:${pagePrefix}`, JSON.stringify({ kind: "session" }));
    expect(loadViewerChoice(pagePrefix)).toEqual({ kind: "run" });
  });
});
