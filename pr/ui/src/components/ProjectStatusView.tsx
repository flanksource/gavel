import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button, FilterBar, SplitButton, type FilterBarFilter } from '@flanksource/clicky-ui/components';
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
import { copyText } from '../clipboard';
import { fetchJSON, mutationJSON, queryKeys } from '../query';
import { useDocumentVisible } from '../useDocumentVisible';
import { useGitFocus } from '../useGitFocus';
import { useIsMobile } from '../useIsMobile';
import { ProjectActionDialog, type ProjectAction } from './ProjectActionDialog';
import { type ProjectActionStatus } from './ProjectActionFeedback';
import { ProjectActionRunDialog, type ProjectRunnerAction } from './ProjectActionRunDialog';
import { ProjectCommitTasks } from './ProjectCommitTasks';
import { ProjectDiffView } from './ProjectDiffView';
import { ProjectChangesView } from './ProjectChangesView';
import { ProjectFileTree } from './ProjectFileTree';
import { ProjectGitLockPrompt, useProjectGitLockRequest } from './ProjectGitLockPrompt';
import { useRebaseAction } from './rebaseAction';
import {
  emptyProjectFileFilters,
  filterProjectFiles,
  scopeCounts,
  UNSCOPED_SCOPE,
  type ProjectFileFilters,
} from './projectFileFilter';
import { projectApiUrl } from './projectApiUrl';
import { invalidateProjectData, projectDiffQueryKey, projectEditorsQueryKey } from './projectMutations';
import { stageRequestPaths, stagedCommitPaths, type IgnoreRequest, type ProjectEditor } from './projectStage';

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
  linkedDependencies?: { name: string; target: string }[];
}

export interface ProjectStatusResponse {
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

// An index or worktree mutation: each one changes what status and diffs report.
type ProjectIndexRequest =
  | { resource: 'stage'; body: { paths: string[]; unstage: boolean }; context: string }
  | { resource: 'discard'; body: { paths: string[] }; context: string }
  | { resource: 'ignore'; body: IgnoreRequest; context: string };

interface Props {
  project: Project;
  diffPath?: string;
  showResults?: boolean;
  // Absolute path of the linked worktree to review and commit in. Omitted for
  // the project's main checkout.
  worktree?: string;
  onDiffPathChange?: (path: string) => void;
  onChanged?: () => void;
}

const ignoreDiffPathChange = () => {};
const ignoreChanged = () => {};
const STATUS_POLL_INTERVAL_MS = 750;
const noEditors: ProjectEditor[] = [];
// The lint/test action endpoint is not worktree-scoped, so offering it here would
// silently run against the main checkout.
const RUNNER_MAIN_CHECKOUT_ONLY = 'Lint and test run in the main checkout only';

export function ProjectStatusView({ project, diffPath = '', showResults = false, worktree, onDiffPathChange = ignoreDiffPathChange, onChanged = ignoreChanged }: Props) {
  const queryClient = useQueryClient();
  const gitLock = useProjectGitLockRequest();
  const rebaseAction = useRebaseAction();
  const mobile = useIsMobile();
  const visible = useDocumentVisible();
  const focus = useGitFocus({ project: project.name, worktree });
  const statusQuery = useQuery({
    queryKey: queryKeys.projectStatus(project.name, showResults, worktree),
    queryFn: ({ signal }) => fetchJSON<ProjectStatusResponse>({
      url: projectApiUrl({ projectName: project.name, resource: 'status', query: showResults ? [['includeResults', 'true']] : [], worktree }),
      signal,
      context: `Failed to load project status for ${project.name}`,
    }),
    enabled: visible,
    staleTime: STATUS_POLL_INTERVAL_MS,
    refetchInterval: query => visible && query.state.data?.action.running ? STATUS_POLL_INTERVAL_MS : false,
    refetchIntervalInBackground: false,
  });
  const status = statusQuery.data ?? null;
  const [filters, setFilters] = useState<ProjectFileFilters>(emptyProjectFileFilters);
  const [error, setError] = useState('');
  const [advancedAction, setAdvancedAction] = useState<ProjectAction | null>(null);
  const [activeRun, setActiveRun] = useState<{ action: ProjectRunnerAction; runId: string } | null>(projectActionRunFromLocation);
  const [commitRunId, setCommitRunId] = useState('');
  const [lockedFiles, setLockedFiles] = useState<Map<string, number>>(new Map());
  const [commitTaskError, setCommitTaskError] = useState('');
  const actionWasRunning = useRef(false);

  // The editors are only asked for once a row menu opens, then cached for the session.
  const [editorsWanted, setEditorsWanted] = useState(false);
  const editorsQuery = useQuery({
    queryKey: projectEditorsQueryKey,
    queryFn: ({ signal }) => fetchJSON<ProjectEditor[]>({ url: '/api/editors', signal, context: 'Failed to load editors' }),
    enabled: editorsWanted,
    staleTime: Infinity,
  });
  const editors = editorsQuery.data ?? noEditors;

  const refreshProjectData = useCallback(
    () => invalidateProjectData(queryClient, project.name),
    [project.name, queryClient],
  );

  useEffect(() => {
    setFilters(emptyProjectFileFilters());
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
    if (!diffPath || !status) return;
    const present = status.files.some(file => file.path === diffPath || file.path.startsWith(`${diffPath}/`));
    if (!present) onDiffPathChange('');
  }, [diffPath, onDiffPathChange, status]);

  const visibleFiles = useMemo(() => filterProjectFiles(status?.files ?? [], filters), [filters, status]);
  // A commit and the lint/PR actions take the whole index, filter or not.
  const stagedFiles = useMemo(
    () => commitTaskError ? [] : stagedCommitPaths(status?.files ?? [], lockedFiles),
    [commitTaskError, lockedFiles, status],
  );
  const scopeFilter = useMemo<FilterBarFilter[]>(() => {
    const counts = scopeCounts(status?.files ?? []);
    const scoped = Object.keys(counts).some(scope => scope !== UNSCOPED_SCOPE);
    if (!scoped && Object.keys(filters.scopes).length === 0) return [];
    const options = Object.entries(counts)
      .sort(([aName, aCount], [bName, bCount]) => bCount - aCount || aName.localeCompare(bName))
      .map(([scope, count]) => ({ value: scope, label: `${scope} (${count})` }));
    // A selected scope no file carries any more keeps its option so it can be cleared.
    for (const scope of Object.keys(filters.scopes)) {
      if (!(scope in counts)) options.push({ value: scope, label: `${scope} (0)` });
    }
    return [{
      key: 'scope', kind: 'multi', label: 'Scope', options,
      value: filters.scopes, onChange: scopes => setFilters(current => ({ ...current, scopes })),
    }];
  }, [filters.scopes, status]);

  // react-doctor-disable-next-line react-doctor/query-mutation-missing-invalidation -- queueing only starts a background run; ProjectCommitTasks refreshes project data via onComplete when it finishes
  const queueCommitMutation = useMutation({
    mutationFn: async ({ action, options, files }: {
      action: ProjectCommitAction;
      options?: Record<string, unknown>;
      files: string[];
    }) => {
      const label = action === 'open-pr' ? 'open PR' : 'commit';
      const result = await mutationJSON<ProjectCommitRun>({
        ...gitLock.request(),
        url: `/api/projects/${encodeURIComponent(project.name)}/commit-queue`,
        method: 'POST',
        body: { ...(options ? { action, options } : { action, files }), ...(worktree ? { worktree } : {}) },
        context: `Failed to queue ${label} for ${project.name}`,
      });
      if (typeof result.runId !== 'string' || result.runId === '') {
        throw new Error(`Failed to queue ${label} for ${project.name}: response did not include a run id`);
      }
      return result;
    },
    onSuccess: result => {
      setCommitRunId(result.runId);
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
      queryClient.setQueryData<ProjectStatusResponse>(queryKeys.projectStatus(project.name, showResults, worktree), current => (
        current ? { ...current, action: actionStatus } : current
      ));
      setError('');
      if (!actionStatus.running) await refreshProjectData();
    },
  });

  // react-doctor-disable-next-line react-doctor/query-mutation-missing-invalidation -- onSuccess invalidates the project status and diff queries via refreshProjectData
  const indexMutation = useMutation({
    mutationFn: ({ resource, body, context }: ProjectIndexRequest) => mutationJSON({
      ...(resource === 'ignore' ? {} : gitLock.request()),
      url: projectApiUrl({ projectName: project.name, resource, worktree }),
      method: 'POST',
      body,
      context,
    }),
    onSuccess: async () => {
      setError('');
      await refreshProjectData();
    },
  });

  const openMutation = useMutation({
    mutationFn: ({ path, editor }: { path: string; editor: string }) => mutationJSON({
      url: projectApiUrl({ projectName: project.name, resource: 'open', worktree }),
      method: 'POST',
      body: { path, editor },
      context: `Failed to open ${path} in ${editor}`,
    }),
    onSuccess: () => setError(''),
  });

  // Commits go to the per-project queue rather than the one-shot action so a
  // second batch can be staged and handed over while the first is still committing.
  const queueCommit = useCallback(async (action: ProjectCommitAction, options?: Record<string, unknown>) => {
    await queueCommitMutation.mutateAsync({ action, options, files: stagedFiles });
  }, [queueCommitMutation.mutateAsync, stagedFiles]);

  const startAction = useCallback(async (action: ProjectAction, options?: Record<string, unknown>) => {
    if (action === 'commit') return queueCommit('commit', options);
    const body = options
      ? { action, options }
      : action === 'lint' && stagedFiles.length > 0
        ? { action, files: stagedFiles }
        : { action };
    await actionMutation.mutateAsync({ action, body });
  }, [actionMutation.mutateAsync, queueCommit, stagedFiles]);

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

  const queueStaged = useCallback(async (action: ProjectCommitAction) => {
    try {
      await queueCommit(action);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : `Failed to queue ${action}`);
    }
  }, [queueCommit]);

  const refreshAfterCommit = useCallback(() => {
    void refreshProjectData().catch(cause => setError(cause instanceof Error ? cause.message : 'Failed to refresh project data'));
  }, [refreshProjectData]);

  // Every row action reports its failure inline instead of swallowing it.
  const attempt = useCallback(async (task: () => Promise<unknown>) => {
    try {
      await task();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
    }
  }, []);

  const stage = useCallback((paths: string[], unstage: boolean) => attempt(() => indexMutation.mutateAsync({
    resource: 'stage',
    body: { paths, unstage },
    context: `Failed to ${unstage ? 'unstage' : 'stage'} ${paths.length === 1 ? paths[0] : `${paths.length} files`} in ${project.name}`,
  })), [attempt, indexMutation.mutateAsync, project.name]);

  const toggleStaged = useCallback((file: ProjectFileStatus) => {
    const unstage = file.state === 'staged';
    return stage(stageRequestPaths({ files: [file], locked: lockedFiles, unstage }), unstage);
  }, [lockedFiles, stage]);

  const discard = useCallback((path: string) => attempt(() => indexMutation.mutateAsync({
    resource: 'discard',
    body: { paths: [path] },
    context: `Failed to discard ${path} in ${project.name}`,
  })), [attempt, indexMutation.mutateAsync, project.name]);

  const ignore = useCallback((request: IgnoreRequest) => attempt(() => indexMutation.mutateAsync({
    resource: 'ignore',
    body: request,
    context: `Failed to ignore ${request.path} in ${project.name}`,
  })), [attempt, indexMutation.mutateAsync, project.name]);

  const openInEditor = useCallback((path: string, editor: string) => attempt(() => openMutation.mutateAsync({ path, editor })), [attempt, openMutation.mutateAsync]);

  const copyPath = useCallback((path: string) => attempt(() => copyText(path)), [attempt]);

  const statusError = statusQuery.error instanceof Error ? statusQuery.error.message : '';
  if (!status) return (
    <div className="flex h-full min-h-0 flex-col">
      {mobile && diffPath && <Button type="button" variant="ghost" className="min-h-11 self-start" onClick={() => onDiffPathChange('')}>Back to files</Button>}
      <Centered>{statusQuery.isPending ? <><Spinner /> Loading project status…</> : error || statusError || 'Project status is unavailable.'}</Centered>
    </div>
  );

  const busy = status.action.running || actionMutation.isPending;
  const treeDisabled = busy || indexMutation.isPending || commitTaskError !== '';
  const reviewing = mobile && diffPath !== '';
  const editorsError = editorsQuery.error instanceof Error ? editorsQuery.error.message : '';
  const notices = <>
    {(error || statusError || editorsError) && <div role="alert" className="mt-2 text-xs text-red-600 dark:text-red-400">{error || statusError || editorsError}</div>}
    {focus.error && <div role="alert" className="mt-2 text-xs text-amber-600 dark:text-amber-400">Git tracker focus failed: {focus.error}</div>}
    {showResults && status.resultsStale && <div className="mt-2 text-xs text-amber-600">Test or lint results are from an earlier commit.</div>}
  </>;
  const commitButton = <SplitButton
    variant="default"
    size={mobile ? 'default' : 'sm'}
    className={mobile ? '[&_button]:min-h-11 [&_button]:min-w-11' : undefined}
    disabled={queueCommitMutation.isPending || (!mobile && stagedFiles.length === 0)}
    primaryDisabled={stagedFiles.length === 0}
    label={<span className="flex items-center gap-1"><UiGitCommit />Commit staged ({stagedFiles.length})</span>}
    title="Commit options"
    onClick={() => void run('commit')}
    items={[
      ...(mobile ? [{
        label: stagedFiles.length > 0 ? `Commit & open PR (${stagedFiles.length})` : 'Open PR',
        disabled: commitTaskError !== '',
        onSelect: () => void queueStaged('open-pr'),
      }] : []),
      ...(mobile && rebaseAction ? [rebaseAction] : []),
      { label: <span className="flex items-center gap-2"><UiCog />Advanced…</span>, onSelect: () => setAdvancedAction('commit') },
    ]}
  />;
  return (
    <div className="flex h-full min-h-0 min-w-0 flex-col">
      {!reviewing && (!mobile || error || statusError || editorsError || focus.error || (showResults && status.resultsStale)) && <div className="shrink-0 border-b border-border px-3 py-2 md:px-4 md:py-3">
        <div className="flex flex-wrap items-start justify-between gap-3">
          {!mobile && <div className="min-w-0">
            <h2 className="truncate text-base font-semibold text-foreground">{project.name}</h2>
            <div className="mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
              <span className="flex items-center gap-1"><UiGitBranch />{status.branch || '(detached)'}</span>
              <span className="hidden truncate md:inline" title={status.workDir}>{status.workDir}</span>
              <span>{status.files.length} change{status.files.length === 1 ? '' : 's'}</span>
            </div>
          </div>}
          {!mobile && <div className="flex flex-wrap items-center gap-1.5">
            <SplitButton
              variant="outline"
              size="sm"
              disabled={busy || !!worktree}
              label={<span className="flex items-center gap-1"><UiWarningTriangle />{stagedFiles.length > 0 ? 'Lint staged' : 'Lint project'}</span>}
              title={worktree ? RUNNER_MAIN_CHECKOUT_ONLY : 'Lint options'}
              onClick={() => void run('lint')}
              items={[{
                label: <span className="flex items-center gap-2"><UiCog />Advanced…</span>,
                onSelect: () => setAdvancedAction('lint'),
              }]}
            />
            <SplitButton
              variant="outline"
              size="sm"
              disabled={busy || !!worktree}
              label={<span className="flex items-center gap-1"><UiBeaker />Test changed</span>}
              title={worktree ? RUNNER_MAIN_CHECKOUT_ONLY : 'Test options'}
              onClick={() => void run('test')}
              items={[{
                label: <span className="flex items-center gap-2"><UiCog />Advanced…</span>,
                onSelect: () => setAdvancedAction('test'),
              }]}
            />
            {commitButton}
            <SplitButton
              variant="outline"
              size="sm"
              primaryDisabled={queueCommitMutation.isPending || commitTaskError !== ''}
              title="Pull request options"
              label={<span className="flex items-center gap-1" title={stagedFiles.length > 0 ? 'Commit the staged files, push, and open a PR' : 'Push local commits and open or update the branch PR'}><UiGitPr />{stagedFiles.length > 0 ? `Commit & open PR (${stagedFiles.length})` : 'Open PR'}</span>}
              onClick={() => void queueStaged('open-pr')}
              items={rebaseAction ? [rebaseAction] : []}
            />
            <Button type="button" variant="ghost" size="icon" disabled={busy} onClick={refreshAfterCommit} aria-label="Refresh project status">
              <UiRefresh />
            </Button>
          </div>}
        </div>
        {notices}
      </div>}
      {reviewing && <div className="shrink-0 px-3">{notices}{commitTaskError && <div role="alert" className="text-xs text-destructive">{commitTaskError}</div>}</div>}

      <div hidden={reviewing} className={reviewing ? 'hidden' : 'contents'}>
      <ProjectCommitTasks
        key={`${project.name}\0${worktree ?? ''}`}
        projectName={project.name}
        worktree={worktree}
        preferredRunId={commitRunId}
        onLockedFilesChange={setLockedFiles}
        onErrorChange={setCommitTaskError}
        onComplete={refreshAfterCommit}
        onRunChange={setCommitRunId}
      />
      </div>

      <ProjectChangesView
        files={visibleFiles}
        locked={lockedFiles}
        disabled={treeDisabled}
        diffPath={diffPath}
        onDiffPathChange={onDiffPathChange}
        onToggleStaged={file => void toggleStaged(file)}
        fileList={
          <div className="flex h-full min-h-0 flex-col">
            <div className="min-h-0 flex-1 overflow-auto">
              {status.files.length === 0 ? (
                <Centered><UiCheck className="mr-1 text-green-600" />Working tree clean</Centered>
              ) : (
                <ProjectFileTree
                  files={visibleFiles}
                  search={filters.search}
                  onSearchChange={search => setFilters(current => ({ ...current, search }))}
                  filterActions={scopeFilter.length > 0 && <FilterBar className="min-w-0" filters={scopeFilter} />}
                  locked={lockedFiles}
                  disabled={treeDisabled}
                  showResults={showResults}
                  diffPath={diffPath}
                  editors={editors}
                  onDiffPathChange={onDiffPathChange}
                  onMenuOpen={() => setEditorsWanted(true)}
                  onStage={(paths, unstage) => void stage(paths, unstage)}
                  onDiscard={path => void discard(path)}
                  onIgnore={request => void ignore(request)}
                  onOpenInEditor={(path, editor) => void openInEditor(path, editor)}
                  onCopyPath={path => void copyPath(path)}
                />
              )}
            </div>
          </div>
        }
        diff={<ProjectDiffView projectName={project.name} path={diffPath} worktree={worktree} />}
      />
      {mobile && <div className="flex shrink-0 items-center justify-between gap-2 border-t border-border bg-background px-3 py-2 pb-[max(0.5rem,env(safe-area-inset-bottom))]">
        <Button type="button" variant="ghost" size="icon" disabled={busy} onClick={refreshAfterCommit} aria-label="Refresh project status" className="min-h-11 min-w-11"><UiRefresh /></Button>
        {commitButton}
      </div>}

      <ProjectActionDialog
        projectName={project.name}
        action={advancedAction}
        stagedFiles={stagedFiles}
        onClose={() => setAdvancedAction(null)}
        onRun={(action, options) => startAction(action, options)}
      />
      <ProjectGitLockPrompt owner={gitLock.owner} enabled={indexMutation.isPending || queueCommitMutation.isPending} />
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

function Centered({ children }: { children: React.ReactNode }) {
  return <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground">{children}</div>;
}
