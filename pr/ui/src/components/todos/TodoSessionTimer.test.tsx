import type React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { CmuxSessionButton, cmuxSurfaceLabel } from './TodoSessionTimer';

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
  vi.restoreAllMocks();
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

    render(<CmuxSessionButton dir="/repo" sessionId="sess-9" agent="claude" onResume={vi.fn()} />);

    const focus = await screen.findByRole('button', { name: /Focus in cmux/ });
    expect((focus as HTMLButtonElement).disabled).toBe(true);
    expect(screen.getByText(/session terminal may have been closed/)).toBeTruthy();
    // Resume stays available to reopen the closed terminal.
    expect((screen.getByRole('button', { name: /Resume in cmux/ }) as HTMLButtonElement).disabled).toBe(false);
  });
});
