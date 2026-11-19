import { render, screen, within } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, expect, it, vi } from 'vitest';
import type { AgentSession } from '../../types';
import { SessionDetail } from './SessionDetail';

vi.mock('@flanksource/clicky-ui/ai', () => ({
  formatCost: (usd: number) => `$${usd.toFixed(2)}`,
  ContextMeter: () => <div data-testid="header-context-meter" />,
  SessionInspector: ({ hostTabs, session: input }: {
    hostTabs?: { id: string; label: string }[];
    session?: { currentSessionId: string; sessions: { id: string; parentId?: string }[] };
  }) => (
    <div
      data-testid="inspector"
      data-current={input?.currentSessionId}
      data-items={input?.sessions.map(item => `${item.id}<${item.parentId ?? ''}`).join(',')}
    >
      {(hostTabs ?? []).map(tab => <span key={tab.id}>{tab.label}</span>)}
    </div>
  ),
}));
vi.mock('../BranchLandActions', () => ({ BranchLandActions: () => <div data-testid="land-actions" /> }));

const clean = { staged: 0, unstaged: 0, both: 0, untracked: 0, conflict: 0, adds: 0, dels: 0 };
const dirtyFiles = 221;

function session(overrides: Partial<AgentSession> = {}): AgentSession {
  return {
    id: 'session-1', title: 'Export the permission matrix', project: 'mission-control-oipa', source: 'codex',
    model: 'gpt-6.2-sol', effort: 'xhigh', lifecycleStatus: 'running', activityState: 'working', processActive: true,
    pid: 59126, lastActivityAt: '2026-11-18T21:58:00Z', durationMs: 6 * 3_600_000 + 21 * 60_000,
    context: { usedTokens: 44_000, windowTokens: 100_000, freePercent: 56 },
    tokens: { input: 0, output: 0, cacheRead: 0, total: 0 }, costUsd: 19.09,
    worktree: { path: '/repo', branch: 'main', base: 'main', primary: true },
    git: { changes: { ...clean, untracked: dirtyFiles }, ahead: 0, behind: 0, baseCheckedOut: true, worktreeLive: true },
    ...overrides,
  };
}

function renderDetail(value: AgentSession, sessions: AgentSession[] = [value]) {
  return render(
    <QueryClientProvider client={new QueryClient()}>
      <SessionDetail session={value} sessions={sessions} onOpenTodo={vi.fn()} onOpenWorktree={vi.fn()} />
    </QueryClientProvider>,
  );
}

describe('SessionDetail', () => {
  it("hands the inspector the session with its sub-sessions nested under it, for the inspector's menu", () => {
    const parent = session();
    const sub = session({ id: 'sub', parentId: 'session-1', title: 'Subagent' });
    const nested = session({ id: 'nested', parentId: 'sub', title: 'Nested subagent' });
    const unrelated = session({ id: 'other', title: 'Other session' });
    renderDetail(parent, [unrelated, nested, parent, sub]);

    const inspector = screen.getByTestId('inspector');
    expect(inspector.getAttribute('data-current')).toBe('session-1');
    expect(inspector.getAttribute('data-items')).toBe('session-1<,sub<session-1,nested<sub');
  });

  it('leaves the model and context meter to the transcript toolbar instead of repeating them in the header', () => {
    renderDetail(session());
    const header = screen.getByRole('banner');
    expect(screen.queryByTestId('header-context-meter')).toBeNull();
    expect(within(header).queryByText(/gpt-6\.2-sol/)).toBeNull();
    expect(within(header).getByText('$19.09')).toBeTruthy();
  });

  it('hides merge and PR actions while nothing is ahead of the base', () => {
    renderDetail(session());
    expect(screen.queryByTestId('land-actions')).toBeNull();
  });

  it('offers merge and PR actions once the branch is ahead of its base', () => {
    renderDetail(session({
      worktree: { path: '/wt', branch: 'feat', base: 'main', primary: false },
      git: { changes: clean, ahead: 2, behind: 0, baseCheckedOut: true, worktreeLive: true },
    }));
    expect(screen.getByTestId('land-actions')).toBeTruthy();
  });

  it('puts Branch changes in the inspector toolbar, only when the branch has commits ahead', () => {
    const { unmount } = renderDetail(session());
    expect(screen.queryByRole('navigation', { name: 'Session detail' })).toBeNull();
    expect(within(screen.getByTestId('inspector')).queryByText('Branch changes')).toBeNull();
    unmount();

    renderDetail(session({
      worktree: { path: '/wt', branch: 'feat', base: 'main', primary: false },
      git: { changes: clean, ahead: 2, behind: 0, baseCheckedOut: true, worktreeLive: true },
    }));
    expect(within(screen.getByTestId('inspector')).getByText('Branch changes')).toBeTruthy();
  });

  it('puts conflicts and abandoned changes on the one-line work state instead of repeating them in the header', () => {
    renderDetail(session({
      processActive: false,
      lifecycleStatus: 'completed',
      activityState: 'idle',
      git: { changes: { ...clean, unstaged: 3, conflict: 1, adds: 40, dels: 2 }, ahead: 0, behind: 2, baseCheckedOut: true, worktreeLive: true },
    }));

    const line = screen.getByRole('region', { name: 'Work state' });
    const item = (label: string) => line.querySelector(`[data-work-item="${label}"]`)?.textContent;
    expect({
      uncommitted: item('Uncommitted'),
      unmerged: item('Unmerged'),
      pr: item('Pull request'),
      headerWarnings: within(screen.getByRole('banner')).queryByText(/conflicts|left behind/),
    }).toEqual({
      uncommitted: '1 conflicted4 uncommitted+40−2left behind',
      unmerged: '↑0 ↓2',
      pr: 'no PR',
      headerWarnings: null,
    });
  });

  it('summarises uncommitted work as a header toggle for the work-state panel', () => {
    renderDetail(session());
    const toggle = screen.getByRole('button', { name: `Work state: ${dirtyFiles} uncommitted` });
    expect(toggle.getAttribute('aria-expanded')).toBe('false');
  });
});
