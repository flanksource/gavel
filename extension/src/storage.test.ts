import { afterEach, expect, it, vi } from "vitest";
import { SITE_SETTINGS_KEY } from "../host/site-settings";
import { readSiteSettings, writeSiteSettings } from "./storage";

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
