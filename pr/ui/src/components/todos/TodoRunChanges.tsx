import { useMemo, useState, type ReactNode } from 'react';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { Select, SplitPane } from '@flanksource/clicky-ui/components';
import { GitDiffPanel, Tree, type GitDiffPayload } from '@flanksource/clicky-ui/data';
import { UiDiff } from '@flanksource/clicky-ui/icons';
import type { TodoCommitFile, TodoCommitFilesResponse, TodoRunCommit, TodoRunLanding, TodoRunWorktree } from '../../types';
import { HttpError, fetchJSON } from '../../query';
import { Spinner } from '../../icons/Spinner';
import { ProjectFileIcon } from '../../icons/ProjectFileIcon';
import { buildFileTree, filesBelow, findFileTreeNode, firstFilePath, sortedChildren, type FileTreeNode } from '../fileTree';
import { todoQuery } from './format';
import { commitSubject, shortSha } from './runWorkspace';
import { TodoRunLandingSummary } from './TodoRunLandingSummary';
import { todoQueryKeys } from './todoQueries';

const ALL_CHANGES = 'all';

type ChangeNode = FileTreeNode<TodoCommitFile>;

interface ChangeRange {
  hash: string;
  base: string;
}

// fileStatusView maps a file's change kind to its label and chip colors,
// matching how the rest of the dashboard signals add/edit/delete state.
function fileStatusView(status: TodoCommitFile['status']) {
  switch (status) {
    case 'added':
      return { label: 'Added', className: 'bg-green-500/10 text-green-600' };
    case 'deleted':
      return { label: 'Deleted', className: 'bg-red-500/10 text-red-600' };
    case 'renamed':
      return { label: 'Renamed', className: 'bg-blue-500/10 text-blue-600' };
    default:
      return { label: 'Modified', className: 'bg-amber-500/10 text-amber-600' };
  }
}

function isGone(error: unknown) {
  return error instanceof HttpError && error.status === 410;
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

function Chip({ children, className }: { children: ReactNode; className: string }) {
  return <span className={`inline-flex shrink-0 items-center rounded px-1 text-[10px] font-medium leading-tight ${className}`}>{children}</span>;
}

function Notice({ children }: { children: ReactNode }) {
  return <div className="px-3 py-3 text-xs text-muted-foreground">{children}</div>;
}

function ChangeRow({ node, directory }: { node: ChangeNode; directory: boolean }) {
  const file = node.file;
  const status = file ? fileStatusView(file.status) : null;
  const count = directory ? filesBelow(node).length : 0;
  return (
    <>
      <ProjectFileIcon
        path={node.path}
        directory={directory}
        className={`size-4 shrink-0 ${directory ? 'text-sky-600 dark:text-sky-400' : 'text-muted-foreground'}`}
      />
      <span className="flex min-w-0 flex-1 items-center gap-2">
        <span className={`truncate text-sm ${directory ? 'font-medium' : 'font-mono'}`} title={file?.previousPath ? `${file.previousPath} → ${node.path}` : node.path}>
          {node.name}
        </span>
        {status && <Chip className={status.className}>{status.label}</Chip>}
        {file?.language && <Chip className="bg-primary/10 text-primary">{file.language}</Chip>}
        {file?.scopes?.map(scope => <Chip key={scope} className="bg-muted text-muted-foreground">{scope}</Chip>)}
        {file && (
          <span className="flex shrink-0 gap-1 text-[11px]">
            {file.adds > 0 && <span className="text-green-600">+{file.adds}</span>}
            {file.dels > 0 && <span className="text-red-600">−{file.dels}</span>}
          </span>
        )}
        {directory && <span className="text-[10px] text-muted-foreground">{count} {count === 1 ? 'file' : 'files'}</span>}
      </span>
    </>
  );
}

function ChangeTree({ roots, selected, onSelect }: { roots: ChangeNode[]; selected: ChangeNode | null; onSelect: (path: string) => void }) {
  return (
    <Tree<ChangeNode>
      roots={roots}
      ariaLabel="Changed files"
      getChildren={sortedChildren}
      getKey={node => node.path}
      getAriaLabel={node => node.path}
      getSearchText={node => node.path}
      defaultOpen={() => true}
      selected={selected}
      revealSelected
      onSelect={node => onSelect(node.path)}
      renderRow={({ node, hasChildren }) => <ChangeRow node={node} directory={hasChildren} />}
      rowClass={(_node, isSelected) => `group min-h-8 border-l-2 py-1 pr-3 ${
        isSelected
          ? 'border-primary bg-primary/10'
          : 'border-transparent hover:border-primary/40 hover:bg-muted/40'
      }`}
      className="pb-1"
    />
  );
}

function ChangeDiff({ path, diff }: { path: string; diff: UseQueryResult<GitDiffPayload> }) {
  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex shrink-0 items-center gap-2 border-b border-border px-3 py-2">
        <UiDiff className="shrink-0 text-muted-foreground" />
        <h3 className="min-w-0 flex-1 truncate font-mono text-xs font-semibold" title={path}>{path}</h3>
      </div>
      <GitDiffPanel
        loading={diff.isPending}
        payload={diff.data ?? null}
        error={diff.error?.message ?? ''}
        className="min-h-0 flex-1 overflow-auto border-t-0"
        maxHeightClassName="max-h-none"
      />
    </div>
  );
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
 * its commits. Selecting a file or directory in the tree diffs that path.
 */
export function TodoRunChanges({ dir, worktree, commits, landing }: TodoRunChangesProps) {
  const combined: ChangeRange | null = worktree?.setup && worktree.head ? { base: worktree.setup, hash: worktree.head } : null;
  const ranges = new Map<string, ChangeRange>();
  if (combined) ranges.set(ALL_CHANGES, combined);
  for (const commit of commits) ranges.set(commit.sha, { hash: commit.sha, base: '' });
  const [picked, setPicked] = useState('');
  const [selectedPath, setSelectedPath] = useState('');
  const activeId = ranges.has(picked) ? picked : ranges.keys().next().value;
  const range = activeId === undefined ? undefined : ranges.get(activeId);

  const filesQuery = useQuery({
    queryKey: todoQueryKeys.commitFiles(dir, range?.hash ?? '', range?.base ?? ''),
    queryFn: async ({ signal }) => {
      if (!range) throw new Error('No range to load files for');
      const data = await fetchJSON<TodoCommitFilesResponse>({
        url: `/api/todos/commits/files?${changeParams(dir, range).toString()}`,
        signal,
        context: `Failed to load files for ${rangeLabel(range)}`,
      });
      if (!Array.isArray(data.files)) throw new Error(`Failed to load files for ${rangeLabel(range)}: response has no files list`);
      return data.files;
    },
    enabled: !!range,
    staleTime: Infinity,
  });
  const roots = useMemo(() => buildFileTree(filesQuery.data ?? []), [filesQuery.data]);
  const activePath = findFileTreeNode(roots, selectedPath)?.path ?? firstFilePath(roots);
  const diffQuery = useQuery({
    queryKey: todoQueryKeys.commitDiff(dir, range?.hash ?? '', range?.base ?? '', activePath),
    queryFn: ({ signal }) => {
      if (!range) throw new Error('No range to load a diff for');
      const params = changeParams(dir, range);
      params.set('file', activePath);
      return fetchJSON<GitDiffPayload>({
        url: `/api/todos/commits/diff?${params.toString()}`,
        signal,
        context: `Failed to load ${activePath} diff for ${rangeLabel(range)}`,
      });
    },
    enabled: !!range && activePath !== '',
    staleTime: Infinity,
  });

  if (!range || activeId === undefined) return null;

  const gone = isGone(filesQuery.error) || isGone(diffQuery.error);
  const options = [...ranges.keys()].map(id => ({
    value: id,
    label: id === ALL_CHANGES
      ? 'All changes'
      : `${shortSha(id)} ${commitSubject(commits.find(commit => commit.sha === id)?.message)}`.trim(),
  }));

  return (
    <div className="flex flex-col">
      <div className="flex items-center gap-2 border-b border-border px-3 py-2">
        <Select
          aria-label="Changes to show"
          value={activeId}
          onChange={event => setPicked(event.currentTarget.value)}
          options={options}
          className="h-8 min-w-0 max-w-xl flex-1 text-xs"
        />
        <span className="shrink-0 text-[11px] tabular-nums text-muted-foreground">
          {commits.length} {commits.length === 1 ? 'commit' : 'commits'}
        </span>
      </div>
      {gone ? (
        <UnreachableNotice landing={landing} />
      ) : filesQuery.isPending ? (
        <div className="flex items-center gap-2 px-3 py-3 text-xs text-muted-foreground">
          <Spinner className="text-xs" />
          Loading files…
        </div>
      ) : filesQuery.error ? (
        <div className="whitespace-pre-wrap px-3 py-3 text-xs text-red-600">{filesQuery.error.message}</div>
      ) : roots.length === 0 ? (
        <Notice>No file changes</Notice>
      ) : (
        <SplitPane
          className="h-[32rem]"
          defaultSplit={45}
          minLeft={25}
          minRight={30}
          leftClass="overflow-auto"
          rightClass="overflow-hidden"
          left={<ChangeTree roots={roots} selected={findFileTreeNode(roots, activePath)} onSelect={setSelectedPath} />}
          right={<ChangeDiff path={activePath} diff={diffQuery} />}
        />
      )}
    </div>
  );
}
