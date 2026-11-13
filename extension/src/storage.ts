import { SITE_SETTINGS_KEY, parseSiteSettings, serializeSiteSettings, type SiteSettings } from "../host/site-settings";
import { DEFAULT_GAVEL, parseGlobalDefaults, type GlobalDefaults } from "./settings";

const GLOBAL_KEY = "globalDefaults";
const SEEDED: GlobalDefaults = { server: DEFAULT_GAVEL };

/** The extension-wide defaults. The first run uses the local Gavel dashboard. */
export async function readGlobalDefaults(): Promise<GlobalDefaults> {
  const stored = (await chrome.storage.local.get(GLOBAL_KEY))[GLOBAL_KEY];
  if (stored !== undefined) return parseGlobalDefaults(stored);
  await chrome.storage.local.set({ [GLOBAL_KEY]: SEEDED });
  return SEEDED;
}

/** Validates and stores global defaults, returning what was stored. */
export async function writeGlobalDefaults(defaults: GlobalDefaults): Promise<GlobalDefaults> {
  const checked = parseGlobalDefaults(defaults);
  await chrome.storage.local.set({ [GLOBAL_KEY]: checked });
  return checked;
}

/** Forgets the stored global defaults, e.g. a corrupt or outdated entry; the next read seeds them again. */
export async function discardGlobalDefaults(): Promise<void> {
  await chrome.storage.local.remove(GLOBAL_KEY);
}

type PageResult = { ok: true; value: string | null } | { ok: false; error: string };

// The two functions below run inside the page (ISOLATED world, which shares the
// page origin's localStorage), so they must not reference anything outside
// themselves. Storage can throw (blocked, sandboxed, file:// without access).
function readPageEntry(key: string): PageResult {
  try {
    return { ok: true, value: window.localStorage.getItem(key) };
  } catch (cause) {
    return { ok: false, error: cause instanceof Error ? cause.message : String(cause) };
  }
}

function writePageEntry(key: string, value: string | null): PageResult {
  try {
    if (value === null) window.localStorage.removeItem(key);
    else window.localStorage.setItem(key, value);
    return { ok: true, value: null };
  } catch (cause) {
    return { ok: false, error: cause instanceof Error ? cause.message : String(cause) };
  }
}

function answer(tabId: number, results: { result?: PageResult }[]): string | null {
  const first = results[0]?.result;
  if (!first) throw new Error(`The page in tab ${tabId} did not answer; reload it and try again`);
  if (!first.ok) throw new Error(`This page's storage is not available to the grabber: ${first.error}`);
  return first.value;
}

/** The site's overrides, from the page's `localStorage`. */
export async function readSiteSettings(tabId: number): Promise<SiteSettings> {
  const results = await chrome.scripting.executeScript({ target: { tabId }, world: "ISOLATED", func: readPageEntry, args: [SITE_SETTINGS_KEY] });
  return parseSiteSettings(answer(tabId, results));
}

/** Stores the site's overrides in the page's `localStorage`; `null` removes the entry. */
export async function writeSiteSettings(tabId: number, site: SiteSettings | null): Promise<void> {
  const results = await chrome.scripting.executeScript({ target: { tabId }, world: "ISOLATED", func: writePageEntry, args: [SITE_SETTINGS_KEY, site === null ? null : serializeSiteSettings(site)] });
  answer(tabId, results);
}
