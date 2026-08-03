import type { ComponentType, ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
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
vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, onClick, title, disabled, loading, ...rest }: any) => (
    <button onClick={onClick} title={title} disabled={disabled || loading} aria-label={rest['aria-label']}>{children}</button>
  ),
  SplitButton: ({ label, onClick, title, disabled }: { label: ReactNode; onClick?: () => void; title?: string; disabled?: boolean }) => (
    <button onClick={onClick} title={title} disabled={disabled}>{label}</button>
  ),
  DropdownMenu: ({ trigger }: { trigger: ReactNode }) => <div>{trigger}</div>,
  Modal: ({ children, footer, title }: { children: ReactNode; footer: ReactNode; title: string }) => (
    <div role="dialog" aria-label={title}>{children}{footer}</div>
  ),
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
