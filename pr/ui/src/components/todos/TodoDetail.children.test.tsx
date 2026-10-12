import type React from 'react';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem } from '../../types';
import { StatefulTodoDetail } from './todoDetailTestHarness';
import type { TodoDetailProps } from './TodoDetail';
import { queryTestWrapper } from './queryTestWrapper';

const update = vi.hoisted(() => ({ mutateAsync: vi.fn() }));

vi.mock('./TodoTimeline', () => ({ TodoTimeline: () => null }));
vi.mock('./TodoRunBranch', () => ({ TodoRunBranch: () => <div data-testid="run-branch" /> }));
vi.mock('./TodoSession', () => ({ TodoSession: () => null }));
vi.mock('./TodoPlan', () => ({ TodoPlan: () => null }));
vi.mock('./TodoVerification', () => ({ TodoVerification: () => null }));
vi.mock('./TodoCompose', () => ({ TodoTitleEditor: () => null, TodoBodyEditor: () => null, TodoCommentBox: () => <div data-testid="comment-box" /> }));
vi.mock('./planActions', () => ({ TodoReviewBanner: () => null }));
vi.mock('./TodoPhaseButton', () => ({ TodoPhaseTicks: () => null, TodoPhaseButton: () => null }));
vi.mock('./TodoRunAdvancedDialog', () => ({ TodoRunAdvancedDialog: () => null }));
vi.mock('./CreateTodoDialog', () => ({ CreateTodoDialog: () => null }));
vi.mock('./tagQueries', () => ({ useTodoTagIndex: () => ({}), useTodoTagCounts: () => ({}) }));
vi.mock('./TodoTag', () => ({ TodoTagField: () => null, TodoTagRow: () => null }));
vi.mock('./TodoSessionDetail', () => ({ useTodoSessionDetail: () => ({ detail: undefined, error: '' }) }));
vi.mock('./run', async importOriginal => ({
  ...(await importOriginal<object>()),
  useTodoRunContext: () => ({ context: undefined, loading: false, error: '' }),
  useTodoRun: () => ({ run: vi.fn(), reset: vi.fn(), runBusy: false, runError: '', runMessage: '' }),
}));
vi.mock('./todoMutations', async importOriginal => ({
  ...(await importOriginal<object>()),
  useUpdateTodoMutation: () => ({ isPending: false, mutateAsync: update.mutateAsync }),
  useDeleteTodoMutation: () => ({ isPending: false }),
  useTransferTodoMutation: () => ({ isPending: false }),
  useGithubPushTodoMutation: () => ({ isPending: false }),
  useTodoSessionStop: () => ({ isPending: false }),
  useTodoVerificationRun: () => ({ isPending: false }),
}));
vi.mock('./TodoSessionTimer', async importOriginal => ({
  ...(await importOriginal<object>()),
  useSessionStats: () => ({ stats: null }),
}));

// DropdownMenu unmocked crashes under jsdom; rendering its content inline lets a
// test reach the menu items without opening a popover.
vi.mock('@flanksource/clicky-ui/components', async importOriginal => ({
  ...(await importOriginal<object>()),
  DropdownMenu: ({ trigger, children }: { trigger: React.ReactNode; children: (close: () => void) => React.ReactNode }) => (
    <div>{trigger}{children(() => {})}</div>
  ),
}));

vi.mock('@flanksource/clicky-ui/data', async importOriginal => ({
  ...(await importOriginal<object>()),
  Markdown: ({ text }: { text: string }) => <div>{text}</div>,
}));

const PARENT_ID = '4e2b9a7c-6d1f-4c83-a5e0-7f9b1d3c5a20';
const DIR = '/repos/billing';

const parent: TodoItem = { ref: 'parent-ref', id: PARENT_ID, shortId: 'p123', title: 'Migrate ledger', status: 'in_progress', priority: 'medium' };
const otherTopLevel: TodoItem = { ref: 'other-ref', id: 'other-id', shortId: 'o456', title: 'Rotate keys', status: 'pending', priority: 'low' };
const firstChild: TodoItem = { ref: 'first-child', id: 'first-id', parentId: PARENT_ID, title: 'Backfill rows', status: 'completed', priority: 'medium' };
const secondChild: TodoItem = { ref: 'second-child', id: 'second-id', parentId: PARENT_ID, title: 'Reconcile totals', status: 'pending', priority: 'medium' };

async function renderDetail(props: Partial<TodoDetailProps> & { todo: TodoItem }) {
  render(
    <StatefulTodoDetail loading={false} dir={DIR} onChanged={() => {}} onDeleted={() => {}} {...props} />,
    { wrapper: queryTestWrapper() },
  );
  await act(async () => {});
}

beforeEach(() => {
  update.mutateAsync.mockReset();
  update.mutateAsync.mockImplementation(async () => parent);
  vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => ({}) }) as Response));
  const store = new Map<string, string>();
  vi.stubGlobal('localStorage', { getItem: (key: string) => store.get(key) ?? null, setItem: (key: string, value: string) => store.set(key, value) });
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('TodoDetail children section', () => {
  it('lists the children of a parent between the comment box and the run branch', async () => {
    await renderDetail({ todo: parent, childTodos: [firstChild, secondChild] });

    const section = screen.getByRole('heading', { name: 'Children' }).closest('section')!;
    const follows = (before: Element, after: Element) => !!(before.compareDocumentPosition(after) & Node.DOCUMENT_POSITION_FOLLOWING);
    expect(follows(screen.getByTestId('comment-box'), section)).toBe(true);
    expect(follows(section, screen.getByTestId('run-branch'))).toBe(true);
    expect(within(section).getByText('1/2')).toBeTruthy();
    expect(within(section).getByRole('link', { name: 'Backfill rows' })).toBeTruthy();
    expect(within(section).getByRole('link', { name: 'Reconcile totals' })).toBeTruthy();
    expect(within(section).getByRole('button', { name: /add child/i })).toBeTruthy();
  });

  it('offers Add child on a top-level todo that has no children yet', async () => {
    await renderDetail({ todo: parent, childTodos: [] });

    expect(screen.getByText('No children yet')).toBeTruthy();
    expect(screen.getByRole('button', { name: /add child/i })).toBeTruthy();
  });

  it('shows no Children section on a child that has none of its own', async () => {
    await renderDetail({ todo: firstChild, childTodos: [], parentTodo: parent });

    expect(screen.queryByRole('heading', { name: 'Children' })).toBeNull();
  });
});

describe('TodoDetail parent link', () => {
  it('names the parent above the title and opens it in place', async () => {
    const onOpenTodo = vi.fn();
    await renderDetail({ todo: firstChild, childTodos: [], parentTodo: parent, onOpenTodo });

    const link = screen.getByRole('link', { name: /Migrate ledger/ });
    expect(link.textContent).toBe('↑ Migrate ledger');
    expect(link.getAttribute('href')).toBe('/todos/parent-ref');
    expect(link.compareDocumentPosition(screen.getByRole('heading', { level: 1 })) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();

    fireEvent.click(link);
    expect(onOpenTodo).toHaveBeenCalledExactlyOnceWith('parent-ref');
  });

  it('shows no parent link on a top-level todo', async () => {
    await renderDetail({ todo: parent, childTodos: [] });

    expect(screen.queryByText(/^↑/)).toBeNull();
  });

  it.each([
    ['is still loading', { state: 'loading' }, '↑ loading parent…'],
    ['could not be listed', { state: 'error', message: 'workspace list unavailable' }, '↑ parent unavailable'],
    ['is genuinely missing from the list', { state: 'ready' }, '↑ parent not found'],
  ] as const)('tells apart a parent that %s', async (_name, familyStatus, label) => {
    await renderDetail({ todo: firstChild, childTodos: [], parentTodo: null, familyStatus });

    const note = screen.getByText(label);
    expect(note.tagName).toBe('SPAN');
    expect(note.getAttribute('title')).toBe(familyStatus.state === 'error' ? familyStatus.message : null);
  });
});

describe('TodoDetail children section while the family is unknown', () => {
  it('shows loading, not "No children yet", for a parent whose list has not arrived', async () => {
    await renderDetail({ todo: parent, childTodos: [], familyStatus: { state: 'loading' } });

    expect(screen.getByRole('status').textContent).toBe('Loading children…');
    expect(screen.queryByText('No children yet')).toBeNull();
  });

  it('shows the failure, not "No children yet", when the workspace list failed', async () => {
    await renderDetail({ todo: parent, childTodos: [], familyStatus: { state: 'error', message: 'workspace list unavailable' } });

    expect(screen.getByRole('alert').textContent).toContain('workspace list unavailable');
    expect(screen.queryByText('No children yet')).toBeNull();
  });
});

describe('TodoDetail parent menu', () => {
  const menuItem = (label: string) => screen.getAllByText(label)[0].closest('button') as HTMLButtonElement;

  it('sets the parent to the chosen top-level todo', async () => {
    await renderDetail({ todo: firstChild, childTodos: [], parentTodo: null, parentCandidates: [parent, otherTopLevel] });

    fireEvent.click(screen.getAllByRole('button', { name: /Rotate keys/ })[0]);

    expect(update.mutateAsync).toHaveBeenCalledOnce();
    const request = update.mutateAsync.mock.calls[0][0] as { ref: string; body: string };
    expect(request.ref).toBe('first-child');
    expect(JSON.parse(request.body)).toEqual({ ref: 'first-child', parent: 'other-ref' });
  });

  it('detaches a child with an empty parent', async () => {
    await renderDetail({ todo: firstChild, childTodos: [], parentTodo: parent, parentCandidates: [parent, otherTopLevel] });

    fireEvent.click(menuItem('Remove parent'));

    const request = update.mutateAsync.mock.calls[0][0] as { body: string };
    expect(JSON.parse(request.body)).toEqual({ ref: 'first-child', parent: '' });
  });

  it('offers Remove parent only on a child', async () => {
    await renderDetail({ todo: parent, childTodos: [], parentCandidates: [otherTopLevel] });

    expect(screen.queryByText('Remove parent')).toBeNull();
    expect(screen.getAllByText('Set parent…').length).toBeGreaterThan(0);
  });

  it('hides Set parent while the todo has children, since a parent cannot be a child', async () => {
    await renderDetail({ todo: parent, childTodos: [firstChild], parentCandidates: [otherTopLevel] });

    expect(screen.queryByText('Set parent…')).toBeNull();
  });

  it('hides Set parent when there is no other top-level todo to pick', async () => {
    await renderDetail({ todo: firstChild, childTodos: [], parentTodo: parent, parentCandidates: [] });

    expect(screen.queryByText('Set parent…')).toBeNull();
  });
});
