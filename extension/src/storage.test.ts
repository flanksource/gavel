import { afterEach, expect, it, vi } from "vitest";
import { SITE_SETTINGS_KEY } from "../host/site-settings";
import { readGlobalDefaults, readSiteSettings, writeGlobalSession, writeSiteSettings } from "./storage";

function stubLocalStorage(initial: Record<string, unknown>) {
  const store = { ...initial };
  vi.stubGlobal("chrome", { storage: { local: {
    get: async (key: string) => (key in store ? { [key]: store[key] } : {}),
    set: async (values: Record<string, unknown>) => { Object.assign(store, values); },
  } } });
  return store;
}

afterEach(() => {
  localStorage.removeItem(SITE_SETTINGS_KEY);
  vi.unstubAllGlobals();
});

it("retains a newer session selection when separately saving other site settings", async () => {
  vi.stubGlobal("chrome", { scripting: {
    executeScript: async ({ func, args }: { func: (...args: unknown[]) => unknown; args: unknown[] }) =>
      [{ result: func(...args) }],
  } });
  await writeSiteSettings(7, { enabled: true, session: "sidebar-choice" });
  await writeSiteSettings(7, { enabled: true, grouping: "shared", session: "stale-popup-choice" }, { preserveSession: true });
  expect(await readSiteSettings(7)).toEqual({ enabled: true, grouping: "shared", session: "sidebar-choice" });
});

it("stores the global session beside the other global defaults and clears it again", async () => {
  const defaults = { server: "http://localhost:9092", project: "acme", grouping: "shared" };
  const store = stubLocalStorage({ globalDefaults: defaults });
  expect(await writeGlobalSession("global-1")).toEqual({ ...defaults, session: "global-1" });
  expect(await readGlobalDefaults()).toEqual({ ...defaults, session: "global-1" });
  expect(await writeGlobalSession("global-2")).toEqual({ ...defaults, session: "global-2" });
  expect(await writeGlobalSession(null)).toEqual(defaults);
  expect(store["globalDefaults"]).toEqual(defaults);
});

it("refuses an empty global session instead of storing it", async () => {
  const store = stubLocalStorage({ globalDefaults: { server: "http://localhost:9092" } });
  await expect(writeGlobalSession("")).rejects.toThrow(/non-empty/);
  expect(store["globalDefaults"]).toEqual({ server: "http://localhost:9092" });
});
