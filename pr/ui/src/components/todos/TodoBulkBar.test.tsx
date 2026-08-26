import type React from 'react';
import { act, fireEvent, render, renderHook, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { BulkTodoResponse } from './todoMutations';
import { queryTestWrapper } from './queryTestWrapper';
import { TodoBulkBar, TodoBulkToggle } from './TodoBulkBar';
import { useTodoSelection, type TodoSelection } from './todoSelection';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({
    children,
    loading: _loading,
    ...props
  }: React.ButtonHTMLAttributes<HTMLButtonElement> & { loading?: boolean }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>{children}</button>
  ),
  // The real DropdownMenu portals and measures; render the trigger and the open
  // menu inline so a menu item is directly clickable.
  DropdownMenu: ({ trigger, children }: {
    trigger: React.ReactNode;
    children: (close: () => void) => React.ReactNode;
  }) => (
    <div>
      {trigger}
      {children(() => {})}
    </div>
  ),
}));

const alpha = { dir: '/repos/alpha', ref: 'todo-1' };
const beta = { dir: '/repos/beta', ref: 'todo-2' };

function bulkResponse(overrides: Partial<BulkTodoResponse> = {}): BulkTodoResponse {
  return { updated: 2, failed: 0, results: [], ...overrides };
}

function selectionWith(...todos: { dir: string; ref: string }[]) {
  const { result } = renderHook(() => useTodoSelection());
  act(() => {
    result.current.setBulkMode(true);
    result.current.setGroupSelected(todos, true);
  });
  return result;
}

function renderBar(selection: TodoSelection, onApplied?: () => void) {
  return render(<TodoBulkBar selection={selection} onApplied={onApplied} />, { wrapper: queryTestWrapper() });
}

describe('TodoBulkBar', () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn(async () => new Response(JSON.stringify(bulkResponse()), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('renders nothing until bulk-edit mode is on', () => {
    const { result } = renderHook(() => useTodoSelection());
    const { container } = renderBar(result.current);
    expect(container.innerHTML).toBe('');
  });

  it('reports how many todos are checked', () => {
    renderBar(selectionWith(alpha, beta).current);
    expect(screen.getByText('2 selected')).toBeTruthy();
  });

  it('posts every checked todo to the bulk endpoint with the chosen priority', async () => {
    renderBar(selectionWith(alpha, beta).current);

    fireEvent.click(screen.getByText('high'));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe('/api/todos/bulk');
    expect(JSON.parse(init.body as string)).toEqual({
      items: expect.arrayContaining([alpha, beta]),
      priority: 'high',
    });
  });

  it('offers only the statuses the server lets a caller assign', () => {
    renderBar(selectionWith(alpha).current);
    for (const status of ['draft', 'pending', 'verified', 'completed', 'skipped']) {
      expect(screen.getByText(status)).toBeTruthy();
    }
    for (const projected of ['in progress', 'review', 'ask', 'failed', 'unverified']) {
      expect(screen.queryByText(projected)).toBeNull();
    }
  });

  it('sends the chosen status rather than a priority', async () => {
    renderBar(selectionWith(alpha).current);

    fireEvent.click(screen.getByText('completed'));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(JSON.parse((fetchMock.mock.calls[0][1] as RequestInit).body as string))
      .toEqual({ items: [alpha], status: 'completed' });
  });

  it('names the todos the server refused, alongside the ones it updated', async () => {
    fetchMock.mockResolvedValueOnce(new Response(JSON.stringify(bulkResponse({
      updated: 1,
      failed: 1,
      results: [
        { ...alpha, todo: undefined, error: 'todo "todo-1" not found' },
        { ...beta, todo: undefined },
      ],
    })), { status: 200, headers: { 'Content-Type': 'application/json' } }));

    renderBar(selectionWith(alpha, beta).current);
    fireEvent.click(screen.getByText('low'));

    const status = await screen.findByRole('status');
    expect(status.textContent).toContain('Updated 1 todo');
    expect(status.textContent).toContain('1 failed: todo-1 (todo "todo-1" not found)');
  });

  it('surfaces a rejected request instead of reporting a silent success', async () => {
    fetchMock.mockResolvedValueOnce(new Response(JSON.stringify({ error: 'unknown priority "urgent"' }), {
      status: 400,
      headers: { 'Content-Type': 'application/json' },
    }));

    renderBar(selectionWith(alpha).current);
    fireEvent.click(screen.getByText('medium'));

    expect((await screen.findByRole('alert')).textContent).toContain('unknown priority "urgent"');
    expect(screen.queryByRole('status')).toBeNull();
  });

  it('keeps the selection after an apply so a second field can be set on the same batch', async () => {
    const selection = selectionWith(alpha, beta);
    const { rerender } = renderBar(selection.current);

    fireEvent.click(screen.getByText('high'));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));

    rerender(<TodoBulkBar selection={selection.current} />);
    expect(screen.getByText('2 selected')).toBeTruthy();
  });

  it('clears the checked set without leaving bulk-edit mode', () => {
    const selection = selectionWith(alpha, beta);
    const { rerender } = renderBar(selection.current);

    fireEvent.click(screen.getByText('Clear'));

    rerender(<TodoBulkBar selection={selection.current} />);
    expect(screen.getByText('0 selected')).toBeTruthy();
    expect(selection.current.bulkMode).toBe(true);
  });

  it('drops the selection when bulk-edit mode is left', () => {
    const selection = selectionWith(alpha, beta);
    const { rerender } = renderBar(selection.current);

    fireEvent.click(screen.getByLabelText('Exit bulk edit'));

    rerender(<TodoBulkBar selection={selection.current} />);
    expect(selection.current.bulkMode).toBe(false);
    expect(selection.current.targets).toEqual([]);
  });

  it('cannot apply an edit with nothing checked', () => {
    const { result } = renderHook(() => useTodoSelection());
    act(() => result.current.setBulkMode(true));
    renderBar(result.current);

    fireEvent.click(screen.getByText('high'));

    expect(fetchMock).not.toHaveBeenCalled();
  });
});

describe('TodoBulkToggle', () => {
  it('advertises whether bulk-edit mode is on', () => {
    const { result } = renderHook(() => useTodoSelection());
    const { rerender } = render(<TodoBulkToggle selection={result.current} />);
    expect(screen.getByRole('button').getAttribute('aria-pressed')).toBe('false');

    fireEvent.click(screen.getByRole('button'));

    rerender(<TodoBulkToggle selection={result.current} />);
    expect(screen.getByRole('button').getAttribute('aria-pressed')).toBe('true');
  });
});
