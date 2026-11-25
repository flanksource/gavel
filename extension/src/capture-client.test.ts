import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { requestTabImage } from "./capture-client";
import { CAPTURE_REQUEST_EVENT, CAPTURE_RESULT_EVENT, type CaptureRequestDetail, type CaptureResultDetail } from "./messages";

const listeners: EventListener[] = [];
function bridgeAnswers(answer: (request: CaptureRequestDetail) => Omit<CaptureResultDetail, "id"> | null) {
  const listener = ((event: CustomEvent<CaptureRequestDetail>) => {
    const reply = answer(event.detail);
    if (reply) window.dispatchEvent(new CustomEvent<CaptureResultDetail>(CAPTURE_RESULT_EVENT, { detail: { id: event.detail.id, ...reply } as CaptureResultDetail }));
  }) as EventListener;
  window.addEventListener(CAPTURE_REQUEST_EVENT, listener);
  listeners.push(listener);
}

beforeEach(() => vi.useFakeTimers());
afterEach(() => {
  vi.useRealTimers();
  for (const listener of listeners.splice(0)) window.removeEventListener(CAPTURE_REQUEST_EVENT, listener);
});

it("resolves with the tab image the bridge returns for this request", async () => {
  bridgeAnswers(() => ({ ok: true, dataUrl: "data:image/png;base64,AAAA" }));
  await expect(requestTabImage()).resolves.toBe("data:image/png;base64,AAAA");
});

it("ignores answers to other requests", async () => {
  bridgeAnswers((request) => {
    window.dispatchEvent(new CustomEvent<CaptureResultDetail>(CAPTURE_RESULT_EVENT, { detail: { id: `not-${request.id}`, ok: true, dataUrl: "data:image/png;base64,WRONG" } }));
    return { ok: true, dataUrl: "data:image/png;base64,RIGHT" };
  });
  await expect(requestTabImage()).resolves.toBe("data:image/png;base64,RIGHT");
});

it("rejects with the reason the extension gave", async () => {
  bridgeAnswers(() => ({ ok: false, error: "tab is not active" }));
  await expect(requestTabImage()).rejects.toThrow("tab is not active");
});

it("rejects when nothing answers, naming the missing bridge", async () => {
  const pending = expect(requestTabImage()).rejects.toThrow(/bridge did not answer/);
  await vi.advanceTimersByTimeAsync(10_000);
  await pending;
});
