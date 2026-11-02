import { UiDiff, UiFolderGit, UiGitBranch, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import type { ProjectGitSummaryView } from '../projectGitQueries';
import type { ProjectGitSummary } from '../types';

export type ProjectGitBadgesState =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'ready'; summary: ProjectGitSummary };

// projectGitBadgesState picks one project's badge state out of the shared
// summary query. A project absent from a loaded response is new since the
// response was computed, so it reads as loading rather than as zero work.
export function projectGitBadgesState(summaries: ProjectGitSummaryView, projectName: string): ProjectGitBadgesState {
  const summary = summaries.byProject.get(projectName);
  if (summary) return { status: 'ready', summary };
  if (summaries.error) return { status: 'error', message: summaries.error };
  return { status: 'loading' };
}

const plural = (count: number, noun: string, many = `${noun}s`) => `${count} ${count === 1 ? noun : many}`;

// ProjectGitBadges summarises a project's work: lines added and removed by the
// main checkout's uncommitted changes, linked worktrees, unmerged branches and
// branches with uncommitted files.
// Loading renders nothing, a failure renders a warning with the reason, and a
// project with no work renders only the counts that are non-zero.
export function ProjectGitBadges({ state }: { state: ProjectGitBadgesState }) {
  if (state.status === 'loading') return null;
  const message = state.status === 'error' ? state.message : state.summary.error;
  if (message) {
    return (
      <span role="img" aria-label="Git summary unavailable" title={message} className="inline-flex shrink-0 items-center text-amber-600 dark:text-amber-400">
        <UiWarningTriangle className="text-xs" />
      </span>
    );
  }
  if (state.status !== 'ready') return null;
  const { adds, dels, worktrees, branches, uncommittedBranches, base } = state.summary;
  if (adds === 0 && dels === 0 && worktrees === 0 && branches === 0 && uncommittedBranches === 0) return null;
  return (
    <span className="inline-flex shrink-0 items-center gap-1.5 whitespace-nowrap text-[10px] tabular-nums">
      {adds > 0 && <span className="font-medium text-green-600 dark:text-green-400" title={`${plural(adds, 'line')} added by uncommitted changes in the checkout`}>+{adds}</span>}
      {dels > 0 && <span className="font-medium text-red-600 dark:text-red-400" title={`${plural(dels, 'line')} removed by uncommitted changes in the checkout`}>−{dels}</span>}
      {worktrees > 0 && (
        <span className="inline-flex items-center gap-0.5 text-muted-foreground" aria-label={plural(worktrees, 'linked worktree')} title={plural(worktrees, 'linked worktree')}>
          <UiFolderGit />{worktrees}
        </span>
      )}
      {branches > 0 && (
        <span
          className="inline-flex items-center gap-0.5 text-muted-foreground"
          aria-label={plural(branches, 'unmerged branch', 'unmerged branches')}
          title={`${plural(branches, 'branch', 'branches')} with commits not in ${base}`}
        >
          <UiGitBranch />{branches}
        </span>
      )}
      {uncommittedBranches > 0 && (
        <span
          className="inline-flex items-center gap-0.5 rounded bg-amber-500/15 px-1 font-medium text-amber-600 dark:text-amber-400"
          aria-label={`${plural(uncommittedBranches, 'branch', 'branches')} with uncommitted changes`}
          title={`${plural(uncommittedBranches, 'branch', 'branches')} with uncommitted changes`}
        >
          <UiDiff />{uncommittedBranches}
        </span>
      )}
    </span>
  );
}
