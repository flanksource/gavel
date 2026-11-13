import type { UploadedAttachment } from "../comments/gavel-todo";
import { uploadAttachment } from "./attachment-upload";
import { CLOSED_EVENT, type BackgroundRequest, type BackgroundResponse } from "./messages";
import { effectiveSettings, hostLaunch, type Activation, type HostLaunch } from "./settings";
import { readGlobalDefaults, readSiteSettings } from "./storage";
import { startTodoSync, todoRequest, warmTodoCatalog } from "./todos/background";

startTodoSync();

// The tabs where the grabber is showing, with the settings it was opened with.
// The relay and panel frames in them may only act for these settings, and only
// they may ask for screenshots. A full navigation to a site that is enabled
// activates the tab again through the early script; closing the grabber or the
// tab removes the record.
const ACTIVATIONS_KEY = "activations";
type Activations = Record<string, Activation>;

async function activations(): Promise<Activations> {
  const stored = (await chrome.storage.session.get(ACTIVATIONS_KEY))[ACTIVATIONS_KEY];
  if (stored === undefined) return {};
  if (typeof stored !== "object" || stored === null || Array.isArray(stored)) throw new Error(`Stored grabber activations are not an object: ${JSON.stringify(stored)}`);
  return stored as Activations;
}

async function setActive(tabId: number, activation: Activation | null) {
  const { [String(tabId)]: _previous, ...others } = await activations();
  await chrome.storage.session.set({ [ACTIVATIONS_KEY]: activation ? { ...others, [String(tabId)]: activation } : others });
  await chrome.action.setBadgeText({ tabId, text: activation ? "on" : "" });
  await chrome.action.setBadgeBackgroundColor({ tabId, color: "#2563eb" });
  await chrome.action.setTitle({ tabId, title: activation ? "Gavel: active on this tab" : "Gavel" });
}

// A popup switch and the early script's signal can both ask for the same tab
// while it loads; they share one activation instead of racing two.
const activating = new Map<number, Promise<void>>();

function activate(tabId: number): Promise<void> {
  const running = activating.get(tabId);
  if (running) return running;
  const work = install(tabId).finally(() => activating.delete(tabId));
  activating.set(tabId, work);
  return work;
}

async function install(tabId: number) {
  const settings = effectiveSettings(await readGlobalDefaults(), await readSiteSettings(tabId));
  warmTodoCatalog(settings.server);
  const launch = hostLaunch(settings, { panel: chrome.runtime.getURL("panel.html"), todoRelay: chrome.runtime.getURL("todo-relay.html") });
  const target = { tabId };
  // The bridge is also a static content script; this covers tabs that were open before the extension loaded, and is a no-op where it already runs.
  await chrome.scripting.executeScript({ target, files: ["bridge.js"] });
  const [loaded] = await chrome.scripting.executeScript({ target, world: "MAIN", args: [], func: () => "GavelGrabber" in window });
  // Extension files run regardless of the page's script-src policy.
  if (!loaded?.result) await chrome.scripting.executeScript({ target, world: "MAIN", files: ["grabber.js"] });
  // The host frames the relay or panel as it installs, and they ask for this record as soon as they load, so it is written first and dropped again if installing fails.
  await setActive(tabId, { server: settings.server, ...(settings.project ? { project: settings.project } : {}) });
  try {
    await installHost(tabId, launch);
  } catch (cause) {
    await setActive(tabId, null);
    throw cause;
  }
}

async function installHost(tabId: number, launch: HostLaunch) {
  const [installed] = await chrome.scripting.executeScript({
    target: { tabId },
    world: "MAIN",
    args: [launch, CLOSED_EVENT],
    func: (launch: HostLaunch, closed: string) => {
      const grabber = (window as { GavelGrabber?: { installBookmarklet(options: HostLaunch & { onClose: () => void; capture: unknown; upload: unknown }): void; extensionCapture: unknown; extensionUpload: unknown } }).GavelGrabber;
      if (!grabber) return "grabber.js did not define GavelGrabber";
      try {
        grabber.installBookmarklet({ ...launch, capture: grabber.extensionCapture, upload: grabber.extensionUpload, onClose: () => window.dispatchEvent(new CustomEvent(closed)) });
        return null;
      } catch (cause) {
        return cause instanceof Error ? cause.message : String(cause);
      }
    },
  });
  if (installed?.result) throw new Error(installed.result);
}

/** Closes the tab's grabber (restoring the page's own react-grab state) and forgets the activation. */
async function deactivate(tabId: number) {
  await chrome.scripting.executeScript({ target: { tabId }, world: "MAIN", args: [], func: () => (window as { __GAVEL_GRABBER__?: { close(): void } }).__GAVEL_GRABBER__?.close() });
  await setActive(tabId, null);
}

/** A PNG of what the tab shows, for a frame of the tab's own active grabber. */
async function captureTab(sender: chrome.runtime.MessageSender): Promise<string> {
  const tab = sender.tab;
  if (tab?.id === undefined || tab.windowId === undefined) throw new Error("Only a page inside a tab can ask for a screenshot");
  if (!(String(tab.id) in (await activations()))) throw new Error("The grabber is not active on this tab");
  if (!tab.active) throw new Error("Only the visible tab can be captured; switch to this tab and try again");
  return chrome.tabs.captureVisibleTab(tab.windowId, { format: "png" });
}

/** Uploads a screenshot for a tab's gavel-backed grabber, to the server the tab was activated with (never one the page names). */
async function uploadForTab(sender: chrome.runtime.MessageSender, request: { dataUrl: unknown; name: unknown }): Promise<UploadedAttachment> {
  const tabId = sender.tab?.id;
  if (tabId === undefined) throw new Error("Only a page inside a tab can upload a screenshot");
  const activation = (await activations())[String(tabId)];
  if (!activation) throw new Error("The grabber is not active on this tab");
  if (typeof request.dataUrl !== "string" || typeof request.name !== "string" || request.name === "") throw new Error("A screenshot upload needs a PNG data URL and a file name");
  return uploadAttachment({ server: activation.server, dataUrl: request.dataUrl, name: request.name });
}

async function report(tabId: number, cause: unknown) {
  const message = cause instanceof Error ? cause.message : String(cause);
  console.error(`Gavel failed on tab ${tabId}:`, cause);
  await chrome.action.setBadgeText({ tabId, text: "!" });
  await chrome.action.setBadgeBackgroundColor({ tabId, color: "#dc2626" });
  await chrome.action.setTitle({ tabId, title: `Gavel: ${message}` });
}

chrome.tabs.onRemoved.addListener(tabId => {
  void activations().then(({ [String(tabId)]: _closed, ...others }) => chrome.storage.session.set({ [ACTIVATIONS_KEY]: others }));
});

function answer<T>(sendResponse: (response: BackgroundResponse<T>) => void, work: Promise<T>) {
  work.then(
    value => sendResponse({ ok: true, value }),
    (cause: unknown) => sendResponse({ ok: false, error: cause instanceof Error ? cause.message : String(cause) }),
  );
  return true;
}

chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  const todo = todoRequest(message as BackgroundRequest, sender);
  if (todo) return answer(sendResponse, todo);
  const request = message as { type?: BackgroundRequest["type"] | typeof CLOSED_EVENT; tabId?: unknown } | null;
  const senderTab = sender.tab?.id;
  if (request?.type === CLOSED_EVENT) {
    if (senderTab !== undefined) setActive(senderTab, null).catch(cause => report(senderTab, cause));
    return;
  }
  if (request?.type === "activate" || request?.type === "deactivate") {
    const tabId = request.tabId;
    if (typeof tabId !== "number") return answer(sendResponse, Promise.reject(new Error(`${request.type === "activate" ? "Activation" : "Deactivation"} requires a tab id`)));
    const work = request.type === "activate" ? activate(tabId) : deactivate(tabId);
    return answer(sendResponse, work.catch(async (cause: unknown) => { await report(tabId, cause); throw cause; }));
  }
  if (request?.type === "page-enabled") {
    if (senderTab === undefined || sender.frameId !== 0) return answer(sendResponse, Promise.reject(new Error("Only the top frame of a tab can ask to be activated")));
    return answer(sendResponse, activate(senderTab).catch(async (cause: unknown) => { await report(senderTab, cause); throw cause; }));
  }
  if (request?.type === "capture") return answer(sendResponse, captureTab(sender));
  if (request?.type === "upload-attachment") return answer(sendResponse, uploadForTab(sender, request as { dataUrl: unknown; name: unknown }));
  if (request?.type === "activation") {
    if (senderTab === undefined) return answer(sendResponse, Promise.reject(new Error("Only a frame inside a tab can ask for its activation")));
    return answer(sendResponse, activations().then(active => active[String(senderTab)]));
  }
  console.error("Grabber background received an unknown message", message);
});
