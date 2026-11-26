// The Runs tab of a hosted organization: its sandbox runs (push checks and agent runs) newest
// first in a list, and the selected run's details, why it failed, its time against its timeout,
// its log and its spec.
import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button, CopyButton, DetailEmptyState, SegmentedControl, SplitPane, Tabs } from '@flanksource/clicky-ui/components';
import { JsonView, LogViewer, ProgressBar, StatusDot, Timestamp } from '@flanksource/clicky-ui/data';
import { fetchJSON, queryKeys, responseError } from '../query';
import { useNow } from '../useNow';

export interface SandboxRun {
  id: string;
  kind: string;
  status: string;
  // statusMessage says why the run waits, or why it was refused or failed before it ran.
  statusMessage?: string;
  exitCode?: number;
  ref?: string;
  sha?: string;
  repo?: string;
  user?: string;
  todoId?: string;
  todoTitle?: string;
  backend: string;
  hasLog: boolean;
  createdAt: string;
  startedAt?: string;
  finishedAt?: string;
  // backendRef names the run's job in its backend.
  backendRef?: string;
  template?: string;
  cpuMillicores?: number;
  memoryBytes?: number;
  spec?: { args?: string[]; step?: string; pr?: string; [key: string]: unknown };
  // timeoutSeconds is the longest the run may work.
  timeoutSeconds?: number;
  // lost marks a run still recorded as running past its lifetime: its job can no longer report.
  lost?: boolean;
  // events are what happened to the run, oldest first.
  events?: RunEvent[];
  // phase is the status the run is shown with, which its events refine: waiting for worker,
  // starting, rejected, failed to assign or failed to start, else status (sandbox.RunPhase).
  phase?: string;
}

// RunEvent is one thing that happened to a run. source is gavel, whose kind is the status the run
// moved to or waiting/stopping, or substrate, whose kind is what it reported of the run's sandbox.
export interface RunEvent {
  at: string;
  source: string;
  kind: string;
  message: string;
}

export const kindLabels: Record<string, string> = { 'push-check': 'Push check', agent: 'Agent run', test: 'Test', lint: 'Lint' };

type DotStatus = 'success' | 'error' | 'warning' | 'info';

const statusDots: Record<string, DotStatus> = { succeeded: 'success', failed: 'error', provisioning: 'info', running: 'info', cancelled: 'warning' };

// A run is active while it waits for room, for its sandbox, or runs.
const activeStatuses = ['queued', 'provisioning', 'running'];

function isActive(run: SandboxRun): boolean {
  return !run.lost && activeStatuses.includes(run.status);
}

// While a run is queued or running the list refreshes, so its status and log follow it.
const activeRefreshMs = 5000;

export function runStatus(run: SandboxRun): string {
  if (run.lost) return 'lost';
  const status = run.phase || run.status;
  return run.status === 'failed' && run.exitCode !== undefined ? `${status} (exit ${run.exitCode})` : status;
}

// ranJob says whether a run's job ran. A run recorded before runs had events, whose history lacks
// the queued event that creates every run now, is taken to have (sandbox.RunPhase).
function ranJob(run: SandboxRun): boolean {
  const first = run.events?.[0];
  if (first?.source !== 'gavel' || first.kind !== 'queued') return true;
  return run.events!.some(event => event.source === 'gavel' && event.kind === 'running');
}

function runDot(run: SandboxRun): DotStatus | undefined {
  return run.lost ? 'warning' : statusDots[run.status];
}

// formatSeconds is a duration the way people read it: 45s, 12m 3s, 3h 20m, 3d 21h.
export function formatSeconds(total: number): string {
  const s = Math.max(0, Math.floor(total));
  if (s < 60) return `${s}s`;
  if (s < 3600) return `${Math.floor(s / 60)}m ${s % 60}s`;
  if (s < 86400) return `${Math.floor(s / 3600)}h ${Math.floor((s % 3600) / 60)}m`;
  return `${Math.floor(s / 86400)}d ${Math.floor((s % 86400) / 3600)}h`;
}

function seconds(from: string, to: string | number): number {
  return (new Date(to).getTime() - new Date(from).getTime()) / 1000;
}

function branch(ref?: string): string {
  return ref?.replace(/^refs\/heads\//, '') ?? '';
}

async function fetchRunLog(id: string, signal: AbortSignal): Promise<string> {
  const response = await fetch(`/api/sandbox-runs/${encodeURIComponent(id)}/log`, { signal });
  if (!response.ok) {
    const body = await response.text();
    let payload: unknown = body;
    try {
      payload = JSON.parse(body);
    } catch {
      // Not JSON: responseError reports the status.
    }
    throw new Error(responseError(payload, response.status));
  }
  return response.text();
}

// useSandboxRuns loads the organization's runs, refreshing while any is queued or running.
export function useSandboxRuns() {
  return useQuery({
    queryKey: queryKeys.sandboxRuns(),
    queryFn: ({ signal }) => fetchJSON<SandboxRun[]>({ url: '/api/sandbox-runs', signal, context: 'load sandbox runs' }),
    refetchInterval: query => (query.state.data ?? []).some(isActive) ? activeRefreshMs : false,
  });
}

type RunFilter = 'all' | 'active' | 'failed';

function matchesFilter(run: SandboxRun, filter: RunFilter): boolean {
  if (filter === 'active') return isActive(run);
  if (filter === 'failed') return run.status === 'failed' || !!run.lost;
  return true;
}

// Elapsed is a run's time so far: fixed once it ends, ticking while it runs.
function Elapsed({ run }: { run: SandboxRun }) {
  const now = useNow();
  if (run.lost) return <>lost</>;
  if (!run.startedAt) return <>{run.status === 'queued' ? `queued ${formatSeconds(seconds(run.createdAt, now))}` : ''}</>;
  return <>{formatSeconds(seconds(run.startedAt, run.finishedAt ?? now))}</>;
}

function RunList({ runs, selectedId, onSelect }: { runs: SandboxRun[]; selectedId: string; onSelect: (id: string) => void }) {
  const [filter, setFilter] = useState<RunFilter>('all');
  const shown = runs.filter(run => matchesFilter(run, filter));
  const step = (from: number, by: number) => {
    const next = shown[Math.min(shown.length - 1, Math.max(0, from + by))];
    if (!next) return;
    onSelect(next.id);
    document.querySelector<HTMLElement>(`[data-run-id="${next.id}"]`)?.focus();
  };

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="border-b border-border p-2">
        <SegmentedControl<RunFilter>
          aria-label="Show runs"
          size="sm"
          value={filter}
          onChange={setFilter}
          options={[
            { id: 'all', label: `All ${runs.length}` },
            { id: 'active', label: 'Active' },
            { id: 'failed', label: 'Failed' },
          ]}
        />
      </div>
      <div role="listbox" aria-label="Runs" className="min-h-0 flex-1 overflow-y-auto">
        {shown.length === 0 && <p className="p-3 text-xs text-muted-foreground">No runs match.</p>}
        {shown.map((run, i) => (
          <button
            key={run.id}
            type="button"
            role="option"
            data-run-id={run.id}
            aria-selected={run.id === selectedId}
            onClick={() => onSelect(run.id)}
            onKeyDown={e => {
              if (e.key === 'ArrowDown' || e.key === 'j') { e.preventDefault(); step(i, 1); }
              if (e.key === 'ArrowUp' || e.key === 'k') { e.preventDefault(); step(i, -1); }
            }}
            className={`grid w-full grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-x-2 border-b border-border px-3 py-2 text-left text-xs hover:bg-muted focus-visible:outline-2 focus-visible:outline-primary ${run.id === selectedId ? 'bg-primary/10' : ''}`}
          >
            {runDot(run) ? <StatusDot status={runDot(run)!} size="sm" /> : <span className="size-2 rounded-full bg-muted-foreground/40" />}
            <span className="truncate">
              <span className="font-medium">{kindLabels[run.kind] ?? run.kind}</span>{' '}
              <span className="text-muted-foreground">{run.repo}</span>
            </span>
            <span className={`tabular-nums ${run.lost ? 'text-amber-600' : 'text-muted-foreground'}`}><Elapsed run={run} /></span>
            <span className="col-start-2 col-end-4 truncate text-[11px] text-muted-foreground">
              <span className="font-mono">{branch(run.ref)}{run.sha ? `@${run.sha.slice(0, 7)}` : ''}</span>
              {' · '}{runStatus(run)}{' · '}<Timestamp value={run.createdAt} format="relative" />
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}

// RunTime shows a run's time queued and working against its timeout.
function RunTime({ run }: { run: SandboxRun }) {
  const now = useNow();
  const timeout = run.timeoutSeconds ?? 0;
  const queued = seconds(run.createdAt, run.startedAt ?? (run.status === 'queued' ? now : run.finishedAt ?? now));
  const worked = run.startedAt ? seconds(run.startedAt, run.finishedAt ?? now) : 0;
  const total = queued + Math.max(timeout, worked);
  if (total <= 0) return null;
  const color = run.lost ? 'bg-amber-500' : run.status === 'failed' ? 'bg-red-500' : run.status === 'succeeded' ? 'bg-green-500' : 'bg-blue-500';
  let caption = '';
  if (run.lost) caption = `Still recorded as running ${formatSeconds(worked - timeout)} past its ${formatSeconds(timeout)} timeout. Its job is gone and can no longer report how it ended.`;
  else if (run.status === 'provisioning') caption = `Waiting ${formatSeconds(worked)} for its sandbox to start. The job has not started yet.`;
  else if (run.status === 'running' && timeout) caption = `${formatSeconds(worked)} of ${formatSeconds(timeout)} used. Stopped at ${new Date(new Date(run.startedAt!).getTime() + timeout * 1000).toLocaleTimeString()} if still running.`;

  return (
    <div className="flex flex-col gap-1">
      <div className="flex justify-between text-xs text-muted-foreground">
        <span>Queued {formatSeconds(queued)}{run.startedAt ? `, ${ranJob(run) ? 'ran' : 'waited for its sandbox'} ${formatSeconds(worked)}` : ''}</span>
        {timeout > 0 && <span>Timeout {formatSeconds(timeout)}</span>}
      </div>
      <div className="relative">
        <ProgressBar
          label="Run time"
          total={total}
          formatValue={formatSeconds}
          segments={[
            { count: queued, color: 'bg-zinc-400', label: 'queued' },
            { count: worked, color, label: run.finishedAt ? 'ran' : run.status },
          ]}
        />
        {timeout > 0 && (
          <div title="Timeout" className="absolute -top-1 h-4 w-px bg-foreground/60" style={{ left: `${((queued + timeout) / total) * 100}%` }} />
        )}
      </div>
      {caption && <span className="text-xs text-muted-foreground">{caption}</span>}
    </div>
  );
}

const lostReason = 'Still recorded as running past its timeout and the time it had to report back: its job is gone and its record was never closed.';

// reason is the one line under a run's status that says why it waits or how it ended.
function reason(run: SandboxRun): string | undefined {
  if (run.lost) return lostReason;
  if (run.status === 'failed' || run.status === 'cancelled') {
    if (run.statusMessage) return run.statusMessage;
    return run.exitCode !== undefined ? `The job exited with code ${run.exitCode}; the errors are in its log.` : undefined;
  }
  if (run.status === 'queued') return run.statusMessage || undefined;
  return undefined;
}

function formatSize(run: SandboxRun): string {
  const parts = [
    run.cpuMillicores ? `${run.cpuMillicores / 1000} CPU` : '',
    run.memoryBytes ? `${+(run.memoryBytes / 2 ** 30).toFixed(1)} GiB` : '',
  ].filter(Boolean);
  return parts.join(' · ');
}

const errorLine = /\b(error|err|fail(ed|ure)?|panic|fatal)\b|^--- FAIL|exit status [1-9]/i;

function RunLog({ run }: { run: SandboxRun }) {
  const [errorsOnly, setErrorsOnly] = useState(false);
  const log = useQuery({
    queryKey: [...queryKeys.sandboxRunLog(run.id), run.status],
    queryFn: ({ signal }) => fetchRunLog(run.id, signal),
    retry: false,
  });
  const shown = useMemo(
    () => (errorsOnly && log.data ? log.data.split('\n').filter(line => errorLine.test(line)).join('\n') : log.data),
    [errorsOnly, log.data],
  );

  if (log.error) {
    return <p className="text-xs text-muted-foreground">{isActive(run) ? 'The log is kept once the run ends. ' : ''}{log.error.message}</p>;
  }
  if (log.data === undefined) return <p className="text-xs text-muted-foreground">Loading the log…</p>;
  const download = () => {
    const url = URL.createObjectURL(new Blob([log.data!], { type: 'text/plain' }));
    const link = Object.assign(document.createElement('a'), { href: url, download: `run-${run.id}.log` });
    link.click();
    URL.revokeObjectURL(url);
  };
  return (
    <div className="flex min-h-0 flex-col gap-2">
      <div className="flex items-center justify-between gap-2">
        <SegmentedControl
          aria-label="Log lines"
          size="sm"
          value={errorsOnly ? 'errors' : 'all'}
          onChange={id => setErrorsOnly(id === 'errors')}
          options={[{ id: 'all', label: 'All lines' }, { id: 'errors', label: 'Errors' }]}
        />
        <Button variant="outline" size="sm" onClick={download}>Download</Button>
      </div>
      <div role="log">
        {shown ? <LogViewer logs={shown} maxHeight="60vh" variant="terminal" /> : <p className="text-xs text-muted-foreground">No error lines in this log.</p>}
      </div>
    </div>
  );
}

const eventDots: Record<string, DotStatus> = { ...statusDots, waiting: 'info', stopping: 'warning', assigned: 'info' };

// RunEventChain shows what happened to a run as a chain of badges, oldest first, with the time
// between each; a run still active ends in the time since its last event. Clicking a badge shows
// its message under the chain; the run's reason line already says how it ended.
function RunEventChain({ run }: { run: SandboxRun }) {
  const now = useNow();
  const events = run.events ?? [];
  const [picked, setPicked] = useState<number>();
  if (events.length === 0) return null;
  const selected = picked !== undefined && picked < events.length ? picked : undefined;
  const event = selected === undefined ? undefined : events[selected];
  const gap = (from: string, to: string | number) => (
    <span className="flex items-center gap-0.5 px-0.5 text-[11px] tabular-nums text-muted-foreground">
      +{formatSeconds(seconds(from, to))}<span aria-hidden>›</span>
    </span>
  );
  return (
    <div className="flex flex-col gap-1.5">
      <ol aria-label="Run events" className="flex flex-wrap items-center gap-y-1 text-xs">
        {events.map((e, i) => (
          <li key={i} className="flex items-center">
            {i > 0 && gap(events[i - 1].at, e.at)}
            <button
              type="button"
              aria-pressed={i === selected}
              onClick={() => setPicked(i === selected ? undefined : i)}
              title={e.message}
              className={`flex items-center gap-1.5 rounded-full px-2 py-0.5 ring-1 ring-inset ring-border hover:bg-muted focus-visible:outline-2 focus-visible:outline-primary ${i === selected ? 'bg-muted' : ''}`}
            >
              {eventDots[e.kind] ? <StatusDot status={eventDots[e.kind]} size="sm" /> : <span className="size-2 rounded-full bg-muted-foreground/40" />}
              {e.kind}
              {e.source !== 'gavel' && <span className="text-[11px] text-muted-foreground">{e.source}</span>}
            </button>
          </li>
        ))}
        {isActive(run) && <li className="flex items-center">{gap(events[events.length - 1].at, now)}<span className="text-[11px] text-muted-foreground">now</span></li>}
      </ol>
      {event && <p className="text-xs text-muted-foreground">
        <span className="text-foreground">{event.kind}</span> <Timestamp value={event.at} format="short" />: <span className="break-words">{event.message || '—'}</span>
      </p>}
    </div>
  );
}

function RunDetail({ run }: { run: SandboxRun }) {
  const [tab, setTab] = useState('log');
  const why = reason(run);
  const args = run.spec?.args?.join(' ');

  return (
    <div className="flex flex-col gap-4 p-4 text-sm">
      <header className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            {runDot(run) && <StatusDot status={runDot(run)!} label={runStatus(run)} />}
            {!runDot(run) && <span className="text-muted-foreground">{runStatus(run)}</span>}
            <h2 className="font-medium">{kindLabels[run.kind] ?? run.kind} · {run.repo}</h2>
          </div>
          {why && <p title={why} className={`mt-0.5 truncate text-xs ${run.status === 'failed' || run.lost ? 'text-red-700' : 'text-muted-foreground'}`}>{why}</p>}
          <p className="mt-0.5 text-xs text-muted-foreground">
            <span className="font-mono">{branch(run.ref)}{run.sha ? `@${run.sha.slice(0, 7)}` : ''}</span>
            {run.user ? ` · by ${run.user}` : ''}
          </p>
        </div>
        <CopyButton value={run.id} label="Copy run ID" />
      </header>

      <RunEventChain run={run} />

      <RunTime run={run} />

      <dl className="grid grid-cols-2 gap-x-6 gap-y-2 text-xs sm:grid-cols-3 xl:grid-cols-4">
        {[
          { label: 'Created', value: <Timestamp value={run.createdAt} format="short" /> },
          { label: 'Started', value: run.startedAt ? <Timestamp value={run.startedAt} format="short" /> : '—' },
          { label: 'Finished', value: run.finishedAt ? <Timestamp value={run.finishedAt} format="short" /> : '—' },
          { label: 'Size', value: formatSize(run) },
          { label: 'Template', value: run.template, mono: true },
          { label: 'Backend', value: run.backend },
          { label: 'Job', value: run.backendRef, mono: true },
          { label: 'Args', value: args, mono: true },
          { label: 'Step', value: run.spec?.step },
          { label: 'Todo', value: run.todoTitle },
          { label: 'PR', value: run.spec?.pr && <a className="text-primary hover:underline" href={run.spec.pr} target="_blank" rel="noreferrer">{run.spec.pr.replace(/^https:\/\/github\.com\//, '')}</a> },
        ].filter(item => item.value).map(item => (
          <div key={item.label} className="min-w-0">
            <dt className="text-muted-foreground">{item.label}</dt>
            <dd className={`truncate ${item.mono ? 'font-mono' : ''}`} title={typeof item.value === 'string' ? item.value : undefined}>{item.value}</dd>
          </div>
        ))}
      </dl>

      <div className="flex flex-col gap-3">
        <Tabs tabs={[{ id: 'log', label: 'Log' }, { id: 'spec', label: 'Spec' }]} value={tab} onChange={setTab} />
        {tab === 'log' ? <RunLog run={run} /> : <JsonView data={run.spec ?? {}} defaultOpenDepth={2} />}
      </div>
    </div>
  );
}

export function RunsView({ selectedId, onSelect }: { selectedId: string; onSelect: (id: string) => void }) {
  const runs = useSandboxRuns();
  const selected = (runs.data ?? []).find(run => run.id === selectedId);

  if (runs.error) return <p role="alert" className="p-4 text-sm text-red-600">{runs.error.message}</p>;
  if (!runs.data) return <p className="p-4 text-sm text-muted-foreground">Loading runs…</p>;
  if (runs.data.length === 0) {
    return <p className="p-4 text-sm text-muted-foreground">No sandbox runs yet. Pushes and hosted agent runs show up here.</p>;
  }
  return (
    <SplitPane
      className="h-full min-h-0"
      defaultSplit={28}
      minLeft={18}
      minRight={40}
      leftClass="border-r border-border"
      left={<RunList runs={runs.data} selectedId={selectedId} onSelect={onSelect} />}
      right={selected
        ? <RunDetail key={selected.id} run={selected} />
        : <DetailEmptyState label={selectedId ? 'This run is not in the list' : 'Select a run'} description="Its status, timing, details and log show here." />}
    />
  );
}
