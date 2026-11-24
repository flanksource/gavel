import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import type { InboxItem } from '@flanksource/clicky-ui/comments/gavel';
import { Button, Modal } from '@flanksource/clicky-ui/components';
import { ActionError } from '@flanksource/clicky-ui/data';
import { UiChatDots } from '@flanksource/clicky-ui/icons';
import { fetchJSON, mutationJSON } from '../../query';
import { projectApiUrl } from '../projectApiUrl';
import { projectEditorsQueryKey } from '../projectMutations';
import type { ProjectEditor } from '../projectStage';
import { commentListKey, type CommentListings } from './queries';
import { CommentDetails } from './CommentDetails';

function externalTarget(item: InboxItem): URL | null {
  const address = item.thread.content?.page?.url ?? (item.uri.startsWith(`${item.scope}/pages/`) ? item.label : item.uri);
  if (!URL.canParse(address)) return null;
  const url = new URL(address);
  return ['http:', 'https:'].includes(url.protocol) && !url.username && !url.password ? url : null;
}

function filePath(item: InboxItem): string | null {
  const prefix = `${item.scope}/files/`;
  return item.uri.startsWith(prefix) ? item.uri.slice(prefix.length) : null;
}

export function CommentsDetailPane({ query, selectedId, local, projectsLoaded, projectError }: {
  query: CommentListings;
  selectedId: string;
  local: boolean;
  projectsLoaded: boolean;
  projectError: string;
}) {
  const item = query.items.find(candidate => candidate.thread.id === selectedId);
  if (item) return <SelectedComment key={item.thread.id} item={item} local={local} />;
  const error = [projectError, ...query.errors].filter(Boolean).join(' · ');

  return (
    <div className="flex h-full flex-col items-center justify-center gap-2 p-4 text-center text-sm text-muted-foreground">
      <UiChatDots className="size-8" />
      {!selectedId ? <p>Select a comment to view its thread</p>
        : !projectsLoaded || query.pending ? <p role="status">Loading comment…</p>
        : error ? <p role="alert">Could not load this comment. {error}</p>
        : <p>Comment not found in this project scope.</p>}
    </div>
  );
}

function SelectedComment({ item, local }: { item: InboxItem; local: boolean }) {
  const [opening, setOpening] = useState(false);
  const cache = useQueryClient();
  const url = externalTarget(item);
  const target = item.line === undefined ? item.label : `${item.label}:${item.line}`;
  return (
    <section aria-label="Comment thread" className="flex h-full min-h-0 min-w-0 flex-col">
      <header className="flex shrink-0 flex-wrap items-center justify-between gap-x-3 gap-y-2 border-b border-border px-4 py-2">
        <div className="min-w-0 flex-1 basis-36">
          <div className="truncate text-xs text-muted-foreground" title={item.project}>{item.project}</div>
          <h1 className="truncate font-mono text-sm font-medium" title={target}>{target}</h1>
        </div>
        {url ? <Button asChild variant="outline" size="sm">
          <a href={url.href} target="_blank" rel="noopener noreferrer" aria-label={`Open ${target}`}>Open target</a>
        </Button> : local && filePath(item) ? (
          <Button variant="outline" size="sm" aria-label={`Open ${target}`} onClick={() => setOpening(true)}>Open in editor</Button>
        ) : null}
      </header>
      <div className="min-h-0 min-w-0 flex-1 overflow-y-auto p-3 [overflow-wrap:anywhere] sm:p-4">
        <CommentDetails item={item} local={local} refresh={() => cache.invalidateQueries({ queryKey: commentListKey(item.scope) })} />
      </div>
      {opening && <OpenCommentFile item={item} onClose={() => setOpening(false)} />}
    </section>
  );
}

function OpenCommentFile({ item, onClose }: { item: InboxItem; onClose: () => void }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);
  const editors = useQuery({
    queryKey: projectEditorsQueryKey,
    queryFn: ({ signal }) => fetchJSON<ProjectEditor[]>({ url: '/api/editors', signal, context: 'Load editors' }),
    staleTime: Infinity,
  });
  async function open(editor: string) {
    setBusy(true);
    setError(null);
    try {
      const path = filePath(item);
      if (!path) throw new Error('The comment does not name a project file');
      await mutationJSON({ url: projectApiUrl({ projectName: item.project, resource: 'open' }), method: 'POST', body: { path, editor, line: item.line }, context: `Open ${path}` });
      onClose();
    } catch (cause) {
      setError(cause);
    } finally {
      setBusy(false);
    }
  }

  return <Modal open title="Open in editor" onClose={() => { if (!busy) onClose(); }}>
    <p className="mb-3 break-all font-mono text-sm">{item.label}{item.line ? `:${item.line}` : ''}</p>
    {editors.isPending && <p role="status">Loading editors…</p>}
    {editors.data?.length === 0 && <p>No editors detected.</p>}
    <div className="flex flex-wrap gap-2">
      {editors.data?.map(editor => <Button key={editor.id} variant="outline" disabled={busy} onClick={() => void open(editor.id)}>{editor.label}</Button>)}
    </div>
    {(error || editors.error) && <ActionError error={error || editors.error} />}
  </Modal>;
}
