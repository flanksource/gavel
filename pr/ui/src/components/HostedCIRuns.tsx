// A hosted repository's CI runs, as its GitHub Actions jobs upload them to /api/ci/results, and
// one run's results. A run still streaming is re-read every few seconds, so a member can watch it.
import { useQuery } from '@tanstack/react-query';
import { Button } from '@flanksource/clicky-ui/components';
import { fetchJSON } from '../query';
import type { Project } from '../types';
import { LogViewer } from './LogViewer';
import { TestRunResults } from './tests/TestRunResults';
import type { RunSnapshot } from './tests/types';

// CIRun mirrors the Go pr/ui.ciRunView.
interface CIRun {
  runId: string;
  sha?: string;
  started?: string;
  ended?: string;
  updated: string;
  passed: number;
  failed: number;
  skipped: number;
  total: number;
}

// livePollMs is how often a streaming run, and the list while one streams, is re-read. Jobs
// upload every 15s, so this sees each upload within a few seconds.
const livePollMs = 5_000;

// staleAfterMs is how long a run may go without an upload before it is shown as lost: a job
// uploads every 15s while it runs, so a run this quiet was most likely killed.
const staleAfterMs = 2 * 60_000;

function ciRunsURL(project: Project, runId?: string): string {
  const base = `/api/projects/${encodeURIComponent(project.name)}/ci-runs`;
  return runId ? `${base}/${encodeURIComponent(runId)}` : base;
}

function runState(run: CIRun, now: number): 'done' | 'running' | 'lost' {
  if (run.ended) return 'done';
  return now - new Date(run.updated).getTime() > staleAfterMs ? 'lost' : 'running';
}

const stateLabels = { done: '', running: 'running', lost: 'no upload for 2m+' } as const;

export function CIRunsList({ project, onSelect }: { project: Project; onSelect: (runId: string) => void }) {
  const runs = useQuery({
    queryKey: ['ci-runs', project.name],
    queryFn: ({ signal }) => fetchJSON<CIRun[]>({ url: ciRunsURL(project), signal, context: `load the CI runs of ${project.name}` }),
    refetchInterval: query => (query.state.data ?? []).some(run => runState(run, Date.now()) === 'running') ? livePollMs : 30_000,
  });
  const now = Date.now();
  return (
    <section className="flex flex-col gap-1">
      <h3 className="text-xs font-semibold uppercase text-muted-foreground">CI runs</h3>
      {runs.error && <p className="text-muted-foreground">{runs.error.message}</p>}
      {runs.data && runs.data.length === 0 && (
        <p className="text-muted-foreground">No CI runs yet. Set the gavel action's gavel-server input to this dashboard's URL.</p>
      )}
      <ul aria-label="CI runs" className="flex flex-col">
        {(runs.data ?? []).map(run => {
          const state = runState(run, now);
          return (
            <li key={run.runId}>
              <button type="button" onClick={() => onSelect(run.runId)} className="w-full rounded px-2 py-1 text-left text-xs hover:bg-muted">
                <span className={run.failed > 0 ? 'text-red-600' : ''}>{run.passed} passed, {run.failed} failed</span>
                {run.skipped > 0 && <span>, {run.skipped} skipped</span>}
                {state !== 'done' && <span className={state === 'lost' ? 'ml-2 text-amber-600' : 'ml-2 text-blue-600'}>{stateLabels[state]}</span>}
                <span className="ml-2 text-muted-foreground">
                  {run.sha && <span className="font-mono">{run.sha.slice(0, 7)} · </span>}
                  {new Date(run.started ?? run.updated).toLocaleString()}
                </span>
              </button>
            </li>
          );
        })}
      </ul>
    </section>
  );
}

export function CIRunView({ project, runId, onBack }: { project: Project; runId: string; onBack: () => void }) {
  const run = useQuery({
    queryKey: ['ci-runs', project.name, runId],
    queryFn: ({ signal }) => fetchJSON<RunSnapshot>({ url: ciRunsURL(project, runId), signal, context: `load CI run ${runId}` }),
    refetchInterval: query => query.state.data?.status?.running ? livePollMs : false,
  });
  const snapshot = run.data;
  const running = Boolean(snapshot?.status?.running);
  const logTail = (snapshot as { log_tail?: string } | undefined)?.log_tail ?? '';
  return (
    <div className="flex h-full min-h-0 flex-col text-sm">
      <div className="flex flex-wrap items-center gap-2 border-b border-border px-4 py-2">
        <Button type="button" variant="ghost" onClick={onBack}>← {project.name}</Button>
        <span className="font-mono text-xs">{runId}</span>
        {running && <span className="text-xs text-blue-600">running, refreshing every {livePollMs / 1000}s</span>}
      </div>
      {run.error && <p role="alert" className="p-4 text-red-600">{run.error.message}</p>}
      {!snapshot && !run.error && <p className="p-4 text-muted-foreground">Loading…</p>}
      {logTail && (
        <details className="border-b border-border px-4 py-2">
          <summary className="cursor-pointer text-xs font-semibold uppercase text-muted-foreground">Diagnostics</summary>
          <LogViewer logs={logTail} collapsedLines={40} />
        </details>
      )}
      {snapshot && (
        <div className="min-h-0 flex-1">
          <TestRunResults
            snapshot={snapshot}
            done={!running}
            runKey={runId}
            projectName={project.name}
            projectDir={project.dir}
            emptyMessage={running ? 'No test results uploaded yet.' : undefined}
          />
        </div>
      )}
    </div>
  );
}
