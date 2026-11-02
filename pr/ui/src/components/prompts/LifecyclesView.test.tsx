import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { LifecyclesView } from './LifecyclesView';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, loading: _loading, ...props }: { children: ReactNode; loading?: boolean }) => <button {...props}>{children}</button>,
  Tabs: () => null,
  JsonSchemaForm: () => <div>Lifecycle fields</div>,
}));

afterEach(() => vi.unstubAllGlobals());

it('requires a fresh preview after reordering and saves only the scoped override', async () => {
  const steps = ['plan', 'verify', 'run'].map(name => ({ name, prompt: `todos.${name}` }));
  const view = { path: '/work/acme/.gavel.yaml', revision: 'original', source: {}, inherited: { steps }, effective: { steps } };
  const calls: RequestInit[] = [];
  vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
    if (url === '/api/settings/schema') return Response.json({ properties: { todos: { properties: { lifecycle: { type: 'object' } } } } });
    if (init?.method !== 'GET') calls.push(init ?? {});
    return Response.json(view);
  }));
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(<QueryClientProvider client={client}><LifecyclesView scopeQuery="project=acme" kind="todos" onSelect={vi.fn()} /></QueryClientProvider>);
  const save = await screen.findByRole('button', { name: 'Save lifecycle' });
  expect(save.hasAttribute('disabled')).toBe(true);
  expect(screen.getByRole('button', { name: 'Disable verify' }).hasAttribute('disabled')).toBe(true);
  fireEvent.click(screen.getByRole('button', { name: 'Edit plan' }));
  expect(screen.getByRole('button', { name: 'Move plan up' }).hasAttribute('disabled')).toBe(true);
  fireEvent.click(screen.getByRole('button', { name: 'Reset plan' }));
  fireEvent.click(screen.getByRole('button', { name: 'Move run up' }));
  fireEvent.click(screen.getByRole('button', { name: 'Preview lifecycle' }));
  await waitFor(() => expect(save.hasAttribute('disabled')).toBe(false));
  fireEvent.click(save);
  await waitFor(() => expect(calls.some(call => call.method === 'PUT')).toBe(true));
  const payload = JSON.parse(String(calls.find(call => call.method === 'PUT')?.body));
  expect(payload).toEqual({ revision: 'original', source: { steps: [], order: ['plan', 'run', 'verify'] } });
  fireEvent.click(screen.getByRole('button', { name: 'Disable plan' }));
  expect(save.hasAttribute('disabled')).toBe(true);
});
