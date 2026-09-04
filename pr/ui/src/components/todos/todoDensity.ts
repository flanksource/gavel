import type { ComponentType } from 'react';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import { UiListFlat, UiRows } from '@flanksource/clicky-ui/icons';
import type { TodoDensity } from '../../types';

// Row density is a per-user view preference for the todo lists, persisted
// alongside the status filter so it survives reloads. 'comfortable' keeps the
// two-line layout (status + title, then metadata); 'compact'
// collapses each todo onto a single line.
export const DENSITY_OPTIONS: { value: TodoDensity; label: string; icon: ComponentType<IconProps> }[] = [
  { value: 'comfortable', label: 'Comfortable', icon: UiRows },
  { value: 'compact', label: 'Compact', icon: UiListFlat },
];

const STORAGE_KEY = 'gavel.pr-ui.todoDensity.v1';

export function defaultDensity(): TodoDensity {
  return 'comfortable';
}

// Persistence is best-effort: localStorage can throw (private mode / disabled),
// so a failure falls back to the comfortable default rather than breaking.
export function loadDensity(): TodoDensity {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw === 'compact' || raw === 'comfortable' ? raw : defaultDensity();
  } catch {
    return defaultDensity();
  }
}

export function saveDensity(density: TodoDensity): void {
  try {
    localStorage.setItem(STORAGE_KEY, density);
  } catch {
    // best-effort: storage unavailable — skip persisting.
  }
}
