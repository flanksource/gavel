// Recent todos filed from the focused /todos/new window, remembered per source
// host: a React Grab capture from localhost:5173 offers the issues last filed
// from that app as one-click "add to existing" chips.

export interface RecentTodo {
  ref: string;
  shortId?: string;
  title: string;
  dir: string;
  at: number;
}

const RECENT_TODOS_KEY = 'gavel.pr-ui.todoNew.recent.v1';
export const RECENT_TODOS_LIMIT = 8;

type RecentTodoStore = Record<string, RecentTodo[]>;

// recentTodoHost is the page the window was opened from (the grabbed page's
// ?sourceUrl=), or gavel's own host when opened directly.
export function recentTodoHost(sourceUrl: string): string {
  if (sourceUrl) {
    try {
      return new URL(sourceUrl).host;
    } catch {
      // An unparseable sourceUrl identifies no host; file under gavel's own.
    }
  }
  return window.location.host;
}

// Storage is best-effort (private mode / disabled storage) like the rest of the
// UI's localStorage state: an unreadable store starts empty.
function readStore(): RecentTodoStore {
  try {
    const parsed = JSON.parse(localStorage.getItem(RECENT_TODOS_KEY) || '{}');
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed as RecentTodoStore : {};
  } catch {
    return {};
  }
}

export function loadRecentTodos(host: string): RecentTodo[] {
  const entries = readStore()[host];
  return Array.isArray(entries) ? entries : [];
}

// rememberRecentTodo puts entry first for host, replacing an older entry for the
// same todo, and keeps the newest RECENT_TODOS_LIMIT. Returns the new list.
export function rememberRecentTodo(host: string, entry: RecentTodo): RecentTodo[] {
  const store = readStore();
  const next = [entry, ...(store[host] ?? []).filter(e => e.ref !== entry.ref || e.dir !== entry.dir)]
    .slice(0, RECENT_TODOS_LIMIT);
  try {
    localStorage.setItem(RECENT_TODOS_KEY, JSON.stringify({ ...store, [host]: next }));
  } catch {
    // best-effort: storage unavailable — the chips just won't persist.
  }
  return next;
}
