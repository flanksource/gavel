import { useEffect, useState } from 'react';
import type { TodoCommit, TodoCommitsResponse } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { RelativeTime } from '../RelativeTime';
import { todoQuery } from './format';

// useTodoCommits fetches the git commits linked to a todo via its Gavel-Issue-Id
// trailer. It refetches when the todo ref changes and reports nothing for a todo
// with no linked commits (e.g. file-backed todos that carry no id).
function useTodoCommits(dir: string, provider: string, todoRef: string) {
  const [commits, setCommits] = useState<TodoCommit[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!todoRef) {
      setCommits([]);
      setError('');
      return;
    }
    let cancelled = false;
    const controller = new AbortController();
    setLoading(true);
    setError('');
    const params = new URLSearchParams(todoQuery(dir, provider));
    params.set('ref', todoRef);
    fetch(`/api/todos/commits?${params.toString()}`, { signal: controller.signal })
      .then(async res => {
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Failed to load commits');
        if (!cancelled) setCommits((data as TodoCommitsResponse).commits ?? []);
      })
      .catch((err: any) => {
        if (!cancelled && err?.name !== 'AbortError') setError(err?.message || 'Failed to load commits');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [dir, provider, todoRef]);

  return { commits, loading, error };
}

// TodoCommits lists the commits that reference this todo through their
// Gavel-Issue-Id git trailer, each linking to the commit on the origin remote.
// It renders nothing until at least one commit is found, so todos with no linked
// commits (or file-backed todos without an id) show no empty section.
export function TodoCommits({ dir, provider, todoRef }: { dir: string; provider: string; todoRef: string }) {
  const { commits, error } = useTodoCommits(dir, provider, todoRef);

  if (!error && commits.length === 0) return null;

  return (
    <section className="rounded-md border border-border bg-background">
      <div className="flex items-center gap-2 px-3 py-2">
        <GavelIcon name="codicon:git-commit" className="shrink-0 text-xs text-muted-foreground" />
        <span className="min-w-0 flex-1 truncate text-xs font-semibold uppercase text-muted-foreground">Commits</span>
        <span className="text-xs tabular-nums text-muted-foreground">{commits.length}</span>
      </div>
      {error ? (
        <div className="border-t border-border px-3 py-2 text-xs text-red-600">{error}</div>
      ) : (
        <ul className="divide-y divide-border border-t border-border">
          {commits.map(commit => (
            <li key={commit.hash} className="flex items-start gap-2 px-3 py-2">
              <GavelIcon name="codicon:git-commit" className="mt-0.5 shrink-0 text-xs text-muted-foreground" />
              <div className="min-w-0 flex-1">
                <div className="truncate text-sm text-foreground" title={commit.subject}>{commit.subject}</div>
                <div className="mt-0.5 flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
                  {commit.url ? (
                    <a
                      href={commit.url}
                      target="_blank"
                      rel="noreferrer"
                      className="font-mono text-primary hover:underline"
                      title="Open commit"
                    >
                      {commit.shortHash}
                    </a>
                  ) : (
                    <span className="font-mono">{commit.shortHash}</span>
                  )}
                  {commit.author && <span className="truncate">{commit.author}</span>}
                  {commit.date && <RelativeTime iso={commit.date} />}
                </div>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
