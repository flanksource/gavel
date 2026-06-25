import type { TodoItem } from '../../types';
import { ageMs } from './todoGroup';

// TodoTimeRange holds the clicky-ui TimeRange tokens (e.g. `now-7d` / `now`) the
// user applied, persisted so the activity filter survives reloads. null means no
// range is set and every todo is shown.
export interface TodoTimeRange {
  from: string;
  to: string;
}

// ResolvedRange is a TodoTimeRange resolved to absolute epoch-millis bounds; a
// null bound is open-ended on that side.
export interface ResolvedRange {
  from: number | null;
  to: number | null;
}

const STORAGE_KEY = 'gavel.pr-ui.todoTimeRange.v1';

const UNIT_MS: Record<string, number> = {
  s: 1000,
  m: 60_000,
  h: 3_600_000,
  d: 86_400_000,
  w: 604_800_000,
  M: 2_592_000_000, // ~30 days
  y: 31_536_000_000, // ~365 days
};

// resolveTimeToken turns a clicky-ui TimeRange token into epoch millis. It
// accepts `now`, relative `now-<n><unit>` / `now+<n><unit>` expressions (unit one
// of s,m,h,d,w,M,y), and absolute ISO date/datetime strings. Returns null when
// the token cannot be parsed.
export function resolveTimeToken(token: string, now: number): number | null {
  const trimmed = token.trim();
  if (trimmed === '' || trimmed === 'now') return trimmed === 'now' ? now : null;
  const rel = /^now([+-])(\d+)([smhdwMy])$/.exec(trimmed);
  if (rel) {
    const [, sign, amount, unit] = rel;
    const delta = Number(amount) * UNIT_MS[unit];
    return sign === '-' ? now - delta : now + delta;
  }
  const abs = Date.parse(trimmed);
  return Number.isNaN(abs) ? null : abs;
}

// resolveRange resolves the stored tokens to absolute bounds, or null when no
// range is set (so callers can skip filtering entirely).
export function resolveRange(range: TodoTimeRange | null, now: number): ResolvedRange | null {
  if (!range) return null;
  return {
    from: resolveTimeToken(range.from, now),
    to: resolveTimeToken(range.to, now),
  };
}

// withinActivityRange reports whether a todo's age anchor (its created time, else
// last activity) falls inside the resolved bounds. A todo with no recorded
// timestamp never matches an active range — a time filter is an explicit request
// to narrow by activity, so undated todos drop out rather than leak through.
export function withinActivityRange(todo: TodoItem, range: ResolvedRange): boolean {
  const ms = ageMs(todo);
  if (ms === null) return false;
  if (range.from !== null && ms < range.from) return false;
  if (range.to !== null && ms > range.to) return false;
  return true;
}

export function loadTimeRange(): TodoTimeRange | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as TodoTimeRange;
    return parsed && typeof parsed.from === 'string' && typeof parsed.to === 'string' ? parsed : null;
  } catch {
    return null;
  }
}

export function saveTimeRange(range: TodoTimeRange | null): void {
  try {
    if (range) localStorage.setItem(STORAGE_KEY, JSON.stringify(range));
    else localStorage.removeItem(STORAGE_KEY);
  } catch {
    // best-effort: storage unavailable — skip persisting.
  }
}
