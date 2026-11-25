import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { queryKeys } from '../query';
import { projectDiffQueryKey } from './projectMutations';
import { ProjectDiffView } from './ProjectDiffView';

interface Selection {
  isSelected: (target: { side: 'old' | 'new'; line: number }) => boolean;
  onChange: (targets: { side: 'old' | 'new'; line: number }[], selected: boolean) => void;
  disabled?: boolean;
  label?: (target: { side: 'old' | 'new'; line: number }, selected: boolean) => string;
  hunkLabel?: (hunk: number, selected: boolean) => string;
}

// A stand-in for clicky-ui's GitDiffPanel that exposes the diff text and drives
// its lineSelection the way a gutter click does.
vi.mock('@flanksource/clicky-ui/data', () => ({
  GitDiffPanel: ({ loading, payload, error, lineSelection }: { loading: boolean; payload: { diff?: string } | null; error: string; lineSelection?: Selection }) => (
    <div data-testid={lineSelection ? 'gutter-panel' : 'plain-panel'}>
      <span>{loading ? 'loading' : error || payload?.diff || 'empty'}</span>
      {lineSelection && (
        <>
          <input
            type="checkbox"
            aria-label={lineSelection.label?.({ side: 'new', line: 3 }, lineSelection.isSelected({ side: 'new', line: 3 }))}
            checked={lineSelection.isSelected({ side: 'new', line: 3 })}
            onChange={event => lineSelection.onChange([{ side: 'new', line: 3 }, { side: 'old', line: 2 }], event.target.checked)}
          />
          <input
            type="checkbox"
            aria-label={lineSelection.label?.({ side: 'old', line: 2 }, lineSelection.isSelected({ side: 'old', line: 2 }))}
            checked={lineSelection.isSelected({ side: 'old', line: 2 })}
            onChange={event => lineSelection.onChange([{ side: 'old', line: 2 }], event.target.checked)}
          />
          <input
            type="checkbox"
            aria-label={lineSelection.hunkLabel?.(1, lineSelection.isSelected({ side: 'new', line: 3 }))}
            checked={lineSelection.isSelected({ side: 'new', line: 3 })}
            onChange={event => lineSelection.onChange([{ side: 'new', line: 3 }], event.target.checked)}
          />
        </>
      )}
    </div>
  ),
}));

vi.mock('@flanksource/clicky-ui/icons', () => ({ UiDiff: () => <span /> }));

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

function createClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: Infinity }, mutations: { retry: false } } });
}

function Provider({ client, children }: { client: QueryClient; children: ReactNode }) {
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

const json = (value: unknown, status = 200) => new Response(JSON.stringify(value), { status, headers: { 'Content-Type': 'application/json' } });
const payload = (extra: Record<string, unknown> = {}) => ({ path: 'one.go', diff: 'combined diff', staged: '', unstaged: '', ...extra });

describe('ProjectDiffView queries', () => {
  it('deduplicates concurrent consumers and reuses the cached diff after remount', async () => {
    const fetchMock = vi.fn(async () => json(payload({ diff: 'diff:one.go' })));
    vi.stubGlobal('fetch', fetchMock);
    const client = createClient();

    const first = render(
      <Provider client={client}>
        <ProjectDiffView projectName="gavel" path="one.go" />
        <ProjectDiffView projectName="gavel" path="one.go" />
      </Provider>,
    );

    expect(await screen.findAllByText('diff:one.go')).toHaveLength(2);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    first.unmount();

    render(
      <Provider client={client}>
        <ProjectDiffView projectName="gavel" path="one.go" />
      </Provider>,
    );

    expect(await screen.findByText('diff:one.go')).toBeTruthy();
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it('cancels the prior request when the diff identity changes', async () => {
    const signals: AbortSignal[] = [];
    vi.stubGlobal('fetch', vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
      signals.push(init?.signal as AbortSignal);
      return new Promise<Response>(() => undefined);
    }));
    const client = createClient();

    const view = render(
      <Provider client={client}>
        <ProjectDiffView projectName="gavel" path="one.go" />
      </Provider>,
    );
    await waitFor(() => expect(signals).toHaveLength(1));

    view.rerender(
      <Provider client={client}>
        <ProjectDiffView projectName="gavel" path="two.go" />
      </Provider>,
    );

    await waitFor(() => expect(signals).toHaveLength(2));
    expect(signals[0].aborted).toBe(true);
  });
});

describe('ProjectDiffView gutter staging', () => {
  function setup({ diff = payload({ staged: 'STAGED PATCH', unstaged: 'UNSTAGED PATCH' }), failStage = false, worktree }: { diff?: unknown; failStage?: boolean; worktree?: string } = {}) {
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') return failStage ? json({ error: 'patch does not apply' }, 400) : json({ ok: true });
      return json(diff);
    });
    vi.stubGlobal('fetch', fetchMock);
    const client = createClient();
    const invalidate = vi.spyOn(client, 'invalidateQueries');
    render(<Provider client={client}><ProjectDiffView projectName="gavel" path="one.go" {...(worktree ? { worktree } : {})} /></Provider>);
    const posts = () => fetchMock.mock.calls.filter(([, init]) => init?.method === 'POST').map(([url, init]) => [String(url), JSON.parse(String(init?.body))]);
    return { fetchMock, invalidate, posts };
  }

  it('renders an Unstaged and a Staged section, each with a gutter', async () => {
    setup();
    const unstaged = (await screen.findByRole('region', { name: 'Unstaged changes' }));
    const staged = screen.getByRole('region', { name: 'Staged changes' });
    expect(within(unstaged).getByText('UNSTAGED PATCH')).toBeTruthy();
    expect(within(staged).getByText('STAGED PATCH')).toBeTruthy();
    expect(screen.getAllByTestId('gutter-panel')).toHaveLength(2);
    expect(screen.queryByText('combined diff')).toBeNull();
  });

  it.each([
    { name: 'staged', diff: payload({ staged: 'STAGED PATCH' }), shown: 'Staged changes', hidden: 'Unstaged changes' },
    { name: 'unstaged', diff: payload({ unstaged: 'UNSTAGED PATCH' }), shown: 'Unstaged changes', hidden: 'Staged changes' },
  ])('hides the empty section of a file with only $name changes', async ({ diff, shown, hidden }) => {
    setup({ diff });
    expect(await screen.findByRole('region', { name: shown })).toBeTruthy();
    expect(screen.queryByRole('region', { name: hidden })).toBeNull();
  });

  it('keeps the combined diff without a gutter for a folder target', async () => {
    setup({ diff: payload({ path: 'pkg', diff: 'folder diff' }) });
    expect(await screen.findByText('folder diff')).toBeTruthy();
    expect(screen.getByTestId('plain-panel')).toBeTruthy();
    expect(screen.queryByTestId('gutter-panel')).toBeNull();
    expect(screen.queryByRole('region', { name: 'Unstaged changes' })).toBeNull();
  });

  it('labels the unstaged gutter as Stage and the staged gutter as Unstage', async () => {
    setup();
    const unstaged = within(await screen.findByRole('region', { name: 'Unstaged changes' }));
    const staged = within(screen.getByRole('region', { name: 'Staged changes' }));
    expect(unstaged.getByRole('checkbox', { name: 'Stage added line 3' })).toBeTruthy();
    expect(unstaged.getByRole('checkbox', { name: 'Stage removed line 2' })).toBeTruthy();
    expect(unstaged.getByRole('checkbox', { name: 'Stage hunk 1' })).toBeTruthy();
    expect((unstaged.getByRole('checkbox', { name: 'Stage added line 3' }) as HTMLInputElement).checked).toBe(false);
    expect(staged.getByRole('checkbox', { name: 'Unstage added line 3' })).toBeTruthy();
    expect(staged.getByRole('checkbox', { name: 'Unstage removed line 2' })).toBeTruthy();
    expect(staged.getByRole('checkbox', { name: 'Unstage hunk 1' })).toBeTruthy();
    expect((staged.getByRole('checkbox', { name: 'Unstage added line 3' }) as HTMLInputElement).checked).toBe(true);
  });

  it('stages the clicked lines through stage/lines and refreshes status and diffs', async () => {
    const { posts, invalidate } = setup();
    const unstaged = within(await screen.findByRole('region', { name: 'Unstaged changes' }));
    fireEvent.click(unstaged.getByRole('checkbox', { name: 'Stage added line 3' }));
    await waitFor(() => expect(posts()).toHaveLength(1));
    expect(posts()[0]).toEqual(['/api/projects/gavel/stage/lines', {
      path: 'one.go',
      lines: [{ side: 'new', line: 3 }, { side: 'old', line: 2 }],
    }]);
    await waitFor(() => {
      expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.projectStatusScope('gavel') });
      expect(invalidate).toHaveBeenCalledWith({ queryKey: projectDiffQueryKey('gavel') });
    });
  });

  it('unstages the clicked lines with unstage=true', async () => {
    const { posts } = setup();
    const staged = within(await screen.findByRole('region', { name: 'Staged changes' }));
    fireEvent.click(staged.getByRole('checkbox', { name: 'Unstage removed line 2' }));
    await waitFor(() => expect(posts()).toHaveLength(1));
    expect(posts()[0]).toEqual(['/api/projects/gavel/stage/lines', {
      path: 'one.go',
      unstage: true,
      lines: [{ side: 'old', line: 2 }],
    }]);
  });

  it('scopes the request to the worktree', async () => {
    const { posts } = setup({ worktree: '/work/gavel-wt' });
    fireEvent.click(within(await screen.findByRole('region', { name: 'Unstaged changes' })).getByRole('checkbox', { name: 'Stage hunk 1' }));
    await waitFor(() => expect(posts()).toHaveLength(1));
    expect(posts()[0]![0]).toBe('/api/projects/gavel/stage/lines?worktree=%2Fwork%2Fgavel-wt');
  });

  it('shows a failed stage inline and leaves the sections in place', async () => {
    setup({ failStage: true });
    fireEvent.click(within(await screen.findByRole('region', { name: 'Unstaged changes' })).getByRole('checkbox', { name: 'Stage added line 3' }));
    expect((await screen.findByRole('alert')).textContent).toBe('Failed to stage lines of one.go: patch does not apply');
    expect(screen.getByRole('region', { name: 'Unstaged changes' })).toBeTruthy();
  });
});
