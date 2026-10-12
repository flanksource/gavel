import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { GIT_FOCUS_RENEW_MS, useGitFocus } from './useGitFocus';

type Visibility = 'visible' | 'hidden';

function setVisibility(state: Visibility) {
  Object.defineProperty(document, 'visibilityState', { configurable: true, get: () => state });
  act(() => { document.dispatchEvent(new Event('visibilitychange')); });
}

const focusBodies = (fetchMock: ReturnType<typeof vi.fn>) =>
  fetchMock.mock.calls.map(([url, init]) => {
    expect(url).toBe('/api/git/focus');
    expect((init as RequestInit).method).toBe('POST');
    return JSON.parse((init as RequestInit).body as string);
  });

const noContent = () => Promise.resolve(new Response(null, { status: 204 }));

let fetchMock: ReturnType<typeof vi.fn>;

beforeEach(() => {
  vi.useFakeTimers();
  fetchMock = vi.fn(noContent);
  vi.stubGlobal('fetch', fetchMock);
  setVisibility('visible');
});

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('useGitFocus', () => {
  it('posts the focus on mount, omitting worktree for the main checkout', async () => {
    renderHook(() => useGitFocus({ project: 'gavel' }));
    await act(async () => {});

    expect(focusBodies(fetchMock)).toEqual([{ project: 'gavel' }]);
  });

  it('posts the worktree path when one is shown', async () => {
    renderHook(() => useGitFocus({ project: 'gavel', worktree: '/work/gavel-feat' }));
    await act(async () => {});

    expect(focusBodies(fetchMock)).toEqual([{ project: 'gavel', worktree: '/work/gavel-feat' }]);
  });

  it('renews the lease on every interval tick', async () => {
    renderHook(() => useGitFocus({ project: 'gavel' }));
    await act(async () => {});

    await act(async () => { await vi.advanceTimersByTimeAsync(GIT_FOCUS_RENEW_MS * 3); });

    expect(fetchMock).toHaveBeenCalledTimes(4);
  });

  it('stops renewing after unmount', async () => {
    const { unmount } = renderHook(() => useGitFocus({ project: 'gavel' }));
    await act(async () => {});
    unmount();

    await act(async () => { await vi.advanceTimersByTimeAsync(GIT_FOCUS_RENEW_MS * 3); });

    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it('does not post while the document is hidden and resumes when it is visible again', async () => {
    setVisibility('hidden');
    renderHook(() => useGitFocus({ project: 'gavel' }));
    await act(async () => { await vi.advanceTimersByTimeAsync(GIT_FOCUS_RENEW_MS * 3); });
    expect(fetchMock).not.toHaveBeenCalled();

    setVisibility('visible');
    await act(async () => {});

    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it('stops renewing once the document is hidden', async () => {
    renderHook(() => useGitFocus({ project: 'gavel' }));
    await act(async () => {});
    setVisibility('hidden');

    await act(async () => { await vi.advanceTimersByTimeAsync(GIT_FOCUS_RENEW_MS * 3); });

    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it('moves the lease to the new worktree when it changes', async () => {
    const { rerender } = renderHook(({ worktree }) => useGitFocus({ project: 'gavel', worktree }), {
      initialProps: { worktree: '/work/a' },
    });
    await act(async () => {});

    rerender({ worktree: '/work/b' });
    await act(async () => {});

    expect(focusBodies(fetchMock)).toEqual([
      { project: 'gavel', worktree: '/work/a' },
      { project: 'gavel', worktree: '/work/b' },
    ]);
  });

  it('does not post when disabled', async () => {
    renderHook(() => useGitFocus({ project: 'gavel', enabled: false }));
    await act(async () => { await vi.advanceTimersByTimeAsync(GIT_FOCUS_RENEW_MS * 2); });

    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('logs and exposes a failed post, and keeps renewing on the normal interval', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
    fetchMock.mockImplementation(() => Promise.resolve(Response.json({ error: 'boom' }, { status: 500 })));
    const { result } = renderHook(() => useGitFocus({ project: 'gavel' }));
    await act(async () => {});

    expect(result.current.error).toContain('boom');
    expect(consoleError).toHaveBeenCalled();

    await act(async () => { await vi.advanceTimersByTimeAsync(GIT_FOCUS_RENEW_MS); });
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('stops renewing after a 503 instead of retrying an unavailable tracker', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    fetchMock.mockImplementation(() => Promise.resolve(Response.json({ error: 'git tracker disabled' }, { status: 503 })));
    const { result } = renderHook(() => useGitFocus({ project: 'gavel' }));
    await act(async () => {});

    await act(async () => { await vi.advanceTimersByTimeAsync(GIT_FOCUS_RENEW_MS * 3); });

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(result.current.error).toContain('git tracker disabled');
  });

  it('clears the error after a successful renewal', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    fetchMock.mockImplementationOnce(() => Promise.resolve(Response.json({ error: 'boom' }, { status: 500 })));
    const { result } = renderHook(() => useGitFocus({ project: 'gavel' }));
    await act(async () => {});
    expect(result.current.error).toContain('boom');

    await act(async () => { await vi.advanceTimersByTimeAsync(GIT_FOCUS_RENEW_MS); });

    expect(result.current.error).toBe('');
  });
});
