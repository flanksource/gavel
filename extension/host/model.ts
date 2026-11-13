import type { CommentElementCaptureContext } from "../contracts/comments-model";
import type { AgentPin } from "@flanksource/page-runtime/schema";

export const BRIDGE = "gavel-extension-v1";

export function externalTitle(options: { url: string; title: string }): string {
  if (options.title.trim()) return options.title;
  const url = new URL(options.url);
  return url.protocol === "file:" ? decodeURIComponent(url.pathname).split("/").filter(Boolean).at(-1) ?? url.href : url.hostname;
}

export function externalUrl(input: string): { url: string; base: string } {
  const url = new URL(input);
  if (
    !["http:", "https:", "file:"].includes(url.protocol) ||
    url.username ||
    url.password
  ) {
    throw new Error(
      "External pages require an HTTP(S) or file URL without credentials",
    );
  }
  url.searchParams.sort();
  const origin = url.protocol === "file:" ? `file://${url.host}` : url.origin;
  return { url: url.href, base: `${origin}${url.pathname}${url.hash}` };
}

export type Selection = {
  anchor: string;
  label: string;
  element: CommentElementCaptureContext;
};

export type HostState = { url: string; title: string; selection?: Selection };
export type Presentation = "compact" | "draft" | "rail";
/** The presentations a user picks for the panel; `compact` is the hover popover. */
export type PanelMode = Exclude<Presentation, "draft">;

export function panelMode(input: unknown): PanelMode {
  if (input !== "rail" && input !== "compact") throw new Error(`Invalid panel mode ${JSON.stringify(input)}`);
  return input;
}
export type AnchorState = {
  anchor: string;
  label: string;
  available: boolean;
  top?: number;
};
export type HostMessage =
  | { type: "presentation"; mode: Presentation }
  | { type: "fab-position"; x: number; y: number }
  | { type: "sidebar-width"; width: number }
  | { type: "hover-size"; width: number; height: number }
  | { type: "state"; state: HostState }
  | { type: "anchors"; anchors: AnchorState[] }
  | { type: "focus"; anchor: string }
  | { type: "error"; error: string };
export type PanelMessage =
  | { type: "open-rail" }
  | { type: "dialog"; open: boolean }
  | { type: "fab-position"; x: number; y: number }
  | { type: "minimize" }
  | { type: "draft"; active: boolean; posted?: boolean; anchor?: string }
  | { type: "draft-size"; height: number }
  /** `mode` and `hover` restore the panel's remembered presentation; a panel without them opens as a hover popover. */
  | { type: "ready"; width: number; mode?: PanelMode; hover?: { width: number; height: number } }
  /** The user switched an open panel between the side rail and the hover popover. */
  | { type: "presentation-request"; mode: PanelMode }
  | { type: "sidebar-layout"; width: number; maximized: boolean }
  | { type: "close" }
  | { type: "pick"; screenshot?: boolean }
  | { type: "page" }
  /** The panel changed this site's agent session; the host persists it in the page's grabber settings. `null` clears it. */
  | { type: "settings"; session: string | null }
  | {
      type: "scroll";
      anchor: string;
      behavior?: ScrollBehavior;
      block?: "start" | "center";
      offset?: number;
    }
  | { type: "highlight"; anchor: string | null }
  | { type: "pins"; anchors: AgentPin[] };

export function envelope(session: string, message: HostMessage | PanelMessage) {
  return { bridge: BRIDGE, session, message };
}

export function bridgeMessage(
  data: unknown,
  session: string,
): Record<string, unknown> | null {
  if (!data || typeof data !== "object") return null;
  const value = data as Record<string, unknown>;
  if (value["bridge"] !== BRIDGE || value["session"] !== session) return null;
  const message = value["message"];
  return message && typeof message === "object" && !Array.isArray(message)
    ? (message as Record<string, unknown>)
    : null;
}
