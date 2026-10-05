import { useCallback, useEffect, useMemo, useRef } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { TaskProgress, type TaskControlAction, type TaskSnapshot } from '@flanksource/clicky-ui/data';
import { useTaskRun, useTaskRuns } from '@flanksource/clicky-ui/hooks';
import { queryKeys } from '../query';
import {
  projectCommitLockedFiles,
  projectCommitTaskControlKey,
  projectCommitTaskKeys,
  requestProjectCommitTaskControl,
  type ProjectCommitTaskControl,
} from './project-commit-tasks';

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
  return (
    <section aria-label="Commit tasks" className="shrink-0 border-b border-border bg-muted/30 px-4 py-3">
      {reportedError && <div role="alert" className="mb-2 text-xs text-red-600 dark:text-red-400">{reportedError}</div>}
      {!ownership.error && (
        <TaskProgress
          snapshots={snapshots}
          compact
          onControl={controlGroup}
          onTaskControl={controlTask}
          metricsBaseUrl="/api/v1/tasks/metrics/"
        />
      )}
    </section>
  );
}
