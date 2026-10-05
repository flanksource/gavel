import { UiGitBranch } from '@flanksource/clicky-ui/icons';
import { queryKeys } from '../query';
import type { BranchMergeResult, GitBranchInfo } from '../types';
import { BranchLandActions } from './BranchLandActions';
import { CommitRangeChanges, type CommitRangeRequest } from './CommitRangeChanges';
import { projectApiUrl } from './projectApiUrl';

const ALL_CHANGES = 'all';

export interface ProjectBranchChangesProps {
  projectName: string;
  base: string;
  branch: GitBranchInfo;
  baseCheckedOut: boolean;
  /** Uncommitted files in the branch's worktree; 0 when it has none. */
  dirtyCount: number;
  worktreePath?: string;
  onMerged: (result: BranchMergeResult) => void;
}

/**
 * A project branch's committed changes against its base — the files it adds,
 * edits and deletes beside a diff viewer — with the actions that land it
 * (merge into the base, or open a PR).
 */
export function ProjectBranchChanges({ projectName, base, branch, baseCheckedOut, dirtyCount, worktreePath, onMerged }: ProjectBranchChangesProps) {
  const filesRequest = (): CommitRangeRequest => ({
    queryKey: queryKeys.projectBranchFiles(projectName, branch.name, branch.head),
    url: projectApiUrl({ projectName, resource: 'branch/files', query: [['branch', branch.name]] }),
    context: `Failed to load files of ${branch.name}`,
  });
  const diffRequest = (_id: string, file: string): CommitRangeRequest => ({
    queryKey: queryKeys.projectBranchDiff(projectName, branch.name, branch.head, file),
    url: projectApiUrl({ projectName, resource: 'branch/diff', query: [['branch', branch.name], ['file', file]] }),
    context: `Failed to load ${file} diff of ${branch.name}`,
  });
  const { commits, files, adds, dels } = branch.diff;

  return (
    <div className="flex min-h-0 flex-col">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border px-4 py-3">
        <div className="flex min-w-0 flex-col gap-0.5">
          <span className="flex min-w-0 items-center gap-1 text-sm font-semibold">
            <UiGitBranch className="shrink-0 text-muted-foreground" />
            <span className="truncate" title={branch.name}>{branch.name}</span>
          </span>
          {branch.behind > 0 && <span className="text-[11px] text-amber-600">{branch.behind} behind {base}</span>}
        </div>
        <BranchLandActions
          projectName={projectName}
          branch={branch.name}
          base={base}
          ahead={branch.ahead}
          dirtyCount={dirtyCount}
          baseCheckedOut={baseCheckedOut}
          worktreePath={worktreePath}
          onMerged={onMerged}
        />
      </div>
      <CommitRangeChanges
        options={[{ id: ALL_CHANGES, label: `All changes vs ${base}` }]}
        showPicker={false}
        summary={`${commits} ${commits === 1 ? 'commit' : 'commits'} · ${files} ${files === 1 ? 'file' : 'files'} · +${adds} −${dels}`}
        filesRequest={filesRequest}
        diffRequest={diffRequest}
        unreachable={<div role="status" className="px-3 py-3 text-xs text-muted-foreground">Branch no longer exists</div>}
      />
    </div>
  );
}
