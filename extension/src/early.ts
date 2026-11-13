import { SITE_SETTINGS_KEY } from "../host/site-settings";
import { earlyDecision } from "./early-decision";
import { ENABLED_EVENT } from "./messages";

// Runs in the page's own world at document_start on every page, before any page
// script: react-dom only registers with a DevTools hook that already exists.
// It does nothing unless this site enabled the grabber in its own localStorage.

/** bippy's `install-hook-only` script, inlined by the extension build (the same script the playground shell runs first). */
declare const __INSTALL_REACT_DEVTOOLS_HOOK__: () => void;

const decision = earlyDecision({
  read: () => window.localStorage.getItem(SITE_SETTINGS_KEY),
  hookDefined: (window as { __REACT_DEVTOOLS_GLOBAL_HOOK__?: unknown }).__REACT_DEVTOOLS_GLOBAL_HOOK__ !== undefined,
});

if (decision.kind === "failed") console.error(decision.message);
else if (decision.kind === "on") {
  if (decision.installHook) __INSTALL_REACT_DEVTOOLS_HOOK__();
  // The grabber mounts into document.body, so the isolated-world bridge is told once the document is parsed.
  const signal = () => window.dispatchEvent(new CustomEvent(ENABLED_EVENT));
  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", signal, { once: true });
  else signal();
}
