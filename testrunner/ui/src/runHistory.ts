import { apiUrl } from './config';

/** One recorded run from `/api/runs` (snapshots.RunInfo plus the URL of its view). */
export interface RunInfo {
  runId: string;
  path: string;
  started?: string;
  ended?: string;
  kind: 'test' | 'lint' | 'test+lint';
  repo?: string;
  sha?: string;
  frameworks?: string[];
  passed: number;
  failed: number;
  skipped: number;
  warned: number;
  total: number;
  lintViolations: number;
  lintLinters: number;
  href: string;
}

/** One page of `/api/runs`. */
export interface RunsPage {
  /** The run this page is viewing; absent for the default (latest) view. */
  current?: string;
  /** The runs merged into the default view at latestHref. */
  latest?: string[];
  latestHref: string;
  runs: RunInfo[];
  /** Cursor for the following page; absent when none follows. */
  next?: string;
}

/**
 * Fetches a page of the run history. Resolves null when the host serves no
 * history (a live `gavel test --ui` run, or a page embedded in another app).
 */
export async function fetchRunsPage({ before, signal }: { before?: string; signal?: AbortSignal } = {}): Promise<RunsPage | null> {
  const query = before ? `?before=${encodeURIComponent(before)}` : '';
  const res = await fetch(apiUrl(`/api/runs${query}`), signal ? { signal } : {});
  if (res.status === 404) return null;
  if (!res.ok) throw new Error(`GET /api/runs: ${res.status} ${(await res.text()).trim()}`);
  return (await res.json()) as RunsPage;
}

/** The URL that opens a run on the tab holding its results: Lint for a lint-only run, Tests otherwise. */
export function runViewHref(run: Pick<RunInfo, 'href' | 'kind'>): string {
  return run.kind === 'lint' ? `${run.href}lint` : run.href;
}

/** "Oct 22, 15:31" — the run's start in the viewer's locale, or its id when unknown. */
export function runStartLabel(run: Pick<RunInfo, 'runId' | 'started'>): string {
  if (!run.started) return run.runId;
  return new Date(run.started).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}

/** A run's results in a few words: "2 failed / 40 tests · 3 lint". */
export function runResultLabel(run: RunInfo): string {
  const parts: string[] = [];
  if (run.total > 0) parts.push(run.failed > 0 ? `${run.failed} failed / ${run.total} tests` : `${run.total} tests`);
  if (run.kind !== 'test') parts.push(`${run.lintViolations} lint`);
  return parts.join(' · ') || 'no results';
}

export function runDurationMs(run: Pick<RunInfo, 'started' | 'ended'>): number | undefined {
  if (!run.started || !run.ended) return undefined;
  const ms = Date.parse(run.ended) - Date.parse(run.started);
  return ms >= 0 ? ms : undefined;
}
