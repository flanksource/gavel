import type { TodoItem } from '../../types';

export function TodoPRBadge({ pr, link = false }: { pr: NonNullable<TodoItem['pr']>; link?: boolean }) {
  const className = 'shrink-0 whitespace-nowrap rounded border border-border bg-muted px-1.5 py-0.5 text-xs text-primary';
  return link
    ? <a href={pr.url} target="_blank" rel="noreferrer" title={pr.url} className={className} onClick={event => event.stopPropagation()}>PR #{pr.number}</a>
    : <span title={pr.url} className={className}>PR #{pr.number}</span>;
}
