import type { AgentSessionGrouping, AgentSessionRuntime } from "../contracts/agent-session-store";

export function bookmarkletParentOrigin(url: string): string {
  const parsed = new URL(url);
  return parsed.protocol === "file:" ? "null" : parsed.origin;
}

/** The extension page a host frames: its bundled gavel panel, configured by the settings it was activated with. */
export type GavelPanelFrame = {
  url: string;
  /** The extension page that frames the dashboard's new-todo form for the host, so a page whose CSP blocks the dashboard can still open it. */
  todoRelay: string;
  server: string;
  project: string;
  /** The runtime new agent sessions start with; absent means gavel's default. JSON-encoded in the frame URL. */
  runtime?: AgentSessionRuntime;
  grouping?: AgentSessionGrouping;
  /** The Captain session this site's comments go to; distinct from the bridge session. */
  agentSession?: string;
};

function extensionPage(input: string, what: string): { url: URL; origin: string } {
  const url = new URL(input);
  if (url.protocol !== "chrome-extension:") throw new Error(`Grabber ${what} must be a chrome-extension page, got ${url.href}`);
  // URL.origin is "null" for chrome-extension: outside Chrome itself.
  return { url, origin: `${url.protocol}//${url.host}` };
}

/** The bundled panel URL and the origin it exchanges messages with. */
export function bookmarkletFrame(
  options: { documentUrl: string; session: string; panel: GavelPanelFrame },
): { src: string; origin: string } {
  const parentOrigin = bookmarkletParentOrigin(options.documentUrl);
    const { url, origin } = extensionPage(options.panel.url, "panel");
    url.search = new URLSearchParams({
      server: options.panel.server,
      project: options.panel.project,
      ...(options.panel.runtime ? { runtime: JSON.stringify(options.panel.runtime) } : {}),
      ...(options.panel.grouping ? { grouping: options.panel.grouping } : {}),
      ...(options.panel.agentSession ? { session: options.panel.agentSession } : {}),
      parentOrigin,
      bridgeSession: options.session,
    }).toString();
    return { src: url.href, origin };
}

export function matchesBookmarkletParent(options: { expected: string; received: string }): boolean {
  // Chromium with file access serializes file message origins as file://.
  return options.expected === "null" ? options.received === "null" || options.received === "file://" : options.received === options.expected;
}
