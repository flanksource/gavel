import type { RunTraces, TraceRef, TracesStatus } from './types';

// The run's trace store is served through commons-db query at status.traces.mount:
// its result profiles, trace exports and session checkpoints under /api/
// (traceserve.Handler), and the query app over them where gavel's serve wiring
// mounts it (queryapp.Serve), all under that mount. A mount is a URL path as the browser
// reaches it: rebaseTraces puts the page's base path (and a MultiServer run id)
// in front of what the server reports relative to itself.

export type TraceRecord = Record<string, unknown>;

export interface TraceExportFormat {
  name: string;
  ext: string;
  mime: string;
}

export interface TraceKindInfo {
  name: string;
  title: string;
  exportFormats: TraceExportFormat[];
}

// TraceStreamNotFound is a window whose stream has committed no record yet
// (or does not exist): the export answers 404.
export class TraceStreamNotFound extends Error {}

/** Run runId's store as status reaches it: its mount, or why it has none. */
export function runTraces(status: TracesStatus | undefined, runId: string): RunTraces {
  const run = status?.runs[runId];
  if (run) return run;
  return { unavailable: `The trace store of run ${runId || '(unnamed)'} is not served by this view.` };
}

/** The mount of the store holding the Processes tab's capture; undefined when not served. */
export function processMount(status: TracesStatus | undefined): string | undefined {
  if (!status?.process_run || !status.process) return undefined;
  return status.runs[status.process_run]?.mount;
}

/**
 * status with every mount under base: a host that serves the runner's server
 * under a prefix of its own (the page's base path, a MultiServer run id, or
 * pr/ui's /api/project-runs/<id>) rebases the mounts the server reports
 * relative to itself.
 */
export function rebaseTraces(status: TracesStatus | undefined, base: string): TracesStatus | undefined {
  if (!status) return undefined;
  if (!base) return status;
  const runs = Object.fromEntries(Object.entries(status.runs).map(([id, run]) => [id, run.mount ? { ...run, mount: `${base}${run.mount}` } : run]));
  return { ...status, runs };
}

export function queryUrl(mount: string, path: string): string {
  return `${mount}${path}`;
}

function windowQuery(ref: TraceRef, extra: Record<string, string> = {}): string {
  const params = new URLSearchParams({ stream: ref.stream });
  if (ref.afterSeq) params.set('afterSeq', String(ref.afterSeq));
  if (ref.toSeq) params.set('toSeq', String(ref.toSeq));
  for (const [key, value] of Object.entries(extra)) params.set(key, value);
  return params.toString();
}

/** The download of ref's window in one of its kind's export formats. */
export function traceExportUrl(mount: string, ref: TraceRef, format: string): string {
  return queryUrl(mount, `/api/v1/traces/${encodeURIComponent(ref.kind)}/export?${windowQuery(ref, { format })}`);
}

/** The query app's page for ref's window: its kind's result profile, filtered to the window. */
export function openInQueryUrl(mount: string, ref: TraceRef): string {
  return queryUrl(mount, `/traces/${encodeURIComponent(ref.kind)}?${windowQuery(ref)}`);
}

export async function failure(res: Response, what: string): Promise<Error> {
  const text = (await res.text()).trim();
  return new Error(`${what} failed (${res.status})${text ? `: ${text}` : ''}`);
}

/** Parses an NDJSON body: one record per non-empty line. */
export function parseNdjson(body: string): TraceRecord[] {
  return body.split('\n').filter(line => line.trim() !== '').map(line => JSON.parse(line) as TraceRecord);
}

/**
 * The records of ref's window as its stream holds them: the NDJSON export.
 * The result profile's JSON rows leave out the columns a kind hides (a
 * process's key and stack, a terminal event's data), which the viewers need.
 */
export async function fetchTraceRecords(mount: string, ref: TraceRef, signal?: AbortSignal): Promise<TraceRecord[]> {
  const res = await fetch(traceExportUrl(mount, ref, 'ndjson'), signal ? { signal } : undefined);
  if (res.status === 404) throw new TraceStreamNotFound(`${ref.kind} stream ${ref.stream} has no records`);
  if (!res.ok) throw await failure(res, `export ${ref.kind} window`);
  return parseNdjson(await res.text());
}

/**
 * The id the query server surfaces a kind's result profile (`traces/<kind>`)
 * under: commons-db `profiles.SurfaceKey`, `profile-` plus the slugged name.
 */
export function traceProfileId(kind: string): string {
  const slug = `traces/${kind}`.toLowerCase().replace(/[ _/.-]/g, '-').replace(/[^a-z0-9-]/g, '');
  return `profile-${slug.replace(/^-+|-+$/g, '')}`;
}

/** The highest seq ref's stream has committed, read from its result profile; 0 when it has none. */
export async function fetchLatestSeq(mount: string, ref: TraceRef): Promise<number> {
  const query = new URLSearchParams({ stream: ref.stream, sort: 'seq', order: 'desc', limit: '1' });
  const res = await fetch(queryUrl(mount, `/api/v1/profile/${traceProfileId(ref.kind)}?${query}`), { headers: { Accept: 'application/json' } });
  if (!res.ok) throw await failure(res, `read ${ref.kind} stream`);
  const rows = await res.json() as Array<{ seq?: number }>;
  return rows[0]?.seq ?? 0;
}

/** The trace kinds the store serves, with the formats each exports in. */
export async function fetchTraceKinds(mount: string): Promise<TraceKindInfo[]> {
  const res = await fetch(queryUrl(mount, '/api/v1/traces/kinds'), { headers: { Accept: 'application/json' } });
  if (!res.ok) throw await failure(res, 'list trace kinds');
  return res.json() as Promise<TraceKindInfo[]>;
}

export interface CheckpointRequest {
  mount: string;
  session: string;
  /**
   * The kind's checkpoint params: `{key, transport, reason}` takes a stack dump
   * of one process on a `stacks` session, `{key}` forces a full facet reading
   * of one process on a `process_samples` session.
   */
  params: Record<string, string>;
}

/** Has the running session take a reading now, as its kind's checkpoint params describe it. */
export async function checkpointSession({ mount, session, params }: CheckpointRequest): Promise<void> {
  const res = await fetch(queryUrl(mount, `/api/v1/sessions/${encodeURIComponent(session)}/checkpoint`), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  });
  if (!res.ok) throw await failure(res, 'checkpoint');
}
