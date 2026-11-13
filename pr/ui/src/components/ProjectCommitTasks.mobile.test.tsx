import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ProjectCommitTasks } from './ProjectCommitTasks';

const run = vi.hoisted(() => ({ status: 'success', complete: true }));
beforeEach(() => { run.status = 'success'; run.complete = true; });

vi.mock('../useIsMobile', () => ({ useIsMobile: () => true }));
vi.mock('@flanksource/clicky-ui/hooks', async importOriginal => ({
  ...await importOriginal<typeof import('@flanksource/clicky-ui/hooks')>(),
  useTaskRuns: () => ({ runs: [{ id: 'finished-commit' }], status: '' }),
  useTaskRun: () => ({
    snapshots: [{ id: 'finished-commit', groupId: 'finished-commit', name: 'Commit project', type: 'group', status: run.status }],
    isComplete: run.complete, status: '',
  }),
}));

describe('mobile commit history', () => {
  it.each([
    { status: 'running', complete: false, preferredRunId: undefined },
    { status: 'failed', complete: true, preferredRunId: undefined },
    { status: 'success', complete: true, preferredRunId: 'finished-commit' },
  ])('keeps $status run details visible (current: $preferredRunId)', async ({ status, complete, preferredRunId }) => {
    run.status = status;
    run.complete = complete;
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(<QueryClientProvider client={client}>
      <ProjectCommitTasks projectName="example" preferredRunId={preferredRunId} onLockedFilesChange={vi.fn()} onErrorChange={vi.fn()} onComplete={vi.fn()} onRunChange={vi.fn()} />
    </QueryClientProvider>);
    expect(await screen.findByText('Commit project')).toBeTruthy();
    expect(screen.queryByRole('button', { name: /Commit complete/ })).toBeNull();
  });

  it('collapses an old successful run and lets its details expand', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(<QueryClientProvider client={client}>
      <ProjectCommitTasks projectName="example" onLockedFilesChange={vi.fn()} onErrorChange={vi.fn()} onComplete={vi.fn()} onRunChange={vi.fn()} />
    </QueryClientProvider>);

    const toggle = await screen.findByRole('button', { name: /Commit complete/ });
    expect(toggle.getAttribute('aria-expanded')).toBe('false');
    expect(screen.queryByText('Commit project')).toBeNull();
    fireEvent.click(toggle);
    expect(toggle.getAttribute('aria-expanded')).toBe('true');
    expect(await screen.findByText('Commit project')).toBeTruthy();
  });
});
