import type { AgentSessionGrouping, AgentSessionRuntime } from "../contracts/agent-session-store";

/** The page's `localStorage` key holding the grabber settings that override the extension's global defaults for that site. */
export const SITE_SETTINGS_KEY = "flanksource.grabber";

export const GROUPINGS: readonly AgentSessionGrouping[] = ["shared", "page"];

/**
 * Only the fields that differ from the global defaults, plus the site's agent
 * session and whether the extension is enabled on this site. `enabled` is stored
 * only as `true`: switching it off removes the key.
 */
export type SiteSettings = { server?: string; project?: string; runtime?: AgentSessionRuntime; grouping?: AgentSessionGrouping; session?: string; enabled?: true };

const TEXT_FIELDS = ["server", "project", "session"] as const;
const FIELDS = [...TEXT_FIELDS, "runtime", "grouping", "enabled"];

/** The old string `agent` setting ("claude" | "codex") named no model, so it cannot be carried over; it is refused rather than guessed at. */
export function removedAgentError(where: string): Error {
  return new Error(`${where} still has the removed "agent" field: the "agent" setting was replaced by "runtime"; discard it in the extension popup and re-pick the model there.`);
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === "object" && !Array.isArray(value);
}

export function parseGrouping(value: unknown, label: string): AgentSessionGrouping {
  if (!GROUPINGS.includes(value as AgentSessionGrouping)) throw new Error(`${label} grouping must be ${GROUPINGS.join(" or ")}, got ${JSON.stringify(value)}`);
  return value as AgentSessionGrouping;
}

/** The runtime new agent sessions start with: a spec naming a model, and optional presets. Anything else is a corrupt entry. */
export function parseRuntime(value: unknown, label: string): AgentSessionRuntime {
  if (!isRecord(value)) throw new Error(`${label} runtime must be an object, got ${JSON.stringify(value)}`);
  for (const key of Object.keys(value)) if (key !== "spec" && key !== "presets") throw new Error(`${label} runtime has unknown field "${key}"`);
  const spec = value["spec"];
  if (!isRecord(spec) || typeof spec["model"] !== "string" || spec["model"] === "")
    throw new Error(`${label} runtime spec.model must be a non-empty string, got ${JSON.stringify(spec)}`);
  const presets = value["presets"];
  if (presets === undefined) return { spec: spec as AgentSessionRuntime["spec"] };
  if (!Array.isArray(presets) || !presets.every(entry => typeof entry === "string" && entry !== ""))
    throw new Error(`${label} runtime presets must be a list of non-empty strings, got ${JSON.stringify(presets)}`);
  return { spec: spec as AgentSessionRuntime["spec"], presets: presets as string[] };
}

export function parseSiteSettings(raw: string | null): SiteSettings {
  if (raw === null) return {};
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch (cause) {
    throw new Error(`Grabber settings in ${SITE_SETTINGS_KEY} are not valid JSON: ${String(cause)}`);
  }
  if (!isRecord(parsed)) throw new Error(`Grabber settings in ${SITE_SETTINGS_KEY} must be an object`);
  if ("agent" in parsed) throw removedAgentError(`Grabber settings in ${SITE_SETTINGS_KEY}`);
  for (const key of Object.keys(parsed)) if (key !== "backend" && !FIELDS.includes(key)) throw new Error(`Grabber settings in ${SITE_SETTINGS_KEY} have unknown field "${key}"`);
  const site: SiteSettings = {};
  if (parsed["backend"] !== undefined) {
    if (parsed["backend"] !== "gavel") throw new Error("The playground backend was removed; discard this site's settings and configure Gavel");
  }
  for (const field of TEXT_FIELDS) {
    const text = parsed[field];
    if (text === undefined) continue;
    if (typeof text !== "string" || text === "") throw new Error(`Grabber settings ${field} must be a non-empty string, got ${JSON.stringify(text)}`);
    site[field] = text;
  }
  if (parsed["runtime"] !== undefined) site.runtime = parseRuntime(parsed["runtime"], "Grabber settings");
  if (parsed["grouping"] !== undefined) site.grouping = parseGrouping(parsed["grouping"], "Grabber settings");
  if (parsed["enabled"] !== undefined) {
    if (parsed["enabled"] !== true) throw new Error(`Grabber settings enabled must be true (remove the key to disable), got ${JSON.stringify(parsed["enabled"])}`);
    site.enabled = true;
  }
  return site;
}

/** The text to store, or `null` when the entry should be removed because nothing overrides the defaults. */
export function serializeSiteSettings(site: SiteSettings): string | null {
  const entries = FIELDS.flatMap(field => {
    const value = site[field as keyof SiteSettings];
    return value === undefined ? [] : [[field, value] as const];
  });
  return entries.length === 0 ? null : JSON.stringify(Object.fromEntries(entries));
}

/** The stored entry with only its session changed; `null` clears the session. */
export function withSiteSession(raw: string | null, session: string | null): string | null {
  const site = parseSiteSettings(raw);
  if (session === null) delete site.session;
  else if (session === "") throw new Error("Grabber session must be a non-empty string");
  else site.session = session;
  return serializeSiteSettings(site);
}

/** The stored entry with only `enabled` changed; disabling removes the key, and the entry once empty. */
export function withSiteEnabled(raw: string | null, enabled: boolean): string | null {
  const site = parseSiteSettings(raw);
  if (enabled) site.enabled = true;
  else delete site.enabled;
  return serializeSiteSettings(site);
}
