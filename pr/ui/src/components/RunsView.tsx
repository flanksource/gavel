// The Runs tab of a hosted organization: its sandbox runs (push checks and agent runs) newest
// first, and the selected run's log.
import { useQuery } from '@tanstack/react-query';
import { AnsiHtml } from '@flanksource/clicky-ui/data';
import { fetchJSON, queryKeys, responseError } from '../query';
import { formatDuration } from './WorkflowView';

export interface SandboxRun {
  id: string;
  kind: string;
  status: string;
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
}

export const kindLabels: Record<string, string> = { 'push-check': 'Push check', agent: 'Agent run', test: 'Test', lint: 'Lint' };

const statusClass: Record<string, string> = {
  succeeded: 'text-green-600',
  failed: 'text-red-600',
  running: 'text-blue-600',
  queued: 'text-muted-foreground',
};

// While a run is queued or running the list refreshes, so its status and log follow it.
const activeRefreshMs = 5000;

export function runStatus(run: SandboxRun): string {
  return run.status === 'failed' && run.exitCode !== undefined ? `failed (exit ${run.exitCode})` : run.status;
}

async function fetchRunLog(id: string, signal: AbortSignal): Promise<string> {
  const response = await fetch(`/api/sandbox-runs/${encodeURIComponent(id)}/log`, { signal });
  const body = await response.text();
  if (response.ok) return body;
  let payload: unknown = body;
  try {
    payload = JSON.parse(body);
  } catch {
    // Not JSON: responseError reports the status.
  }
  throw new Error(responseError(payload, response.status));
}

// useSandboxRuns loads the organization's runs, refreshing while any is queued or running.
export function useSandboxRuns() {
  return useQuery({
    queryKey: queryKeys.sandboxRuns(),
    queryFn: ({ signal }) => fetchJSON<SandboxRun[]>({ url: '/api/sandbox-runs', signal, context: 'load sandbox runs' }),
    refetchInterval: query => (query.state.data ?? []).some(run => run.status === 'queued' || run.status === 'running') ? activeRefreshMs : false,
  });
}

export function RunsView({ selectedId, onSelect }: { selectedId: string; onSelect: (id: string) => void }) {
  const runs = useSandboxRuns();
  const selected = (runs.data ?? []).find(run => run.id === selectedId);
  const log = useQuery({
    queryKey: queryKeys.sandboxRunLog(selectedId),
    queryFn: ({ signal }) => fetchRunLog(selectedId, signal),
    enabled: selectedId !== '',
    retry: false,
  });

  return (
    <div className="flex h-full min-h-0 flex-col gap-3 overflow-y-auto p-4 text-sm">
      {runs.error && <p role="alert" className="text-red-600">{runs.error.message}</p>}
      {runs.data?.length === 0 && (
        <p className="text-muted-foreground">No sandbox runs yet. Pushes and hosted agent runs show up here.</p>
      )}
      {!!runs.data?.length && (
        <table className="w-full border-collapse text-left text-xs">
          <thead className="text-muted-foreground">
            <tr>
              {['Run', 'Status', 'Repository', 'Branch', 'Commit', 'Todo', 'By', 'Started', 'Took'].map(heading => (
                <th key={heading} className="border-b border-border px-2 py-1.5 font-medium">{heading}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {runs.data.map(run => (
              <tr
                key={run.id}
                onClick={() => onSelect(run.id)}
                aria-selected={run.id === selectedId}
                className={`cursor-pointer border-b border-border hover:bg-muted ${run.id === selectedId ? 'bg-primary/10' : ''}`}
              >
                <td className="px-2 py-1.5 font-medium">{kindLabels[run.kind] ?? run.kind}</td>
                <td className={`px-2 py-1.5 ${statusClass[run.status] ?? ''}`}>{runStatus(run)}</td>
                <td className="px-2 py-1.5">{run.repo}</td>
                <td className="px-2 py-1.5 font-mono">{run.ref?.replace(/^refs\/heads\//, '')}</td>
                <td className="px-2 py-1.5 font-mono">{run.sha?.slice(0, 7)}</td>
                <td className="px-2 py-1.5">{run.todoTitle}</td>
                <td className="px-2 py-1.5">{run.user}</td>
                <td className="px-2 py-1.5 text-muted-foreground">{new Date(run.startedAt ?? run.createdAt).toLocaleString()}</td>
                <td className="px-2 py-1.5 text-muted-foreground">{formatDuration(run.startedAt, run.finishedAt)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {selectedId && (
        <section className="flex min-h-0 flex-col gap-2">
          <h2 className="text-xs font-semibold uppercase text-muted-foreground">
            Log{selected ? ` · ${kindLabels[selected.kind] ?? selected.kind} ${selected.sha?.slice(0, 7) ?? ''}` : ''}
          </h2>
          {log.error && <p className="text-muted-foreground">{log.error.message}</p>}
          {log.data !== undefined && (
            <pre role="log" className="max-h-[60vh] overflow-auto rounded border border-border bg-muted/30 p-3 font-mono text-xs whitespace-pre-wrap">
              <AnsiHtml as="span" text={log.data} />
            </pre>
          )}
        </section>
      )}
    </div>
  );
}
