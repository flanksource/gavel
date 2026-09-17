import { afterEach, describe, expect, it, vi } from 'vitest';
import { act, renderHook } from '@testing-library/react';
import { todoMutationStream } from './todoLaunch';
import { useTodoRun } from './run';
import { queryTestWrapper } from './queryTestWrapper';
import type { TodoRunResponse } from '../../types';

function streamResponse(chunks: string[]) {
  const encoder = new TextEncoder();
  return new Response(new ReadableStream<Uint8Array>({
    start(controller) {
      for (const chunk of chunks) controller.enqueue(encoder.encode(chunk));
      controller.close();
    },
  }), { headers: { 'Content-Type': 'text/event-stream' } });
}

afterEach(() => vi.unstubAllGlobals());

describe('todoMutationStream', () => {
  it('reports the exact resolved spec before returning the admitted attempt', async () => {
    const response = streamResponse([
      'event: resolved\ndata: {"spec":{"model":"gpt-5.5"},"specYaml":"model: gpt-5.5\\n","step":"run"}\n',
      '\nevent: admitted\ndata: {"status":"started","sessionId":"session-1","promptRunId":"prompt-1"}\n\n',
    ]);
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response));
    const resolved = vi.fn();
    const result = await todoMutationStream<{ promptRunId: string }>('/api/todos/run', { method: 'POST' }, 'Run failed', resolved);
    expect(resolved).toHaveBeenCalledWith({ spec: { model: 'gpt-5.5' }, specYaml: 'model: gpt-5.5\n', step: 'run' });
    expect(result).toEqual({ status: 'started', sessionId: 'session-1', promptRunId: 'prompt-1' });
  });

  it('surfaces a conflict after resolution with its status', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(streamResponse([
      'event: resolved\ndata: {"spec":{},"specYaml":"{}\\n","step":"run"}\n\n',
      'event: failed\ndata: {"error":"run already active","status":409}\n\n',
    ])));
    await expect(todoMutationStream('/api/todos/run', { method: 'POST' }, 'Run failed', vi.fn()))
      .rejects.toMatchObject({ message: 'run already active', status: 409 });
  });

  it('offers the existing parallel-run confirmation and retries with force', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(streamResponse([
        'event: resolved\ndata: {"spec":{},"specYaml":"{}\\n","step":"run"}\n\n',
        'event: failed\ndata: {"error":"run already active","status":409}\n\n',
      ]))
      .mockResolvedValueOnce(streamResponse([
        'event: resolved\ndata: {"spec":{},"specYaml":"{}\\n","step":"run"}\n\n',
        'event: admitted\ndata: {"status":"started","promptRunId":"run-new","sessionId":"session-new"}\n\n',
      ]));
    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('confirm', vi.fn().mockReturnValue(true));
    const { result } = renderHook(() => useTodoRun('/workspace'), { wrapper: queryTestWrapper() });
    let admitted: TodoRunResponse | null = null;
    await act(async () => { admitted = await result.current.run('todo-1', { step: 'run', spec: {} }); });
    expect(admitted).toMatchObject({ status: 'started', promptRunId: 'run-new' });
    expect(window.confirm).toHaveBeenCalledWith(expect.stringContaining('run already active'));
    expect(JSON.parse(String(fetchMock.mock.calls[1]?.[1]?.body))).toMatchObject({ ref: 'todo-1', force: true });
  });
});
