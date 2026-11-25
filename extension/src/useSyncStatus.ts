import { useEffect, useMemo, useState } from "react";
import { syncStatus, type QueueEntry, type SyncStatus } from "./sync-status";
import { TodoStore } from "./todos/store";

const POLL_INTERVAL_MS = 1500;

type QueueReader = { queue: () => Promise<QueueEntry[]>; commentQueue: () => Promise<QueueEntry[]> };
type Reading = { entries: QueueEntry[]; peak: number; readError?: string };

let extensionStore: TodoStore | undefined;

/** The popup shares the background worker's IndexedDB, so it reads the same queues the toolbar badge counts. */
function defaultStore(): QueueReader {
  extensionStore ??= new TodoStore();
  return extensionStore;
}

/**
 * The popup's one-line sync state. The background worker's change callback lives in another JS context, so the popup polls
 * the queues while mounted, as `useCommentQueue` does.
 */
export function useSyncStatus({ connectionError, connecting, store, intervalMs = POLL_INTERVAL_MS }: {
  connectionError?: string | undefined;
  connecting?: boolean | undefined;
  store?: QueueReader;
  intervalMs?: number;
}): SyncStatus {
  const reader = useMemo(() => store ?? defaultStore(), [store]);
  const [reading, setReading] = useState<Reading>({ entries: [], peak: 0 });

  useEffect(() => {
    let active = true;
    async function read() {
      let entries: QueueEntry[];
      try {
        const [todos, comments] = await Promise.all([reader.queue(), reader.commentQueue()]);
        entries = [...todos, ...comments];
      } catch (cause) {
        if (active) setReading(previous => ({ ...previous, readError: cause instanceof Error ? cause.message : String(cause) }));
        return;
      }
      const unsynced = entries.filter(entry => entry.state !== "synced").length;
      if (active) setReading(previous => ({ entries, peak: unsynced ? Math.max(previous.peak, unsynced) : 0 }));
    }
    void read();
    const timer = window.setInterval(() => void read(), intervalMs);
    return () => { active = false; window.clearInterval(timer); };
  }, [reader, intervalMs]);

  const { readError, ...queue } = reading;
  if (!connectionError && !connecting && readError) return { kind: "sync-error", failed: 0, detail: readError };
  return syncStatus({ connectionError, connecting, ...queue });
}
