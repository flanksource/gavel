import { queryOptions, useQuery } from '@tanstack/react-query';
import { todoMutationJSON } from './todoMutations';

// NewTodoStep is one lifecycle step the server says applies to a todo that is
// about to be created with a given status. `prompt` names the template it runs;
// `auxiliary` marks steps that sit beside the main pipeline.
export interface NewTodoStep {
  name: string;
  label: string;
  prompt: string;
  auxiliary: boolean;
}

interface NewTodoStepsResponse {
  steps: NewTodoStep[];
}

export const newTodoStepsKey = (dir: string, status: string) => ['todos', 'new', 'steps', { dir: dir.trim(), status }] as const;

export function newTodoStepsQueryOptions(dir: string, status: string) {
  return queryOptions({
    queryKey: newTodoStepsKey(dir, status),
    queryFn: async ({ signal }) => {
      const params = new URLSearchParams();
      if (dir.trim()) params.set('dir', dir.trim());
      params.set('status', status);
      const response = await todoMutationJSON<NewTodoStepsResponse>(
        `/api/todos/new/steps?${params.toString()}`,
        { signal },
        'Failed to load lifecycle steps',
      );
      if (!Array.isArray(response?.steps)) throw new Error('Failed to load lifecycle steps: response omitted steps');
      return response.steps;
    },
    staleTime: 30_000,
  });
}

/** The lifecycle steps a new todo of `status` can be run through after it is
 *  created. The key carries the status, so changing it in the dialog fetches the
 *  set for the new status. */
export function useNewTodoSteps({ dir, status, enabled = true }: { dir: string; status: string; enabled?: boolean }) {
  return useQuery({ ...newTodoStepsQueryOptions(dir, status), enabled });
}
