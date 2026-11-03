import { useState, type ReactNode } from 'react';
import { Button, Combobox, SegmentedControl, type ComboboxOption } from '@flanksource/clicky-ui/components';
import { UiArrowDown, UiArrowUp, UiCheck, UiEdit, UiFolderGit, UiGitBranch, UiGitCommit, UiHome, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import { useProjectGit } from '../projectGitQueries';
import { parseProjectRef } from '../projectRef';
import type { BranchMergeResult, GitBranchInfo, GitWorktree, Project, ProjectGit } from '../types';
import { GIT_HOT_REFRESH_MS, useGitFocus } from '../useGitFocus';
import { useNow } from '../useNow';
import { useIsMobile } from '../useIsMobile';
import { ProjectBranchChanges } from './ProjectBranchChanges';
import { ProjectStatusView } from './ProjectStatusView';
import { ProjectRebase } from './ProjectRebase';
import { dirtyFileCount, gitStateAge, gitStateScannedAt, projectRefAges, projectRefEntries, type ProjectRefEntry } from './projectGitView';

type WorktreeView = 'work' | 'branch';

interface Props {
  project: Project;
  // Raw ?ref= value: "" (main checkout), "wt:<path>" or "br:<name>".
  projectRef: string;
  diffPath: string;
  showResults: boolean;
  onRefChange: (ref: string) => void;
  onDiffPathChange: (path: string) => void;
  onChanged: () => void;
}

interface Landed {
  branch: string;
  result: BranchMergeResult;
}

const shortSha = (sha: string) => sha.slice(0, 7);

/**
 * The project detail pane for one git ref. A picker above the content switches
 * between the main checkout, each linked worktree and each unmerged branch
 * without a worktree. The main checkout and a worktree review and commit
 * working-tree changes; a checkout that is ahead of the base, and a branch,
 * also show their committed changes with the actions that land them.
 */
export function ProjectRefPane({ project, projectRef, diffPath, showResults, onRefChange, onDiffPathChange, onChanged }: Props) {
  const reviewing = useIsMobile() && diffPath !== '';
  const gitQuery = useProjectGit(project.name, { refetchIntervalMs: GIT_HOT_REFRESH_MS });
  const [landed, setLanded] = useState<Landed | null>(null);
  const [viewChoice, setViewChoice] = useState<{ ref: string; view: WorktreeView } | null>(null);
  const ref = parseProjectRef(projectRef);
  const git = gitQuery.data;
  // ProjectStatusView holds the lease for the main checkout and for a worktree
  // shown by its working changes; a branch without a worktree has no status
  // view, so it keeps the project's main checkout fresh.
  const focus = useGitFocus({ project: project.name, enabled: ref.kind === 'branch' });
  const shownWorktree = ref.kind === 'worktree'
    ? git?.worktrees.find(candidate => !candidate.primary && candidate.path === ref.path)
    : ref.kind === 'main' ? git?.worktrees.find(candidate => candidate.primary) : undefined;
  const notices = [
    git?.error && `Git refresh failed: ${git.error}`,
    shownWorktree?.statusError && `Working tree status failed: ${shownWorktree.statusError}`,
    focus.error && `Git tracker focus failed: ${focus.error}`,
  ].filter((notice): notice is string => !!notice);

  const onMerged = (branch: string) => (result: BranchMergeResult) => {
    setLanded({ branch, result });
    onRefChange('');
  };

  let body;
  if (git && shownWorktree && (ref.kind === 'worktree' || (shownWorktree.ahead > 0 && git.branches.some(branch => branch.name === shownWorktree.branch)))) {
    const view = viewChoice?.ref === projectRef ? viewChoice.view : undefined;
    body = <WorktreeBody project={project} git={git} worktree={shownWorktree} view={view} onViewChange={next => setViewChoice({ ref: projectRef, view: next })} diffPath={diffPath} showResults={showResults} onDiffPathChange={onDiffPathChange} onChanged={onChanged} onMerged={onMerged(shownWorktree.branch)} />;
  } else if (ref.kind === 'main') {
    body = <ProjectStatusView key={`${project.name}\0`} project={project} diffPath={diffPath} showResults={showResults} onDiffPathChange={onDiffPathChange} onChanged={onChanged} />;
  } else if (!git) {
    body = gitQuery.error
      ? <Gone message={gitQuery.error.message} onBack={() => onRefChange('')} />
      : <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground">Loading git state…</div>;
  } else if (ref.kind === 'worktree') {
    body = <Gone message={`Worktree ${ref.path} no longer exists.`} onBack={() => onRefChange('')} />;
  } else {
    const branch = git.branches.find(candidate => candidate.name === ref.name && candidate.worktree === '');
    body = branch
      ? <div className="h-full min-h-0 overflow-auto"><ProjectBranchChanges projectName={project.name} base={git.base} branch={branch} baseCheckedOut={git.baseCheckedOut} dirtyCount={0} onMerged={onMerged(branch.name)} /></div>
      : <Gone message={`Branch ${ref.name} no longer has commits to review.`} onBack={() => onRefChange('')} />;
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div hidden={reviewing} className={reviewing ? 'hidden' : 'contents'}>
      <RefBar projectName={project.name} git={git} scannedAt={git ? gitStateScannedAt(git.computedAt, shownWorktree) : ''} loading={gitQuery.isPending} error={gitQuery.error?.message ?? ''} notices={notices} value={projectRef} onChange={onRefChange}
        actions={<ProjectRebase key={`${project.name}:${shownWorktree?.path ?? projectRef}`} projectName={project.name} worktree={shownWorktree?.path} branch={shownWorktree?.branch} generation={git?.generation} />}
      />
      {landed && <LandedBanner landed={landed} onDismiss={() => setLanded(null)} />}
      </div>
      <div className="min-h-0 flex-1">{body}</div>
    </div>
  );
}

function RefBar({ projectName, git, scannedAt, loading, error, notices, value, onChange, actions }: { projectName: string; git: ProjectGit | undefined; scannedAt: string; loading: boolean; error: string; notices: string[]; value: string; onChange: (ref: string) => void; actions: ReactNode }) {
  const entries = git ? projectRefEntries(git) : [];
  const selected = entries.find(entry => entry.value === value);
  const options: ComboboxOption[] = entries.map(entry => ({
    value: optionValue(entry.value),
    label: entry.label,
    group: entry.group,
    title: entry.title,
    icon: GROUP_ICONS[entry.group],
    selectedLabel: `${projectName} · ${entry.label}`,
    layout: 'responsive',
    trailing: <RefSummary entry={entry} />,
  }));
  return (
    <div className="@container/project-ref shrink-0 border-b border-border">
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1.5 px-3 py-2 md:px-4">
        {error ? <span role="alert" className="flex min-w-0 items-center gap-1 text-xs text-destructive">
          <UiWarningTriangle />Worktrees and branches unavailable: {error}
        </span> : loading ? <span className="text-xs text-muted-foreground">Loading git state…</span> : entries.length > 1 ? <Combobox
          ariaLabel="Worktree or branch"
          value={optionValue(value)}
          onChange={next => onChange(next === MAIN_OPTION ? '' : next)}
          options={selected ? options : [...options, { value, label: 'Unavailable', disabled: true }]}
          allowCustomValue={false}
          required
          placeholder="Worktree or branch"
          menuWidth={720}
          className="w-full min-w-0 @min-[48rem]/project-ref:max-w-[42rem] @min-[48rem]/project-ref:flex-1"
        /> : <span className="font-semibold">{projectName}</span>}
        {selected && <RefSummary entry={selected} compact />}
        <GitNotices notices={notices} />
        {actions}
        {scannedAt && <GitStateAge scannedAt={scannedAt} />}
      </div>
    </div>
  );
}

// A failed scan leaves the rest of the state stale rather than empty, so the
// failure is shown beside the age that tells how stale it is.
function GitNotices({ notices }: { notices: string[] }) {
  return (
    <>
      {notices.map(notice => (
        <span key={notice} role="alert" className="flex min-w-0 items-center gap-1 text-xs text-amber-700 dark:text-amber-400" title={notice}>
          <UiWarningTriangle className="shrink-0" /><span className="truncate">{notice}</span>
        </span>
      ))}
    </>
  );
}

function GitStateAge({ scannedAt }: { scannedAt: string }) {
  useNow();
  return (
    <span className="ml-auto shrink-0 whitespace-nowrap text-xs text-muted-foreground" title={`Git state scanned ${new Date(scannedAt).toLocaleString()}`}>
      {gitStateAge(scannedAt)}
    </span>
  );
}

// Combobox treats "" as no selection, so the main checkout's empty ref is
// carried as a sentinel inside the picker only.
const MAIN_OPTION = '@main';
const optionValue = (ref: string) => ref || MAIN_OPTION;

const GROUP_ICONS: Record<ProjectRefEntry['group'], ReactNode> = {
  Checkout: <UiHome />,
  Worktrees: <UiFolderGit />,
  Branches: <UiGitBranch />,
};

const plural = (n: number, noun: string) => `${n} ${noun}${n === 1 ? '' : 's'}`;

/** Counts and ages wrap as groups when the picker has limited space. */
function RefSummary({ entry: { stats, lastCommitAt, touchedAt }, className = 'flex', compact = false }: { entry: ProjectRefEntry; className?: string; compact?: boolean }) {
  useNow();
  const { committed, touched } = projectRefAges({ lastCommitAt, touchedAt });
  const upstream = stats.upstream;
  const upstreamLabel = upstream && `${plural(upstream.ahead, 'commit')} ahead of ${upstream.name}, ${upstream.behind} behind`;
  return (
    <span className={`${className} min-w-0 flex-wrap items-center gap-x-2 gap-y-1 whitespace-nowrap text-xs tabular-nums text-muted-foreground`}>
      {upstream && (upstream.ahead > 0 || upstream.behind > 0) && (
        <span className="flex items-center gap-1" title={upstreamLabel} aria-label={upstreamLabel}>
          {upstream.ahead > 0 && <span className="flex items-center"><UiArrowUp />{upstream.ahead}</span>}
          {upstream.behind > 0 && <span className="flex items-center"><UiArrowDown />{upstream.behind}</span>}
        </span>
      )}
      {stats.commits > 0 && <span>{plural(stats.commits, 'commit')}</span>}
      {stats.uncommitted > 0 && <span>{stats.uncommitted} uncommitted</span>}
      {stats.lines && (
        <span className="flex items-center gap-1 font-medium">
          <span className="text-green-600 dark:text-green-400">+{stats.lines.adds}</span>
          <span className="text-red-600 dark:text-red-400">−{stats.lines.dels}</span>
        </span>
      )}
      {committed && <span className={`${compact ? 'hidden md:flex' : 'flex'} items-center gap-0.5`} title={`Last commit ${new Date(lastCommitAt).toLocaleString()}`}><UiGitCommit />{committed}</span>}
      {touched && touchedAt && <span className={`${compact ? 'hidden md:flex' : 'flex'} items-center gap-0.5`} title={`Uncommitted files last touched ${new Date(touchedAt).toLocaleString()}`}><UiEdit />{touched}</span>}
    </span>
  );
}

function WorktreeBody({ project, git, worktree, view, onViewChange, diffPath, showResults, onDiffPathChange, onChanged, onMerged }: {
  project: Project;
  git: ProjectGit;
  worktree: GitWorktree;
  view: WorktreeView | undefined;
  onViewChange: (view: WorktreeView) => void;
  diffPath: string;
  showResults: boolean;
  onDiffPathChange: (path: string) => void;
  onChanged: () => void;
  onMerged: (result: BranchMergeResult) => void;
}) {
  const dirty = dirtyFileCount(worktree.changes);
  const branch: GitBranchInfo | undefined = git.branches.find(candidate => candidate.name === worktree.branch);
  const canReviewBranch = worktree.ahead > 0 && !!branch;
  const active = canReviewBranch ? view ?? (dirty > 0 ? 'work' : 'branch') : 'work';
  // The status view holds the lease while it is mounted; the committed-changes
  // view has none of its own.
  const checkoutPath = worktree.primary ? undefined : worktree.path;
  const focus = useGitFocus({ project: project.name, worktree: checkoutPath, enabled: active === 'branch' });

  return (
    <div className="flex h-full min-h-0 flex-col">
      {canReviewBranch && (
        <div className="shrink-0 border-b border-border px-4 py-1.5">
          <SegmentedControl<WorktreeView>
            aria-label="Worktree changes"
            size="sm"
            value={active}
            onChange={onViewChange}
            options={[
              { id: 'work', label: 'Working changes' },
              { id: 'branch', label: `Committed changes (${worktree.ahead} ahead)` },
            ]}
          />
        </div>
      )}
      {focus.error && <div role="alert" className="shrink-0 border-b border-border px-4 py-1.5 text-xs text-amber-700 dark:text-amber-400">Git tracker focus failed: {focus.error}</div>}
      <div className="min-h-0 flex-1 overflow-auto">
        {active === 'branch' && branch ? (
          <ProjectBranchChanges projectName={project.name} base={git.base} branch={branch} baseCheckedOut={git.baseCheckedOut} dirtyCount={dirty} worktreePath={worktree.path} onMerged={onMerged} />
        ) : (
          <ProjectStatusView key={`${project.name}\0${checkoutPath ?? ''}`} project={project} worktree={checkoutPath} diffPath={diffPath} showResults={showResults} onDiffPathChange={onDiffPathChange} onChanged={onChanged} />
        )}
      </div>
    </div>
  );
}

function Gone({ message, onBack }: { message: string; onBack: () => void }) {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 p-6 text-sm text-muted-foreground">
      <span role="alert">{message}</span>
      <Button type="button" variant="outline" size="sm" onClick={onBack}>Back to main checkout</Button>
    </div>
  );
}

function LandedBanner({ landed: { branch, result }, onDismiss }: { landed: Landed; onDismiss: () => void }) {
  const cleanup = [result.worktreeRemoved && 'worktree removed', result.branchDeleted && 'branch deleted'].filter(Boolean).join(', ');
  return (
    <div role="status" className="flex shrink-0 items-center gap-2 border-b border-border bg-green-500/10 px-4 py-1.5 text-xs text-green-700 dark:text-green-300">
      <UiCheck className="shrink-0" />
      <span className="min-w-0 flex-1">
        Merged {branch} into {result.targetBranch}{result.landedSha && ` as ${shortSha(result.landedSha)}`} ({result.commits} {result.commits === 1 ? 'commit' : 'commits'}, {result.mode}){cleanup && `; ${cleanup}`}.
      </span>
      <Button type="button" variant="ghost" size="sm" onClick={onDismiss}>Dismiss</Button>
    </div>
  );
}
