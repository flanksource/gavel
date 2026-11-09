import { useEffect, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { openEventStream } from '../../eventHub';

interface Snapshot { epoch: string; catalog: number; workspaces: Record<string, { revision: number }>; error?: string }
const snapshots = new WeakMap<object, Snapshot>();

export function useWorkspaceChanges(enabled: boolean) {
  const client = useQueryClient();
  const [ready, setReady] = useState(false);
  const [error, setError] = useState('');
  useEffect(() => {
    if (!enabled) return;
    const dirty = new Set<string>();
    let timer: ReturnType<typeof setTimeout> | undefined;
    const flush = () => {
      timer = undefined;
      const dirs = new Set(dirty); dirty.clear();
      void client.invalidateQueries({ predicate: query => query.queryKey[0] === 'todos' && query.queryKey.some(part => !!part && typeof part === 'object' && 'dir' in part && dirs.has(String(part.dir))) }, { cancelRefetch: false });
    };
    const stream = openEventStream('/api/todos/changes');
    const onMessage = (event: Event) => {
      try {
        const next: Snapshot = JSON.parse((event as MessageEvent<string>).data);
        if (!next.epoch || !Number.isInteger(next.catalog) || !next.workspaces) throw new Error('Invalid workspace change snapshot');
        if (next.error) { setError(next.error); return; }
        const previous = snapshots.get(client);
        const resync = !!previous && previous.epoch !== next.epoch;
        for (const dir of new Set([...Object.keys(previous?.workspaces ?? {}), ...Object.keys(next.workspaces)])) {
          if (resync || previous?.workspaces[dir]?.revision !== next.workspaces[dir]?.revision) dirty.add(dir);
        }
        if (previous && (resync || previous.catalog !== next.catalog)) {
          void client.invalidateQueries({ queryKey: ['projects'] });
        }
        snapshots.set(client, next);
        if (previous && dirty.size && !timer) timer = setTimeout(flush, 1000);
        else if (!previous) dirty.clear();
        setError(''); setReady(true);
      } catch (cause) { setError(cause instanceof Error ? cause.message : 'Workspace change feed failed'); }
    };
    stream.addEventListener('message', onMessage);
    stream.onerror = () => setError('Workspace change feed disconnected; reconnecting…');
    return () => { if (timer) clearTimeout(timer); stream.removeEventListener('message', onMessage); stream.close(); };
  }, [client, enabled]);
  return { ready, error };
}
