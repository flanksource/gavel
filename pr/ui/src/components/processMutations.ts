import { useCallback, useRef } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { mutationJSON, queryKeys } from '../query';
import type { ProcStatus, Project } from '../types';

export type ProcessAction = 'start' | 'stop' | 'restart';

interface ProcessControl {
  action: ProcessAction;
  names?: string[];
  profile?: string;
}

export function useProcessControl(project: Project, onChanged: () => void) {
  const queryClient = useQueryClient();
  const inFlight = useRef(false);
  const mutation = useMutation({
    mutationKey: ['processes', project.name, 'control'],
    mutationFn: async ({ action, names, profile }: ProcessControl) => parseProcessStatus(
      await mutationJSON<unknown>({
        url: `/api/proc/${action}`,
        method: 'POST',
        body: { project: project.name, names, profile },
        context: `Failed to ${action} processes for ${project.name}`,
      }),
      action,
      project.name,
    ),
    onSuccess: async status => {
      await Promise.all([
        queryClient.cancelQueries({ queryKey: queryKeys.processStatuses(), exact: true }),
        queryClient.cancelQueries({ queryKey: queryKeys.processStatus(project.name), exact: true }),
      ]);
      queryClient.setQueryData(queryKeys.processStatus(project.name), status);
      queryClient.setQueryData<Record<string, ProcStatus>>(queryKeys.processStatuses(), current => {
        const updated: Record<string, ProcStatus> = current ? { ...current } : {};
        updated[project.name] = status;
        for (const repo of project.repos) updated[repo] = status;
        return updated;
      });
      await queryClient.invalidateQueries({ queryKey: queryKeys.projectStatusScope(project.name) });
      onChanged();
    },
  });
  const mutate = mutation.mutate;

  const control = useCallback((request: ProcessControl) => {
    if (inFlight.current) return;
    inFlight.current = true;
    mutate(request, { onSettled: () => { inFlight.current = false; } });
  }, [mutate]);

  return {
    control,
    busy: mutation.isPending,
    error: mutation.error instanceof Error ? mutation.error.message : '',
  };
}

function parseProcessStatus(payload: unknown, action: ProcessAction, project: string): ProcStatus {
  if (!payload || typeof payload !== 'object' || !('hasProcfile' in payload) || typeof payload.hasProcfile !== 'boolean' || !('running' in payload) || typeof payload.running !== 'boolean') {
    throw new Error(`Failed to ${action} processes for ${project}: invalid response`);
  }
  return payload as ProcStatus;
}
