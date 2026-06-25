import type { TodoSessionEvent } from '../../types';

// The session browser groups every streamed event into one of a few broad
// categories the user can toggle, plus a per-tool dimension for tool_use events.
// Categories answer "what kind of activity", tools answer "which tool".
export type SessionCategory = 'assistant' | 'thinking' | 'tool' | 'lifecycle' | 'error';

export interface SessionCategoryDef {
  key: SessionCategory;
  label: string;
}

// Listed in display order; only categories present in the stream are shown.
export const SESSION_CATEGORIES: SessionCategoryDef[] = [
  { key: 'assistant', label: 'Assistant' },
  { key: 'thinking', label: 'Thinking' },
  { key: 'tool', label: 'Tools' },
  { key: 'lifecycle', label: 'Turn ends' },
  { key: 'error', label: 'Errors' },
];

// Explore subagent calls dominate the transcript with read-only fan-out and are
// hidden by default; the user opts them back in via the filter menu.
export const DEFAULT_HIDDEN_TOOLS = ['Explore'];

export function eventCategory(ev: TodoSessionEvent): SessionCategory {
  switch (ev.kind) {
    case 'thinking':
      return 'thinking';
    case 'tool_use':
      return 'tool';
    case 'turn_end':
      return 'lifecycle';
    case 'error':
      return 'error';
    default:
      return 'assistant';
  }
}

// toolKey is the filterable identity of a tool_use event: the subagent type for
// Task/Agent dispatches (so "Explore" is filterable apart from generic "Task"),
// otherwise the tool name itself. Empty for non-tool events.
export function toolKey(ev: TodoSessionEvent): string {
  if (ev.kind !== 'tool_use') return '';
  return ev.subagent || ev.tool || '';
}

export function eventVisible(
  ev: TodoSessionEvent,
  hiddenCategories: Set<SessionCategory>,
  hiddenTools: Set<string>,
): boolean {
  if (hiddenCategories.has(eventCategory(ev))) return false;
  const key = toolKey(ev);
  if (key && hiddenTools.has(key)) return false;
  return true;
}

export interface SessionFacets {
  categories: { def: SessionCategoryDef; count: number }[];
  tools: { key: string; count: number }[];
}

// computeFacets tallies which categories and tools actually appear in the
// stream so the menu only offers toggles that affect something, with counts.
export function computeFacets(events: TodoSessionEvent[]): SessionFacets {
  const catCounts = new Map<SessionCategory, number>();
  const toolCounts = new Map<string, number>();
  for (const ev of events) {
    const cat = eventCategory(ev);
    catCounts.set(cat, (catCounts.get(cat) ?? 0) + 1);
    const key = toolKey(ev);
    if (key) toolCounts.set(key, (toolCounts.get(key) ?? 0) + 1);
  }
  return {
    categories: SESSION_CATEGORIES.filter(d => catCounts.has(d.key)).map(def => ({
      def,
      count: catCounts.get(def.key) ?? 0,
    })),
    tools: [...toolCounts.entries()]
      .sort((a, b) => a[0].localeCompare(b[0]))
      .map(([key, count]) => ({ key, count })),
  };
}
