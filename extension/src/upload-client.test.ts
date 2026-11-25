import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { extensionUpload } from "./upload-client";
import { UPLOAD_REQUEST_EVENT, UPLOAD_RESULT_EVENT, type UploadRequestDetail, type UploadResultDetail } from "./messages";

const listeners: EventListener[] = [];
function bridgeAnswers(answer: (request: UploadRequestDetail) => Omit<UploadResultDetail, "id"> | null) {
  const listener = ((event: CustomEvent<UploadRequestDetail>) => {
    const reply = answer(event.detail);
    if (reply) window.dispatchEvent(new CustomEvent<UploadResultDetail>(UPLOAD_RESULT_EVENT, { detail: { id: event.detail.id, ...reply } as UploadResultDetail }));
  }) as EventListener;
  window.addEventListener(UPLOAD_REQUEST_EVENT, listener);
  listeners.push(listener);
}

beforeEach(() => vi.useFakeTimers());
afterEach(() => {
  vi.useRealTimers();
  for (const listener of listeners.splice(0)) window.removeEventListener(UPLOAD_REQUEST_EVENT, listener);
});

const png = { dataUrl: "data:image/png;base64,iVBORw0KGgo=", name: "screenshot.png" };

it("sends the PNG and resolves with the attachment the extension stored", async () => {
  const seen: UploadRequestDetail[] = [];
  bridgeAnswers((request) => {
    seen.push(request);
    return { ok: true, attachment: { url: "http://localhost:9092/files/a.png", filename: "screenshot.png" } };
  });
  await expect(extensionUpload(png)).resolves.toEqual({ url: "http://localhost:9092/files/a.png", filename: "screenshot.png" });
  expect(seen).toHaveLength(1);
  expect([seen[0]!.dataUrl, seen[0]!.name, typeof seen[0]!.id]).toEqual([png.dataUrl, png.name, "string"]);
});

it("ignores answers to other uploads", async () => {
  bridgeAnswers((request) => {
    window.dispatchEvent(new CustomEvent<UploadResultDetail>(UPLOAD_RESULT_EVENT, { detail: { id: `not-${request.id}`, ok: true, attachment: { url: "http://wrong" } } }));
    return { ok: true, attachment: { url: "http://right" } };
  });
  await expect(extensionUpload(png)).resolves.toEqual({ url: "http://right" });
});

it("rejects with the reason the extension gave", async () => {
  bridgeAnswers(() => ({ ok: false, error: "attachment too large" }));
  await expect(extensionUpload(png)).rejects.toThrow("attachment too large");
});

it("rejects when nothing answers, naming the missing bridge", async () => {
  const pending = expect(extensionUpload(png)).rejects.toThrow(/bridge did not answer the screenshot upload/);
  await vi.advanceTimersByTimeAsync(30_000);
  await pending;
});
