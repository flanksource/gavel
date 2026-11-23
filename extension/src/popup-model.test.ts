import { describe, expect, it } from "vitest";
import { editField, isSiteOnly, promoteField, resetField, type Stored } from "./popup-model";

const server = "http://localhost:9092";
const otherServer = "https://other.example.com";

const runtimeA = { spec: { model: "claude-example", effort: "high" }, presets: ["fast"] };
const runtimeB = { spec: { model: "codex-example" } };
const untouched = { session: "page-1", target: "auto" as const, enabled: true as const };
const globalSession = { session: "global-1" };

const stored = (global: Stored["global"], site: Stored["site"] = {}): Stored => ({ global, site });

describe("isSiteOnly", () => {
  it.each([
    { field: "server" as const, site: { server: otherServer } },
    { field: "project" as const, site: { project: "docs" } },
    { field: "runtime" as const, site: { runtime: runtimeA } },
    { field: "grouping" as const, site: { grouping: "shared" as const } },
  ])("is true for $field exactly when the site stores it", ({ field, site }) => {
    expect(isSiteOnly(stored({ server }), field)).toBe(false);
    expect(isSiteOnly(stored({ server }, site), field)).toBe(true);
    expect(isSiteOnly(stored({ server }, { session: "page-1", target: "global", enabled: true }), field)).toBe(false);
  });
});

describe("editField", () => {
  it("never mutates its input", () => {
    const input = stored({ server, project: "acme", runtime: runtimeA }, { project: "docs" });
    const snapshot = structuredClone(input);
    editField(input, "runtime", runtimeB);
    editField(input, "server", otherServer);
    expect(input).toEqual(snapshot);
  });

  it.each([
    { name: "runtime", field: "runtime" as const, global: { server, runtime: runtimeA }, equal: { presets: ["fast"], spec: { effort: "high", model: "claude-example" } }, other: runtimeB, key: "runtime" as const },
    { name: "grouping", field: "grouping" as const, global: { server, grouping: "shared" as const }, equal: "shared" as const, other: "page" as const, key: "grouping" as const },
    { name: "project", field: "project" as const, global: { server, project: "acme" }, equal: "acme", other: "docs", key: "project" as const },
  ])("stores a differing $name as a site override and removes the override when it equals the default", ({ field, global, equal, other, key }) => {
    const base = stored(global, { ...untouched });
    const edited = editField(base, field, other as never);
    expect(edited).toEqual({ global, site: { ...untouched, [key]: other } });
    expect(editField(edited, field, equal as never)).toEqual({ global, site: untouched });
  });

  it("compares grouping against the built-in default when the extension has none", () => {
    const base = stored({ server });
    expect(editField(base, "grouping", "shared")).toEqual({ global: { server }, site: { grouping: "shared" } });
    expect(editField(stored({ server }, { grouping: "shared" }), "grouping", "page")).toEqual({ global: { server }, site: {} });
  });

  it.each([
    { name: "project", field: "project" as const, value: "acme", global: { project: "acme" } },
    { name: "runtime", field: "runtime" as const, value: runtimeA, global: { runtime: runtimeA } },
  ])("makes the first $name ever chosen the default instead of a site override", ({ field, value, global }) => {
    expect(editField(stored({ server }, { ...untouched }), field, value as never)).toEqual({ global: { server, ...global }, site: untouched });
    expect(editField(stored({ server }, { [field]: value }), field, value as never)).toEqual({ global: { server, ...global }, site: {} });
  });

  it("normalises the server and drops the project of the old dashboard when the server changes", () => {
    const base = stored({ server, project: "acme" }, { ...untouched, project: "docs" });
    expect(editField(base, "server", ` ${otherServer}/path?x=1 `)).toEqual({ global: { server, project: "acme" }, site: { ...untouched, server: otherServer } });
  });

  it("removes the server override and its project when the server equals the default", () => {
    const base = stored({ server, project: "acme" }, { ...untouched, server: otherServer, project: "docs" });
    expect(editField(base, "server", `${server}/`)).toEqual({ global: { server, project: "acme" }, site: untouched });
  });

  it("keeps the site's project when the server is edited to the value it already has", () => {
    const base = stored({ server, project: "acme" }, { server: otherServer, project: "docs" });
    expect(editField(base, "server", otherServer)).toEqual(base);
    expect(editField(stored({ server }, { project: "docs" }), "server", server)).toEqual(stored({ server }, { project: "docs" }));
  });

  it("rejects an invalid server instead of storing it", () => {
    for (const invalid of ["localhost:9092", "http://user:pw@example.com", "not a url"]) expect(() => editField(stored({ server }), "server", invalid)).toThrow();
  });

  it("always keeps the project a site value while the site has its own server, even when it equals the default", () => {
    const base = stored({ server, project: "acme" }, { server: otherServer });
    expect(editField(base, "project", "acme")).toEqual({ global: { server, project: "acme" }, site: { server: otherServer, project: "acme" } });
    expect(editField(stored({ server }, { server: otherServer }), "project", "docs")).toEqual({ global: { server }, site: { server: otherServer, project: "docs" } });
  });

  it("refuses an empty project", () => {
    expect(() => editField(stored({ server, project: "acme" }), "project", "")).toThrow(/project/i);
    expect(() => editField(stored({ server }, { server: otherServer }), "project", "")).toThrow(/project/i);
  });

  it("leaves the sessions, target, enabled flag and global session alone", () => {
    const base = stored({ server, project: "acme", ...globalSession }, { ...untouched });
    for (const edited of [editField(base, "runtime", runtimeB), editField(base, "grouping", "shared"), editField(base, "project", "docs"), editField(base, "server", otherServer)]) {
      expect(edited.site).toMatchObject(untouched);
      expect(edited.global.session).toBe("global-1");
    }
  });
});

describe("resetField", () => {
  it("removes only the field's own key", () => {
    const site = { project: "docs", runtime: runtimeA, grouping: "shared" as const, ...untouched };
    const base = stored({ server, project: "acme" }, site);
    expect(resetField(base, "runtime")).toEqual({ global: base.global, site: { ...site, runtime: undefined } });
    expect(resetField(base, "grouping").site).not.toHaveProperty("grouping");
    expect(resetField(base, "project").site).toEqual({ runtime: runtimeA, grouping: "shared", ...untouched });
  });

  it("removes the server together with the project of its dashboard", () => {
    const base = stored({ server }, { server: otherServer, project: "docs", grouping: "shared", ...untouched });
    expect(resetField(base, "server")).toEqual({ global: { server }, site: { grouping: "shared", ...untouched } });
  });

  it("refuses to reset the project of a site's own dashboard: there is no default project to fall back to", () => {
    for (const site of [{ server: otherServer, project: "docs", ...untouched }, { server: otherServer, ...untouched }])
      expect(() => resetField(stored({ server, project: "acme" }, site), "project")).toThrow(/project.*dashboard/i);
  });

  it("returns an equal state for a field the site does not override, without mutating", () => {
    const base = stored({ server, project: "acme" }, { ...untouched });
    for (const field of ["server", "project", "runtime", "grouping"] as const) expect(resetField(base, field)).toEqual(base);
  });
});

describe("promoteField", () => {
  it("moves the site's server and project into the defaults", () => {
    const base = stored({ server, project: "acme", ...globalSession }, { server: otherServer, project: "docs", ...untouched });
    expect(promoteField(base, "server")).toEqual({ global: { server: otherServer, project: "docs", ...globalSession }, site: untouched });
  });

  it("drops the default project when the site has a server but no project, since it belonged to the old dashboard", () => {
    const base = stored({ server, project: "acme" }, { server: otherServer, ...untouched });
    const promoted = promoteField(base, "server");
    expect(promoted).toEqual({ global: { server: otherServer }, site: untouched });
    expect(promoted.global).not.toHaveProperty("project");
  });

  it("refuses to promote the project alone while the site has its own server: the dashboard is what gets promoted", () => {
    for (const site of [{ server: otherServer, project: "docs", ...untouched }, { server: otherServer, ...untouched }])
      expect(() => promoteField(stored({ server, project: "acme" }, site), "project")).toThrow(/project.*dashboard/i);
  });

  it("moves a project of the shared dashboard into the defaults", () => {
    const base = stored({ server, project: "acme" }, { project: "docs", ...untouched });
    expect(promoteField(base, "project")).toEqual({ global: { server, project: "docs" }, site: untouched });
  });

  it.each([
    { field: "runtime" as const, site: { runtime: runtimeB }, global: { runtime: runtimeA }, promoted: { runtime: runtimeB } },
    { field: "grouping" as const, site: { grouping: "shared" as const }, global: { grouping: "page" as const }, promoted: { grouping: "shared" as const } },
  ])("moves the site's $field into the defaults", ({ field, site, global, promoted }) => {
    const base = stored({ server, ...global }, { ...site, ...untouched });
    expect(promoteField(base, field)).toEqual({ global: { server, ...promoted }, site: untouched });
  });

  it("throws when promoting a field the site does not override", () => {
    const base = stored({ server, project: "acme" }, { ...untouched });
    for (const field of ["server", "project", "runtime", "grouping"] as const) expect(() => promoteField(base, field)).toThrow(new RegExp(field));
  });

  it("never mutates its input", () => {
    const input = stored({ server, project: "acme" }, { server: otherServer, project: "docs", ...untouched });
    const snapshot = structuredClone(input);
    promoteField(input, "server");
    expect(input).toEqual(snapshot);
  });
});
