import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { reactDevtoolsHookScript } from "../scripts/react-devtools-hook";
import { ENABLED_EVENT } from "./messages";

type HookWindow = { __REACT_DEVTOOLS_GLOBAL_HOOK__?: unknown };
const hookWindow = window as unknown as HookWindow;

async function runEarlyScript() {
  vi.resetModules();
  await import("./early");
}

// The extension build inlines bippy's real installer as this global; the test runs the same script.
// Vitest runs from the playground root, which resolves bippy the same way the build does.
const hookScript = reactDevtoolsHookScript(process.cwd());

beforeEach(() => {
  localStorage.clear();
  delete hookWindow.__REACT_DEVTOOLS_GLOBAL_HOOK__;
  vi.stubGlobal("__INSTALL_REACT_DEVTOOLS_HOOK__", () => new Function(hookScript)());
});
afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  localStorage.clear();
  delete hookWindow.__REACT_DEVTOOLS_GLOBAL_HOOK__;
});

function listenForEnabled() {
  const heard = vi.fn();
  window.addEventListener(ENABLED_EVENT, heard, { once: true });
  return heard;
}

it("leaves a page that did not enable the grabber completely alone", async () => {
  localStorage.setItem("flanksource.grabber", JSON.stringify({ project: "docs" }));
  const heard = listenForEnabled();
  await runEarlyScript();
  expect(hookWindow.__REACT_DEVTOOLS_GLOBAL_HOOK__).toBeUndefined();
  expect(heard).not.toHaveBeenCalled();
  window.removeEventListener(ENABLED_EVENT, heard);
});

it("installs the DevTools hook and signals the bridge on an enabled site", async () => {
  localStorage.setItem("flanksource.grabber", JSON.stringify({ enabled: true }));
  const heard = listenForEnabled();
  await runEarlyScript();
  expect(hookWindow.__REACT_DEVTOOLS_GLOBAL_HOOK__).toBeDefined();
  expect(heard).toHaveBeenCalledTimes(1);
});

it("signals but keeps the hook that already exists", async () => {
  const existing = { supportsFiber: true, renderers: new Map(), inject: vi.fn() };
  hookWindow.__REACT_DEVTOOLS_GLOBAL_HOOK__ = existing;
  localStorage.setItem("flanksource.grabber", JSON.stringify({ enabled: true }));
  const heard = listenForEnabled();
  await runEarlyScript();
  expect(hookWindow.__REACT_DEVTOOLS_GLOBAL_HOOK__).toBe(existing);
  expect(existing.inject).not.toHaveBeenCalled();
  expect(heard).toHaveBeenCalledTimes(1);
});

it("logs unreadable storage and does nothing else", async () => {
  vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => { throw new DOMException("denied", "SecurityError"); });
  const error = vi.spyOn(console, "error").mockImplementation(() => {});
  const heard = listenForEnabled();
  await runEarlyScript();
  expect(error).toHaveBeenCalledWith(expect.stringMatching(/storage is not available.*denied/));
  expect(hookWindow.__REACT_DEVTOOLS_GLOBAL_HOOK__).toBeUndefined();
  expect(heard).not.toHaveBeenCalled();
  window.removeEventListener(ENABLED_EVENT, heard);
});
