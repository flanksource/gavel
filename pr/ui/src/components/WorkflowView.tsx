import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import type { WorkflowRun, Job } from '../types';
import { statusIcon, statusColor } from '../utils';
import { useNow } from '../useNow';
import { LogViewer } from './LogViewer';
import { UiChevronDown, UiChevronRight, UiLinkExternal } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../icons/Spinner';
import { jobLogsQuery } from './oneShotQueries';

export function formatDuration(start?: string, end?: string): string {
  if (!start) return '';
  const s = new Date(start);
  const e = end ? new Date(end) : new Date();
  const ms = e.getTime() - s.getTime();
  const secs = Math.floor(ms / 1000);
  if (!end) return `(running ${secs}s...)`;
  if (secs < 60) return `(${secs}s)`;
  return `(${Math.floor(secs / 60)}m ${secs % 60}s)`;
}

// JobDuration shows a job's elapsed time. A still-running job (no completedAt)
// renders a leaf that subscribes to the shared useNow() clock so its
// '(running Xs...)' counter advances each second without reconciling the job
// row; a completed job's duration is fixed, so it stays a plain static span.
function JobDuration({ startedAt, completedAt }: { startedAt?: string; completedAt?: string }) {
  if (completedAt) {
    const fixed = formatDuration(startedAt, completedAt);
    return fixed ? <span className="text-muted-foreground">{fixed}</span> : null;
  }
  return <RunningDuration startedAt={startedAt} />;
}

function RunningDuration({ startedAt }: { startedAt?: string }) {
  useNow();
  if (!startedAt) return null;
  return <span className="text-muted-foreground">{formatDuration(startedAt)}</span>;
}

function IndeterminateProgress() {
  return (
    <div className="ml-4 mt-1 mb-1">
      <div className="flex items-center gap-1.5 text-[10px] text-blue-600 mb-0.5">
        <Spinner />
        <span>Fetching logs from GitHub…</span>
      </div>
      <div className="h-1 w-full max-w-xs bg-blue-100 rounded overflow-hidden relative">
        <div className="gavel-progress-bar absolute inset-y-0 w-1/3 bg-blue-500 rounded" />
      </div>
    </div>
  );
}

export function runSummary(run: WorkflowRun): string {
  const jobs = run.jobs || [];
  if (jobs.length === 0) return '';
  const failed = jobs.filter(j => j.conclusion?.toLowerCase() === 'failure').length;
  if (failed > 0) return `${jobs.length} jobs, ${failed} failing`;
  return `${jobs.length} jobs`;
}

export function WorkflowRunView({ run, repo }: { run: WorkflowRun; repo: string }) {
  const isFailure = run.conclusion?.toLowerCase() === 'failure';
  const [expanded, setExpanded] = useState(isFailure);
  const summary = runSummary(run);
  const ChevronIcon = expanded ? UiChevronDown : UiChevronRight;

  return (
    <div className="mb-3">
      <div
        className="flex items-center gap-1.5 text-sm font-medium cursor-pointer hover:bg-muted rounded px-1 -mx-1 py-0.5"
        onClick={() => setExpanded(!expanded)}
      >
        <ChevronIcon className="text-muted-foreground text-[10px]" />
        <span className={statusColor(run.status, run.conclusion)}>
          {statusIcon(run.status, run.conclusion)}
        </span>
        <span>{run.name}</span>
        {summary && <span className="text-muted-foreground text-xs font-normal">· {summary}</span>}
        {run.url && (
          <a
            href={run.url}
            target="_blank"
            rel="noopener"
            className="text-muted-foreground hover:text-primary"
            onClick={e => e.stopPropagation()}
          >
            <UiLinkExternal className="text-xs" />
          </a>
        )}
      </div>
      {expanded && run.jobs && run.jobs.map(job => (
        <JobView key={job.databaseId} job={job} repo={repo} runId={run.databaseId} />
      ))}
    </div>
  );
}

function JobView({ job, repo, runId }: { job: Job; repo: string; runId: number }) {
  const failed = job.conclusion?.toLowerCase() === 'failure';

  const [logsRequested, setLogsRequested] = useState(false);
  const [expandedSteps, setExpandedSteps] = useState<Set<number>>(new Set());
  const [expandedJobFallback, setExpandedJobFallback] = useState(false);
  const logsResult = useQuery({
    ...jobLogsQuery(repo, runId, job.databaseId),
    enabled: logsRequested,
  });
  const loading = logsResult.isFetching;
  const loaded = logsResult.isSuccess;
  const jobLogs = logsResult.data?.logs || '';
  const stepLogs = new Map(
    (logsResult.data?.steps ?? [])
      .filter(step => !!step.logs)
      .map(step => [step.number, step.logs!]),
  );
  const error = logsResult.error instanceof Error ? logsResult.error.message : null;

  function toggleStep(stepNum: number) {
    setLogsRequested(true);
    setExpandedSteps(prev => {
      const next = new Set(prev);
      if (next.has(stepNum)) next.delete(stepNum);
      else next.add(stepNum);
      return next;
    });
  }

  function toggleJobFallback() {
    setLogsRequested(true);
    setExpandedJobFallback(v => !v);
  }

  const hasSteps = failed && job.steps && job.steps.some(s => s.conclusion?.toLowerCase() === 'failure');

  return (
    <div className="ml-4 mt-1">
      <div
        className={`flex items-center gap-1.5 text-xs ${failed && !hasSteps ? 'cursor-pointer hover:bg-muted rounded px-1 -mx-1' : ''}`}
        onClick={failed && !hasSteps ? toggleJobFallback : undefined}
      >
        <span className={statusColor(job.status, job.conclusion)}>
          {statusIcon(job.status, job.conclusion)}
        </span>
        <span className={failed ? 'text-red-700 font-medium' : 'text-foreground'}>{job.name}</span>
        <JobDuration startedAt={job.startedAt} completedAt={job.completedAt} />
        {job.url && (
          <a
            href={job.url}
            target="_blank"
            rel="noopener"
            className="text-muted-foreground hover:text-primary"
            onClick={e => e.stopPropagation()}
          >
            <UiLinkExternal className="text-[10px]" />
          </a>
        )}
      </div>
      {failed && !hasSteps && expandedJobFallback && loading && !loaded && <IndeterminateProgress />}
      {failed && job.steps && job.steps.map(step => {
        const stepFailed = step.conclusion?.toLowerCase() === 'failure';
        if (!stepFailed) return null;
        const isOpen = expandedSteps.has(step.number);
        const logs = stepLogs.get(step.number) || jobLogs;
        const isFallback = !stepLogs.get(step.number) && !!jobLogs;
        const ChevronIcon = isOpen ? UiChevronDown : UiChevronRight;
        return (
          <div key={step.number} className="ml-4 mt-0.5 text-xs">
            <div
              className="cursor-pointer hover:bg-muted rounded px-1 -mx-1 inline-flex items-center gap-1"
              onClick={() => toggleStep(step.number)}
            >
              <ChevronIcon className="text-muted-foreground text-[9px]" />
              <span className={statusColor(step.status, step.conclusion)}>
                {statusIcon(step.status, step.conclusion)}
              </span>
              <span className="text-red-600">{step.name}</span>
            </div>
            {isOpen && loading && !loaded && <IndeterminateProgress />}
            {isOpen && loaded && logs && (
              <>
                {isFallback && (
                  <div className="ml-4 mt-0.5 text-[10px] text-muted-foreground italic">Showing job log tail (step-level logs unavailable)</div>
                )}
                <LogViewer logs={logs} />
              </>
            )}
            {isOpen && loaded && !logs && !error && (
              <div className="ml-4 mt-0.5 text-[10px] text-muted-foreground">No logs captured for this step.</div>
            )}
            {isOpen && error && (
              <div className="ml-4 mt-0.5 text-[10px] text-red-500">{error}</div>
            )}
          </div>
        );
      })}
      {failed && !hasSteps && expandedJobFallback && loaded && jobLogs && (
        <LogViewer logs={jobLogs} />
      )}
      {failed && !hasSteps && expandedJobFallback && loaded && !jobLogs && !error && (
        <div className="ml-4 mt-0.5 text-[10px] text-muted-foreground">No logs captured for this job.</div>
      )}
      {failed && !hasSteps && expandedJobFallback && error && (
        <div className="ml-4 mt-0.5 text-[10px] text-red-500">{error}</div>
      )}
    </div>
  );
}
