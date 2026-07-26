import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Button, SplitButton, SplitPane } from '@flanksource/clicky-ui/components';
import {
  UiBeaker,
  UiCheck,
  UiCog,
  UiGitBranch,
  UiGitCommit,
  UiGitPr,
  UiRefresh,
  UiWarningTriangle,
} from '@flanksource/clicky-ui/icons';
import type { Project } from '../types';
import { Spinner } from '../icons/Spinner';
import { CommitQueuePanel, commitQueueLockedFiles, type CommitQueueAction, type CommitQueueStatus } from './CommitQueuePanel';
import { ProjectActionDialog, type ProjectAction } from './ProjectActionDialog';
import { type ProjectActionStatus } from './ProjectActionFeedback';
import { ProjectActionRunDialog, type ProjectRunnerAction } from './ProjectActionRunDialog';
import { ProjectDiffView } from './ProjectDiffView';
import { ProjectFileTree } from './ProjectFileTree';

export type FileState = 'staged' | 'unstaged' | 'both' | 'untracked' | 'conflict';

export interface ProjectFileStatus {
  path: string;
  previousPath?: string;
  state: FileState;
  stagedKind?: string;
  workKind?: string;
  adds: number;
  dels: number;
  modifiedAt?: string;
  language?: string;
  scopes?: string[];
  testStatus: { passed: number; failed: number; skipped: number };
  lintStatus: { errors: number; warnings: number; infos: number };
  problems?: { kind: string; severity: string; label: string; line?: number; message?: string }[];
  resultsStale: boolean;
  repomapError?: string;
  conflictReason?: string;
}

interface ProjectStatusResponse {
  project: Project;
  workDir: string;
  branch: string;
  files: ProjectFileStatus[];
  resultsStale: boolean;
  action: ProjectActionStatus;
  commitQueue: CommitQueueStatus;
}

interface Props {
  project: Project;
  diffPath?: string;
  onDiffPathChange?: (path: string) => void;
  onChanged?: () => void;
}

const ignoreDiffPathChange = () => {};
const ignoreChanged = () => {};

export function ProjectStatusView({ project, diffPath = '', onDiffPathChange = ignoreDiffPathChange, onChanged = ignoreChanged }: Props) {
  const [status, setStatus] = useState<ProjectStatusResponse | null>(null);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [advancedAction, setAdvancedAction] = useState<ProjectAction | null>(null);
  const [activeRun, setActiveRun] = useState<{ action: ProjectRunnerAction; runId: string } | null>(projectActionRunFromLocation);
  const [diffRevision, setDiffRevision] = useState(0);
  const actionWasRunning = useRef(false);

  const load = useCallback(async (refreshDiff = false) => {
    try {
      const response = await fetch(`/api/projects/${encodeURIComponent(project.name)}/status`);
      const payload = await response.json();
      if (!response.ok) throw new Error(payload.error || 'Failed to load project status');
      const next = payload as ProjectStatusResponse;
      setStatus(next);
      const locked = commitQueueLockedFiles(next.commitQueue);
      const paths = new Set(next.files.filter(file => file.state !== 'conflict' && !locked.has(file.path)).map(file => file.path));
      setSelected(current => new Set([...current].filter(path => paths.has(path))));
      setError('');
      if (refreshDiff) setDiffRevision(current => current + 1);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Failed to load project status');
    } finally {
      setLoading(false);
    }
  }, [project.name]);

  useEffect(() => {
    setStatus(null);
    setSelected(new Set());
    setActiveRun(projectActionRunFromLocation());
    setLoading(true);
    void load();
  }, [load]);

  const actionRunning = status?.action.running ?? false;
  const commitQueueRunning = status?.commitQueue?.running ?? false;

  useEffect(() => {
    if (!actionRunning && !commitQueueRunning) {
      if (actionWasRunning.current) {
        actionWasRunning.current = false;
        setDiffRevision(current => current + 1);
      }
      return;
    }
    actionWasRunning.current = true;
    const timer = window.setInterval(() => { void load(); }, 750);
    return () => window.clearInterval(timer);
  }, [actionRunning, commitQueueRunning, load]);

  useEffect(() => {
    if (!diffPath || !status) return;
    const present = status.files.some(file => file.path === diffPath || file.path.startsWith(`${diffPath}/`));
    if (!present) onDiffPathChange('');
  }, [diffPath, onDiffPathChange, status]);

  const lockedFiles = useMemo(() => commitQueueLockedFiles(status?.commitQueue), [status?.commitQueue]);
  const selectable = useMemo(
    () => status?.files.filter(file => file.state !== 'conflict' && !lockedFiles.has(file.path)) ?? [],
    [lockedFiles, status],
  );
  const allSelected = selectable.length > 0 && selectable.every(file => selected.has(file.path));
  const selectedFiles = useMemo(() => [...selected].sort(), [selected]);

  // Commits go to the per-project queue rather than the one-shot action so a
  // second selection can be handed over while the first is still committing.
  const queueCommit = useCallback(async (action: CommitQueueAction, options?: Record<string, unknown>) => {
    const response = await fetch(`/api/projects/${encodeURIComponent(project.name)}/commit-queue`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(options ? { action, options } : { action, files: selectedFiles }),
    });
    const payload = await response.json();
    if (!response.ok) throw new Error(payload.error || `Failed to queue ${action === 'open-pr' ? 'Open PR' : 'commit'}`);
    setStatus(current => current ? { ...current, commitQueue: payload as CommitQueueStatus } : current);
    setSelected(new Set());
    setError('');
  }, [project.name, selectedFiles]);

  const cancelCommitGroup = useCallback(async (id: string) => {
    try {
      const response = await fetch(`/api/projects/${encodeURIComponent(project.name)}/commit-queue/${encodeURIComponent(id)}`, { method: 'DELETE' });
      const payload = await response.json();
      if (!response.ok) throw new Error(payload.error || 'Failed to cancel commit group');
      setStatus(current => current ? { ...current, commitQueue: payload as CommitQueueStatus } : current);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Failed to cancel commit group');
    }
  }, [project.name]);

  const startAction = useCallback(async (action: ProjectAction, options?: Record<string, unknown>) => {
    if (action === 'commit') return queueCommit('commit', options);
    const body = options
      ? { action, options }
      : action === 'lint' && selectedFiles.length > 0
        ? { action, files: selectedFiles }
        : { action };
    const response = await fetch(`/api/projects/${encodeURIComponent(project.name)}/actions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    const payload = await response.json();
    if (!response.ok) throw new Error(payload.error || `Failed to start ${action}`);
    const actionStatus = payload as ProjectActionStatus;
    if (!actionStatus.runId) throw new Error(`${action} did not return a test-runner id`);
    setActiveRun({ action, runId: actionStatus.runId });
    setStatus(current => current ? { ...current, action: actionStatus } : current);
    setError('');
  }, [project.name, queueCommit, selectedFiles]);

  const closeActiveRun = useCallback(() => {
    setActiveRun(null);
    const url = new URL(window.location.href);
    url.searchParams.delete('action');
    url.searchParams.delete('run');
    window.history.replaceState({}, '', `${url.pathname}${url.search}`);
  }, []);

  const run = useCallback(async (action: ProjectAction) => {
    try {
      await startAction(action);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : `Failed to start ${action}`);
    }
  }, [startAction]);

  const queueSelected = useCallback(async (action: CommitQueueAction) => {
    try {
      await queueCommit(action);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : `Failed to queue ${action}`);
    }
  }, [queueCommit]);

  const ignore = useCallback(async (path: string, directory: boolean) => {
    try {
      const response = await fetch(`/api/projects/${encodeURIComponent(project.name)}/ignore`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path, directory }),
      });
      const payload = await response.json();
      if (!response.ok) throw new Error(payload.error || `Failed to ignore ${path}`);
      setSelected(current => new Set([...current].filter(selectedPath => directory ? !selectedPath.startsWith(`${path}/`) : selectedPath !== path)));
      await load(true);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : `Failed to ignore ${path}`);
    }
  }, [load, project.name]);

  if (loading && !status) return <Centered><Spinner /> Loading project status…</Centered>;
  if (!status) return <Centered>{error || 'Project status is unavailable.'}</Centered>;

  const busy = status.action.running;
  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="shrink-0 border-b border-border px-4 py-3">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <h2 className="truncate text-base font-semibold text-foreground">{project.name}</h2>
            <div className="mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
              <span className="flex items-center gap-1"><UiGitBranch />{status.branch || '(detached)'}</span>
              <span className="truncate" title={status.workDir}>{status.workDir}</span>
              <span>{status.files.length} change{status.files.length === 1 ? '' : 's'}</span>
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-1.5">
            <SplitButton
              variant="outline"
              size="sm"
              disabled={busy}
              label={<span className="flex items-center gap-1"><UiWarningTriangle />{selected.size > 0 ? 'Lint selected' : 'Lint project'}</span>}
              title="Lint options"
              onClick={() => void run('lint')}
              items={[{
                label: <span className="flex items-center gap-2"><UiCog />Advanced…</span>,
                onSelect: () => setAdvancedAction('lint'),
              }]}
            />
            <SplitButton
              variant="outline"
              size="sm"
              disabled={busy}
              label={<span className="flex items-center gap-1"><UiBeaker />Test changed</span>}
              title="Test options"
              onClick={() => void run('test')}
              items={[{
                label: <span className="flex items-center gap-2"><UiCog />Advanced…</span>,
                onSelect: () => setAdvancedAction('test'),
              }]}
            />
            <SplitButton
              variant="default"
              size="sm"
              disabled={selected.size === 0}
              label={<span className="flex items-center gap-1"><UiGitCommit />Commit selected ({selected.size})</span>}
              title="Commit options"
              onClick={() => void run('commit')}
              items={[
                {
                  label: <span className="flex items-center gap-2"><UiGitPr />Open PR</span>,
                  onSelect: () => void queueSelected('open-pr'),
                },
                {
                  label: <span className="flex items-center gap-2"><UiCog />Advanced…</span>,
                  onSelect: () => setAdvancedAction('commit'),
                },
              ]}
            />
            <Button type="button" variant="ghost" size="icon" disabled={busy} onClick={() => void load(true)} aria-label="Refresh project status">
              <UiRefresh />
            </Button>
          </div>
        </div>
        {error && <div role="alert" className="mt-2 text-xs text-red-600 dark:text-red-400">{error}</div>}
        {status.resultsStale && <div className="mt-2 text-xs text-amber-600">Test or lint results are from an earlier commit.</div>}
      </div>

      <CommitQueuePanel queue={status.commitQueue} onCancel={id => void cancelCommitGroup(id)} />

      <SplitPane
        className="min-h-0 flex-1"
        defaultSplit={55}
        minLeft={35}
        minRight={30}
        leftClass="overflow-hidden"
        left={
          <div className="flex h-full min-h-0 flex-col">
            <div className="min-h-0 flex-1 overflow-auto">
              {status.files.length === 0 ? (
                <Centered><UiCheck className="mr-1 text-green-600" />Working tree clean</Centered>
              ) : (
                <div className="divide-y divide-border">
                  <label className="sticky top-0 z-10 flex items-center gap-2 bg-muted px-4 py-2 text-xs font-medium text-muted-foreground">
                    <input
                      type="checkbox"
                      aria-label="Select all files"
                      checked={allSelected}
                      disabled={busy || selectable.length === 0}
                      onChange={() => setSelected(allSelected ? new Set() : new Set(selectable.map(file => file.path)))}
                    />
                    Select all committable files
                  </label>
                  <ProjectFileTree
                    files={status.files}
                    selected={selected}
                    locked={lockedFiles}
                    disabled={busy}
                    diffPath={diffPath}
                    onDiffPathChange={onDiffPathChange}
                    onToggleFile={path => setSelected(current => togglePath(current, path))}
                    onToggleFiles={paths => setSelected(current => togglePaths(current, paths))}
                    onIgnore={(path, directory) => void ignore(path, directory)}
                  />
                </div>
              )}
            </div>
          </div>
        }
        right={<ProjectDiffView projectName={project.name} path={diffPath} refreshKey={diffRevision} />}
      />

      <ProjectActionDialog
        projectName={project.name}
        action={advancedAction}
        selectedFiles={selectedFiles}
        onClose={() => setAdvancedAction(null)}
        onRun={(action, options) => startAction(action, options)}
      />
      {activeRun && (
        <ProjectActionRunDialog
          projectName={project.name}
          projectDir={status.workDir}
          action={activeRun.action}
          runId={activeRun.runId}
          onClose={closeActiveRun}
          onTodoCreated={onChanged}
        />
      )}
    </div>
  );
}

function projectActionRunFromLocation(): { action: ProjectRunnerAction; runId: string } | null {
  if (typeof window === 'undefined') return null;
  const params = new URLSearchParams(window.location.search);
  const action = params.get('action');
  const runId = params.get('run');
  if ((action !== 'test' && action !== 'lint') || !runId) return null;
  return { action, runId };
}

function togglePath(current: Set<string>, path: string) {
  const next = new Set(current);
  if (next.has(path)) next.delete(path);
  else next.add(path);
  return next;
}

function togglePaths(current: Set<string>, paths: string[]) {
  const next = new Set(current);
  const remove = paths.every(path => next.has(path));
  paths.forEach(path => remove ? next.delete(path) : next.add(path));
  return next;
}

function Centered({ children }: { children: React.ReactNode }) {
  return <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground">{children}</div>;
}
