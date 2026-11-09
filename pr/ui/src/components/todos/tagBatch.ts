import type { QueryClient } from '@tanstack/react-query';
import type { TodoTagListResponse } from '../../types';

interface Read {
  dir: string;
  signal: AbortSignal;
  resolve: (value: TodoTagListResponse) => void;
  reject: (error: unknown) => void;
}

const queues = new WeakMap<QueryClient, Read[]>();

export function readTagBatch(client: QueryClient, dir: string, signal: AbortSignal): Promise<TodoTagListResponse> {
  return new Promise((resolve, reject) => {
    let queue = queues.get(client);
    if (!queue) {
      queue = [];
      queues.set(client, queue);
      queueMicrotask(() => {
        queues.delete(client);
        void flush(queue!);
      });
    }
    queue.push({ dir, signal, resolve, reject });
  });
}

async function flush(queued: Read[]) {
  const reads = queued.filter(read => {
    if (!read.signal.aborted) return true;
    read.reject(read.signal.reason);
    return false;
  });
  if (!reads.length) return;
  const controller = new AbortController();
  const onAbort = () => {
    for (const read of reads) if (read.signal.aborted) read.reject(read.signal.reason);
    if (reads.every(read => read.signal.aborted)) controller.abort();
  };
  for (const read of reads) read.signal.addEventListener('abort', onAbort, { once: true });
  try {
    const dirs = [...new Set(reads.map(read => read.dir))];
    const response = await fetch('/api/todos/labels/batch', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ dirs }), signal: controller.signal,
    });
    if (!response.ok) throw new Error(`Failed to load tag definitions: HTTP ${response.status}`);
    const payload: { results: { dir: string; labels?: TodoTagListResponse; error?: { message: string } }[]; palette: string[] } = await response.json();
    if (!Array.isArray(payload.results) || payload.results.length !== dirs.length
      || payload.results.some((result, i) => result.dir !== dirs[i]
        || (!result.error && !Array.isArray(result.labels?.definitions)))) {
      throw new Error('Invalid tag batch response');
    }
    const byDir = new Map(payload.results.map(result => [result.dir, result]));
    for (const read of reads) {
      if (read.signal.aborted) continue;
      const result = byDir.get(read.dir)!;
      if (result.error) read.reject(new Error(result.error.message));
      else read.resolve({ ...result.labels!, palette: payload.palette });
    }
  } catch (error) {
    for (const read of reads) read.reject(error);
  } finally {
    for (const read of reads) read.signal.removeEventListener('abort', onAbort);
  }
}
