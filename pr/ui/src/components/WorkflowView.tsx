import { useState } from 'preact/hooks';
import type { WorkflowRun, Job } from '../types';
import { statusIcon, statusColor } from '../utils';
import { LogViewer } from './LogViewer';

interface JobLogsResponse {
  jobId: number;
  logs?: string;
  steps?: { number: number; logs?: string }[];
  error?: string;
}

async function fetchJobLogs(repo: string, runId: number, jobId: number, tail = 100): Promise<JobLogsResponse> {
  const url = `/api/prs/job-logs?repo=${encodeURIComponent(repo)}&runId=${runId}&jobId=${jobId}&tail=${tail}`;
  const r = await fetch(url);
  if (!r.ok) throw new Error(`job-logs ${r.status}`);
  return r.json();
}

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

function IndeterminateProgress() {
  return (
    <div class="ml-4 mt-1 mb-1">
      <div class="flex items-center gap-1.5 text-[10px] text-blue-600 mb-0.5">
        <iconify-icon icon="svg-spinners:ring-resize" />
        <span>Fetching logs from GitHub…</span>
      </div>
      <div class="h-1 w-full max-w-xs bg-blue-100 rounded overflow-hidden relative">
        <div class="gavel-progress-bar absolute inset-y-0 w-1/3 bg-blue-500 rounded" />
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

  return (
    <div class="mb-3">
      <div
        class="flex items-center gap-1.5 text-sm font-medium cursor-pointer hover:bg-gray-50 rounded px-1 -mx-1 py-0.5"
        onClick={() => setExpanded(!expanded)}
      >
        <iconify-icon
          icon={expanded ? 'codicon:chevron-down' : 'codicon:chevron-right'}
          class="text-gray-400 text-[10px]"
        />
        <span class={statusColor(run.status, run.conclusion)}>
          {statusIcon(run.status, run.conclusion)}
        </span>
        <span>{run.name}</span>
        {summary && <span class="text-gray-400 text-xs font-normal">· {summary}</span>}
        {run.url && (
          <a
            href={run.url}
            target="_blank"
            rel="noopener"
            class="text-gray-400 hover:text-blue-500"
            onClick={e => e.stopPropagation()}
          >
            <iconify-icon icon="codicon:link-external" class="text-xs" />
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
  const duration = formatDuration(job.startedAt, job.completedAt);

  const [loading, setLoading] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const [stepLogs, setStepLogs] = useState<Map<number, string>>(new Map());
  const [jobLogs, setJobLogs] = useState<string>('');
  const [error, setError] = useState<string | null>(null);
  const [expandedSteps, setExpandedSteps] = useState<Set<number>>(new Set());
  const [expandedJobFallback, setExpandedJobFallback] = useState(false);

  async function ensureLogs() {
    if (loaded || loading) return;
    setLoading(true);
    setError(null);
    try {
      const resp = await fetchJobLogs(repo, runId, job.databaseId);
      if (resp.error) {
        setError(resp.error);
      } else {
        const m = new Map<number, string>();
        for (const s of resp.steps || []) {
          if (s.logs) m.set(s.number, s.logs);
        }
        setStepLogs(m);
        setJobLogs(resp.logs || '');
      }
      setLoaded(true);
    } catch (e) {
      setError(String(e));
      setLoaded(true);
    } finally {
      setLoading(false);
    }
  }

  async function toggleStep(stepNum: number) {
    await ensureLogs();
    setExpandedSteps(prev => {
      const next = new Set(prev);
      if (next.has(stepNum)) next.delete(stepNum);
      else next.add(stepNum);
      return next;
    });
  }

  async function toggleJobFallback() {
    await ensureLogs();
    setExpandedJobFallback(v => !v);
  }

  const hasSteps = failed && job.steps && job.steps.some(s => s.conclusion?.toLowerCase() === 'failure');

  return (
    <div class="ml-4 mt-1">
      <div
        class={`flex items-center gap-1.5 text-xs ${failed && !hasSteps ? 'cursor-pointer hover:bg-gray-50 rounded px-1 -mx-1' : ''}`}
        onClick={failed && !hasSteps ? toggleJobFallback : undefined}
      >
        <span class={statusColor(job.status, job.conclusion)}>
          {statusIcon(job.status, job.conclusion)}
        </span>
        <span class={failed ? 'text-red-700 font-medium' : 'text-gray-700'}>{job.name}</span>
        {duration && <span class="text-gray-400">{duration}</span>}
        {job.url && (
          <a
            href={job.url}
            target="_blank"
            rel="noopener"
            class="text-gray-400 hover:text-blue-500"
            onClick={e => e.stopPropagation()}
          >
            <iconify-icon icon="codicon:link-external" class="text-[10px]" />
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
        return (
          <div key={step.number} class="ml-4 mt-0.5 text-xs">
            <div
              class="cursor-pointer hover:bg-gray-50 rounded px-1 -mx-1 inline-flex items-center gap-1"
              onClick={() => toggleStep(step.number)}
            >
              <iconify-icon
                icon={isOpen ? 'codicon:chevron-down' : 'codicon:chevron-right'}
                class="text-gray-400 text-[9px]"
              />
              <span class={statusColor(step.status, step.conclusion)}>
                {statusIcon(step.status, step.conclusion)}
              </span>
              <span class="text-red-600">{step.name}</span>
            </div>
            {isOpen && loading && !loaded && <IndeterminateProgress />}
            {isOpen && loaded && logs && (
              <>
                {isFallback && (
                  <div class="ml-4 mt-0.5 text-[10px] text-gray-400 italic">Showing job log tail (step-level logs unavailable)</div>
                )}
                <LogViewer logs={logs} bgClass={isFallback ? 'bg-red-50' : 'bg-gray-50'} borderClass={isFallback ? 'border-red-100' : 'border-gray-100'} />
              </>
            )}
            {isOpen && loaded && !logs && !error && (
              <div class="ml-4 mt-0.5 text-[10px] text-gray-400">No logs captured for this step.</div>
            )}
            {isOpen && error && (
              <div class="ml-4 mt-0.5 text-[10px] text-red-500">Failed to load logs: {error}</div>
            )}
          </div>
        );
      })}
      {failed && !hasSteps && expandedJobFallback && loaded && jobLogs && (
        <LogViewer logs={jobLogs} bgClass="bg-red-50" borderClass="border-red-100" />
      )}
      {failed && !hasSteps && expandedJobFallback && loaded && !jobLogs && !error && (
        <div class="ml-4 mt-0.5 text-[10px] text-gray-400">No logs captured for this job.</div>
      )}
      {failed && !hasSteps && expandedJobFallback && error && (
        <div class="ml-4 mt-0.5 text-[10px] text-red-500">Failed to load logs: {error}</div>
      )}
    </div>
  );
}
