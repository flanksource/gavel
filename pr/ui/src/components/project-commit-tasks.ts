import type { TaskControlAction, TaskSnapshot } from '@flanksource/clicky-ui/data';

interface ProjectCommitTaskDetails {
  taskId: string;
  action: 'commit' | 'open-pr';
  files: string[];
}

const activeStatuses = new Set(['pending', 'running']);

export const projectCommitTaskKeys = {
  runs: (projectName: string) => ['projects', projectName, 'commit-tasks', 'runs'] as const,
  run: (projectName: string, runId: string) => ['projects', projectName, 'commit-tasks', runId] as const,
};

export interface ProjectCommitTaskControl {
  runId: string;
  taskId?: string;
  action: TaskControlAction;
}

export function projectCommitTaskControlKey({ runId, taskId, action }: ProjectCommitTaskControl) {
  return `${runId}:${taskId ?? 'group'}:${action}`;
}

export async function requestProjectCommitTaskControl({ runId, taskId, action }: ProjectCommitTaskControl) {
  const runPath = `/api/v1/tasks/${encodeURIComponent(runId)}`;
  const response = await fetch(
    taskId ? `${runPath}/tasks/${encodeURIComponent(taskId)}/control` : `${runPath}/control`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ action }),
    },
  ).catch(cause => {
    throw new Error(`Failed to ${action} commit task: ${cause instanceof Error ? cause.message : 'request failed'}`, { cause });
  });
  if (!response.ok) {
    throw new Error(`Failed to ${action} commit task: ${(await response.text()).trim() || `HTTP ${response.status}`}`);
  }
}

export function projectCommitLockedFiles(snapshots: TaskSnapshot[]) {
  const active = snapshots.filter(snapshot => snapshot.type === 'task' && activeStatuses.has(snapshot.status));
  if (active.length === 0) return new Map<string, number>();

  const group = snapshots.find(snapshot => snapshot.type === 'group');
  const entries = projectCommitEntries(group?.details);
  if (!entries) {
    throw new Error('Active commit tasks are missing file ownership metadata.');
  }

  const positions = new Map(entries.map((entry, index) => [entry.taskId, { entry, position: index + 1 }]));
  const locked = new Map<string, number>();
  for (const task of active) {
    const owned = positions.get(task.id);
    if (!owned) {
      throw new Error(`Active commit task ${task.id} is missing file ownership metadata.`);
    }
    for (const file of owned.entry.files) locked.set(file, owned.position);
  }
  return locked;
}

function projectCommitEntries(details: TaskSnapshot['details']): ProjectCommitTaskDetails[] | null {
  if (!isRecord(details) || !Array.isArray(details.entries)) return null;
  return details.entries.map((value, index) => {
    if (!isRecord(value) || typeof value.taskId !== 'string' || !Array.isArray(value.files)) {
      throw new Error(`Commit task metadata entry ${index + 1} is invalid.`);
    }
    if (value.action !== 'commit' && value.action !== 'open-pr') {
      throw new Error(`Commit task metadata entry ${index + 1} has an invalid action.`);
    }
    if (!value.files.every(file => typeof file === 'string')) {
      throw new Error(`Commit task metadata entry ${index + 1} has invalid files.`);
    }
    return { taskId: value.taskId, action: value.action, files: value.files };
  });
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}
