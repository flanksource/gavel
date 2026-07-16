import type React from 'react';
import { fireEvent, render, renderHook, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useSessionStatus } from './TodoSession';
import { SessionErrorDetails } from './SessionErrorDetails';

vi.mock('@flanksource/clicky-ui/ai', () => ({
  SessionInspector: () => <div data-testid="session-inspector" />,
  SessionViewer: () => <div data-testid="session-viewer" />,
}));

vi.mock('@flanksource/clicky-ui/components', () => ({
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
}));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => {
  const Icon = (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />;
  return {
    ...(await importOriginal<object>()),
    UiCancel: Icon,
    UiCheck: Icon,
    UiChevronDown: Icon,
    UiChevronRight: Icon,
    UiCircleFilled: Icon,
    UiComment: Icon,
    UiCopy: Icon,
    UiError: Icon,
    UiLightbulb: Icon,
    UiPass: Icon,
    UiShield: Icon,
    UiWarningTriangle: Icon,
  };
});

describe('session errors', () => {
  it('keeps the complete backend error from a failed stats request', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
      text: async () => JSON.stringify({
        error: 'captain session conflict: provider session ID "session-1" is ambiguous',
        matches: ['captain-1', 'captain-2'],
      }),
    });
    vi.stubGlobal('fetch', fetchMock);

    const { result, unmount } = renderHook(() => useSessionStatus('/repo', 'session-1', true));

    await waitFor(() => expect(result.current.error).toContain('HTTP 500 Internal Server Error'));
    expect(result.current.error).toContain('provider session ID "session-1" is ambiguous');
    expect(result.current.error).toContain('"captain-1"');
    expect(result.current.error).toContain('"captain-2"');
    unmount();
    vi.unstubAllGlobals();
  });

  it('reveals and copies every labeled session error', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText },
    });
    render(<SessionErrorDetails errors={[
      { source: 'Session stream', message: 'stream failed\nprovider session ID is ambiguous' },
      { source: 'Session status', message: 'HTTP 500 Internal Server Error\nsecond backend detail' },
    ]} />);

    expect(screen.queryByText('provider session ID is ambiguous')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Show details' }));
    expect(screen.getByText(/provider session ID is ambiguous/)).toBeTruthy();
    expect(screen.getByText(/second backend detail/)).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Copy error details' }));
    await waitFor(() => expect(writeText).toHaveBeenCalledWith([
      'Session stream',
      'stream failed\nprovider session ID is ambiguous',
      '',
      'Session status',
      'HTTP 500 Internal Server Error\nsecond backend detail',
    ].join('\n')));
  });

  it('uses the DOM copy path when clipboard permission is denied', async () => {
    const writeText = vi.fn().mockRejectedValue(new Error('clipboard permission denied'));
    const execCommand = vi.fn().mockReturnValue(true);
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText },
    });
    Object.defineProperty(document, 'execCommand', {
      configurable: true,
      value: execCommand,
    });
    render(<SessionErrorDetails errors={[
      { source: 'Session status', message: 'HTTP 500 Internal Server Error' },
    ]} />);

    fireEvent.click(screen.getByRole('button', { name: 'Show details' }));
    fireEvent.click(screen.getByRole('button', { name: 'Copy error details' }));

    await waitFor(() => expect(execCommand).toHaveBeenCalledWith('copy'));
    expect(screen.getByRole('button', { name: 'Copy error details' }).textContent).toContain('Copied');
  });
});
