import type React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem } from '../../types';
import { AcceptanceCriteria } from './AcceptanceCriteria';
import { emptyCounts } from './format';
import { queryTestWrapper } from './queryTestWrapper';
import { todoQueryKeys } from './todoQueries';
import { workspaceTodoBatchKeys } from './workspaceTodoQueries';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>{children}</button>
  ),
}));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => {
  const Icon = (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />;
  return {
    ...(await importOriginal<object>()),
    UiEdit: Icon,
    UiListDashes: Icon,
    UiTrash: Icon,
  };
});

const todo: TodoItem = {
  ref: 'todo-criteria',
  title: 'Preserve rollback',
  status: 'pending',
  priority: 'medium',
  criteria: [{ text: 'Focused tests pass', done: false }],
};

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('AcceptanceCriteria', () => {
  it('rolls an optimistic toggle back and surfaces contextual server errors', async () => {
    let resolveResponse!: (response: Response) => void;
    const fetchMock = vi.fn(() => new Promise<Response>(resolve => {
      resolveResponse = resolve;
    }));
    vi.stubGlobal('fetch', fetchMock);
    const onChanged = vi.fn();
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const list = { dir: '/repo', counts: { ...emptyCounts, total: 1, open: 1, pending: 1 }, items: [todo] };
    client.setQueryData(todoQueryKeys.item('/repo', todo.ref), todo);
    client.setQueryData(todoQueryKeys.list('/repo'), list);
    client.setQueryData(workspaceTodoBatchKeys.list(['/repo']), {
      byDir: { '/repo': list },
      errorsByDir: {},
      error: '',
    });
    render(<AcceptanceCriteria dir="/repo" todo={todo} onChanged={onChanged} />, {
      wrapper: ({ children }) => <QueryClientProvider client={client}>{children}</QueryClientProvider>,
    });
    const checkbox = screen.getByRole('checkbox', { name: 'Mark done' }) as HTMLInputElement;

    fireEvent.click(checkbox);
    expect(checkbox.checked).toBe(true);
    expect(client.getQueryData<TodoItem>(todoQueryKeys.item('/repo', todo.ref))?.criteria?.[0]?.done).toBe(true);
    expect(client.getQueryData<{ byDir: Record<string, { items: TodoItem[] }> }>(
      workspaceTodoBatchKeys.list(['/repo']),
    )?.byDir['/repo'].items[0].criteria?.[0]?.done).toBe(true);
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    await act(async () => {
      resolveResponse({
        ok: false,
        status: 409,
        json: async () => ({ error: 'criteria version changed' }),
        text: async () => JSON.stringify({ error: 'criteria version changed' }),
      } as Response);
    });

    await waitFor(() => expect(checkbox.checked).toBe(false));
    expect(client.getQueryData<TodoItem>(todoQueryKeys.item('/repo', todo.ref))?.criteria?.[0]?.done).toBe(false);
    expect(client.getQueryData<{ byDir: Record<string, { items: TodoItem[] }> }>(
      workspaceTodoBatchKeys.list(['/repo']),
    )?.byDir['/repo'].items[0].criteria?.[0]?.done).toBe(false);
    expect(screen.getByText(/acceptance criteria update failed.*criteria version changed/i)).toBeTruthy();
    expect(onChanged).not.toHaveBeenCalled();
  });

  it('adopts the authoritative criteria returned by the mutation', async () => {
    const updated = { ...todo, criteria: [{ text: 'Focused tests pass', done: true }] };
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => updated }) as Response));
    const onChanged = vi.fn();
    render(<AcceptanceCriteria dir="/repo" todo={todo} onChanged={onChanged} />, {
      wrapper: queryTestWrapper(),
    });

    fireEvent.click(screen.getByRole('checkbox', { name: 'Mark done' }));

    await waitFor(() => expect(onChanged).toHaveBeenCalledWith(updated));
    expect((screen.getByRole('checkbox', { name: 'Mark not done' }) as HTMLInputElement).checked).toBe(true);
  });
});
