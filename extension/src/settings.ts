import type { AgentSessionGrouping, AgentSessionRuntime } from "../contracts/agent-session-store";
import { validateParentOrigin } from "../host/bridge-parse";
import { matchesBookmarkletParent, type GavelPanelFrame } from "../host/origin";
import { parseGrouping, parseRuntime, removedAgentError, type SiteSettings } from "../host/site-settings";

export const DEFAULT_GAVEL = "http://localhost:9092";
/** How agent sessions are grouped when neither the site nor the global defaults say. */
export const DEFAULT_GROUPING: AgentSessionGrouping = "page";

/** The extension-wide defaults, in `chrome.storage.local`. A site overrides single fields in its own `localStorage`. */
export type GlobalDefaults = { server: string; project?: string; runtime?: AgentSessionRuntime; grouping?: AgentSessionGrouping };
/** What a tab's grabber runs with: global defaults with the site's entry on top, validated. */
export type EffectiveSettings = GlobalDefaults & { session?: string; enabled?: true };

function canonical(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(canonical);
  if (value && typeof value === "object") return Object.fromEntries(Object.entries(value).sort(([left], [right]) => (left < right ? -1 : 1)).map(([key, entry]) => [key, canonical(entry)]));
  return value;
}

/** Deep equality of JSON-shaped values: key order does not matter, `undefined` members are absent. */
export function sameJson(left: unknown, right: unknown): boolean {
  return JSON.stringify(canonical(left)) === JSON.stringify(canonical(right));
}

/** Validates a Gavel dashboard address and returns its origin. */
export function serverOrigin(input: string): string {
  const url = new URL(input.trim());
  if (!["http:", "https:"].includes(url.protocol) || url.username || url.password)
    throw new Error(`Server address must be an HTTP(S) URL without credentials, got ${input}`);
  return url.origin;
}

/** Settings may omit the comment project until the panel is activated. */
function usable(settings: EffectiveSettings): EffectiveSettings {
  const server = serverOrigin(settings.server);
  return { ...settings, server };
}

export function effectiveSettings(global: GlobalDefaults, site: SiteSettings): EffectiveSettings {
  const merged: EffectiveSettings = { server: site.server ?? global.server };
  const project = site.project ?? global.project;
  const runtime = site.runtime ?? global.runtime;
  const grouping = site.grouping ?? global.grouping;
  if (project !== undefined) merged.project = project;
  if (runtime !== undefined) merged.runtime = runtime;
  if (grouping !== undefined) merged.grouping = grouping;
  if (site.session !== undefined) merged.session = site.session;
  if (site.enabled) merged.enabled = true;
  return usable(merged);
}

/**
 * The site entry that makes `chosen` effective: only fields that differ from
 * the global defaults, and the session and enabled flag whenever set. `null` means the
 * site follows the defaults entirely, so its entry should be removed.
 */
export function siteOverrides(global: GlobalDefaults, chosen: EffectiveSettings): SiteSettings | null {
  const settings = usable(chosen);
  const defaults = usable(global);
  const site: SiteSettings = {};
  if (settings.server !== defaults.server) site.server = settings.server;
  if (settings.project !== undefined && settings.project !== defaults.project) site.project = settings.project;
  if (settings.runtime !== undefined && !sameJson(settings.runtime, defaults.runtime)) site.runtime = settings.runtime;
  if (settings.grouping !== undefined && settings.grouping !== (defaults.grouping ?? DEFAULT_GROUPING)) site.grouping = settings.grouping;
  if (settings.session !== undefined) site.session = settings.session;
  if (settings.enabled) site.enabled = true;
  return Object.keys(site).length === 0 ? null : site;
}

/** Reads stored global defaults; a value that is not exactly this shape is a corrupt store, not something to default over. */
export function parseGlobalDefaults(stored: unknown): GlobalDefaults {
  if (!stored || typeof stored !== "object" || Array.isArray(stored)) throw new Error(`Stored global defaults must be an object, got ${JSON.stringify(stored)}`);
  const value = stored as Record<string, unknown>;
  if ("agent" in value) throw removedAgentError("Stored global defaults");
  for (const key of Object.keys(value))
    if (!["backend", "server", "project", "runtime", "grouping"].includes(key)) throw new Error(`Stored global defaults have unknown field "${key}"`);
  if (value["backend"] !== undefined && value["backend"] !== "gavel") throw new Error("The playground backend was removed; discard these defaults and configure Gavel");
  if (typeof value["server"] !== "string") throw new Error(`Global default server must be a string, got ${JSON.stringify(value["server"])}`);
  const project = value["project"];
  if (project !== undefined && (typeof project !== "string" || project === "")) throw new Error(`Global default project must be a non-empty string, got ${JSON.stringify(project)}`);
  const runtime = value["runtime"] === undefined ? undefined : parseRuntime(value["runtime"], "Global default");
  const grouping = value["grouping"] === undefined ? undefined : parseGrouping(value["grouping"], "Global default");
  const checked = usable({ server: value["server"], ...(project ? { project } : {}) });
  return { server: checked.server, ...(project ? { project } : {}), ...(runtime ? { runtime } : {}), ...(grouping ? { grouping } : {}) };
}

/** The bundled panel and TODO form the injected host frames. */
export type HostLaunch = { gavel: GavelPanelFrame };

export function hostLaunch(settings: EffectiveSettings, urls: { panel: string; todoRelay: string }): HostLaunch {
  if (!settings.project) throw new Error("The gavel backend needs a project");
  return {
    gavel: {
      url: urls.panel,
      todoRelay: urls.todoRelay,
      server: settings.server,
      project: settings.project,
      ...(settings.runtime ? { runtime: settings.runtime } : {}),
      ...(settings.grouping ? { grouping: settings.grouping } : {}),
      ...(settings.session ? { agentSession: settings.session } : {}),
    },
  };
}

/** What the background recorded when the user activated the grabber on a tab. */
export type Activation = { server: string; project?: string };

/** An extension frame may only act for the settings the user activated its tab's grabber with. */
export function assertActivated(options: { record: Activation | undefined; request: Activation }) {
  const { record, request } = options;
  if (!record) throw new Error("The grabber was not activated on this tab");
  if (record.server !== request.server || record.project !== request.project)
    throw new Error("The grabber was activated with different settings than this frame requested");
}

/** What the bundled gavel panel runs with, read from the frame URL its host built. */
export type PanelTarget = { server: string; project: string; runtime?: AgentSessionRuntime; grouping: AgentSessionGrouping; agentSession?: string; parentOrigin: string; bridgeSession: string };

/**
 * Like the relay, the panel is web-accessible: it only serves the page that
 * actually frames it, and only with a complete, valid configuration.
 */
export function panelTarget(options: { search: string; ancestorOrigin: string | undefined }): PanelTarget {
  if (!options.ancestorOrigin) throw new Error("Grabber panel must be framed by a page");
  const params = new URLSearchParams(options.search);
  const parentOrigin = params.get("parentOrigin");
  if (!parentOrigin) throw new Error("Grabber panel requires the origin of the page that frames it");
  validateParentOrigin(parentOrigin);
  if (!matchesBookmarkletParent({ expected: parentOrigin, received: options.ancestorOrigin }))
    throw new Error(`Grabber panel belongs to ${parentOrigin}, not the framing page ${options.ancestorOrigin}`);
  const server = params.get("server");
  if (!server) throw new Error("Grabber panel requires a gavel server");
  const project = params.get("project");
  if (!project) throw new Error("Grabber panel requires a gavel project");
  const bridgeSession = params.get("bridgeSession");
  if (!bridgeSession) throw new Error("Grabber panel requires a bridge session");
  const agentSession = params.get("session");
  const encodedRuntime = params.get("runtime");
  let runtime: AgentSessionRuntime | undefined;
  if (encodedRuntime) {
    let decoded: unknown;
    try {
      decoded = JSON.parse(encodedRuntime);
    } catch (cause) {
      throw new Error(`Grabber panel runtime is not valid JSON: ${String(cause)}`);
    }
    runtime = parseRuntime(decoded, "Grabber panel");
  }
  const encodedGrouping = params.get("grouping");
  const grouping = encodedGrouping === null ? DEFAULT_GROUPING : parseGrouping(encodedGrouping, "Grabber panel");
  return { server: serverOrigin(server), project, ...(runtime ? { runtime } : {}), grouping, ...(agentSession ? { agentSession } : {}), parentOrigin, bridgeSession };
}
