import { useState, useMemo } from 'preact/hooks';
import type { PRItem, PRDetail, WorkflowRun, Job, PRComment } from '../types';
import { stateColor, reviewColor, timeAgo, statusIcon, statusColor, severityIcon } from '../utils';
import { Markdown } from './Markdown';

interface Props {
  pr: PRItem;
  detail: PRDetail | null;
  loading: boolean;
}

export function PRDetailPanel({ pr, detail, loading }: Props) {
  return (
    <div class="p-4 bg-white h-full overflow-y-auto">
      <PRHeader pr={pr} detail={detail} />

      {loading && (
        <div class="flex items-center gap-2 text-sm text-gray-400 mt-4">
          <iconify-icon icon="svg-spinners:ring-resize" class="text-blue-500" />
          Loading details...
        </div>
      )}

      {detail?.error && (
        <div class="mt-3 p-2 bg-red-50 border border-red-100 rounded text-xs text-red-700">
          <iconify-icon icon="codicon:error" class="mr-1" />
          {detail.error}
        </div>
      )}

      {detail?.runs && Object.keys(detail.runs).length > 0 && (
        <Section title="Workflows">
          {Object.values(detail.runs).map(run => (
            <WorkflowRunView key={run.databaseId} run={run} />
          ))}
        </Section>
      )}

      {detail?.comments && detail.comments.length > 0 && (
        <CommentsSection comments={detail.comments} />
      )}

      <div class="pt-3 mt-3 border-t border-gray-100">
        <a href={pr.url} target="_blank" rel="noopener"
          class="inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800 hover:underline">
          <iconify-icon icon="codicon:link-external" />
          Open on GitHub
        </a>
      </div>
    </div>
  );
}

function PRHeader({ pr, detail }: { pr: PRItem; detail: PRDetail | null }) {
  const info = detail?.pr;
  return (
    <div>
      <div class="flex items-start gap-2 mb-2">
        <span class={`text-lg ${stateColor(pr.state, pr.isDraft)}`}>
          {pr.isDraft ? '○' : '●'}
        </span>
        <div class="flex-1">
          <h2 class="text-base font-semibold text-gray-900">{pr.title}</h2>
          <div class="text-xs text-gray-500 mt-0.5">
            <a href={pr.url} target="_blank" rel="noopener" class="text-blue-600 hover:underline">
              {pr.repo}#{pr.number}
            </a>
            <span class="mx-1">·</span>
            <span>@{pr.author}</span>
            <span class="mx-1">·</span>
            <span>{timeAgo(pr.updatedAt)}</span>
          </div>
        </div>
      </div>

      <div class="flex flex-wrap gap-x-4 gap-y-1 text-sm mb-3">
        <div>
          <span class="text-cyan-600 font-mono text-xs">{pr.source}</span>
          <span class="text-gray-400 mx-1">→</span>
          <span class="text-cyan-600 font-mono text-xs">{pr.target}</span>
        </div>
        <div class="flex gap-2">
          <span class={stateColor(pr.state, pr.isDraft)}>
            {pr.isDraft ? 'Draft' : pr.state}
          </span>
          {(pr.reviewDecision || info?.reviewDecision) && (
            <>
              <span class="text-gray-300">|</span>
              <span class={reviewColor(pr.reviewDecision || info?.reviewDecision || '')}>
                {(pr.reviewDecision || info?.reviewDecision || '').replace(/_/g, ' ')}
              </span>
            </>
          )}
          {(pr.mergeable || info?.mergeable) && (() => {
            const m = pr.mergeable || info?.mergeable || '';
            return (
              <>
                <span class="text-gray-300">|</span>
                <span class={m === 'MERGEABLE' ? 'text-green-600' : m === 'CONFLICTING' ? 'text-red-600' : 'text-yellow-600'}>
                  {m === 'CONFLICTING' && <iconify-icon icon="codicon:git-merge" class="mr-0.5" />}
                  {m}
                </span>
              </>
            );
          })()}
        </div>
      </div>
    </div>
  );
}

function WorkflowRunView({ run }: { run: WorkflowRun }) {
  return (
    <div class="mb-3">
      <div class="flex items-center gap-1.5 text-sm font-medium">
        <span class={statusColor(run.status, run.conclusion)}>
          {statusIcon(run.status, run.conclusion)}
        </span>
        <span>{run.name}</span>
        {run.url && (
          <a href={run.url} target="_blank" rel="noopener" class="text-gray-400 hover:text-blue-500">
            <iconify-icon icon="codicon:link-external" class="text-xs" />
          </a>
        )}
      </div>
      {run.jobs && run.jobs.map(job => (
        <JobView key={job.databaseId} job={job} />
      ))}
    </div>
  );
}

function JobView({ job }: { job: Job }) {
  const failed = job.conclusion?.toLowerCase() === 'failure';
  const duration = formatDuration(job.startedAt, job.completedAt);

  return (
    <div class="ml-4 mt-1">
      <div class="flex items-center gap-1.5 text-xs">
        <span class={statusColor(job.status, job.conclusion)}>
          {statusIcon(job.status, job.conclusion)}
        </span>
        <span class={failed ? 'text-red-700 font-medium' : 'text-gray-700'}>{job.name}</span>
        {duration && <span class="text-gray-400">{duration}</span>}
        {job.url && (
          <a href={job.url} target="_blank" rel="noopener" class="text-gray-400 hover:text-blue-500">
            <iconify-icon icon="codicon:link-external" class="text-[10px]" />
          </a>
        )}
      </div>
      {failed && job.steps && job.steps.map(step => (
        <div key={step.number} class="ml-4 mt-0.5 text-xs">
          <span class={statusColor(step.status, step.conclusion)}>
            {statusIcon(step.status, step.conclusion)}
          </span>
          <span class={step.conclusion?.toLowerCase() === 'failure' ? 'text-red-600 ml-1' : 'text-gray-500 ml-1'}>
            {step.name}
          </span>
          {step.logs && (
            <pre class="mt-0.5 ml-4 text-[11px] text-gray-500 bg-gray-50 rounded p-1.5 whitespace-pre-wrap max-h-40 overflow-y-auto border border-gray-100">
              {step.logs}
            </pre>
          )}
        </div>
      ))}
      {failed && !hasStepLogs(job) && job.logs && (
        <pre class="ml-4 mt-1 text-[11px] text-gray-500 bg-red-50 rounded p-1.5 whitespace-pre-wrap max-h-40 overflow-y-auto border border-red-100">
          {job.logs}
        </pre>
      )}
    </div>
  );
}

function hasStepLogs(job: Job): boolean {
  return !!job.steps?.some(s => !!s.logs);
}

function CommentView({ comment }: { comment: PRComment }) {
  const [expanded, setExpanded] = useState(false);
  const resolved = comment.isResolved || comment.isOutdated;
  const title = extractTitle(comment.body);

  return (
    <div class={`text-xs border-b border-gray-50 ${resolved ? 'opacity-50' : ''}`}>
      <div
        class="flex items-start gap-1.5 py-1.5 cursor-pointer hover:bg-gray-50 rounded px-1 -mx-1"
        onClick={() => setExpanded(!expanded)}
      >
        <span class="shrink-0 mt-0.5">{severityIcon(comment.severity)}</span>
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-1.5">
            {comment.path && (
              <span class="text-cyan-600 font-mono truncate">{comment.path}
                {comment.line ? `:${comment.line}` : ''}
              </span>
            )}
            {resolved && <span class="text-gray-400 text-[10px]">({comment.isOutdated ? 'outdated' : 'resolved'})</span>}
          </div>
          <div class={`mt-0.5 ${resolved ? 'line-through text-gray-400' : 'text-gray-700'}`}>
            {title.length > 120 ? title.slice(0, 117) + '...' : title}
          </div>
        </div>
        <span class="text-gray-400 shrink-0">@{comment.author}</span>
        <iconify-icon
          icon={expanded ? 'codicon:chevron-up' : 'codicon:chevron-down'}
          class="text-gray-400 shrink-0 text-[10px] mt-1"
        />
      </div>
      {expanded && (
        <div class="ml-5 mb-2 mt-1">
          {comment.path && (
            <div class="text-[11px] text-cyan-700 bg-cyan-50 rounded px-2 py-1 mb-2 font-mono">
              {comment.path}{comment.line ? `:${comment.line}` : ''}
            </div>
          )}
          <Markdown text={comment.body} class="text-xs text-gray-700" />
        </div>
      )}
    </div>
  );
}

function extractTitle(body: string): string {
  const clean = body.replace(/<[^>]+>/g, '').trim();
  for (const line of clean.split('\n').slice(0, 15)) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('_') || trimmed.startsWith('> [!')) continue;
    const bold = trimmed.match(/\*\*(.+?)\*\*/);
    if (bold) return bold[1];
    if (trimmed.startsWith('# ')) return trimmed.slice(2).trim();
    return trimmed;
  }
  return clean.split('\n')[0] || '';
}

function formatDuration(start?: string, end?: string): string {
  if (!start) return '';
  const s = new Date(start);
  const e = end ? new Date(end) : new Date();
  const ms = e.getTime() - s.getTime();
  const secs = Math.floor(ms / 1000);
  if (!end) return `(running ${secs}s...)`;
  if (secs < 60) return `(${secs}s)`;
  return `(${Math.floor(secs / 60)}m ${secs % 60}s)`;
}

const SEVERITY_DEFS = [
  { key: 'critical', icon: '🔴', label: 'Critical', color: 'border-red-300 bg-red-50' },
  { key: 'major', icon: '🟠', label: 'Major', color: 'border-orange-300 bg-orange-50' },
  { key: 'minor', icon: '🟡', label: 'Minor', color: 'border-yellow-300 bg-yellow-50' },
  { key: 'nitpick', icon: '🧹', label: 'Nitpick', color: 'border-gray-300 bg-gray-50' },
];

function CommentsSection({ comments }: { comments: PRComment[] }) {
  const [showOutdated, setShowOutdated] = useState(false);
  const [severityFilter, setSeverityFilter] = useState<Set<string>>(new Set());

  const severityCounts = useMemo(() => {
    const c: Record<string, number> = {};
    let outdated = 0;
    for (const comment of comments) {
      const sev = comment.severity || '';
      c[sev] = (c[sev] || 0) + 1;
      if (comment.isResolved || comment.isOutdated) outdated++;
    }
    return { ...c, _outdated: outdated };
  }, [comments]);

  const filtered = useMemo(() => {
    return comments.filter(c => {
      if (!showOutdated && (c.isResolved || c.isOutdated)) return false;
      if (severityFilter.size > 0 && !severityFilter.has(c.severity || '')) return false;
      return true;
    });
  }, [comments, showOutdated, severityFilter]);

  function toggleSeverity(key: string) {
    setSeverityFilter(prev => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  }

  return (
    <Section title={`Comments (${filtered.length}/${comments.length})`}>
      <div class="flex items-center gap-1.5 flex-wrap mb-2">
        {SEVERITY_DEFS.map(sf => {
          const count = severityCounts[sf.key] || 0;
          if (count === 0) return null;
          const active = severityFilter.has(sf.key);
          return (
            <button key={sf.key}
              class={`inline-flex items-center gap-1 text-[11px] px-1.5 py-0.5 rounded-full border transition-colors ${
                active ? `${sf.color} font-medium` : 'border-gray-200 text-gray-500 hover:bg-gray-50'
              }`}
              onClick={() => toggleSeverity(sf.key)}
            >
              <span>{sf.icon}</span>
              <span>{count}</span>
            </button>
          );
        })}
        {severityCounts[''] > 0 && (
          <button
            class={`inline-flex items-center gap-1 text-[11px] px-1.5 py-0.5 rounded-full border transition-colors ${
              severityFilter.has('') ? 'border-gray-300 bg-gray-50 font-medium' : 'border-gray-200 text-gray-500 hover:bg-gray-50'
            }`}
            onClick={() => toggleSeverity('')}
          >
            <span>💬</span>
            <span>{severityCounts['']}</span>
          </button>
        )}
        {severityCounts._outdated > 0 && (
          <>
            <span class="text-gray-300 mx-0.5">|</span>
            <button
              class={`inline-flex items-center gap-1 text-[11px] px-1.5 py-0.5 rounded-full border transition-colors ${
                showOutdated ? 'border-gray-300 bg-gray-50 font-medium' : 'border-gray-200 text-gray-400 hover:bg-gray-50'
              }`}
              onClick={() => setShowOutdated(!showOutdated)}
            >
              <iconify-icon icon="codicon:eye" class="text-[10px]" />
              {severityCounts._outdated} resolved
            </button>
          </>
        )}
        {severityFilter.size > 0 && (
          <button class="text-[11px] text-gray-400 hover:text-gray-600 ml-0.5"
            onClick={() => setSeverityFilter(new Set())}>
            Clear
          </button>
        )}
      </div>
      {filtered.map(c => (
        <CommentView key={c.id} comment={c} />
      ))}
      {filtered.length === 0 && (
        <div class="text-xs text-gray-400 py-2">No comments match filters</div>
      )}
    </Section>
  );
}

function Section({ title, children }: { title: string; children: any }) {
  return (
    <div class="mt-4">
      <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-2 border-b border-gray-100 pb-1">
        {title}
      </h3>
      {children}
    </div>
  );
}
