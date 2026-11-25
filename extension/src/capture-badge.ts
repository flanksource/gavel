import { entryFailed, type QueueEntry } from "./sync-status";

type BadgeTab = { id: number; active: boolean; error?: string };
type BadgeState = { text: string; color: string; title: string };

export function createCaptureBadge({ store, tabs }: {
  store: { queue: () => Promise<QueueEntry[]>; commentQueue: () => Promise<QueueEntry[]> };
  tabs: () => Promise<BadgeTab[]>;
}): () => Promise<void> {
  let applied = new Map<number | undefined, BadgeState>();
  let running: Promise<void> | undefined;
  let requested = false;

  async function refresh() {
    const [todos, comments, openTabs] = await Promise.all([store.queue(), store.commentQueue(), tabs()]);
    const queued = [...todos, ...comments].filter(item => item.state !== "synced");
    const failures = queued.filter(entryFailed).length;
    const count = queued.length > 99 ? "99+" : String(queued.length);
    const status = queued.length ? `Gavel — ${queued.length} queued${failures ? `, ${failures} with errors` : ""}` : "Gavel";
    const next = new Map<number | undefined, BadgeState>();
    const results = await Promise.allSettled([undefined, ...openTabs].map(async tab => {
      const badge = {
        text: queued.length ? `${count}${failures ? "!" : ""}` : tab?.error ? "!" : tab?.active ? "on" : "",
        color: failures || (!queued.length && tab?.error) ? "#dc2626" : "#2563eb",
        title: tab?.error ? `${status} · ${tab.error}` : tab?.active ? `${status} · Active on this tab` : status,
      };
      const previous = applied.get(tab?.id);
      if (previous?.text !== badge.text || previous.color !== badge.color || previous.title !== badge.title) {
        const target = tab ? { tabId: tab.id } : {};
        await Promise.all([
          chrome.action.setBadgeText({ ...target, text: badge.text }),
          chrome.action.setBadgeBackgroundColor({ ...target, color: badge.color }),
          chrome.action.setTitle({ ...target, title: badge.title }),
        ]);
      }
      next.set(tab?.id, badge);
    }));
    applied = next;
    const errors = results.flatMap(result => result.status === "rejected" ? [result.reason] : []);
    if (errors.length) throw new AggregateError(errors, "Could not update the Gavel toolbar badge");
  }

  return () => {
    requested = true;
    running ??= (async () => {
      while (requested) {
        requested = false;
        await refresh();
      }
    })().finally(() => { running = undefined; });
    return running;
  };
}
