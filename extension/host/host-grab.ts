import { init } from "@flanksource/react-grab/core";
import type { ReactGrabAPI } from "@flanksource/react-grab";

/**
 * The picker the grabber host drives: the page's own react-grab when it has one,
 * otherwise a full-toolbar instance of the vendored fork that the host owns (and
 * must dispose). It is not published as `window.__REACT_GRAB__`, so the page and
 * its other tools never see it.
 */
export function acquireGrab(): { grab: ReactGrabAPI; created: boolean } {
  const existing = window.__REACT_GRAB__;
  if (existing) return { grab: existing, created: false };
  return { grab: init({ enabled: true, telemetry: false, freezeReactUpdates: false }), created: true };
}
