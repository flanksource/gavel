import { useState, useMemo, useRef } from 'preact/hooks';
import type { PRItem, PRDetail, PRComment, GavelResultsSummary } from '../types';
import { stateColor, reviewColor, timeAgo, severityIcon } from '../utils';
import { Markdown } from './Markdown';
import { Avatar } from './Avatar';
import { WorkflowRunView } from './WorkflowView';
import { BotCommentBody, BotBadge } from './BotComment';

interface Props {
  pr: PRItem;
  detail: PRDetail | null;
  loading: boolean;
}

export function PRDetailPanel({ pr, detail, loading }: Props) {
  return (
    <div class="p-4 bg-white h-full overflow-y-auto">
      <PRHeader pr={pr} detail={detail} />

      {loading && !detail && (
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
            <WorkflowRunView key={run.databaseId} run={run} repo={pr.repo} />
          ))}
        </Section>
      )}

      {detail?.gavelResults && (
        <GavelResultsSection results={detail.gavelResults} pr={pr} />
      )}

      {detail?.comments && <DeploymentsSection comments={detail.comments} />}

      {detail?.comments && detail.comments.length > 0 && (
        <CommentsSection comments={detail.comments.filter(c => !isDeploymentComment(c))} />
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
  const authorAvatarUrl = pr.authorAvatarUrl || info?.author?.avatarUrl;
  return (
    <div>
      <div class="flex items-start gap-3 mb-2">
        <Avatar
          src={pr.repoAvatarUrl}
          alt={pr.repo}
          size={36}
          rounded="md"
          href={`https://github.com/${pr.repo}`}
          title={pr.repo}
        />
        <div class="flex-1 min-w-0">
          <h2 class="text-base font-semibold text-gray-900">{pr.title}</h2>
          <div class="flex items-center gap-1 text-xs text-gray-500 mt-0.5">
            <a href={pr.url} target="_blank" rel="noopener" class="text-blue-600 hover:underline">
              {pr.repo}#{pr.number}
            </a>
            <span class="mx-1">·</span>
            <Avatar
              src={authorAvatarUrl}
              alt={pr.author}
              size={16}
              href={`https://github.com/${pr.author}`}
              title={pr.author}
            />
            <a
              href={`https://github.com/${pr.author}`}
              target="_blank"
              rel="noopener"
              class="hover:text-blue-600 hover:underline"
            >
              @{pr.author}
            </a>
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
        <span class="inline-flex items-center gap-1 text-gray-400 shrink-0">
          <Avatar
            src={comment.avatarUrl}
            alt={comment.author}
            size={14}
            href={`https://github.com/${comment.author}`}
            title={comment.author}
          />
          @{comment.author}
          {comment.botType && <BotBadge botType={comment.botType} />}
        </span>
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
          {comment.botType
            ? <BotCommentBody comment={comment} />
            : <Markdown text={comment.body} class="text-xs text-gray-700" />}
        </div>
      )}
    </div>
  );
}

function extractTitle(body: string): string {
  const clean = body.replace(/[<>]/g, '').trim();
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

function isDeploymentComment(c: PRComment): boolean {
  return c.botType === 'vercel' && c.body.startsWith('[vc]:');
}

interface VercelProject {
  name: string;
  previewUrl: string;
  inspectorUrl: string;
  status: string; // DEPLOYED, BUILDING, ERROR, CANCELED, QUEUED
}

// Vercel comments embed a base64 JSON blob: [vc]: #<hash>:<base64json>\n...
function parseVercelProjects(body: string): VercelProject[] {
  const m = body.match(/^\[vc\]:\s*#[^:]+:(\S+)/);
  if (!m) return [];
  try {
    const data = JSON.parse(atob(m[1]));
    return (data.projects || []).map((p: any) => ({
      name: p.name || '',
      previewUrl: p.previewUrl?.startsWith('http') ? p.previewUrl : `https://${p.previewUrl}`,
      inspectorUrl: p.inspectorUrl || '',
      status: (p.nextCommitStatus || '').toUpperCase(),
    }));
  } catch { return []; }
}

const deployStatusConfig: Record<string, { icon: string; color: string; label: string }> = {
  DEPLOYED:  { icon: 'codicon:pass',           color: 'text-green-600', label: 'Deployed' },
  BUILDING:  { icon: 'svg-spinners:ring-resize', color: 'text-yellow-600', label: 'Building' },
  QUEUED:    { icon: 'codicon:clock',           color: 'text-gray-500',  label: 'Queued' },
  ERROR:     { icon: 'codicon:error',           color: 'text-red-600',   label: 'Error' },
  CANCELED:  { icon: 'codicon:circle-slash',    color: 'text-gray-500',  label: 'Canceled' },
};

function DeploymentRow({ project }: { project: VercelProject }) {
  const [hover, setHover] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const st = deployStatusConfig[project.status] || deployStatusConfig.QUEUED;

  return (
    <div class="relative" ref={ref}
      onMouseEnter={() => setHover(true)} onMouseLeave={() => setHover(false)}
    >
      <div class="flex items-center gap-2 py-1.5 px-1 -mx-1 rounded hover:bg-gray-50 text-sm transition-colors">
        <iconify-icon icon={st.icon} class={`${st.color} text-xs`} />
        <a href={project.previewUrl} target="_blank" rel="noopener"
          class="text-blue-600 hover:underline font-medium flex-1 truncate"
        >
          {project.name}
        </a>
        <a href={project.inspectorUrl} target="_blank" rel="noopener"
          class="text-gray-400 hover:text-gray-600 p-0.5 rounded hover:bg-gray-100 transition-colors"
          title="Build output"
          onClick={(e: Event) => e.stopPropagation()}
        >
          <iconify-icon icon="codicon:server-process" class="text-xs" />
        </a>
      </div>
      {hover && (
        <div class="absolute left-0 top-full z-50 mt-0.5 w-72 bg-white border border-gray-200 rounded-lg shadow-lg p-3 text-xs">
          <div class="flex items-center gap-1.5 mb-2">
            <iconify-icon icon="simple-icons:vercel" class="text-sm" />
            <span class="font-semibold text-gray-900">{project.name}</span>
            <span class={`ml-auto inline-flex items-center gap-1 ${st.color}`}>
              <iconify-icon icon={st.icon} class="text-[10px]" />
              {st.label}
            </span>
          </div>
          <div class="space-y-1.5 text-gray-600">
            <div class="flex items-center gap-1.5">
              <iconify-icon icon="codicon:link-external" class="text-gray-400 text-[10px] shrink-0" />
              <a href={project.previewUrl} target="_blank" rel="noopener"
                class="text-blue-600 hover:underline truncate">
                {project.previewUrl.replace(/^https?:\/\//, '')}
              </a>
            </div>
            <div class="flex items-center gap-1.5">
              <iconify-icon icon="codicon:server-process" class="text-gray-400 text-[10px] shrink-0" />
              <a href={project.inspectorUrl} target="_blank" rel="noopener"
                class="text-blue-600 hover:underline truncate">
                Build output
              </a>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function DeploymentsSection({ comments }: { comments: PRComment[] }) {
  const projects = useMemo(() => {
    for (const c of comments) {
      if (!isDeploymentComment(c)) continue;
      const p = parseVercelProjects(c.body);
      if (p.length > 0) return p;
    }
    return [];
  }, [comments]);
  if (projects.length === 0) return null;

  return (
    <Section title="Deployments">
      {projects.map(p => <DeploymentRow key={p.name} project={p} />)}
    </Section>
  );
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

  const severityCounts = useMemo<Record<string, number>>(() => {
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

function GavelResultsSection({ results, pr }: { results: GavelResultsSummary; pr: PRItem }) {
  if (results.error) {
    return (
      <Section title="Gavel Results">
        <div class="text-xs text-gray-400 py-1">
          <iconify-icon icon="codicon:warning" class="text-yellow-500 mr-1" />
          {results.error}
        </div>
      </Section>
    );
  }

  const backTo = `${window.location.pathname}${window.location.search}`;
  const basePath = `/results/${pr.repo}/${results.artifactId}`;

  const rows: { tab: string; icon: string; color: string; label: string }[] = [];

  if (results.testsTotal > 0) {
    const parts: string[] = [];
    if (results.testsPassed > 0) parts.push(`${results.testsPassed} passed`);
    if (results.testsFailed > 0) parts.push(`${results.testsFailed} failed`);
    if (results.testsSkipped > 0) parts.push(`${results.testsSkipped} skipped`);
    rows.push({
      tab: 'tests',
      icon: results.testsFailed > 0 ? 'codicon:error' : 'codicon:pass',
      color: results.testsFailed > 0 ? 'text-red-600' : 'text-green-600',
      label: `Tests: ${parts.join(', ')}`,
    });
  }

  if (results.lintLinters > 0) {
    rows.push({
      tab: 'lint',
      icon: results.lintViolations > 0 ? 'codicon:warning' : 'codicon:pass',
      color: results.lintViolations > 0 ? 'text-yellow-600' : 'text-green-600',
      label: results.lintViolations > 0
        ? `Lint: ${results.lintViolations} violation${results.lintViolations !== 1 ? 's' : ''} from ${results.lintLinters} linter${results.lintLinters !== 1 ? 's' : ''}`
        : `Lint: ${results.lintLinters} linter${results.lintLinters !== 1 ? 's' : ''} clean`,
    });
  }

  if (results.hasBench) {
    rows.push({
      tab: 'bench',
      icon: (results.benchRegressions ?? 0) > 0 ? 'codicon:warning' : 'codicon:graph',
      color: (results.benchRegressions ?? 0) > 0 ? 'text-red-600' : 'text-blue-600',
      label: (results.benchRegressions ?? 0) > 0
        ? `Bench: ${results.benchRegressions} regression${results.benchRegressions !== 1 ? 's' : ''}`
        : 'Bench: no regressions',
    });
  }

  if (rows.length === 0) return null;

  return (
    <Section title="Gavel Results">
      {rows.map(row => (
        <a
          key={row.tab}
          href={`${basePath}/${row.tab}?backTo=${encodeURIComponent(backTo)}`}
          class="flex items-center gap-2 py-1.5 px-1 -mx-1 rounded hover:bg-gray-50 text-sm group transition-colors"
        >
          <iconify-icon icon={row.icon} class={row.color} />
          <span class="text-gray-700 flex-1">{row.label}</span>
          <iconify-icon icon="codicon:chevron-right" class="text-gray-300 group-hover:text-gray-500 text-xs" />
        </a>
      ))}
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
