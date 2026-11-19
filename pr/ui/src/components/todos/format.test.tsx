import type React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { SessionStats, TodoItem } from '../../types';
import { TodoActionCounts, TodoCountsBar, TodoRow, emptyCounts, todoQuery } from './format';

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
    <div data-testid="todo-row" {...props}>{children}</div>
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
  it.each(['comfortable', 'compact'] as const)('indicates a linked PR in %s rows without fetching it', density => {
    const pr = { number: 27, url: 'https://github.com/acme/service/pull/27' };
    render(<TodoRow todo={{ ...baseTodo, pr }} density={density} active={false} onSelect={() => {}} />);
    expect(screen.getByTitle(pr.url).textContent).toBe('PR #27');
  });
  it('shows a plan indicator when hasPlan is true', () => {
    render(<TodoRow todo={{ ...baseTodo, hasPlan: true, hasVerification: false }} active={false} onSelect={() => {}} />);

    expect(screen.getByTitle('Plan available')).toBeTruthy();
    expect(screen.queryByTitle('Verification fixture defined')).toBeNull();
  });

  it('shows a verification indicator when hasVerification is true', () => {
    render(<TodoRow todo={{ ...baseTodo, hasPlan: false, hasVerification: true }} active={false} onSelect={() => {}} />);

    expect(screen.queryByTitle('Plan available')).toBeNull();
    expect(screen.getByTitle('Verification fixture defined')).toBeTruthy();
  });

  it('shows neither indicator when the todo has no plan or verification fixture', () => {
    render(<TodoRow todo={{ ...baseTodo, hasPlan: false, hasVerification: false }} active={false} onSelect={() => {}} />);

    expect(screen.queryByTitle('Plan available')).toBeNull();
    expect(screen.queryByTitle('Verification fixture defined')).toBeNull();
  });

  it('shows the plan indicator, verification indicator, and diff badge together', () => {
    render(
      <TodoRow
        todo={{ ...baseTodo, hasPlan: true, hasVerification: true, diff: { commits: 1, files: 2, adds: 3, dels: 1 } }}
        active={false}
        onSelect={() => {}}
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
        todo={{ ...baseTodo, status: 'draft', executionState: 'failed', sessionId: SESSION_ID, sessionSummary: { found: true, inProgress: false, state: 'completed', durationMs: RUN_MS, costUsd: RUN_COST, updatedAt: new Date().toISOString() } }}
        active={false}
        onSelect={() => {}}
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
        onSelect={() => {}}
        dir="/work/repo"
      />,
    );

    expect(useSessionStats).not.toHaveBeenCalled();
  });

  it('falls back to the todo status icon when a settled todo has no session stats', () => {
    render(
      <TodoRow
        todo={{ ...baseTodo, status: 'draft', sessionId: SESSION_ID }}
        active={false}
        onSelect={() => {}}
        dir="/work/repo"
      />,
    );

    expect(screen.getByTitle('draft')).toBeTruthy();
    expect(screen.queryByText('In progress')).toBeNull();
  });

  it('keeps the live chip without fetching while a running todo has no session stats yet', () => {
    render(
      <TodoRow
        todo={{ ...baseTodo, status: 'in_progress', executionState: 'running', sessionId: SESSION_ID }}
        active={false}
        onSelect={() => {}}
        dir="/work/repo"
      />,
    );

    expect(screen.getByTitle('In progress · agent session')).toBeTruthy();
    expect(useSessionStats).not.toHaveBeenCalled();
  });

  it('never reads session stats for a todo that has no session', () => {
    render(
      <TodoRow todo={{ ...baseTodo, status: 'draft' }} active={false} onSelect={() => {}} dir="/work/repo" />,
    );

    expect(useSessionStats).not.toHaveBeenCalled();
    expect(screen.getByTitle('draft')).toBeTruthy();
  });
});

const DAY_MS = 86_400_000;
const daysAgo = (days: number) => new Date(Date.now() - days * DAY_MS).toISOString();

describe('TodoRow on a phone', () => {
  const settledRun = {
    found: true, inProgress: false, state: 'completed' as const, durationMs: RUN_MS, costUsd: RUN_COST, updatedAt: new Date().toISOString(),
  };

  it('leads the title line with the status icon and moves the run duration and cost to the meta line', () => {
    render(
      <TodoRow
        todo={{ ...baseTodo, status: 'review', sessionId: SESSION_ID, sessionSummary: settledRun }}
        active={false}
        onSelect={() => {}}
        dir="/work/repo"
        mobile
      />,
    );

    const titleLine = screen.getByText('Ship the feature').parentElement!;
    expect(titleLine.firstElementChild?.getAttribute('aria-label')).toBe('review');
    expect(titleLine.textContent).not.toContain('5m 15s');
    expect(screen.queryByTitle('Done · agent session')).toBeNull();
    const run = screen.getByTitle('Agent run: 5m 15s, $1.54');
    expect(run.textContent).toBe('5m 15s$1.54');
    expect(titleLine.contains(run)).toBe(false);
  });

  it('shows one compact last-activity age and keeps the created age in its tooltip', () => {
    render(
      <TodoRow
        todo={{ ...baseTodo, created: daysAgo(10), lastRun: daysAgo(3) }}
        active={false}
        onSelect={() => {}}
        mobile
      />,
    );

    const age = screen.getByText('3d');
    expect(age.closest('[title]')?.getAttribute('title')).toMatch(/^Updated .+\nCreated .+$/);
    expect(screen.queryByText(/ago/)).toBeNull();
    expect(screen.queryByText('1w')).toBeNull();
  });

  it('drops the severity edge so the status icon is the row\'s one indicator', () => {
    render(<TodoRow todo={{ ...baseTodo, priority: 'high' }} active={false} onSelect={() => {}} mobile />);
    expect((screen.getByTestId('todo-row') as HTMLElement).style.borderLeftColor).toBe('');
  });

  it('keeps the severity edge on desktop', () => {
    render(<TodoRow todo={{ ...baseTodo, priority: 'high' }} active={false} onSelect={() => {}} />);
    expect((screen.getByTestId('todo-row') as HTMLElement).style.borderLeftColor).not.toBe('');
  });
});

describe('TodoCountsBar', () => {
  it('names the status behind every count for assistive tech', () => {
    render(<TodoCountsBar counts={{ ...emptyCounts, total: 9, open: 9, failed: 7, ask: 2 }} />);
    expect(screen.getByLabelText('7 failed')).toBeTruthy();
    expect(screen.getByLabelText('2 ask')).toBeTruthy();
    expect(screen.getByLabelText('9 open todos')).toBeTruthy();
  });
});

describe('TodoActionCounts', () => {
  const counts = { ...emptyCounts, total: 148, open: 148, draft: 17, failed: 7, ask: 2, review: 1, verified: 40, completed: 51 };

  it('shows only the statuses that need action, each named for assistive tech', () => {
    render(<TodoActionCounts counts={counts} statusFilter={{}} onFocus={() => {}} />);
    expect(screen.getAllByRole('button').map(b => b.getAttribute('aria-label'))).toEqual([
      '7 failed', '2 need input', '1 needs review',
    ]);
    expect(screen.queryByText('148')).toBeNull();
    expect(screen.queryByText('17')).toBeNull();
  });

  it('omits an actionable status with no todos', () => {
    render(<TodoActionCounts counts={{ ...counts, ask: 0 }} statusFilter={{}} onFocus={() => {}} />);
    expect(screen.getAllByRole('button').map(b => b.getAttribute('aria-label'))).toEqual(['7 failed', '1 needs review']);
  });

  it('applies the tapped status as the list filter and marks the chip whose filter is applied', () => {
    const onFocus = vi.fn();
    render(<TodoActionCounts counts={counts} statusFilter={{ failed: 'include' }} onFocus={onFocus} />);
    fireEvent.click(screen.getByLabelText('2 need input'));
    expect(onFocus).toHaveBeenCalledWith('ask');
    expect(screen.getByLabelText('7 failed').getAttribute('aria-pressed')).toBe('true');
    expect(screen.getByLabelText('2 need input').getAttribute('aria-pressed')).toBe('false');
  });
});
