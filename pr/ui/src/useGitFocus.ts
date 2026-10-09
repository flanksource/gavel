import { useEffect, useState } from 'react';
import { HttpError, mutationJSON } from './query';
import { useDocumentVisible } from './useDocumentVisible';

// The server treats a focus as a 30s lease; renewing every third of that
// survives two lost renewals.
export const GIT_FOCUS_RENEW_MS = 10_000;
// A focused worktree and its repo's refs are rescanned at the tracker's hot
// cadence (gitstate.DefaultHot); a page re-reads them as often.
export const GIT_HOT_REFRESH_MS = 5_000;

interface GitFocusTarget {
  project: string;
  // Absolute path of the linked worktree being shown; omitted for the main checkout.
  worktree?: string;
  // False when another component already holds the lease for this target.
  enabled?: boolean;
}

/**
 * Tells the server which project worktree is on screen so its tracker scans it
 * eagerly. Posts on mount and renews the lease while the document is visible;
 * the lease is simply left to expire on unmount or when the tab is hidden.
 *
 * A failed post is logged and returned as `error` so the caller can show it.
 * A 503 means the tracker is unavailable, so renewing stops rather than
 * hammering it; any other failure retries on the next scheduled renewal.
 */
export function useGitFocus({ project, worktree, enabled = true }: GitFocusTarget): { error: string } {
  const visible = useDocumentVisible();
  const [error, setError] = useState('');

  useEffect(() => {
    if (!enabled || !visible) return;
    let timer: ReturnType<typeof setInterval> | undefined;
    let stopped = false;
    const renew = async () => {
      try {
        await mutationJSON({
          url: '/api/git/focus',
          method: 'POST',
          body: worktree ? { project, worktree } : { project },
          context: `Focus git tracker on ${project}${worktree ? ` (${worktree})` : ''}`,
        });
        if (!stopped) setError('');
      } catch (cause) {
        console.error('[git focus] renewal failed:', cause);
        if (stopped) return;
        setError(cause instanceof Error ? cause.message : 'Failed to focus the git tracker');
        if (cause instanceof HttpError && cause.status === 503) clearInterval(timer);
      }
    };
    void renew();
    timer = setInterval(() => void renew(), GIT_FOCUS_RENEW_MS);
    return () => {
      stopped = true;
      clearInterval(timer);
    };
  }, [enabled, visible, project, worktree]);

  return { error };
}
