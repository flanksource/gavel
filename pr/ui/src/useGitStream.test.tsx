import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import type { PropsWithChildren } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { queryKeys } from './query';
import { useGitStream } from './useGitStream';

// The multiplexing hub has its own spec (eventHub.test.ts); here the stream is
// a plain fake so the spec drives frames directly.
vi.mock('./eventHub', () => ({
  openEventStream: (url: string) => new FakeEventSource(url),
}));

class FakeEventSource {
  static instances: FakeEventSource[] = [];
  readonly listeners = new Map<string, EventListener>();
  onerror: ((event: Event) => void) | null = null;
  readonly close = vi.fn();

  constructor(readonly url: string) {
    FakeEventSource.instances.push(this);
  }

  addEventListener(type: string, listener: EventListener) {
    this.listeners.set(type, listener);
  }

  emit(data: unknown) {
    const payload = typeof data === 'string' ? data : JSON.stringify(data);
    this.listeners.get('message')?.(new MessageEvent('message', { data: payload }));
  }
}

function setup(enabled = true) {
  const client = new QueryClient();
  const invalidate = vi.spyOn(client, 'invalidateQueries').mockResolvedValue();
  const wrapper = ({ children }: PropsWithChildren) => <QueryClientProvider client={client}>{children}</QueryClientProvider>;
  const hook = renderHook(({ on }) => useGitStream({ enabled: on }), { wrapper, initialProps: { on: enabled } });
  const stream = () => FakeEventSource.instances[FakeEventSource.instances.length - 1]!;
  const invalidatedKeys = () => invalidate.mock.calls.map(([filters]) => filters?.queryKey);
  return { hook, stream, invalidate, invalidatedKeys };
}

afterEach(() => {
  FakeEventSource.instances = [];
  vi.restoreAllMocks();
});

describe('useGitStream', () => {
  it('subscribes to /api/git/stream only while enabled', () => {
    const { hook } = setup(false);
    expect(FakeEventSource.instances).toHaveLength(0);

    hook.rerender({ on: true });

    expect(FakeEventSource.instances.map(instance => instance.url)).toEqual(['/api/git/stream']);
  });

  it('closes the stream on unmount', () => {
    const { hook, stream } = setup();
    hook.unmount();
    expect(stream().close).toHaveBeenCalledTimes(1);
  });

  it('records the first frame without invalidating anything', () => {
    const { stream, invalidate } = setup();

    act(() => stream().emit({ gavel: 4, clicky: 9 }));

    expect(invalidate).not.toHaveBeenCalled();
  });

  it('invalidates exactly the changed project git and status keys plus the summary', () => {
    const { stream, invalidatedKeys } = setup();
    act(() => stream().emit({ gavel: 4, clicky: 9 }));

    act(() => stream().emit({ gavel: 5, clicky: 9 }));

    expect(invalidatedKeys()).toEqual([
      queryKeys.projectGit('gavel'),
      queryKeys.projectStatusScope('gavel'),
      queryKeys.projectGitSummary(),
      queryKeys.agentSessions(),
    ]);
  });

  it('invalidates the summary once when several projects change in one frame', () => {
    const { stream, invalidatedKeys } = setup();
    act(() => stream().emit({ gavel: 4, clicky: 9 }));

    act(() => stream().emit({ gavel: 5, clicky: 10 }));

    expect(invalidatedKeys()).toEqual([
      queryKeys.projectGit('gavel'),
      queryKeys.projectStatusScope('gavel'),
      queryKeys.projectGit('clicky'),
      queryKeys.projectStatusScope('clicky'),
      queryKeys.projectGitSummary(),
      queryKeys.agentSessions(),
    ]);
  });

  it('invalidates nothing when a frame repeats the generations already seen', () => {
    const { stream, invalidate } = setup();
    act(() => stream().emit({ gavel: 4 }));

    act(() => stream().emit({ gavel: 4 }));

    expect(invalidate).not.toHaveBeenCalled();
  });

  it('treats a project that first appears after the first frame as changed', () => {
    const { stream, invalidatedKeys } = setup();
    act(() => stream().emit({ gavel: 4 }));

    act(() => stream().emit({ gavel: 4, clicky: 1 }));

    expect(invalidatedKeys()).toEqual([
      queryKeys.projectGit('clicky'),
      queryKeys.projectStatusScope('clicky'),
      queryKeys.projectGitSummary(),
      queryKeys.agentSessions(),
    ]);
  });

  it('keeps the seen generations across a disable and re-enable so a change in the gap is caught', () => {
    const { hook, stream, invalidatedKeys } = setup();
    act(() => stream().emit({ gavel: 4 }));
    hook.rerender({ on: false });
    hook.rerender({ on: true });

    act(() => stream().emit({ gavel: 6 }));

    expect(invalidatedKeys()).toEqual([
      queryKeys.projectGit('gavel'),
      queryKeys.projectStatusScope('gavel'),
      queryKeys.projectGitSummary(),
      queryKeys.agentSessions(),
    ]);
  });

  it.each([
    ['malformed JSON', '{not json'],
    ['a non-object frame', [1, 2]],
    ['a non-numeric generation', { gavel: 'four' }],
  ])('reports %s as a stream error instead of ignoring it', (_label, frame) => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
    const { hook, stream, invalidate } = setup();

    act(() => stream().emit(frame));

    expect(hook.result.current.error).toMatch(/invalid update/i);
    expect(consoleError).toHaveBeenCalled();
    expect(invalidate).not.toHaveBeenCalled();
  });

  it('reports a dropped connection and clears it on the next valid frame', () => {
    const { hook, stream } = setup();

    act(() => stream().onerror?.(new Event('error')));
    expect(hook.result.current.error).toMatch(/disconnected/i);

    act(() => stream().emit({ gavel: 1 }));
    expect(hook.result.current.error).toBe('');
  });

  it('clears the stream error on the next valid frame', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    const { hook, stream } = setup();
    act(() => stream().emit('{not json'));

    act(() => stream().emit({ gavel: 1 }));

    expect(hook.result.current.error).toBe('');
  });
});
