import { passFacet, type FacetModes } from '../utils';
import type { ProjectFileStatus } from './ProjectStatusView';

// UNSCOPED_SCOPE buckets files repomap assigned no scope, so an include-only
// scope filter can still bring them back.
export const UNSCOPED_SCOPE = '(none)';

export interface ProjectFileFilters {
  search: string;
  scopes: FacetModes;
}

export const emptyProjectFileFilters = (): ProjectFileFilters => ({ search: '', scopes: {} });

export function projectFileFiltersActive(filters: ProjectFileFilters): boolean {
  return filters.search.trim() !== '' || Object.keys(filters.scopes).length > 0;
}

export function fileScopes(file: Pick<ProjectFileStatus, 'scopes'>): string[] {
  return file.scopes?.length ? file.scopes : [UNSCOPED_SCOPE];
}

export function filterProjectFiles<T extends Pick<ProjectFileStatus, 'path' | 'previousPath' | 'scopes'>>(files: T[], filters: ProjectFileFilters): T[] {
  const search = filters.search.trim().toLowerCase();
  return files.filter(file => (
    (!search || file.path.toLowerCase().includes(search) || (file.previousPath?.toLowerCase().includes(search) ?? false))
    && passFacet(filters.scopes, fileScopes(file))
  ));
}

export function scopeCounts(files: Pick<ProjectFileStatus, 'scopes'>[]): Record<string, number> {
  const counts: Record<string, number> = {};
  for (const file of files) {
    for (const scope of fileScopes(file)) counts[scope] = (counts[scope] ?? 0) + 1;
  }
  return counts;
}
