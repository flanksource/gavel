import { useQuery } from '@tanstack/react-query';
import { GitDiffPanel, type GitDiffPayload } from '@flanksource/clicky-ui/data';
import { UiDiff } from '@flanksource/clicky-ui/icons';
import { fetchJSON, queryKeys } from '../query';

interface Props {
  projectName: string;
  path: string;
  refreshKey?: number;
}

export function ProjectDiffView({ projectName, path, refreshKey = 0 }: Props) {
  const diff = useQuery({
    queryKey: queryKeys.projectDiff(projectName, path, refreshKey),
    queryFn: ({ signal }) => fetchJSON<GitDiffPayload>({
      url: `/api/projects/${encodeURIComponent(projectName)}/diff?path=${encodeURIComponent(path)}`,
      signal,
      context: `Failed to load diff for ${path}`,
    }),
    enabled: path !== '',
    staleTime: Infinity,
  });

  if (!path) {
    return (
      <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground">
        <div className="text-center"><UiDiff className="mx-auto mb-2 text-3xl" />Select a file or folder to view its diff</div>
      </div>
    );
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex shrink-0 items-center gap-2 border-b border-border px-4 py-3">
        <UiDiff className="shrink-0 text-muted-foreground" />
        <h3 className="min-w-0 flex-1 truncate font-mono text-sm font-semibold" title={path}>{path}</h3>
      </div>
      <GitDiffPanel
        loading={diff.isPending}
        payload={diff.data ?? null}
        error={diff.error instanceof Error ? diff.error.message : ''}
        className="min-h-0 flex-1 overflow-auto border-t-0"
        maxHeightClassName="max-h-none"
      />
    </div>
  );
}
