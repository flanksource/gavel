import { expect, it } from "vitest";
import { parseGrouping, parseRuntime, parseSiteSettings, parseSessionTarget, serializeSiteSettings, withSiteEnabled, withSiteSession, withSiteTarget } from "./site-settings";

const runtime = { spec: { model: "example-model", effort: "high" }, presets: ["fast"] };

it("reads an absent entry as no overrides", () => {
  expect(parseSiteSettings(null)).toEqual({});
});

it("reads every known field", () => {
  expect(parseSiteSettings(JSON.stringify({ backend: "gavel", server: "http://localhost:9092", project: "site", runtime, grouping: "shared", session: "s-1", target: "global", enabled: true }))).toEqual({
    server: "http://localhost:9092", project: "site", runtime, grouping: "shared", session: "s-1", target: "global", enabled: true,
  });
  expect(parseSiteSettings(JSON.stringify({ target: "page" }))).toEqual({ target: "page" });
  expect(parseSiteSettings(JSON.stringify({ session: "s-1", target: "auto" }))).toEqual({ session: "s-1", target: "auto" });
});

it.each([
  { name: "malformed JSON", raw: "{nope", error: /not valid JSON/ },
  { name: "an array", raw: "[]", error: /must be an object/ },
  { name: "an unknown field", raw: JSON.stringify({ colour: "red" }), error: /unknown field "colour"/ },
  { name: "the removed agent field", raw: JSON.stringify({ agent: "codex" }), error: /"agent" setting was replaced.*re-pick the model/s },
  { name: "a removed backend", raw: JSON.stringify({ backend: "playground" }), error: /playground backend was removed/ },
  { name: "a non-string project", raw: JSON.stringify({ project: 3 }), error: /project must be a non-empty string/ },
  { name: "an empty session", raw: JSON.stringify({ session: "" }), error: /session must be a non-empty string/ },
  { name: "enabled: false, which is stored as no key", raw: JSON.stringify({ enabled: false }), error: /enabled must be true/ },
  { name: "a non-boolean enabled", raw: JSON.stringify({ enabled: "yes" }), error: /enabled must be true/ },
  { name: "an unknown grouping", raw: JSON.stringify({ grouping: "global" }), error: /grouping must be shared or page/ },
  { name: "an unknown target", raw: JSON.stringify({ target: "everywhere" }), error: /target must be global, page or auto, got "everywhere"/ },
  { name: "a non-string target", raw: JSON.stringify({ target: true }), error: /target must be global, page or auto, got true/ },
  { name: "a runtime that is a string", raw: JSON.stringify({ runtime: "codex" }), error: /runtime must be an object/ },
  { name: "a runtime without a model", raw: JSON.stringify({ runtime: { spec: {} } }), error: /runtime spec.model must be a non-empty string/ },
  { name: "a runtime with an unknown field", raw: JSON.stringify({ runtime: { spec: { model: "m" }, model: "m" } }), error: /runtime has unknown field "model"/ },
  { name: "runtime presets that are not strings", raw: JSON.stringify({ runtime: { spec: { model: "m" }, presets: [1] } }), error: /runtime presets must be a list of non-empty strings/ },
])("rejects $name instead of ignoring it", ({ raw, error }) => {
  expect(() => parseSiteSettings(raw)).toThrow(error);
});

it("validates a runtime and a grouping wherever they are stored, naming where", () => {
  expect(parseRuntime(runtime, "Global default")).toEqual(runtime);
  expect(parseGrouping("page", "Global default")).toBe("page");
  expect(() => parseRuntime({ spec: {} }, "Global default")).toThrow(/Global default runtime spec.model/);
  expect(() => parseGrouping(1, "Global default")).toThrow(/Global default grouping must be shared or page, got 1/);
  expect(parseSessionTarget("global", "Grabber settings")).toBe("global");
  expect(parseSessionTarget("auto", "Grabber settings")).toBe("auto");
  expect(() => parseSessionTarget("site", "Grabber settings")).toThrow(/Grabber settings target must be global, page or auto, got "site"/);
});

it("removes the entry when nothing overrides the defaults", () => {
  expect(serializeSiteSettings({})).toBeNull();
  expect(serializeSiteSettings({ project: "site" })).toBe('{"project":"site"}');
  expect(serializeSiteSettings({ enabled: true })).toBe('{"enabled":true}');
  expect(JSON.parse(serializeSiteSettings({ runtime, grouping: "shared" }) ?? "")).toEqual({ runtime, grouping: "shared" });
});

it("changes only enabled in the stored entry and removes the entry once empty", () => {
  expect(JSON.parse(withSiteEnabled(JSON.stringify({ project: "site" }), true) ?? "")).toEqual({ project: "site", enabled: true });
  expect(JSON.parse(withSiteEnabled(null, true) ?? "")).toEqual({ enabled: true });
  expect(withSiteEnabled(JSON.stringify({ project: "site", enabled: true }), false)).toBe('{"project":"site"}');
  expect(withSiteEnabled(JSON.stringify({ enabled: true }), false)).toBeNull();
  expect(withSiteEnabled(null, false)).toBeNull();
});

it("changes only the target of the stored entry: global and auto are stored, page (the default) removes the key and the entry once empty", () => {
  const stored = JSON.stringify({ project: "site", session: "s-2" });
  expect(JSON.parse(withSiteTarget(stored, "global") ?? "")).toEqual({ project: "site", session: "s-2", target: "global" });
  expect(JSON.parse(withSiteTarget(stored, "auto") ?? "")).toEqual({ project: "site", session: "s-2", target: "auto" });
  expect(JSON.parse(withSiteTarget(null, "global") ?? "")).toEqual({ target: "global" });
  expect(JSON.parse(withSiteTarget(JSON.stringify({ target: "global" }), "auto") ?? "")).toEqual({ target: "auto" });
  expect(withSiteTarget(JSON.stringify({ project: "site", target: "global" }), "page")).toBe('{"project":"site"}');
  expect(withSiteTarget(JSON.stringify({ project: "site", target: "auto" }), "page")).toBe('{"project":"site"}');
  expect(withSiteTarget(JSON.stringify({ target: "global" }), "page")).toBeNull();
  expect(() => withSiteTarget(null, "nowhere" as never)).toThrow(/target must be global, page or auto/);
});

it("changes only the session of the stored entry and removes the entry once empty", () => {
  const stored = JSON.stringify({ project: "site" });
  expect(JSON.parse(withSiteSession(stored, "s-2") ?? "")).toEqual({ project: "site", session: "s-2" });
  expect(JSON.parse(withSiteSession(null, "s-2") ?? "")).toEqual({ session: "s-2" });
  expect(withSiteSession(JSON.stringify({ project: "site", session: "s-2" }), null)).toBe('{"project":"site"}');
  expect(withSiteSession(JSON.stringify({ session: "s-2" }), null)).toBeNull();
  expect(() => withSiteSession(null, "")).toThrow(/session/);
});
