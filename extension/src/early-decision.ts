import { parseSiteSettings } from "../host/site-settings";

export type EarlyDecision =
  | { kind: "off" }
  | { kind: "on"; installHook: boolean }
  | { kind: "failed"; message: string };

/**
 * What the early script does for the page it was injected into: nothing unless
 * the site's own settings say the extension is enabled, and then install the
 * React DevTools hook only when nothing defined one yet. Unreadable settings are
 * reported, never treated as enabled.
 */
export function earlyDecision(options: { read: () => string | null; hookDefined: boolean }): EarlyDecision {
  let raw: string | null;
  try {
    raw = options.read();
  } catch (cause) {
    return { kind: "failed", message: `Gavel: this page's storage is not available, so it cannot tell whether the grabber is enabled here: ${cause instanceof Error ? cause.message : String(cause)}` };
  }
  if (raw === null) return { kind: "off" };
  try {
    return parseSiteSettings(raw).enabled ? { kind: "on", installHook: !options.hookDefined } : { kind: "off" };
  } catch (cause) {
    return { kind: "failed", message: `Gavel: ${cause instanceof Error ? cause.message : String(cause)}` };
  }
}
