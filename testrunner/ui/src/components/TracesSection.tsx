import { Component, useEffect, useState, type ReactNode } from 'react';
import { TraceView } from '@flanksource/clicky-ui/data';
import {
  fetchTraceKinds,
  fetchTraceRecords,
  openInQueryUrl,
  runTraces,
  traceExportUrl,
  type TraceExportFormat,
  type TraceRecord,
} from '../traces';
import type { TraceResult, TracesStatus } from '../types';

interface Props {
  windows: TraceResult[];
  // traces is status.traces: where the run's store is served, or why it is not.
  traces?: TracesStatus;
}

// One fixture's trace windows: what each captured, a link into the query app
// at exactly that window, its downloads and an inline viewer of its records.
export function TracesSection({ windows, traces }: Props) {
  // Every run's store serves the same kinds, so any served one lists the formats.
  const formats = useExportFormats(windows.map(window => runTraces(traces, window.run_id).mount).find(Boolean));
  const unavailable = [...new Set(windows.map(window => runTraces(traces, window.run_id))
    .filter(run => !run.mount).map(run => run.unavailable || "This run's trace store is not served, so its windows cannot be opened."))];
  return (
    <div>
      <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">Traces</h3>
      {unavailable.map(reason => (
        <div key={reason} className="mb-2 rounded border border-amber-200 bg-amber-50 px-2 py-1 text-xs text-amber-800">{reason}</div>
      ))}
      <div className="space-y-2">
        {windows.map(window => (
          <TraceRow key={`${window.run_id}:${window.kind}:${window.stream}:${window.afterSeq ?? 0}`} window={window}
            mount={runTraces(traces, window.run_id).mount} formats={formats.byKind[window.kind] ?? []} />
        ))}
      </div>
      {formats.error && <div className="mt-1 text-xs text-red-700">{formats.error}</div>}
    </div>
  );
}

function useExportFormats(mount: string | undefined) {
  const [byKind, setByKind] = useState<Record<string, TraceExportFormat[]>>({});
  const [error, setError] = useState<string | undefined>(undefined);
  useEffect(() => {
    if (!mount) return;
    fetchTraceKinds(mount)
      .then(kinds => setByKind(Object.fromEntries(kinds.map(kind => [kind.name, kind.exportFormats]))))
      .catch((e: unknown) => setError(e instanceof Error ? e.message : String(e)));
  }, [mount]);
  return { byKind, error };
}

function TraceRow({ window, mount, formats }: { window: TraceResult; mount?: string; formats: TraceExportFormat[] }) {
  const [open, setOpen] = useState(false);
  return (
    <div className="rounded border border-gray-200" data-trace-kind={window.kind}>
      <div className="flex flex-wrap items-center gap-2 px-2 py-1.5 text-sm">
        <span className="font-mono font-medium text-gray-800">{window.kind}</span>
        <span className="text-xs text-gray-500">{window.count} {window.count === 1 ? 'record' : 'records'}</span>
        <SummaryValues summary={window.summary} />
        <span className="flex-1" />
        {mount && (
          <>
            <a className="text-xs text-blue-600 hover:underline" href={openInQueryUrl(mount, window)} target="_blank" rel="noreferrer">
              Open in query
            </a>
            {formats.map(format => (
              <a key={format.name} className="text-xs text-blue-600 hover:underline" href={traceExportUrl(mount, window, format.name)} download>
                {format.name.toUpperCase()}
              </a>
            ))}
            <button className="text-xs text-blue-600 hover:underline" onClick={() => setOpen(!open)}>
              {open ? 'Hide' : 'View'}
            </button>
          </>
        )}
      </div>
      {window.error && <div className="px-2 pb-1.5 text-xs text-red-700 whitespace-pre-wrap">{window.error}</div>}
      {open && mount && <TraceWindowView mount={mount} window={window} />}
    </div>
  );
}

// SummaryValues shows the kind's summary of the window, less the count shown beside it.
function SummaryValues({ summary }: { summary?: Record<string, unknown> }) {
  const entries = Object.entries(summary ?? {}).filter(([key, value]) => key !== 'count' && isScalar(value)).slice(0, 6);
  return (
    <>
      {entries.map(([key, value]) => (
        <span key={key} className="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600">
          {key}: <span className="font-mono">{String(value)}</span>
        </span>
      ))}
    </>
  );
}

function isScalar(value: unknown): boolean {
  return typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean';
}

function TraceWindowView({ mount, window }: { mount: string; window: TraceResult }) {
  const [rows, setRows] = useState<TraceRecord[] | null>(null);
  const [error, setError] = useState<string | undefined>(undefined);
  useEffect(() => {
    const abort = new AbortController();
    fetchTraceRecords(mount, window, abort.signal)
      .then(setRows)
      .catch((e: unknown) => { if (!abort.signal.aborted) setError(e instanceof Error ? e.message : String(e)); });
    return () => abort.abort();
  }, [mount, window.kind, window.stream, window.afterSeq, window.toSeq]);
  if (error) return <div className="border-t px-2 py-1.5 text-xs text-red-700 whitespace-pre-wrap">{error}</div>;
  if (!rows) return <div className="border-t px-2 py-1.5 text-xs text-gray-400">Loading records...</div>;
  return (
    <div className="border-t" style={{ height: 420 }}>
      <ViewerBoundary>
        <TraceView kind={window.kind} rows={rows} />
      </ViewerBoundary>
    </div>
  );
}

// ViewerBoundary keeps a viewer that refuses its records (TraceView throws on a
// row its kind's viewer cannot render) from taking the whole panel down.
class ViewerBoundary extends Component<{ children: ReactNode }, { error?: string }> {
  state: { error?: string } = {};

  static getDerivedStateFromError(error: unknown) {
    return { error: error instanceof Error ? error.message : String(error) };
  }

  render() {
    if (this.state.error) return <div className="px-2 py-1.5 text-xs text-red-700">{this.state.error}</div>;
    return this.props.children;
  }
}
