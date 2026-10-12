import type { Tab } from './routes';

// The compact (mobile / native menubar) layout has four tabs. Three share a
// route tab with the desktop app; Processes has no route of its own, so it
// lives at /projects, the desktop tab that hosts the same workspaces.
export const MOBILE_TABS = ['processes', 'prs', 'todos', 'sessions'] as const;
export type MobileTab = (typeof MOBILE_TABS)[number];

const ROUTE_TAB: Record<MobileTab, Tab> = {
  processes: 'projects',
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
