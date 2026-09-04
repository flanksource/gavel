import type React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem } from '../../types';
import { TodoPlan } from './TodoPlan';
import { queryTestWrapper } from './queryTestWrapper';
import { todoQueryKeys } from './todoQueries';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>{children}</button>
  ),
}));

vi.mock('@flanksource/clicky-ui/mdx-editor', () => ({
  MdxEditorField: ({ value, onChange }: { value: string; onChange: (value: string) => void }) => (
    <textarea aria-label="Plan content" value={value} onChange={event => onChange(event.currentTarget.value)} />
  ),
}));

const todo: TodoItem = {
  ref: 'native-id',
  version: 3,
  title: 'Native issue',
  status: 'pending',
  priority: 'medium',
};

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('TodoPlan', () => {
  it('renders the actionable database error when plan loading fails', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({
      ok: false,
      status: 503,
      json: async () => ({ error: 'native TODO storage requires PostgreSQL' }),
    }) as Response));

    render(<TodoPlan dir="/repo" todo={todo} active />, { wrapper: queryTestWrapper() });
    expect((await screen.findByRole('alert')).textContent).toContain('native TODO storage requires PostgreSQL');
    expect(screen.queryByText('No plan yet. Run this todo in Plan mode to produce one.')).toBeNull();
  });

  it('writes a saved revision into the exact plan cache and projects its todo', async () => {
    const updatedTodo = { ...todo, version: 4, status: 'review' as const };
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const onChanged = vi.fn();
    vi.stubGlobal('fetch', vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        return {
          ok: true,
          json: async () => ({ found: true, content: 'updated plan', version: 4, todo: updatedTodo }),
        } as Response;
      }
      return { ok: true, json: async () => ({ found: true, content: 'original plan', version: 3 }) } as Response;
    }));

    render(<TodoPlan dir="/repo" todo={todo} active onChanged={onChanged} />, {
      wrapper: ({ children }) => <QueryClientProvider client={client}>{children}</QueryClientProvider>,
    });
    fireEvent.change(await screen.findByRole('textbox', { name: 'Plan content' }), { target: { value: 'updated plan' } });
    fireEvent.click(screen.getByRole('button', { name: /save/i }));

    await waitFor(() => expect(onChanged).toHaveBeenCalledWith(updatedTodo));
    expect(client.getQueryData<{ content?: string }>(todoQueryKeys.plan('/repo', 'native-id'))?.content).toBe('updated plan');
  });

  it('surfaces plan save failures with action context', async () => {
    vi.stubGlobal('fetch', vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        return {
          ok: false,
          status: 409,
          json: async () => ({ error: 'plan version changed' }),
          text: async () => JSON.stringify({ error: 'plan version changed' }),
        } as Response;
      }
      return { ok: true, json: async () => ({ found: true, content: 'original plan', version: 3 }) } as Response;
    }));
    render(<TodoPlan dir="/repo" todo={todo} active />, { wrapper: queryTestWrapper() });
    fireEvent.change(await screen.findByRole('textbox', { name: 'Plan content' }), { target: { value: 'updated plan' } });
    fireEvent.click(screen.getByRole('button', { name: /save/i }));

    expect(await screen.findByText(/plan save failed.*plan version changed/i)).toBeTruthy();
  });
});
