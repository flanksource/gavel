import { useState } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ProjectStatusView } from './ProjectStatusView';
import { projectStatusResponse } from './ProjectStatusView.fixture';

vi.mock('../useIsMobile', () => ({ useIsMobile: () => true }));
vi.mock('../useGitFocus', () => ({ useGitFocus: () => ({ error: '' }) }));
vi.mock('./ProjectCommitTasks', () => ({ ProjectCommitTasks: () => null }));
vi.mock('./ProjectDiffView', () => ({ ProjectDiffView: ({ path }: { path: string }) => <div>Reviewing {path}</div> }));

afterEach(() => vi.unstubAllGlobals());

function Review({ worktree }: { worktree?: string }) {
  const [path, setPath] = useState('');
  return <ProjectStatusView project={projectStatusResponse.project} worktree={worktree} diffPath={path} onDiffPathChange={setPath} />;
}

describe('mobile project commit selection', () => {
  it('offers a way back to files when the status request fails during a diff deep link', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ error: 'status unavailable' }), {
      status: 503, headers: { 'Content-Type': 'application/json' },
    })));
    const onDiffPathChange = vi.fn();
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(<QueryClientProvider client={client}><ProjectStatusView project={projectStatusResponse.project} diffPath="one.go" onDiffPathChange={onDiffPathChange} /></QueryClientProvider>);
    await screen.findByText(/status unavailable/);
    fireEvent.click(screen.getByRole('button', { name: 'Back to files' }));
    expect(onDiffPathChange).toHaveBeenCalledWith('');
  });

  it.each([undefined, '/work/feature'])('queues the reviewed selection in the correct checkout (%s)', async worktree => {
    const fetchMock = vi.fn(async (_url: string, init?: RequestInit) => new Response(JSON.stringify(
      init?.method === 'POST' ? { runId: 'commit-review' } : projectStatusResponse,
    ), { status: 200, headers: { 'Content-Type': 'application/json' } }));
    vi.stubGlobal('fetch', fetchMock);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(<QueryClientProvider client={client}><Review worktree={worktree} /></QueryClientProvider>);

    const row = await screen.findByRole('treeitem', { name: 'one.go' });
    fireEvent.click(within(row).getByText('one.go'));
    expect(await screen.findByText('Reviewing one.go')).toBeTruthy();
    fireEvent.click(screen.getByRole('checkbox', { name: 'Include one.go in commit' }));
    fireEvent.click(screen.getByRole('button', { name: 'Back to files' }));
    expect((await screen.findByRole('checkbox', { name: 'Select one.go' }) as HTMLInputElement).checked).toBe(true);
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected (1)' }));

    await waitFor(() => expect(fetchMock.mock.calls.filter(([, init]) => init?.method === 'POST')).toHaveLength(1));
    const request = fetchMock.mock.calls.find(([, init]) => init?.method === 'POST');
    expect(request?.[0]).toBe('/api/projects/gavel/commit-queue');
    expect(JSON.parse(String(request?.[1]?.body))).toEqual({ action: 'commit', files: ['one.go'], ...(worktree ? { worktree } : {}) });
  });
});
