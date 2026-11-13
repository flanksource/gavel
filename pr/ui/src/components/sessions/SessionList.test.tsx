import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { AgentSession } from '../../types';
import { timeAgo } from '../../utils';
import { SessionList } from './SessionList';

const session: AgentSession = {
  id: 'finished-session', title: 'Plan the page-runtime switch', source: 'claude', lifecycleStatus: 'completed',
  activityState: 'idle', processActive: false, lastActivityAt: '2026-10-11T12:00:00Z', project: 'facet', cwd: '/src/facet',
  worktree: { path: '/src/facet', branch: 'main', base: 'main', primary: true },
  tokens: { input: 0, output: 0, cacheRead: 0, total: 0 }, costUsd: 33.01,
};

describe('SessionList stacked row', () => {
  it('ends the second line with the last activity and the state, right of the checkout', () => {
    render(<SessionList families={[{ session, children: [] }]} selectedId="" onSelect={() => {}} />);

    const state = screen.getByRole('img', { name: 'Finished' });
    const trailing = state.parentElement!;
    expect(trailing.lastElementChild).toBe(state);
    expect(trailing.className).toContain('ml-auto');
    expect(trailing.textContent).toBe(timeAgo(session.lastActivityAt));
    expect(trailing.previousElementSibling?.textContent).toContain('main');
  });
});
