import { afterEach, expect, it, vi } from "vitest";
import { createCaptureBadge } from "./capture-badge";

afterEach(() => vi.unstubAllGlobals());

function setup() {
  const action = {
    setBadgeText: vi.fn(async (_details: { tabId?: number; text: string }) => {}),
    setBadgeBackgroundColor: vi.fn(async (_details: { tabId?: number; color: string }) => {}),
    setTitle: vi.fn(async (_details: { tabId?: number; title: string }) => {}),
  };
  vi.stubGlobal("chrome", { action });
  const queue = {
    queue: vi.fn<() => Promise<{ state: string; error?: string | undefined }[]>>().mockResolvedValue([]),
    commentQueue: vi.fn<() => Promise<{ state: string; error?: string | undefined }[]>>().mockResolvedValue([]),
  };
  const tabs = vi.fn().mockResolvedValue([{ id: 12, active: true }, { id: 34, active: false, error: "Activation failed" }]);
  return { action, queue, tabs, refresh: createCaptureBadge({ store: queue, tabs }) };
}

it("shows one count across all tabs and flags rejected, interrupted, and retrying submissions", async () => {
  const { action, queue, refresh } = setup();
  queue.queue.mockResolvedValue([{ state: "pending" }, { state: "sending" }, { state: "uncertain" }, { state: "run_error" }, { state: "synced" }]);
  queue.commentQueue.mockResolvedValue([{ state: "error" }, { state: "pending", error: "Server unavailable" }, { state: "synced" }]);
  await refresh();

  expect(action.setBadgeText.mock.calls.map(([details]) => details)).toEqual([
    { text: "6!" }, { tabId: 12, text: "6!" }, { tabId: 34, text: "6!" },
  ]);
  expect(action.setBadgeBackgroundColor).toHaveBeenCalledWith({ color: "#dc2626" });
  expect(action.setTitle).toHaveBeenCalledWith({ title: "Gavel — 6 queued, 4 with errors" });
  expect(action.setTitle).toHaveBeenCalledWith({ tabId: 34, title: "Gavel — 6 queued, 4 with errors · Activation failed" });
});

it("replaces the error indicator during retry and restores the tab badge after sync", async () => {
  const { action, queue, refresh } = setup();
  queue.queue.mockResolvedValue([{ state: "error" }]);
  await refresh();
  queue.queue.mockResolvedValue([{ state: "sending" }]);
  await refresh();
  expect(action.setBadgeText).toHaveBeenCalledWith({ text: "1" });
  expect(action.setBadgeBackgroundColor).toHaveBeenCalledWith({ color: "#2563eb" });
  queue.queue.mockResolvedValue([{ state: "synced" }]);
  await refresh();
  expect(action.setBadgeText.mock.calls.slice(-3).map(([details]) => details)).toEqual([
    { text: "" }, { tabId: 12, text: "on" }, { tabId: 34, text: "!" },
  ]);
});
