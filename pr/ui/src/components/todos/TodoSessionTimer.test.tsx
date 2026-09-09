import type React from 'react';
import { act, fireEvent, render, renderHook, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { CmuxSessionButton, cmuxSurfaceLabel, useSessionStats } from './TodoSessionTimer';
import { queryTestWrapper } from './queryTestWrapper';

// DropdownMenu is mocked to render its trigger and menu content inline and to
// report itself open (so the lazy cmux-surface fetch runs), letting the test
// exercise the menu items without driving the floating-ui open/close machinery.
vi.mock('@flanksource/clicky-ui/components', async () => {
  const React = await import('react');
  return {
    Button: ({
      children,
      variant: _variant,
      size: _size,
      ...props
    }: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: string; size?: string }) => (
      // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
      <button type="button" {...props}>
        {children}
      </button>
    ),
    DropdownMenu: ({
      trigger,
      children,
      onOpenChange,
    }: {
      trigger: React.ReactNode;
      children: (close: () => void) => React.ReactNode;
      onOpenChange?: (open: boolean) => void;
    }) => {
      React.useEffect(() => {
        onOpenChange?.(true);
      }, [onOpenChange]);
      return (
        <div>
          {trigger}
          {children(() => {})}
        </div>
      );
    },
  };
});

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => {
  const Icon = (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />;
  return {
    ...(await importOriginal<object>()),
    UiChevronDown: Icon,
    UiClock: Icon,
    UiCollapseAll: Icon,
    UiDebugStepOver: Icon,
    UiError: Icon,
    UiEye: Icon,
    UiHubot: Icon,
    UiUnknown: Icon,
  };
});

function mockFetch(handlers: Record<string, () => unknown>) {
  return vi.fn(async (input: RequestInfo | URL) => {
    const url = typeof input === 'string' ? input : input.toString();
    const key = Object.keys(handlers).find(fragment => url.includes(fragment));
    const body = key ? handlers[key]() : {};
    return { ok: true, json: async () => body } as Response;
  });
}

afterEach(() => {
	vi.useRealTimers();
  vi.restoreAllMocks();
});

describe('useSessionStats', () => {
  it('keeps failed response details out of stats render state', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
      text: async () => JSON.stringify({
        error: 'captain session conflict: provider session ID "session-1" is ambiguous',
      }),
    }));

    const { result, unmount } = renderHook(() => useSessionStats({ dir: '/repo', sessionId: 'session-1', active: true }), {
      wrapper: queryTestWrapper(),
    });

    await waitFor(() => expect(result.current.error).toContain('HTTP 500 Internal Server Error'));
    expect(result.current.error).toContain('provider session ID "session-1" is ambiguous');
    expect(result.current.stats).toBeNull();
    unmount();
  });

  it('continues the live clock across server polls', async () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-07-13T12:00:00Z'));
    const durations = [30_000, 32_000];
    const fetchMock = vi.fn(async () => {
      const durationMs = durations.shift() ?? 32_000;
      return {
        ok: true,
        json: async () => ({
          found: true,
          inProgress: true,
          durationMs,
          inputTokens: 0,
          outputTokens: 0,
          cacheReadTokens: 0,
          cacheCreationTokens: 0,
          totalTokens: 0,
          contextTokens: 0,
          contextWindow: 0,
          turns: 0,
          compactions: 0,
          costUsd: 0,
        }),
      } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    const { result, unmount } = renderHook(() => useSessionStats({ dir: '/repo', sessionId: 'session-clock', active: true }), {
      wrapper: queryTestWrapper(),
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });
    expect(result.current.elapsedMs).toBe(30_000);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1_000);
    });
    expect(result.current.elapsedMs).toBe(31_000);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1_000);
    });
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(result.current.elapsedMs).toBeGreaterThanOrEqual(32_000);
    unmount();
  });
});

describe('cmuxSurfaceLabel', () => {
  it('joins the resolved workspace and surface after the Cmux prefix', () => {
    expect(cmuxSurfaceLabel('workspace:2', 'surface:1')).toBe('Cmux workspace:2 surface:1');
  });

  it('drops missing parts so a workspace with no surface still labels', () => {
    expect(cmuxSurfaceLabel('workspace:2', undefined)).toBe('Cmux workspace:2');
    expect(cmuxSurfaceLabel(undefined, undefined)).toBe('Cmux');
  });
});

describe('CmuxSessionButton', () => {
  it('captions the resolved workspace/surface for the session and focuses on demand', async () => {
    const fetchMock = mockFetch({
      '/session/cmux': () => ({ found: true, workspace: 'workspace:2', surface: 'surface:1' }),
      '/session/focus': () => ({ focused: true }),
    });
    vi.stubGlobal('fetch', fetchMock);

    render(
      <CmuxSessionButton dir="/repo" sessionId="sess-1234" agent="claude" onResume={vi.fn()} />,
      { wrapper: queryTestWrapper() },
    );

    // The surface comment maps the session to its cmux terminal.
    expect(await screen.findByText('Cmux workspace:2 surface:1')).toBeTruthy();
    expect(screen.getByText('for session: sess-1234')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: /Focus in cmux/ }));
    await waitFor(() =>
      expect(fetchMock.mock.calls.some(([url]) => String(url).includes('/api/todos/session/focus'))).toBe(true),
    );
  });

  it('resumes via the supplied handler and hides Resume without one', async () => {
    const fetchMock = mockFetch({ '/session/cmux': () => ({ found: true, workspace: 'workspace:2', surface: 'surface:1' }) });
    vi.stubGlobal('fetch', fetchMock);
    const onResume = vi.fn();

    const { rerender } = render(
      <CmuxSessionButton dir="/repo" sessionId="sess-1234" agent="claude" onResume={onResume} />,
      { wrapper: queryTestWrapper() },
    );

    fireEvent.click(await screen.findByRole('button', { name: /Resume in cmux/ }));
    expect(onResume).toHaveBeenCalledTimes(1);

    rerender(<CmuxSessionButton dir="/repo" sessionId="sess-1234" agent="claude" />);
    expect(screen.queryByRole('button', { name: /Resume in cmux/ })).toBeNull();
  });

  it('disables Focus and explains when the terminal is gone', async () => {
    const fetchMock = mockFetch({
      '/session/cmux': () => ({ found: false, reason: 'no cmux workspace; the session terminal may have been closed' }),
    });
    vi.stubGlobal('fetch', fetchMock);

    render(<CmuxSessionButton dir="/repo" sessionId="sess-9" agent="claude" onResume={vi.fn()} />, {
      wrapper: queryTestWrapper(),
    });

    const focus = await screen.findByRole('button', { name: /Focus in cmux/ });
    await waitFor(() => expect((focus as HTMLButtonElement).disabled).toBe(true));
    expect(screen.getByText(/session terminal may have been closed/)).toBeTruthy();
    // Resume stays available to reopen the closed terminal.
    expect((screen.getByRole('button', { name: /Resume in cmux/ }) as HTMLButtonElement).disabled).toBe(false);
  });

  it('surfaces the contextual server error when focus fails', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/session/cmux')) {
        return { ok: true, json: async () => ({ found: true, workspace: 'workspace:2', surface: 'surface:1' }) };
      }
      if (url.includes('/session/focus')) {
        return {
          ok: false,
          status: 502,
          json: async () => ({ error: 'cmux socket is unavailable' }),
          text: async () => JSON.stringify({ error: 'cmux socket is unavailable' }),
        };
      }
      throw new Error(`unexpected fetch ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);

    render(<CmuxSessionButton dir="/repo" sessionId="sess-1234" agent="claude" />, {
      wrapper: queryTestWrapper(),
    });
    fireEvent.click(await screen.findByRole('button', { name: /Focus in cmux/ }));

    expect(await screen.findByText(/could not focus cmux session.*cmux socket is unavailable/i)).toBeTruthy();
  });
});
