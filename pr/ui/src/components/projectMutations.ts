import type { QueryClient } from '@tanstack/react-query';
import { queryKeys } from '../query';

export const projectDiffQueryKey = (projectName: string) => ['projects', projectName, 'diff'] as const;

// The editors detected on the gavel host; the same for every project.
export const projectEditorsQueryKey = ['editors'] as const;

// invalidateProjectData refreshes everything a git or index change can alter:
// the status tree and every cached diff of the project.
export async function invalidateProjectData(queryClient: QueryClient, projectName: string): Promise<void> {
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: queryKeys.projectStatusScope(projectName) }),
    queryClient.invalidateQueries({ queryKey: projectDiffQueryKey(projectName) }),
  ]);
}
