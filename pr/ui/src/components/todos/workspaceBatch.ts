import type { QueryClient } from '@tanstack/react-query';
import type { TodoListResponse } from '../../types';

interface Read { dir: string; signal: AbortSignal; resolve: (value: TodoListResponse) => void; reject: (error: unknown) => void }
const queues = new WeakMap<QueryClient, Read[]>();

export function readWorkspaceBatch(client: QueryClient, dir: string, signal: AbortSignal): Promise<TodoListResponse> {
  return new Promise((resolve, reject) => {
    let queue = queues.get(client);
    if (!queue) {
      queue = []; queues.set(client, queue);
      queueMicrotask(() => { queues.delete(client); void flush(queue!); });
    }
    queue.push({ dir, signal, resolve, reject });
  });
}

async function flush(queued: Read[]) {
  const reads = queued.filter(read => {
    if (!read.signal.aborted) return true;
    read.reject(read.signal.reason); return false;
  });
  if (!reads.length) return;
  const controller = new AbortController();
  const abort = () => {
    for (const read of reads) if (read.signal.aborted) read.reject(read.signal.reason);
    if (reads.every(read => read.signal.aborted)) controller.abort();
  };
  for (const read of reads) read.signal.addEventListener('abort', abort, { once: true });
  try {
    const dirs = [...new Set(reads.map(read => read.dir))];
    const response = await fetch('/api/todos/batch', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ dirs }), signal: controller.signal });
    if (!response.ok) throw new Error(`Load TODO workspaces: HTTP ${response.status}`);
    const payload: { results: (TodoListResponse & { error?: { message: string } })[] } = await response.json();
    if (!Array.isArray(payload.results) || payload.results.length !== dirs.length || payload.results.some((result, index) => result.dir !== dirs[index] || (!result.error && (!result.counts || !Array.isArray(result.items))))) throw new Error('Invalid TODO batch response');
    for (const read of reads) {
      if (read.signal.aborted) continue;
      const result = payload.results[dirs.indexOf(read.dir)];
      if (result.error) read.reject(new Error(result.error.message)); else read.resolve(result);
    }
  } catch (error) { for (const read of reads) read.reject(error); }
  finally { for (const read of reads) read.signal.removeEventListener('abort', abort); }
}
