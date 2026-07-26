import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { CommitQueuePanel, commitQueueLockedFiles, type CommitQueueEntry, type CommitQueueEntryStatus } from './CommitQueuePanel';

const entry = (id: string, files: string[], status: CommitQueueEntryStatus, extra: Partial<CommitQueueEntry> = {}): CommitQueueEntry => ({
  id,
  action: 'commit',
  files,
  status,
  ...extra,
});

describe('CommitQueuePanel', () => {
  it('renders nothing until a group is queued', () => {
    const { container } = render(<CommitQueuePanel queue={{ running: false }} onCancel={vi.fn()} />);
    expect(container.firstChild).toBeNull();
  });

  it('lists groups in queue order with their position and status', () => {
    render(
      <CommitQueuePanel
        queue={{
          runId: 'commit-run-1',
          href: '/tasks/commit-run-1',
          running: true,
          entries: [
            entry('a', ['one.go'], 'success', { exitCode: 0 }),
            entry('b', ['two.go', 'three.go'], 'running'),
            entry('c', ['four.go'], 'pending'),
          ],
        }}
        onCancel={vi.fn()}
      />,
    );

    const rows = within(screen.getByRole('region', { name: 'Commit queue' })).getAllByRole('listitem');
    expect(rows.map(row => within(row).getByRole('button', { expanded: false }).textContent)).toEqual([
      'one.go',
      '2 files',
      'four.go',
    ]);
    expect(rows.map(row => row.textContent?.startsWith(String(rows.indexOf(row) + 1)))).toEqual([true, true, true]);
    expect(within(rows[1]).getByText('running')).toBeTruthy();
    expect(screen.getByText('3 groups, 1 waiting')).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Details' }).getAttribute('href')).toBe('/tasks/commit-run-1');
  });

  it('expands a group to its streamed command output', () => {
    render(
      <CommitQueuePanel
        queue={{ running: false, entries: [entry('a', ['one.go'], 'failed', { exitCode: 1, output: 'staging one.go\nprecommit failed\n', error: 'exit status 1' })] }}
        onCancel={vi.fn()}
      />,
    );

    expect(screen.getByText('Exit 1')).toBeTruthy();
    expect(screen.getByText('exit status 1')).toBeTruthy();
    expect(screen.queryByText(/precommit failed/)).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: 'one.go' }));
    expect(screen.getByText(/precommit failed/)).toBeTruthy();
  });

  it('labels Open PR groups separately from regular commits', () => {
    render(
      <CommitQueuePanel
        queue={{ running: true, entries: [entry('a', ['one.go'], 'running', { action: 'open-pr' })] }}
        onCancel={vi.fn()}
      />,
    );

    expect(screen.getByText('Open PR')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Cancel open PR group 1' })).toBeTruthy();
  });

  it('cancels queued and running groups but offers no cancel once terminal', () => {
    const onCancel = vi.fn();
    render(
      <CommitQueuePanel
        queue={{
          running: true,
          entries: [entry('a', ['one.go'], 'success'), entry('b', ['two.go'], 'running'), entry('c', ['three.go'], 'pending')],
        }}
        onCancel={onCancel}
      />,
    );

    expect(screen.queryByRole('button', { name: 'Cancel commit group 1' })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Cancel commit group 2' }));
    expect(onCancel).toHaveBeenLastCalledWith('b');
    fireEvent.click(screen.getByRole('button', { name: 'Cancel commit group 3' }));
    expect(onCancel).toHaveBeenLastCalledWith('c');
  });
});

describe('commitQueueLockedFiles', () => {
  it('locks the files of groups that have not committed yet, keyed by queue position', () => {
    const locked = commitQueueLockedFiles({
      running: true,
      entries: [
        entry('a', ['done.go'], 'success'),
        entry('b', ['running.go', 'shared.go'], 'running'),
        entry('c', ['queued.go'], 'pending'),
        entry('d', ['canceled.go'], 'canceled'),
      ],
    });

    expect([...locked.entries()]).toEqual([
      ['running.go', 2],
      ['shared.go', 2],
      ['queued.go', 3],
    ]);
  });

  it('locks nothing without a queue', () => {
    expect(commitQueueLockedFiles(undefined).size).toBe(0);
  });
});
