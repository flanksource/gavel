import { useEffect, useRef, useState } from 'react';
import { useQueryClient, type QueryClient } from '@tanstack/react-query';
import { openEventStream } from './eventHub';
import { queryKeys } from './query';

// Project name -> the generation of its git rows, as pushed by /api/git/stream.
type Generations = Readonly<Record<string, number>>;

function parseGenerations(payload: unknown): Generations {
  if (typeof payload !== 'object' || payload === null || Array.isArray(payload)) {
    throw new Error(`expected an object of project generations, got ${JSON.stringify(payload)}`);
  }
  for (const [project, generation] of Object.entries(payload)) {
    if (typeof generation !== 'number') {
      throw new Error(`generation of ${project} is not a number: ${JSON.stringify(generation)}`);
    }
  }
  return payload as Generations;
}

// Every query that renders a project's git state is invalidated rather than
// patched: the frame only says that something changed, not what. Session rows
// carry their worktree's git state, so the sessions list refetches too.
function invalidateChangedProjects({ queryClient, seen, next }: { queryClient: QueryClient; seen: Generations; next: Generations }) {
  const changed = Object.keys(next).filter(project => seen[project] !== next[project]);
  for (const project of changed) {
    void queryClient.invalidateQueries({ queryKey: queryKeys.projectGit(project) });
    void queryClient.invalidateQueries({ queryKey: queryKeys.projectStatusScope(project) });
  }
  if (changed.length > 0) {
    void queryClient.invalidateQueries({ queryKey: queryKeys.projectGitSummary() });
    void queryClient.invalidateQueries({ queryKey: queryKeys.agentSessions() });
  }
}

/**
 * Subscribes to /api/git/stream, whose frames map each tracked project to the
 * generation of its git rows, and invalidates that project's git queries when
 * its generation moves. The first frame only records the baseline. The seen
 * generations survive a disable/re-enable (hidden tab) and a reconnect, so a
 * change that happened in the gap is caught by the next frame.
 *
 * Returns a non-empty message while the stream is disconnected or sent a frame
 * that could not be read.
 */
export function useGitStream({ enabled }: { enabled: boolean }): { error: string } {
  const queryClient = useQueryClient();
  const seen = useRef<Generations | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!enabled) return;
    const stream = openEventStream('/api/git/stream');
    stream.addEventListener('message', event => {
      try {
        const next = parseGenerations(JSON.parse((event as MessageEvent<string>).data));
        if (seen.current) invalidateChangedProjects({ queryClient, seen: seen.current, next });
        seen.current = next;
        setError('');
      } catch (cause) {
        console.error('[git stream] invalid update:', cause);
        setError('Git change stream received an invalid update.');
      }
    });
    stream.onerror = () => setError(current => current || 'Git change stream disconnected; reconnecting…');
    return () => stream.close();
  }, [enabled, queryClient]);

  return { error };
}
