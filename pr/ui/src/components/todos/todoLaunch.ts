import { useQuery, type QueryClient } from '@tanstack/react-query';
import type { AISpecRuntimeValue } from '@flanksource/clicky-ui/ai';
import { TodoMutationError, todoMutationResponseError } from './todoMutations';
import { todoQueryKeys } from './todoQueries';

export interface TodoLaunchResolved {
  spec: AISpecRuntimeValue;
  specYaml: string;
  step: string;
  reason?: string;
}

export interface TodoLaunchProgress {
  status: 'preparing' | 'resolved' | 'admitted' | 'failed';
  step: string;
  requestedSpec?: AISpecRuntimeValue;
  spec?: AISpecRuntimeValue;
  specYaml?: string;
  promptRunId?: string;
  sessionId?: string;
  error?: string;
}

export function useTodoLaunchProgress(dir: string, ref: string) {
  return useQuery<TodoLaunchProgress | null>({
    queryKey: todoQueryKeys.launch(dir, ref),
    queryFn: async () => null,
    enabled: false,
  }).data ?? null;
}

export function setTodoLaunchProgress(client: QueryClient, dir: string, ref: string, progress: TodoLaunchProgress | null) {
  client.setQueryData(todoQueryKeys.launch(dir, ref), progress);
}

export function updateTodoLaunchProgress(client: QueryClient, dir: string, ref: string, update: (previous: TodoLaunchProgress) => TodoLaunchProgress) {
  client.setQueryData<TodoLaunchProgress | null>(todoQueryKeys.launch(dir, ref), previous => previous ? update(previous) : previous);
}

export async function todoMutationStream<T>(
  url: string,
  init: RequestInit,
  context: string,
  onResolved: (resolved: TodoLaunchResolved) => void,
): Promise<T> {
  let response: Response;
  const headers = new Headers(init.headers);
  headers.set('Accept', 'text/event-stream');
  try {
    response = await fetch(url, {
      ...init,
      headers,
    });
  } catch (cause) {
    throw new Error(`${context}: ${cause instanceof Error ? cause.message : 'request failed'}`, { cause });
  }
  if (!response.ok) throw await todoMutationResponseError(response, context);
  if (!response.headers.get('Content-Type')?.includes('text/event-stream') || !response.body) {
    throw new Error(`${context}: expected a launch progress stream`);
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let pending = '';
  let admitted: T | undefined;
  try {
    for (;;) {
      const { done, value } = await reader.read();
      pending = (pending + decoder.decode(value, { stream: !done })).replaceAll('\r\n', '\n');
      for (let boundary = pending.indexOf('\n\n'); boundary >= 0; boundary = pending.indexOf('\n\n')) {
        const frame = pending.slice(0, boundary);
        pending = pending.slice(boundary + 2);
        const event = frame.split('\n').find(line => line.startsWith('event: '))?.slice(7);
        const data = frame.split('\n').find(line => line.startsWith('data: '))?.slice(6);
        if (!event || !data) throw new Error(`${context}: invalid launch progress event`);
        const payload = JSON.parse(data) as Record<string, unknown>;
        if (event === 'resolved') onResolved(payload as unknown as TodoLaunchResolved);
        if (event === 'admitted') admitted = payload as T;
        if (event === 'failed') throw new TodoMutationError(String(payload.error || context), Number(payload.status || 500));
      }
      if (done) break;
    }
  } finally {
    reader.releaseLock();
  }
  if (admitted === undefined) throw new Error(`${context}: launch ended before admission`);
  return admitted;
}
