export const workspaceTodoBatchKeys = {
  all: ['todos', 'batch'] as const,
  list: (dirs: readonly string[]) => ['todos', 'batch', { dirs: [...dirs] }] as const,
};
