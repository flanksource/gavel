import { useMemo, useState, type ReactNode } from 'react';
import { useQuery, type QueryKey, type UseQueryResult } from '@tanstack/react-query';
import { Select, SplitPane } from '@flanksource/clicky-ui/components';
import { GitDiffPanel, Tree, type GitDiffPayload } from '@flanksource/clicky-ui/data';
import { UiDiff } from '@flanksource/clicky-ui/icons';
import type { TodoCommitFile, TodoCommitFilesResponse } from '../types';
import { HttpError, fetchJSON } from '../query';
import { Spinner } from '../icons/Spinner';
import { ProjectFileIcon } from '../icons/ProjectFileIcon';
import { buildFileTree, filesBelow, findFileTreeNode, firstFilePath, sortedChildren, type FileTreeNode } from './fileTree';

type ChangeNode = FileTreeNode<TodoCommitFile>;

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

/** One GET the viewer makes: its cache key, URL, and the error context. */
export interface CommitRangeRequest {
  queryKey: QueryKey;
  url: string;
  context: string;
}

/** A scope of changes the picker offers, e.g. "All changes" or a single commit. */
export interface CommitRangeOption {
  id: string;
  label: string;
}

export interface CommitRangeChangesProps {
  /** Picker entries; the first is shown until another is picked. */
  options: CommitRangeOption[];
  /** The file-list request of one option; the response is `{files: TodoCommitFile[]}`. */
  filesRequest: (id: string) => CommitRangeRequest;
  /** The diff request of one file (or directory) path in one option; the response is a GitDiffPayload. */
  diffRequest: (id: string, file: string) => CommitRangeRequest;
  /** Rendered beside the picker, e.g. a commit count. */
  summary?: ReactNode;
  /** Rendered in place of the changes when either request answers HTTP 410. */
  unreachable?: ReactNode;
  /** Whether to render the picker; a single-option viewer may drop it. */
  showPicker?: boolean;
}

/**
 * The files a range of commits changed, as a tree beside a diff viewer, for any
 * backend that answers a file list and a per-path diff. The caller owns which
 * ranges exist and the URL of each request, so the same viewer serves a todo
 * run's commits and a project branch. Selecting a file or directory in the tree
 * diffs that path; the selection survives a picker change while the path still
 * exists.
 */
export function CommitRangeChanges({ options, filesRequest, diffRequest, summary, unreachable, showPicker = true }: CommitRangeChangesProps) {
  const [picked, setPicked] = useState('');
  const [selectedPath, setSelectedPath] = useState('');
  const activeId = options.some(option => option.id === picked) ? picked : options[0]?.id;

  const filesQuery = useQuery({
    queryKey: activeId === undefined ? ['commit-range', 'none'] : filesRequest(activeId).queryKey,
    queryFn: async ({ signal }) => {
      if (activeId === undefined) throw new Error('No range to load files for');
      const request = filesRequest(activeId);
      const data = await fetchJSON<TodoCommitFilesResponse>({ url: request.url, signal, context: request.context });
      if (!Array.isArray(data.files)) throw new Error(`${request.context}: response has no files list`);
      return data.files;
    },
    enabled: activeId !== undefined,
    staleTime: Infinity,
  });
  const roots = useMemo(() => buildFileTree(filesQuery.data ?? []), [filesQuery.data]);
  const activePath = findFileTreeNode(roots, selectedPath)?.path ?? firstFilePath(roots);
  const diffQuery = useQuery({
    queryKey: activeId === undefined ? ['commit-range', 'none'] : diffRequest(activeId, activePath).queryKey,
    queryFn: ({ signal }) => {
      if (activeId === undefined) throw new Error('No range to load a diff for');
      const request = diffRequest(activeId, activePath);
      return fetchJSON<GitDiffPayload>({ url: request.url, signal, context: request.context });
    },
    enabled: activeId !== undefined && activePath !== '',
    staleTime: Infinity,
  });

  if (activeId === undefined) return null;

  const gone = unreachable !== undefined && (isGone(filesQuery.error) || isGone(diffQuery.error));
  return (
    <div className="flex flex-col">
      {(showPicker || summary) && (
        <div className="flex items-center gap-2 border-b border-border px-3 py-2">
          {showPicker && (
            <Select
              aria-label="Changes to show"
              value={activeId}
              onChange={event => setPicked(event.currentTarget.value)}
              options={options.map(option => ({ value: option.id, label: option.label }))}
              className="h-8 min-w-0 max-w-xl flex-1 text-xs"
            />
          )}
          {summary && <span className="shrink-0 text-[11px] tabular-nums text-muted-foreground">{summary}</span>}
        </div>
      )}
      {gone ? (
        unreachable
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
