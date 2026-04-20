import { useState, useMemo, useRef } from 'preact/hooks';
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

interface MetricCardProps {
  href: string;
  icon: string;
  label: string;
  value: string | number;
  sub?: string;
  tone: 'pass' | 'fail' | 'warn' | 'info' | 'neutral';
}

function MetricCard({ href, icon, label, value, sub, tone }: MetricCardProps) {
  const toneClass = {
    pass: 'bg-green-50 border-green-200 text-green-700 hover:bg-green-100',
    fail: 'bg-red-50 border-red-200 text-red-700 hover:bg-red-100',
    warn: 'bg-yellow-50 border-yellow-200 text-yellow-700 hover:bg-yellow-100',
    info: 'bg-blue-50 border-blue-200 text-blue-700 hover:bg-blue-100',
    neutral: 'bg-gray-50 border-gray-200 text-gray-600 hover:bg-gray-100',
  }[tone];
  return (
    <a
      href={href}
      class={`group block rounded-lg border px-3 py-2 transition-colors ${toneClass}`}
    >
      <div class="flex items-center justify-between">
        <iconify-icon icon={icon} class="text-lg" />
        <iconify-icon icon="codicon:chevron-right" class="text-xs opacity-30 group-hover:opacity-70" />
      </div>
      <div class="text-2xl font-semibold tabular-nums leading-tight mt-1">{value}</div>
      <div class="text-[11px] font-medium uppercase tracking-wide opacity-80">{label}</div>
      {sub && <div class="text-[11px] mt-0.5 opacity-70 truncate">{sub}</div>}
    </a>
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
  const link = (tab: string) => `${basePath}/${tab}?backTo=${encodeURIComponent(backTo)}`;

  const cards: MetricCardProps[] = [];

  if (results.testsTotal > 0) {
    cards.push({
      href: link('tests?filter=passed'),
      icon: 'codicon:pass',
      label: 'Passed',
      value: results.testsPassed,
      sub: `of ${results.testsTotal} test${results.testsTotal !== 1 ? 's' : ''}`,
      tone: 'pass',
    });
    cards.push({
      href: link('tests?filter=failed'),
      icon: 'codicon:error',
      label: 'Failed',
      value: results.testsFailed,
      sub: results.testsFailed > 0 ? 'need triage' : 'none',
      tone: results.testsFailed > 0 ? 'fail' : 'neutral',
    });
    if (results.testsSkipped > 0) {
      cards.push({
        href: link('tests?filter=skipped'),
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
      href: link('lint'),
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
      href: link('bench'),
      icon: regs > 0 ? 'codicon:arrow-down' : 'codicon:graph',
      label: 'Bench',
      value: regs,
      sub: regs > 0
        ? `regression${regs !== 1 ? 's' : ''}`
        : 'no regressions',
      tone: regs > 0 ? 'fail' : 'info',
    });
  }

  if (cards.length === 0) return null;

  const failures = results.topFailures ?? [];
  const lintHits = results.topLintViolations ?? [];

  return (
    <Section
      title="Gavel Results"
      actions={{
        json: () => results,
        text: () => formatGavelText(results),
        markdown: () => formatGavelMarkdown(results),
      }}
    >
      <div class="grid grid-cols-2 md:grid-cols-3 gap-2">
        {cards.map((c, i) => (
          <MetricCard key={i} {...c} />
        ))}
      </div>
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
    </Section>
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
    <div class="mt-3">
      <div class="flex items-center gap-1.5 text-[11px] uppercase tracking-wide text-gray-500 mb-1">
        <iconify-icon icon={icon} class={iconColor} />
        <span class="font-semibold">{title}</span>
        <span class="text-gray-400 normal-case tracking-normal">
          showing {shown} of {total}
        </span>
      </div>
      <div class="divide-y divide-gray-100 border border-gray-100 rounded">
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
    <div class="flex items-start gap-2 py-1.5 px-2 text-xs">
      {withChevron && (
        <iconify-icon icon="codicon:chevron-right" class="text-gray-400 mt-0.5 shrink-0 transition-transform group-open:rotate-90" />
      )}
      <iconify-icon icon="codicon:error" class="text-red-600 mt-0.5 shrink-0" />
      <div class="flex-1 min-w-0">
        <div class="font-medium text-gray-800 truncate" title={f.name}>
          {f.suite ? <span class="text-gray-400">{f.suite} › </span> : null}
          {f.name}
        </div>
        <div class="text-[11px] text-gray-500 truncate font-mono" title={`${location}${plainMsg ? ' — ' + plainMsg : ''}`}>
          {location && <span>{location}</span>}
          {location && plainMsg && <span class="mx-1">·</span>}
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
    <details class="group">
      <summary class="list-none cursor-pointer hover:bg-gray-50">
        <FailureHeader f={f} withChevron={true} />
      </summary>
      <pre
        class="text-[11px] font-mono text-gray-100 bg-[#1e1e1e] px-3 py-2 overflow-x-auto whitespace-pre-wrap border-t border-gray-200"
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
    <div class="flex items-start gap-2 py-1.5 px-2 text-xs">
      <iconify-icon icon="codicon:warning" class="text-yellow-600 mt-0.5 shrink-0" />
      <div class="flex-1 min-w-0">
        <div class="font-medium text-gray-800 truncate">
          <span class="text-gray-400">{v.linter}</span>
          {v.rule && <span class="ml-1 text-gray-500">({v.rule})</span>}
        </div>
        <div class="text-[11px] text-gray-500 truncate font-mono" title={`${location}${plainMsg ? ' — ' + plainMsg : ''}`}>
          {location && <span>{location}</span>}
          {location && plainMsg && <span class="mx-1">·</span>}
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
    <div class="mt-4">
      <div class="flex items-center justify-between mb-2 border-b border-gray-100 pb-1">
        <h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wide">{title}</h3>
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
    <div class="flex items-center gap-1 text-gray-400">
      {actions.text && (
        <button
          type="button"
          title={copied === 'text' ? 'Copied!' : `Copy ${title} as text`}
          class={`p-0.5 rounded hover:bg-gray-100 hover:text-gray-700 ${copied === 'text' ? 'text-green-600' : ''}`}
          onClick={(e) => { e.stopPropagation(); flash('text', actions.text!()); }}
        >
          <iconify-icon icon={copied === 'text' ? 'codicon:check' : 'codicon:copy'} class="text-sm" />
        </button>
      )}
      {actions.markdown && (
        <button
          type="button"
          title={copied === 'markdown' ? 'Copied!' : `Copy ${title} as Markdown`}
          class={`p-0.5 rounded hover:bg-gray-100 hover:text-gray-700 ${copied === 'markdown' ? 'text-green-600' : ''}`}
          onClick={(e) => { e.stopPropagation(); flash('markdown', actions.markdown!()); }}
        >
          <iconify-icon icon={copied === 'markdown' ? 'codicon:check' : 'codicon:markdown'} class="text-sm" />
        </button>
      )}
      {actions.json && (
        <button
          type="button"
          title={copied === 'json' ? 'Copied!' : `Copy ${title} as JSON`}
          class={`p-0.5 rounded hover:bg-gray-100 hover:text-gray-700 ${copied === 'json' ? 'text-green-600' : ''}`}
          onClick={(e) => { e.stopPropagation(); flash('json', JSON.stringify(actions.json!(), null, 2)); }}
        >
          <iconify-icon icon={copied === 'json' ? 'codicon:check' : 'codicon:json'} class="text-sm" />
        </button>
      )}
    </div>
  );
}
