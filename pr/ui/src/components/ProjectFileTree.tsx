import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { ConfirmDialog, DropdownMenu, type DropdownMenuItem } from '@flanksource/clicky-ui/components';
import { Tree } from '@flanksource/clicky-ui/data';
import { UiCopy, UiDotsVertical, UiEdit, UiEyeClosed, UiFileCode, UiTrash } from '@flanksource/clicky-ui/icons';
import type { ProjectFileStatus, FileState } from './ProjectStatusView';
import { ProjectFileIcon } from '../icons/ProjectFileIcon';
import { ProjectLinkedDepsBadge } from './ProjectLinkedDepsBadge';
import { buildFileTree, filesBelow, findFileTreeNode, sortedChildren, type FileTreeNode } from './fileTree';
import {
  CONFLICT_REASON,
  stageBlockReason,
  stageRequestPaths,
  stageStateOf,
  type IgnoreRequest,
  type ProjectEditor,
  type StageState,
} from './projectStage';

interface ProjectFileTreeProps {
  // files are already filtered by the parent; search is echoed into the tree's
  // own search box so there is a single filter row.
  files: ProjectFileStatus[];
  search: string;
  onSearchChange: (search: string) => void;
  filterActions?: ReactNode;
  // locked maps a path already claimed by a queued commit group to that group's
  // position, so it cannot be staged into a second group.
  locked?: Map<string, number>;
  disabled: boolean;
  showResults: boolean;
  diffPath?: string;
  // editors are the editors detected on the gavel host; none hides Open in editor.
  editors: ProjectEditor[];
  // onMenuOpen fires when a row menu opens, so the host can load the editors lazily.
  onMenuOpen?: () => void;
  onStage: (paths: string[], unstage: boolean) => void;
  onDiscard: (path: string) => void;
  onIgnore: (request: IgnoreRequest) => void;
  onOpenInEditor: (path: string, editor: string) => void;
  onCopyPath: (path: string) => void;
  onDiffPathChange?: (path: string) => void;
}

const noLockedFiles = new Map<string, number>();
const TRACKED_IGNORE_REASON = '.gitignore only hides untracked paths — tracked changes stay';
const ROW_MENU_TRIGGER = 'inline-flex size-8 shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-accent-foreground md:size-6 md:[@media(hover:hover)]:opacity-0 group-hover:opacity-100 focus-visible:opacity-100 aria-expanded:opacity-100';

interface DiscardRequest {
  path: string;
  folder: boolean;
  untracked: boolean;
}

export function ProjectFileTree({
  files,
  search,
  onSearchChange,
  filterActions,
  locked = noLockedFiles,
  disabled,
  showResults,
  diffPath = '',
  editors,
  onMenuOpen,
  onStage,
  onDiscard,
  onIgnore,
  onOpenInEditor,
  onCopyPath,
  onDiffPathChange,
}: ProjectFileTreeProps) {
  const roots = useMemo(() => buildFileTree(files), [files]);
  const diffNode = useMemo(() => findFileTreeNode(roots, diffPath), [diffPath, roots]);
  const [discarding, setDiscarding] = useState<DiscardRequest | null>(null);
  return (
    <>
      <Tree<ProjectFileTreeNode>
        roots={roots}
        ariaLabel="Project files"
        getChildren={sortedChildren}
        getKey={node => node.path}
        getAriaLabel={node => node.path}
        getSearchText={node => [node.path, node.file?.previousPath].filter(Boolean).join(' ')}
        showSearch
        searchQuery={search}
        onSearchQueryChange={onSearchChange}
        actions={<>
          <StageAllCheckbox files={files} locked={locked} disabled={disabled} onStage={onStage} />
          {filterActions}
        </>}
        empty={<div className="px-3 py-4 text-sm text-muted-foreground">No files match the filter.</div>}
        defaultOpen={() => true}
        selected={diffNode}
        revealSelected
        onSelect={node => onDiffPathChange?.(node.path)}
        renderRow={({ node, hasChildren }) => (
          <ProjectFileTreeRow
            node={node}
            directory={hasChildren}
            locked={locked}
            disabled={disabled}
            showResults={showResults}
            editors={editors}
            onMenuOpen={onMenuOpen}
            onStage={onStage}
            onDiscard={setDiscarding}
            onIgnore={onIgnore}
            onOpenInEditor={onOpenInEditor}
            onCopyPath={onCopyPath}
          />
        )}
        indentPx={12}
        rowClass={(_node, isSelected) => `group min-h-9 border-l-2 py-0.5 pr-1 md:min-h-7 ${
          isSelected
            ? 'border-primary bg-primary/10'
            : 'border-transparent hover:border-primary/40 hover:bg-muted/40'
        }`}
        className="pb-1"
      />
      <ConfirmDialog
        open={discarding !== null}
        title={discarding ? `${discarding.untracked ? 'Delete' : 'Discard changes to'} ${discarding.path}${discarding.folder ? '/' : ''}?` : ''}
        warning="This cannot be undone."
        confirmLabel={discarding?.untracked ? 'Delete' : 'Discard changes'}
        variant="destructive"
        onCancel={() => setDiscarding(null)}
        onConfirm={() => {
          if (discarding) onDiscard(discarding.path);
          setDiscarding(null);
        }}
      />
    </>
  );
}

type ProjectFileTreeNode = FileTreeNode<ProjectFileStatus>;

// StageAllCheckbox leads the filter toolbar's action group. The Tree toolbar has
// no leading slot, so it sits first in `actions`, right of the search box.
function StageAllCheckbox({ files, locked, disabled, onStage }: {
  files: ProjectFileStatus[];
  locked: Map<string, number>;
  disabled: boolean;
  onStage: (paths: string[], unstage: boolean) => void;
}) {
  const state = stageStateOf(files);
  const unstage = state === 'all';
  const targets = stageRequestPaths({ files, locked, unstage });
  const label = unstage ? 'Unstage all' : 'Stage all';
  return (
    <label className="flex shrink-0 items-center gap-1 text-[11px] text-muted-foreground" title={`${label} (${targets.length} file${targets.length === 1 ? '' : 's'})`}>
      <StageCheckbox label={label} state={state} disabled={disabled || targets.length === 0} onChange={() => onStage(targets, unstage)} />
      <span className="max-md:sr-only">{label}</span>
    </label>
  );
}

function extensionOf(path: string): string {
  const name = path.slice(path.lastIndexOf('/') + 1);
  const dot = name.lastIndexOf('.');
  return dot > 0 ? name.slice(dot) : '';
}

// rowMenuItems builds the ⋮ menu: File actions, then Ignore rules.
function rowMenuItems({ node, folder, untracked, editors, onCopyPath, onOpenInEditor, onDiscard, onIgnore }: {
  node: ProjectFileTreeNode;
  folder: boolean;
  untracked: boolean;
  editors: ProjectEditor[];
  onCopyPath: (path: string) => void;
  onOpenInEditor: (path: string, editor: string) => void;
  onDiscard: (request: DiscardRequest) => void;
  onIgnore: (request: IgnoreRequest) => void;
}): DropdownMenuItem[] {
  const ignoreGate = untracked ? {} : { disabled: true, title: TRACKED_IGNORE_REASON };
  const extension = folder ? '' : extensionOf(node.path);
  return [
    { label: 'Copy path', icon: UiCopy, group: 'File', onSelect: () => onCopyPath(node.path) },
    ...editors.map(editor => ({ label: `Open in ${editor.label}`, icon: UiEdit, group: 'File', onSelect: () => onOpenInEditor(node.path, editor.id) })),
    {
      label: untracked ? 'Delete' : 'Discard changes',
      icon: UiTrash,
      group: 'File',
      onSelect: () => onDiscard({ path: node.path, folder, untracked }),
    },
    folder
      ? { label: `Ignore ${node.name}/`, icon: UiEyeClosed, group: 'Ignore', ...ignoreGate, onSelect: () => onIgnore({ path: node.path, directory: true }) }
      : { label: `Ignore ${node.name}`, icon: UiEyeClosed, group: 'Ignore', ...ignoreGate, onSelect: () => onIgnore({ path: node.path }) },
    ...(extension ? [{ label: `Ignore all *${extension} files`, icon: UiFileCode, group: 'Ignore', ...ignoreGate, onSelect: () => onIgnore({ path: node.path, extension }) }] : []),
  ];
}

function ProjectFileTreeRow({
  node,
  directory,
  locked,
  disabled,
  showResults,
  editors,
  onMenuOpen,
  onStage,
  onDiscard,
  onIgnore,
  onOpenInEditor,
  onCopyPath,
}: {
  node: ProjectFileTreeNode;
  directory: boolean;
  locked: Map<string, number>;
  disabled: boolean;
  showResults: boolean;
  editors: ProjectEditor[];
  onMenuOpen: (() => void) | undefined;
  onStage: (paths: string[], unstage: boolean) => void;
  onDiscard: (request: DiscardRequest) => void;
  onIgnore: (request: IgnoreRequest) => void;
  onOpenInEditor: (path: string, editor: string) => void;
  onCopyPath: (path: string) => void;
}) {
  const descendants = filesBelow(node);
  const state = stageStateOf(descendants);
  const unstage = state === 'all';
  const targets = stageRequestPaths({ files: descendants, locked, unstage });
  const untracked = descendants.length > 0 && descendants.every(file => file.state === 'untracked');
  const file = node.file;
  const lockedGroup = file ? locked.get(file.path) : undefined;
  const blockReason = targets.length > 0 ? null : file
    ? stageBlockReason({ file, lockedGroup })
    : descendants.some(item => item.state === 'conflict') ? CONFLICT_REASON : 'Nothing here can be staged';
  return (
    <>
      <StageCheckbox
        label={directory ? `Stage directory ${node.path}` : `Stage ${node.path}`}
        title={blockReason ?? undefined}
        state={state}
        disabled={disabled || targets.length === 0}
        onChange={() => onStage(targets, unstage)}
      />
      <ProjectFileIcon
        path={node.path}
        directory={directory}
        className={`size-4 shrink-0 ${directory ? 'text-sky-600 dark:text-sky-400' : 'text-muted-foreground'}`}
      />
      <span className={`min-w-0 flex-1 ${lockedGroup ? 'opacity-60' : ''}`}>
        <span className="flex min-w-0 items-center gap-1.5" data-file-primary>
          <span
            className={`min-w-0 flex-1 truncate text-xs ${directory ? 'font-medium' : 'font-mono'} ${file?.state === 'untracked' ? 'text-muted-foreground' : ''}`}
            title={file?.previousPath ? `${file.previousPath} → ${node.path}` : node.path}
          >
            {node.name}
          </span>
          {file && <ProjectLinkedDepsBadge file={file} />}
          {lockedGroup && (
            <span className="shrink-0 rounded bg-primary/10 px-1 text-[10px] font-medium text-primary">
              queued in group {lockedGroup}
            </span>
          )}
          {file && <StateToken state={file.state} />}
          {file && <DiffSummary file={file} />}
          {directory && <span className="shrink-0 text-[10px] tabular-nums text-muted-foreground" title={`${descendants.length} changed files`}>{descendants.length}</span>}
        </span>
        {file && showResults && <FileDiagnostics file={file} />}
      </span>
      {/* The menu portals through React, so a click inside it would otherwise bubble into row selection. */}
      <span className="shrink-0" onClick={event => event.stopPropagation()}>
        <DropdownMenu
          align="right"
          menuLabel={`Actions for ${node.path}`}
          onOpenChange={open => { if (open) onMenuOpen?.(); }}
          trigger={
            <button type="button" disabled={disabled} className={ROW_MENU_TRIGGER} aria-label={`Actions for ${node.path}`} title="Actions">
              <UiDotsVertical className="size-4" />
            </button>
          }
          items={rowMenuItems({ node, folder: directory, untracked, editors, onCopyPath, onOpenInEditor, onDiscard, onIgnore })}
        />
      </span>
    </>
  );
}

function StageCheckbox({ label, title, state, disabled, onChange }: {
  label: string;
  title?: string | undefined;
  state: StageState;
  disabled: boolean;
  onChange: () => void;
}) {
  const ref = useRef<HTMLInputElement>(null);
  const partial = state === 'partial';
  useEffect(() => {
    if (ref.current) ref.current.indeterminate = partial;
  }, [partial]);
  return (
    <input
      ref={ref}
      type="checkbox"
      aria-label={label}
      title={title}
      checked={state === 'all'}
      disabled={disabled}
      onClick={event => event.stopPropagation()}
      onChange={onChange}
      className="shrink-0"
    />
  );
}

function DiffSummary({ file }: { file: ProjectFileStatus }) {
  if (file.adds === 0 && file.dels === 0) return null;
  return (
    <span className="flex shrink-0 gap-1 text-[10px] tabular-nums">
      <span className="text-emerald-600 dark:text-emerald-400">+{file.adds}</span>
      <span className="text-rose-600 dark:text-rose-400">−{file.dels}</span>
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

const stateTokens: Record<FileState, { letter: string; label: string; tone: string }> = {
  staged: { letter: 'S', label: 'Staged', tone: 'text-emerald-600 dark:text-emerald-400' },
  unstaged: { letter: 'M', label: 'Modified, not staged', tone: 'text-amber-600 dark:text-amber-400' },
  both: { letter: 'M', label: 'Partly staged', tone: 'text-sky-600 dark:text-sky-400' },
  untracked: { letter: 'U', label: 'Untracked', tone: 'text-muted-foreground' },
  conflict: { letter: 'C', label: 'Conflict', tone: 'text-rose-600 dark:text-rose-400' },
};

function StateToken({ state }: { state: FileState }) {
  const token = stateTokens[state];
  return <span className={`w-3 shrink-0 text-center font-mono text-[10px] font-semibold ${token.tone}`} title={token.label}>{token.letter}</span>;
}
