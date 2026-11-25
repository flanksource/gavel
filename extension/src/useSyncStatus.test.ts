import { act, renderHook, waitFor } from "@testing-library/react";
import { IDBFactory } from "fake-indexeddb";
import { afterEach, expect, it, vi } from "vitest";
import type { QueuedComment } from "./comment-sync";
import type { QueueEntry } from "./sync-status";
import type { QueuedTodo } from "./todos/model";
import { TodoStore } from "./todos/store";
import { useSyncStatus } from "./useSyncStatus";

afterEach(() => vi.useRealTimers());

const entries = (state: string, count: number): QueueEntry[] => Array.from({ length: count }, () => ({ state }));

function fakeStore(todos: QueueEntry[] = [], comments: QueueEntry[] = []) {
  const store = {
    todos,
    comments,
    queue: vi.fn(async () => store.todos),
    commentQueue: vi.fn(async () => store.comments),
  };
  return store;
}

it("reads both queues and reports a single connected label when nothing is waiting", async () => {
  const store = fakeStore();
  const { result } = renderHook(() => useSyncStatus({ store }));
  await waitFor(() => expect(store.queue).toHaveBeenCalled());
  expect(store.commentQueue).toHaveBeenCalled();
  expect(result.current).toEqual({ kind: "connected" });
});

it("counts todos and comments together and advances 1 of 9, 2 of 9 as they drain", async () => {
  vi.useFakeTimers();
  const store = fakeStore(entries("pending", 5), entries("pending", 4));
  const { result } = renderHook(() => useSyncStatus({ store, intervalMs: 1000 }));
  await act(() => vi.advanceTimersByTimeAsync(0));
  expect(result.current).toEqual({ kind: "syncing", current: 1, total: 9 });

  store.todos = entries("pending", 4);
  await act(() => vi.advanceTimersByTimeAsync(1000));
  expect(result.current).toEqual({ kind: "syncing", current: 2, total: 9 });

  store.todos = [];
  store.comments = [];
  await act(() => vi.advanceTimersByTimeAsync(1000));
  expect(result.current).toEqual({ kind: "connected" });
});

it("resets the peak once the queue empties, so the next batch counts from its own size", async () => {
  vi.useFakeTimers();
  const store = fakeStore(entries("pending", 9));
  const { result } = renderHook(() => useSyncStatus({ store, intervalMs: 1000 }));
  await act(() => vi.advanceTimersByTimeAsync(0));
  store.todos = [];
  await act(() => vi.advanceTimersByTimeAsync(1000));
  store.todos = entries("pending", 2);
  await act(() => vi.advanceTimersByTimeAsync(1000));
  expect(result.current).toEqual({ kind: "syncing", current: 1, total: 2 });
});

it("reports a failed entry as a sync error with its text", async () => {
  const store = fakeStore([{ state: "error", error: "Rejected: no project" }]);
  const { result } = renderHook(() => useSyncStatus({ store }));
  await waitFor(() => expect(result.current).toEqual({ kind: "sync-error", failed: 1, detail: "Rejected: no project" }));
});

it("lets a connection error outrank the queue", async () => {
  const store = fakeStore([{ state: "error", error: "boom" }]);
  const { result } = renderHook(() => useSyncStatus({ store, connectionError: "Gavel is unreachable" }));
  await waitFor(() => expect(store.queue).toHaveBeenCalled());
  expect(result.current).toEqual({ kind: "connection-error", detail: "Gavel is unreachable" });
});

it("reports connecting while the dashboard has not answered, ahead of the queue", async () => {
  const store = fakeStore([{ state: "error", error: "boom" }]);
  const { result } = renderHook(() => useSyncStatus({ store, connecting: true }));
  await waitFor(() => expect(store.queue).toHaveBeenCalled());
  expect(result.current).toEqual({ kind: "connecting" });
});

it("lets a connection error outrank connecting", async () => {
  const store = fakeStore();
  const { result } = renderHook(() => useSyncStatus({ store, connecting: true, connectionError: "Gavel is unreachable" }));
  await waitFor(() => expect(store.queue).toHaveBeenCalled());
  expect(result.current).toEqual({ kind: "connection-error", detail: "Gavel is unreachable" });
});

it("surfaces a queue read failure as a sync error instead of swallowing it", async () => {
  const store = fakeStore();
  store.queue.mockRejectedValue(new Error("Todo storage transaction aborted"));
  const { result } = renderHook(() => useSyncStatus({ store }));
  await waitFor(() => expect(result.current).toEqual({ kind: "sync-error", failed: 0, detail: "Todo storage transaction aborted" }));
});

it("stops polling when unmounted", async () => {
  vi.useFakeTimers();
  const store = fakeStore();
  const { unmount } = renderHook(() => useSyncStatus({ store, intervalMs: 1000 }));
  await act(() => vi.advanceTimersByTimeAsync(0));
  unmount();
  const reads = store.queue.mock.calls.length;
  await vi.advanceTimersByTimeAsync(5000);
  expect(store.queue).toHaveBeenCalledTimes(reads);
});

it("reads the real extension store: a saved capture and a saved comment show as syncing 1 of 2", async () => {
  const store = new TodoStore(new IDBFactory());
  await store.put({ id: "todo-1", title: "Review the sidebar", state: "pending", createdAt: 1 } as unknown as QueuedTodo);
  await store.putComment({ id: "comment-1", state: "pending", body: "Check the heading", savedAt: 2 } as unknown as QueuedComment);
  const { result } = renderHook(() => useSyncStatus({ store }));
  await waitFor(() => expect(result.current).toEqual({ kind: "syncing", current: 1, total: 2 }));
});
