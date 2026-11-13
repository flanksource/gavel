import type { UploadedAttachment } from "../comments/gavel-todo";
import { UPLOAD_REQUEST_EVENT, UPLOAD_RESULT_EVENT, type UploadRequestDetail, type UploadResultDetail } from "./messages";

const BRIDGE_TIMEOUT_MS = 30_000;

/**
 * The host's `upload` inside the extension: asks the isolated-world bridge, which
 * asks the background worker, to upload a screenshot PNG to the gavel server the
 * tab's grabber was activated with. The page world never reaches the server
 * itself, so the upload is not subject to the page's CSP or CORS.
 */
export function extensionUpload(options: { dataUrl: string; name: string }): Promise<UploadedAttachment> {
  const id = crypto.randomUUID();
  return new Promise((resolve, reject) => {
    const finish = () => { clearTimeout(timeout); window.removeEventListener(UPLOAD_RESULT_EVENT, onResult); };
    const onResult = (event: Event) => {
      const result = (event as CustomEvent<UploadResultDetail>).detail;
      if (result.id !== id) return;
      finish();
      if (result.ok) resolve(result.attachment);
      else reject(new Error(result.error));
    };
    const timeout = setTimeout(() => { finish(); reject(new Error("the grabber's extension bridge did not answer the screenshot upload")); }, BRIDGE_TIMEOUT_MS);
    window.addEventListener(UPLOAD_RESULT_EVENT, onResult);
    window.dispatchEvent(new CustomEvent<UploadRequestDetail>(UPLOAD_REQUEST_EVENT, { detail: { id, dataUrl: options.dataUrl, name: options.name } }));
  });
}
