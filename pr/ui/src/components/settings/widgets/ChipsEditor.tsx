import { useState } from 'react';
import { UiClose } from '@flanksource/clicky-ui/icons';

// ChipsEditor edits a string list (glob / path patterns) as removable chips with
// a trailing free-text input. Entries are shown sorted A–Z and deduped; the
// stored value keeps whatever order the caller had (order is not meaningful for
// append-and-dedupe config lists). Long lists collapse past `collapseAt`.
interface Props {
  items: string[];
  onChange: (next: string[]) => void;
  placeholder?: string;
  collapseAt?: number;
  unit?: string;
}

export function ChipsEditor({
  items,
  onChange,
  placeholder = 'Add pattern…',
  collapseAt = 16,
  unit = 'patterns',
}: Props) {
  const [expanded, setExpanded] = useState(false);
  const [draft, setDraft] = useState('');

  const sorted = [...items].sort((a, b) => a.toLowerCase().localeCompare(b.toLowerCase()));
  const shown = expanded ? sorted : sorted.slice(0, collapseAt);
  const hidden = sorted.length - shown.length;

  const add = () => {
    const value = draft.trim();
    if (value && !items.includes(value)) onChange([...items, value]);
    setDraft('');
  };
  const remove = (item: string) => onChange(items.filter(x => x !== item));

  const moreBtn = 'px-1.5 py-0.5 text-[11.5px] font-semibold text-primary hover:underline';

  return (
    <div className="rounded-lg border border-border bg-muted/40 px-2.5 pb-1.5 pt-2">
      <div className="flex flex-wrap items-center gap-1.5">
        {shown.map(item => (
          <span
            key={item}
            className="inline-flex items-center gap-1 rounded-md border border-border bg-background py-0.5 pl-2 pr-0.5 font-mono text-[11.5px] text-foreground"
          >
            {item}
            <button
              type="button"
              onClick={() => remove(item)}
              title={`Remove ${item}`}
              className="inline-flex h-4 w-4 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-destructive"
            >
              <UiClose className="text-[10px]" />
            </button>
          </span>
        ))}
        {hidden > 0 && (
          <button type="button" className={moreBtn} onClick={() => setExpanded(true)}>
            +{hidden} more
          </button>
        )}
        {expanded && sorted.length > collapseAt && (
          <button type="button" className={moreBtn} onClick={() => setExpanded(false)}>
            Show less
          </button>
        )}
        <input
          value={draft}
          placeholder={placeholder}
          onChange={e => setDraft(e.target.value)}
          onKeyDown={e => {
            if (e.key === 'Enter') {
              e.preventDefault();
              add();
            }
          }}
          onBlur={add}
          className="min-w-[110px] flex-[1_0_110px] bg-transparent px-0.5 py-1 font-mono text-[11.5px] text-foreground outline-none"
        />
      </div>
      <div className="mt-1.5 flex items-center gap-2 border-t border-border pt-1.5 text-[11px] text-muted-foreground">
        <span className="font-mono">{sorted.length}</span> {unit}
        <span className="flex-1" />
        <span>sorted A–Z · deduped</span>
      </div>
    </div>
  );
}
