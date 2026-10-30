import type { ComponentType, ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { PRActions, type ExtraAction } from './PRActions';
import { queryKeys } from '../query';
import type { PRDetail, PRItem } from '../types';
import { UiAdd, type IconProps } from '@flanksource/clicky-ui/icons';

// clicky-ui's SplitButton / DropdownMenu pull @floating-ui/react, which resolves
// a duplicate React 18 under vitest and crashes on render. Stub the components so
// the test exercises PRActions' own collapse branching, not clicky internals. A
// closed DropdownMenu renders only its trigger — mirror that so hidden menu items
// stay out of the DOM.
vi.mock('@flanksource/clicky-ui/components', async (importOriginal) => {
  const { Button } = await importOriginal<typeof import('@flanksource/clicky-ui/components')>();
  return {
  Button,
  SplitButton: ({ label, onClick, title, disabled, items }: { label: ReactNode; onClick?: () => void; title?: string; disabled?: boolean; items?: Array<{ label: ReactNode; onSelect?: () => void }> }) => (
    <div><Button onClick={onClick} title={title} disabled={disabled}>{label}</Button>
      {items?.map((item, index) => <Button key={index} onClick={item.onSelect} disabled={disabled}>{item.label}</Button>)}
    </div>
  ),
  DropdownMenu: ({ trigger }: { trigger: ReactNode }) => <div>{trigger}</div>,
  Modal: ({ children, footer, title }: { children: ReactNode; footer: ReactNode; title: string }) => (
    // oxlint-disable-next-line clicky-ui/no-adhoc-overlay -- this IS the stub standing in for clicky's Modal.
    <div role="dialog" aria-label={title}>{children}{footer}</div>
  ),
  };
});

// SessionTail is clicky-ui's (tested there); the double records which session it follows.
vi.mock('@flanksource/clicky-ui/ai', () => ({
  SessionTail: ({ src }: { src: string }) => <span data-testid="session-tail">{src}</span>,
}));

// The AI rebase run stream is the tab's event hub; this double lets a test play Captain's frames.
const streams = vi.hoisted(() => [] as { url: string; listeners: Map<string, (event: Event) => void> }[]);
vi.mock('../eventHub', () => ({
  openEventStream: (url: string) => {
    const stream = { url, listeners: new Map<string, (event: Event) => void>() };
    streams.push(stream);
    return {
      addEventListener: (name: string, listener: (event: Event) => void) => stream.listeners.set(name, listener),
      removeEventListener: (name: string, listener: (event: Event) => void) => {
        if (stream.listeners.get(name) === listener) stream.listeners.delete(name);
      },
      close: () => {},
    };
  },
}));

function makePR(overrides: Partial<PRItem> = {}): PRItem {
  return {
    number: 7,
    title: 'Test PR',
    author: 'octocat',
    repo: 'acme/widget',
    source: 'feature',
    target: 'main',
    state: 'OPEN',
    isDraft: false,
    url: 'https://github.com/acme/widget/pull/7',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

const newTodo: ExtraAction = { label: 'New todo', icon: UiAdd as ComponentType<IconProps>, onClick: () => {} };

const detail = { pr: { nodeId: 'PR_node_7', state: 'OPEN' } } as PRDetail;

function renderActions(props: Partial<React.ComponentProps<typeof PRActions>> = {}) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return {
    client,
    ...render(
      <QueryClientProvider client={client}>
        <PRActions pr={makePR()} detail={detail} {...props} />
      </QueryClientProvider>,
    ),
  };
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('PRActions responsive collapse', () => {
  it('disables mutations while the parent TODO run is active, including an open confirmation', () => {
    const view = renderActions();
    fireEvent.click(screen.getByRole('button', { name: 'Merge' }));
    view.rerender(<QueryClientProvider client={view.client}><PRActions pr={makePR()} detail={detail} disabled /></QueryClientProvider>);
    expect(screen.getAllByRole('button', { name: 'Merge' }).every(button => (button as HTMLButtonElement).disabled)).toBe(true);
    expect((screen.getByRole('button', { name: 'Approve' }) as HTMLButtonElement).disabled).toBe(true);
  });

  it('confirms an Azure rebase merge without requiring a GitHub node ID', async () => {
    const url = 'https://dev.azure.com/acme/product/_git/service/pullrequest/27';
    const fetchMock = vi.fn(async () => new Response('{"status":"ok"}'));
    vi.stubGlobal('fetch', fetchMock);
    renderActions({ pr: makePR({ number: 27, repo: 'acme/product/service', url }), detail: { pr: { provider: 'azuredevops', state: 'OPEN', headRefOid: 'reviewed-head' } } as PRDetail });
    expect(screen.queryByRole('button', { name: 'Approve' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Enable auto-merge (rebase)' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Update branch' })).toBeNull();
    expect((screen.getByRole('button', { name: 'Merge' }) as HTMLButtonElement).disabled).toBe(false);
    fireEvent.click(screen.getByRole('button', { name: 'Merge' }));
    expect(fetchMock).not.toHaveBeenCalled();
    fireEvent.click(screen.getAllByRole('button', { name: 'Merge' }).at(-1)!);
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith('/api/prs/merge', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url, method: 'rebase', auto: false, headSha: 'reviewed-head' }),
    }));
  });

  it('names rebase, squash and merge-commit strategies on demand', () => {
    renderActions();
    for (const label of ['Rebase and merge', 'Squash and merge', 'Create a merge commit']) {
      expect(screen.getByRole('button', { name: label })).toBeTruthy();
    }
  });
  it('renders every action inline when not collapsed', () => {
    renderActions({ detail: null, collapsed: false, extras: [newTodo] });

    expect(screen.getByText('Approve')).toBeTruthy();
    expect(screen.getByText('Merge')).toBeTruthy();
    expect(screen.getByText('New todo')).toBeTruthy();
    expect(screen.queryByLabelText('Pull request actions')).toBeNull();
  });

  it('collapses the actions behind an overflow trigger when narrow', () => {
    renderActions({ detail: null, collapsed: true, extras: [newTodo] });

    expect(screen.getByLabelText('Pull request actions')).toBeTruthy();
    // The closed menu keeps the individual actions out of the DOM until opened.
    expect(screen.queryByText('Approve')).toBeNull();
    expect(screen.queryByText('New todo')).toBeNull();
  });

  it('keeps a lone action inline even when collapsed', () => {
    // A merged PR has no GitHub actions, so only the New todo extra remains.
    renderActions({ pr: makePR({ state: 'MERGED' }), detail: null, collapsed: true, extras: [newTodo] });

    expect(screen.getByText('New todo')).toBeTruthy();
    expect(screen.queryByLabelText('Pull request actions')).toBeNull();
  });

  it('renders nothing for a merged PR with no extras', () => {
    const { container } = renderActions({ pr: makePR({ state: 'MERGED' }), detail: null, collapsed: false });
    expect(container.firstChild).toBeNull();
  });

  it('submits approval once and invalidates only the selected PR caches', async () => {
    let finishRequest: (response: Response) => void = () => undefined;
    const fetchMock = vi.fn(() => new Promise<Response>(resolve => { finishRequest = resolve; }));
    vi.stubGlobal('fetch', fetchMock);
    const onChanged = vi.fn();
    const { client } = renderActions({ onChanged });
    const invalidate = vi.spyOn(client, 'invalidateQueries');

    fireEvent.click(screen.getByRole('button', { name: 'Approve' }));
    fireEvent.change(screen.getByPlaceholderText('Optional approval comment'), { target: { value: 'Looks good' } });
    const confirm = screen.getAllByRole('button', { name: 'Approve' }).at(-1)!;
    fireEvent.click(confirm);
    fireEvent.click(confirm);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(fetchMock).toHaveBeenCalledWith('/api/prs/approve', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ repo: 'acme/widget', number: 7, nodeId: 'PR_node_7', body: 'Looks good' }),
    });
    finishRequest(new Response(JSON.stringify({ status: 'ok' }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));

    await waitFor(() => expect(onChanged).toHaveBeenCalledTimes(1));
    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.prSnapshot(), exact: true });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.prDetail('acme/widget', 7), exact: true });
    expect(invalidate).toHaveBeenCalledTimes(2);
  });

  it('keeps the confirmation modal open and surfaces contextual mutation errors', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ error: 'branch protection denied the merge' }), {
      status: 409,
      headers: { 'Content-Type': 'application/json' },
    })));
    const onChanged = vi.fn();
    renderActions({ onChanged });

    fireEvent.click(screen.getByRole('button', { name: 'Merge' }));
    fireEvent.click(screen.getAllByRole('button', { name: 'Merge' }).at(-1)!);

    expect(await screen.findByText('Merge pull request: branch protection denied the merge')).toBeTruthy();
    expect(screen.getByRole('dialog', { name: 'Merge pull request' })).toBeTruthy();
    expect(onChanged).not.toHaveBeenCalled();
  });
});

describe('PRActions update branch', () => {
  const behindPR = makePR({ behind: 2 });
  const headDetail = { pr: { nodeId: 'PR_node_7', state: 'OPEN', headRefOid: 'abc123' } } as PRDetail;
  const rebaseVars = {
    repo: '/src/widget', remote: 'origin', number: 7, url: 'https://github.com/acme/widget/pull/7',
    branch: 'feature', base: 'main', baseSha: 'def456', head: 'abc123',
  };
  const conflictBody = (extra: Record<string, unknown>) => new Response(JSON.stringify({
    error: 'update branch (rebase): the branch does not rebase cleanly onto its base', conflict: true, ...extra,
  }), { status: 409, headers: { 'Content-Type': 'application/json' } });

  function updateBranch(fetchMock: ReturnType<typeof vi.fn>, props: Partial<React.ComponentProps<typeof PRActions>> = {}) {
    vi.stubGlobal('fetch', fetchMock);
    renderActions({ pr: behindPR, detail: headDetail, ...props });
    fireEvent.click(screen.getByRole('button', { name: 'Update' }));
    fireEvent.click(screen.getAllByRole('button', { name: 'Update branch' }).at(-1)!);
  }

  it('rebases at the head the UI saw', async () => {
    const fetchMock = vi.fn(async () => new Response('{"status":"ok"}'));
    updateBranch(fetchMock);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(fetchMock).toHaveBeenCalledWith('/api/prs/update-branch', expect.objectContaining({
      body: JSON.stringify({ repo: 'acme/widget', number: 7, nodeId: 'PR_node_7', headSha: 'abc123' }),
    }));
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
  });

  it('offers the AI rebase on a conflict and runs the pr-rebase prompt with the server variables', async () => {
    const fetchMock = vi.fn(async (url: string) => url === '/api/prs/update-branch'
      ? conflictBody({ rebase: rebaseVars })
      : new Response(JSON.stringify({ runId: 'run-1' })));
    const onChanged = vi.fn();
    updateBranch(fetchMock, { onChanged });

    expect((await screen.findByRole('alert')).textContent).toContain('does not rebase cleanly');
    fireEvent.click(screen.getByRole('button', { name: /Rebase with AI/ }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    const [url, init] = fetchMock.mock.calls[1] as unknown as [string, RequestInit];
    expect(url).toBe('/api/v1/prompt/run?args=pr-rebase');
    expect(JSON.parse(String(init.body))).toEqual({ variables: rebaseVars });
    await waitFor(() => expect(streams.at(-1)?.url).toBe('/api/captain/prompt/runs/run-1/stream'));

    act(() => streams.at(-1)!.listeners.get('run')?.(new MessageEvent('run', { data: JSON.stringify({ sessionId: 's-1' }) })));
    expect(screen.getByTestId('session-tail').textContent).toBe('/api/captain/sessions/s-1');

    act(() => streams.at(-1)!.listeners.get('done')?.(new MessageEvent('done', { data: JSON.stringify({ success: true }) })));
    expect(onChanged).toHaveBeenCalledTimes(1);
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
  });

  it('says why the AI rebase cannot run instead of offering it', async () => {
    updateBranch(vi.fn(async () => conflictBody({ rebaseBlocked: 'no project has a local checkout of acme/widget' })));

    expect((await screen.findByRole('alert')).textContent).toContain('AI rebase unavailable: no project has a local checkout of acme/widget');
    expect(screen.queryByRole('button', { name: /Rebase with AI/ })).toBeNull();
    expect((screen.getAllByRole('button', { name: 'Update branch' }).at(-1) as HTMLButtonElement).disabled).toBe(true);
  });
});
