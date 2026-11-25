import { expect, it } from "vitest";
import { bookmarkletFrame, bookmarkletParentOrigin, matchesBookmarkletParent } from "./origin";

const extension = "chrome-extension://abcdefghijklmnopabcdefghijklmnop";
const runtime = { spec: { model: "example-model" }, presets: ["fast"] };

it("frames the extension's bundled gavel panel directly, naming the bridge session apart from the agent session", () => {
  const result = bookmarkletFrame({
    documentUrl: "https://example.com/docs?a=1",
    session: "bridge-1",
    panel: { url: `${extension}/panel.html`, todoRelay: `${extension}/todo-relay.html`, server: "http://localhost:9092", project: "my site", runtime, grouping: "shared", agentSession: "global-1", sessionTarget: "global", pageSession: "agent-7", globalSession: "global-1" },
  });
  expect(result.origin).toBe(extension);
  const url = new URL(result.src);
  expect(`${url.protocol}//${url.host}${url.pathname}`).toBe(`${extension}/panel.html`);
  expect(Object.fromEntries(url.searchParams)).toEqual({
    server: "http://localhost:9092",
    project: "my site",
    runtime: JSON.stringify(runtime),
    grouping: "shared",
    session: "global-1",
    sessionTarget: "global",
    pageSession: "agent-7",
    globalSession: "global-1",
    parentOrigin: "https://example.com",
    bridgeSession: "bridge-1",
  });
});

it("omits the optional runtime, grouping and session from the gavel panel URL and refuses a non-extension panel", () => {
  const panel = { url: `${extension}/panel.html`, todoRelay: `${extension}/todo-relay.html`, server: "http://localhost:9092", project: "site" };
  const url = new URL(bookmarkletFrame({ documentUrl: "file:///docs/a.html", session: "b", panel }).src);
  expect([...url.searchParams.keys()].sort()).toEqual(["bridgeSession", "parentOrigin", "project", "server"]);
  expect(url.searchParams.get("parentOrigin")).toBe("null");
  expect(() => bookmarkletFrame({ documentUrl: "https://example.com/", session: "b", panel: { ...panel, url: "https://example.com/panel.html" } })).toThrow(/chrome-extension/);
});

it("normalizes file documents and accepts only their opaque origin representations", () => {
  expect(bookmarkletParentOrigin("file:///documents/report.html")).toBe("null");
  expect(bookmarkletParentOrigin("https://example.com/report")).toBe("https://example.com");
  expect(matchesBookmarkletParent({ expected: "null", received: "file://" })).toBe(true);
  expect(matchesBookmarkletParent({ expected: "null", received: "null" })).toBe(true);
  expect(matchesBookmarkletParent({ expected: "null", received: "https://example.com" })).toBe(false);
  expect(matchesBookmarkletParent({ expected: "https://example.com", received: "null" })).toBe(false);
});
