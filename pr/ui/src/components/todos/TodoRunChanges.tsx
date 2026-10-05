import type { TodoRunCommit, TodoRunLanding, TodoRunWorktree } from '../../types';
import { CommitRangeChanges, type CommitRangeOption, type CommitRangeRequest } from '../CommitRangeChanges';
import { todoQuery } from './format';
import { commitSubject, shortSha } from './runWorkspace';
import { TodoRunLandingSummary } from './TodoRunLandingSummary';
import { todoQueryKeys } from './todoQueries';

const ALL_CHANGES = 'all';

interface ChangeRange {
  hash: string;
  base: string;
}

function changeParams(dir: string, { hash, base }: ChangeRange) {
  const params = new URLSearchParams(todoQuery(dir));
  params.set('hash', hash);
  if (base) params.set('base', base);
  return params;
}

function rangeLabel({ hash, base }: ChangeRange) {
  return base ? `${shortSha(base)}..${shortSha(hash)}` : shortSha(hash);
}

function UnreachableNotice({ landing }: { landing?: TodoRunLanding }) {
  return (
    <div role="status" className="flex flex-col gap-1.5 px-3 py-3 text-xs text-muted-foreground">
      <span className="font-medium text-foreground">Commits no longer reachable</span>
      <span>The run's branch was removed after landing and its commits have been garbage collected, so their changes can no longer be shown.</span>
      {landing && <TodoRunLandingSummary landing={landing} />}
    </div>
  );
}

export interface TodoRunChangesProps {
  dir: string;
  worktree?: TodoRunWorktree;
  commits: TodoRunCommit[];
  landing?: TodoRunLanding;
}

/**
 * The files a todo run changed, as a tree beside a diff viewer. A picker
 * chooses the scope: every change between the run's setup and head, or one of
 * its commits. The viewer itself is the shared CommitRangeChanges; this only
 * says which ranges a run has and where the todo commit endpoints live.
 */
export function TodoRunChanges({ dir, worktree, commits, landing }: TodoRunChangesProps) {
  const combined: ChangeRange | null = worktree?.setup && worktree.head ? { base: worktree.setup, hash: worktree.head } : null;
  const ranges = new Map<string, ChangeRange>();
  if (combined) ranges.set(ALL_CHANGES, combined);
  for (const commit of commits) ranges.set(commit.sha, { hash: commit.sha, base: '' });
  if (ranges.size === 0) return null;

  const options: CommitRangeOption[] = [...ranges.keys()].map(id => ({
    id,
    label: id === ALL_CHANGES
      ? 'All changes'
      : `${shortSha(id)} ${commitSubject(commits.find(commit => commit.sha === id)?.message)}`.trim(),
  }));
  const rangeOf = (id: string) => {
    const range = ranges.get(id);
    if (!range) throw new Error(`Unknown change range ${id}`);
    return range;
  };
  const filesRequest = (id: string): CommitRangeRequest => {
    const range = rangeOf(id);
    return {
      queryKey: todoQueryKeys.commitFiles(dir, range.hash, range.base),
      url: `/api/todos/commits/files?${changeParams(dir, range).toString()}`,
      context: `Failed to load files for ${rangeLabel(range)}`,
    };
  };
  const diffRequest = (id: string, file: string): CommitRangeRequest => {
    const range = rangeOf(id);
    const params = changeParams(dir, range);
    params.set('file', file);
    return {
      queryKey: todoQueryKeys.commitDiff(dir, range.hash, range.base, file),
      url: `/api/todos/commits/diff?${params.toString()}`,
      context: `Failed to load ${file} diff for ${rangeLabel(range)}`,
    };
  };

  return (
    <CommitRangeChanges
      options={options}
      filesRequest={filesRequest}
      diffRequest={diffRequest}
      summary={`${commits.length} ${commits.length === 1 ? 'commit' : 'commits'}`}
      unreachable={<UnreachableNotice landing={landing} />}
    />
  );
}
