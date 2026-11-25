import { afterEach, beforeEach, expect, it, vi } from "vitest";
import type { Selection } from "../host/model";
import { pageSnapshot } from "./page-snapshot";

const TIMEOUT_MS = 1000;
const selection = (anchor: string): Selection => ({ anchor, label: anchor, element: { source: "https://example.test/:1:1", componentName: "Page", html: "<body></body>", locators: [] } });

beforeEach(() => vi.useFakeTimers());
afterEach(() => vi.useRealTimers());

it("asks the host for the page and resolves with its whole-page selection, which it claims from the draft flow", async () => {
  const requestPage = vi.fn();
  const snapshot = pageSnapshot({ requestPage, timeoutMs: TIMEOUT_MS });
  const pending = snapshot.request();
  expect(requestPage).toHaveBeenCalledOnce();
  expect(snapshot.claim(selection("button.save"))).toBe(false);
  expect(snapshot.claim(selection("__document__"))).toBe(true);
  await expect(pending).resolves.toEqual(selection("__document__"));
  expect(snapshot.claim(selection("__document__"))).toBe(false);
});

it("rejects when the host never reports the page, and then lets a new request start", async () => {
  const requestPage = vi.fn();
  const snapshot = pageSnapshot({ requestPage, timeoutMs: TIMEOUT_MS });
  const pending = snapshot.request();
  vi.advanceTimersByTime(TIMEOUT_MS);
  await expect(pending).rejects.toThrow("did not report its snapshot");
  expect(snapshot.claim(selection("__document__"))).toBe(false);
  const retry = snapshot.request();
  expect(requestPage).toHaveBeenCalledTimes(2);
  snapshot.claim(selection("__document__"));
  await expect(retry).resolves.toEqual(selection("__document__"));
});

it("refuses a second request while one is pending", async () => {
  const snapshot = pageSnapshot({ requestPage: vi.fn(), timeoutMs: TIMEOUT_MS });
  void snapshot.request();
  await expect(snapshot.request()).rejects.toThrow("already being posted");
});
