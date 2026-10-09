import type { AgentSession } from '../../types';
import { dirtyFileCount } from '../projectGitView';

export type SessionGroupKey = 'active' | 'recent' | 'older';

export const SESSION_GROUPS: readonly { key: SessionGroupKey; label: string; defaultOpen: boolean }[] = [
  { key: 'active', label: 'Active', defaultOpen: true },
  { key: 'recent', label: 'Recent · 24h', defaultOpen: true },
  { key: 'older', label: 'Older', defaultOpen: false },
];

const RECENT_WINDOW_MS = 24 * 60 * 60 * 1000;

export function waitingOnInput(session: AgentSession): boolean {
  return session.activityState === 'ask' || session.activityState === 'approval';
}

/** Active = live process or waiting on you; else recent (24h) or older by last activity. */
export function sessionGroup(session: AgentSession, now: number): SessionGroupKey {
  if (session.processActive || waitingOnInput(session)) return 'active';
  return now - Date.parse(session.lastActivityAt) <= RECENT_WINDOW_MS ? 'recent' : 'older';
}

/** The server orders the list; grouping keeps that order within each group. */
export function groupSessions(sessions: AgentSession[], now: number): Record<SessionGroupKey, AgentSession[]> {
  const groups: Record<SessionGroupKey, AgentSession[]> = { active: [], recent: [], older: [] };
  for (const session of sessions) groups[sessionGroup(session, now)].push(session);
  return groups;
}

export function sessionDirtyCount(session: AgentSession): number {
  return session.git ? dirtyFileCount(session.git.changes) : 0;
}

export function sessionFailed(session: AgentSession): boolean {
  return ['failed', 'interrupted'].includes(session.lifecycleStatus);
}

/** Why a session needs a person: input it waits on, or work it left in a bad state. */
export function attentionReasons(session: AgentSession): string[] {
  const reasons: string[] = [];
  if (session.activityState === 'approval') reasons.push('Needs approval');
  if (session.activityState === 'ask') reasons.push('Asked a question');
  if (sessionFailed(session)) reasons.push('Session failed');
  if (session.git?.changes.conflict) reasons.push(`${session.git.changes.conflict} conflicts`);
  if (session.pr?.checkStatus === 'failure' && session.pr.state !== 'merged') reasons.push(`PR #${session.pr.number} checks failing`);
  if (!session.processActive && sessionDirtyCount(session) > 0) reasons.push('Uncommitted changes left behind');
  return reasons;
}

/** Share of the context window used, 0–100; undefined when unknown. */
export function contextUsedPercent(session: AgentSession): number | undefined {
  return session.context ? 100 - session.context.freePercent : undefined;
}

export function contextTone(percent: number): 'danger' | 'warning' | 'ok' {
  if (percent >= 90) return 'danger';
  if (percent >= 70) return 'warning';
  return 'ok';
}

export interface SessionFilters {
  search: string;
  project: string;
  attention: boolean;
}

export function filterSessions(sessions: AgentSession[], { search, project, attention }: SessionFilters): AgentSession[] {
  const query = search.trim().toLowerCase();
  return sessions.filter(session =>
    (!project || session.project === project)
    && (!attention || attentionReasons(session).length > 0)
    && (!query || [session.title, session.project, session.worktree?.branch, session.todo?.title, session.model]
      .some(value => value?.toLowerCase().includes(query))));
}
