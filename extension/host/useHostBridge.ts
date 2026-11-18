import { useEffect, useState } from "react";
import { useCommentRailWidth } from "@flanksource/clicky-ui/comments";
import { sidebarWidth } from "./host-resize";
import { matchesBookmarkletParent } from "./origin";
import { parseAnchors, parseFabPosition, parseHostState, parsePresentation, validateParentOrigin } from "./bridge-parse";
import { hoverSize, type HoverSize } from "./host-hover-resize";
import { bridgeMessage, envelope, panelMode, type AnchorState, type HostState, type PanelMessage, type PanelMode, type Presentation } from "./model";

export type HostConnection = { parentOrigin: string; session: string };
let connection: HostConnection | undefined;

const MODE_KEY = "playground.bookmarklet.mode";
const HOVER_KEY = "playground.bookmarklet.hover";

/** The rail-or-hover mode and hover size the user last left the panel in; throws on a corrupt record. */
export function storedPresentation(): { mode?: PanelMode; hover?: HoverSize } {
  const mode = localStorage.getItem(MODE_KEY);
  const hover = localStorage.getItem(HOVER_KEY);
  return {
    ...(mode !== null ? { mode: panelMode(mode) } : {}),
    ...(hover !== null ? { hover: hoverSize(JSON.parse(hover)) } : {}),
  };
}

/** Asks the host to show the panel as the side rail or the hover popover. */
export function requestPanelMode(mode: PanelMode) {
  sendToHost({ type: "presentation-request", mode });
}

/** Panels with another frame URL contract (the extension's gavel panel) attach before they render. */
export function connectHostBridge(next: HostConnection) {
  connection = next;
}

export function sendToHost(message: PanelMessage) {
  if (!connection || window.parent === window)
    throw new Error("Bookmarklet bridge is not attached to a parent document");
  const { parentOrigin: peer, session } = connection;
  // File documents have opaque origins and require a wildcard target. Incoming
  // messages are still restricted to the parent window and this session.
  window.parent.postMessage(envelope(session, message), peer === "null" ? "*" : peer);
}

export async function saveHostSession(session: string | undefined): Promise<void> {
  const host = connection;
  if (!host) throw new Error("Bookmarklet bridge is not attached to a parent document");
  const requestId = crypto.randomUUID();
  return new Promise((resolve, reject) => {
    const finish = (error?: string) => {
      window.clearTimeout(timer);
      window.removeEventListener("message", receive);
      if (error !== undefined) reject(new Error(error));
      else resolve();
    };
    const receive = (event: MessageEvent) => {
      if (event.source !== window.parent || !matchesBookmarkletParent({ expected: host.parentOrigin, received: event.origin })) return;
      const result = bridgeMessage(event.data, host.session);
      if (result?.["type"] !== "settings-result" || result["requestId"] !== requestId) return;
      finish(typeof result["error"] === "string" ? result["error"] : undefined);
    };
    const timer = window.setTimeout(() => finish("The page did not confirm the session change"), 5000);
    window.addEventListener("message", receive);
    try {
      sendToHost({ type: "settings", requestId, session: session ?? null });
    } catch (cause) {
      finish(cause instanceof Error ? cause.message : String(cause));
    }
  });
}

export function useHostBridge() {
  const [width, setWidth] = useCommentRailWidth(400, "playground.bookmarklet.width");
  const [state, setState] = useState<HostState>();
  const [anchors, setAnchors] = useState<AnchorState[]>([]);
  const [focus, setFocus] = useState<string>();
  const [error, setError] = useState("");
  const [presentation, setPresentation] = useState<Presentation>("compact");
  useEffect(() => {
    if (!connection || window.parent === window) {
      setError("Open the grabber by enabling the Gavel extension on this site.");
      return;
    }
    const { parentOrigin, session } = connection;
    let peer: string;
    try {
      validateParentOrigin(parentOrigin);
      peer = parentOrigin;
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
      return;
    }
    const receive = (event: MessageEvent) => {
      if (!matchesBookmarkletParent({ expected: peer, received: event.origin }) || event.source !== window.parent)
        return;
      const message = bridgeMessage(event.data, session);
      if (!message) return;
      try {
        if (message["type"] === "presentation") {
          const next = parsePresentation(message["mode"]);
          setPresentation(next);
          if (next !== "draft") localStorage.setItem(MODE_KEY, next);
        } else if (message["type"] === "hover-size") {
          localStorage.setItem(HOVER_KEY, JSON.stringify(hoverSize({ width: message["width"], height: message["height"] })));
        } else if (message["type"] === "fab-position") {
          localStorage.setItem("playground.bookmarklet.position", JSON.stringify(parseFabPosition(message)));
        } else if (message["type"] === "sidebar-width") {
          setWidth(sidebarWidth(message["width"]));
        } else if (message["type"] === "state") {
          setState(parseHostState(message["state"], parentOrigin));
          setAnchors([]);
          setFocus(undefined);
          setError("");
        } else if (message["type"] === "anchors") {
          setAnchors(parseAnchors(message["anchors"]));
        } else if (message["type"] === "focus" && typeof message["anchor"] === "string") {
          setFocus(message["anchor"]);
        } else if (message["type"] === "error" && typeof message["error"] === "string") {
          setError(message["error"]);
        }
      } catch (cause) {
        setError(cause instanceof Error ? cause.message : String(cause));
      }
    };
    window.addEventListener("message", receive);
    try {
      sendToHost({ type: "ready", width, ...storedPresentation() });
    } catch (cause) {
      setError(`The remembered panel layout is invalid, so it was reset: ${cause instanceof Error ? cause.message : String(cause)}`);
      localStorage.removeItem(MODE_KEY);
      localStorage.removeItem(HOVER_KEY);
      sendToHost({ type: "ready", width });
    }
    try {
      const stored = localStorage.getItem("playground.bookmarklet.position");
      if (stored) sendToHost({ type: "fab-position", ...JSON.parse(stored) as { x: number; y: number } });
    } catch (cause) { setError(String(cause)); }
    return () => window.removeEventListener("message", receive);
  }, []);
  return { state, anchors, focus, error, floating: presentation === "draft", presentation };
}
