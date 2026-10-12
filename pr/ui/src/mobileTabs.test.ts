import { describe, expect, it } from 'vitest';
import { buildRoute, emptyRouteState, parseRoute, type Tab } from './routes';
import { MOBILE_TABS, mobileTabFromRoute, routeTabFromMobile } from './mobileTabs';

describe('mobile tab <-> route tab', () => {
  it.each([
    ['prs', 'prs'],
    ['todos', 'todos'],
    ['sessions', 'sessions'],
    ['projects', 'processes'],
  ] as const)('route tab %s is the %s mobile tab', (route, mobile) => {
    expect(mobileTabFromRoute(route)).toBe(mobile);
    expect(routeTabFromMobile(mobile)).toBe(route);
  });

  it.each<Tab>(['activity', 'tasks', 'prompts'])('route tab %s has no mobile view and shows the PR list', route => {
    expect(mobileTabFromRoute(route)).toBe('prs');
  });

  it('round-trips every mobile tab', () => {
    for (const tab of MOBILE_TABS) expect(mobileTabFromRoute(routeTabFromMobile(tab))).toBe(tab);
  });

  it('persists the tab and the open session in the URL path', () => {
    const url = buildRoute({ ...emptyRouteState(), tab: routeTabFromMobile('sessions'), selectedPath: 'abc-123', scopeProject: 'gavel' });
    expect(url).toBe('/sessions/abc-123?project=gavel');

    const parsed = parseRoute({ pathname: '/sessions/abc-123', search: '?project=gavel' } as Location);
    expect(mobileTabFromRoute(parsed.tab)).toBe('sessions');
    expect(parsed.selectedPath).toBe('abc-123');
  });
});
