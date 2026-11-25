export type SyncStatus =
  | { kind: "connected" }
  | { kind: "connecting" }
  | { kind: "syncing"; current: number; total: number }
  | { kind: "sync-error"; failed: number; detail?: string }
  | { kind: "connection-error"; detail: string };

/** One queued todo or comment, as far as the status cares. */
export type QueueEntry = { state: string; error?: string | undefined };

const FAILED_STATES = ["error", "uncertain", "run_error"];

/** Whether a queued entry needs attention: it recorded an error or ended in a failed state. */
export function entryFailed(entry: QueueEntry): boolean {
  return !!entry.error || FAILED_STATES.includes(entry.state);
}

export function syncStatus({ connectionError, connecting, entries, peak }: {
  /** Why the dashboard cannot be reached (unreachable, or Chrome withholding access); unset when it can. */
  connectionError?: string | undefined;
  /** The dashboard has not answered yet, so "connected" would be a claim nothing supports. */
  connecting?: boolean | undefined;
  entries: QueueEntry[];
  /** The most unsynced entries seen since the queue was last empty; the hook tracks it. */
  peak: number;
}): SyncStatus {
  if (connectionError) return { kind: "connection-error", detail: connectionError };
  if (connecting) return { kind: "connecting" };
  const unsynced = entries.filter(entry => entry.state !== "synced");
  const failures = unsynced.filter(entryFailed);
  if (failures.length) {
    const detail = failures.find(entry => entry.error)?.error;
    return { kind: "sync-error", failed: failures.length, ...(detail ? { detail } : {}) };
  }
  if (!unsynced.length) return { kind: "connected" };
  const total = Math.max(peak, unsynced.length);
  return { kind: "syncing", current: Math.min(total, Math.max(1, total - unsynced.length + 1)), total };
}
