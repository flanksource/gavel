import { useCallback, useEffect, useMemo, useState } from 'react';
import { TraceWindowList, type TraceWindow, type TraceWindowSource } from '@flanksource/clicky-ui/data';
import {
  fetchTraceKinds,
  fetchTraceRecords,
  openInQueryUrl,
  runTraces,
  traceExportUrl,
  type TraceExportFormat,
} from '../traces';
import type { TraceResult, TracesStatus } from '../types';

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

/**
 * One fixture's trace windows, opened in the stores of the runs they name.
 * `traces` is status.traces with its mounts as the browser reaches them
 * (rebaseTraces), so every link and read goes through the run's own server.
 */
export function TracesPanel({ windows, traces }: { windows: TraceResult[]; traces: TracesStatus | undefined }) {
  // Every run's store serves the same kinds, so any served one lists the formats.
  const formats = useExportFormats(windows.map(window => runTraces(traces, window.run_id).mount).find(Boolean));
  const resolve = useCallback((window: TraceWindow): TraceWindowSource => runTraces(traces, window.run_id), [traces]);
  const mountOf = useCallback((window: TraceWindow): string => {
    const mount = runTraces(traces, window.run_id).mount;
    if (!mount) throw new Error(`trace store of run ${window.run_id} is not served`);
    return mount;
  }, [traces]);
  const props = useMemo(() => ({
    formats: (kind: string) => formats.byKind[kind] ?? [],
    fetchRecords: (window: TraceWindow, signal: AbortSignal) => fetchTraceRecords(mountOf(window), window, signal),
    exportUrl: (window: TraceWindow, format: string) => traceExportUrl(mountOf(window), window, format),
    openUrl: (window: TraceWindow) => openInQueryUrl(mountOf(window), window),
  }), [formats.byKind, mountOf]);

  return (
    <div className="space-y-density-3 p-density-4">
      <TraceWindowList
        windows={windows}
        resolve={resolve}
        {...props}
        {...(formats.error ? { error: formats.error } : {})}
      />
    </div>
  );
}
