import { useEffect, useMemo, useState } from 'preact/hooks';
import type { ComponentChildren } from 'preact';
import type { ProcessNode, RunMeta } from '../types';
import { formatBytes, processLabel, processStateColor, processStateIcon } from '../utils';
import { countGoroutinesByState, parseGoroutineDump, type ParsedGoroutine, type ParsedGoroutineFrame } from '../stacktrace';

interface Props {
  process: ProcessNode | null;
  collectBusy?: boolean;
  onCollectStack?: (pid: number) => void | Promise<void>;
  runMeta?: RunMeta;
}

export function DiagnosticsDetailPanel({ process, collectBusy, onCollectStack, runMeta }: Props) {
  const [search, setSearch] = useState('');
  const [selectedStates, setSelectedStates] = useState<Set<string>>(new Set());
  const [hideRuntimeOnly, setHideRuntimeOnly] = useState(true);
  const stack = process?.stack_capture;
  const parsed = useMemo(() => parseGoroutineDump(stack?.text || ''), [stack?.text]);
  const stateCounts = useMemo(() => countGoroutinesByState(parsed), [parsed]);
  const filtered = useMemo(() => {
    const needle = search.trim().toLowerCase();
    return parsed.filter(goroutine => {
      if (selectedStates.size > 0 && !selectedStates.has(goroutine.state)) return false;
      if (hideRuntimeOnly && goroutine.userFrameCount === 0) return false;
      if (needle && !goroutine.searchText.includes(needle)) return false;
      return true;
    });
  }, [parsed, search, selectedStates, hideRuntimeOnly]);

  useEffect(() => {
    setSearch('');
    setSelectedStates(new Set());
    setHideRuntimeOnly(true);
  }, [stack?.text, process?.pid]);

  if (!process) {
    return (
      <div class="flex items-center justify-center h-full text-gray-400 text-sm">
        <div class="text-center">
          <iconify-icon icon="codicon:server-process" class="text-4xl mb-2 block" />
          Select a process to view diagnostics
        </div>
      </div>
    );
  }

  return (
    <div class="h-full min-h-0 flex flex-col gap-3 p-4">
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <iconify-icon icon={process.is_root ? 'codicon:server-process' : 'codicon:debug-alt'} class="text-2xl text-blue-600 shrink-0" />
            <h2 class="text-lg font-bold text-gray-900 break-words">{processLabel(process)}</h2>
          </div>
          <div class="mt-1 flex items-center gap-2 flex-wrap text-xs text-gray-500">
            <span class="font-mono">pid {process.pid}</span>
            {process.ppid ? <span class="font-mono">ppid {process.ppid}</span> : null}
            {process.status ? (
              <span class={`inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 ${processStateColor(process.status)}`}>
                <iconify-icon icon={processStateIcon(process.status)} />
                {process.status}
              </span>
            ) : null}
          </div>
        </div>
      </div>

      {runMeta && (
        <Section title="Run">
          <div class="grid grid-cols-2 gap-2 text-sm">
            <Metric label={runMeta.kind === 'rerun' ? `Rerun #${runMeta.sequence}` : 'Initial run'} value={runMeta.started_at ? new Date(runMeta.started_at).toLocaleString() : 'Unavailable'} />
            <Metric label="Finished" value={runMeta.finished_at ? new Date(runMeta.finished_at).toLocaleString() : 'In progress'} />
          </div>
        </Section>
      )}

      {process.command && (
        <Section title="Command">
          <pre class="text-xs text-gray-700 whitespace-pre-wrap font-mono bg-blue-50 rounded p-2 break-all">
            {process.command}
          </pre>
        </Section>
      )}

      <Section title="Metrics">
        <div class="grid grid-cols-2 gap-2 text-sm">
          <Metric label="CPU" value={`${(process.cpu_percent || 0).toFixed(1)}%`} />
          <Metric label="RSS" value={formatBytes(process.rss)} />
          <Metric label="Virtual memory" value={formatBytes(process.vms)} />
          <Metric label="Open files" value={process.open_files !== undefined ? String(process.open_files) : 'Unavailable'} />
        </div>
      </Section>

      <Section title="Stack" grow>
        <div class="flex-1 min-h-0 overflow-hidden">
          {!stack?.text ? (
            <div class="h-full min-h-[14rem] flex flex-col justify-center gap-3 p-3">
              <div class="space-y-1.5">
                <div class={`h-3 w-40 rounded ${collectBusy ? 'animate-pulse bg-gray-200' : 'bg-gray-200/70'}`} />
                <div class={`h-3 w-full rounded ${collectBusy ? 'animate-pulse bg-gray-200' : 'bg-gray-200/70'}`} />
                <div class={`h-3 w-5/6 rounded ${collectBusy ? 'animate-pulse bg-gray-200' : 'bg-gray-200/70'}`} />
              </div>
              <div class="flex items-center justify-between gap-3">
                <div class="text-sm text-gray-500">
                  {stack?.error
                    ? stack.error
                    : collectBusy
                      ? 'Collecting stack trace...'
                      : 'No stack trace collected yet.'}
                </div>
                {onCollectStack && (
                  <button
                    class="shrink-0 text-[11px] px-2 py-1 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                    onClick={() => onCollectStack(process.pid)}
                    disabled={collectBusy}
                    title="Collect the latest stack trace"
                  >
                    <iconify-icon icon={collectBusy ? 'svg-spinners:ring-resize' : 'codicon:debug-alt-small'} />
                    {collectBusy ? 'Collecting...' : 'Collect stack trace'}
                  </button>
                )}
              </div>
            </div>
          ) : parsed.length === 0 ? (
            <div class="h-full min-h-0 flex flex-col">
              <div class="flex items-center justify-between gap-2 px-1 py-1 text-[11px] text-gray-500">
                <div class="flex items-center gap-2">
                  <span class={`px-2 py-0.5 rounded-full ${
                    stack.status === 'ready'
                      ? 'bg-green-100 text-green-700'
                      : stack.status === 'unsupported'
                        ? 'bg-yellow-100 text-yellow-700'
                        : 'bg-red-100 text-red-700'
                  }`}>
                    {stack.status}
                  </span>
                  {stack.collected_at && <span>{new Date(stack.collected_at).toLocaleString()}</span>}
                </div>
                {onCollectStack && (
                  <button
                    class="shrink-0 text-[11px] px-2 py-1 rounded border border-gray-200 text-gray-600 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                    onClick={() => onCollectStack(process.pid)}
                    disabled={collectBusy}
                    title="Refresh stack trace"
                  >
                    <iconify-icon icon={collectBusy ? 'svg-spinners:ring-resize' : 'codicon:refresh'} />
                    Refresh
                  </button>
                )}
              </div>
              {stack.error && (
                <div class="mt-2 text-xs text-red-700 whitespace-pre-wrap">
                  {stack.error}
                </div>
              )}
              <pre class="flex-1 min-h-0 overflow-auto py-1 text-[11px] text-gray-700 whitespace-pre-wrap font-mono leading-4">{stack.text}</pre>
            </div>
          ) : (
            <div class="h-full min-h-0 flex flex-col">
              <div class="px-1 py-1.5 space-y-2">
                <div class="flex items-center justify-between gap-2 text-[11px] text-gray-500 flex-wrap">
                  <div class="flex items-center gap-2 flex-wrap">
                    <span class={`px-2 py-0.5 rounded-full ${
                      stack.status === 'ready'
                        ? 'bg-green-100 text-green-700'
                        : stack.status === 'unsupported'
                          ? 'bg-yellow-100 text-yellow-700'
                          : 'bg-red-100 text-red-700'
                    }`}>
                      {stack.status}
                    </span>
                    {stack.collected_at && <span>{new Date(stack.collected_at).toLocaleString()}</span>}
                    <span>{filtered.length} / {parsed.length} goroutines</span>
                  </div>
                  {onCollectStack && (
                    <button
                      class="shrink-0 text-[11px] px-2 py-1 rounded border border-gray-200 text-gray-600 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                      onClick={() => onCollectStack(process.pid)}
                      disabled={collectBusy}
                      title="Refresh stack trace"
                    >
                      <iconify-icon icon={collectBusy ? 'svg-spinners:ring-resize' : 'codicon:refresh'} />
                      Refresh
                    </button>
                  )}
                </div>
                <div class="flex items-center gap-1.5 flex-wrap">
                  <div class="relative min-w-[14rem] flex-1">
                    <iconify-icon icon="codicon:search" class="absolute left-2 top-1/2 -translate-y-1/2 text-gray-400 text-xs" />
                    <input
                      class="w-full rounded-md border border-gray-200 bg-gray-50 py-1 pl-7 pr-2 text-xs text-gray-700 outline-none focus:border-blue-400 focus:bg-white"
                      placeholder="Filter by goroutine id, function, or file"
                      value={search}
                      onInput={(e) => setSearch((e.target as HTMLInputElement).value)}
                    />
                  </div>
                  <label class="inline-flex items-center gap-1.5 rounded-md border border-gray-200 bg-gray-50 px-2 py-1 text-[11px] text-gray-600">
                    <input
                      type="checkbox"
                      checked={hideRuntimeOnly}
                      onChange={(e) => setHideRuntimeOnly((e.target as HTMLInputElement).checked)}
                    />
                    Hide runtime-only
                  </label>
                  {(search || selectedStates.size > 0 || !hideRuntimeOnly) && (
                    <button
                      class="text-[11px] text-gray-500 hover:text-gray-700"
                      onClick={() => {
                        setSearch('');
                        setSelectedStates(new Set());
                        setHideRuntimeOnly(true);
                      }}
                    >
                      Clear
                    </button>
                  )}
                </div>
                <div class="flex items-center gap-1 flex-wrap">
                  {Array.from(stateCounts.entries()).sort((a, b) => b[1] - a[1]).map(([state, count]) => {
                    const active = selectedStates.has(state);
                    return (
                      <button
                        key={state}
                        class={`inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[11px] transition-colors ${
                          active
                            ? 'border-blue-400 bg-blue-50 text-blue-700'
                            : 'border-gray-200 bg-gray-50 text-gray-600 hover:bg-white'
                        }`}
                        onClick={() => {
                          setSelectedStates(prev => {
                            const next = new Set(prev);
                            if (next.has(state)) next.delete(state);
                            else next.add(state);
                            return next;
                          });
                        }}
                      >
                        <span class={`h-2 w-2 rounded-full ${goroutineStateDot(state)}`} />
                        {state}
                        <span class="text-[10px] opacity-70">{count}</span>
                      </button>
                    );
                  })}
                </div>
              </div>
              {stack.error && (
                <div class="mt-2 text-xs text-red-700 whitespace-pre-wrap">
                  {stack.error}
                </div>
              )}
              <div class="flex-1 min-h-0 overflow-auto py-1 space-y-1">
                {filtered.length === 0 && (
                  <div class="py-3 text-center text-xs text-gray-500">
                    No goroutines match the current filters.
                  </div>
                )}
                {filtered.map(goroutine => (
                  <GoroutineCard
                    key={goroutine.id}
                    goroutine={goroutine}
                    search={search}
                    hideRuntimeOnly={hideRuntimeOnly}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      </Section>
    </div>
  );
}

function Section({ title, children, grow }: { title: string; children: ComponentChildren; grow?: boolean }) {
  return (
    <section class={grow ? 'flex min-h-0 flex-1 flex-col' : ''}>
      <div class="text-[11px] font-semibold uppercase tracking-wide text-gray-500 mb-1.5">{title}</div>
      {children}
    </section>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div class="border rounded-lg bg-gray-50 px-2.5 py-2">
      <div class="text-[10px] uppercase tracking-wide text-gray-500">{label}</div>
      <div class="text-xs font-medium text-gray-800 mt-0.5">{value}</div>
    </div>
  );
}

function GoroutineCard({ goroutine, search, hideRuntimeOnly }: {
  goroutine: ParsedGoroutine;
  search: string;
  hideRuntimeOnly: boolean;
}) {
  const frames = hideRuntimeOnly
    ? goroutine.frames.filter(frame => !frame.runtime || frame.kind === 'created_by')
    : goroutine.frames;
  const defaultOpen = goroutine.state === 'running' || !!search;

  return (
    <details class="border-0 bg-transparent" open={defaultOpen}>
      <summary class="cursor-pointer list-none px-0 py-1">
        <div class="flex items-center gap-2 flex-wrap">
          <span class="font-mono text-xs font-semibold text-gray-900">g{goroutine.id}</span>
          <span class={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] ${goroutineStateBadge(goroutine.state)}`}>
            <span class={`h-2 w-2 rounded-full ${goroutineStateDot(goroutine.state)}`} />
            {goroutine.rawState}
          </span>
          <span class="text-[11px] text-gray-500">{frames.length}f</span>
          {goroutine.userFrameCount > 0 && <span class="text-[11px] text-gray-500">{goroutine.userFrameCount}u</span>}
          {goroutine.topFunction && <span class="truncate text-[11px] text-gray-600">{goroutine.topFunction}</span>}
        </div>
      </summary>
      <div class="pl-3 py-1 space-y-0.5">
        {frames.map((frame, index) => (
          <FrameRow key={`${goroutine.id}-${index}`} frame={frame} />
        ))}
      </div>
    </details>
  );
}

function FrameRow({ frame }: { frame: ParsedGoroutineFrame }) {
  return (
    <div class={`${frame.runtime ? 'text-gray-500' : 'text-gray-800'}`}>
      <div class="flex items-start gap-1.5">
        <iconify-icon icon={frame.kind === 'created_by' ? 'codicon:debug-restart' : frame.runtime ? 'codicon:debug-step-over' : 'codicon:symbol-method'} class="shrink-0 mt-0.5 text-[11px]" />
        <div class="min-w-0">
          <div class="break-all font-mono text-[11px] font-semibold leading-4">
            {frame.displayName}
            {frame.location && <span class="ml-2 text-[10px] font-normal opacity-80">{frame.location}</span>}
          </div>
        </div>
      </div>
    </div>
  );
}

function goroutineStateBadge(state: string): string {
  if (state.includes('running')) return 'bg-green-50 text-green-700';
  if (state.includes('chan') || state.includes('wait')) return 'bg-blue-50 text-blue-700';
  if (state.includes('sleep')) return 'bg-amber-50 text-amber-700';
  if (state.includes('select')) return 'bg-violet-50 text-violet-700';
  return 'bg-gray-100 text-gray-700';
}

function goroutineStateDot(state: string): string {
  if (state.includes('running')) return 'bg-green-500';
  if (state.includes('chan') || state.includes('wait')) return 'bg-blue-500';
  if (state.includes('sleep')) return 'bg-amber-500';
  if (state.includes('select')) return 'bg-violet-500';
  return 'bg-gray-400';
}
