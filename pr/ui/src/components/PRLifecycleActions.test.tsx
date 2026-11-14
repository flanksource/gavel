import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { PRLifecycleActions } from './PRLifecycleActions';
import type { PRItem } from '../types';

vi.mock('@flanksource/clicky-ui/components', () => ({ Button: ({ children, ...props }: { children: ReactNode }) => <button {...props}>{children}</button> }));
afterEach(() => vi.unstubAllGlobals());

it('loads only when opened and carries the preview identity into manual kickoff', async () => {
  const calls: Array<{ url: string; body?: Record<string, unknown> }> = [];
  vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
    const body = init?.body ? JSON.parse(String(init.body)) as Record<string, unknown> : undefined;
    calls.push({ url, body });
    if (body?.operation === 'preview') return Response.json({ spec: { prompt: 'Inspect the diff' }, specHash: 'spec-a' });
    return Response.json({ headSha: 'head-a', definition: 'definition-a', actions: [{ step: { name: 'review' }, event: { id: 'update:head-a', type: 'pr.update', body: '', author: '' }, reason: 'true' }], runs: [] });
  }));
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(<QueryClientProvider client={client}><PRLifecycleActions pr={{ repo: 'acme/cli', number: 7 } as PRItem} projects={[{ name: 'cli', dir: '/work/cli', repos: ['acme/cli'] }]} /></QueryClientProvider>);
  expect(calls).toHaveLength(0);
  fireEvent.click(screen.getByText('Lifecycle actions'));
  fireEvent.click(await screen.findByRole('button', { name: 'Preview review' }));
  fireEvent.click(await screen.findByRole('button', { name: 'Start review' }));
  await waitFor(() => expect(calls.some(call => call.body?.operation === 'start')).toBe(true));
  expect(calls.find(call => call.body?.operation === 'start')?.body).toEqual({ operation: 'start', headSha: 'head-a', definition: 'definition-a', step: 'review', eventId: 'update:head-a', specHash: 'spec-a' });
});
