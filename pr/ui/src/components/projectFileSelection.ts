import type { ProjectFileStatus } from './ProjectStatusView';

export function isProjectFileCommittable(file: ProjectFileStatus): boolean {
  return file.state !== 'conflict' && !file.linkedDependencies?.length;
}
