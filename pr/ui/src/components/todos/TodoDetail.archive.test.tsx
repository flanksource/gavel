import type React from 'react';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem } from '../../types';
import { StatefulTodoDetail } from './todoDetailTestHarness';
import type { TodoDetailProps } from './TodoDetail';
import { queryTestWrapper } from './queryTestWrapper';
import { TodoMutationError } from './todoMutations';

// `reset` is stable, as the real hook's is: the detail clears its transient state
// whenever it changes, which would close the prompt the instant it opened.
const del = vi.hoisted(() => ({ mutateAsync: vi.fn(), run: vi.fn(), reset: vi.fn() }));

vi.mock('./TodoTimeline', () => ({ TodoTimeline: () => null }));
vi.mock('./TodoRunBranch', () => ({ TodoRunBranch: () => null }));
vi.mock('./TodoSession', () => ({ TodoSession: () => null }));
vi.mock('./TodoPlan', () => ({ TodoPlan: () => null }));
vi.mock('./TodoVerification', () => ({ TodoVerification: () => null }));
vi.mock('./TodoCompose', () => ({ TodoTitleEditor: () => null, TodoBodyEditor: () => null, TodoCommentBox: () => null }));
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
  useTodoRun: () => ({ run: del.run, reset: del.reset, runBusy: false, runError: '', runMessage: '' }),
}));
vi.mock('./todoMutations', async importOriginal => ({
  ...(await importOriginal<object>()),
  useUpdateTodoMutation: () => ({ isPending: false, mutateAsync: vi.fn() }),
  useDeleteTodoMutation: () => ({ isPending: false, mutateAsync: del.mutateAsync }),
  useTransferTodoMutation: () => ({ isPending: false }),
  useGithubPushTodoMutation: () => ({ isPending: false }),
  useTodoSessionStop: () => ({ isPending: false }),
  useTodoVerificationRun: () => ({ isPending: false }),
}));
vi.mock('./TodoSessionTimer', async importOriginal => ({
  ...(await importOriginal<object>()),
  useSessionStats: () => ({ stats: null }),
}));
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

const PARENT_ID = '9b3d5f71-2a4c-4e86-b0d1-3f5a7c9e1b24';
const DIR = '/repos/billing';

const parent: TodoItem = { ref: 'parent-ref', id: PARENT_ID, title: 'Migrate ledger', status: 'in_progress', priority: 'medium' };
const child = (ref: string, title: string, status: TodoItem['status']): TodoItem =>
  ({ ref, id: `id-${ref}`, parentId: PARENT_ID, title, status, priority: 'medium' });
const openOne = child('c1', 'Backfill rows', 'pending');
const openTwo = child('c2', 'Reconcile totals', 'in_progress');
const closed = child('c3', 'Old cleanup', 'completed');
const plain: TodoItem = { ref: 'plain-ref', id: 'plain-id', title: 'Rotate keys', status: 'pending', priority: 'low' };

const onDeleted = vi.fn();
const confirm = vi.fn(() => true);

async function renderDetail(props: Partial<TodoDetailProps> & { todo: TodoItem }) {
  render(
    <StatefulTodoDetail loading={false} dir={DIR} onChanged={() => {}} onDeleted={onDeleted} {...props} />,
    { wrapper: queryTestWrapper() },
  );
  await act(async () => {});
}

const archive = async () => {
  fireEvent.click(screen.getAllByText('Archive todo')[0].closest('button')!);
  await act(async () => {});
};
const dialog = () => screen.queryByRole('dialog');
// The buttons that carry a label: the modal's own icon-only controls are not choices.
const choices = (box: HTMLElement) => within(box).getAllByRole('button').map(button => button.textContent).filter(Boolean);

beforeEach(() => {
  del.mutateAsync.mockReset();
  del.mutateAsync.mockResolvedValue(undefined);
  onDeleted.mockReset();
  confirm.mockReset();
  confirm.mockReturnValue(true);
  vi.stubGlobal('confirm', confirm);
  vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => ({}) }) as Response));
  const store = new Map<string, string>();
  vi.stubGlobal('localStorage', { getItem: (key: string) => store.get(key) ?? null, setItem: (key: string, value: string) => store.set(key, value) });
});

afterEach(() => vi.unstubAllGlobals());

describe('archiving a todo with no open children', () => {
  it.each([
    ['no children at all', []],
    ['only closed children', [closed]],
  ])('archives as before, with the plain confirmation and no children parameter, for %s', async (_name, childTodos) => {
    await renderDetail({ todo: parent, childTodos });
    await archive();

    expect(dialog()).toBeNull();
    expect(confirm).toHaveBeenCalledWith('Archive this todo?');
    expect(del.mutateAsync).toHaveBeenCalledExactlyOnceWith({ ref: 'parent-ref', children: undefined });
    expect(onDeleted).toHaveBeenCalledOnce();
  });

  it('archives a plain todo the same way', async () => {
    await renderDetail({ todo: plain });
    await archive();

    expect(del.mutateAsync).toHaveBeenCalledExactlyOnceWith({ ref: 'plain-ref', children: undefined });
  });
});

describe('archiving a parent with open children', () => {
  it('asks what to do with them, naming the parent and its open children, and sends nothing yet', async () => {
    await renderDetail({ todo: parent, childTodos: [openOne, openTwo, closed] });
    await archive();

    const box = within(screen.getByRole('dialog', { name: /Migrate ledger/ }));
    expect(box.getByText(/2 open children/)).toBeTruthy();
    expect(box.getByText('Backfill rows')).toBeTruthy();
    expect(box.getByText('Reconcile totals')).toBeTruthy();
    expect(box.queryByText('Old cleanup')).toBeNull();
    expect(choices(screen.getByRole('dialog'))).toEqual([
      'Cancel', 'Make them full todos', 'Archive children too',
    ]);
    expect(confirm).not.toHaveBeenCalled();
    expect(del.mutateAsync).not.toHaveBeenCalled();
  });

  it.each([
    ['Archive children too', 'archive'],
    ['Make them full todos', 'detach'],
  ])('sends children=%s when "%s" is chosen and then leaves the archived todo', async (label, children) => {
    await renderDetail({ todo: parent, childTodos: [openOne, openTwo] });
    await archive();

    fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: label }));

    await waitFor(() => expect(onDeleted).toHaveBeenCalledOnce());
    expect(del.mutateAsync).toHaveBeenCalledExactlyOnceWith({ ref: 'parent-ref', children });
    expect(dialog()).toBeNull();
  });

  it('sends nothing when cancelled', async () => {
    await renderDetail({ todo: parent, childTodos: [openOne] });
    await archive();

    fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Cancel' }));

    expect(dialog()).toBeNull();
    expect(del.mutateAsync).not.toHaveBeenCalled();
    expect(onDeleted).not.toHaveBeenCalled();
  });

  it('says "1 open child" for a single one', async () => {
    await renderDetail({ todo: parent, childTodos: [openOne] });
    await archive();

    expect(within(screen.getByRole('dialog')).getByText(/1 open child\b/)).toBeTruthy();
  });

  it('lists only the first few titles when there are many', async () => {
    const many = ['a', 'b', 'c', 'd', 'e', 'f', 'g'].map(ref => child(ref, `Task ${ref}`, 'pending'));
    await renderDetail({ todo: parent, childTodos: many });
    await archive();

    const box = within(screen.getByRole('dialog'));
    expect(box.getByText(/7 open children/)).toBeTruthy();
    expect(box.getByText('Task e')).toBeTruthy();
    expect(box.queryByText('Task f')).toBeNull();
    expect(box.getByText('and 2 more')).toBeTruthy();
  });
});

describe('archiving before the children are known', () => {
  const conflict = new TodoMutationError('Failed to archive todo parent-ref: todo has 2 open children; choose archive or detach', 409);

  it.each([
    ['still loading', { state: 'loading' }],
    ['failed to load', { state: 'error', message: 'workspace list unavailable' }],
  ] as const)('sends without children while the list is %s, then asks on the 409 and retries with the choice', async (_name, familyStatus) => {
    del.mutateAsync.mockRejectedValueOnce(conflict);
    await renderDetail({ todo: parent, childTodos: [], familyStatus });
    await archive();

    await waitFor(() => expect(screen.getByRole('dialog')).toBeTruthy());
    expect(del.mutateAsync).toHaveBeenNthCalledWith(1, { ref: 'parent-ref', children: undefined });
    const box = within(screen.getByRole('dialog'));
    expect(box.getByText(/todo has 2 open children; choose archive or detach/)).toBeTruthy();
    expect(onDeleted).not.toHaveBeenCalled();

    fireEvent.click(box.getByRole('button', { name: 'Archive children too' }));

    await waitFor(() => expect(onDeleted).toHaveBeenCalledOnce());
    expect(del.mutateAsync).toHaveBeenNthCalledWith(2, { ref: 'parent-ref', children: 'archive' });
  });

  it('offers the same choices and cancels without a second request', async () => {
    del.mutateAsync.mockRejectedValueOnce(conflict);
    await renderDetail({ todo: parent, childTodos: [] });
    await archive();
    await waitFor(() => expect(screen.getByRole('dialog')).toBeTruthy());

    expect(choices(screen.getByRole('dialog'))).toEqual([
      'Cancel', 'Make them full todos', 'Archive children too',
    ]);
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));

    expect(dialog()).toBeNull();
    expect(del.mutateAsync).toHaveBeenCalledOnce();
  });

  it('reports any other failure instead of opening the dialog', async () => {
    del.mutateAsync.mockRejectedValueOnce(new TodoMutationError('Failed to archive todo parent-ref: db is down', 500));
    await renderDetail({ todo: parent, childTodos: [] });
    await archive();

    await waitFor(() => expect(screen.getByText(/db is down/)).toBeTruthy());
    expect(dialog()).toBeNull();
  });
});
