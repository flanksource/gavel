import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { OrgChooser } from './OrgChooser';
import type { SearchConfig } from '../types';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('OrgChooser', () => {
  it('loads the full organization list once and reuses it when reopened', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => [{ login: 'acme', avatarUrl: '' }],
    });
    vi.stubGlobal('fetch', fetchMock);
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const config = { all: false, org: '', repos: [], ignoredOrgs: [] } as SearchConfig;
    render(
      <QueryClientProvider client={queryClient}>
        <OrgChooser config={config} onChange={vi.fn()} />
      </QueryClientProvider>,
    );

    const trigger = screen.getByTitle('Switch GitHub org / scope');
    fireEvent.click(trigger);
    expect(await screen.findByText('acme')).toBeTruthy();
    fireEvent.click(trigger);
    fireEvent.click(trigger);
    await waitFor(() => expect(screen.getByText('acme')).toBeTruthy());

    expect(fetchMock).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenCalledWith('/api/orgs?include-ignored=1', {
      signal: expect.any(AbortSignal),
    });
  });
});
