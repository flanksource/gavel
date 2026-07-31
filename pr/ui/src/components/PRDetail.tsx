import { useState, useMemo, useRef, type ReactNode, type ComponentType } from 'react';
import type { PRItem, PRDetail, PRComment, PRCommitInfo, PRFileInfo, GavelResultsSummary, TestFailure, LintViolation, Project } from '../types';
import { stateColor, reviewColor, severityIcon, extractCommentTitle, isDeploymentComment } from '../utils';
import { CreateTodoFromPRDialog } from './todos/CreateTodoFromPRDialog';
import { PRActions, type ExtraAction } from './PRActions';
import { RelativeTime } from './RelativeTime';
import { ansiToHtml, stripAnsi } from '../ansi';
import { Markdown } from './Markdown';
import { Avatar } from './Avatar';
import { WorkflowRunView } from './WorkflowView';
import { BotCommentBody, BotBadge } from './BotComment';
import type { WorkflowRun } from '../types';
import {
  UiActivity,
  UiAdd,
  UiArrowDown,
  UiBeaker,
  UiCancel,
  UiCheck,
  UiChevronDown,
  UiChevronRight,
  UiChevronUp,
  UiClock,
  UiClose,
  UiComment,
  UiCopy,
  UiDebugStepOver,
  UiDiff,
  UiError,
  UiEye,
  UiGitGraph,
  UiGitMerge,
  UiGraph,
  UiJson,
  UiLinkExternal,
  UiMarkdown,
  UiPass,
  UiServerProcess,
  UiWarningTriangle,
} from '@flanksource/clicky-ui/icons';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../icons/Spinner';
import { VercelIcon } from '../icons/VercelIcon';
import { Button } from '@flanksource/clicky-ui/components';
import { AnsiHtml, GitChangedFilesSummary, GitCommitList, GitFileList, type GitCommitItem, type GitDiffPayload, type GitFileChangeItem } from '@flanksource/clicky-ui/data';
import { useTimeoutFlash } from '../useTimeoutFlash';
import { copyText } from '../clipboard';
import { useContainerWidth } from '../useContainerWidth';

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

type PRDiffAPIResponse = GitDiffPayload & { error?: string };

async function loadGitDiff(url: string): Promise<GitDiffPayload> {
  const response = await fetch(url);
  let payload: PRDiffAPIResponse = { diff: '' };
  try {
    payload = await response.json();
  } catch {
    payload = { diff: '', error: await response.text().catch(() => '') };
  }
  if (!response.ok) {
    throw new Error(payload.error || `Failed to load diff (${response.status})`);
  }
  return {
    diff: payload.diff || '',
    truncated: !!payload.truncated,
    binary: !!payload.binary,
  };
}

interface Props {
  pr: PRItem;
  detail: PRDetail | null;
  loading: boolean;
  // projects are the configured workspaces a PR-derived todo can be created in.
  // The dialog can register one when this list is absent or empty.
  projects?: Project[];
  // onTodoCreated lets the host refresh todo counts after one is added.
  onTodoCreated?: () => void;
  // onActionDone fires after a merge/approve/auto-merge action lands so the host
  // can re-fetch the PR detail and reflect the new state.
  onActionDone?: () => void;
  // onClose, when set, renders a close button in the header that dismisses the
  // panel. Only the desktop split-pane layout wires it; the menu-bar view has
  // its own back button and leaves this unset.
  onClose?: () => void;
}

type DetailTab = 'overview' | 'commits' | 'files';

export function PRDetailPanel({ pr, detail, loading, projects, onTodoCreated, onActionDone, onClose }: Props) {
  const [showCreate, setShowCreate] = useState(false);
  const [activeTab, setActiveTab] = useState<DetailTab>('overview');
  const workspaces = useMemo(() => (projects ?? []).filter(p => !!p.dir), [projects]);
  const [containerRef, containerWidth] = useContainerWidth<HTMLDivElement>();

  // Below this the header can't lay Approve / Merge / Update / New todo out
  // beside the title, so they collapse into an overflow dropdown. The menu-bar
  // window is 760 wide by default (min 520) and the desktop split pane can be
  // dragged narrower still, so this tracks the panel's own width, not the viewport.
  const collapsedActions = containerWidth > 0 && containerWidth < 560;
  const extraActions = useMemo<ExtraAction[]>(
    () => [{
      label: 'New todo',
      icon: UiAdd,
      onClick: () => setShowCreate(true),
      title: "Create a todo from this PR's failures and comments",
    }],
    [],
  );

  const info = detail?.pr;
  const commits = info?.prCommits ?? [];
  const files = info?.prFiles ?? [];

  return (
    <div ref={containerRef} className="p-4 bg-card h-full overflow-y-auto">
      <PRHeader
        pr={pr}
        detail={detail}
        action={
          <div className="flex items-center gap-2">
            <PRActions
              pr={pr}
              detail={detail}
              onChanged={onActionDone}
              collapsed={collapsedActions}
              extras={extraActions}
            />
            {onClose && (
              <Button
                variant="ghost"
                size="icon"
                type="button"
                onClick={onClose}
                title="Close"
                aria-label="Close pull request details"
                className="h-8 w-8 shrink-0 rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
              >
                <UiClose className="text-base" />
              </Button>
            )}
          </div>
        }
      />

      {/* PR Description / Body */}
      {info?.body && info.body.trim().length > 0 && (
        <PRBodySection body={info.body} />
      )}

      {loading && !detail && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground mt-4">
          <Spinner className="text-blue-500" />
          Loading details...
        </div>
      )}

      {detail?.error && (
        <div className="mt-3 p-2 bg-red-50 border border-red-100 rounded text-xs text-red-700">
          <UiError className="mr-1" />
          {detail.error}
        </div>
      )}

      {/* Detail tabs: Overview | Commits | Files */}
      <DetailTabBar
        active={activeTab}
        onChange={setActiveTab}
        commitCount={commits.length}
        fileCount={files.length}
      />

      {activeTab === 'overview' && (
        <>
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
        </>
      )}

      {activeTab === 'commits' && (
        <CommitsTab commits={commits} pr={pr} />
      )}

      {activeTab === 'files' && (
        <FilesTab files={files} pr={pr} info={info} />
      )}

      <div className="pt-3 mt-3 border-t border-border">
        <a href={pr.url} target="_blank" rel="noopener"
          className="inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800 hover:underline">
          <UiLinkExternal />
          Open on GitHub
        </a>
      </div>

      <CreateTodoFromPRDialog
        open={showCreate}
        onClose={() => setShowCreate(false)}
        pr={pr}
        detail={detail}
        workspaces={workspaces}
        onCreated={onTodoCreated}
        onProjectsChanged={onTodoCreated}
      />
    </div>
  );
}

function PRBodySection({ body }: { body: string }) {
  const [expanded, setExpanded] = useState(false);
  const isLong = body.length > 500;

  return (
    <Section title="Description">
      <div className={`relative ${!expanded && isLong ? 'max-h-40 overflow-hidden' : ''}`}>
        <Markdown text={body} className="text-xs text-foreground" />
        {!expanded && isLong && (
          <div className="absolute bottom-0 left-0 right-0 h-12 bg-gradient-to-t from-card to-transparent" />
        )}
      </div>
      {isLong && (
        <Button
          variant="ghost"
          className="text-xs text-blue-600 hover:text-blue-800 mt-1 h-auto p-0"
          onClick={() => setExpanded(!expanded)}
        >
          {expanded ? 'Show less' : 'Show more'}
        </Button>
      )}
    </Section>
  );
}

function DetailTabBar({ active, onChange, commitCount, fileCount }: {
  active: DetailTab;
  onChange: (t: DetailTab) => void;
  commitCount: number;
  fileCount: number;
}) {
  const tabs: { id: DetailTab; label: string; icon: ComponentType<IconProps>; count?: number }[] = [
    { id: 'overview', label: 'Overview', icon: UiActivity },
    { id: 'commits', label: 'Commits', icon: UiGitGraph, count: commitCount },
    { id: 'files', label: 'Files', icon: UiDiff, count: fileCount },
  ];

  return (
    <div className="flex items-center gap-1 mt-3 mb-1 border-b border-border">
      {tabs.map(t => {
        const Icon = t.icon;
        return (
        <Button
          key={t.id}
          variant="ghost"
          className={`inline-flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-none border-b-2 transition-colors h-auto ${
            active === t.id
              ? 'border-blue-600 text-blue-600 font-medium'
              : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
          }`}
          onClick={() => onChange(t.id)}
        >
          <Icon className="text-xs" />
          {t.label}
          {t.count != null && t.count > 0 && (
            <span className="text-[10px] bg-muted rounded-full px-1.5 py-0 font-medium">{t.count}</span>
          )}
        </Button>
        );
      })}
    </div>
  );
}

function CommitsTab({ commits, pr }: { commits: PRCommitInfo[]; pr: PRItem }) {
  const items = useMemo<GitCommitItem[]>(() => commits.map(commit => ({
    id: commit.oid,
    sha: commit.oid,
    shortSha: commit.oid.slice(0, 7),
    title: commit.messageHeadline,
    body: commit.messageBody,
    authorName: commit.authorName,
    authorLogin: commit.authorLogin,
    authorAvatarUrl: commit.authorAvatarUrl,
    committedAt: commit.committedDate,
    additions: commit.additions,
    deletions: commit.deletions,
    changedFiles: commit.changedFiles,
    href: `https://github.com/${pr.repo}/commit/${commit.oid}`,
  })), [commits, pr.repo]);

  const loadDiff = (commit: GitCommitItem) => loadGitDiff(
    `/api/prs/commits/diff?repo=${encodeURIComponent(pr.repo)}&sha=${encodeURIComponent(commit.sha)}`,
  );

  return (
    <GitCommitList
      commits={items}
      loadDiff={loadDiff}
      renderTime={iso => <RelativeTime iso={iso} />}
      className="mt-2"
      diffMaxHeightClassName="max-h-[560px]"
    />
  );
}

function FilesTab({ files, pr, info }: { files: PRFileInfo[]; pr: PRItem; info?: import('../types').PRInfo }) {
  const totalAdds = info?.additions ?? files.reduce((s, f) => s + f.additions, 0);
  const totalDels = info?.deletions ?? files.reduce((s, f) => s + f.deletions, 0);
  const items = useMemo<GitFileChangeItem[]>(() => files.map(file => ({
    id: file.path,
    path: file.path,
    status: normalizePRFileStatus(file.changeType),
    additions: file.additions,
    deletions: file.deletions,
    href: `${pr.url}/files#diff-${btoa(file.path).replace(/=/g, '')}`,
  })), [files, pr.url]);

  const loadDiff = (file: GitFileChangeItem) => loadGitDiff(
    `/api/prs/files/diff?repo=${encodeURIComponent(pr.repo)}&number=${pr.number}&path=${encodeURIComponent(file.path)}`,
  );

  return (
    <div className="mt-2 space-y-2">
      <GitChangedFilesSummary files={files.length} additions={totalAdds} deletions={totalDels} />
      <GitFileList
        files={items}
        loadDiff={loadDiff}
        diffMaxHeightClassName="max-h-[560px]"
      />
    </div>
  );
}

function normalizePRFileStatus(changeType: string): string {
  switch (changeType.toUpperCase()) {
    case 'ADDED':
      return 'added';
    case 'DELETED':
    case 'REMOVED':
      return 'deleted';
    case 'RENAMED':
      return 'renamed';
    case 'COPIED':
      return 'copied';
    default:
      return 'modified';
  }
}

function PRHeader({ pr, detail, action }: { pr: PRItem; detail: PRDetail | null; action?: ReactNode }) {
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
          <h2 className="text-base font-semibold text-foreground">{pr.title}</h2>
          <div className="flex items-center gap-1 text-xs text-muted-foreground mt-0.5">
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
            <RelativeTime iso={pr.updatedAt} />
          </div>
        </div>
        {action}
      </div>

      <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm mb-3">
        <div>
          <span className="text-cyan-600 font-mono text-xs">{pr.source}</span>
          <span className="text-muted-foreground mx-1">→</span>
          <span className="text-cyan-600 font-mono text-xs">{pr.target}</span>
        </div>
        <div className="flex gap-2">
          <span className={stateColor(pr.state, pr.isDraft)}>
            {pr.isDraft ? 'Draft' : pr.state}
          </span>
          {(pr.reviewDecision || info?.reviewDecision) && (
            <>
              <span className="text-muted-foreground/50">|</span>
              <span className={reviewColor(pr.reviewDecision || info?.reviewDecision || '')}>
                {(pr.reviewDecision || info?.reviewDecision || '').replace(/_/g, ' ')}
              </span>
            </>
          )}
          {(pr.mergeable || info?.mergeable) && (() => {
            const m = pr.mergeable || info?.mergeable || '';
            return (
              <>
                <span className="text-muted-foreground/50">|</span>
                <span className={m === 'MERGEABLE' ? 'text-green-600' : m === 'CONFLICTING' ? 'text-red-600' : 'text-yellow-600'}>
                  {m === 'CONFLICTING' && <UiGitMerge className="mr-0.5" />}
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
  const title = extractCommentTitle(comment.body);

  return (
    <div className={`text-xs border-b border-border ${resolved ? 'opacity-50' : ''}`}>
      <div
        className="flex items-start gap-1.5 py-1.5 cursor-pointer hover:bg-muted rounded px-1 -mx-1"
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
            {resolved && <span className="text-muted-foreground text-[10px]">({comment.isOutdated ? 'outdated' : 'resolved'})</span>}
          </div>
          <div className={`mt-0.5 ${resolved ? 'line-through text-muted-foreground' : 'text-foreground'}`}>
            {title.length > 120 ? title.slice(0, 117) + '...' : title}
          </div>
        </div>
        <span className="inline-flex items-center gap-1 text-muted-foreground shrink-0">
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
        {expanded
          ? <UiChevronUp className="text-muted-foreground shrink-0 text-[10px] mt-1" />
          : <UiChevronDown className="text-muted-foreground shrink-0 text-[10px] mt-1" />}
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
            : <Markdown text={comment.body} className="text-xs text-foreground" />}
        </div>
      )}
    </div>
  );
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

const deployStatusConfig: Record<string, { icon: ComponentType<IconProps>; color: string; label: string }> = {
  DEPLOYED:  { icon: UiPass,    color: 'text-green-600', label: 'Deployed' },
  BUILDING:  { icon: Spinner,   color: 'text-yellow-600', label: 'Building' },
  QUEUED:    { icon: UiClock,   color: 'text-muted-foreground',  label: 'Queued' },
  ERROR:     { icon: UiError,   color: 'text-red-600',   label: 'Error' },
  CANCELED:  { icon: UiCancel,  color: 'text-muted-foreground',  label: 'Canceled' },
};

function DeploymentRow({ project }: { project: VercelProject }) {
  const [hover, setHover] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const st = deployStatusConfig[project.status] || deployStatusConfig.QUEUED;
  const StatusIcon = st.icon;

  return (
    <div className="relative" ref={ref}
      onMouseEnter={() => setHover(true)} onMouseLeave={() => setHover(false)}
    >
      <div className="flex items-center gap-2 py-1.5 px-1 -mx-1 rounded hover:bg-muted text-sm transition-colors">
        <StatusIcon className={`${st.color} text-xs`} />
        <a href={project.previewUrl} target="_blank" rel="noopener"
          className="text-blue-600 hover:underline font-medium flex-1 truncate"
        >
          {project.name}
        </a>
        <a href={project.inspectorUrl} target="_blank" rel="noopener"
          className="text-muted-foreground hover:text-muted-foreground p-0.5 rounded hover:bg-muted transition-colors"
          title="Build output"
          onClick={(e) => e.stopPropagation()}
        >
          <UiServerProcess className="text-xs" />
        </a>
      </div>
      {hover && (
        <div className="absolute left-0 top-full z-50 mt-0.5 w-72 bg-popover border border-border rounded-lg shadow-lg p-3 text-xs">
          <div className="flex items-center gap-1.5 mb-2">
            <VercelIcon className="text-sm" />
            <span className="font-semibold text-foreground">{project.name}</span>
            <span className={`ml-auto inline-flex items-center gap-1 ${st.color}`}>
              <StatusIcon className="text-[10px]" />
              {st.label}
            </span>
          </div>
          <div className="space-y-1.5 text-muted-foreground">
            <div className="flex items-center gap-1.5">
              <UiLinkExternal className="text-muted-foreground text-[10px] shrink-0" />
              <a href={project.previewUrl} target="_blank" rel="noopener"
                className="text-blue-600 hover:underline truncate">
                {project.previewUrl.replace(/^https?:\/\//, '')}
              </a>
            </div>
            <div className="flex items-center gap-1.5">
              <UiServerProcess className="text-muted-foreground text-[10px] shrink-0" />
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
  { key: 'nitpick', icon: '🧹', label: 'Nitpick', color: 'border-border bg-muted' },
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
            <Button key={sf.key}
              variant="ghost"
              className={`inline-flex items-center gap-1 text-[11px] px-1.5 py-0.5 rounded-full border transition-colors h-auto ${
                active ? `${sf.color} font-medium` : 'border-border text-muted-foreground hover:bg-muted'
              }`}
              onClick={() => toggleSeverity(sf.key)}
            >
              <span>{sf.icon}</span>
              <span>{count}</span>
            </Button>
          );
        })}
        {severityCounts[''] > 0 && (
          <Button
            variant="ghost"
            className={`inline-flex items-center gap-1 text-[11px] px-1.5 py-0.5 rounded-full border transition-colors h-auto ${
              severityFilter.has('') ? 'border-border bg-muted font-medium' : 'border-border text-muted-foreground hover:bg-muted'
            }`}
            onClick={() => toggleSeverity('')}
          >
            <UiComment className="text-[11px]" />
            <span>{severityCounts['']}</span>
          </Button>
        )}
        {severityCounts._outdated > 0 && (
          <>
            <span className="text-muted-foreground/50 mx-0.5">|</span>
            <Button
              variant="ghost"
              className={`inline-flex items-center gap-1 text-[11px] px-1.5 py-0.5 rounded-full border transition-colors h-auto ${
                showOutdated ? 'border-border bg-muted font-medium' : 'border-border text-muted-foreground hover:bg-muted'
              }`}
              onClick={() => setShowOutdated(!showOutdated)}
            >
              <UiEye className="text-[10px]" />
              {severityCounts._outdated} resolved
            </Button>
          </>
        )}
        {severityFilter.size > 0 && (
          <Button variant="ghost" className="text-[11px] text-muted-foreground hover:text-muted-foreground ml-0.5 h-auto p-0"
            onClick={() => setSeverityFilter(new Set())}>
            Clear
          </Button>
        )}
      </div>
      {filtered.map(c => (
        <CommentView key={c.id} comment={c} />
      ))}
      {filtered.length === 0 && (
        <div className="text-xs text-muted-foreground py-2">No comments match filters</div>
      )}
    </Section>
  );
}

interface MetricCardProps {
  href?: string;
  icon: ComponentType<IconProps>;
  label: string;
  value: string | number;
  sub?: string;
  tone: 'pass' | 'fail' | 'warn' | 'info' | 'neutral';
}

function MetricCard({ href, icon: Icon, label, value, sub, tone }: MetricCardProps) {
  const toneClass = {
    pass: href ? 'bg-green-50 border-green-200 text-green-700 hover:bg-green-100' : 'bg-green-50 border-green-200 text-green-700',
    fail: href ? 'bg-red-50 border-red-200 text-red-700 hover:bg-red-100' : 'bg-red-50 border-red-200 text-red-700',
    warn: href ? 'bg-yellow-50 border-yellow-200 text-yellow-700 hover:bg-yellow-100' : 'bg-yellow-50 border-yellow-200 text-yellow-700',
    info: href ? 'bg-blue-50 border-blue-200 text-blue-700 hover:bg-blue-100' : 'bg-blue-50 border-blue-200 text-blue-700',
    neutral: href ? 'bg-muted border-border text-muted-foreground hover:bg-muted' : 'bg-muted border-border text-muted-foreground',
  }[tone];
  const body = (
    <>
      <div className="flex items-center justify-between">
        <Icon className="text-lg" />
        {href && <UiChevronRight className="text-xs opacity-30 group-hover:opacity-70" />}
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
  // A shard that produced no cards but carries an error knows why it produced
  // nothing — say that instead of the generic "no data" text.
  const errored = shards.filter(s => s.error);

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
        <div className="grid grid-cols-[repeat(auto-fill,minmax(7rem,1fr))] gap-2">
          {headerCards.map((c, i) => (
            <MetricCard key={i} {...c} />
          ))}
        </div>
      ) : errored.length > 0 ? (
        <div className="divide-y divide-border">
          {errored.map(s => (
            <div key={s.stickyId || s.artifactId} className="py-1 first:pt-0">
              {multi && (
                <div className="text-[11px] font-mono text-muted-foreground">
                  {s.stickyId || `artifact ${s.artifactId}`}
                </div>
              )}
              <ShardError results={s} href={shardLink(pr, s, 'tests')} />
            </div>
          ))}
        </div>
      ) : (
        <div className="text-xs text-muted-foreground py-1">
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
          <Button
            type="button"
            variant="ghost"
            className="flex items-center gap-1 text-[11px] uppercase tracking-wide text-muted-foreground hover:text-foreground justify-start h-auto p-0"
            onClick={() => setBreakdownOpen(o => !o)}
            aria-expanded={breakdownOpen}
          >
            {breakdownOpen
              ? <UiChevronDown className="text-muted-foreground" />
              : <UiChevronRight className="text-muted-foreground" />}
            <span className="font-semibold">Per-shard breakdown</span>
            <span className="text-muted-foreground normal-case tracking-normal">
              ({shards.length} shard{shards.length !== 1 ? 's' : ''})
            </span>
          </Button>
          {breakdownOpen && (
            <div className="mt-2 divide-y divide-border border border-border rounded">
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
      icon: UiPass,
      label: 'Passed',
      value: results.testsPassed,
      sub: `of ${results.testsTotal} test${results.testsTotal !== 1 ? 's' : ''}`,
      tone: 'pass',
    });
    cards.push({
      href: href('tests?filter=failed'),
      icon: UiError,
      label: 'Failed',
      value: results.testsFailed,
      sub: results.testsFailed > 0 ? 'need triage' : 'none',
      tone: results.testsFailed > 0 ? 'fail' : 'neutral',
    });
    if (results.testsSkipped > 0) {
      cards.push({
        href: href('tests?filter=skipped'),
        icon: UiDebugStepOver,
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
      icon: results.lintViolations > 0 ? UiWarningTriangle : UiPass,
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
      icon: regs > 0 ? UiArrowDown : UiGraph,
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

// LOG_TAIL_LINES bounds how much of a crash log is rendered inline; the full
// log is uploaded alongside the artifact.
const LOG_TAIL_LINES = 40;

// ShardError explains an artifact that carries no results: either gavel crashed
// before producing any (exit code + log tail) or the artifact could not be read.
function ShardError({ results, href }: { results: GavelResultsSummary; href?: string }) {
  const lines = (results.logTail ?? '').replace(/\n+$/, '').split('\n');
  const tail = lines.slice(-LOG_TAIL_LINES).join('\n');
  const dropped = lines.length - LOG_TAIL_LINES;

  return (
    <div className="text-xs py-1 space-y-1.5">
      <div className="flex items-start gap-1.5">
        <UiWarningTriangle className="text-yellow-500 shrink-0 mt-0.5" />
        <span className="text-foreground">{results.error}</span>
      </div>
      {results.exitCode !== undefined && (
        <div className="text-muted-foreground">
          Exit code <span className="font-mono tabular-nums text-foreground">{results.exitCode}</span>
        </div>
      )}
      {tail && (
        <details>
          <summary className="cursor-pointer text-muted-foreground hover:text-foreground">
            Log tail{dropped > 0 ? ` (last ${LOG_TAIL_LINES} of ${lines.length} lines)` : ''}
          </summary>
          <AnsiHtml
            as="pre"
            text={tail}
            className="mt-1 max-h-64 overflow-auto rounded bg-black p-2 text-[11px] leading-snug text-gray-100 whitespace-pre-wrap"
          />
        </details>
      )}
      {href && <a className="text-blue-600 hover:underline" href={href}>Open results</a>}
    </div>
  );
}

function ShardExtras({ results }: { results: GavelResultsSummary }) {
  const failures = results.topFailures ?? [];
  const lintHits = results.topLintViolations ?? [];
  return (
    <>
      {failures.length > 0 && (
        <FailureList
          title="Test failures"
          icon={UiBeaker}
          iconColor="text-red-600"
          total={results.testsFailed}
        >
          {failures.map((f, i) => <TestFailureRow key={i} f={f} />)}
        </FailureList>
      )}
      {lintHits.length > 0 && (
        <FailureList
          title="Lint violations"
          icon={UiWarningTriangle}
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
        <UiWarningTriangle />
      </span>
    );
  }
  const items: { icon: ComponentType<IconProps>; color: string; count: number; title: string }[] = [];
  if (g.testsPassed > 0) {
    items.push({ icon: UiPass, color: 'text-green-600', count: g.testsPassed, title: `${g.testsPassed} passed` });
  }
  if (g.testsFailed > 0) {
    items.push({ icon: UiError, color: 'text-red-600', count: g.testsFailed, title: `${g.testsFailed} failed` });
  }
  if (g.testsSkipped > 0) {
    items.push({ icon: UiDebugStepOver, color: 'text-muted-foreground', count: g.testsSkipped, title: `${g.testsSkipped} skipped` });
  }
  if (g.lintViolations > 0) {
    items.push({ icon: UiWarningTriangle, color: 'text-yellow-600', count: g.lintViolations, title: `${g.lintViolations} lint` });
  }
  if ((g.benchRegressions ?? 0) > 0) {
    items.push({ icon: UiArrowDown, color: 'text-red-600', count: g.benchRegressions ?? 0, title: `${g.benchRegressions} bench regression` });
  }
  if (items.length === 0) return null;
  return (
    <span className="inline-flex items-center gap-1 tabular-nums">
      {items.map((it, i) => {
        const Icon = it.icon;
        return (
        <span key={i} className={`inline-flex items-center ${it.color} leading-none`} title={it.title}>
          <Icon className="text-[12px]" />
          <span className="text-[11px] font-medium">{it.count}</span>
        </span>
        );
      })}
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
      <Button
        type="button"
        variant="ghost"
        className="w-full flex items-center gap-2 px-2 py-1.5 hover:bg-muted text-left justify-start h-auto"
        onClick={() => setOpen(o => !o)}
        aria-expanded={open}
      >
        {open
          ? <UiChevronDown className="text-muted-foreground text-[12px] shrink-0" />
          : <UiChevronRight className="text-muted-foreground text-[12px] shrink-0" />}
        <span className="text-xs font-mono text-foreground truncate">{label}</span>
        <ShardSummaryBadges g={results} />
      </Button>
      {open && (
        <div className="px-2 pb-3 pt-1">
          {results.error ? (
            <ShardError results={results} href={link('tests')} />
          ) : cards.length === 0 ? (
            <div className="text-xs text-muted-foreground py-1">
              No test, lint, or bench data in this artifact.{' '}
              <a className="text-blue-600 hover:underline" href={link('tests')}>Open results</a>
            </div>
          ) : (
            <>
              <div className="grid grid-cols-[repeat(auto-fill,minmax(7rem,1fr))] gap-2">
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

function FailureList({ title, icon: Icon, iconColor, total, children }: {
  title: string;
  icon: ComponentType<IconProps>;
  iconColor: string;
  total: number;
  children: any;
}) {
  const rows = Array.isArray(children) ? children : [children];
  const shown = rows.length;
  return (
    <div className="mt-3">
      <div className="flex items-center gap-1.5 text-[11px] uppercase tracking-wide text-muted-foreground mb-1">
        <Icon className={iconColor} />
        <span className="font-semibold">{title}</span>
        <span className="text-muted-foreground normal-case tracking-normal">
          showing {shown} of {total}
        </span>
      </div>
      <div className="divide-y divide-border border border-border rounded">
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
        <UiChevronRight className="text-muted-foreground mt-0.5 shrink-0 transition-transform group-open:rotate-90" />
      )}
      <UiError className="text-red-600 mt-0.5 shrink-0" />
      <div className="flex-1 min-w-0">
        <div className="font-medium text-foreground truncate" title={f.name}>
          {f.suite ? <span className="text-muted-foreground">{f.suite} › </span> : null}
          {f.name}
        </div>
        <div className="text-[11px] text-muted-foreground truncate font-mono" title={`${location}${plainMsg ? ' — ' + plainMsg : ''}`}>
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
      <summary className="list-none cursor-pointer hover:bg-muted">
        <FailureHeader f={f} withChevron={true} />
      </summary>
      <pre
        className="text-[11px] font-mono text-gray-100 bg-black px-3 py-2 overflow-x-auto whitespace-pre-wrap border-t border-border"
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
      <UiWarningTriangle className="text-yellow-600 mt-0.5 shrink-0" />
      <div className="flex-1 min-w-0">
        <div className="font-medium text-foreground truncate">
          <span className="text-muted-foreground">{v.linter}</span>
          {v.rule && <span className="ml-1 text-muted-foreground">({v.rule})</span>}
        </div>
        <div className="text-[11px] text-muted-foreground truncate font-mono" title={`${location}${plainMsg ? ' — ' + plainMsg : ''}`}>
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
      <div className="flex items-center justify-between mb-2 border-b border-border pb-1">
        <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">{title}</h3>
        {actions && <SectionActionsBar actions={actions} title={title} />}
      </div>
      {children}
    </div>
  );
}

type CopyKind = 'text' | 'markdown' | 'json';
type CopyFlash = { kind: CopyKind; status: 'copied' | 'error' };

function SectionActionsBar({ actions, title }: { actions: SectionActions; title: string }) {
  const [flash, setFlash] = useTimeoutFlash<CopyFlash | null>(null, 1200);

  const copy = (kind: CopyKind, content: string) => {
    copyText(content)
      .then(() => setFlash({ kind, status: 'copied' }))
      .catch(() => setFlash({ kind, status: 'error' }));
  };

  return (
    <div className="flex items-center gap-1 text-muted-foreground">
      {actions.text && (
        <CopyActionButton kind="text" idle={UiCopy} label={`Copy ${title} as text`} flash={flash}
          onCopy={() => copy('text', actions.text!())} />
      )}
      {actions.markdown && (
        <CopyActionButton kind="markdown" idle={UiMarkdown} label={`Copy ${title} as Markdown`} flash={flash}
          onCopy={() => copy('markdown', actions.markdown!())} />
      )}
      {actions.json && (
        <CopyActionButton kind="json" idle={UiJson} label={`Copy ${title} as JSON`} flash={flash}
          onCopy={() => copy('json', JSON.stringify(actions.json!(), null, 2))} />
      )}
    </div>
  );
}

function CopyActionButton({ kind, idle: Idle, label, flash, onCopy }: {
  kind: CopyKind;
  idle: ComponentType<IconProps>;
  label: string;
  flash: CopyFlash | null;
  onCopy: () => void;
}) {
  const status = flash?.kind === kind ? flash.status : null;
  const tone = status === 'copied' ? 'text-green-600' : status === 'error' ? 'text-red-600' : '';
  return (
    <Button
      type="button"
      variant="ghost"
      title={status === 'copied' ? 'Copied!' : status === 'error' ? 'Copy failed' : label}
      className={`p-0.5 rounded hover:bg-muted hover:text-foreground h-auto ${tone}`}
      onClick={(e) => { e.stopPropagation(); onCopy(); }}
    >
      {status === 'copied'
        ? <UiCheck className="text-sm" />
        : status === 'error'
          ? <UiCancel className="text-sm" />
          : <Idle className="text-sm" />}
    </Button>
  );
}
