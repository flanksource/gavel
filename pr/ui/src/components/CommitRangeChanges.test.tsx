import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoCommitFile } from '../types';
import { CommitRangeChanges, type CommitRangeRequest } from './CommitRangeChanges';
import { queryTestWrapper } from './todos/queryTestWrapper';

vi.mock('@flanksource/clicky-ui/data', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/clicky-ui/data')>()),
  GitDiffPanel: ({ loading, payload, error }: { loading: boolean; payload: { diff: string } | null; error: string }) => (
    <div data-testid="diff-panel">{loading ? 'loading' : error || payload?.diff || 'empty'}</div>
  ),
}));

const files: TodoCommitFile[] = [
  { path: 'docs/a.md', status: 'added', adds: 4, dels: 0 },
  { path: 'cmd/main.go', status: 'modified', adds: 3, dels: 1, language: 'go' },
];

const options = [{ id: 'all', label: 'All changes vs main' }, { id: 'abc', label: 'abc1234 first commit' }];

const filesRequest = (id: string): CommitRangeRequest => ({
  queryKey: ['test', 'files', id],
  url: `/api/example/files?range=${id}`,
  context: `Failed to load files for ${id}`,
});

const diffRequest = (id: string, file: string): CommitRangeRequest => ({
  queryKey: ['test', 'diff', id, file],
  url: `/api/example/diff?range=${id}&file=${encodeURIComponent(file)}`,
  context: `Failed to load ${file} diff for ${id}`,
});

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
}

function stub(overrides: { files?: (url: URL) => Response; diff?: (url: URL) => Response } = {}) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
    const url = new URL(String(input), 'http://localhost');
    if (url.pathname === '/api/example/files') return overrides.files?.(url) ?? json({ files });
    if (url.pathname === '/api/example/diff') return overrides.diff?.(url) ?? json({ diff: `diff:${url.searchParams.get('file')}` });
    throw new Error(`unexpected request ${url.pathname}`);
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

const requested = (fetchMock: ReturnType<typeof stub>) => fetchMock.mock.calls.map(([input]) => String(input));

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('CommitRangeChanges', () => {
  it('loads the files and the first file diff from the caller-built URLs', async () => {
    const fetchMock = stub();

    render(<CommitRangeChanges options={options} filesRequest={filesRequest} diffRequest={diffRequest} />, { wrapper: queryTestWrapper() });

    expect(await screen.findByText('diff:cmd/main.go')).toBeTruthy();
    expect(requested(fetchMock)).toEqual([
      '/api/example/files?range=all',
      '/api/example/diff?range=all&file=cmd%2Fmain.go',
    ]);
  });

  it('requests the files and diff of the range picked in the picker', async () => {
    const fetchMock = stub();
    render(<CommitRangeChanges options={options} filesRequest={filesRequest} diffRequest={diffRequest} />, { wrapper: queryTestWrapper() });
    await screen.findByText('diff:cmd/main.go');

    fireEvent.change(screen.getByRole('combobox', { name: 'Changes to show' }), { target: { value: 'abc' } });

    await waitFor(() => expect(requested(fetchMock)).toContain('/api/example/diff?range=abc&file=cmd%2Fmain.go'));
    expect(requested(fetchMock)).toContain('/api/example/files?range=abc');
  });

  it('requests the clicked file', async () => {
    const fetchMock = stub();
    render(<CommitRangeChanges options={options} filesRequest={filesRequest} diffRequest={diffRequest} />, { wrapper: queryTestWrapper() });
    await screen.findByText('diff:cmd/main.go');

    fireEvent.click(screen.getByText('a.md'));

    expect(await screen.findByText('diff:docs/a.md')).toBeTruthy();
    expect(requested(fetchMock)).toContain('/api/example/diff?range=all&file=docs%2Fa.md');
  });

  it('hides the picker when asked to and shows the summary beside the files', async () => {
    stub();

    render(
      <CommitRangeChanges
        options={options.slice(0, 1)}
        showPicker={false}
        summary="2 commits"
        filesRequest={filesRequest}
        diffRequest={diffRequest}
      />,
      { wrapper: queryTestWrapper() },
    );

    expect(await screen.findByText('diff:cmd/main.go')).toBeTruthy();
    expect(screen.queryByRole('combobox')).toBeNull();
    expect(screen.getByText('2 commits')).toBeTruthy();
  });

  it('shows the caller\'s unreachable notice on HTTP 410 instead of the tree', async () => {
    stub({ files: () => json({ error: 'gone' }, 410), diff: () => json({ error: 'gone' }, 410) });

    render(
      <CommitRangeChanges options={options} unreachable={<span>branch was deleted</span>} filesRequest={filesRequest} diffRequest={diffRequest} />,
      { wrapper: queryTestWrapper() },
    );

    expect(await screen.findByText('branch was deleted')).toBeTruthy();
    expect(screen.queryByTestId('diff-panel')).toBeNull();
  });

  it('renders other failures as an error, not the unreachable notice', async () => {
    stub({ files: () => json({ error: 'git exploded' }, 500) });

    render(
      <CommitRangeChanges options={options} unreachable={<span>branch was deleted</span>} filesRequest={filesRequest} diffRequest={diffRequest} />,
      { wrapper: queryTestWrapper() },
    );

    expect((await screen.findByText(/git exploded/)).className).toContain('text-red-600');
    expect(screen.queryByText('branch was deleted')).toBeNull();
  });

  it('rejects a files response without a files list', async () => {
    stub({ files: () => json({ hash: 'x' }) });

    render(<CommitRangeChanges options={options} filesRequest={filesRequest} diffRequest={diffRequest} />, { wrapper: queryTestWrapper() });

    expect(await screen.findByText(/response has no files list/)).toBeTruthy();
  });
});
