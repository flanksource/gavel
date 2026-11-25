import { SITE_SETTINGS_KEY, withSiteEnabled } from "../host/site-settings";
import type { UploadedAttachment } from "../comments/gavel-todo";
import {
  CAPTURE_REQUEST_EVENT,
  CAPTURE_RESULT_EVENT,
  CLOSED_EVENT,
  ENABLED_EVENT,
  UPLOAD_REQUEST_EVENT,
  UPLOAD_RESULT_EVENT,
  requestBackground,
  type CaptureRequestDetail,
  type CaptureResultDetail,
  type UploadRequestDetail,
  type UploadResultDetail,
} from "./messages";

function describe(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause);
}

/** Closing the grabber disables the extension for the site, in the page's own storage (this world shares it). */
function clearEnabled() {
  const stored = withSiteEnabled(window.localStorage.getItem(SITE_SETTINGS_KEY), false);
  if (stored === null) window.localStorage.removeItem(SITE_SETTINGS_KEY);
  else window.localStorage.setItem(SITE_SETTINGS_KEY, stored);
}

/**
 * Any page script can fire the capture request, so the bridge itself hides the
 * grabber's root (which holds the comment panel frame) instead of trusting the
 * page-world host to have done so; the host's own hiding still covers react-grab's overlays.
 */
async function withPanelHidden<T>(work: () => Promise<T>): Promise<T> {
  const root = document.getElementById("gavel-grabber");
  if (!root) return work();
  const previous = { value: root.style.getPropertyValue("display"), priority: root.style.getPropertyPriority("display") };
  root.style.setProperty("display", "none", "important");
  try {
    await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())));
    return await work();
  } finally {
    if (previous.value) root.style.setProperty("display", previous.value, previous.priority);
    else root.style.removeProperty("display");
  }
}

async function answerCapture(id: string) {
  const result: CaptureResultDetail = await withPanelHidden(() => requestBackground<string>({ type: "capture" })).then(
    (dataUrl): CaptureResultDetail => ({ id, ok: true, dataUrl }),
    (cause: unknown): CaptureResultDetail => ({ id, ok: false, error: describe(cause) }),
  );
  window.dispatchEvent(new CustomEvent<CaptureResultDetail>(CAPTURE_RESULT_EVENT, { detail: result }));
}

async function answerUpload(request: UploadRequestDetail) {
  const result: UploadResultDetail = await requestBackground<UploadedAttachment>({ type: "upload-attachment", dataUrl: request.dataUrl, name: request.name }).then(
    (attachment): UploadResultDetail => ({ id: request.id, ok: true, attachment }),
    (cause: unknown): UploadResultDetail => ({ id: request.id, ok: false, error: describe(cause) }),
  );
  window.dispatchEvent(new CustomEvent<UploadResultDetail>(UPLOAD_RESULT_EVENT, { detail: result }));
}

// Runs in the extension's isolated world, a static content script (and injected
// again when the user enables a tab that was open before the extension loaded):
// the page-world grabber cannot reach chrome.runtime, so it signals with DOM
// events this script forwards.
const marker = globalThis as { __gavelGrabberBridge?: true };
if (!marker.__gavelGrabberBridge) {
  marker.__gavelGrabberBridge = true;
  window.addEventListener(ENABLED_EVENT, () => {
    requestBackground({ type: "page-enabled" }).catch((cause: unknown) => console.error(`Gavel could not start on this page: ${describe(cause)}`));
  });
  window.addEventListener(CLOSED_EVENT, () => {
    try {
      clearEnabled();
    } catch (cause) {
      console.error(`Gavel could not clear its enabled flag, so it will reopen on reload: ${describe(cause)}`);
    }
    void chrome.runtime.sendMessage({ type: CLOSED_EVENT });
  });
  window.addEventListener(CAPTURE_REQUEST_EVENT, (event) => {
    const id = (event as CustomEvent<Partial<CaptureRequestDetail> | null>).detail?.id;
    if (typeof id !== "string" || id === "") console.error("Gavel ignored a screenshot request without an id");
    else void answerCapture(id);
  });
  window.addEventListener(UPLOAD_REQUEST_EVENT, (event) => {
    const detail = (event as CustomEvent<Partial<UploadRequestDetail> | null>).detail;
    if (typeof detail?.id !== "string" || detail.id === "") console.error("Gavel ignored a screenshot upload without an id");
    else if (typeof detail.dataUrl !== "string" || typeof detail.name !== "string") void answerUpload({ id: detail.id, dataUrl: "", name: "" });
    else void answerUpload({ id: detail.id, dataUrl: detail.dataUrl, name: detail.name });
  });
}
