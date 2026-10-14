import type { ComponentType, ReactNode } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@flanksource/clicky-ui/components';
import { UiClose, UiCopy, UiGitPr, UiLink, type IconProps } from '@flanksource/clicky-ui/icons';
import type { TodoLink, TodoPRLinkItem } from '../../types';
import { StatusIcon } from './format';
import { TodoRouteLink } from './TodoChildren';
import { todoLinksQueryOptions, todoPRLinksQueryOptions, useUnlinkTodo, useUnlinkTodoPR } from './todoLinkQueries';
import { TodoPRBadge } from './TodoPRBadge';

const chipClass = 'shrink-0 rounded border border-border bg-muted px-1.5 py-0.5 text-[10px] uppercase tracking-wide text-muted-foreground';

function LinkGroup({ title, icon: Icon, children }: { title: string; icon: ComponentType<IconProps>; children: ReactNode }) {
  return (
    <section className="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <header className="flex items-center gap-2 border-b border-border bg-muted/30 px-3 py-2.5">
        <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground">
          <Icon className="text-xs" />
        </span>
        <h2 className="min-w-0 flex-1 truncate text-xs font-semibold uppercase text-muted-foreground">{title}</h2>
      </header>
      <ul className="divide-y divide-border">{children}</ul>
    </section>
  );
}

function UnlinkButton({ label, disabled, pending, onClick }: { label: string; disabled?: boolean; pending: boolean; onClick: () => void }) {
  return (
    <Button
      variant="ghost"
      size="icon"
      type="button"
      title={label}
      aria-label={label}
      disabled={disabled || pending}
      onClick={onClick}
      className="h-6 w-6 shrink-0 text-muted-foreground hover:text-foreground"
    >
      <UiClose className="text-xs" />
    </Button>
  );
}

function PRLinkRow({ link, disabled, pending, onUnlink }: { link: TodoPRLinkItem; disabled?: boolean; pending: boolean; onUnlink: () => void }) {
  return (
    <li className="flex flex-wrap items-center gap-x-2 gap-y-1 px-3 py-2 text-xs">
      <span className={chipClass}>{link.relation}</span>
      <TodoPRBadge pr={{ number: link.number, url: link.url }} link />
      <span className="min-w-0 flex-1 basis-40 truncate text-sm text-foreground" title={link.title ?? link.url}>
        {link.title ?? <span className="text-muted-foreground">{link.repo}</span>}
      </span>
      {link.state && <span className="text-[11px] text-muted-foreground">{link.state}</span>}
      <span className="text-[11px] text-muted-foreground" title={`Linked by ${link.source}`}>{link.source}</span>
      <UnlinkButton label={`Unlink PR #${link.number}`} disabled={disabled} pending={pending} onClick={onUnlink} />
    </li>
  );
}

function TodoLinkRow({ link, disabled, pending, onOpenTodo, onUnlink }: {
  link: TodoLink;
  disabled?: boolean;
  pending: boolean;
  onOpenTodo?: (ref: string) => void;
  onUnlink: () => void;
}) {
  const title = link.target_title || link.target_short_id || link.target_id;
  return (
    <li className="flex flex-wrap items-center gap-x-2 gap-y-1 px-3 py-2 text-xs">
      {link.target_status && <StatusIcon status={link.target_status} />}
      <TodoRouteLink
        todoRef={link.target_short_id || link.target_id}
        onOpen={onOpenTodo}
        title={title}
        className="min-w-0 flex-1 basis-40 truncate text-sm font-medium text-foreground hover:underline"
      >
        {title}
      </TodoRouteLink>
      {link.target_short_id && <span className="font-mono text-[10px] text-muted-foreground">{link.target_short_id}</span>}
      <UnlinkButton label={`Unlink ${title}`} disabled={disabled} pending={pending} onClick={onUnlink} />
    </li>
  );
}

// What this todo is linked to: the pull requests that implement or touch it, the
// todos it duplicates, and the todos it merely relates to. A group with no
// entries is not drawn, and a todo with none at all draws nothing. Dependencies
// are not shown here — they have their own surface.
export function TodoLinks({ todoRef, dir, disabled, onOpenTodo }: {
  todoRef: string;
  dir: string;
  disabled?: boolean;
  onOpenTodo?: (ref: string) => void;
}) {
  const todoLinks = useQuery(todoLinksQueryOptions(dir, todoRef));
  const prLinks = useQuery(todoPRLinksQueryOptions(dir, todoRef));
  const unlinkTodo = useUnlinkTodo({ dir, todoRef });
  const unlinkPR = useUnlinkTodoPR({ dir, todoRef });

  const duplicates = (todoLinks.data ?? []).filter(link => link.relation === 'duplicates');
  const related = (todoLinks.data ?? []).filter(link => link.relation === 'related_to');
  const pullRequests = prLinks.data ?? [];
  const errors = [todoLinks.error, prLinks.error, unlinkTodo.error, unlinkPR.error].filter((error): error is Error => !!error);

  return (
    <>
      {errors.map(error => <p key={error.message} className="text-xs text-red-600">{error.message}</p>)}
      {pullRequests.length > 0 && (
        <LinkGroup title="Pull requests" icon={UiGitPr}>
          {pullRequests.map(link => (
            <PRLinkRow
              key={link.url}
              link={link}
              disabled={disabled}
              pending={unlinkPR.isPending}
              onUnlink={() => unlinkPR.mutate({ url: link.url })}
            />
          ))}
        </LinkGroup>
      )}
      {duplicates.length > 0 && (
        <LinkGroup title="Duplicates" icon={UiCopy}>
          {duplicates.map(link => (
            <TodoLinkRow
              key={link.target_id}
              link={link}
              disabled={disabled}
              pending={unlinkTodo.isPending}
              onOpenTodo={onOpenTodo}
              onUnlink={() => unlinkTodo.mutate({ target: link.target_id, relation: link.relation })}
            />
          ))}
        </LinkGroup>
      )}
      {related.length > 0 && (
        <LinkGroup title="Related" icon={UiLink}>
          {related.map(link => (
            <TodoLinkRow
              key={link.target_id}
              link={link}
              disabled={disabled}
              pending={unlinkTodo.isPending}
              onOpenTodo={onOpenTodo}
              onUnlink={() => unlinkTodo.mutate({ target: link.target_id, relation: link.relation })}
            />
          ))}
        </LinkGroup>
      )}
    </>
  );
}
