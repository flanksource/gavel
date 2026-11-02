import type { Tab } from './routes';

export const MOBILE_TABS = ['projects', 'processes', 'prs', 'todos', 'sessions'] as const;
export type MobileTab = (typeof MOBILE_TABS)[number];

const ROUTE_TAB: Record<MobileTab, Tab> = {
  projects: 'projects',
  processes: 'processes',
  prs: 'prs',
  todos: 'todos',
  sessions: 'sessions',
};

export function routeTabFromMobile(tab: MobileTab): Tab {
  return ROUTE_TAB[tab];
}

/** Route tabs the compact layout has no view for (activity, tasks, prompts) show the PR list, its landing tab. */
export function mobileTabFromRoute(tab: Tab): MobileTab {
  return MOBILE_TABS.find(candidate => ROUTE_TAB[candidate] === tab) ?? 'prs';
}
