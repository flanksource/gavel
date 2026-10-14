import type React from 'react';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem, TodoPRLinkItem } from '../../types';
import { StatefulTodoDetail } from './todoDetailTestHarness';
import type { TodoDetailProps } from './TodoDetail';
import { queryTestWrapper } from './queryTestWrapper';

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
  useTodoRun: () => ({ run: vi.fn(), reset: vi.fn(), runBusy: false, runError: '', runMessage: '' }),
}));
vi.mock('./todoMutations', async importOriginal => ({
  ...(await importOriginal<object>()),
  useUpdateTodoMutation: () => ({ isPending: false, mutateAsync: vi.fn() }),
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

const DIR = '/repos/billing';
const todo: TodoItem = { ref: 'abc123', id: 'abc-full', shortId: 'abc123', title: 'Fix rounding', status: 'pending', priority: 'medium' };
const prLinks: TodoPRLinkItem[] = [
  { provider: 'github', repo: 'acme/ledger', number: 42, url: 'https://github.com/acme/ledger/pull/42', relation: 'implements', source: 'manual', createdAt: '2026-10-01T00:00:00Z', title: 'Rework ledger rounding', state: 'OPEN' },
];

let backendPRLinks: TodoPRLinkItem[];
let fetchMock: ReturnType<typeof vi.fn>;

function respond(body: unknown, status = 200) {
  return Promise.resolve({
    ok: status < 400,
    status,
    json: async () => body,
    text: async () => (body === undefined ? '' : JSON.stringify(body)),
    clone() { return this; },
  } as unknown as Response);
}

async function renderDetail(props: Partial<TodoDetailProps> = {}) {
  render(
    <StatefulTodoDetail todo={todo} loading={false} dir={DIR} onChanged={() => {}} onDeleted={() => {}} {...props} />,
    { wrapper: queryTestWrapper() },
  );
  await waitFor(() => expect(fetchMock).toHaveBeenCalled());
  await act(async () => { await new Promise(resolve => setTimeout(resolve, 0)); });
}

const menuButton = (label: string) => screen.getAllByText(label)[0]!.closest('button') as HTMLButtonElement;

beforeEach(() => {
  backendPRLinks = prLinks;
  fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (init?.method === 'DELETE') return respond(undefined, 204);
    if (url.startsWith('/api/todos/pr-links')) return respond(backendPRLinks);
    if (url.startsWith('/api/todos/links')) return respond({ links: [] });
    return respond([]);
  });
  vi.stubGlobal('fetch', fetchMock);
  const store = new Map<string, string>();
  vi.stubGlobal('localStorage', { getItem: (key: string) => store.get(key) ?? null, setItem: (key: string, value: string) => store.set(key, value) });
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('TodoDetail link menu', () => {
  it('offers Link PR… and opens the picker dialog', async () => {
    await renderDetail();
    expect(screen.queryByLabelText('Search pull requests')).toBeNull();

    fireEvent.click(menuButton('Link PR…'));

    expect(await screen.findByLabelText('Search pull requests')).toBeTruthy();
    expect(screen.getByLabelText('Pull request URL')).toBeTruthy();
  });

  it('lists the current PR links under Unlink PR and unlinks the chosen one', async () => {
    await renderDetail();

    expect(screen.getAllByText('Unlink PR').length).toBeGreaterThan(0);
    fireEvent.click(screen.getAllByRole('button', { name: /PR #42.*Rework ledger rounding/ })[0]!);

    await waitFor(() => {
      const call = fetchMock.mock.calls.find(([, init]) => init?.method === 'DELETE');
      expect(call).toBeTruthy();
      const url = new URL(String(call![0]), 'http://x');
      expect(url.pathname).toBe('/api/todos/pr-links');
      expect(Object.fromEntries(url.searchParams)).toEqual({ dir: DIR, id: 'abc123', url: prLinks[0]!.url });
    });
  });

  it('hides Unlink PR when the todo has no PR links', async () => {
    backendPRLinks = [];
    await renderDetail();

    expect(screen.getAllByText('Link PR…').length).toBeGreaterThan(0);
    expect(screen.queryByText('Unlink PR')).toBeNull();
  });
});

describe('TodoDetail overview links', () => {
  it('shows the Pull requests group on the overview', async () => {
    await renderDetail();

    const group = screen.getByRole('heading', { name: 'Pull requests' }).closest('section')!;
    expect(within(group).getByText('Rework ledger rounding')).toBeTruthy();
  });
});
