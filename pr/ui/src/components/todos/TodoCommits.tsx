import { useEffect, useState } from 'react';
import { AnsiHtml } from '@flanksource/clicky-ui/data';
import type { TodoCommit, TodoCommitDiffResponse, TodoCommitsResponse } from '../../types';
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

// CommitDiff lazily fetches and renders one commit's diff (the ANSI-colored
// `git show` output) once its row is expanded. The diff is shown in a black
// terminal panel via AnsiHtml, matching how process logs render.
function CommitDiff({ dir, provider, hash }: { dir: string; provider: string; hash: string }) {
  const [diff, setDiff] = useState('');
  const [truncated, setTruncated] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let cancelled = false;
    const controller = new AbortController();
    setLoading(true);
    setError('');
    const params = new URLSearchParams(todoQuery(dir, provider));
    params.set('hash', hash);
    fetch(`/api/todos/commits/diff?${params.toString()}`, { signal: controller.signal })
      .then(async res => {
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Failed to load diff');
        if (!cancelled) {
          const payload = data as TodoCommitDiffResponse;
          setDiff(payload.diff ?? '');
          setTruncated(!!payload.truncated);
        }
      })
      .catch((err: any) => {
        if (!cancelled && err?.name !== 'AbortError') setError(err?.message || 'Failed to load diff');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [dir, provider, hash]);

  if (loading) {
    return (
      <div className="flex items-center gap-2 px-3 py-2 text-[11px] text-muted-foreground">
        <GavelIcon name="svg-spinners:ring-resize" className="text-xs" />
        Loading diff…
      </div>
    );
  }
  if (error) {
    return <div className="px-3 py-2 text-[11px] text-red-600">{error}</div>;
  }
  if (!diff.trim()) {
    return <div className="px-3 py-2 text-[11px] text-muted-foreground">No changes</div>;
  }
  return (
    <div className="px-3 pb-3">
      <AnsiHtml
        as="pre"
        text={diff}
        className="max-h-[28rem] overflow-auto whitespace-pre rounded bg-black p-3 text-[11px] leading-snug text-gray-100"
      />
      {truncated && <div className="mt-1 text-[10px] text-muted-foreground">Diff truncated — open the commit to see the rest.</div>}
    </div>
  );
}

// CommitRow renders one linked commit with an expand toggle that reveals its
// diff. The short hash still links out to the commit on the origin remote.
function CommitRow({ dir, provider, commit }: { dir: string; provider: string; commit: TodoCommit }) {
  const [open, setOpen] = useState(false);
  return (
    <li>
      <div className="flex items-start gap-2 px-3 py-2">
        <button
          type="button"
          onClick={() => setOpen(o => !o)}
          aria-expanded={open}
          title={open ? 'Hide changes' : 'Show changes'}
          className="mt-0.5 inline-flex h-4 w-4 shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <GavelIcon name={open ? 'codicon:chevron-down' : 'codicon:chevron-right'} className="text-xs" />
        </button>
        <GavelIcon name="codicon:git-commit" className="mt-0.5 shrink-0 text-xs text-muted-foreground" />
        <div className="min-w-0 flex-1">
          <button
            type="button"
            onClick={() => setOpen(o => !o)}
            className="block w-full truncate text-left text-sm text-foreground hover:underline"
            title={commit.subject}
          >
            {commit.subject}
          </button>
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
      {open && <CommitDiff dir={dir} provider={provider} hash={commit.hash} />}
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
            <CommitRow key={commit.hash} dir={dir} provider={provider} commit={commit} />
          ))}
        </ul>
      )}
    </section>
  );
}
