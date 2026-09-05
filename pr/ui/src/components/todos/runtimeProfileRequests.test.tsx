import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useTodoRun } from './run';
import { useTodoVerificationRun } from './todoMutations';
import { usePlanActions } from './planActions';
import { queryTestWrapper } from './queryTestWrapper';

afterEach(() => vi.unstubAllGlobals());

describe('runtime profile request serialization', () => {
  const options = { runtimeProfile: 'review-profile', spec: { budget: { maxTurns: 8 } } };

  function captureRequest() {
    const fetch = vi.fn(async () => new Response(JSON.stringify({ status: 'started', todo: { ref: 'example-todo' } }), { status: 200 }));
    vi.stubGlobal('fetch', fetch);
    return () => JSON.parse(String((fetch.mock.calls[0] as unknown as [string, RequestInit])[1].body));
  }

  it('sends a profile with the ordinary named-step run', async () => {
    const body = captureRequest();
    const { result } = renderHook(() => useTodoRun('/repo'), { wrapper: queryTestWrapper() });
    await act(async () => { await result.current.run('example-todo', { ...options, step: 'plan' }); });
    expect(body()).toEqual({ ref: 'example-todo', step: 'plan', ...options });
  });

  it('sends a profile through verification', async () => {
    const body = captureRequest();
    const { result } = renderHook(() => useTodoVerificationRun('/repo', 'example-todo'), { wrapper: queryTestWrapper() });
    await act(async () => { await result.current.mutateAsync({ ref: 'example-todo', ...options }); });
    expect(body()).toEqual({ ref: 'example-todo', step: 'verify', ...options });
  });

  it('preserves a profile when approving a plan and continuing into run', async () => {
    const body = captureRequest();
    const { result } = renderHook(() => usePlanActions('/repo'), { wrapper: queryTestWrapper() });
    await act(async () => { await result.current.approve('example-todo', { run: true, options }); });
    expect(body()).toEqual({ ref: 'example-todo', run: true, options: { step: 'run', ...options } });
  });
});
