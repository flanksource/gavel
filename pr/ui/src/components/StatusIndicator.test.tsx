import type React from 'react';
import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, render } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { StatusIndicator } from './StatusIndicator';

vi.hoisted(() => {
  Object.assign(globalThis, { __GAVEL_UI_VERSION__: 'test', __GAVEL_UI_COMMIT__: 'test' });
});

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => <button {...props}>{children}</button>,
  DropdownMenu: ({ trigger, children }: { trigger: ReactNode; children: () => ReactNode }) => <div>{trigger}{children()}</div>,
}));

vi.mock('@flanksource/clicky-ui/data', () => ({ Version: () => <span>version</span> }));

vi.mock('@flanksource/clicky-ui/icons', () => ({
  UiWarningTriangle: () => <span />,
  UiPause: () => <span />,
  UiPlay: () => <span />,
  UiSync: () => <span />,
  UiRefresh: () => <span />,
  UiLoader: () => <span />,
}));

vi.mock('../useNow', () => ({ useNow: () => 0 }));
vi.mock('./RelativeTime', () => ({ RelativeTime: () => <span>now</span> }));

const health = {
  overall: 'ok',
  database: { severity: 'ok', message: 'ready' },
  github: { severity: 'ok', message: 'ready' },
  checkedAt: '2026-08-02T08:00:00Z',
};

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'visible' });
});

function createClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: Infinity } } });
}

function indicator() {
  return <StatusIndicator fetchedAt="" nextFetchIn={30} paused={false} onRefresh={() => undefined} onPause={() => undefined} />;
}

function deferredResponse() {
  let resolve!: (value: Response) => void;
  const promise = new Promise<Response>(complete => { resolve = complete; });
  return { promise, resolve };
}

describe('StatusIndicator health query', () => {
  it('deduplicates concurrent consumers', async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify(health), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);
    const client = createClient();

    render(<QueryClientProvider client={client}>{indicator()}{indicator()}</QueryClientProvider>);
    await act(async () => undefined);

    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it('does not overlap a slow health poll', async () => {
    vi.useFakeTimers();
    const pending = deferredResponse();
    const fetchMock = vi.fn(() => pending.promise);
    vi.stubGlobal('fetch', fetchMock);

    render(<QueryClientProvider client={createClient()}>{indicator()}</QueryClientProvider>);
    await act(async () => undefined);
    expect(fetchMock).toHaveBeenCalledTimes(1);

    await act(async () => vi.advanceTimersByTimeAsync(30_000));
    expect(fetchMock).toHaveBeenCalledTimes(1);

    pending.resolve(new Response(JSON.stringify(health), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));
    await act(async () => undefined);
  });

  it('pauses polling while hidden and refreshes when visible again', async () => {
    vi.useFakeTimers();
    const fetchMock = vi.fn(async () => new Response(JSON.stringify(health), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);

    render(<QueryClientProvider client={createClient()}>{indicator()}</QueryClientProvider>);
    await act(async () => undefined);
    expect(fetchMock).toHaveBeenCalledTimes(1);

    act(() => {
      Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'hidden' });
      document.dispatchEvent(new Event('visibilitychange'));
    });
    await act(async () => vi.advanceTimersByTimeAsync(30_000));
    expect(fetchMock).toHaveBeenCalledTimes(1);

    act(() => {
      Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'visible' });
      document.dispatchEvent(new Event('visibilitychange'));
    });
    await act(async () => undefined);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
});
