import { useEffect, useMemo, useRef } from 'react';
import { Tree } from '@flanksource/clicky-ui/data';
import { UiEyeClosed } from '@flanksource/clicky-ui/icons';
import type { ProjectFileStatus, FileState } from './ProjectStatusView';
import { ProjectFileIcon } from '../icons/ProjectFileIcon';

interface ProjectFileTreeProps {
  files: ProjectFileStatus[];
  selected: Set<string>;
  // locked maps a path already claimed by a queued commit group to that group's
  // position, so it cannot be selected into a second group.
  locked?: Map<string, number>;
  disabled: boolean;
  showResults: boolean;
  diffPath?: string;
  onToggleFile: (path: string) => void;
  onToggleFiles: (paths: string[]) => void;
  onIgnore: (path: string, directory: boolean) => void;
  onDiffPathChange?: (path: string) => void;
}

const noLockedFiles = new Map<string, number>();

export function ProjectFileTree({
  files,
  selected,
  locked = noLockedFiles,
  disabled,
  showResults,
  diffPath = '',
  onToggleFile,
  onToggleFiles,
  onIgnore,
  onDiffPathChange,
}: ProjectFileTreeProps) {
  const roots = useMemo(() => buildFileTree(files), [files]);
  const diffNode = useMemo(() => findFileTreeNode(roots, diffPath), [diffPath, roots]);
  return (
    <Tree<ProjectFileTreeNode>
      roots={roots}
      ariaLabel="Project files"
      getChildren={node => [...node.children.values()].sort(compareNodes)}
      getKey={node => node.path}
      getAriaLabel={node => node.path}
      getSearchText={node => node.path}
      defaultOpen={() => true}
      selected={diffNode}
      revealSelected
      onSelect={node => onDiffPathChange?.(node.path)}
      renderRow={({ node, hasChildren }) => (
        <ProjectFileTreeRow
          node={node}
          directory={hasChildren}
          selected={selected}
          locked={locked}
          disabled={disabled}
          showResults={showResults}
          onToggleFile={onToggleFile}
          onToggleFiles={onToggleFiles}
          onIgnore={onIgnore}
        />
      )}
      rowClass={(_node, isSelected) => `group min-h-8 border-l-2 py-1 pr-3 ${
        isSelected
          ? 'border-primary bg-primary/10'
          : 'border-transparent hover:border-primary/40 hover:bg-muted/40'
      }`}
      className="pb-1"
    />
  );
}

interface ProjectFileTreeNode {
  name: string;
  path: string;
  children: Map<string, ProjectFileTreeNode>;
  file?: ProjectFileStatus;
}

function ProjectFileTreeRow({
  node,
  directory,
  selected,
  locked,
  disabled,
  showResults,
  onToggleFile,
  onToggleFiles,
  onIgnore,
}: {
  node: ProjectFileTreeNode;
  directory: boolean;
  selected: Set<string>;
  locked: Map<string, number>;
  disabled: boolean;
  showResults: boolean;
  onToggleFile: (path: string) => void;
  onToggleFiles: (paths: string[]) => void;
  onIgnore: (path: string, directory: boolean) => void;
}) {
  const descendants = filesBelow(node);
  const selectable = descendants.filter(file => file.state !== 'conflict' && !locked.has(file.path));
  const selectedCount = selectable.filter(file => selected.has(file.path)).length;
  const ignorable = descendants.length > 0 && descendants.every(file => file.state === 'untracked');
  const file = node.file;
  const lockedGroup = file ? locked.get(file.path) : undefined;
  const inputDisabled = disabled || selectable.length === 0;
  return (
    <>
      <SelectionCheckbox
        label={directory ? `Select directory ${node.path}` : `Select ${node.path}`}
        checked={selectable.length > 0 && selectedCount === selectable.length}
        partial={selectedCount > 0 && selectedCount < selectable.length}
        disabled={inputDisabled}
        onChange={() => directory ? onToggleFiles(selectable.map(item => item.path)) : onToggleFile(node.path)}
      />
      <ProjectFileIcon
        path={node.path}
        directory={directory}
        className={`size-4 shrink-0 ${directory ? 'text-sky-600 dark:text-sky-400' : 'text-muted-foreground'}`}
      />
      <span className={`min-w-0 flex-1 ${lockedGroup ? 'opacity-60' : ''}`}>
        <span className="flex min-w-0 items-center gap-2" data-file-primary>
          <span className={`truncate text-sm ${directory ? 'font-medium' : 'font-mono'}`} title={node.path}>{node.name}</span>
          {file && <StateBadge state={file.state} />}
          {lockedGroup && (
            <span className="shrink-0 rounded bg-primary/10 px-1.5 py-0.5 text-[10px] font-medium text-primary">
              queued in group {lockedGroup}
            </span>
          )}
          {file && <DiffSummary file={file} />}
          {directory && <span className="text-[10px] text-muted-foreground">{descendants.length} files</span>}
        </span>
        {file && showResults && <FileDiagnostics file={file} />}
      </span>
      {ignorable && <IgnoreButton node={node} directory={directory} disabled={disabled} onIgnore={onIgnore} />}
    </>
  );
}

function IgnoreButton({ node, directory, disabled, onIgnore }: {
  node: ProjectFileTreeNode;
  directory: boolean;
  disabled: boolean;
  onIgnore: (path: string, directory: boolean) => void;
}) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={event => {
        event.stopPropagation();
        onIgnore(node.path, directory);
      }}
      className="inline-flex h-7 shrink-0 items-center gap-1 rounded px-2 text-[11px] text-muted-foreground hover:bg-accent hover:text-accent-foreground disabled:opacity-50"
      aria-label={`Ignore ${directory ? 'directory' : 'file'} ${node.path}`}
      title={`Add ${node.path}${directory ? '/' : ''} to .gitignore`}
    >
      <UiEyeClosed className="size-3.5" />Ignore
    </button>
  );
}

function SelectionCheckbox({ label, checked, partial, disabled, onChange }: {
  label: string;
  checked: boolean;
  partial: boolean;
  disabled: boolean;
  onChange: () => void;
}) {
  const ref = useRef<HTMLInputElement>(null);
  useEffect(() => {
    if (ref.current) ref.current.indeterminate = partial;
  }, [partial]);
  return (
    <input
      ref={ref}
      type="checkbox"
      aria-label={label}
      checked={checked}
      disabled={disabled}
      onClick={event => event.stopPropagation()}
      onChange={onChange}
      className="shrink-0"
    />
  );
}

function DiffSummary({ file }: { file: ProjectFileStatus }) {
  return (
    <span className="flex shrink-0 gap-1 text-[11px]">
      <span className="text-green-600">+{file.adds}</span>
      <span className="text-red-600">−{file.dels}</span>
    </span>
  );
}

function FileDiagnostics({ file }: { file: ProjectFileStatus }) {
  const lintCount = file.lintStatus.errors + file.lintStatus.warnings;
  const testCount = file.testStatus.passed + file.testStatus.failed + file.testStatus.skipped;
  if (testCount === 0 && lintCount === 0 && !file.resultsStale) return null;
  return (
    <span className="mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-muted-foreground">
      {testCount > 0 && (
        <span className={file.testStatus.failed > 0 ? 'text-red-600' : 'text-green-600'}>
          {file.testStatus.passed} passed{file.testStatus.failed > 0 ? `, ${file.testStatus.failed} failed` : ''}
        </span>
      )}
      {lintCount > 0 && <span className={file.lintStatus.errors > 0 ? 'text-red-600' : 'text-amber-600'}>{lintCount} lint</span>}
      {file.resultsStale && <span className="text-amber-600">stale results</span>}
    </span>
  );
}

function StateBadge({ state }: { state: FileState }) {
  const tones: Record<FileState, string> = {
    staged: 'bg-green-500/10 text-green-600',
    unstaged: 'bg-amber-500/10 text-amber-600',
    both: 'bg-blue-500/10 text-blue-600',
    untracked: 'bg-purple-500/10 text-purple-600',
    conflict: 'bg-red-500/10 text-red-600',
  };
  return <span className={`rounded px-1.5 py-0.5 text-[10px] font-medium ${tones[state]}`}>{state}</span>;
}

function buildFileTree(files: ProjectFileStatus[]) {
  const roots = new Map<string, ProjectFileTreeNode>();
  for (const file of files) {
    const parts = file.path.split('/').filter(Boolean);
    let nodes = roots;
    let currentPath = '';
    parts.forEach((part, index) => {
      currentPath = currentPath ? `${currentPath}/${part}` : part;
      const node: ProjectFileTreeNode = nodes.get(part) ?? { name: part, path: currentPath, children: new Map() };
      nodes.set(part, node);
      if (index === parts.length - 1) node.file = file;
      nodes = node.children;
    });
  }
  return [...roots.values()].sort(compareNodes);
}

function compareNodes(left: ProjectFileTreeNode, right: ProjectFileTreeNode) {
  const leftDirectory = left.children.size > 0;
  const rightDirectory = right.children.size > 0;
  if (leftDirectory !== rightDirectory) return leftDirectory ? -1 : 1;
  return left.name.localeCompare(right.name);
}

function filesBelow(node: ProjectFileTreeNode): ProjectFileStatus[] {
  if (node.file) return [node.file];
  return [...node.children.values()].flatMap(filesBelow);
}

function findFileTreeNode(nodes: ProjectFileTreeNode[], path: string): ProjectFileTreeNode | null {
  if (!path) return null;
  for (const node of nodes) {
    if (node.path === path) return node;
    const child = findFileTreeNode([...node.children.values()], path);
    if (child) return child;
  }
  return null;
}
