import { expect, it } from "vitest";
import { assertActivated, effectiveSettings, hostLaunch, panelTarget, parseGlobalDefaults, sameJson, serverOrigin, siteOverrides, type GlobalDefaults } from "./settings";
import { parseSiteSettings } from "../host/site-settings";

it("allows first-run settings and offline todo capture before selecting a comment project", () => {
  expect(effectiveSettings(parseGlobalDefaults({ server: "http://localhost:9092" }), {})).toEqual({ server: "http://localhost:9092" });
});

it("requires a project before activating the comment panel", () => {
  expect(() => hostLaunch({ server: "http://localhost:9092" }, { panel: "chrome-extension://example/panel.html", todoRelay: "chrome-extension://example/todo-relay.html" })).toThrow(/project/i);
});

it("refuses removed playground configuration explicitly", () => {
  expect(() => parseGlobalDefaults({ backend: "playground", server: "http://localhost:5274" })).toThrow(/playground.*removed/i);
  expect(() => parseSiteSettings(JSON.stringify({ backend: "playground" }))).toThrow(/playground.*removed/i);
});

const server = "http://localhost:9092";
const runtime = { spec: { model: "example-model", effort: "high" }, presets: ["fast"] };
const defaults: GlobalDefaults = { server, project: "acme", runtime };

it("normalizes server addresses and rejects credentials or non-HTTP URLs", () => {
  expect(serverOrigin(` ${server}/path?query=1 `)).toBe(server);
  for (const invalid of ["file:///docs/example.html", "http://user:password@example.com", "localhost:9092"]) {
    expect(() => serverOrigin(invalid)).toThrow();
  }
});

it("merges site settings while keeping sessions and enablement per site", () => {
  expect(effectiveSettings(defaults, {})).toEqual(defaults);
  expect(effectiveSettings(defaults, { project: "docs", session: "session-1", enabled: true })).toEqual({ ...defaults, project: "docs", session: "session-1", enabled: true });
  expect(() => effectiveSettings(defaults, { server: "not a URL" })).toThrow();
});

it("stores differing values and removes overrides once they match the defaults", () => {
  expect(siteOverrides(defaults, defaults)).toBeNull();
  expect(siteOverrides(defaults, { ...defaults, server: `${server}/`, project: "docs", session: "session-1", enabled: true })).toEqual({ project: "docs", session: "session-1", enabled: true });
  expect(siteOverrides(defaults, { ...defaults, runtime: { presets: ["fast"], spec: { effort: "high", model: "example-model" } }, grouping: "page" })).toBeNull();
  expect(siteOverrides(defaults, { ...defaults, grouping: "shared" })).toEqual({ grouping: "shared" });
  expect(siteOverrides({ ...defaults, grouping: "shared" }, { ...defaults, grouping: "page" })).toEqual({ grouping: "page" });
});

it("migrates stored Gavel defaults without dropping runtime or grouping", () => {
  expect(parseGlobalDefaults({ backend: "gavel", server: `${server}/path`, project: "acme", runtime, grouping: "shared" })).toEqual({ ...defaults, grouping: "shared" });
});

it.each([
  { value: "invalid", error: /object/ },
  { value: { server, session: "session-1" }, error: /unknown field "session"/ },
  { value: { server, project: "" }, error: /non-empty string/ },
  { value: { server, runtime: { spec: {} } }, error: /spec.model/ },
  { value: { server, grouping: "all" }, error: /shared or page/ },
  { value: { server, agent: "codex" }, error: /re-pick the model/ },
])("rejects invalid stored defaults $value", ({ value, error }) => {
  expect(() => parseGlobalDefaults(value)).toThrow(error);
});

it("compares JSON settings by value while preserving array order", () => {
  expect(sameJson({ spec: { a: 1, b: [1, { c: 2 }] } }, { spec: { b: [1, { c: 2 }], a: 1 }, presets: undefined })).toBe(true);
  expect(sameJson({ a: [1, 2] }, { a: [2, 1] })).toBe(false);
  expect(sameJson(undefined, undefined)).toBe(true);
  expect(sameJson(undefined, { a: 1 })).toBe(false);
});

const page = "https://example.com";
const frameParams = { server, project: "acme", runtime: JSON.stringify(runtime), grouping: "shared", session: "session-1", parentOrigin: page, bridgeSession: "bridge-1" };
const framed = (params: Record<string, string>, ancestorOrigin: string | undefined) => ({ search: `?${new URLSearchParams(params)}`, ancestorOrigin });
const without = (key: string) => Object.fromEntries(Object.entries(frameParams).filter(([name]) => name !== key));

it("accepts a bundled panel only for its framing document, including file pages", () => {
  expect(panelTarget(framed(frameParams, page))).toEqual({ server, project: "acme", runtime, grouping: "shared", agentSession: "session-1", parentOrigin: page, bridgeSession: "bridge-1" });
  expect(panelTarget(framed({ server, project: "acme", parentOrigin: "null", bridgeSession: "bridge-1" }, "file://"))).toEqual({ server, project: "acme", grouping: "page", parentOrigin: "null", bridgeSession: "bridge-1" });
});

it.each([
  { name: "an unframed panel", options: framed(frameParams, undefined), error: /framed by a page/ },
  { name: "a panel for another page", options: framed(frameParams, "https://other.example"), error: /not the framing page/ },
  { name: "a missing project", options: framed(without("project"), page), error: /project/ },
  { name: "a missing bridge session", options: framed(without("bridgeSession"), page), error: /bridge session/ },
  { name: "an invalid server", options: framed({ ...frameParams, server: "invalid" }, page), error: /./ },
  { name: "an invalid parent origin", options: framed({ ...frameParams, parentOrigin: `${page}/path` }, `${page}/path`), error: /Invalid parent origin/ },
  { name: "a runtime that is not JSON", options: framed({ ...frameParams, runtime: "{invalid" }, page), error: /not valid JSON/ },
  { name: "a runtime without a model", options: framed({ ...frameParams, runtime: JSON.stringify({ spec: {} }) }, page), error: /spec.model/ },
  { name: "an unknown grouping", options: framed({ ...frameParams, grouping: "all" }, page), error: /shared or page/ },
])("refuses $name", ({ options, error }) => expect(() => panelTarget(options)).toThrow(error));

it("launches only the bundled Gavel panel with the site's runtime and session", () => {
  const urls = { panel: "chrome-extension://example/panel.html", todoRelay: "chrome-extension://example/todo-relay.html" };
  expect(hostLaunch({ ...defaults, grouping: "shared", session: "session-1" }, urls)).toEqual({ gavel: { url: urls.panel, todoRelay: urls.todoRelay, server, project: "acme", runtime, grouping: "shared", agentSession: "session-1" } });
});

it("authorizes a frame only for the activated server and project", () => {
  const record = { server, project: "acme" };
  expect(() => assertActivated({ record, request: record })).not.toThrow();
  expect(() => assertActivated({ record: undefined, request: record })).toThrow(/not activated/);
  expect(() => assertActivated({ record, request: { ...record, project: "other" } })).toThrow(/different settings/);
  expect(() => assertActivated({ record, request: { ...record, server: "https://other.example" } })).toThrow(/different settings/);
});
