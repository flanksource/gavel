import { useMemo, useState, type ReactNode } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { Timeline, type TimelineItem } from '@flanksource/clicky-ui/data';
import { UiAdd, UiCircleFilled, UiComment, UiDiff } from '@flanksource/clicky-ui/icons';
import type { TodoEvent } from '../../types';
import { timeAgo } from '../../utils';
import { GavelIcon } from '../GavelIcon';
import { Markdown } from '../Markdown';
import { RelativeTime } from '../RelativeTime';

type EventVisual = Pick<TimelineItem, 'icon' | 'tone'> & { action: string };

// Grite emits CamelCase event kinds (IssueCreated, IssueUpdated, CommentAdded).
// Map the known ones to an icon + disc tone + verb phrase; anything else falls
// back to a neutral dot with the kind humanized ("StatusChanged" -> "status changed").
function eventVisual(kind?: string): EventVisual {
  switch (kind) {
    case 'IssueCreated':
      return { icon: UiAdd, tone: 'success', action: 'created this issue' };
    case 'IssueUpdated':
      return { icon: UiDiff, tone: 'info', action: 'updated this issue' };
    case 'CommentAdded':
      return { icon: UiComment, tone: 'neutral', action: 'commented' };
    default:
      return { icon: UiCircleFilled, tone: 'neutral', action: humanizeKind(kind) };
  }
}

function humanizeKind(kind?: string): string {
  if (!kind) return 'logged an event';
  return kind.replace(/([a-z0-9])([A-Z])/g, '$1 $2').toLowerCase();
}

// Grite records the actor as an opaque id — a 32-char UUID without dashes or the
// canonical dashed form. Those carry no meaning in the timeline, so surface the
// actor only when it's a human-readable name; an id renders nothing.
const OPAQUE_ACTOR_ID = /^[0-9a-f]{32}$|^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function actorName(actor?: string): string | undefined {
  const name = actor?.trim();
  if (!name || OPAQUE_ACTOR_ID.test(name)) return undefined;
  return name;
}

// Provider events carry the zero-time sentinel when grite omits a timestamp.
// Returns the epoch ms (null when absent, used to sort), a relative age for
// display ("2h ago"), and the absolute timestamp for the hover title.
function eventTime(timestamp?: string): { ms: number | null; age: string; full: string } {
  if (!timestamp || timestamp.startsWith('0001-01-01')) return { ms: null, age: '', full: '' };
  const parsed = new Date(timestamp);
  const ms = parsed.getTime();
  if (Number.isNaN(ms)) return { ms: null, age: '', full: '' };
  return { ms, age: timeAgo(timestamp), full: parsed.toLocaleString() };
}

// Sort by timestamp (desc = newest first by default). Events without a usable
// timestamp keep their original relative order, pinned to the bottom either way.
function sortEvents(events: TodoEvent[], desc: boolean): TodoEvent[] {
  return events
    .map((event, index) => ({ event, index, ms: eventTime(event.timestamp).ms }))
    .sort((a, b) => {
      if (a.ms === null || b.ms === null) {
        if (a.ms === b.ms) return a.index - b.index;
        return a.ms === null ? 1 : -1;
      }
      return desc ? b.ms - a.ms : a.ms - b.ms;
    })
    .map(entry => entry.event);
}

const labelChip = (text: string, extra = '') => (
  <span className={`rounded bg-muted px-1.5 py-0.5 font-mono text-[11px] text-foreground ${extra}`}>{text}</span>
);

function toTimelineItem(event: TodoEvent, index: number): TimelineItem {
  const visual = eventVisual(event.kind);
  const time = eventTime(event.timestamp);
  const title = event.title?.trim();
  const body = event.body?.trim();
  const label = event.label?.trim();
  const oldLabel = event.old_label?.trim();
  const newLabel = event.new_label?.trim();

  // Timeline only renders the body bubble when `body` is set, so a title with
  // no body content is surfaced in the body slot rather than the header.
  let bodyHeader: ReactNode;
  let bodyNode: ReactNode;
  if (body) {
    bodyHeader = title || undefined;
    bodyNode = <Markdown text={body} className="text-xs" />;
  } else if (title) {
    bodyNode = title;
  }

  return {
    id: event.id || index,
    icon: visual.icon,
    tone: visual.tone,
    actor: actorName(event.actor),
    action: (
      <>
        {visual.action}
        {oldLabel || newLabel ? (
          <span className="ml-1.5 inline-flex items-center gap-1">
            {oldLabel && labelChip(oldLabel, 'line-through opacity-70')}
            <span className="text-muted-foreground">→</span>
            {newLabel && labelChip(newLabel)}
          </span>
        ) : (
          label && <span className="ml-1.5">{labelChip(label)}</span>
        )}
        {event.short_id && <span className="ml-1.5 font-mono text-[11px]">{event.short_id}</span>}
      </>
    ),
    timestamp: time.age && event.timestamp ? <RelativeTime iso={event.timestamp} title={time.full} /> : undefined,
    bodyHeader,
    body: bodyNode,
  };
}

export function TodoTimeline({ events }: { events: TodoEvent[] }) {
  const [desc, setDesc] = useState(true);
  const items = useMemo(() => sortEvents(events, desc).map(toTimelineItem), [events, desc]);

  return (
    <section className="rounded-md border border-border bg-background">
      <div className="flex items-center gap-2 px-3 py-2">
        <GavelIcon name="codicon:history" className="shrink-0 text-xs text-muted-foreground" />
        <span className="min-w-0 flex-1 truncate text-xs font-semibold uppercase text-muted-foreground">History</span>
        <span className="text-xs tabular-nums text-muted-foreground">{events.length}</span>
        <Button
          variant="ghost"
          type="button"
          onClick={() => setDesc(d => !d)}
          className="inline-flex h-auto items-center gap-1 rounded px-1.5 py-0.5 text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground"
          title={desc ? 'Newest first' : 'Oldest first'}
          aria-label="Toggle timeline sort order"
        >
          <GavelIcon name={desc ? 'codicon:chevron-down' : 'codicon:chevron-up'} className="text-xs" />
          {desc ? 'Newest' : 'Oldest'}
        </Button>
      </div>
      <div className="border-t border-border px-3 py-3">
        <Timeline items={items} />
      </div>
    </section>
  );
}
