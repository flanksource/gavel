import { Button } from '@flanksource/clicky-ui/components';
import type { TodoCounts, TodoStatus } from '../../types';
import { statusClass } from './format';
import { STATUS_FILTER_DEFS } from './todoFilter';

// TodoFilterBar renders one toggle pill per status that has todos, plus the
// trailing "Closed" pill. A pill is active (colored) when its status is shown
// and dimmed when hidden; clicking flips it. Completed starts hidden so the list
// defaults to open work only.
export function TodoFilterBar({ counts, hidden, onToggle }: {
  counts: TodoCounts;
  hidden: Set<TodoStatus>;
  onToggle: (status: TodoStatus) => void;
}) {
  const defs = STATUS_FILTER_DEFS.filter(d => counts[d.countKey] > 0);
  if (defs.length === 0) return null;
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      {defs.map(def => {
        const active = !hidden.has(def.status);
        const count = counts[def.countKey];
        return (
          <Button
            key={def.status}
            variant="ghost"
            type="button"
            onClick={() => onToggle(def.status)}
            aria-pressed={active}
            title={active ? `Hide ${def.label.toLowerCase()}` : `Show ${def.label.toLowerCase()}`}
            className={`inline-flex h-auto items-center gap-1 rounded-full border px-2 py-0.5 text-[11px] font-medium transition-colors ${
              active ? statusClass(def.status) : 'border-border bg-transparent text-muted-foreground opacity-60 hover:opacity-100'
            }`}
          >
            <span>{def.label}</span>
            <span className="tabular-nums">{count}</span>
          </Button>
        );
      })}
    </div>
  );
}
