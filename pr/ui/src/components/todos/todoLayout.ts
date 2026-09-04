import type { ComponentType } from 'react';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import { UiSidebar, UiTable } from '@flanksource/clicky-ui/icons';
import type { TodoLayout } from '../../types';

// Layout is a per-user view preference for the dashboard's Todos tab, persisted
// alongside density and grouping so it survives reloads. 'split' is the
// master-detail default: the list lives in the AppShell body sidebar beside the
// detail pane. 'full' drops the sidebar entirely — the todos render as a
// full-viewport table under a full-width filter bar, and selecting one opens it
// as a full-page detail behind a back arrow.
//
// The menubar/mobile layout ignores this: it is already single-column and
// full-width, so a second axis of choice there would mean nothing.
export const LAYOUT_OPTIONS: { value: TodoLayout; label: string; icon: ComponentType<IconProps> }[] = [
  { value: 'split', label: 'Split', icon: UiSidebar },
  { value: 'full', label: 'Full width', icon: UiTable },
];

const STORAGE_KEY = 'gavel.pr-ui.todoLayout.v1';

export function defaultLayout(): TodoLayout {
  return 'split';
}

// Persistence is best-effort: localStorage can throw (private mode / disabled),
// so a failure falls back to the split default rather than breaking.
export function loadLayout(): TodoLayout {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw === 'split' || raw === 'full' ? raw : defaultLayout();
  } catch {
    return defaultLayout();
  }
}

export function saveLayout(layout: TodoLayout): void {
  try {
    localStorage.setItem(STORAGE_KEY, layout);
  } catch {
    // best-effort: storage unavailable — skip persisting.
  }
}
