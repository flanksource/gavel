import type { UploadedAttachment } from "../comments/gavel-todo";
import type { CommentSubmission } from "./comment-sync";
import type { Capture } from "./todos/model";
import type { TodoRun } from "./todos/runtime";

/** Dispatched by the page-world grabber when the user closes it; the isolated-world bridge clears the site's `enabled` flag and forwards it to the background. */
export const CLOSED_EVENT = "gavel-grabber:closed";
/** Dispatched by the page-world early script once the document is ready and the site is enabled; the bridge forwards it as a `page-enabled` request. */
export const ENABLED_EVENT = "gavel-grabber:enabled";
/** Dispatched by the page-world grabber to ask for a screenshot of the visible tab; the bridge answers with `CAPTURE_RESULT_EVENT`. */
export const CAPTURE_REQUEST_EVENT = "gavel-grabber:capture";
export const CAPTURE_RESULT_EVENT = "gavel-grabber:captured";

/** Dispatched by the page-world grabber to upload a screenshot to the tab's gavel server; the bridge answers with `UPLOAD_RESULT_EVENT`. */
export const UPLOAD_REQUEST_EVENT = "gavel-grabber:upload";
export const UPLOAD_RESULT_EVENT = "gavel-grabber:uploaded";

/** `id` pairs a result with the request it answers. */
export type CaptureRequestDetail = { id: string };
export type CaptureResultDetail = { id: string } & ({ ok: true; dataUrl: string } | { ok: false; error: string });
export type UploadRequestDetail = { id: string; dataUrl: string; name: string };
export type UploadResultDetail = { id: string } & ({ ok: true; attachment: UploadedAttachment } | { ok: false; error: string });

/** What extension pages (popup, relay, panel) and the bridge ask the background worker. */
export type BackgroundRequest =
  | { type: "todo-snapshot"; server: string; dir: string; sourceUrl: string; status?: string }
  | { type: "todo-refresh"; server: string; dir: string }
  | { type: "todo-runtime-refresh"; server: string; dir: string; status: string }
  | { type: "todo-run-choice"; server: string; dir: string; run: TodoRun | null }
  | { type: "todo-capture"; server: string; capture: Capture }
  | { type: "todo-sync"; server: string; retryId?: string }
  | { type: "comment-capture"; draft: CommentSubmission }
  | { type: "comment-queue"; pageUrl: string }
  | { type: "comment-retry"; pageUrl: string; id: string }
  /** Show the grabber on `tabId` with that tab's effective settings. */
  | { type: "activate"; tabId: number }
  /** Close the grabber on `tabId` and forget its activation. */
  | { type: "deactivate"; tabId: number }
  /** The sender tab's document loaded while enabled for its site: activate it. */
  | { type: "page-enabled" }
  /** A PNG data URL of the sender tab's visible area. */
  | { type: "capture" }
  /** Upload a PNG data URL to the sender tab's gavel server as a todo attachment. */
  | { type: "upload-attachment"; dataUrl: string; name: string }
  /** The settings the sender tab's grabber was activated with. */
  | { type: "activation" }
  /** The panel started or chose the extension-wide global session; the background stores it in the global defaults. `null` clears it. */
  | { type: "global-session"; session: string | null };

export type BackgroundResponse<T> = { ok: true; value: T } | { ok: false; error: string };

export async function requestBackground<T>(request: BackgroundRequest): Promise<T> {
  const response = (await chrome.runtime.sendMessage(request)) as BackgroundResponse<T> | undefined;
  if (!response) throw new Error("The grabber's background worker did not answer");
  if (!response.ok) throw new Error(response.error);
  return response.value;
}
