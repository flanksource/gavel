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
  preferredRunId?: string;
  onLockedFilesChange: (files: Map<string, number>) => void;
  onErrorChange: (error: string) => void;
  onComplete: () => void;
}

export function ProjectCommitTasks({
  projectName,
  preferredRunId,
  onLockedFilesChange,
  onErrorChange,
  onComplete,
}: Props) {
  const queryClient = useQueryClient();
  const labels = useMemo(() => ({ project: projectName }), [projectName]);
  const { runs, status: runsStatus } = useTaskRuns({
    basePath: '/api/v1',
    kind: 'gavel-commit',
    labels,
    enabled: !preferredRunId,
  });
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
    onSuccess: async () => queryClient.invalidateQueries({
      queryKey: queryKeys.projectStatusScope(projectName),
    }),
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
    if (!preferredRunId) queryClient.setQueryData(projectCommitTaskKeys.runs(projectName), runs);
  }, [preferredRunId, projectName, queryClient, runs]);
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
    void control({ runId, action });
  };
  const controlTask = (action: TaskControlAction, task: TaskSnapshot) => {
    void control({ runId, taskId: task.id, action });
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
