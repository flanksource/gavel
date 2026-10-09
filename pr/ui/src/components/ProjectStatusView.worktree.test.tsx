import type React from 'react';
import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { projectStatusResponse as statusResponse } from './ProjectStatusView.fixture';
import { ProjectStatusView } from './ProjectStatusView';

const WORKTREE = '/work/gavel-feat';
const ENCODED = '%2Fwork%2Fgavel-feat';

/* oxlint-disable clicky-ui/prefer-clicky-components --
   Raw elements are the test doubles for clicky-ui's Button/SplitButton: the real
   SplitButton mounts floating-ui, which crashes under vitest. */
vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, onClick, loading: _loading, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { loading?: boolean }) => (
    <button onClick={onClick} {...props}>{children}</button>
  ),
  SplitButton: ({ label, onClick, disabled, title }: { label: ReactNode; onClick: () => void; disabled?: boolean; title?: string }) => (
    <button type="button" onClick={onClick} disabled={disabled} title={title}>{label}</button>
  ),
  Modal: () => null,
  JsonSchemaForm: () => null,
  SplitPane: ({ left, right }: { left: ReactNode; right: ReactNode }) => <div>{left}{right}</div>,
}));
/* oxlint-enable clicky-ui/prefer-clicky-components */

vi.mock('./ProjectCommitTasks', () => ({ ProjectCommitTasks: () => null }));

// The focus lease has its own spec (useGitFocus.test.tsx); here it would only
// add POSTs to the request counts these tests assert on.
const { useGitFocusMock } = vi.hoisted(() => ({ useGitFocusMock: vi.fn(() => ({ error: '' })) }));
vi.mock('../useGitFocus', () => ({ useGitFocus: useGitFocusMock }));

let queryClient: QueryClient;

beforeEach(() => {
  queryClient = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: Infinity } } });
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

const renderWorktree = (props: Partial<React.ComponentProps<typeof ProjectStatusView>> = {}) => render(
  <QueryClientProvider client={queryClient}>
    <ProjectStatusView project={statusResponse.project} worktree={WORKTREE} {...props} />
  </QueryClientProvider>,
);

function stubFetch(extra: (url: string, init?: RequestInit) => unknown = () => undefined) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    const body = extra(url, init);
    return new Response(JSON.stringify(body ?? (init?.method === 'POST' ? { runId: 'commit-run-1' } : statusResponse)), {
      status: init?.method === 'POST' ? 202 : 200,
      headers: { 'Content-Type': 'application/json' },
    });
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

describe('ProjectStatusView in a linked worktree', () => {
  it.each([
    ['holds the git focus lease for the worktree on screen', WORKTREE, { project: 'gavel', worktree: WORKTREE }],
    ['holds the git focus lease for the main checkout without a worktree', undefined, { project: 'gavel', worktree: undefined }],
  ])('%s', async (_label, worktree, expected) => {
    stubFetch();

    renderWorktree({ worktree });

    await screen.findByText('feature/projects');
    expect(useGitFocusMock).toHaveBeenCalledWith(expected);
  });

  it('scopes the status request to the worktree', async () => {
    const fetchMock = stubFetch();

    renderWorktree();

    expect(await screen.findByText('feature/projects')).toBeTruthy();
    expect(fetchMock).toHaveBeenCalledWith(`/api/projects/gavel/status?worktree=${ENCODED}`, expect.anything());
  });

  it('keeps the results flag and the worktree in one query string', async () => {
    const fetchMock = stubFetch();

    renderWorktree({ showResults: true });

    await screen.findByText('feature/projects');
    expect(fetchMock).toHaveBeenCalledWith(`/api/projects/gavel/status?includeResults=true&worktree=${ENCODED}`, expect.anything());
  });

  it('scopes the selected file diff to the worktree', async () => {
    const fetchMock = stubFetch(url => url.includes('/diff?') ? { path: 'one.go', diff: '-old\n+new\n' } : undefined);

    renderWorktree({ diffPath: 'one.go' });

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(`/api/projects/gavel/diff?path=one.go&worktree=${ENCODED}`, expect.anything()));
  });

  it('queues the commit with the worktree in the request body', async () => {
    const fetchMock = stubFetch();
    renderWorktree();

    fireEvent.click(await screen.findByRole('checkbox', { name: 'Select one.go' }));
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected (1)' }));

    await waitFor(() => expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(true));
    const call = fetchMock.mock.calls.find(([, init]) => init?.method === 'POST')!;
    expect(call[0]).toBe('/api/projects/gavel/commit-queue');
    expect(JSON.parse(String(call[1]?.body))).toEqual({ action: 'commit', files: ['one.go'], worktree: WORKTREE });
  });

  it('scopes ignoring a directory to the worktree', async () => {
    const fetchMock = stubFetch(url => url.includes('/ignore') ? { path: 'generated', directory: true, rule: '/generated/', added: true } : undefined);
    renderWorktree();

    fireEvent.click(await screen.findByRole('button', { name: 'Ignore directory generated' }));

    await waitFor(() => expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(true));
    const call = fetchMock.mock.calls.find(([, init]) => init?.method === 'POST')!;
    expect(call[0]).toBe(`/api/projects/gavel/ignore?worktree=${ENCODED}`);
    expect(JSON.parse(String(call[1]?.body))).toEqual({ path: 'generated', directory: true });
  });

  it('disables lint and test, which only run in the main checkout, saying why', async () => {
    stubFetch();

    renderWorktree();

    await screen.findByText('feature/projects');
    for (const name of [/Lint project/, /Test changed/]) {
      const button = screen.getByRole('button', { name }) as HTMLButtonElement;
      expect(button.disabled).toBe(true);
      expect(button.title).toMatch(/main checkout/);
    }
  });

  it('leaves the main checkout requests unscoped', async () => {
    const fetchMock = stubFetch();

    renderWorktree({ worktree: undefined });

    await screen.findByText('feature/projects');
    expect(fetchMock).toHaveBeenCalledWith('/api/projects/gavel/status', expect.anything());
  });
});
