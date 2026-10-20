import type React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoRunWorkspace, TodoSessionAttempt } from '../../types';
import { TodoRunBranch } from './TodoRunBranch';
import { latestRunWorkspace } from './runWorkspace';
import { queryTestWrapper } from './queryTestWrapper';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, variant: _variant, size: _size, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: string; size?: string }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>
      {children}
    </button>
  ),
}));

vi.mock('./TodoRunChanges', () => ({
  TodoRunChanges: ({ dir, worktree, commits, landing }: {
    dir: string;
    worktree?: { setup?: string; head?: string };
    commits: Array<{ sha: string }>;
    landing?: { via: string };
  }) => (
    <div data-testid="run-changes">
      {JSON.stringify({ dir, setup: worktree?.setup, head: worktree?.head, commits: commits.map(commit => commit.sha), landing: landing?.via })}
    </div>
  ),
}));

const SETUP = '5e7a0000aaaa1111bbbb2222cccc3333dddd4444';
const HEAD = 'd0b5c966aaaa1111bbbb2222cccc3333dddd4444';

function attempt(ordinal: number, overrides: Partial<TodoSessionAttempt> = {}): TodoSessionAttempt {
  return {
    promptRunId: `run-${ordinal}`,
    ordinal,
    step: 'run',
    requested: {},
    resolved: {},
    status: 'completed',
    processActive: false,
    state: 'succeeded',
    phase: 'finished',
    queuedAt: `2026-09-14T1${ordinal}:00:00Z`,
    admissionSessionId: `admission-${ordinal}`,
    createdAt: `2026-09-14T1${ordinal}:00:00Z`,
    updatedAt: `2026-09-14T1${ordinal}:01:00Z`,
    verification: null,
    ...overrides,
  };
}

const committed: TodoRunWorkspace = {
  worktree: { repo: '/repo', branch: 'shell/289107d9', base: 'ba5e', setup: SETUP, head: HEAD, removed: true },
  commits: [{ sha: HEAD, message: 'feat(todos): record the workspace\n\nGavel-Issue-Id: 42' }],
};

function stubAttempts(attempts: TodoSessionAttempt[], changes = { staged: 0, unstaged: 0, both: 0, untracked: 0, conflict: 0, adds: 0, dels: 0 }) {
  vi.stubGlobal('fetch', vi.fn(async (input: string | URL | Request) => new Response(JSON.stringify(String(input) === '/api/projects'
    ? [{ name: 'repo', dir: '/repo', repos: [] }]
    : String(input) === '/api/projects/repo/git'
      ? {
        base: 'main', currentBranch: 'main', baseCheckedOut: true, branches: [],
        computedAt: '2026-10-20T08:00:00Z', generation: 1,
        worktrees: attempts.flatMap(run => run.workspace?.worktree?.path ? [{
          ...run.workspace.worktree, head: HEAD, primary: false, detached: false, prunable: false,
          changes, ahead: 0, lastCommitAt: '2026-10-20T08:00:00Z', statusScannedAt: '2026-10-20T08:00:00Z',
        }] : []),
      }
      : { attempts }), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  })));
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('latestRunWorkspace', () => {
  it('picks the newest run attempt that recorded a workspace, skipping verify steps and runs without one', () => {
    const newestRun = attempt(4, { workspace: committed });
    const latest = latestRunWorkspace([
      attempt(6, { step: 'verify', workspace: { cwd: '/repo' } }),
      attempt(5),
      newestRun,
      attempt(3, { workspace: { worktree: { branch: 'shell/older' } } }),
    ]);
    expect(latest).toEqual({ attempt: newestRun, workspace: committed });
  });

  it('is null when no run recorded a worktree or commits', () => {
    expect(latestRunWorkspace([attempt(1, { workspace: { cwd: '/repo' } })])).toBeNull();
  });
});

describe('TodoRunBranch', () => {
  it.each(['running', 'failed', 'cancelled'])('shows a %s attempt’s worktree and live changes before any commit', async status => {
    const path = '/repo/.worktrees/active';
    stubAttempts([attempt(1, { status, workspace: { worktree: { repo: '/repo', path, branch: 'feature/active' } } })], {
      staged: 2, unstaged: 1, both: 0, untracked: 1, conflict: 0, adds: 36, dels: 1,
    });
    render(<TodoRunBranch dir="/repo" todoRef="todo-1" />, { wrapper: queryTestWrapper() });
    expect(await screen.findByText(path)).toBeTruthy();
    expect(await screen.findByText('2 staged')).toBeTruthy();
    expect(screen.getByText('1 unstaged')).toBeTruthy();
    expect(screen.getByText('1 untracked')).toBeTruthy();
    expect(screen.getByText('+36')).toBeTruthy();
    expect(screen.getByText('−1')).toBeTruthy();
  });

  it('shows a clean active worktree with zero commits', async () => {
    stubAttempts([attempt(1, { status: 'running', workspace: { worktree: { path: '/repo/.worktrees/clean', branch: 'feature/clean' } } })]);
    render(<TodoRunBranch dir="/repo" todoRef="todo-1" />, { wrapper: queryTestWrapper() });
    expect(await screen.findByText('Working tree clean')).toBeTruthy();
    expect(screen.getByText('0 commits')).toBeTruthy();
  });

  it('retains the path and removed status without requesting live Git state', async () => {
    stubAttempts([attempt(1, { workspace: { worktree: { path: '/repo/.worktrees/removed', branch: 'feature/removed', removed: true } } })]);
    render(<TodoRunBranch dir="/repo" todoRef="todo-1" />, { wrapper: queryTestWrapper() });
    expect(await screen.findByText('Worktree removed')).toBeTruthy();
    expect(screen.getByText('/repo/.worktrees/removed')).toBeTruthy();
    expect(vi.mocked(fetch).mock.calls.some(([url]) => String(url).startsWith('/api/projects'))).toBe(false);
  });

  it('shows the branch chip, setup…head and hands the range, commits and landing to the changes viewer', async () => {
    stubAttempts([attempt(2, {
      workspace: committed,
      landing: { promptRunId: 'run-2', via: 'merge', targetBranch: 'main', landedSha: HEAD, commitCount: 1, landedAt: '2026-09-14T12:00:00Z' },
    })]);

    render(<TodoRunBranch dir="/repo" todoRef="todo-1" />, { wrapper: queryTestWrapper() });

    expect(await screen.findByText('shell/289107d9')).toBeTruthy();
    expect(screen.getByText('5e7a000…d0b5c96')).toBeTruthy();
    expect(screen.queryByRole('alert')).toBeNull();
    expect(JSON.parse(screen.getByTestId('run-changes').textContent ?? '')).toEqual({
      dir: '/repo', setup: SETUP, head: HEAD, commits: [HEAD], landing: 'merge',
    });
  });

  it('shows the changes viewer for a setup..head range even when no commit was recorded', async () => {
    stubAttempts([attempt(2, { workspace: { worktree: { branch: 'shell/3', setup: SETUP, head: HEAD } } })]);

    render(<TodoRunBranch dir="/repo" todoRef="todo-1" />, { wrapper: queryTestWrapper() });

    await screen.findByText('shell/3');
    expect(JSON.parse(screen.getByTestId('run-changes').textContent ?? '').commits).toEqual([]);
  });

  it('omits the changes viewer when the worktree recorded neither commits nor a range', async () => {
    stubAttempts([attempt(2, { workspace: { worktree: { branch: 'shell/4' } } })]);

    render(<TodoRunBranch dir="/repo" todoRef="todo-1" />, { wrapper: queryTestWrapper() });

    await screen.findByText('shell/4');
    expect(screen.queryByTestId('run-changes')).toBeNull();
  });

  it('warns about a kept worktree with its path, reason and dirty count', async () => {
    stubAttempts([attempt(1, {
      workspace: {
        worktree: {
          branch: 'shell/9', setup: SETUP, head: SETUP, path: '/repo/.worktrees/shell-9',
          kept: true, keptReason: 'commit failed', dirty: ['go.mod', 'main.go'],
        },
      },
    })]);

    render(<TodoRunBranch dir="/repo" todoRef="todo-1" />, { wrapper: queryTestWrapper() });

    const warning = await screen.findByRole('alert');
    expect(warning.textContent).toContain('/repo/.worktrees/shell-9');
    expect(warning.textContent).toContain('commit failed');
    expect(warning.textContent).toContain('2 uncommitted paths');
  });

  it('renders the actions slot with the latest run workspace', async () => {
    stubAttempts([attempt(3, { workspace: committed })]);

    render(
      <TodoRunBranch dir="/repo" todoRef="todo-1" renderActions={({ attempt: run, workspace }) => <span>land #{run.ordinal} {workspace.worktree?.branch}</span>} />,
      { wrapper: queryTestWrapper() },
    );

    expect(await screen.findByText('land #3 shell/289107d9')).toBeTruthy();
  });

  it('renders nothing for a todo whose runs recorded no workspace', async () => {
    stubAttempts([attempt(1)]);

    const { container } = render(<TodoRunBranch dir="/repo" todoRef="todo-1" />, { wrapper: queryTestWrapper() });

    await waitFor(() => expect(vi.mocked(fetch)).toHaveBeenCalled());
    expect(container.textContent).toBe('');
  });
});
