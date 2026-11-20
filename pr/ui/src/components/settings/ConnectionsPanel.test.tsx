// Tests that the connections panel reads its form schema and dashboard from the server's
// connection catalog.
import { render, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ConnectionsPanel } from './ConnectionsPanel';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('ConnectionsPanel', () => {
  it('reads the connection catalog the server serves under /api/catalog', async () => {
    const fetchMock = vi.fn().mockImplementation(async () => new Response(JSON.stringify({ connections: [] }), {
      status: 200, headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <ConnectionsPanel />
      </QueryClientProvider>,
    );

    await waitFor(() => {
      const urls = fetchMock.mock.calls.map(([url]) => String(url instanceof Request ? url.url : url));
      expect(urls.some(url => url.includes('/api/catalog/connections/dashboard'))).toBe(true);
      expect(urls.some(url => url.endsWith('/api/catalog/connection') || url.includes('/api/catalog/connection?'))).toBe(true);
    });
  });
});
