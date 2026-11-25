import { pngBlob } from "../comments/png-blob";
import type { UploadedAttachment } from "../comments/gavel-todo";

/** The server may answer with a path relative to itself; the todo body needs a URL that works anywhere. */
export function absoluteAttachmentUrl(pathOrUrl: string, server: string): string {
  return new URL(pathOrUrl, server).toString();
}

/**
 * Uploads a screenshot to gavel's todo attachments API (multipart field
 * `attachment`) from an extension context, where host permissions apply and the
 * page's CSP and CORS do not. `server` comes from the tab's activation record,
 * never from the page.
 */
export async function uploadAttachment(options: { server: string; dataUrl: string; name: string; fetcher?: typeof fetch }): Promise<UploadedAttachment> {
  const form = new FormData();
  form.append("attachment", pngBlob(options.dataUrl), options.name);
  const response = await (options.fetcher ?? fetch)(`${options.server}/api/todos/attachments`, { method: "POST", body: form });
  const data = (await response.json().catch(() => ({}))) as { error?: string; attachments?: { url?: string; filename?: string }[] };
  if (!response.ok) throw new Error(data.error || `screenshot upload failed (HTTP ${response.status})`);
  const attachment = data.attachments?.[0];
  if (!attachment?.url) throw new Error("screenshot upload returned no URL");
  return { url: absoluteAttachmentUrl(attachment.url, options.server), ...(attachment.filename ? { filename: attachment.filename } : {}) };
}
