import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoCommitFile, TodoRunCommit, TodoRunLanding } from '../../types';
import { TodoRunChanges } from './TodoRunChanges';
import { queryTestWrapper } from './queryTestWrapper';

vi.mock('@flanksource/clicky-ui/data', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/clicky-ui/data')>()),
  GitDiffPanel: ({ loading, payload, error }: { loading: boolean; payload: { diff: string } | null; error: string }) => (
    <div data-testid="diff-panel">{loading ? 'loading' : error || payload?.diff || 'empty'}</div>
  ),
}));

const SETUP = '5e7a0000aaaa1111bbbb2222cccc3333dddd4444';
const HEAD = 'd0b5c966aaaa1111bbbb2222cccc3333dddd4444';
const FIRST = 'aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111';

const commits: TodoRunCommit[] = [
  { sha: FIRST, message: 'feat(todos): add the picker\n\nbody' },
  { sha: HEAD, message: 'fix(todos): keep the selection' },
];
const worktree = { branch: 'shell/1', setup: SETUP, head: HEAD };

const files: TodoCommitFile[] = [
  { path: 'docs/a.md', status: 'added', adds: 4, dels: 0, language: 'markdown' },
  { path: 'cmd/main.go', status: 'modified', adds: 3, dels: 1, language: 'go', scopes: ['cli'] },
  { path: 'docs/b.md', status: 'deleted', adds: 0, dels: 9 },
];

const landing: TodoRunLanding = {
  promptRunId: 'run-1', via: 'merge', targetBranch: 'main', landedSha: 'abc1234def5678', commitCount: 2, landedAt: '2026-09-14T10:00:00Z',
};

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
}

type Handler = (url: URL) => Response;

function stubChanges(overrides: { files?: Handler; diff?: Handler } = {}) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
    const url = new URL(String(input), 'http://localhost');
    if (url.pathname === '/api/todos/commits/files') {
      return overrides.files?.(url) ?? json({ hash: url.searchParams.get('hash'), files });
    }
    if (url.pathname === '/api/todos/commits/diff') {
      return overrides.diff?.(url) ?? json({ diff: `diff:${url.searchParams.get('file')}` });
    }
    throw new Error(`unexpected request ${url.pathname}`);
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

function requests(fetchMock: ReturnType<typeof stubChanges>, pathname: string) {
  return fetchMock.mock.calls
    .map(([input]) => new URL(String(input), 'http://localhost'))
    .filter(url => url.pathname === pathname)
    .map(url => Object.fromEntries(url.searchParams));
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('TodoRunChanges', () => {
  it('loads the combined setup..head range by default and diffs the first file in tree order', async () => {
    const fetchMock = stubChanges();

    render(<TodoRunChanges dir="/repo" worktree={worktree} commits={commits} />, { wrapper: queryTestWrapper() });

    expect(await screen.findByText('diff:cmd/main.go')).toBeTruthy();
    expect(requests(fetchMock, '/api/todos/commits/files')).toEqual([{ dir: '/repo', hash: HEAD, base: SETUP }]);
    expect(requests(fetchMock, '/api/todos/commits/diff')).toEqual([{ dir: '/repo', hash: HEAD, base: SETUP, file: 'cmd/main.go' }]);
    expect(screen.getByText('2 commits')).toBeTruthy();
    expect((screen.getByRole('combobox', { name: 'Changes to show' }) as HTMLSelectElement).value).toBe('all');
  });

  it('shows each file with its change kind, language, scope and line counts, and directories with a file count', async () => {
    stubChanges();

    render(<TodoRunChanges dir="/repo" worktree={worktree} commits={commits} />, { wrapper: queryTestWrapper() });

    await screen.findByText('main.go');
    expect(screen.getByText('Modified')).toBeTruthy();
    expect(screen.getByText('Added')).toBeTruthy();
    expect(screen.getByText('Deleted')).toBeTruthy();
    expect(screen.getByText('go')).toBeTruthy();
    expect(screen.getByText('cli')).toBeTruthy();
    expect(screen.getByText('+3')).toBeTruthy();
    expect(screen.getByText('−1')).toBeTruthy();
    expect(screen.getByText('1 file')).toBeTruthy();
    expect(screen.getByText('2 files')).toBeTruthy();
  });

  it('drops the base when a single commit is picked and requests that commit', async () => {
    const fetchMock = stubChanges();
    render(<TodoRunChanges dir="/repo" worktree={worktree} commits={commits} />, { wrapper: queryTestWrapper() });
    await screen.findByText('diff:cmd/main.go');

    fireEvent.change(screen.getByRole('combobox', { name: 'Changes to show' }), { target: { value: FIRST } });

    await waitFor(() => expect(requests(fetchMock, '/api/todos/commits/files')).toContainEqual({ dir: '/repo', hash: FIRST }));
    await waitFor(() => expect(requests(fetchMock, '/api/todos/commits/diff')).toContainEqual({ dir: '/repo', hash: FIRST, file: 'cmd/main.go' }));
    expect(screen.getByRole('option', { name: `${FIRST.slice(0, 7)} feat(todos): add the picker` })).toBeTruthy();
  });

  it('requests the clicked file, then a directory as a directory path', async () => {
    const fetchMock = stubChanges();
    render(<TodoRunChanges dir="/repo" worktree={worktree} commits={commits} />, { wrapper: queryTestWrapper() });
    await screen.findByText('diff:cmd/main.go');

    fireEvent.click(screen.getByText('a.md'));
    expect(await screen.findByText('diff:docs/a.md')).toBeTruthy();

    fireEvent.click(screen.getByText('docs'));
    expect(await screen.findByText('diff:docs')).toBeTruthy();
    const files = requests(fetchMock, '/api/todos/commits/diff').map(request => request.file);
    expect(files).toEqual(['cmd/main.go', 'docs/a.md', 'docs']);
  });

  it('keeps the selected path across a picker change while it still exists, else selects the first file', async () => {
    const fetchMock = stubChanges({
      files: url => json({
        hash: url.searchParams.get('hash'),
        files: url.searchParams.get('hash') === FIRST ? files.filter(file => file.path !== 'docs/a.md') : files,
      }),
    });
    render(<TodoRunChanges dir="/repo" worktree={worktree} commits={commits} />, { wrapper: queryTestWrapper() });
    await screen.findByText('diff:cmd/main.go');
    fireEvent.click(screen.getByText('a.md'));
    await screen.findByText('diff:docs/a.md');

    fireEvent.change(screen.getByRole('combobox', { name: 'Changes to show' }), { target: { value: HEAD } });
    await waitFor(() => expect(requests(fetchMock, '/api/todos/commits/diff')).toContainEqual({ dir: '/repo', hash: HEAD, file: 'docs/a.md' }));

    fireEvent.change(screen.getByRole('combobox', { name: 'Changes to show' }), { target: { value: FIRST } });
    await waitFor(() => expect(requests(fetchMock, '/api/todos/commits/diff')).toContainEqual({ dir: '/repo', hash: FIRST, file: 'cmd/main.go' }));
    expect(requests(fetchMock, '/api/todos/commits/diff')).not.toContainEqual({ dir: '/repo', hash: FIRST, file: 'docs/a.md' });
  });

  it('offers no "All changes" entry without a setup..head range and defaults to the first commit', async () => {
    const fetchMock = stubChanges();

    render(<TodoRunChanges dir="/repo" worktree={{ branch: 'shell/1', head: HEAD }} commits={commits} />, { wrapper: queryTestWrapper() });

    expect(await screen.findByText('diff:cmd/main.go')).toBeTruthy();
    expect(screen.queryByRole('option', { name: 'All changes' })).toBeNull();
    expect(requests(fetchMock, '/api/todos/commits/files')).toEqual([{ dir: '/repo', hash: FIRST }]);
  });

  it('shows only "All changes" when the run has a range but recorded no commits', async () => {
    stubChanges();

    render(<TodoRunChanges dir="/repo" worktree={worktree} commits={[]} />, { wrapper: queryTestWrapper() });

    expect(await screen.findByText('diff:cmd/main.go')).toBeTruthy();
    expect(screen.getAllByRole('option').map(option => option.textContent)).toEqual(['All changes']);
  });

  it('explains unreachable commits on HTTP 410 and shows how they landed', async () => {
    stubChanges({ files: () => json({ error: 'commit is gone' }, 410), diff: () => json({ error: 'commit is gone' }, 410) });

    render(<TodoRunChanges dir="/repo" worktree={worktree} commits={commits} landing={landing} />, { wrapper: queryTestWrapper() });

    expect(await screen.findByText('Commits no longer reachable')).toBeTruthy();
    expect(screen.getByText('Merged into')).toBeTruthy();
    expect(screen.getByText('main')).toBeTruthy();
    expect(screen.getByText('abc1234')).toBeTruthy();
    expect(screen.queryByTestId('diff-panel')).toBeNull();
  });

  it('explains unreachable commits without landing details when none were recorded', async () => {
    stubChanges({ files: () => json({ error: 'commit is gone' }, 410) });

    render(<TodoRunChanges dir="/repo" worktree={worktree} commits={commits} />, { wrapper: queryTestWrapper() });

    expect(await screen.findByText('Commits no longer reachable')).toBeTruthy();
    expect(screen.queryByText('Merged into')).toBeNull();
  });

  it('renders other file-list failures as an error, not the unreachable notice', async () => {
    stubChanges({ files: () => json({ error: 'git exploded' }, 500) });

    render(<TodoRunChanges dir="/repo" worktree={worktree} commits={commits} />, { wrapper: queryTestWrapper() });

    const message = await screen.findByText(/git exploded/);
    expect(message.className).toContain('text-red-600');
    expect(screen.queryByText('Commits no longer reachable')).toBeNull();
  });

  it('passes a diff failure to the diff panel while keeping the tree', async () => {
    stubChanges({ diff: () => json({ error: 'diff exploded' }, 500) });

    render(<TodoRunChanges dir="/repo" worktree={worktree} commits={commits} />, { wrapper: queryTestWrapper() });

    await waitFor(() => expect(screen.getByTestId('diff-panel').textContent).toContain('diff exploded'));
    expect(screen.getByText('main.go')).toBeTruthy();
  });

  it('says so when the range changed no files', async () => {
    stubChanges({ files: url => json({ hash: url.searchParams.get('hash'), files: [] }) });

    render(<TodoRunChanges dir="/repo" worktree={worktree} commits={commits} />, { wrapper: queryTestWrapper() });

    expect(await screen.findByText('No file changes')).toBeTruthy();
  });
});
