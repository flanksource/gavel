import { useState, type ReactNode } from 'react';
import { ContextMenu, type DropdownMenuItem } from '@flanksource/clicky-ui/components';
import { cn } from '@flanksource/clicky-ui/utils';

export interface ListingCardProps {
  /** Leading identity, about 32px square: a repo avatar, a provider meter. */
  icon: ReactNode;
  /** Pinned to the icon's bottom-right corner: a health dot, a spinner. */
  status?: ReactNode;
  title: ReactNode;
  /** Hover text of the title, which truncates. */
  titleHint?: string;
  /** Follows the title on its line. */
  titleBadges?: ReactNode;
  /** Starts the second line: where the item lives (repository, branch). */
  meta?: ReactNode;
  /** Follows the meta on the second line; clips before the meta truncates. */
  badges?: ReactNode;
  /** Ends the second line. */
  age?: ReactNode;
  /** Before the icon, outside the selection target: an expand chevron. */
  leading?: ReactNode;
  /** After the content, outside the selection target: row buttons. */
  actions?: ReactNode;
  /** Item actions offered on right-click of the whole card, keeping the row free of buttons. */
  menuItems?: DropdownMenuItem[];
  selected?: boolean;
  /** Accessible name of the selection button. */
  label: string;
  onSelect: () => void;
}

/**
 * One row of a sidebar listing, laid out like a session row: an icon with a
 * status corner, a title line and a meta line ending in the item's age.
 * The whole card selects the item; its title is the keyboard-focusable
 * button. Interactive slot content (links, buttons) stops propagation so it
 * never doubles as a selection.
 */
export function ListingCard({ icon, status, title, titleHint, titleBadges, meta, badges, age, leading, actions, menuItems, selected, label, onSelect }: ListingCardProps) {
  const [card, setCard] = useState<HTMLDivElement | null>(null);
  return (
    // The title button is the keyboard and screen-reader target; the card's
    // click only widens its pointer target.
    <div
      ref={setCard}
      data-testid="listing-card"
      onClick={onSelect}
      className={cn(
        'group flex min-w-0 cursor-pointer items-center gap-1 border-b border-border/50 py-2 pl-3 pr-2 hover:bg-muted/50',
        selected && 'bg-accent hover:bg-accent',
      )}
    >
      {leading ? <span className="-ml-2 inline-flex shrink-0" onClick={event => event.stopPropagation()}>{leading}</span> : null}
      <span className="flex min-w-0 flex-1 items-center gap-2">
        <span data-testid="listing-card-icon" className="relative inline-flex size-8 shrink-0 items-center justify-center">
          {icon}
          {status ? (
            <span data-testid="listing-card-status" className="absolute -bottom-1 -right-1 flex rounded-full bg-background p-0.5">{status}</span>
          ) : null}
        </span>
        <span className="min-w-0 flex-1 space-y-1">
          <span data-testid="listing-card-title" className="flex min-w-0 items-center gap-2">
            <button
              type="button"
              aria-label={label}
              aria-current={selected ? 'true' : undefined}
              title={titleHint}
              className="min-w-0 truncate rounded-sm text-left text-sm font-medium text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              {title}
            </button>
            {titleBadges}
          </span>
          {meta || badges || age ? (
            <span data-testid="listing-card-meta" className="flex min-w-0 items-center gap-2 text-xs text-muted-foreground">
              {meta ? <span className="flex min-w-0 items-center gap-1 overflow-hidden">{meta}</span> : null}
              {badges ? <span className="flex min-w-0 shrink-[1000] items-center gap-1 overflow-hidden">{badges}</span> : null}
              {age ? <span className="ml-auto shrink-0 whitespace-nowrap tabular-nums">{age}</span> : null}
            </span>
          ) : null}
        </span>
      </span>
      {actions ? <span className="flex shrink-0 items-center" onClick={event => event.stopPropagation()}>{actions}</span> : null}
      {/* React bubbles portal events through the tree, so menu picks must not reach the card's select. */}
      {menuItems?.length ? <span className="contents" onClick={event => event.stopPropagation()}><ContextMenu contextTarget={card} menuLabel={`${label} actions`} menuItems={menuItems} /></span> : null}
    </div>
  );
}
