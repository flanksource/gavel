import { useQuery } from '@tanstack/react-query';
import type { Project, TodoRunWorktree } from '../../types';
import { fetchJSON, queryKeys } from '../../query';
import { useProjectGit } from '../../projectGitQueries';
import { useGitFocus } from '../../useGitFocus';

export function TodoWorktreeStatus({ dir, worktree }: { dir: string; worktree: TodoRunWorktree }) {
  return (
    <div className="flex flex-col gap-1.5 border-b border-border px-3 py-2 text-xs text-muted-foreground">
      {worktree.path && <span className="break-all font-mono">{worktree.path}</span>}
      {worktree.removed
        ? <span>Worktree removed</span>
        : worktree.path ? <LiveWorktreeStatus dir={dir} path={worktree.path} /> : <span>Worktree path was not recorded</span>}
    </div>
  );
}

function LiveWorktreeStatus({ dir, path }: { dir: string; path: string }) {
  const projects = useQuery({
    queryKey: queryKeys.projects(),
    queryFn: ({ signal }) => fetchJSON<Project[]>({ url: '/api/projects', signal, context: 'Load workspace projects' }),
    staleTime: Infinity,
  });
  const project = projects.data?.find(candidate => candidate.dir.replace(/\/+$/, '') === dir.replace(/\/+$/, ''));
  const git = useProjectGit(project?.name ?? '', { enabled: !!project });
  const focus = useGitFocus({ project: project?.name ?? '', worktree: path, enabled: !!project });
  const worktree = git.data?.worktrees.find(candidate => candidate.path === path);
  const error = projects.error?.message || git.error?.message || git.data?.error || worktree?.statusError || focus.error;
  if (error) return <span role="alert">{error}</span>;
  if (projects.isPending || (project && git.isPending)) return <span>Loading worktree status…</span>;
  if (!project) return <span role="alert">No configured project for {dir}</span>;
  if (!worktree) return <span>Worktree unavailable</span>;
  if (!worktree.statusScannedAt) return <span>Waiting for Git status…</span>;
  const changes = worktree.changes;
  const staged = changes.staged + changes.both;
  const unstaged = changes.unstaged + changes.both;
  if (staged + unstaged + changes.untracked + changes.conflict === 0) return <span>Working tree clean</span>;
  return (
    <span className="flex flex-wrap gap-x-3 gap-y-1 tabular-nums">
      <span>{staged} staged</span>
      <span>{unstaged} unstaged</span>
      <span>{changes.untracked} untracked</span>
      {changes.conflict > 0 && <span className="text-red-600">{changes.conflict} conflicts</span>}
      <span className="text-green-600">+{changes.adds}</span>
      <span className="text-red-600">−{changes.dels}</span>
    </span>
  );
}
