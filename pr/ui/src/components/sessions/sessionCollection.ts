import type { SessionCollectionInput, SessionCollectionItem } from '@flanksource/clicky-ui/ai';
import type { AgentSession } from '../../types';
import { captainSessionUrl } from '../todos/TodoSessionDetail';
import { sessionState } from './sessionView';

/**
 * The inspector's view of one session: the session itself, current, with every
 * sub-session it delegated to nested beneath it. The sessions list shows only
 * top-level sessions, so this is where sub-sessions are reached — the
 * inspector's ⋮ menu lists them under "Session content", unselected until
 * picked. Each item loads from its own Captain session.
 */
export function sessionCollection(session: AgentSession, sessions: AgentSession[]): SessionCollectionInput {
  const byParent = new Map<string, AgentSession[]>();
  for (const candidate of sessions) {
    if (!candidate.parentId) continue;
    byParent.set(candidate.parentId, [...(byParent.get(candidate.parentId) ?? []), candidate]);
  }
  const items: SessionCollectionItem[] = [];
  const visit = (node: AgentSession, parentId?: string) => {
    items.push(collectionItem(node, parentId));
    for (const child of byParent.get(node.id) ?? []) visit(child, node.id);
  };
  visit(session);
  return { kind: 'session-collection', id: `session:${session.id}`, currentSessionId: session.id, sessions: items };
}

function collectionItem(session: AgentSession, parentId: string | undefined): SessionCollectionItem {
  const status = sessionState(session);
  return {
    id: session.id,
    ...(parentId ? { parentId } : {}),
    label: session.title,
    status,
    summary: {
      provider: session.provider || session.source,
      ...(session.model ? { model: session.model } : {}),
      ...(session.effort ? { effort: session.effort } : {}),
      status,
      ...(session.pid ? { pid: session.pid } : {}),
      ...(session.durationMs !== undefined ? { durationMs: session.durationMs } : {}),
      updatedAt: session.lastActivityAt,
      cost: session.costUsd,
    },
    src: captainSessionUrl(session.id),
  };
}
