export type FilterMode = 'include' | 'exclude';
export type FilterState<T extends string = string> = Map<T, FilterMode>;

export function decodeFilterState<T extends string = string>(tokens: string[]): FilterState<T> {
  const state = new Map<T, FilterMode>();
  for (const token of tokens) {
    const value = token.startsWith('!') ? token.slice(1) : token;
    if (!value) continue;
    state.set(value as T, token.startsWith('!') ? 'exclude' : 'include');
  }
  return state;
}

export function encodeFilterState<T extends string = string>(state: FilterState<T>): string[] {
  return Array.from(state.entries())
    .sort(([a], [b]) => String(a).localeCompare(String(b)))
    .map(([value, mode]) => mode === 'exclude' ? `!${value}` : String(value));
}

export function cycleFilterState<T extends string = string>(state: FilterState<T>, key: T): FilterState<T> {
  const next = new Map(state);
  const current = next.get(key);
  if (current === 'include') {
    next.set(key, 'exclude');
  } else if (current === 'exclude') {
    next.delete(key);
  } else {
    next.set(key, 'include');
  }
  return next;
}

export function matchesFilterState<T extends string = string>(value: T | null | undefined, state: FilterState<T>): boolean {
  if (state.size === 0) return true;

  const include = new Set<T>();
  const exclude = new Set<T>();
  for (const [key, mode] of state.entries()) {
    if (mode === 'include') include.add(key);
    if (mode === 'exclude') exclude.add(key);
  }

  if (value && exclude.has(value)) return false;
  if (include.size === 0) return true;
  return !!value && include.has(value);
}
