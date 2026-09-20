import type { ProcStatus, Project, TodoItem, TodoStatus } from '../../types';

// Pure helpers behind the focused /todos/new form: reading its query params,
// resolving where it returns to, detecting the project from the source page,
// and ordering the existing-issue picker.

// firstParam reads the first present query value across a set of aliases so
// external callers can use the field name they have (e.g. ?body= or ?text=).
export function firstParam(params: URLSearchParams, ...keys: string[]): string {
  for (const key of keys) {
    const value = params.get(key);
    if (value !== null && value.trim() !== '') return value.trim();
  }
  return '';
}

// listParam collects every value of every alias, splitting on commas, so both
// ?labels=a&labels=b and ?labels=a,b work — the same rule the server applies.
export function listParam(params: URLSearchParams, ...keys: string[]): string[] {
  const out: string[] = [];
  for (const key of keys) {
    for (const value of params.getAll(key)) {
      for (const item of value.split(',')) {
        const trimmed = item.trim();
        if (trimmed && !out.includes(trimmed)) out.push(trimmed);
      }
    }
  }
  return out;
}

export function parseBool(value: string): boolean {
  return /^(1|true|yes|on)$/i.test(value.trim());
}

// returnTarget resolves where the form returns to after a create or cancel: an
// explicit ?return= path wins, otherwise the (same-origin) referer that opened
// the page. Cross-origin referers and the new-todo page itself are ignored so a
// hard-load or external link falls through to the caller's default.
export function returnTarget(params: URLSearchParams): string | null {
  const candidate = firstParam(params, 'return', 'returnTo') || (typeof document !== 'undefined' ? document.referrer : '');
  if (!candidate) return null;
  try {
    const url = new URL(candidate, window.location.origin);
    if (url.origin !== window.location.origin) return null;
    if (url.pathname.startsWith('/todos/new')) return null;
    return `${url.pathname}${url.search}${url.hash}`;
  } catch {
    return null;
  }
}

export function oneOf<T extends string>(value: string, allowed: readonly T[], fallback: T): T {
  return (allowed as readonly string[]).includes(value) ? (value as T) : fallback;
}

export type TodoNewMode = 'new' | 'existing';

export const closedStatuses = new Set<TodoStatus>(['completed', 'skipped']);

export function modeFromParams(params: URLSearchParams): TodoNewMode {
  const mode = firstParam(params, 'mode', 'target');
  if (mode === 'existing' || firstParam(params, 'ref', 'todo', 'issue')) return 'existing';
  return 'new';
}

function sourcePort(raw: string): number | null {
  if (!raw) return null;
  try {
    const url = new URL(raw);
    if (url.port) return Number(url.port);
    if (url.protocol === 'http:') return 80;
    if (url.protocol === 'https:') return 443;
  } catch {
    return null;
  }
  return null;
}

function statusForProject(project: Project, procStatus: Record<string, ProcStatus>): ProcStatus | undefined {
  if (procStatus[project.name]) return procStatus[project.name];
  for (const repo of project.repos || []) {
    if (procStatus[repo]) return procStatus[repo];
  }
  return undefined;
}

export function detectProjectDir(sourceUrl: string, workspaces: Project[], procStatus: Record<string, ProcStatus>): string {
  const port = sourcePort(sourceUrl);
  if (!port) return '';
  for (const workspace of workspaces) {
    const status = statusForProject(workspace, procStatus);
    if (status?.processes?.some(proc => proc.ports?.includes(port))) return workspace.dir;
  }
  return '';
}

export function todoMatchesSearch(todo: TodoItem, query: string): boolean {
  const q = query.trim().toLowerCase();
  if (!q) return true;
  return todo.title.toLowerCase().includes(q) ||
    todo.ref.toLowerCase().includes(q) ||
    (todo.shortId || '').toLowerCase().includes(q) ||
    (todo.id || '').toLowerCase().includes(q);
}

function todoActivityMs(todo: TodoItem): number | null {
  const raw = todo.lastRun || todo.created;
  if (!raw) return null;
  const ms = Date.parse(raw);
  return Number.isNaN(ms) ? null : ms;
}

export function compareRecentTodos(a: TodoItem, b: TodoItem): number {
  const am = todoActivityMs(a);
  const bm = todoActivityMs(b);
  if (am !== bm) {
    if (am === null) return 1;
    if (bm === null) return -1;
    return bm - am;
  }
  return (a.title || '').localeCompare(b.title || '');
}
