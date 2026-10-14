import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@flanksource/clicky-ui/components';
import type { PRDetail, PRItem, TodoItem } from '../../types';
import { fetchJSON, queryKeys } from '../../query';
import { PRActions } from '../PRActions';
import { TodoPRBadge } from './TodoPRBadge';

export function TodoPRStatus({ pr, disabled }: { pr: NonNullable<TodoItem['pr']>; disabled?: boolean }) {
  const [expanded, setExpanded] = useState(false);
  const query = useQuery({
    queryKey: queryKeys.linkedPR(pr.url),
    queryFn: ({ signal }) => fetchJSON<PRDetail & { repo: string }>({
      url: `/api/prs/linked?url=${encodeURIComponent(pr.url)}`, signal, context: 'Load linked PR',
    }),
    enabled: expanded,
    refetchOnWindowFocus: false,
    retry: false,
  });
  const info = query.data?.pr;
  const error = query.error?.message || query.data?.error;
  const threads = new Set(info?.reviewThreads?.filter(thread => !thread.isResolved).map(thread => thread.threadId ?? thread.id)).size;
  const item: PRItem | undefined = info && query.data && {
    number: info.number, title: info.title, repo: query.data.repo,
    author: info.author.login, source: info.headRefName, target: info.baseRefName,
    state: info.state, isDraft: info.isDraft, mergeable: info.mergeable, url: pr.url, updatedAt: '',
  };
  return <div className="space-y-2">
    <div className="flex flex-wrap items-center gap-2">
      <TodoPRBadge pr={pr} link />
      <Button variant="ghost" size="sm" aria-expanded={expanded} onClick={() => setExpanded(!expanded)}>PR status and actions</Button>
    </div>
    {expanded && <section aria-label="Linked pull request" className="space-y-2 rounded-md border border-border bg-card p-3 text-xs">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="font-medium text-foreground">{info?.title ?? 'Pull request'}</span>
        <Button variant="outline" size="sm" loading={query.isFetching} onClick={() => void query.refetch()}>Refresh PR status</Button>
      </div>
      {error && <div role="alert" className="text-destructive">{error}</div>}
      {query.isFetching && !info && <p className="text-muted-foreground">Loading PR status…</p>}
      {info && <>
        <div className="flex flex-wrap gap-x-3 gap-y-1 text-muted-foreground">
          <span>{info.isDraft ? 'DRAFT' : info.state}</span>
          <span className="break-all">{info.headRefName} → {info.baseRefName}</span>
          {info.headRefOid && <span title={info.headRefOid}>HEAD {info.headRefOid.slice(0, 8)}</span>}
          {info.mergeable && <span>{info.mergeable}</span>}
          {info.reviewDecision && <span>{info.reviewDecision}</span>}
          <span>{threads} unresolved review thread{threads === 1 ? '' : 's'}</span>
        </div>
        {info.mergeReadiness?.reasons?.map(reason => <p key={reason} className="text-muted-foreground">{reason}</p>)}
        <div className="flex flex-wrap gap-2">
          {info.statusCheckRollup?.map((check, index) => <a key={index} href={check.detailsUrl || pr.url} target="_blank" rel="noreferrer" className="text-primary hover:underline">
            {check.name}: {check.conclusion || check.status}
          </a>)}
          {!info.statusCheckRollup?.length && <span className="text-muted-foreground">No checks reported</span>}
        </div>
        {query.data?.gavelResults?.map(result => <p key={result.stickyId ?? result.artifactId} className={result.error || result.testsFailed || result.lintViolations ? 'text-destructive' : 'text-muted-foreground'}>
          Gavel: {result.testsPassed}/{result.testsTotal} tests passed · {result.lintViolations} lint violations{result.error && ` · ${result.error}`}
        </p>)}
        {item && !error && <div className="flex flex-wrap gap-2">
          <PRActions pr={item} detail={query.data ?? null} disabled={disabled || query.isFetching} onChanged={() => void query.refetch()} />
        </div>}
      </>}
    </section>}
  </div>;
}
