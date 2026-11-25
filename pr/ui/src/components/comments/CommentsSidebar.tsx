import { useLayoutEffect, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { DEFAULT_COMMENT_STATUSES, DriftBadge } from '@flanksource/clicky-ui/comments';
import {
  EMPTY_INBOX_FILTERS, filterInbox, groupInbox, inboxAuthorOptions, inboxPrompt, writeClipboard,
  INBOX_KIND_OPTIONS, INBOX_PHASE_OPTIONS, INBOX_STATUS_OPTIONS,
  type InboxFilters, type InboxItem, type InboxOption, type InboxPromptTarget,
} from '@flanksource/clicky-ui/comments/gavel';
import { Button, FilterBar, SplitButton, type FilterBarFilter } from '@flanksource/clicky-ui/components';
import { Badge } from '@flanksource/clicky-ui/data';
import { UiChatDots, UiFile, UiGlobe, UiImage } from '@flanksource/clicky-ui/icons';
import { cn } from '@flanksource/clicky-ui/utils';
import type { Project } from '../../types';
import { useCopyFeedback } from '../../useCopyFeedback';
import { ageShort } from '../../utils';
import { commentClient, commentListKey, type CommentListings } from './queries';

const targetIcons = { file: UiFile, page: UiGlobe, image: UiImage } as const;
const targetKey = ({ scope, uri }: { scope: string; uri: string }) => `${scope}\n${uri}`;
const phaseLabels = new Map(INBOX_PHASE_OPTIONS.map(({ value, label }) => [value, label]));
const statuses = new Map(DEFAULT_COMMENT_STATUSES.map(status => [status.value, status]));

function CommentRow({ item, selected, onSelect }: { item: InboxItem; selected: boolean; onSelect: () => void }) {
  const { thread, anchor, line, activity } = item;
  const title = thread.bodyMarkdown.trim().split('\n', 1)[0] || `Comment on ${item.label}`;
  const status = statuses.get(thread.status);
  const phase = phaseLabels.get(thread.phase) ?? thread.phase;
  const replies = thread.messages?.length ?? 0;
  const meta = [thread.author.label ?? thread.author.id, line && `Line ${line}`, phase, replies > 0 && `${replies} ${replies === 1 ? 'reply' : 'replies'}`].filter(Boolean).join(' · ');
  const statusBadge = <Badge tone={status?.tone === 'default' ? undefined : status?.tone} variant="soft" size="xs" clickToCopy={false} className="whitespace-nowrap">{status?.label ?? thread.status}</Badge>;
  return (
    <li>
      <Button type="button" variant="ghost" aria-current={selected || undefined} aria-label={`View ${title}`}
        aria-description={[thread.unread && 'Unread', status?.label ?? thread.status, meta].filter(Boolean).join(' · ')} onClick={onSelect}
        className={cn('grid h-auto min-h-12 w-full min-w-0 items-center gap-2 rounded-none border-b border-border/50 px-3 py-2 text-left hover:bg-muted/50 focus-visible:ring-inset @max-[560px]:grid-cols-[minmax(0,1fr)_auto] @min-[560px]:grid-cols-[minmax(0,1fr)_6rem_3rem]', selected && 'bg-accent')}>
        <span className="flex min-w-0 items-start gap-2">
          <span className={cn('mt-1.5 size-2 shrink-0 rounded-full', thread.unread && 'bg-primary')} title={thread.unread ? 'Unread' : undefined} />
          <span className="min-w-0 flex-1 space-y-0.5">
            <span className="block truncate text-sm font-medium" title={title}>{title}</span>
            <span className="flex min-w-0 items-center gap-1.5 text-xs text-muted-foreground">
              <span className="shrink-0 @min-[560px]:hidden">{statusBadge}</span>
              <span className="truncate" title={meta}>{meta}</span>
              {anchor && anchor.state !== 'matched' && <DriftBadge state={anchor.state} />}
            </span>
          </span>
        </span>
        <span className="@max-[560px]:hidden">{statusBadge}</span>
        <time dateTime={activity} title={`Last activity ${activity}`} className="whitespace-nowrap text-right text-xs font-normal tabular-nums text-muted-foreground @max-[560px]:self-end">{ageShort(activity)}</time>
      </Button>
    </li>
  );
}

function CopyComments({ items, listings, filters }: { items: InboxItem[]; listings: CommentListings['listings']; filters: InboxFilters }) {
  const { copyState, copyError, beginCopy, resetCopyFeedback } = useCopyFeedback();
  async function copy(target: InboxPromptTarget) {
    beginCopy();
    try {
      await writeClipboard(Promise.resolve().then(() => inboxPrompt({
        items, scopes: listings.map(({ scope }) => scope), filters, target,
        baseUrl: window.location.origin, url: window.location.href,
      })));
      resetCopyFeedback('copied');
    } catch (cause) {
      resetCopyFeedback('error', cause instanceof Error ? cause.message : String(cause));
    }
  }
  return <>
    <SplitButton label={copyState === 'copied' ? 'Copied' : 'Copy for LLM'} size="sm" variant="outline"
      title="Copy agent instructions" onClick={() => void copy('LLM')}
      items={(['Claude', 'Codex'] as const).map(target => ({ label: `Copy for ${target}`, onSelect: () => void copy(target) }))} />
    {copyError && <p role="alert" className="basis-full text-xs text-destructive">{copyError}</p>}
  </>;
}

export function CommentsSidebar({ query, projects, projectsLoaded, projectError, filters, selectedId, onSelect, onFiltersChange }: {
  query: CommentListings;
  projects: Project[];
  projectsLoaded: boolean;
  projectError: string;
  filters: InboxFilters;
  selectedId: string;
  onSelect: (id: string) => void;
  onFiltersChange: (filters: InboxFilters) => void;
}) {
  const { items, listings, errors, pending } = query;
  const [readError, setReadError] = useState('');
  const cache = useQueryClient();
  const latest = useRef(filters);
  useLayoutEffect(() => { latest.current = filters; }, [filters]);
  // FilterBar clears each field in one event. Keep those patches together.
  function set(patch: Partial<InboxFilters>) {
    latest.current = { ...latest.current, ...patch };
    onFiltersChange(latest.current);
  }
  async function markRead({ scope, uri }: { scope: string; uri: string }) {
    setReadError('');
    try {
      await commentClient(scope).markRead(uri);
    } catch (cause) {
      setReadError(cause instanceof Error ? cause.message : String(cause));
    } finally {
      await cache.invalidateQueries({ queryKey: commentListKey(scope) });
    }
  }

  const matched = filterInbox(items, filters);
  const groups = groupInbox(filters.review ? matched.filter(item => item.thread.phase === 'awaiting_outcome') : matched);
  const reviewCount = filterInbox(items, { ...filters, review: true }).filter(item => item.thread.phase === 'awaiting_outcome').length;
  const unresolved = items.filter(item => statuses.get(item.thread.status)?.unresolved).length;
  const unreadTargets = new Set(items.filter(item => item.thread.unread).map(targetKey));
  const error = [projectError, ...errors].filter(Boolean).join(' · ');
  const choice = (key: 'project' | 'status' | 'phase' | 'author' | 'kind', label: string, options: InboxOption[]): FilterBarFilter => ({
    key, kind: 'enum', label, value: filters[key], options, onChange: value => set({ [key]: value }),
  });
  return (
    <section aria-label="Comment inbox" className="@container flex h-full min-h-0 min-w-0 flex-col">
      <header className="flex shrink-0 flex-wrap items-center justify-between gap-2 px-3 pt-2">
        <h1 className="flex min-w-0 flex-wrap items-baseline gap-x-2 text-sm font-semibold">
          Comments <span className="text-xs font-normal text-muted-foreground">{listings.length > 0 && `${unresolved} unresolved`}</span>
        </h1>
        <CopyComments items={groups.flatMap(group => group.items)} listings={listings} filters={filters} />
      </header>
      <FilterBar className="shrink-0 border-b border-border px-3"
        search={{ value: filters.q, onChange: q => set({ q }), placeholder: 'Comment text or path…', ariaLabel: 'Search comments' }}
        filters={[
          choice('project', 'Project', projects.map(({ name }) => ({ value: name, label: name }))),
          choice('status', 'Status', INBOX_STATUS_OPTIONS),
          choice('phase', 'Phase', INBOX_PHASE_OPTIONS),
          choice('author', 'Author', inboxAuthorOptions(items)),
          choice('kind', 'Target', INBOX_KIND_OPTIONS),
          { key: 'unread', kind: 'boolean', label: 'Unread only', value: filters.unread, onChange: unread => set({ unread }) },
          { key: 'review', kind: 'boolean', label: `Needs review (${reviewCount})`, value: filters.review, onChange: review => set({ review }) },
        ]} />
      {error && <p role="alert" className="shrink-0 break-words px-3 py-2 text-xs text-destructive">{listings.length ? 'Refreshing comments failed' : 'Could not load comments'}. {error}</p>}
      {readError && <p role="alert" className="shrink-0 break-words px-3 py-2 text-xs text-destructive">{readError}</p>}
      <div className="min-h-0 flex-1 overflow-y-auto">
        {groups.map(group => {
          const TargetIcon = targetIcons[group.kind];
          return <div key={group.key}>
            <header className="flex min-w-0 items-center gap-2 border-b border-border bg-muted/40 px-3 py-1.5 text-xs text-muted-foreground">
              <TargetIcon className="size-4 shrink-0" />
              <h2 className="min-w-0 flex-1">
                <span className="block truncate font-medium text-foreground" title={group.label}>{group.label}</span>
                <span className="block truncate" title={group.project}>{group.project}</span>
              </h2>
              {unreadTargets.has(group.key) && <Button size="sm" variant="ghost" className="shrink-0 [@media(pointer:coarse)]:min-h-11" aria-label={`Mark ${group.label} read`} onClick={() => void markRead(group)}>Mark read</Button>}
            </header>
            <ol className="m-0 list-none p-0">
              {group.items.map(item => <CommentRow key={item.thread.id} item={item} selected={item.thread.id === selectedId} onSelect={() => onSelect(item.thread.id)} />)}
            </ol>
          </div>;
        })}
        {groups.length === 0 && (!projectsLoaded || pending ? (
          <p role="status" className="p-3 text-sm text-muted-foreground">Loading comments…</p>
        ) : (!error || listings.length > 0) && (
          <div className="flex h-full flex-col items-center justify-center gap-3 p-4 text-center">
            <UiChatDots className="size-6 text-muted-foreground" />
            <p className="text-sm font-medium">{items.length ? 'No threads match these filters' : 'No comment threads yet'}</p>
            {items.length > 0 && <Button size="sm" variant="outline" onClick={() => onFiltersChange(EMPTY_INBOX_FILTERS)}>Clear filters</Button>}
          </div>
        ))}
      </div>
    </section>
  );
}
