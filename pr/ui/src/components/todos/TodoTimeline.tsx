import type { ReactNode } from 'react';
import { Timeline, type TimelineItem } from '@flanksource/clicky-ui/data';
import { UiAdd, UiCircleFilled, UiComment, UiDiff } from '@flanksource/clicky-ui/icons';
import type { TodoEvent } from '../../types';
import { Markdown } from '../Markdown';

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

// Provider events carry the zero-time sentinel when grite omits a timestamp.
function eventTimestamp(timestamp?: string): string {
  if (!timestamp || timestamp.startsWith('0001-01-01')) return '';
  const parsed = new Date(timestamp);
  return Number.isNaN(parsed.getTime()) ? '' : parsed.toLocaleString();
}

function toTimelineItem(event: TodoEvent, index: number): TimelineItem {
  const visual = eventVisual(event.kind);
  const title = event.title?.trim();
  const body = event.body?.trim();
  const label = event.label?.trim();

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
    actor: event.actor || undefined,
    action: (
      <>
        {visual.action}
        {label && <span className="ml-1.5 rounded bg-muted px-1.5 py-0.5 font-mono text-[11px] text-foreground">{label}</span>}
        {event.short_id && <span className="ml-1.5 font-mono text-[11px]">{event.short_id}</span>}
      </>
    ),
    timestamp: eventTimestamp(event.timestamp) || undefined,
    bodyHeader,
    body: bodyNode,
  };
}

export function TodoTimeline({ events }: { events: TodoEvent[] }) {
  return <Timeline items={events.map(toTimelineItem)} />;
}
