import type React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { SessionStats, TodoItem } from '../../types';
import { TodoRow, todoQuery } from './format';

const useSessionStats = vi.hoisted(() => vi.fn(() => ({ stats: null as SessionStats | null, elapsedMs: 0, error: '' })));

// Only the polling hook is stubbed; formatDuration/formatCost stay real so the
// row's rendered text is checked against the shared formatter, not a copy of it.
vi.mock('./TodoSessionTimer', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useSessionStats,
}));

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>
      {children}
    </button>
  ),
  ListMenuItem: ({ children, active: _active, selected: _selected, ...props }: { children?: React.ReactNode; active?: boolean; selected?: boolean }) => (
    <div {...props}>{children}</div>
  ),
}));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  UiListDashes: (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
  UiBeaker: (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
}));

const baseTodo: TodoItem = {
  ref: 'todo-1',
  title: 'Ship the feature',
  status: 'pending',
  priority: 'medium',
};

const SESSION_ID = '7657484f-e2e6-4f71-85c7-c244577a4028';
const RUN_MS = 5 * 60_000 + 15_000;
const RUN_COST = 1.54;

function sessionStats(overrides: Partial<SessionStats> = {}): SessionStats {
  return {
    durationMs: RUN_MS,
    inputTokens: 0,
    outputTokens: 0,
    cacheReadTokens: 0,
    cacheCreationTokens: 0,
    totalTokens: 0,
    contextTokens: 0,
    contextWindow: 0,
    turns: 0,
    compactions: 0,
    costUsd: RUN_COST,
    inProgress: false,
    found: true,
    ...overrides,
  };
}

beforeEach(() => {
  useSessionStats.mockClear();
  useSessionStats.mockReturnValue({ stats: null, elapsedMs: 0, error: '' });
});

describe('todoQuery', () => {
  it('carries only the native workspace directory', () => {
    expect(todoQuery('/work/repo')).toBe('dir=%2Fwork%2Frepo');
    expect(todoQuery('')).toBe('');
  });
});

describe('TodoRow', () => {
  it('shows a plan indicator when hasPlan is true', () => {
    render(<TodoRow todo={{ ...baseTodo, hasPlan: true, hasVerification: false }} active={false} onClick={() => {}} />);

    expect(screen.getByTitle('Plan available')).toBeTruthy();
    expect(screen.queryByTitle('Verification fixture defined')).toBeNull();
  });

  it('shows a verification indicator when hasVerification is true', () => {
    render(<TodoRow todo={{ ...baseTodo, hasPlan: false, hasVerification: true }} active={false} onClick={() => {}} />);

    expect(screen.queryByTitle('Plan available')).toBeNull();
    expect(screen.getByTitle('Verification fixture defined')).toBeTruthy();
  });

  it('shows neither indicator when the todo has no plan or verification fixture', () => {
    render(<TodoRow todo={{ ...baseTodo, hasPlan: false, hasVerification: false }} active={false} onClick={() => {}} />);

    expect(screen.queryByTitle('Plan available')).toBeNull();
    expect(screen.queryByTitle('Verification fixture defined')).toBeNull();
  });

  it('shows the plan indicator, verification indicator, and diff badge together', () => {
    render(
      <TodoRow
        todo={{ ...baseTodo, hasPlan: true, hasVerification: true, diff: { commits: 1, files: 2, adds: 3, dels: 1 } }}
        active={false}
        onClick={() => {}}
      />,
    );

    expect(screen.getByTitle('Plan available')).toBeTruthy();
    expect(screen.getByTitle('Verification fixture defined')).toBeTruthy();
    expect(screen.getByTitle('1 commit, 2 files changed')).toBeTruthy();
  });
});

describe('TodoRow session badge', () => {
  it('shows a settled run duration and cost for a todo that is no longer in progress', () => {
    useSessionStats.mockReturnValue({
      stats: sessionStats({ state: 'completed' }),
      elapsedMs: RUN_MS,
      error: '',
    });

    render(
      <TodoRow
        todo={{ ...baseTodo, status: 'draft', executionState: 'failed', sessionId: SESSION_ID }}
        active={false}
        onClick={() => {}}
        dir="/work/repo"
      />,
    );

    const badge = screen.getByTitle('Done · agent session');
    expect(badge.textContent).toContain('5m 15s');
    expect(badge.textContent).toContain('$1.54');
    expect(screen.queryByTitle('draft')).toBeNull();
  });

  it('does not keep a finished run polling for a session that never appeared', () => {
    render(
      <TodoRow
        todo={{ ...baseTodo, status: 'draft', executionState: 'failed', sessionId: SESSION_ID }}
        active={false}
        onClick={() => {}}
        dir="/work/repo"
      />,
    );

    expect(useSessionStats).toHaveBeenCalledWith(expect.objectContaining({ sessionId: SESSION_ID, expectLive: false }));
  });

  it('falls back to the todo status icon when a settled todo has no session stats', () => {
    render(
      <TodoRow
        todo={{ ...baseTodo, status: 'draft', sessionId: SESSION_ID }}
        active={false}
        onClick={() => {}}
        dir="/work/repo"
      />,
    );

    expect(screen.getByTitle('draft')).toBeTruthy();
    expect(screen.queryByText('In progress')).toBeNull();
  });

  it('keeps the live chip and polling while a running todo has no session stats yet', () => {
    render(
      <TodoRow
        todo={{ ...baseTodo, status: 'in_progress', executionState: 'running', sessionId: SESSION_ID }}
        active={false}
        onClick={() => {}}
        dir="/work/repo"
      />,
    );

    expect(screen.getByTitle('In progress · agent session')).toBeTruthy();
    expect(useSessionStats).toHaveBeenCalledWith(expect.objectContaining({ sessionId: SESSION_ID, expectLive: true }));
  });

  it('never reads session stats for a todo that has no session', () => {
    render(
      <TodoRow todo={{ ...baseTodo, status: 'draft' }} active={false} onClick={() => {}} dir="/work/repo" />,
    );

    expect(useSessionStats).not.toHaveBeenCalled();
    expect(screen.getByTitle('draft')).toBeTruthy();
  });
});
