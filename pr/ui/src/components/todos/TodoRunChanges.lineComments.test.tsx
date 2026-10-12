import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import type { DiffLineTarget, DiffLineWidget } from '@flanksource/clicky-ui/data';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoCommitFile, TodoEvent, TodoRunCommit } from '../../types';
import { lineComments, type LineCommentAnchor } from './lineComments';
import { TodoRunChanges } from './TodoRunChanges';
import { queryTestWrapper } from './queryTestWrapper';

const LINE: DiffLineTarget = { side: 'new', line: 7, content: 'return nil' };

// The real diff renderer is clicky-ui's; here it only proves the two seams the
// viewer drives: the per-line "+" action and the widgets rendered under a line.
vi.mock('@flanksource/clicky-ui/data', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/clicky-ui/data')>()),
  // Streamdown pulls its own React copy under vitest; the body text is all these tests read.
  Markdown: ({ text }: { text: string }) => <div>{text}</div>,
  GitDiffPanel: ({ payload, onLineAction, lineWidgets }: {
    payload: { diff: string } | null;
    onLineAction?: (target: DiffLineTarget) => void;
    lineWidgets?: DiffLineWidget[];
  }) => (
    <div>
      <div data-testid="diff-panel">{payload?.diff ?? 'loading'}</div>
      {onLineAction && (
        // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test stand-in for the diff gutter button.
        <button type="button" onClick={() => onLineAction(LINE)}>gutter plus</button>
      )}
      {lineWidgets?.map(widget => (
        <div key={widget.key} data-testid="line-widget" data-side={widget.side} data-line={widget.line}>{widget.node}</div>
      ))}
    </div>
  ),
}));

const SETUP = '5e7a0000aaaa1111bbbb2222cccc3333dddd4444';
const HEAD = 'd0b5c966aaaa1111bbbb2222cccc3333dddd4444';
const FIRST = 'aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111';
const ATTEMPT = 'run-7';

const commits: TodoRunCommit[] = [{ sha: FIRST, message: 'feat: first' }, { sha: HEAD, message: 'fix: second' }];
const worktree = { branch: 'shell/1', setup: SETUP, head: HEAD };
const files: TodoCommitFile[] = [
  { path: 'docs/a.md', status: 'added', adds: 4, dels: 0 },
  { path: 'cmd/main.go', status: 'modified', adds: 3, dels: 1 },
];

function json(body: unknown) {
  return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } });
}

function stubChanges() {
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
    const url = new URL(String(input), 'http://localhost');
    if (url.pathname === '/api/todos/commits/files') return json({ hash: url.searchParams.get('hash'), files });
    if (url.pathname === '/api/todos/commits/diff') return json({ diff: `diff:${url.searchParams.get('file')}` });
    throw new Error(`unexpected request ${url.pathname}`);
  }));
}

function comment(id: string, body: string, anchor: Partial<LineCommentAnchor>, resolved = false): TodoEvent[] {
  const created: TodoEvent = {
    id, kind: 'comment', actor: 'reviewer', timestamp: '2026-10-01T10:00:00Z', body,
    payload: { anchor: { path: 'cmd/main.go', side: 'new', line: 7, lineText: 'return nil', base: SETUP, commit: HEAD, ...anchor } },
  };
  return resolved ? [created, { id: `${id}-r`, kind: 'comment_resolved', payload: { commentId: id, resolved: true } }] : [created];
}

function renderChanges(events: TodoEvent[] = []) {
  const onCreate = vi.fn(async () => undefined);
  const onResolve = vi.fn(async () => undefined);
  render(
    <TodoRunChanges
      dir="/repo"
      worktree={worktree}
      commits={commits}
      attemptId={ATTEMPT}
      lineComments={{ comments: lineComments(events), onCreate, onResolve }}
    />,
    { wrapper: queryTestWrapper() },
  );
  return { onCreate, onResolve };
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('TodoRunChanges line comments', () => {
  it('offers no gutter action when the viewer is read-only', async () => {
    stubChanges();
    render(<TodoRunChanges dir="/repo" worktree={worktree} commits={commits} />, { wrapper: queryTestWrapper() });

    await screen.findByText('diff:cmd/main.go');

    expect(screen.queryByRole('button', { name: 'gutter plus' })).toBeNull();
  });

  it('saves a comment from the gutter composer with the all-changes range as its anchor', async () => {
    stubChanges();
    const { onCreate } = renderChanges();
    await screen.findByText('diff:cmd/main.go');

    fireEvent.click(screen.getByRole('button', { name: 'gutter plus' }));
    expect(screen.getByTestId('line-widget').getAttribute('data-line')).toBe('7');
    fireEvent.change(screen.getByRole('textbox', { name: 'Line comment body' }), { target: { value: '  handle the error  ' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(onCreate).toHaveBeenCalledTimes(1));
    expect(onCreate).toHaveBeenCalledWith(
      { path: 'cmd/main.go', side: 'new', line: 7, lineText: 'return nil', base: SETUP, commit: HEAD, branch: 'shell/1', attemptId: ATTEMPT },
      'handle the error',
    );
    await waitFor(() => expect(screen.queryByTestId('diff-line-composer')).toBeNull());
  });

  it('anchors a single-commit range with no base and that commit', async () => {
    stubChanges();
    const { onCreate } = renderChanges();
    await screen.findByText('diff:cmd/main.go');
    fireEvent.change(screen.getByRole('combobox', { name: 'Changes to show' }), { target: { value: FIRST } });
    await screen.findByRole('button', { name: 'gutter plus' });

    fireEvent.click(screen.getByRole('button', { name: 'gutter plus' }));
    fireEvent.change(screen.getByRole('textbox', { name: 'Line comment body' }), { target: { value: 'nit' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(onCreate).toHaveBeenCalled());
    expect(onCreate).toHaveBeenCalledWith(
      { path: 'cmd/main.go', side: 'new', line: 7, lineText: 'return nil', base: undefined, commit: FIRST, branch: 'shell/1', attemptId: ATTEMPT },
      'nit',
    );
  });

  it('keeps the composer open and shows the failure when saving rejects', async () => {
    stubChanges();
    const { onCreate } = renderChanges();
    onCreate.mockRejectedValueOnce(new Error('The comment was not saved'));
    await screen.findByText('diff:cmd/main.go');

    fireEvent.click(screen.getByRole('button', { name: 'gutter plus' }));
    fireEvent.change(screen.getByRole('textbox', { name: 'Line comment body' }), { target: { value: 'oops' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect((await screen.findByRole('alert')).textContent).toBe('The comment was not saved');
    expect(screen.getByTestId('diff-line-composer')).toBeTruthy();
  });

  it('renders under their line only the comments matching the selected range and file', async () => {
    stubChanges();
    renderChanges([
      ...comment('here', 'matches', {}),
      ...comment('other-file', 'wrong file', { path: 'docs/a.md' }),
      ...comment('other-range', 'wrong range', { base: undefined, commit: FIRST }),
      ...comment('other-commit', 'wrong head', { commit: FIRST }),
    ]);
    await screen.findByText('diff:cmd/main.go');

    expect(screen.getAllByTestId('line-widget').map(widget => widget.textContent)).toEqual([expect.stringContaining('matches')]);
    expect(screen.queryByText('wrong file')).toBeNull();

    fireEvent.change(screen.getByRole('combobox', { name: 'Changes to show' }), { target: { value: FIRST } });
    await screen.findByText('wrong range');
    expect(screen.getAllByTestId('line-widget').map(widget => widget.textContent)).toEqual([expect.stringContaining('wrong range')]);
  });

  it('shows no widgets and no gutter action when a directory is selected', async () => {
    stubChanges();
    renderChanges(comment('here', 'matches', {}));
    await screen.findByText('matches');

    fireEvent.click(screen.getByText('docs'));
    await screen.findByText('diff:docs');

    expect(screen.queryByRole('button', { name: 'gutter plus' })).toBeNull();
    expect(screen.queryByTestId('line-widget')).toBeNull();
  });

  it('resolves an open comment and reopens a resolved one through the card button', async () => {
    stubChanges();
    const { onResolve } = renderChanges([
      ...comment('open', 'still open', { line: 7 }),
      ...comment('done', 'was fixed', { line: 9 }, true),
    ]);
    await screen.findByText('still open');

    const [open, done] = screen.getAllByTestId('diff-line-comment');
    expect(open!.getAttribute('data-resolved')).toBe('false');
    expect(done!.getAttribute('data-resolved')).toBe('true');
    fireEvent.click(within(open!).getByRole('button', { name: 'Resolve' }));
    await waitFor(() => expect(onResolve).toHaveBeenCalledWith('open', true));

    fireEvent.click(within(done!).getByRole('button', { name: 'Reopen' }));
    await waitFor(() => expect(onResolve).toHaveBeenCalledWith('done', false));
  });

  it('collapses a resolved comment to a one-line summary until expanded', async () => {
    stubChanges();
    renderChanges(comment('done', 'first line\nsecond line', {}, true));
    const card = await screen.findByTestId('diff-line-comment');

    expect(within(card).getByText('first line')).toBeTruthy();
    expect(within(card).queryByText('second line')).toBeNull();

    fireEvent.click(within(card).getByRole('button', { name: 'Expand comment' }));

    expect(within(card).getByText(/second line/)).toBeTruthy();
  });
});
