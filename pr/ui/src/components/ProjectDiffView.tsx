import { useEffect, useState } from 'react';
import { GitDiffPanel, type GitDiffPayload } from '@flanksource/clicky-ui/data';
import { UiDiff } from '@flanksource/clicky-ui/icons';

interface Props {
  projectName: string;
  path: string;
  refreshKey?: number;
}

export function ProjectDiffView({ projectName, path, refreshKey = 0 }: Props) {
  const [payload, setPayload] = useState<GitDiffPayload | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!path) {
      setPayload(null);
      setError('');
      return;
    }
    const controller = new AbortController();
    setPayload(null);
    setLoading(true);
    setError('');
    fetch(`/api/projects/${encodeURIComponent(projectName)}/diff?path=${encodeURIComponent(path)}`, { signal: controller.signal })
      .then(async response => {
        const next = await response.json();
        if (!response.ok) throw new Error(next.error || `Failed to load diff for ${path}`);
        setPayload(next as GitDiffPayload);
      })
      .catch(cause => {
        if (cause instanceof DOMException && cause.name === 'AbortError') return;
        setError(cause instanceof Error ? cause.message : `Failed to load diff for ${path}`);
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [path, projectName, refreshKey]);

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
        loading={loading}
        payload={payload}
        error={error}
        className="min-h-0 flex-1 overflow-auto border-t-0"
        maxHeightClassName="max-h-none"
      />
    </div>
  );
}
