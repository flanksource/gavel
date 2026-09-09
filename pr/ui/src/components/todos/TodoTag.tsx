import { useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { UiAdd, UiClose } from '@flanksource/clicky-ui/icons';

import { TagIcon } from '../../icons/tags';
import { TagPicker } from './TagPicker';
import { tagClasses, tagTextClasses, normalizeTag } from './tagPalette';
import type { ResolvedTag, TagIndex } from './tagResolve';
import { tagCandidates, todoReservedLabels, todoVisibleLabels } from './tagResolve';
import type { TodoItem } from '../../types';

// TodoTag is the one tag chip. Every tag surface — the detail header, the mobile
// menu, list rows — renders through it, so a tag looks the same everywhere.
//
// It is deliberately NOT built on clicky-ui's TagList: that component hardcodes
// tone="neutral" and forwards neither colour nor icon to its Badge, and it mounts
// a floating-ui HoverCard per chip, which is unusable inside list rows. Adding
// colour/icon support to TagList upstream is the right long-term home; until
// then this composes the same primitives directly.
export function TodoTag({ tag, size = 'xs', showKey = true, glyphOnly = false, className = '' }: {
  tag: ResolvedTag;
  size?: 'xxs' | 'xs';
  /** Compact surfaces drop the namespace key and show only the value. */
  showKey?: boolean;
  /** Render only the coloured glyph — for single-line rows with no room for text. */
  glyphOnly?: boolean;
  className?: string;
}) {
  const title = tag.description ? `${tag.token} — ${tag.description}` : tag.token;
  const label = showKey ? tag.token : tag.value;

  if (glyphOnly) {
    return (
      <span
        className={`inline-flex shrink-0 items-center ${tagTextClasses(tag.color)} ${className}`}
        title={title}
        data-tag={tag.token}
      >
        <TagIcon iconKey={tag.icon} iconify={tag.iconify} className="text-[13px]" />
      </span>
    );
  }

  const padding = size === 'xxs' ? 'px-1.5 py-0 text-[10px]' : 'px-2 py-0.5 text-[11px]';
  return (
    <span
      className={`inline-flex max-w-[14rem] shrink-0 items-center gap-1 rounded font-medium ${padding} ${tagClasses(tag.color)} ${className}`}
      title={title}
      data-tag={tag.token}
    >
      <TagIcon iconKey={tag.icon} iconify={tag.iconify} className="shrink-0 text-[12px]" />
      <span className="truncate">{label}</span>
    </span>
  );
}

// TodoTagRow renders a todo's tags with an overflow count.
//
// The "+N" is a plain span rather than clicky's HoverCard: this renders inside
// list rows, and a floating-ui popover per row is both a performance and a
// jsdom-stability problem.
export function TodoTagRow({
  labels,
  index,
  max = 6,
  size = 'xs',
  showKey = true,
  glyphOnly = false,
  className = '',
}: {
  labels: string[];
  index: TagIndex;
  max?: number;
  size?: 'xxs' | 'xs';
  showKey?: boolean;
  glyphOnly?: boolean;
  className?: string;
}) {
  if (labels.length === 0) return null;

  const visible = labels.slice(0, max);
  const overflow = labels.length - visible.length;

  return (
    <span className={`inline-flex min-w-0 items-center gap-1 ${className}`}>
      {visible.map(label => (
        <TodoTag
          key={label}
          tag={index.resolve(label)}
          size={size}
          showKey={showKey}
          glyphOnly={glyphOnly}
        />
      ))}
      {overflow > 0 && (
        <span
          className="inline-flex shrink-0 items-center rounded bg-muted/30 px-1.5 py-0.5 text-[10px] text-muted-foreground"
          title={labels.slice(max).join(', ')}
        >
          +{overflow}
        </span>
      )}
    </span>
  );
}

// TodoTagField is the editable tag surface on the detail header: the chips plus
// add/remove.
//
// It sends the full replacement set — the edited visible labels PLUS the todo's
// reserved lifecycle labels (status:/priority:/session:). The API replaces the
// whole set, so dropping the reserved half here would silently delete the
// todo's lifecycle state along with a tag edit.
export function TodoTagField({ todo, index, counts = {}, disabled = false, onChange }: {
  todo: TodoItem;
  index: TagIndex;
  /** Per-label todo counts, so the picker can lead with this project's own vocabulary. */
  counts?: Record<string, number>;
  disabled?: boolean;
  /** Receives the complete label set to persist, reserved labels included. */
  onChange: (labels: string[]) => void;
}) {
  const [adding, setAdding] = useState(false);
  const visible = todoVisibleLabels(todo);
  const reserved = todoReservedLabels(todo);

  const commit = (next: string[]) => onChange([...reserved, ...next]);

  function add(name: string) {
    setAdding(false);
    const value = normalizeTag(name);
    if (!value || visible.some(label => normalizeTag(label) === value)) return;
    commit([...visible, value]);
  }

  // Everything offerable — definitions plus labels already in use here — minus
  // what this todo already carries.
  const candidates = tagCandidates(index, counts)
    .filter(def => !visible.some(label => normalizeTag(label) === def.name));

  return (
    <span className="relative inline-flex flex-wrap items-center gap-1">
      {visible.map(label => (
        <span key={label} className="group inline-flex items-center">
          <TodoTag tag={index.resolve(label)} />
          {!disabled && (
            <Button
              type="button"
              size="icon"
              variant="ghost"
              title={`Remove ${label}`}
              aria-label={`Remove ${label}`}
              onClick={() => commit(visible.filter(other => other !== label))}
              className="ml-0.5 h-4 w-4 text-muted-foreground opacity-0 transition-opacity hover:text-destructive focus-visible:opacity-100 group-hover:opacity-100"
            >
              <UiClose className="text-[9px]" />
            </Button>
          )}
        </span>
      ))}

      {!disabled && (
        <span className="relative inline-flex">
          <Button
            type="button"
            size="icon"
            variant="ghost"
            title="Add tag"
            aria-label="Add tag"
            aria-expanded={adding}
            onClick={() => setAdding(open => !open)}
            className="h-5 w-5 text-muted-foreground hover:text-foreground"
          >
            <UiAdd className="text-[10px]" />
          </Button>
          {adding && (
            <TagPicker
              candidates={candidates}
              counts={counts}
              index={index}
              onPick={add}
              onClose={() => setAdding(false)}
              placeholder="Filter or create…"
              emptyLabel="This todo already has every defined tag."
            />
          )}
        </span>
      )}
    </span>
  );
}
