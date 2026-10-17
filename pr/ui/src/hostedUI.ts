// Decides which parts of the dashboard a hosted (clerk mode) organization sees. The
// server serves only org-scoped routes there, so host-only views and controls are hidden.
import type { AuthMode } from './query';
import type { Tab } from './routes';

// hostedTabs are the tabs whose routes the hosted allowlist serves.
const hostedTabs: readonly Tab[] = ['prs', 'todos'];

export function visibleTabs<T extends Tab>(mode: AuthMode, tabs: readonly T[]): T[] {
  return mode === 'clerk' ? tabs.filter(tab => hostedTabs.includes(tab)) : [...tabs];
}

// tabFor maps a requested tab to one the mode can show, so a hosted link to a
// host-only tab opens the PR tab instead of a page of errors.
export function tabFor(mode: AuthMode, tab: Tab): Tab {
  return mode === 'clerk' && !hostedTabs.includes(tab) ? 'prs' : tab;
}

// githubScopeVisible reports whether the GitHub org/project scope chooser is shown. A
// hosted org's repositories come from its GitHub App installations instead.
export function githubScopeVisible(mode: AuthMode): boolean {
  return mode !== 'clerk';
}

// hostControlsVisible reports whether controls that act on the host itself (the
// assistant, task and process managers, ~/.gavel.yaml settings) are shown.
export function hostControlsVisible(mode: AuthMode): boolean {
  return mode !== 'clerk';
}

// todoVisibilityVisible reports whether a todo's visibility (private to its creator, or
// public to the org) is shown and chosen. A local dashboard shows every todo.
export function todoVisibilityVisible(mode: AuthMode): boolean {
  return mode === 'clerk';
}

// runtimeChoiceVisible reports whether a run's runtime, model and presets are chosen in the
// dashboard. A hosted org's runs use the repository's runtime for each step.
export function runtimeChoiceVisible(mode: AuthMode): boolean {
  return mode !== 'clerk';
}
