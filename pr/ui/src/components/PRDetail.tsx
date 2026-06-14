import { useState, useMemo, useRef } from 'react';
import type { PRItem, PRDetail, PRComment, GavelResultsSummary, TestFailure, LintViolation } from '../types';
import { stateColor, reviewColor, timeAgo, severityIcon } from '../utils';
import { ansiToHtml, stripAnsi } from '../ansi';
import { Markdown } from './Markdown';
import { Avatar } from './Avatar';
import { WorkflowRunView } from './WorkflowView';
import { BotCommentBody, BotBadge } from './BotComment';
import type { WorkflowRun } from '../types';

function formatWorkflowsText(runs: WorkflowRun[]): string {
  return runs.map(r => {
    const outcome = r.conclusion || r.status;
    const jobs = (r.jobs || []).map(j => `  - ${j.conclusion || j.status}: ${j.name}`).join('\n');
    return `${r.name} [${outcome}]${jobs ? '\n' + jobs : ''}`;
  }).join('\n\n');
}

function formatWorkflowsMarkdown(runs: WorkflowRun[]): string {
  return runs.map(r => {
    const outcome = r.conclusion || r.status;
    const jobs = (r.jobs || []).map(j => `  - \`${j.conclusion || j.status}\` ${j.name}`).join('\n');
    return `### ${r.name} _(${outcome})_${r.url ? ` · [view](${r.url})` : ''}${jobs ? '\n' + jobs : ''}`;
  }).join('\n\n');
}

function formatGavelText(g: GavelResultsSummary): string {
  const lines: string[] = [];
  if (g.testsTotal > 0) {
    lines.push(`Tests: ${g.testsPassed} passed / ${g.testsFailed} failed / ${g.testsSkipped} skipped (total ${g.testsTotal})`);
  }
  if (g.lintLinters > 0) {
    lines.push(`Lint: ${g.lintViolations} violations across ${g.lintLinters} linters`);
  }
  if (g.hasBench) {
    lines.push(`Bench: ${g.benchRegressions ?? 0} regressions`);
  }
  if (g.topFailures && g.topFailures.length) {
    lines.push('', 'Top failures:');
    for (const f of g.topFailures) {
      lines.push(`- ${f.name}${f.file ? ` (${f.file}${f.line ? ':' + f.line : ''})` : ''}${f.message ? ` — ${stripAnsi(f.message)}` : ''}`);
    }
  }
  if (g.topLintViolations && g.topLintViolations.length) {
    lines.push('', 'Top lint violations:');
    for (const v of g.topLintViolations) {
      lines.push(`- [${v.linter}] ${v.file ?? ''}${v.line ? ':' + v.line : ''}${v.message ? ` — ${stripAnsi(v.message)}` : ''}`);
    }
  }
  return lines.join('\n');
}

function formatGavelMarkdown(g: GavelResultsSummary): string {
  const lines: string[] = [];
  if (g.testsTotal > 0) {
    lines.push(`- **Tests**: ✅ ${g.testsPassed} · ❌ ${g.testsFailed} · ⏭ ${g.testsSkipped} (total ${g.testsTotal})`);
  }
  if (g.lintLinters > 0) {
    lines.push(`- **Lint**: ${g.lintViolations} violations across ${g.lintLinters} linters`);
  }
  if (g.hasBench) {
    lines.push(`- **Bench**: ${g.benchRegressions ?? 0} regressions`);
  }
  if (g.topFailures && g.topFailures.length) {
    lines.push('', '**Top failures**');
    for (const f of g.topFailures) {
      lines.push(`- \`${f.name}\`${f.file ? ` _${f.file}${f.line ? ':' + f.line : ''}_` : ''}${f.message ? ` — ${stripAnsi(f.message)}` : ''}`);
    }
  }
  if (g.topLintViolations && g.topLintViolations.length) {
    lines.push('', '**Top lint violations**');
    for (const v of g.topLintViolations) {
      lines.push(`- \`${v.linter}\` ${v.file ?? ''}${v.line ? ':' + v.line : ''}${v.message ? ` — ${stripAnsi(v.message)}` : ''}`);
    }
  }
  return lines.join('\n');
}

interface Props {
  pr: PRItem;
  detail: PRDetail | null;
  loading: boolean;
}

export function PRDetailPanel({ pr, detail, loading }: Props) {
  return (
    <div className="p-4 bg-white h-full overflow-y-auto">
      <PRHeader pr={pr} detail={detail} />

      {loading && !detail && (
        <div className="flex items-center gap-2 text-sm text-gray-400 mt-4">
          <iconify-icon icon="svg-spinners:ring-resize" className="text-blue-500" />
          Loading details...
        </div>
      )}

      {detail?.error && (
        <div className="mt-3 p-2 bg-red-50 border border-red-100 rounded text-xs text-red-700">
          <iconify-icon icon="codicon:error" className="mr-1" />
          {detail.error}
        </div>
      )}

      {detail?.runs && Object.keys(detail.runs).length > 0 && (
        <Section
          title="Workflows"
          actions={{
            json: () => detail.runs,
            text: () => formatWorkflowsText(Object.values(detail.runs!)),
            markdown: () => formatWorkflowsMarkdown(Object.values(detail.runs!)),
          }}
        >
          {Object.values(detail.runs).map(run => (
            <WorkflowRunView key={run.databaseId} run={run} repo={pr.repo} />
          ))}
        </Section>
      )}

      {detail?.gavelResults && detail.gavelResults.length > 0 && (
        <GavelResultsSection shards={detail.gavelResults} pr={pr} />
      )}

      {detail?.comments && <DeploymentsSection comments={detail.comments} />}

      {detail?.comments && detail.comments.length > 0 && (
        <CommentsSection comments={detail.comments.filter(c => !isDeploymentComment(c))} />
      )}

      <div className="pt-3 mt-3 border-t border-gray-100">
        <a href={pr.url} target="_blank" rel="noopener"
          className="inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800 hover:underline">
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
      <div className="flex items-start gap-3 mb-2">
        <Avatar
          src={pr.repoAvatarUrl}
          alt={pr.repo}
          size={36}
          rounded="md"
          href={`https://github.com/${pr.repo}`}
          title={pr.repo}
        />
        <div className="flex-1 min-w-0">
          <h2 className="text-base font-semibold text-gray-900">{pr.title}</h2>
          <div className="flex items-center gap-1 text-xs text-gray-500 mt-0.5">
            <a href={pr.url} target="_blank" rel="noopener" className="text-blue-600 hover:underline">
              {pr.repo}#{pr.number}
            </a>
            <span className="mx-1">·</span>
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
              className="hover:text-blue-600 hover:underline"
            >
              @{pr.author}
            </a>
            <span className="mx-1">·</span>
            <span>{timeAgo(pr.updatedAt)}</span>
          </div>
        </div>
      </div>

      <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm mb-3">
        <div>
          <span className="text-cyan-600 font-mono text-xs">{pr.source}</span>
          <span className="text-gray-400 mx-1">→</span>
          <span className="text-cyan-600 font-mono text-xs">{pr.target}</span>
        </div>
        <div className="flex gap-2">
          <span className={stateColor(pr.state, pr.isDraft)}>
            {pr.isDraft ? 'Draft' : pr.state}
          </span>
          {(pr.reviewDecision || info?.reviewDecision) && (
            <>
              <span className="text-gray-300">|</span>
              <span className={reviewColor(pr.reviewDecision || info?.reviewDecision || '')}>
                {(pr.reviewDecision || info?.reviewDecision || '').replace(/_/g, ' ')}
              </span>
            </>
          )}
          {(pr.mergeable || info?.mergeable) && (() => {
            const m = pr.mergeable || info?.mergeable || '';
            return (
              <>
                <span className="text-gray-300">|</span>
                <span className={m === 'MERGEABLE' ? 'text-green-600' : m === 'CONFLICTING' ? 'text-red-600' : 'text-yellow-600'}>
                  {m === 'CONFLICTING' && <iconify-icon icon="codicon:git-merge" className="mr-0.5" />}
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
    <div className={`text-xs border-b border-gray-50 ${resolved ? 'opacity-50' : ''}`}>
      <div
        className="flex items-start gap-1.5 py-1.5 cursor-pointer hover:bg-gray-50 rounded px-1 -mx-1"
        onClick={() => setExpanded(!expanded)}
      >
        <span className="shrink-0 mt-0.5">{severityIcon(comment.severity)}</span>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1.5">
            {comment.path && (
              <span className="text-cyan-600 font-mono truncate">{comment.path}
                {comment.line ? `:${comment.line}` : ''}
              </span>
            )}
            {resolved && <span className="text-gray-400 text-[10px]">({comment.isOutdated ? 'outdated' : 'resolved'})</span>}
          </div>
          <div className={`mt-0.5 ${resolved ? 'line-through text-gray-400' : 'text-gray-700'}`}>
            {title.length > 120 ? title.slice(0, 117) + '...' : title}
          </div>
        </div>
        <span className="inline-flex items-center gap-1 text-gray-400 shrink-0">
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
          className="text-gray-400 shrink-0 text-[10px] mt-1"
        />
      </div>
      {expanded && (
        <div className="ml-5 mb-2 mt-1">
          {comment.path && (
            <div className="text-[11px] text-cyan-700 bg-cyan-50 rounded px-2 py-1 mb-2 font-mono">
              {comment.path}{comment.line ? `:${comment.line}` : ''}
            </div>
          )}
          {comment.botType
            ? <BotCommentBody comment={comment} />
            : <Markdown text={comment.body} className="text-xs text-gray-700" />}
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
    <div className="relative" ref={ref}
      onMouseEnter={() => setHover(true)} onMouseLeave={() => setHover(false)}
    >
      <div className="flex items-center gap-2 py-1.5 px-1 -mx-1 rounded hover:bg-gray-50 text-sm transition-colors">
        <iconify-icon icon={st.icon} className={`${st.color} text-xs`} />
        <a href={project.previewUrl} target="_blank" rel="noopener"
          className="text-blue-600 hover:underline font-medium flex-1 truncate"
        >
          {project.name}
        </a>
        <a href={project.inspectorUrl} target="_blank" rel="noopener"
          className="text-gray-400 hover:text-gray-600 p-0.5 rounded hover:bg-gray-100 transition-colors"
          title="Build output"
          onClick={(e) => e.stopPropagation()}
        >
          <iconify-icon icon="codicon:server-process" className="text-xs" />
        </a>
      </div>
      {hover && (
        <div className="absolute left-0 top-full z-50 mt-0.5 w-72 bg-white border border-gray-200 rounded-lg shadow-lg p-3 text-xs">
          <div className="flex items-center gap-1.5 mb-2">
            <iconify-icon icon="simple-icons:vercel" className="text-sm" />
            <span className="font-semibold text-gray-900">{project.name}</span>
            <span className={`ml-auto inline-flex items-center gap-1 ${st.color}`}>
              <iconify-icon icon={st.icon} className="text-[10px]" />
              {st.label}
            </span>
          </div>
          <div className="space-y-1.5 text-gray-600">
            <div className="flex items-center gap-1.5">
              <iconify-icon icon="codicon:link-external" className="text-gray-400 text-[10px] shrink-0" />
              <a href={project.previewUrl} target="_blank" rel="noopener"
                className="text-blue-600 hover:underline truncate">
                {project.previewUrl.replace(/^https?:\/\//, '')}
              </a>
            </div>
            <div className="flex items-center gap-1.5">
              <iconify-icon icon="codicon:server-process" className="text-gray-400 text-[10px] shrink-0" />
              <a href={project.inspectorUrl} target="_blank" rel="noopener"
                className="text-blue-600 hover:underline truncate">
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
    <Section
      title="Deployments"
      actions={{
        json: () => projects,
        text: () => projects.map(p => `${p.name}: ${p.status}${p.previewUrl ? ' ' + p.previewUrl : ''}`).join('\n'),
        markdown: () => projects.map(p => `- **${p.name}** — ${p.status}${p.previewUrl ? ` ([preview](${p.previewUrl}))` : ''}`).join('\n'),
      }}
    >
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
    <Section
      title={`Comments (${filtered.length}/${comments.length})`}
      actions={{
        json: () => filtered,
        text: () => filtered.map(c => `@${c.author}${c.path ? ` [${c.path}${c.line ? ':' + c.line : ''}]` : ''}\n${c.body}`).join('\n\n---\n\n'),
        markdown: () => filtered.map(c => `**@${c.author}**${c.path ? ` _(${c.path}${c.line ? ':' + c.line : ''})_` : ''}\n\n${c.body}`).join('\n\n---\n\n'),
      }}
    >
      <div className="flex items-center gap-1.5 flex-wrap mb-2">
        {SEVERITY_DEFS.map(sf => {
          const count = severityCounts[sf.key] || 0;
          if (count === 0) return null;
          const active = severityFilter.has(sf.key);
          return (
            <button key={sf.key}
              className={`inline-flex items-center gap-1 text-[11px] px-1.5 py-0.5 rounded-full border transition-colors ${
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
            className={`inline-flex items-center gap-1 text-[11px] px-1.5 py-0.5 rounded-full border transition-colors ${
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
            <span className="text-gray-300 mx-0.5">|</span>
            <button
              className={`inline-flex items-center gap-1 text-[11px] px-1.5 py-0.5 rounded-full border transition-colors ${
                showOutdated ? 'border-gray-300 bg-gray-50 font-medium' : 'border-gray-200 text-gray-400 hover:bg-gray-50'
              }`}
              onClick={() => setShowOutdated(!showOutdated)}
            >
              <iconify-icon icon="codicon:eye" className="text-[10px]" />
              {severityCounts._outdated} resolved
            </button>
          </>
        )}
        {severityFilter.size > 0 && (
          <button className="text-[11px] text-gray-400 hover:text-gray-600 ml-0.5"
            onClick={() => setSeverityFilter(new Set())}>
            Clear
          </button>
        )}
      </div>
      {filtered.map(c => (
        <CommentView key={c.id} comment={c} />
      ))}
      {filtered.length === 0 && (
        <div className="text-xs text-gray-400 py-2">No comments match filters</div>
      )}
    </Section>
  );
}

interface MetricCardProps {
  href?: string;
  icon: string;
  label: string;
  value: string | number;
  sub?: string;
  tone: 'pass' | 'fail' | 'warn' | 'info' | 'neutral';
}

function MetricCard({ href, icon, label, value, sub, tone }: MetricCardProps) {
  const toneClass = {
    pass: href ? 'bg-green-50 border-green-200 text-green-700 hover:bg-green-100' : 'bg-green-50 border-green-200 text-green-700',
    fail: href ? 'bg-red-50 border-red-200 text-red-700 hover:bg-red-100' : 'bg-red-50 border-red-200 text-red-700',
    warn: href ? 'bg-yellow-50 border-yellow-200 text-yellow-700 hover:bg-yellow-100' : 'bg-yellow-50 border-yellow-200 text-yellow-700',
    info: href ? 'bg-blue-50 border-blue-200 text-blue-700 hover:bg-blue-100' : 'bg-blue-50 border-blue-200 text-blue-700',
    neutral: href ? 'bg-gray-50 border-gray-200 text-gray-600 hover:bg-gray-100' : 'bg-gray-50 border-gray-200 text-gray-600',
  }[tone];
  const body = (
    <>
      <div className="flex items-center justify-between">
        <iconify-icon icon={icon} className="text-lg" />
        {href && <iconify-icon icon="codicon:chevron-right" className="text-xs opacity-30 group-hover:opacity-70" />}
      </div>
      <div className="text-2xl font-semibold tabular-nums leading-tight mt-1">{value}</div>
      <div className="text-[11px] font-medium uppercase tracking-wide opacity-80">{label}</div>
      {sub && <div className="text-[11px] mt-0.5 opacity-70 truncate">{sub}</div>}
    </>
  );
  const cls = `group block rounded-lg border px-3 py-2 transition-colors ${toneClass}`;
  return href ? <a href={href} className={cls}>{body}</a> : <div className={cls}>{body}</div>;
}

// aggregateShardsForUI rolls up per-shard summaries for the detail card's
// header. The artifactId/Url fields are intentionally zeroed because the
// aggregate has no single artifact to deep-link to — drill-downs should
// happen via the per-shard rows below it.
function aggregateShardsForUI(shards: GavelResultsSummary[]): GavelResultsSummary {
  if (shards.length === 1) return shards[0];
  const agg: GavelResultsSummary = {
    artifactId: 0,
    artifactUrl: '',
    testsPassed: 0,
    testsFailed: 0,
    testsSkipped: 0,
    testsTotal: 0,
    lintViolations: 0,
    lintLinters: 0,
    hasBench: false,
    benchRegressions: 0,
    topFailures: [],
    topLintViolations: [],
  };
  for (const s of shards) {
    agg.testsPassed += s.testsPassed;
    agg.testsFailed += s.testsFailed;
    agg.testsSkipped += s.testsSkipped;
    agg.testsTotal += s.testsTotal;
    agg.lintViolations += s.lintViolations;
    agg.lintLinters += s.lintLinters;
    agg.benchRegressions = (agg.benchRegressions ?? 0) + (s.benchRegressions ?? 0);
    if (s.hasBench) agg.hasBench = true;
    for (const f of s.topFailures ?? []) {
      if ((agg.topFailures!.length) >= 5) break;
      agg.topFailures!.push(f);
    }
    for (const v of s.topLintViolations ?? []) {
      if ((agg.topLintViolations!.length) >= 5) break;
      agg.topLintViolations!.push(v);
    }
  }
  return agg;
}

function GavelResultsSection({ shards, pr }: { shards: GavelResultsSummary[]; pr: PRItem }) {
  if (shards.length === 0) return null;

  const agg = useMemo(() => aggregateShardsForUI(shards), [shards]);
  const multi = shards.length > 1;
  const [breakdownOpen, setBreakdownOpen] = useState(false);

  // The aggregate has no single artifact, so the metric cards in the
  // header don't deep-link when there are multiple shards — clicks on
  // specific results happen via the per-shard rows.
  const headerCards = buildMetricCards(agg, multi ? null : (tab: string) => shardLink(pr, agg, tab));

  return (
    <Section
      title="Gavel Results"
      actions={{
        json: () => (multi ? { aggregate: agg, shards } : agg),
        text: () => formatGavelText(agg),
        markdown: () => formatGavelMarkdown(agg),
      }}
    >
      {headerCards.length > 0 ? (
        <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
          {headerCards.map((c, i) => (
            <MetricCard key={i} {...c} />
          ))}
        </div>
      ) : (
        <div className="text-xs text-gray-400 py-1">
          {multi
            ? `${shards.length} shard${shards.length !== 1 ? 's' : ''} reported but produced no test, lint, or bench data.`
            : 'No test, lint, or bench data in this artifact.'}
        </div>
      )}

      {!multi && (
        <ShardExtras results={shards[0]} />
      )}

      {multi && (
        <div className="mt-3">
          <button
            type="button"
            className="flex items-center gap-1 text-[11px] uppercase tracking-wide text-gray-500 hover:text-gray-700"
            onClick={() => setBreakdownOpen(o => !o)}
            aria-expanded={breakdownOpen}
          >
            <iconify-icon
              icon={breakdownOpen ? 'codicon:chevron-down' : 'codicon:chevron-right'}
              className="text-gray-400"
            />
            <span className="font-semibold">Per-shard breakdown</span>
            <span className="text-gray-400 normal-case tracking-normal">
              ({shards.length} shard{shards.length !== 1 ? 's' : ''})
            </span>
          </button>
          {breakdownOpen && (
            <div className="mt-2 divide-y divide-gray-100 border border-gray-100 rounded">
              {shards.map(s => (
                <GavelShardRow key={s.stickyId || s.artifactId} results={s} pr={pr} />
              ))}
            </div>
          )}
        </div>
      )}
    </Section>
  );
}

function shardLink(pr: PRItem, results: GavelResultsSummary, tab: string): string {
  const backTo = `${window.location.pathname}${window.location.search}`;
  const basePath = `/results/${pr.repo}/${results.artifactId}`;
  return `${basePath}/${tab}?backTo=${encodeURIComponent(backTo)}`;
}

function buildMetricCards(
  results: GavelResultsSummary,
  link: ((tab: string) => string) | null,
): MetricCardProps[] {
  const cards: MetricCardProps[] = [];
  const href = (tab: string) => (link ? link(tab) : undefined);

  if (results.testsTotal > 0) {
    cards.push({
      href: href('tests?filter=passed'),
      icon: 'codicon:pass',
      label: 'Passed',
      value: results.testsPassed,
      sub: `of ${results.testsTotal} test${results.testsTotal !== 1 ? 's' : ''}`,
      tone: 'pass',
    });
    cards.push({
      href: href('tests?filter=failed'),
      icon: 'codicon:error',
      label: 'Failed',
      value: results.testsFailed,
      sub: results.testsFailed > 0 ? 'need triage' : 'none',
      tone: results.testsFailed > 0 ? 'fail' : 'neutral',
    });
    if (results.testsSkipped > 0) {
      cards.push({
        href: href('tests?filter=skipped'),
        icon: 'codicon:debug-step-over',
        label: 'Skipped',
        value: results.testsSkipped,
        sub: 'not run',
        tone: 'warn',
      });
    }
  }

  if (results.lintLinters > 0) {
    cards.push({
      href: href('lint'),
      icon: results.lintViolations > 0 ? 'codicon:warning' : 'codicon:pass',
      label: 'Lint',
      value: results.lintViolations,
      sub: results.lintViolations > 0
        ? `from ${results.lintLinters} linter${results.lintLinters !== 1 ? 's' : ''}`
        : `${results.lintLinters} linter${results.lintLinters !== 1 ? 's' : ''} clean`,
      tone: results.lintViolations > 0 ? 'warn' : 'pass',
    });
  }

  if (results.hasBench) {
    const regs = results.benchRegressions ?? 0;
    cards.push({
      href: href('bench'),
      icon: regs > 0 ? 'codicon:arrow-down' : 'codicon:graph',
      label: 'Bench',
      value: regs,
      sub: regs > 0
        ? `regression${regs !== 1 ? 's' : ''}`
        : 'no regressions',
      tone: regs > 0 ? 'fail' : 'info',
    });
  }

  return cards;
}

function ShardExtras({ results }: { results: GavelResultsSummary }) {
  const failures = results.topFailures ?? [];
  const lintHits = results.topLintViolations ?? [];
  return (
    <>
      {failures.length > 0 && (
        <FailureList
          title="Test failures"
          icon="codicon:beaker-stop"
          iconColor="text-red-600"
          total={results.testsFailed}
        >
          {failures.map((f, i) => <TestFailureRow key={i} f={f} />)}
        </FailureList>
      )}
      {lintHits.length > 0 && (
        <FailureList
          title="Lint violations"
          icon="codicon:warning"
          iconColor="text-yellow-600"
          total={results.lintViolations}
        >
          {lintHits.map((v, i) => <LintViolationRow key={i} v={v} />)}
        </FailureList>
      )}
    </>
  );
}

function ShardSummaryBadges({ g }: { g: GavelResultsSummary }) {
  if (g.error) {
    return (
      <span className="inline-flex items-center text-yellow-600" title={g.error}>
        <iconify-icon icon="codicon:warning" />
      </span>
    );
  }
  const items: { icon: string; color: string; count: number; title: string }[] = [];
  if (g.testsPassed > 0) {
    items.push({ icon: 'codicon:pass', color: 'text-green-600', count: g.testsPassed, title: `${g.testsPassed} passed` });
  }
  if (g.testsFailed > 0) {
    items.push({ icon: 'codicon:error', color: 'text-red-600', count: g.testsFailed, title: `${g.testsFailed} failed` });
  }
  if (g.testsSkipped > 0) {
    items.push({ icon: 'codicon:debug-step-over', color: 'text-gray-500', count: g.testsSkipped, title: `${g.testsSkipped} skipped` });
  }
  if (g.lintViolations > 0) {
    items.push({ icon: 'codicon:warning', color: 'text-yellow-600', count: g.lintViolations, title: `${g.lintViolations} lint` });
  }
  if ((g.benchRegressions ?? 0) > 0) {
    items.push({ icon: 'codicon:arrow-down', color: 'text-red-600', count: g.benchRegressions ?? 0, title: `${g.benchRegressions} bench regression` });
  }
  if (items.length === 0) return null;
  return (
    <span className="inline-flex items-center gap-1 tabular-nums">
      {items.map((it, i) => (
        <span key={i} className={`inline-flex items-center ${it.color} leading-none`} title={it.title}>
          <iconify-icon icon={it.icon} className="text-[12px]" />
          <span className="text-[11px] font-medium">{it.count}</span>
        </span>
      ))}
    </span>
  );
}

function GavelShardRow({ results, pr }: { results: GavelResultsSummary; pr: PRItem }) {
  const [open, setOpen] = useState(false);
  const label = results.stickyId || `artifact ${results.artifactId}`;
  const link = (tab: string) => shardLink(pr, results, tab);
  const cards = buildMetricCards(results, link);

  return (
    <div>
      <button
        type="button"
        className="w-full flex items-center gap-2 px-2 py-1.5 hover:bg-gray-50 text-left"
        onClick={() => setOpen(o => !o)}
        aria-expanded={open}
      >
        <iconify-icon
          icon={open ? 'codicon:chevron-down' : 'codicon:chevron-right'}
          className="text-gray-400 text-[12px] shrink-0"
        />
        <span className="text-xs font-mono text-gray-700 truncate">{label}</span>
        <ShardSummaryBadges g={results} />
      </button>
      {open && (
        <div className="px-2 pb-3 pt-1">
          {results.error ? (
            <div className="text-xs text-gray-400 py-1">
              <iconify-icon icon="codicon:warning" className="text-yellow-500 mr-1" />
              {results.error}
            </div>
          ) : cards.length === 0 ? (
            <div className="text-xs text-gray-400 py-1">
              No test, lint, or bench data in this artifact.{' '}
              <a className="text-blue-600 hover:underline" href={link('tests')}>Open results</a>
            </div>
          ) : (
            <>
              <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
                {cards.map((c, i) => (
                  <MetricCard key={i} {...c} />
                ))}
              </div>
              <ShardExtras results={results} />
            </>
          )}
        </div>
      )}
    </div>
  );
}

function FailureList({ title, icon, iconColor, total, children }: {
  title: string;
  icon: string;
  iconColor: string;
  total: number;
  children: any;
}) {
  const rows = Array.isArray(children) ? children : [children];
  const shown = rows.length;
  return (
    <div className="mt-3">
      <div className="flex items-center gap-1.5 text-[11px] uppercase tracking-wide text-gray-500 mb-1">
        <iconify-icon icon={icon} className={iconColor} />
        <span className="font-semibold">{title}</span>
        <span className="text-gray-400 normal-case tracking-normal">
          showing {shown} of {total}
        </span>
      </div>
      <div className="divide-y divide-gray-100 border border-gray-100 rounded">
        {rows}
      </div>
    </div>
  );
}

function FailureHeader({ f, withChevron }: { f: TestFailure; withChevron: boolean }) {
  const location = f.file ? (f.line ? `${f.file}:${f.line}` : f.file) : '';
  const plainMsg = f.message ?? '';
  const msgHtml = plainMsg ? ansiToHtml(plainMsg) : '';
  return (
    <div className="flex items-start gap-2 py-1.5 px-2 text-xs">
      {withChevron && (
        <iconify-icon icon="codicon:chevron-right" className="text-gray-400 mt-0.5 shrink-0 transition-transform group-open:rotate-90" />
      )}
      <iconify-icon icon="codicon:error" className="text-red-600 mt-0.5 shrink-0" />
      <div className="flex-1 min-w-0">
        <div className="font-medium text-gray-800 truncate" title={f.name}>
          {f.suite ? <span className="text-gray-400">{f.suite} › </span> : null}
          {f.name}
        </div>
        <div className="text-[11px] text-gray-500 truncate font-mono" title={`${location}${plainMsg ? ' — ' + plainMsg : ''}`}>
          {location && <span>{location}</span>}
          {location && plainMsg && <span className="mx-1">·</span>}
          {plainMsg && <span dangerouslySetInnerHTML={{ __html: msgHtml }} />}
        </div>
      </div>
    </div>
  );
}

function TestFailureRow({ f }: { f: TestFailure }) {
  const hasDetails = !!(f.details && f.details.trim().length > 0);
  if (!hasDetails) return <div><FailureHeader f={f} withChevron={false} /></div>;
  const detailsHtml = ansiToHtml(f.details!);
  return (
    <details className="group">
      <summary className="list-none cursor-pointer hover:bg-gray-50">
        <FailureHeader f={f} withChevron={true} />
      </summary>
      <pre
        className="text-[11px] font-mono text-gray-100 bg-[#1e1e1e] px-3 py-2 overflow-x-auto whitespace-pre-wrap border-t border-gray-200"
        dangerouslySetInnerHTML={{ __html: detailsHtml }}
      />
    </details>
  );
}

function LintViolationRow({ v }: { v: LintViolation }) {
  const location = v.file ? (v.line ? `${v.file}:${v.line}` : v.file) : '';
  const plainMsg = v.message ?? '';
  const msgHtml = plainMsg ? ansiToHtml(plainMsg) : '';
  return (
    <div className="flex items-start gap-2 py-1.5 px-2 text-xs">
      <iconify-icon icon="codicon:warning" className="text-yellow-600 mt-0.5 shrink-0" />
      <div className="flex-1 min-w-0">
        <div className="font-medium text-gray-800 truncate">
          <span className="text-gray-400">{v.linter}</span>
          {v.rule && <span className="ml-1 text-gray-500">({v.rule})</span>}
        </div>
        <div className="text-[11px] text-gray-500 truncate font-mono" title={`${location}${plainMsg ? ' — ' + plainMsg : ''}`}>
          {location && <span>{location}</span>}
          {location && plainMsg && <span className="mx-1">·</span>}
          {plainMsg && <span dangerouslySetInnerHTML={{ __html: msgHtml }} />}
        </div>
      </div>
    </div>
  );
}

interface SectionActions {
  text?: () => string;
  markdown?: () => string;
  json?: () => unknown;
}

function Section({ title, children, actions }: { title: string; children: any; actions?: SectionActions }) {
  return (
    <div className="mt-4">
      <div className="flex items-center justify-between mb-2 border-b border-gray-100 pb-1">
        <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wide">{title}</h3>
        {actions && <SectionActionsBar actions={actions} title={title} />}
      </div>
      {children}
    </div>
  );
}

function SectionActionsBar({ actions, title }: { actions: SectionActions; title: string }) {
  const [copied, setCopied] = useState<'text' | 'json' | 'markdown' | null>(null);

  const flash = (kind: 'text' | 'json' | 'markdown', content: string) => {
    navigator.clipboard.writeText(content).then(() => {
      setCopied(kind);
      setTimeout(() => setCopied(null), 1200);
    }).catch(() => {});
  };

  return (
    <div className="flex items-center gap-1 text-gray-400">
      {actions.text && (
        <button
          type="button"
          title={copied === 'text' ? 'Copied!' : `Copy ${title} as text`}
          className={`p-0.5 rounded hover:bg-gray-100 hover:text-gray-700 ${copied === 'text' ? 'text-green-600' : ''}`}
          onClick={(e) => { e.stopPropagation(); flash('text', actions.text!()); }}
        >
          <iconify-icon icon={copied === 'text' ? 'codicon:check' : 'codicon:copy'} className="text-sm" />
        </button>
      )}
      {actions.markdown && (
        <button
          type="button"
          title={copied === 'markdown' ? 'Copied!' : `Copy ${title} as Markdown`}
          className={`p-0.5 rounded hover:bg-gray-100 hover:text-gray-700 ${copied === 'markdown' ? 'text-green-600' : ''}`}
          onClick={(e) => { e.stopPropagation(); flash('markdown', actions.markdown!()); }}
        >
          <iconify-icon icon={copied === 'markdown' ? 'codicon:check' : 'codicon:markdown'} className="text-sm" />
        </button>
      )}
      {actions.json && (
        <button
          type="button"
          title={copied === 'json' ? 'Copied!' : `Copy ${title} as JSON`}
          className={`p-0.5 rounded hover:bg-gray-100 hover:text-gray-700 ${copied === 'json' ? 'text-green-600' : ''}`}
          onClick={(e) => { e.stopPropagation(); flash('json', JSON.stringify(actions.json!(), null, 2)); }}
        >
          <iconify-icon icon={copied === 'json' ? 'codicon:check' : 'codicon:json'} className="text-sm" />
        </button>
      )}
    </div>
  );
}
