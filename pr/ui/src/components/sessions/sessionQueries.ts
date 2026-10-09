import { useQuery } from '@tanstack/react-query';
import { fetchJSON, queryKeys } from '../../query';
import type { AgentSessionsResponse } from '../../types';

// Git state in the rows is read from the database and refreshed by the git
// stream (useGitStream invalidates this key); the poll only picks up session
// activity — new sessions, asks, approvals — which has no stream yet.
const SESSIONS_REFETCH_MS = 15_000;

export function useAgentSessions({ enabled }: { enabled: boolean }) {
  return useQuery<AgentSessionsResponse>({
    queryKey: queryKeys.agentSessions(),
    queryFn: async ({ signal }) => parseAgentSessions(await fetchJSON<unknown>({
      url: '/api/sessions',
      signal,
      context: 'Load agent sessions',
    })),
    enabled,
    retry: false,
    staleTime: SESSIONS_REFETCH_MS,
    refetchInterval: SESSIONS_REFETCH_MS,
  });
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

// Every row is read field by field by the list, so the shape is checked at the
// boundary: a malformed payload fails the query instead of rendering NaN.
export function parseAgentSessions(payload: unknown): AgentSessionsResponse {
  if (!isRecord(payload) || !Array.isArray(payload.sessions) || !Array.isArray(payload.errors)) {
    throw new Error('Load agent sessions: invalid response, expected {sessions, errors}');
  }
  for (const session of payload.sessions) {
    if (!isRecord(session)
      || typeof session.id !== 'string'
      || typeof session.title !== 'string'
      || typeof session.activityState !== 'string'
      || typeof session.processActive !== 'boolean'
      || typeof session.lastActivityAt !== 'string'
      || !isRecord(session.tokens)
      || typeof session.costUsd !== 'number') {
      throw new Error(`Load agent sessions: invalid session ${JSON.stringify(session)}`);
    }
  }
  return payload as unknown as AgentSessionsResponse;
}
