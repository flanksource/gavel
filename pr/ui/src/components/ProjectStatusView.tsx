import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
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
import { fetchJSON, mutationJSON, queryKeys } from '../query';
import { useDocumentVisible } from '../useDocumentVisible';
import { ProjectActionDialog, type ProjectAction } from './ProjectActionDialog';
import { type ProjectActionStatus } from './ProjectActionFeedback';
import { ProjectActionRunDialog, type ProjectRunnerAction } from './ProjectActionRunDialog';
import { ProjectCommitTasks } from './ProjectCommitTasks';
import { ProjectDiffView } from './ProjectDiffView';
import { ProjectFileTree } from './ProjectFileTree';
import { projectDiffQueryKey } from './projectMutations';

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
}

interface ProjectCommitRun {
  runId: string;
}

type ProjectCommitAction = 'commit' | 'open-pr';

interface Props {
  project: Project;
  diffPath?: string;
  showResults?: boolean;
  onDiffPathChange?: (path: string) => void;
  onChanged?: () => void;
}

const ignoreDiffPathChange = () => {};
const ignoreChanged = () => {};
const STATUS_POLL_INTERVAL_MS = 750;

export function ProjectStatusView({ project, diffPath = '', showResults = false, onDiffPathChange = ignoreDiffPathChange, onChanged = ignoreChanged }: Props) {
  const queryClient = useQueryClient();
  const visible = useDocumentVisible();
  const statusQuery = useQuery({
    queryKey: queryKeys.projectStatus(project.name, showResults),
    queryFn: ({ signal }) => fetchJSON<ProjectStatusResponse>({
      url: `/api/projects/${encodeURIComponent(project.name)}/status${showResults ? '?includeResults=true' : ''}`,
      signal,
      context: `Failed to load project status for ${project.name}`,
    }),
    enabled: visible,
    staleTime: STATUS_POLL_INTERVAL_MS,
    refetchInterval: query => visible && query.state.data?.action.running ? STATUS_POLL_INTERVAL_MS : false,
    refetchIntervalInBackground: false,
  });
  const status = statusQuery.data ?? null;
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [error, setError] = useState('');
  const [advancedAction, setAdvancedAction] = useState<ProjectAction | null>(null);
  const [activeRun, setActiveRun] = useState<{ action: ProjectRunnerAction; runId: string } | null>(projectActionRunFromLocation);
  const [commitRunId, setCommitRunId] = useState('');
  const [lockedFiles, setLockedFiles] = useState<Map<string, number>>(new Map());
  const [commitTaskError, setCommitTaskError] = useState('');
  const actionWasRunning = useRef(false);

  const refreshProjectData = useCallback(async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.projectStatusScope(project.name) }),
      queryClient.invalidateQueries({ queryKey: projectDiffQueryKey(project.name) }),
    ]);
  }, [project.name, queryClient]);

  useEffect(() => {
    setSelected(new Set());
    setActiveRun(projectActionRunFromLocation());
    setCommitRunId('');
    setLockedFiles(new Map());
    setCommitTaskError('');
    setError('');
    actionWasRunning.current = false;
  }, [project.name]);

  const actionRunning = status?.action.running ?? false;

  useEffect(() => {
    if (!actionRunning) {
      if (actionWasRunning.current) {
        actionWasRunning.current = false;
        void queryClient.invalidateQueries({ queryKey: projectDiffQueryKey(project.name) });
      }
      return;
    }
    actionWasRunning.current = true;
  }, [actionRunning, project.name, queryClient]);

  useEffect(() => {
    if (!status) return;
    const available = new Set(status.files
      .filter(file => file.state !== 'conflict' && !commitTaskError && !lockedFiles.has(file.path))
      .map(file => file.path));
    setSelected(current => new Set([...current].filter(path => available.has(path))));
  }, [commitTaskError, lockedFiles, status]);

  useEffect(() => {
    if (!diffPath || !status) return;
    const present = status.files.some(file => file.path === diffPath || file.path.startsWith(`${diffPath}/`));
    if (!present) onDiffPathChange('');
  }, [diffPath, onDiffPathChange, status]);

  const selectable = useMemo(
    () => commitTaskError ? [] : status?.files.filter(file => file.state !== 'conflict' && !lockedFiles.has(file.path)) ?? [],
    [commitTaskError, lockedFiles, status],
  );
  const allSelected = selectable.length > 0 && selectable.every(file => selected.has(file.path));
  const selectedFiles = useMemo(() => [...selected].sort(), [selected]);

  const queueCommitMutation = useMutation({
    mutationFn: async ({ action, options, files }: {
      action: ProjectCommitAction;
      options?: Record<string, unknown>;
      files: string[];
    }) => {
      const label = action === 'open-pr' ? 'open PR' : 'commit';
      const result = await mutationJSON<ProjectCommitRun>({
        url: `/api/projects/${encodeURIComponent(project.name)}/commit-queue`,
        method: 'POST',
        body: options ? { action, options } : { action, files },
        context: `Failed to queue ${label} for ${project.name}`,
      });
      if (typeof result.runId !== 'string' || result.runId === '') {
        throw new Error(`Failed to queue ${label} for ${project.name}: response did not include a run id`);
      }
      return result;
    },
    onSuccess: result => {
      setCommitRunId(result.runId);
      setSelected(new Set());
      setError('');
    },
  });

  const actionMutation = useMutation({
    mutationFn: async ({ action, body }: { action: ProjectRunnerAction; body: Record<string, unknown> }) => {
      const result = await mutationJSON<ProjectActionStatus>({
        url: `/api/projects/${encodeURIComponent(project.name)}/actions`,
        method: 'POST',
        body,
        context: `Failed to start ${action} for ${project.name}`,
      });
      if (!result.runId) {
        throw new Error(`Failed to start ${action} for ${project.name}: response did not include a test-runner id`);
      }
      return { actionStatus: result, runId: result.runId };
    },
    onSuccess: async ({ actionStatus, runId }, { action }) => {
      setActiveRun({ action, runId });
      queryClient.setQueryData<ProjectStatusResponse>(queryKeys.projectStatus(project.name, showResults), current => (
        current ? { ...current, action: actionStatus } : current
      ));
      setError('');
      if (!actionStatus.running) await refreshProjectData();
    },
  });

  const ignoreMutation = useMutation({
    mutationFn: ({ path, directory }: { path: string; directory: boolean }) => mutationJSON({
      url: `/api/projects/${encodeURIComponent(project.name)}/ignore`,
      method: 'POST',
      body: { path, directory },
      context: `Failed to ignore ${path} in ${project.name}`,
    }),
    onSuccess: async (_, { path, directory }) => {
      setSelected(current => new Set([...current].filter(selectedPath => directory ? !selectedPath.startsWith(`${path}/`) : selectedPath !== path)));
      setError('');
      await refreshProjectData();
    },
  });

  // Commits go to the per-project queue rather than the one-shot action so a
  // second selection can be handed over while the first is still committing.
  const queueCommit = useCallback(async (action: ProjectCommitAction, options?: Record<string, unknown>) => {
    await queueCommitMutation.mutateAsync({ action, options, files: selectedFiles });
  }, [queueCommitMutation.mutateAsync, selectedFiles]);

  const startAction = useCallback(async (action: ProjectAction, options?: Record<string, unknown>) => {
    if (action === 'commit') return queueCommit('commit', options);
    const body = options
      ? { action, options }
      : action === 'lint' && selectedFiles.length > 0
        ? { action, files: selectedFiles }
        : { action };
    await actionMutation.mutateAsync({ action, body });
  }, [actionMutation.mutateAsync, queueCommit, selectedFiles]);

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

  const queueSelected = useCallback(async (action: ProjectCommitAction) => {
    try {
      await queueCommit(action);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : `Failed to queue ${action}`);
    }
  }, [queueCommit]);

  const refreshAfterCommit = useCallback(() => {
    void refreshProjectData().catch(cause => setError(cause instanceof Error ? cause.message : 'Failed to refresh project data'));
  }, [refreshProjectData]);

  const ignore = useCallback(async (path: string, directory: boolean) => {
    try {
      await ignoreMutation.mutateAsync({ path, directory });
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : `Failed to ignore ${path}`);
    }
  }, [ignoreMutation.mutateAsync]);

  const statusError = statusQuery.error instanceof Error ? statusQuery.error.message : '';
  if (statusQuery.isPending && !status) return <Centered><Spinner /> Loading project status…</Centered>;
  if (!status) return <Centered>{error || statusError || 'Project status is unavailable.'}</Centered>;

  const busy = status.action.running || actionMutation.isPending;
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
              disabled={selected.size === 0 || queueCommitMutation.isPending}
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
            <Button type="button" variant="ghost" size="icon" disabled={busy} onClick={refreshAfterCommit} aria-label="Refresh project status">
              <UiRefresh />
            </Button>
          </div>
        </div>
        {(error || statusError) && <div role="alert" className="mt-2 text-xs text-red-600 dark:text-red-400">{error || statusError}</div>}
        {showResults && status.resultsStale && <div className="mt-2 text-xs text-amber-600">Test or lint results are from an earlier commit.</div>}
      </div>

      <ProjectCommitTasks
        key={project.name}
        projectName={project.name}
        preferredRunId={commitRunId}
        onLockedFilesChange={setLockedFiles}
        onErrorChange={setCommitTaskError}
        onComplete={refreshAfterCommit}
      />

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
                      disabled={busy || ignoreMutation.isPending || commitTaskError !== '' || selectable.length === 0}
                      onChange={() => setSelected(allSelected ? new Set() : new Set(selectable.map(file => file.path)))}
                    />
                    Select all committable files
                  </label>
                  <ProjectFileTree
                    files={status.files}
                    selected={selected}
                    locked={lockedFiles}
                    disabled={busy || ignoreMutation.isPending || commitTaskError !== ''}
                    showResults={showResults}
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
        right={<ProjectDiffView projectName={project.name} path={diffPath} />}
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
