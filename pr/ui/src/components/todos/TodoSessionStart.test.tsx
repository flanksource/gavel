import type React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { TodoSessionStart } from './TodoSessionStart';
import type { TodoItem } from '../../types';

const RESOLVED_OPTIONS = { driver: 'claude-headless', backend: 'claude-agent', model: 'claude-sonnet-5', effort: 'medium' };

vi.mock('./run', () => ({
  useTodoRunContext: () => ({ context: { backends: [], efforts: [], defaultBackend: '', tools: [] }, loading: false, error: '' }),
  TodoRunContextError: ({ error }: { error: string }) => error ? <div role="alert">{error}</div> : null,
  loadLastTodoRunOptions: () => ({ ...RESOLVED_OPTIONS }),
  todoRunButtonPresentation: () => ({ provider: undefined, model: 'sonnet-5', effort: 'medium' }),
  todoRunModeLabel: () => 'Agent',
  useTodoRunPreview: () => ({
    isPending: false,
    mutate: (
      { body, signal }: { body: Record<string, unknown>; signal?: AbortSignal },
      callbacks: { onSuccess: (data: { prompt?: string }) => void; onError: (error: Error) => void },
    ) => {
      fetch('/api/todos/run/preview?dir=%2Frepo', {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body), signal,
      }).then(async response => {
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || 'Preview failed');
        callbacks.onSuccess(data);
      }).catch(callbacks.onError);
    },
  }),
  TodoRunEffortBadge: ({ effort }: { effort?: string }) => <span data-testid="effort-badge">{effort}</span>,
  TodoRunActionButton: ({ action, onRun }: { action: string; onRun?: (options?: unknown) => void }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test stub for the run button.
    <button type="button" onClick={() => onRun?.({ action })}>
      {action}
    </button>
  ),
}));

vi.mock('@flanksource/clicky-ui/data', () => ({
  Markdown: ({ text }: { text: string }) => <div data-testid="markdown">{text}</div>,
}));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => {
  const Icon = (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />;
  return {
    ...(await importOriginal<object>()),
    UiHubot: Icon,
    UiRobotAi: Icon,
  };
});

vi.mock('../../icons/Spinner', () => ({
  Spinner: () => <span data-testid="spinner" />,
}));

const todo: TodoItem = { ref: 'todo:abc', title: 'Wire the thing', status: 'pending', priority: 'medium' };

afterEach(() => {
  vi.restoreAllMocks();
});

function stubPreviewFetch(prompt: string) {
  const fetchMock = vi.fn(async (_url: RequestInfo | URL, _init?: RequestInit) => ({
    ok: true,
    json: async () => ({ prompt, mode: 'inline', agent: 'claude', count: 1 }),
  }));
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

describe('TodoSessionStart', () => {
  it('renders the model, runtime, and effort a run would use', async () => {
    stubPreviewFetch('## Implement thing');
    render(<TodoSessionStart dir="/repo" todo={todo} onRun={vi.fn()} />);

    expect(screen.getByText('sonnet-5')).toBeTruthy();
    expect(screen.getByText('Agent')).toBeTruthy();
    expect(screen.getByTestId('effort-badge').textContent).toBe('medium');
    // Flush the async preview so its state update settles inside act().
    await screen.findByText('## Implement thing');
  });

  it('fetches and renders the exact prompt the run would send', async () => {
    const fetchMock = stubPreviewFetch('## Implement thing');
    render(<TodoSessionStart dir="/repo" todo={todo} onRun={vi.fn()} />);

    expect(await screen.findByText('## Implement thing')).toBeTruthy();
    const [url, init] = fetchMock.mock.calls[0]!;
    expect(String(url)).toContain('/api/todos/run/preview');
    expect(JSON.parse(init!.body as string)).toMatchObject({ ref: 'todo:abc', runMode: 'run' });
  });

  it('invokes onRun when the Run action is clicked', async () => {
    stubPreviewFetch('## Implement thing');
    const onRun = vi.fn();
    render(<TodoSessionStart dir="/repo" todo={todo} onRun={onRun} />);

    fireEvent.click(screen.getByRole('button', { name: 'run' }));
    await waitFor(() => expect(onRun).toHaveBeenCalledWith({ action: 'run' }));
  });

  it('surfaces a preview error but still offers the run actions', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => ({ ok: false, json: async () => ({ error: 'boom' }) })),
    );
    render(<TodoSessionStart dir="/repo" todo={todo} onRun={vi.fn()} />);

    expect(await screen.findByText('boom')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'run' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'plan' })).toBeTruthy();
  });
});
