import type { AgentSession } from '../../types';
import { passFacet, type FacetModes } from '../../utils';
import { dirtyFileCount } from '../projectGitView';

export type SessionState = 'working' | 'approval' | 'ask' | 'completed' | 'failed';

export const SESSION_STATE_LABELS: Record<SessionState, string> = {
  working: 'Running',
  approval: 'Approval',
  ask: 'Needs input',
  completed: 'Finished',
  failed: 'Failed',
};

export const SESSION_SORTS = [
  { value: 'activity-desc', label: 'Recent activity' },
  { value: 'activity-asc', label: 'Oldest activity' },
  { value: 'started-desc', label: 'Newest started' },
  { value: 'started-asc', label: 'Oldest started' },
  { value: 'title-asc', label: 'Title A–Z' },
  { value: 'cost-desc', label: 'Highest cost' },
] as const;

export type SessionSort = (typeof SESSION_SORTS)[number]['value'];

export function waitingOnInput(session: AgentSession): boolean {
  return session.activityState === 'ask' || session.activityState === 'approval';
}

export function sessionFailed(session: AgentSession): boolean {
  return ['failed', 'interrupted'].includes(session.lifecycleStatus);
}

/** Input the session waits on comes first; a live process is running; else how it ended. */
export function sessionState(session: AgentSession): SessionState {
  if (session.activityState === 'approval') return 'approval';
  if (session.activityState === 'ask') return 'ask';
  if (session.processActive) return 'working';
  return sessionFailed(session) ? 'failed' : 'completed';
}

export function sessionDirtyCount(session: AgentSession): number {
  return session.git ? dirtyFileCount(session.git.changes) : 0;
}

/** The most pressing work-state fact, for the header badge that opens the panel on narrow screens. */
export function workStateBadge(session: AgentSession): { label: string; tone?: 'warning' } {
  const { git, pr } = session;
  if (!git) return { label: 'no git' };
  if (git.changes.conflict) return { label: `${git.changes.conflict} conflicted`, tone: 'warning' };
  const dirty = sessionDirtyCount(session);
  if (dirty > 0) return { label: `${dirty} uncommitted`, tone: 'warning' };
  if (git.ahead > 0) return { label: `${git.ahead} ahead` };
  return { label: pr ? `PR #${pr.number}` : 'clean' };
}

/**
 * Why a session needs a person: input it waits on, or work it left in a bad
 * state. `work: false` leaves out the work-state reasons (conflicts, failing PR
 * checks, abandoned changes) for a view that shows the work state itself.
 */
export function attentionReasons(session: AgentSession, { work = true }: { work?: boolean } = {}): string[] {
  const reasons: string[] = [];
  if (session.activityState === 'approval') reasons.push('Needs approval');
  if (session.activityState === 'ask') reasons.push('Asked a question');
  if (sessionFailed(session)) reasons.push('Session failed');
  if (!work) return reasons;
  if (session.git?.changes.conflict) reasons.push(`${session.git.changes.conflict} conflicts`);
  if (session.pr?.checkStatus === 'failure' && session.pr.state !== 'merged') reasons.push(`PR #${session.pr.number} checks failing`);
  if (!session.processActive && sessionDirtyCount(session) > 0) reasons.push('Uncommitted changes left behind');
  return reasons;
}

/** Share of the context window used, 0–100; undefined when unknown. */
export function contextUsedPercent(session: AgentSession): number | undefined {
  return session.context ? 100 - session.context.freePercent : undefined;
}

/** A session with the subagents it delegated to, nested. */
export interface SessionFamily {
  session: AgentSession;
  children: SessionFamily[];
}

/** Nests each subagent under its parent. The server lists every parent it
 *  names, so a missing parent is a broken payload. */
export function sessionFamilies(sessions: AgentSession[]): SessionFamily[] {
  const nodes = new Map(sessions.map(session => [session.id, { session, children: [] as SessionFamily[] }]));
  const roots: SessionFamily[] = [];
  for (const node of nodes.values()) {
    const { parentId } = node.session;
    if (!parentId) {
      roots.push(node);
      continue;
    }
    const parent = nodes.get(parentId);
    if (!parent) throw new Error(`Session ${node.session.id} names parent ${parentId}, which is not listed`);
    parent.children.push(node);
  }
  return roots;
}

export function flattenFamilies(families: SessionFamily[]): AgentSession[] {
  return families.flatMap(family => [family.session, ...flattenFamilies(family.children)]);
}

/** The most recent activity anywhere in the family, as epoch ms. */
export function latestActivity(family: SessionFamily): number {
  return Math.max(Date.parse(family.session.lastActivityAt), ...family.children.map(latestActivity));
}

function startedAt(session: AgentSession): number {
  return Date.parse(session.startedAt ?? session.lastActivityAt);
}

export interface SessionFilters {
  search: string;
  /** Projects to include or exclude; absent projects are neutral. */
  project: FacetModes;
  /** States to include or exclude; absent states are neutral. */
  state: FacetModes;
  sort: SessionSort;
}

export const DEFAULT_SESSION_FILTERS: SessionFilters = { search: '', project: {}, state: {}, sort: 'activity-desc' };

function compareFamilies(sort: SessionSort): (a: SessionFamily, b: SessionFamily) => number {
  switch (sort) {
    case 'activity-desc': return (a, b) => latestActivity(b) - latestActivity(a);
    case 'activity-asc': return (a, b) => latestActivity(a) - latestActivity(b);
    case 'started-desc': return (a, b) => startedAt(b.session) - startedAt(a.session);
    case 'started-asc': return (a, b) => startedAt(a.session) - startedAt(b.session);
    case 'title-asc': return (a, b) => a.session.title.localeCompare(b.session.title);
    case 'cost-desc': return (a, b) => b.session.costUsd - a.session.costUsd;
  }
}

/** Keeps the families with a matching session — a match keeps its ancestors
 *  so it is never shown without its parent — sorted at every level. */
export function visibleFamilies(families: SessionFamily[], filters: SessionFilters): SessionFamily[] {
  const query = filters.search.trim().toLowerCase();
  const matches = ({ session }: SessionFamily) =>
    passFacet(filters.project, session.project ? [session.project] : [])
    && passFacet(filters.state, [sessionState(session)])
    && (!query || [session.title, session.project, session.worktree?.branch, session.todo?.title, session.model]
      .some(value => value?.toLowerCase().includes(query)));
  return families
    .flatMap(family => {
      const children = visibleFamilies(family.children, filters);
      return matches(family) || children.length > 0 ? [{ ...family, children }] : [];
    })
    .sort(compareFamilies(filters.sort));
}
