import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import type { DiffLineTarget, DiffLineWidget } from '@flanksource/clicky-ui/data';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoEvent, TodoSessionAttempt } from '../../types';
import { TodoRunBranch } from './TodoRunBranch';
import { queryTestWrapper } from './queryTestWrapper';

vi.mock('@flanksource/clicky-ui/data', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/clicky-ui/data')>()),
  Markdown: ({ text }: { text: string }) => <div>{text}</div>,
  GitDiffPanel: ({ onLineAction, lineWidgets }: { onLineAction?: (target: DiffLineTarget) => void; lineWidgets?: DiffLineWidget[] }) => (
    <div>
      {onLineAction && (
        // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test stand-in for the diff gutter button.
        <button type="button" onClick={() => onLineAction({ side: 'old', line: 3, content: 'old line' })}>gutter plus</button>
      )}
      {lineWidgets?.map(widget => <div key={widget.key}>{widget.node}</div>)}
    </div>
  ),
}));

const SETUP = '5e7a0000aaaa1111bbbb2222cccc3333dddd4444';
const HEAD = 'd0b5c966aaaa1111bbbb2222cccc3333dddd4444';

const attempt: TodoSessionAttempt = {
  promptRunId: 'run-2',
  ordinal: 2,
  step: 'run',
  requested: {},
  resolved: {},
  status: 'completed',
  processActive: false,
  state: 'succeeded',
  phase: 'finished',
  queuedAt: '2026-09-14T12:00:00Z',
  admissionSessionId: 'admission-2',
  createdAt: '2026-09-14T12:00:00Z',
  updatedAt: '2026-09-14T12:01:00Z',
  verification: null,
  workspace: { worktree: { branch: 'shell/289107d9', setup: SETUP, head: HEAD } },
};

const openComment: TodoEvent = {
  id: 'c1', kind: 'comment', actor: 'reviewer', timestamp: '2026-10-01T10:00:00Z', body: 'rename this',
  payload: { anchor: { path: 'main.go', side: 'old', line: 3, lineText: 'old line', base: SETUP, commit: HEAD } },
};

function stubFetch() {
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
    const url = new URL(String(input), 'http://localhost');
    const body = url.pathname === '/api/todos/session/detail'
      ? { attempts: [attempt] }
      : url.pathname === '/api/todos/commits/files'
        ? { hash: HEAD, files: [{ path: 'main.go', status: 'modified', adds: 1, dels: 1 }] }
        : { diff: 'diff' };
    return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } });
  }));
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('TodoRunBranch line comments', () => {
  it('patches a new line comment with the full anchor of the run branch range', async () => {
    stubFetch();
    const onPatch = vi.fn(async () => true);
    render(<TodoRunBranch dir="/repo" todoRef="todo-1" events={[]} onPatch={onPatch} />, { wrapper: queryTestWrapper() });

    fireEvent.click(await screen.findByRole('button', { name: 'gutter plus' }));
    fireEvent.change(screen.getByRole('textbox', { name: 'Line comment body' }), { target: { value: 'use the new name' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(onPatch).toHaveBeenCalledTimes(1));
    expect(onPatch).toHaveBeenCalledWith({
      lineComment: {
        anchor: { path: 'main.go', side: 'old', line: 3, lineText: 'old line', base: SETUP, commit: HEAD, branch: 'shell/289107d9', attemptId: 'run-2' },
        body: 'use the new name',
      },
    });
  });

  it('patches a resolution for the comment id', async () => {
    stubFetch();
    const onPatch = vi.fn(async () => true);
    render(<TodoRunBranch dir="/repo" todoRef="todo-1" events={[openComment]} onPatch={onPatch} />, { wrapper: queryTestWrapper() });

    const card = await screen.findByTestId('diff-line-comment');
    fireEvent.click(within(card).getByRole('button', { name: 'Resolve' }));

    await waitFor(() => expect(onPatch).toHaveBeenCalledWith({ resolveComment: { id: 'c1', resolved: true } }));
  });

  it('patches a reopen once an event resolved the comment', async () => {
    stubFetch();
    const onPatch = vi.fn(async () => true);
    const resolved: TodoEvent = { id: 'r1', kind: 'comment_resolved', payload: { commentId: 'c1', resolved: true } };
    render(<TodoRunBranch dir="/repo" todoRef="todo-1" events={[openComment, resolved]} onPatch={onPatch} />, { wrapper: queryTestWrapper() });

    const card = await screen.findByTestId('diff-line-comment');
    fireEvent.click(within(card).getByRole('button', { name: 'Reopen' }));

    await waitFor(() => expect(onPatch).toHaveBeenCalledWith({ resolveComment: { id: 'c1', resolved: false } }));
  });

  it('reports a rejected patch and keeps the draft open', async () => {
    stubFetch();
    const onPatch = vi.fn(async () => false);
    render(<TodoRunBranch dir="/repo" todoRef="todo-1" events={[]} onPatch={onPatch} />, { wrapper: queryTestWrapper() });

    fireEvent.click(await screen.findByRole('button', { name: 'gutter plus' }));
    fireEvent.change(screen.getByRole('textbox', { name: 'Line comment body' }), { target: { value: 'draft' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect((await screen.findByRole('alert')).textContent).toBe('The comment was not saved');
    expect((screen.getByRole('textbox', { name: 'Line comment body' }) as HTMLTextAreaElement).value).toBe('draft');
  });

  it('stays read-only without a patch handler', async () => {
    stubFetch();
    render(<TodoRunBranch dir="/repo" todoRef="todo-1" />, { wrapper: queryTestWrapper() });

    await screen.findByText('shell/289107d9');
    expect((await screen.findAllByText('main.go')).length).toBeGreaterThan(0);

    expect(screen.queryByRole('button', { name: 'gutter plus' })).toBeNull();
  });
});
