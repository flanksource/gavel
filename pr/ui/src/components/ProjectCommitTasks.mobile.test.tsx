import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, cleanup, fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { CommitStatusSlotContext } from './commitStatusSlot';
import { ProjectCommitTasks } from './ProjectCommitTasks';

const run = vi.hoisted(() => ({ status: 'success', complete: true }));
const CONFIRMATION_MS = 4000;

beforeEach(() => {
  run.status = 'success';
  run.complete = true;
  vi.useFakeTimers();
});
afterEach(() => { cleanup(); vi.useRealTimers(); });

vi.mock('../useIsMobile', () => ({ useIsMobile: () => true }));
vi.mock('@flanksource/clicky-ui/hooks', async importOriginal => ({
  ...await importOriginal<typeof import('@flanksource/clicky-ui/hooks')>(),
  useTaskRuns: () => ({ runs: [{ id: 'finished-commit' }], status: '' }),
  useTaskRun: () => ({
    snapshots: [{ id: 'finished-commit', groupId: 'finished-commit', name: 'Commit project', type: 'group', status: run.status }],
    isComplete: run.complete, status: '',
  }),
}));

const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
const tasks = (preferredRunId?: string) => (
  <QueryClientProvider client={client}>
    <ProjectCommitTasks projectName="example" preferredRunId={preferredRunId} onLockedFilesChange={vi.fn()} onErrorChange={vi.fn()} onComplete={vi.fn()} onRunChange={vi.fn()} />
  </QueryClientProvider>
);

describe('mobile commit run', () => {
  it.each([
    { status: 'running', complete: false },
    { status: 'failed', complete: true },
  ])('keeps $status run details visible with no collapse chip', ({ status, complete }) => {
    run.status = status;
    run.complete = complete;
    render(tasks());

    expect(screen.getByText('Commit project')).toBeTruthy();
    act(() => { vi.advanceTimersByTime(CONFIRMATION_MS * 2); });
    expect(screen.getByText('Commit project')).toBeTruthy();
    expect(screen.queryByRole('button', { name: /commit details/ })).toBeNull();
  });

  it('collapses a run that finished before it was opened without a confirmation, keeping its details one tap away', () => {
    render(tasks());

    const chip = screen.getByRole('button', { name: 'Show commit details' });
    expect(chip.getAttribute('aria-expanded')).toBe('false');
    expect(screen.queryByRole('status')).toBeNull();
    expect(screen.queryByText('Commit project')).toBeNull();

    fireEvent.click(chip);

    expect(screen.getByRole('button', { name: 'Hide commit details' }).getAttribute('aria-expanded')).toBe('true');
    expect(screen.getByText('Commit project')).toBeTruthy();
  });

  it.each([undefined, 'finished-commit'])('confirms a run that succeeds while watched (preferred run: %s), then auto-collapses to the chip', preferredRunId => {
    run.status = 'running';
    run.complete = false;
    const { rerender } = render(tasks(preferredRunId));
    expect(screen.getByText('Commit project')).toBeTruthy();

    run.status = 'success';
    run.complete = true;
    rerender(tasks(preferredRunId));

    expect(screen.getByRole('status').textContent).toBe('Commit complete');
    expect(screen.queryByText('Commit project')).toBeNull();
    act(() => { vi.advanceTimersByTime(CONFIRMATION_MS - 1); });
    expect(screen.getByRole('status')).toBeTruthy();
    act(() => { vi.advanceTimersByTime(1); });

    expect(screen.queryByRole('status')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Show commit details' }));
    expect(screen.getByText('Commit project')).toBeTruthy();
  });

  it('puts the chip in the ref bar slot and leaves no row behind once collapsed', () => {
    const slot = document.createElement('span');
    document.body.append(slot);
    const { container } = render(<CommitStatusSlotContext.Provider value={slot}>{tasks()}</CommitStatusSlotContext.Provider>);

    expect(within(slot).getByRole('button', { name: 'Show commit details' })).toBeTruthy();
    expect(container.querySelector('section')).toBeNull();

    fireEvent.click(within(slot).getByRole('button', { name: 'Show commit details' }));
    expect(container.querySelector('section')?.textContent).toContain('Commit project');
    slot.remove();
  });
});
