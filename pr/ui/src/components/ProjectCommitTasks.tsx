import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Icon, TaskProgress, taskStatusColor, taskStatusIcon, type TaskControlAction, type TaskSnapshot } from '@flanksource/clicky-ui/data';
import { useTaskRun, useTaskRuns } from '@flanksource/clicky-ui/hooks';
import { Button } from '@flanksource/clicky-ui/components';
import { UiCheck, UiChevronDown, UiChevronUp } from '@flanksource/clicky-ui/icons';
import { formatDuration } from '@flanksource/clicky-ui/utils';
import { createPortal } from 'react-dom';
import { useIsMobile } from '../useIsMobile';
import { useCommitStatusSlot } from './commitStatusSlot';
import { queryKeys } from '../query';
import {
  projectCommitLockedFiles,
  projectCommitTaskControlKey,
  projectCommitTaskKeys,
  requestProjectCommitTaskControl,
  type ProjectCommitTaskControl,
} from './project-commit-tasks';

const COMMIT_CONFIRMATION_MS = 4000;

interface Props {
  projectName: string;
  // Absolute path of the linked worktree whose commit runs are shown; omitted
  // for the main checkout, which shows only runs that carry no worktree label.
  worktree?: string;
  preferredRunId?: string;
  onLockedFilesChange: (files: Map<string, number>) => void;
  onErrorChange: (error: string) => void;
  onComplete: () => void;
  onRunChange: (runId: string) => void;
}

export function ProjectCommitTasks({
  projectName,
  worktree,
  preferredRunId,
  onLockedFilesChange,
  onErrorChange,
  onComplete,
  onRunChange,
}: Props) {
  const mobile = useIsMobile();
  const [expandedRun, setExpandedRun] = useState('');
  const [watchedRun, setWatchedRun] = useState('');
  const [settledRun, setSettledRun] = useState('');
  const slot = useCommitStatusSlot();
  const queryClient = useQueryClient();
  const labels = useMemo(() => ({ project: projectName, ...(worktree ? { worktree } : {}) }), [projectName, worktree]);
  const listed = useTaskRuns({
    basePath: '/api/v1',
    kind: 'gavel-commit',
    labels,
    enabled: !preferredRunId,
  });
  const runsStatus = listed.status;
  // A label filter can only require a label, not its absence, so the main
  // checkout drops the runs the queue labelled with a worktree here. The same
  // check keeps a listing fetched for the previous ref off screen while a new
  // one is still loading.
  const runs = useMemo(
    () => listed.runs.filter(run => (run.labels?.worktree ?? '') === (worktree ?? '')),
    [listed.runs, worktree],
  );
  const runId = preferredRunId || runs[0]?.id || '';
  const { snapshots, isComplete, status: runStatus } = useTaskRun({
    id: runId,
    basePath: '/api/v1',
    enabled: runId !== '',
  });
  const completionReported = useRef('');
  const pendingControls = useRef(new Set<string>());
  const controlMutation = useMutation({
    mutationFn: requestProjectCommitTaskControl,
    onSuccess: async (followRunId, request) => {
      if (followRunId !== request.runId) onRunChange(followRunId);
      await queryClient.invalidateQueries({ queryKey: queryKeys.projectStatusScope(projectName) });
    },
  });

  const ownership = useMemo(() => {
    try {
      return { locked: projectCommitLockedFiles(snapshots), error: '' };
    } catch (cause) {
      return {
        locked: new Map<string, number>(),
        error: cause instanceof Error ? cause.message : 'Commit task metadata is invalid.',
      };
    }
  }, [snapshots]);

  useEffect(() => {
    if (!preferredRunId) queryClient.setQueryData(projectCommitTaskKeys.runs(projectName, worktree), runs);
  }, [preferredRunId, projectName, queryClient, runs, worktree]);
  useEffect(() => {
    if (runId) queryClient.setQueryData(projectCommitTaskKeys.run(projectName, runId), snapshots);
  }, [projectName, queryClient, runId, snapshots]);

  const controlError = controlMutation.error instanceof Error ? controlMutation.error.message : '';
  const runsError = runsStatus.startsWith('connection lost') ? `Commit task list ${runsStatus}.` : '';
  const runError = runStatus.startsWith('connection lost') ? `Commit task stream ${runStatus}.` : '';
  const reportedError = ownership.error || controlError || runError || runsError;

  const succeeded = isComplete && !reportedError && snapshots.find(snapshot => snapshot.type === 'group')?.status === 'success';
  const confirming = mobile && succeeded && watchedRun === runId && settledRun !== runId;
  useEffect(() => {
    if (runId && !isComplete) setWatchedRun(runId);
  }, [isComplete, runId]);
  useEffect(() => {
    if (!confirming) return;
    const timer = setTimeout(() => setSettledRun(runId), COMMIT_CONFIRMATION_MS);
    return () => clearTimeout(timer);
  }, [confirming, runId]);
  useEffect(() => onLockedFilesChange(ownership.locked), [onLockedFilesChange, ownership.locked]);
  useEffect(() => onErrorChange(reportedError), [onErrorChange, reportedError]);
  useEffect(() => {
    if (!preferredRunId || !runId || !isComplete || completionReported.current === runId) return;
    completionReported.current = runId;
    onComplete();
  }, [isComplete, onComplete, preferredRunId, runId]);

  const control = useCallback(async (request: ProjectCommitTaskControl) => {
    const key = projectCommitTaskControlKey(request);
    if (pendingControls.current.has(key)) return;
    pendingControls.current.add(key);
    controlMutation.reset();
    try {
      await controlMutation.mutateAsync(request);
    } catch {
      // React Query retains the error for the visible task-panel alert.
    } finally {
      pendingControls.current.delete(key);
    }
  }, [controlMutation.mutateAsync, controlMutation.reset]);
  const controlGroup = (action: TaskControlAction) => {
    void control({ projectName, runId, action });
  };
  const controlTask = (action: TaskControlAction, task: TaskSnapshot) => {
    void control({ projectName, runId, taskId: task.id, action });
  };

  if (!runId || snapshots.length === 0) {
    if (!reportedError) return null;
    return (
      <section aria-label="Commit tasks" className="shrink-0 border-b border-border px-4 py-3">
        <div role="alert" className="text-xs text-red-600 dark:text-red-400">{reportedError}</div>
      </section>
    );
  }
  const collapsed = mobile && succeeded;
  // On a wide screen a successful run reads as one line until it is opened; a
  // run in flight keeps its live progress and a failed one its retry controls.
  const summarized = !mobile && succeeded;
  const expanded = (!collapsed && !summarized) || expandedRun === runId;
  const toggleExpanded = () => setExpandedRun(expanded ? '' : runId);
  const chip = collapsed && (
    <Button variant="ghost" size="icon" type="button" className="size-8 shrink-0 text-green-700 dark:text-green-400" aria-expanded={expanded}
      aria-label={expanded ? 'Hide commit details' : 'Show commit details'} title={expanded ? 'Hide commit details' : 'Commit complete · show details'}
      onClick={toggleExpanded}>
      <UiCheck />
    </Button>
  );
  const chipOnly = collapsed && !confirming && !expanded && slot;
  return (
    <>
      {chip && slot && createPortal(chip, slot)}
      {!chipOnly && (
        <section aria-label="Commit tasks" className={`shrink-0 border-b border-border bg-muted/30 ${collapsed ? 'px-3 py-1' : summarized ? 'px-4 py-1.5' : 'px-4 py-3'}`}>
          {summarized && <CommitRunSummary snapshots={snapshots} expanded={expanded} onToggle={toggleExpanded} />}
          {confirming && <div role="status" className="flex items-center gap-1.5 text-xs font-medium text-green-700 dark:text-green-400"><UiCheck />Commit complete</div>}
          {chip && !slot && !confirming && <div className="flex items-center justify-between text-xs"><span className="text-muted-foreground">Commit complete</span>{chip}</div>}
          {reportedError && <div role="alert" className="mb-2 text-xs text-red-600 dark:text-red-400">{reportedError}</div>}
          {!ownership.error && expanded && !confirming && (
            <TaskProgress
              className={summarized ? 'mt-1.5' : undefined}
              snapshots={snapshots}
              compact
              onControl={controlGroup}
              onTaskControl={controlTask}
              metricsBaseUrl="/api/v1/tasks/metrics/"
            />
          )}
        </section>
      )}
    </>
  );
}

/** A finished commit run on one line: its outcome, its steps, how long it took, and a toggle for the full progress. */
function CommitRunSummary({ snapshots, expanded, onToggle }: { snapshots: TaskSnapshot[]; expanded: boolean; onToggle: () => void }) {
  const group = snapshots.find(snapshot => snapshot.type === 'group');
  if (!group) throw new Error('commit run has no group snapshot');
  const tasks = snapshots.filter(snapshot => snapshot.type === 'task');
  const done = tasks.filter(task => task.status !== 'running' && task.status !== 'pending').length;
  const elapsed = group.startedAt && group.finishedAt ? Date.parse(group.finishedAt) - Date.parse(group.startedAt) : undefined;
  return (
    <button type="button" aria-expanded={expanded} onClick={onToggle} data-commit-run-summary=""
      title={expanded ? 'Hide commit details' : 'Show commit details'}
      className="flex w-full min-w-0 items-center gap-2 text-left text-xs hover:text-foreground">
      <Icon icon={taskStatusIcon(group.status)} className={`shrink-0 ${taskStatusColor(group.status)}`} />
      <span className="shrink-0 font-medium text-foreground">{group.name}</span>
      <span className="flex min-w-0 flex-1 items-center gap-2 truncate text-muted-foreground">
        {tasks.map(task => (
          <span key={task.id} className="inline-flex min-w-0 items-center gap-0.5" title={task.error || task.status}>
            <Icon icon={taskStatusIcon(task.status)} className={`shrink-0 ${taskStatusColor(task.status)}`} />
            <span className="truncate">{task.name}</span>
          </span>
        ))}
      </span>
      {tasks.length > 0 && <span className="shrink-0 tabular-nums text-muted-foreground">{done}/{tasks.length}</span>}
      {elapsed !== undefined && <span className="shrink-0 tabular-nums text-muted-foreground">{formatDuration(elapsed)}</span>}
      <Icon icon={expanded ? UiChevronUp : UiChevronDown} className="shrink-0 text-muted-foreground" />
    </button>
  );
}
