import { describe, expect, it } from "vitest";
import { entryFailed, syncStatus, type QueueEntry, type SyncStatus } from "./sync-status";

const waiting = (count: number): QueueEntry[] => Array.from({ length: count }, () => ({ state: "pending" }));
const synced = (count: number): QueueEntry[] => Array.from({ length: count }, () => ({ state: "synced" }));

describe("entryFailed", () => {
  it.each<[string, QueueEntry, boolean]>([
    ["a pending entry", { state: "pending" }, false],
    ["a sending entry", { state: "sending" }, false],
    ["a synced entry", { state: "synced" }, false],
    ["an error entry", { state: "error" }, true],
    ["an uncertain entry", { state: "uncertain" }, true],
    ["a run_error entry", { state: "run_error" }, true],
    ["a pending entry that recorded an error", { state: "pending", error: "Server unavailable" }, true],
    ["an entry with an empty error text", { state: "pending", error: "" }, false],
  ])("treats %s as failed=%s", (_name, entry, failed) => {
    expect(entryFailed(entry)).toBe(failed);
  });
});

describe("syncStatus", () => {
  const cases: { name: string; options: Parameters<typeof syncStatus>[0]; want: SyncStatus }[] = [
    { name: "an empty queue is connected", options: { entries: [], peak: 0 }, want: { kind: "connected" } },
    { name: "a queue of only synced entries is connected", options: { entries: synced(3), peak: 3 }, want: { kind: "connected" } },
    {
      name: "a connection error wins over a healthy queue",
      options: { connectionError: "Gavel is unreachable", entries: [], peak: 0 },
      want: { kind: "connection-error", detail: "Gavel is unreachable" },
    },
    {
      name: "a connection error outranks a failed entry",
      options: { connectionError: "Chrome withheld access", entries: [{ state: "error", error: "boom" }], peak: 1 },
      want: { kind: "connection-error", detail: "Chrome withheld access" },
    },
    {
      name: "a dashboard that has not answered yet is connecting, even with a healthy queue",
      options: { connecting: true, entries: [], peak: 0 },
      want: { kind: "connecting" },
    },
    {
      name: "connecting outranks a failed entry and a waiting one",
      options: { connecting: true, entries: [{ state: "error", error: "boom" }, ...waiting(2)], peak: 3 },
      want: { kind: "connecting" },
    },
    {
      name: "a connection error outranks connecting",
      options: { connectionError: "Gavel is unreachable", connecting: true, entries: [], peak: 0 },
      want: { kind: "connection-error", detail: "Gavel is unreachable" },
    },
    {
      name: "connecting false leaves the queue-derived state",
      options: { connecting: false, entries: waiting(1), peak: 1 },
      want: { kind: "syncing", current: 1, total: 1 },
    },
    {
      name: "a failed entry reports sync-error with its text",
      options: { entries: [{ state: "pending" }, { state: "error", error: "Rejected: no project" }], peak: 2 },
      want: { kind: "sync-error", failed: 1, detail: "Rejected: no project" },
    },
    {
      name: "a failed entry without text reports sync-error without detail",
      options: { entries: [{ state: "uncertain" }, { state: "run_error" }], peak: 2 },
      want: { kind: "sync-error", failed: 2 },
    },
    {
      name: "the first failure supplies the detail",
      options: { entries: [{ state: "error", error: "first" }, { state: "error", error: "second" }], peak: 2 },
      want: { kind: "sync-error", failed: 2, detail: "first" },
    },
    {
      name: "a synced entry that carries an error is not a failure",
      options: { entries: [{ state: "synced", error: "stale" }], peak: 1 },
      want: { kind: "connected" },
    },
    {
      name: "a popup opened on nine waiting entries reads 1 of 9",
      options: { entries: waiting(9), peak: 9 },
      want: { kind: "syncing", current: 1, total: 9 },
    },
    {
      name: "progress advances as entries drain",
      options: { entries: waiting(8), peak: 9 },
      want: { kind: "syncing", current: 2, total: 9 },
    },
    {
      name: "the last waiting entry reads N of N",
      options: { entries: [...synced(8), ...waiting(1)], peak: 9 },
      want: { kind: "syncing", current: 9, total: 9 },
    },
    {
      name: "a peak smaller than the remaining entries grows the total",
      options: { entries: waiting(5), peak: 2 },
      want: { kind: "syncing", current: 1, total: 5 },
    },
    {
      name: "a zero peak still reads from 1",
      options: { entries: waiting(1), peak: 0 },
      want: { kind: "syncing", current: 1, total: 1 },
    },
  ];

  it.each(cases)("$name", ({ options, want }) => {
    expect(syncStatus(options)).toEqual(want);
  });
});
