import { useEffect, useState } from 'react';
import { useQueryClient, type Query } from '@tanstack/react-query';
import { openEventStream } from '../../eventHub';

interface Generation { revision: number; labels?: number; config?: number }
interface Snapshot { epoch: string; catalog: number; workspaces: Record<string, Generation>; error?: string }
const snapshots = new WeakMap<object, Snapshot>();
const pending = new WeakMap<object, Set<string>>();

// The all-workspace list is keyed by its exclude filter, not by a workspace dir.
function isWorkspaceListQuery(query: Query): boolean {
  return query.queryKey[1] === 'list' && query.queryKey.some(part => !!part && typeof part === 'object' && 'exclude' in part);
}

function workspaceDir(query: Query): string {
  if (query.queryKey[1] === 'run-context') return String(query.queryKey[2] ?? '');
  const scope = query.queryKey.find(part => !!part && typeof part === 'object' && 'dir' in part);
  if (scope && typeof scope === 'object' && 'dir' in scope) return String(scope.dir);
  const data = query.state.data;
  if (data && typeof data === 'object' && 'cwd' in data) return String(data.cwd);
  return '';
}

export function useWorkspaceChanges(enabled: boolean) {
  const client = useQueryClient();
  const [ready, setReady] = useState(false);
  const [error, setError] = useState('');
  useEffect(() => {
    if (!enabled) return;
    const dirty = pending.get(client) ?? new Set<string>();
    pending.set(client, dirty);
    let timer: ReturnType<typeof setTimeout> | undefined;
    const schedule = () => {
      if (!timer && client.getQueryCache().getAll().some(query => dirty.has(query.queryHash) && query.state.fetchStatus !== 'fetching')) timer = setTimeout(flush, 1000);
    };
    const flush = () => {
      timer = undefined;
      for (const hash of dirty) {
        const query = client.getQueryCache().get(hash);
        if (query?.state.fetchStatus === 'fetching') continue;
        dirty.delete(hash);
        if (query) void client.invalidateQueries({ predicate: candidate => candidate.queryHash === hash }, { cancelRefetch: false });
      }
    };
    const unsubscribe = client.getQueryCache().subscribe(schedule);
    schedule();
    const stream = openEventStream('/api/todos/changes');
    const onMessage = (event: Event) => {
      try {
        const next: Snapshot = JSON.parse((event as MessageEvent<string>).data);
        if (!next.epoch || !Number.isInteger(next.catalog) || !next.workspaces) throw new Error('Invalid workspace change snapshot');
        if (next.error) { setError(next.error); return; }
        const previous = snapshots.get(client);
        const resync = !!previous && previous.epoch !== next.epoch;
        // The flat list spans every workspace, so any workspace's revision, the
        // workspace catalog, or a resync can change what it holds.
        const anyWorkspaceChanged = !!previous && (resync || previous.catalog !== next.catalog
          || [...new Set([...Object.keys(previous.workspaces), ...Object.keys(next.workspaces)])]
            .some(dir => previous.workspaces[dir]?.revision !== next.workspaces[dir]?.revision));
        if (previous) for (const query of client.getQueryCache().getAll()) {
          if (query.queryKey[0] !== 'todos') continue;
          if (isWorkspaceListQuery(query)) {
            if (anyWorkspaceChanged) dirty.add(query.queryHash);
            continue;
          }
          const dir = workspaceDir(query);
          if (!dir) continue;
          const field = query.queryKey[1] === 'tags' ? 'labels' : query.queryKey[1] === 'run-context' ? 'config' : 'revision';
          if (resync || previous.workspaces[dir]?.[field] !== next.workspaces[dir]?.[field]) dirty.add(query.queryHash);
        }
        if (anyWorkspaceChanged) {
          void client.invalidateQueries({ queryKey: ['projects'] });
        }
        snapshots.set(client, next);
        schedule();
        setError(''); setReady(true);
      } catch (cause) { setError(cause instanceof Error ? cause.message : 'Workspace change feed failed'); }
    };
    stream.addEventListener('message', onMessage);
    stream.onerror = () => setError('Workspace change feed disconnected; reconnecting…');
    return () => { if (timer) clearTimeout(timer); unsubscribe(); stream.removeEventListener('message', onMessage); stream.close(); };
  }, [client, enabled]);
  return { ready, error };
}
