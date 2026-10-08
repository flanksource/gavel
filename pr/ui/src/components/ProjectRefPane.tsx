import { useState, type ReactNode } from 'react';
import { Button, Combobox, SegmentedControl, type ComboboxOption } from '@flanksource/clicky-ui/components';
import { UiCheck, UiEdit, UiFolderGit, UiGitBranch, UiGitCommit, UiHome, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import { useProjectGit } from '../projectGitQueries';
import { parseProjectRef } from '../projectRef';
import type { BranchMergeResult, GitBranchInfo, GitWorktree, Project, ProjectGit } from '../types';
import { useGitFocus } from '../useGitFocus';
import { useNow } from '../useNow';
import { ProjectBranchChanges } from './ProjectBranchChanges';
import { ProjectStatusView } from './ProjectStatusView';
import { dirtyFileCount, gitStateAge, projectRefAges, projectRefEntries, type ProjectRefEntry } from './projectGitView';

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
 * working-tree changes; a worktree that is ahead of the base, and a branch,
 * also show their committed changes with the actions that land them.
 */
export function ProjectRefPane({ project, projectRef, diffPath, showResults, onRefChange, onDiffPathChange, onChanged }: Props) {
  const gitQuery = useProjectGit(project.name);
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
  if (ref.kind === 'main') {
    body = <ProjectStatusView key={`${project.name}\0`} project={project} diffPath={diffPath} showResults={showResults} onDiffPathChange={onDiffPathChange} onChanged={onChanged} />;
  } else if (!git) {
    body = gitQuery.error
      ? <Gone message={gitQuery.error.message} onBack={() => onRefChange('')} />
      : <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground">Loading git state…</div>;
  } else if (ref.kind === 'worktree') {
    const worktree = git.worktrees.find(candidate => !candidate.primary && candidate.path === ref.path);
    const view = viewChoice?.ref === projectRef ? viewChoice.view : undefined;
    body = worktree
      ? <WorktreeBody project={project} git={git} worktree={worktree} view={view} onViewChange={next => setViewChoice({ ref: projectRef, view: next })} diffPath={diffPath} showResults={showResults} onDiffPathChange={onDiffPathChange} onChanged={onChanged} onMerged={onMerged(worktree.branch)} />
      : <Gone message={`Worktree ${ref.path} no longer exists.`} onBack={() => onRefChange('')} />;
  } else {
    const branch = git.branches.find(candidate => candidate.name === ref.name && candidate.worktree === '');
    body = branch
      ? <div className="h-full min-h-0 overflow-auto"><ProjectBranchChanges projectName={project.name} base={git.base} branch={branch} baseCheckedOut={git.baseCheckedOut} dirtyCount={0} onMerged={onMerged(branch.name)} /></div>
      : <Gone message={`Branch ${ref.name} no longer has commits to review.`} onBack={() => onRefChange('')} />;
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      <RefBar git={git} loading={gitQuery.isPending} error={gitQuery.error?.message ?? ''} notices={notices} value={projectRef} onChange={onRefChange} />
      {landed && <LandedBanner landed={landed} onDismiss={() => setLanded(null)} />}
      <div className="min-h-0 flex-1">{body}</div>
    </div>
  );
}

function RefBar({ git, loading, error, notices, value, onChange }: { git: ProjectGit | undefined; loading: boolean; error: string; notices: string[]; value: string; onChange: (ref: string) => void }) {
  if (error) {
    return (
      <div role="alert" className="flex shrink-0 items-center gap-1 border-b border-border px-4 py-1.5 text-xs text-red-600 dark:text-red-400">
        <UiWarningTriangle />Worktrees and branches unavailable: {error}
      </div>
    );
  }
  if (loading) return <div className="shrink-0 border-b border-border px-4 py-1.5 text-xs text-muted-foreground">Loading git state…</div>;
  if (!git) return null;
  const entries = projectRefEntries(git);
  if (entries.length === 1) {
    return notices.length > 0
      ? <div className="flex shrink-0 items-center gap-3 border-b border-border px-4 py-1.5"><GitNotices notices={notices} /></div>
      : null;
  }
  const selected = entries.find(entry => entry.value === value);
  const options: ComboboxOption[] = entries.map(entry => ({
    value: optionValue(entry.value),
    label: entry.label,
    group: entry.group,
    title: entry.title,
    icon: GROUP_ICONS[entry.group],
    selectedLabel: `${GROUP_NOUNS[entry.group]} ${entry.label}`,
    trailing: <RefSummary entry={entry} />,
  }));
  return (
    <div className="flex shrink-0 items-center gap-3 border-b border-border px-4 py-1.5">
      {/* The open menu takes the control's width, so the width here is what
          keeps every option's stats and ages on a single line. */}
      <Combobox
        ariaLabel="Worktree or branch"
        value={optionValue(value)}
        onChange={next => onChange(next === MAIN_OPTION ? '' : next)}
        options={selected ? options : [...options, { value, label: 'Unavailable', disabled: true }]}
        allowCustomValue={false}
        required
        placeholder="Worktree or branch"
        className="w-[42rem] min-w-0 max-w-full shrink"
      />
      {selected && <RefSummary entry={selected} />}
      <GitNotices notices={notices} />
      <GitStateAge computedAt={git.computedAt} />
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

function GitStateAge({ computedAt }: { computedAt: string }) {
  useNow();
  return (
    <span className="ml-auto shrink-0 whitespace-nowrap text-xs text-muted-foreground" title={`Git state computed ${new Date(computedAt).toLocaleString()}`}>
      {gitStateAge(computedAt)}
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

const GROUP_NOUNS: Record<ProjectRefEntry['group'], string> = { Checkout: 'Checkout', Worktrees: 'Worktree', Branches: 'Branch' };

const plural = (n: number, noun: string) => `${n} ${noun}${n === 1 ? '' : 's'}`;

/** One line: commits · uncommitted · +adds −dels, then the commit and touched ages. */
function RefSummary({ entry: { stats, lastCommitAt, touchedAt } }: { entry: ProjectRefEntry }) {
  useNow();
  const { committed, touched } = projectRefAges({ lastCommitAt, touchedAt });
  return (
    <span className="flex shrink-0 items-center gap-2 whitespace-nowrap text-xs tabular-nums text-muted-foreground">
      {stats.commits > 0 && <span>{plural(stats.commits, 'commit')}</span>}
      {stats.uncommitted > 0 && <span>{stats.uncommitted} uncommitted</span>}
      {stats.lines && (
        <span className="flex items-center gap-1 font-medium">
          <span className="text-green-600 dark:text-green-400">+{stats.lines.adds}</span>
          <span className="text-red-600 dark:text-red-400">−{stats.lines.dels}</span>
        </span>
      )}
      {committed && <span className="flex items-center gap-0.5" title={`Last commit ${new Date(lastCommitAt).toLocaleString()}`}><UiGitCommit />{committed}</span>}
      {touched && touchedAt && <span className="flex items-center gap-0.5" title={`Uncommitted files last touched ${new Date(touchedAt).toLocaleString()}`}><UiEdit />{touched}</span>}
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
  const focus = useGitFocus({ project: project.name, worktree: worktree.path, enabled: active === 'branch' });

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
          <ProjectStatusView key={`${project.name}\0${worktree.path}`} project={project} worktree={worktree.path} diffPath={diffPath} showResults={showResults} onDiffPathChange={onDiffPathChange} onChanged={onChanged} />
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
        Merged {branch} into {result.targetBranch} as {shortSha(result.landedSha)} ({result.commits} {result.commits === 1 ? 'commit' : 'commits'}, {result.mode}){cleanup && `; ${cleanup}`}.
      </span>
      <Button type="button" variant="ghost" size="sm" onClick={onDismiss}>Dismiss</Button>
    </div>
  );
}
