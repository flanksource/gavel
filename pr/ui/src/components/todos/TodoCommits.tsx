import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@flanksource/clicky-ui/components';
import { UiChevronDown, UiChevronRight, UiGitGraph } from '@flanksource/clicky-ui/icons';
import type { TodoCommit, TodoCommitsResponse } from '../../types';
import { fetchJSON } from '../../query';
import { RelativeTime } from '../RelativeTime';
import { CommitFiles } from './TodoCommitFiles';
import { todoQuery } from './format';
import { todoQueryKeys } from './todoQueries';

// useTodoCommits fetches the git commits linked to a todo via its Gavel-Issue-Id
// trailer. It refetches when the todo ref changes and reports nothing for a todo
// with no linked commits.
function useTodoCommits(dir: string, todoRef: string) {
  const query = useQuery({
    queryKey: todoQueryKeys.commits(dir, todoRef),
    queryFn: async ({ signal }) => {
      const params = new URLSearchParams(todoQuery(dir));
      params.set('ref', todoRef);
      const data = await fetchJSON<TodoCommitsResponse>({
        url: `/api/todos/commits?${params.toString()}`,
        signal,
        context: 'Failed to load Todo commits',
      });
      return data.commits ?? [];
    },
    enabled: !!todoRef,
    staleTime: 5_000,
  });

  return { commits: query.data ?? [], loading: query.isFetching, error: query.error?.message ?? '' };
}

// CommitRow renders one linked commit with an expand toggle that reveals its
// per-file repomap status (each file revealing its own diff on hover). The short
// hash still links out to the commit on the origin remote.
function CommitRow({ dir, commit }: { dir: string; commit: TodoCommit }) {
  const [open, setOpen] = useState(false);
  const ChevronIcon = open ? UiChevronDown : UiChevronRight;
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
          <ChevronIcon className="text-xs" />
        </Button>
        <span className="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border border-border bg-muted/30 text-muted-foreground">
          <UiGitGraph className="text-xs" />
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
      {open && <CommitFiles dir={dir} hash={commit.hash} />}
    </li>
  );
}

// TodoCommits lists the commits that reference this todo through their
// Gavel-Issue-Id git trailer, each linking to the commit on the origin remote
// and expandable to show its diff. It renders nothing until at least one commit
// is found, so todos with no linked commits show no empty section.
export function TodoCommits({ dir, todoRef }: { dir: string; todoRef: string }) {
  const { commits, error } = useTodoCommits(dir, todoRef);

  if (!error && commits.length === 0) return null;

  return (
    <section className="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <div className="flex items-center gap-2 border-b border-border bg-muted/30 px-3 py-2.5">
        <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground">
          <UiGitGraph className="text-xs" />
        </span>
        <span className="min-w-0 flex-1 truncate text-xs font-semibold uppercase tracking-wide text-muted-foreground">Commits</span>
        <span className="rounded-full border border-border bg-background px-1.5 py-0.5 text-[11px] tabular-nums text-muted-foreground">{commits.length}</span>
      </div>
      {error ? (
        <div className="px-3 py-2 text-xs text-red-600">{error}</div>
      ) : (
        <ul className="divide-y divide-border">
          {commits.map(commit => (
            <CommitRow key={commit.hash} dir={dir} commit={commit} />
          ))}
        </ul>
      )}
    </section>
  );
}
