import { fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { AgentSession } from '../../types';
import type { SessionFamily } from './sessionView';
import { SessionList } from './SessionList';

const NOW = new Date('2026-10-11T14:00:00Z');
const TWO_HOURS_AGO = '2026-10-11T12:00:00Z';
const BELOW_BREAKPOINT_HIDDEN = '@max-[760px]:hidden';
const ABOVE_BREAKPOINT_HIDDEN = '@min-[760px]:hidden';

const changes = { staged: 1, unstaged: 2, both: 0, untracked: 0, conflict: 0, adds: 15644, dels: 5190 };

function makeSession(overrides: Partial<AgentSession> & Pick<AgentSession, 'id' | 'title'>): AgentSession {
  return {
    source: 'claude', model: 'claude-opus-5-5', lifecycleStatus: 'completed', activityState: 'idle', processActive: false,
    lastActivityAt: TWO_HOURS_AGO, project: 'mission-control-oipa', cwd: '/src/mission-control-oipa',
    worktree: { path: '/src/mission-control-oipa', branch: 'main', base: 'main', primary: true },
    tokens: { input: 0, output: 0, cacheRead: 0, total: 0 }, costUsd: 8.66,
    ...overrides,
  };
}

const parent = makeSession({
  id: 'parent', title: 'Global settings git cache',
  git: { changes, ahead: 0, behind: 0, baseCheckedOut: true, worktreeLive: true },
});

function byClass(token: string): string {
  return `[class~="${token}"]`;
}

function renderList(families: SessionFamily[]) {
  return render(<SessionList families={families} selectedId="" onSelect={() => {}} />);
}

function row(sessionId: string): HTMLElement {
  const item = document.querySelector<HTMLElement>(`[data-session-id="${sessionId}"]`);
  if (!item) throw new Error(`no row for session ${sessionId}`);
  return item.querySelector<HTMLElement>('button[aria-label^="View"]')!;
}

function metaLine(sessionId: string): HTMLElement {
  return within(row(sessionId)).getByTestId('session-meta');
}

beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] });
  vi.setSystemTime(NOW);
});

afterEach(() => {
  vi.useRealTimers();
});

describe('SessionList stacked row', () => {
  it('ends the narrow meta line with the compact age only, and keeps the cost off the title line', () => {
    renderList([{ session: parent, children: [] }]);

    const trailing = within(metaLine('parent')).getByTitle("This session's last activity");
    expect(trailing.textContent).toBe('2h');
    expect(trailing.closest(byClass(ABOVE_BREAKPOINT_HIDDEN))).not.toBeNull();
    expect(within(row('parent')).getByTestId('session-title-line').textContent).not.toContain('$');
  });

  it('shows the cost in every layout as a badge right after the git changes, not a right-aligned column', () => {
    renderList([{ session: parent, children: [] }]);

    const badge = within(metaLine('parent')).getByTestId('session-cost');
    expect({
      text: badge.textContent,
      follows: badge.previousElementSibling?.getAttribute('data-testid'),
      layoutGated: badge.matches(`${byClass(BELOW_BREAKPOINT_HIDDEN)},${byClass(ABOVE_BREAKPOINT_HIDDEN)}`),
      rightAligned: badge.matches(byClass('ml-auto')),
      columns: [...row('parent').children].filter(child => child.textContent?.trim() === '$8.66').length,
    }).toEqual({ text: '$8.66', follows: 'session-location', layoutGated: false, rightAligned: false, columns: 0 });
  });

  it('folds the state into the left icon as the only narrow-layout indicator', () => {
    renderList([{ session: makeSession({ id: 'failed', title: 'Broke', lifecycleStatus: 'failed' }), children: [] }]);

    const states = within(row('failed')).getAllByRole('img', { name: 'Failed' });
    expect(states).toHaveLength(1);
    expect(within(row('failed')).getByTestId('session-context').contains(states[0]!)).toBe(true);
    expect(states[0]!.closest(byClass(ABOVE_BREAKPOINT_HIDDEN))).not.toBeNull();
    expect(within(metaLine('failed')).queryByText('Failed')).toBeNull();
  });

  it('distinguishes needs-input, failed and finished in the left icon', () => {
    renderList([
      { session: makeSession({ id: 'ask', title: 'Ask', activityState: 'ask' }), children: [] },
      { session: makeSession({ id: 'fail', title: 'Fail', lifecycleStatus: 'interrupted' }), children: [] },
      { session: makeSession({ id: 'done', title: 'Done' }), children: [] },
    ]);

    expect(within(row('ask')).getByRole('img', { name: 'Needs input' })).toBeTruthy();
    expect(within(row('fail')).getByRole('img', { name: 'Failed' })).toBeTruthy();
    expect(within(row('done')).getByRole('img', { name: 'Finished' })).toBeTruthy();
  });

  it('shows the wide state column through a wrapper that carries no base display utility', () => {
    renderList([{ session: parent, children: [] }]);

    const column = within(row('parent')).getByTestId('session-wide-state');
    expect(column.className.split(' ')).toEqual([BELOW_BREAKPOINT_HIDDEN]);
    expect(within(column).getByText('Finished')).toBeTruthy();
  });

  it('hides the model badge below the breakpoint and keeps the model in the title tooltip', () => {
    renderList([{ session: parent, children: [] }]);

    const badge = within(row('parent')).getByText('claude-opus-5-5');
    expect(badge.closest(byClass(BELOW_BREAKPOINT_HIDDEN))).not.toBeNull();
    expect(within(row('parent')).getByText(parent.title).getAttribute('title')).toBe(`${parent.title} · claude-opus-5-5`);
  });

  it('truncates the project, clips the location and drops the line counts below the breakpoint', () => {
    renderList([{ session: parent, children: [] }]);

    const location = within(metaLine('parent')).getByTestId('session-location');
    expect(location.className).toContain('overflow-hidden');
    const project = within(location).getByText('mission-control-oipa');
    expect(project.className).toContain('truncate');
    expect(project.className).not.toContain('shrink-0');
    expect(project.parentElement!.firstElementChild!.getAttribute('aria-label')).toBe('Repository');

    const adds = within(location).getByText('+15644');
    expect(adds.closest(byClass(BELOW_BREAKPOINT_HIDDEN))).not.toBeNull();
    const files = within(location).getByText('· 3f');
    expect(files.closest(byClass(BELOW_BREAKPOINT_HIDDEN))).toBeNull();
  });
});

describe('SessionList sub-sessions', () => {
  const ONE_HOUR_AGO = '2026-10-11T13:00:00Z';
  const child = makeSession({ id: 'child', parentId: 'parent', title: 'Helper subagent', lastActivityAt: ONE_HOUR_AGO });
  const grandchild = makeSession({ id: 'grandchild', parentId: 'child', title: 'Nested helper' });
  const family = { session: parent, children: [{ session: child, children: [{ session: grandchild, children: [] }] }] };

  it('lists only top-level sessions, with no rows or toggles for their sub-sessions', () => {
    renderList([family]);

    expect(row('parent')).toBeTruthy();
    expect(document.querySelector('[data-session-id="child"]')).toBeNull();
    expect(document.querySelector('[data-session-id="grandchild"]')).toBeNull();
    expect(screen.queryByText('Helper subagent')).toBeNull();
    expect(screen.queryByRole('button', { name: /^(Expand|Collapse) / })).toBeNull();
  });

  it("dates a session by the latest activity across its sub-sessions", () => {
    renderList([family]);

    expect(within(metaLine('parent')).getByTitle('Latest activity across this session and its sub-sessions').textContent).toContain('1h');
  });

  it('keeps the parent row when the search matches only a sub-session', () => {
    window.history.replaceState(null, '', '/sessions?sessions.search=nested');
    renderList([family, { session: makeSession({ id: 'other', title: 'Unrelated work' }), children: [] }]);
    window.history.replaceState(null, '', '/sessions');

    expect(row('parent')).toBeTruthy();
    expect(document.querySelector('[data-session-id="other"]')).toBeNull();
    expect(document.querySelector('[data-session-id="grandchild"]')).toBeNull();
  });
});

describe('SessionList toolbar', () => {
  const child = makeSession({ id: 'child', parentId: 'parent', title: 'Helper' });

  const toolbarFamilies = [
    makeSession({ id: 'running', title: 'Running Gavel', project: 'gavel', processActive: true, lastActivityAt: '2026-10-11T13:00:00Z', costUsd: 1 }),
    makeSession({ id: 'failed', title: 'Failed Clicky', project: 'clicky', lifecycleStatus: 'failed', lastActivityAt: '2026-10-11T12:00:00Z', costUsd: 9 }),
    makeSession({ id: 'finished', title: 'Finished Gavel', project: 'gavel', lastActivityAt: '2026-10-11T11:00:00Z', costUsd: 3 }),
    makeSession({ id: 'other', title: 'Running Captain', project: 'captain', processActive: true, lastActivityAt: '2026-10-11T10:00:00Z', costUsd: 5 }),
  ].map(session => ({ session, children: [] }));
  const titles = () => within(screen.getByLabelText('Sessions')).getAllByRole('button', { name: /^View / })
    .map(button => button.getAttribute('aria-label'));

  function withUrl(url: string, body: () => Promise<void> | void) {
    return async () => {
      const oldUrl = window.location.href;
      window.history.replaceState(null, '', url);
      try {
        await body();
      } finally {
        window.history.replaceState(null, '', oldUrl);
      }
    };
  }

  it('restores included and excluded projects and states from the URL', withUrl('/sessions?sessions.project=gavel,clicky&sessions.state=-completed', () => {
    const list = renderList(toolbarFamilies);
    expect(titles()).toEqual(['View Running Gavel', 'View Failed Clicky']);
    list.unmount();

    renderList(toolbarFamilies);
    const params = new URLSearchParams(window.location.search);
    expect({ titles: titles(), project: params.get('sessions.project'), state: params.get('sessions.state') })
      .toEqual({ titles: ['View Running Gavel', 'View Failed Clicky'], project: 'gavel,clicky', state: '-completed' });
  }));

  it('offers Project and State as include/exclude filters and no Sort filter', withUrl('/sessions', () => {
    renderList(toolbarFamilies);

    expect({
      project: screen.getByRole('combobox', { name: 'Project' }).getAttribute('aria-label'),
      state: screen.getByRole('combobox', { name: 'State' }).getAttribute('aria-label'),
      sortFilter: screen.queryByRole('combobox', { name: 'Sort' }),
    }).toEqual({ project: 'Project', state: 'State', sortFilter: null });
  }));

  it('sorts from the sort icon menu and keeps the order in the URL', withUrl('/sessions', async () => {
    renderList(toolbarFamilies);
    expect(titles()).toEqual(['View Running Gavel', 'View Failed Clicky', 'View Finished Gavel', 'View Running Captain']);

    fireEvent.click(screen.getByRole('button', { name: 'Sort: Recent activity' }));
    fireEvent.click(await screen.findByRole('menuitem', { name: 'Highest cost' }));

    expect({ titles: titles(), sort: new URLSearchParams(window.location.search).get('sessions.sort') }).toEqual({
      titles: ['View Failed Clicky', 'View Running Captain', 'View Finished Gavel', 'View Running Gavel'],
      sort: 'cost-desc',
    });
    expect(screen.getByRole('button', { name: 'Sort: Highest cost' })).toBeTruthy();
  }));

  it('does not count the always-set sort as an active filter on a phone', async () => {
    const phoneWidth = 390;
    const media = vi.spyOn(window, 'matchMedia').mockImplementation((query: string) => {
      const maxWidth = /\(max-width:\s*(\d+)px\)/.exec(query);
      return {
        matches: maxWidth ? phoneWidth <= Number(maxWidth[1]) : false,
        media: query, onchange: null,
        addEventListener: () => undefined, removeEventListener: () => undefined,
        addListener: () => undefined, removeListener: () => undefined, dispatchEvent: () => false,
      } as MediaQueryList;
    });
    renderList([{ session: parent, children: [] }]);

    const trigger = await screen.findByRole('button', { name: /^(more )?filters$/i });
    expect(trigger.textContent).not.toMatch(/\d/);
    media.mockRestore();
  });

  it('has no expand or collapse controls, since sub-sessions live in the inspector', () => {
    renderList([{ session: parent, children: [{ session: child, children: [] }] }]);

    expect(screen.queryByRole('button', { name: 'List actions' })).toBeNull();
    expect(screen.queryByText(/Collapse all|Expand all/)).toBeNull();
  });
});
