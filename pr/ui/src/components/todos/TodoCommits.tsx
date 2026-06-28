import { useEffect, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import type { TodoCommit, TodoCommitsResponse } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { RelativeTime } from '../RelativeTime';
import { CommitFiles } from './TodoCommitFiles';
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

// CommitRow renders one linked commit with an expand toggle that reveals its
// per-file repomap status (each file revealing its own diff on hover). The short
// hash still links out to the commit on the origin remote.
function CommitRow({ dir, provider, commit }: { dir: string; provider: string; commit: TodoCommit }) {
  const [open, setOpen] = useState(false);
  return (
    <li>
      <div className="flex items-start gap-2 px-3 py-2.5 hover:bg-muted/30">
        <Button
          variant="ghost"
          size="icon"
          type="button"
          onClick={() => setOpen(o => !o)}
          aria-expanded={open}
          title={open ? 'Hide files' : 'Show files'}
          className="mt-0.5 inline-flex h-4 w-4 shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <GavelIcon name={open ? 'codicon:chevron-down' : 'codicon:chevron-right'} className="text-xs" />
        </Button>
        <span className="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border border-border bg-muted/30 text-muted-foreground">
          <GavelIcon name="codicon:git-commit" className="text-xs" />
        </span>
        <div className="min-w-0 flex-1">
          <Button
            variant="ghost"
            type="button"
            onClick={() => setOpen(o => !o)}
            className="block h-auto w-full truncate p-0 text-left text-sm text-foreground hover:underline"
            title={commit.subject}
          >
            {commit.subject}
          </Button>
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
      </div>
      {open && <CommitFiles dir={dir} provider={provider} hash={commit.hash} />}
    </li>
  );
}

// TodoCommits lists the commits that reference this todo through their
// Gavel-Issue-Id git trailer, each linking to the commit on the origin remote
// and expandable to show its diff. It renders nothing until at least one commit
// is found, so todos with no linked commits show no empty section.
export function TodoCommits({ dir, provider, todoRef }: { dir: string; provider: string; todoRef: string }) {
  const { commits, error } = useTodoCommits(dir, provider, todoRef);

  if (!error && commits.length === 0) return null;

  return (
    <section className="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <div className="flex items-center gap-2 border-b border-border bg-muted/30 px-3 py-2.5">
        <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground">
          <GavelIcon name="codicon:git-commit" className="text-xs" />
        </span>
        <span className="min-w-0 flex-1 truncate text-xs font-semibold uppercase tracking-wide text-muted-foreground">Commits</span>
        <span className="rounded-full border border-border bg-background px-1.5 py-0.5 text-[11px] tabular-nums text-muted-foreground">{commits.length}</span>
      </div>
      {error ? (
        <div className="px-3 py-2 text-xs text-red-600">{error}</div>
      ) : (
        <ul className="divide-y divide-border">
          {commits.map(commit => (
            <CommitRow key={commit.hash} dir={dir} provider={provider} commit={commit} />
          ))}
        </ul>
      )}
    </section>
  );
}
