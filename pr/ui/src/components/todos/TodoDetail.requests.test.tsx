import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { TodoDetail } from './TodoDetail';

afterEach(() => vi.unstubAllGlobals());

it('does not load run configuration before a TODO is resolved', async () => {
  const fetch = vi.fn();
  vi.stubGlobal('fetch', fetch);
  render(<QueryClientProvider client={new QueryClient()}>
    <TodoDetail todo={null} loading={false} dir="" onChanged={vi.fn()} onDeleted={vi.fn()} view={{}} onViewChange={vi.fn()} />
  </QueryClientProvider>);
  await screen.findByText('Select a todo');
  expect(fetch).not.toHaveBeenCalled();
});
