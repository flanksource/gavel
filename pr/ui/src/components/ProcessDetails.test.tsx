import type React from 'react';
import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { ProcProcess } from '../types';
import { ProcExpanded } from './ProcessDetails';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => <button {...props}>{children}</button>,
  Modal: ({ open, children }: { open: boolean; children: ReactNode }) => open ? <section>{children}</section> : null,
}));

vi.mock('@flanksource/clicky-ui/data', () => ({
  AnsiHtml: ({ text }: { text: string }) => <pre>{text}</pre>,
}));

vi.mock('@flanksource/clicky-ui/icons', () => ({ UiFullscreen: () => <span /> }));
vi.mock('../useNow', () => ({ useNow: () => 0 }));

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'visible' });
});

function createClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: Infinity } } });
}

function process(name: string): ProcProcess {
  return { name, command: `${name} --serve`, status: 'running', restarts: 0, logFile: `/work/${name}.log` };
}

describe('ProcessDetails queries', () => {
  it('deduplicates project status, isolates process logs, and reuses both after remount', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.startsWith('/api/proc/status')) {
        return new Response(JSON.stringify({
          hasProcfile: true,
          running: true,
          processes: [
            { name: 'api', tree: [{ pid: 10, ppid: 1, command: 'api --serve' }] },
            { name: 'worker', tree: [{ pid: 20, ppid: 1, command: 'worker --serve' }] },
          ],
        }), { status: 200, headers: { 'Content-Type': 'application/json' } });
      }
      const name = new URL(url, 'http://gavel.local').searchParams.get('name');
      return new Response(`${name} ready\n`, { status: 200 });
    });
    vi.stubGlobal('fetch', fetchMock);
    const client = createClient();
    const content = (
      <QueryClientProvider client={client}>
        <ProcExpanded project="gavel" proc={process('api')} />
        <ProcExpanded project="gavel" proc={process('worker')} />
      </QueryClientProvider>
    );

    const first = render(content);
    expect(await screen.findByText('api ready')).toBeTruthy();
    expect(await screen.findByText('worker ready')).toBeTruthy();
    expect(fetchMock.mock.calls.filter(([input]) => String(input).startsWith('/api/proc/status'))).toHaveLength(1);
    expect(fetchMock.mock.calls.filter(([input]) => String(input).startsWith('/api/proc/logs'))).toHaveLength(2);
    first.unmount();

    render(content);
    expect(await screen.findByText('api ready')).toBeTruthy();
    expect(await screen.findByText('worker ready')).toBeTruthy();
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });

  it('does not overlap slow status or log polls', async () => {
    vi.useFakeTimers();
    const fetchMock = vi.fn(() => new Promise<Response>(() => undefined));
    vi.stubGlobal('fetch', fetchMock);

    render(
      <QueryClientProvider client={createClient()}>
        <ProcExpanded project="gavel" proc={process('api')} />
      </QueryClientProvider>,
    );
    await act(async () => undefined);
    expect(fetchMock).toHaveBeenCalledTimes(2);

    await act(async () => vi.advanceTimersByTimeAsync(9_000));
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('keeps process polling disabled while hidden and starts it when visible', async () => {
    Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'hidden' });
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => (
      String(input).startsWith('/api/proc/status')
        ? new Response(JSON.stringify({ hasProcfile: true, running: true, processes: [] }), {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          })
        : new Response('ready', { status: 200 })
    ));
    vi.stubGlobal('fetch', fetchMock);

    render(
      <QueryClientProvider client={createClient()}>
        <ProcExpanded project="gavel" proc={process('api')} />
      </QueryClientProvider>,
    );
    await act(async () => undefined);
    expect(fetchMock).not.toHaveBeenCalled();

    act(() => {
      Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'visible' });
      document.dispatchEvent(new Event('visibilitychange'));
    });
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
  });
});
