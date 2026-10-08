import { useMemo } from 'react';
import { UiRobotAi } from '@flanksource/clicky-ui/icons';
import type { AgentSessionsResponse } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { SessionDetail, type SessionDetailProps } from './SessionDetail';
import { SessionList } from './SessionList';

export interface SessionsQueryView {
  data: AgentSessionsResponse | undefined;
  error: Error | null;
}

function ProjectErrors({ errors }: { errors: AgentSessionsResponse['errors'] }) {
  if (errors.length === 0) return null;
  return (
    <div role="alert" className="shrink-0 border-b border-amber-200 bg-amber-50 px-3 py-1.5 text-[11px] text-amber-800 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-300">
      {errors.map(entry => <div key={entry.project} className="truncate" title={entry.error}>{entry.project}: {entry.error}</div>)}
    </div>
  );
}

export function SessionsSidebar({ query, selectedId, scopeProject, onSelect }: {
  query: SessionsQueryView;
  selectedId: string;
  scopeProject: string;
  onSelect: (id: string) => void;
}) {
  const sessions = useMemo(
    () => (query.data?.sessions ?? []).filter(session => !scopeProject || session.project === scopeProject),
    [query.data, scopeProject],
  );
  const projects = useMemo(
    () => [...new Set(sessions.map(session => session.project).filter((name): name is string => !!name))].sort(),
    [sessions],
  );
  if (query.error && !query.data) {
    return <div role="alert" className="p-3 text-xs text-red-600 whitespace-pre-wrap">{query.error.message}</div>;
  }
  if (!query.data) {
    return <div role="status" className="flex items-center gap-2 p-3 text-xs text-muted-foreground"><Spinner /> Loading sessions…</div>;
  }
  return (
    <div className="flex h-full min-h-0 flex-col">
      <ProjectErrors errors={query.data.errors} />
      <div className="min-h-0 flex-1">
        <SessionList sessions={sessions} projects={projects} selectedId={selectedId} onSelect={onSelect} />
      </div>
    </div>
  );
}

export function SessionsDetailPane({ query, selectedId, ...actions }: {
  query: SessionsQueryView;
  selectedId: string;
} & Omit<SessionDetailProps, 'session'>) {
  const session = query.data?.sessions.find(candidate => candidate.id === selectedId);
  if (!session) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        <div className="text-center">
          <UiRobotAi className="mb-2 text-4xl" />
          <p>{selectedId && query.data ? 'That session is no longer in the recent list' : 'Select a session to view details'}</p>
        </div>
      </div>
    );
  }
  return <SessionDetail key={session.id} session={session} {...actions} />;
}
