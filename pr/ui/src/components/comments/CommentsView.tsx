import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { CommentInbox, type InboxFilters, type InboxItem } from '@flanksource/clicky-ui/comments/gavel';
import { Button, Modal } from '@flanksource/clicky-ui/components';
import { ActionError } from '@flanksource/clicky-ui/data';
import { fetchJSON, mutationJSON } from '../../query';
import type { Project } from '../../types';
import { projectApiUrl } from '../projectApiUrl';
import { projectEditorsQueryKey } from '../projectMutations';
import type { ProjectEditor } from '../projectStage';
import { capabilitiesQuery, commentClient, commentListKey, useCommentListings } from './queries';
import { CommentDetails } from './CommentDetails';

interface CommentsViewProps {
  projects: Project[];
  scopedProjects: Project[];
  projectsLoaded: boolean;
  projectError: string;
  local: boolean;
  visible: boolean;
  filters: InboxFilters;
  selectedId: string;
  onSelect: (id: string | undefined) => void;
  onFiltersChange: (filters: InboxFilters) => void;
}

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

export function CommentsView({ projects, scopedProjects, projectsLoaded, projectError, local, visible, filters, selectedId, onSelect, onFiltersChange }: CommentsViewProps) {
  const { listings, errors } = useCommentListings(scopedProjects, projectsLoaded && (projects.length > 0 || !projectError), visible);
  const [opening, setOpening] = useState<InboxItem>();
  const [readError, setReadError] = useState('');
  const cache = useQueryClient();
  const refresh = (scope: string) => cache.invalidateQueries({ queryKey: commentListKey(scope) });

  async function review(item: InboxItem, action: 'close' | 'reopen', feedback?: string) {
    try {
      const actions = await cache.fetchQuery(capabilitiesQuery(item.thread.id, item.thread.version));
      if (!actions.includes(action)) throw new Error(`This thread cannot be ${action === 'close' ? 'closed' : 'reopened'}`);
      await commentClient(item.scope).act({ id: item.thread.id, version: item.thread.version, action, ...(feedback ? { bodyMarkdown: feedback } : {}) });
    } finally {
      await refresh(item.scope);
    }
  }

  async function markRead({ scope, uri }: { scope: string; uri: string }) {
    setReadError('');
    try {
      await commentClient(scope).markRead(uri);
    } catch (cause) {
      setReadError(cause instanceof Error ? cause.message : String(cause));
    } finally {
      await refresh(scope);
    }
  }

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden">
      {readError && <p role="alert" className="px-3 text-sm text-destructive">{readError}</p>}
      <CommentInbox
        listings={listings}
        error={[projectError, ...errors].filter(Boolean).join(' · ') || null}
        filters={filters}
        onFiltersChange={onFiltersChange}
        projectOptions={projects.map(({ name }) => ({ value: name, label: name }))}
        baseUrl={window.location.origin}
        expandedThreadId={selectedId || undefined}
        onExpandedThreadChange={onSelect}
        canOpen={(item) => Boolean(externalTarget(item) || (local && filePath(item)))}
        onOpen={(item) => {
          const url = externalTarget(item);
          if (url) window.open(url.href, '_blank', 'noopener,noreferrer');
          else setOpening(item);
        }}
        canReview={(item) => item.thread.phase === 'awaiting_outcome'}
        actions={{ onClose: item => review(item, 'close'), onReopenWithFeedback: (item, feedback) => review(item, 'reopen', feedback), onMarkRead: markRead }}
        renderThread={(item) => <CommentDetails key={item.thread.id} item={item} local={local} refresh={() => refresh(item.scope)} />}
      />
      {opening && <OpenCommentFile item={opening} onClose={() => setOpening(undefined)} />}
    </div>
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
