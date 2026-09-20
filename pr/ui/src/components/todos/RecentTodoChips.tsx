import { Button } from '@flanksource/clicky-ui/components';
import type { RecentTodo } from './recentTodos';

// RecentTodoChips is the one-click "add to existing" row on the new-todo form:
// the todos last created or commented on from the same source host.
export function RecentTodoChips({ label, entries, selectedRef = '', disabled = false, onPick }: {
  label: string;
  entries: RecentTodo[];
  selectedRef?: string;
  disabled?: boolean;
  onPick: (ref: string) => void;
}) {
  if (entries.length === 0) return null;
  return (
    <div className="flex flex-wrap items-center gap-1.5" role="group" aria-label={label}>
      <span className="text-xs text-muted-foreground">{label}</span>
      {entries.map(entry => {
        const selected = entry.ref === selectedRef;
        return (
          <Button
            key={entry.ref}
            type="button"
            variant="outline"
            size="sm"
            disabled={disabled}
            aria-pressed={selected}
            title={entry.title}
            onClick={() => onPick(entry.ref)}
            className={`h-6 max-w-64 rounded-full px-2.5 text-xs ${selected ? 'border-primary bg-primary/10 text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
          >
            <span className="truncate">
              {entry.shortId || entry.ref}
              <span className="text-muted-foreground"> · </span>
              {entry.title}
            </span>
          </Button>
        );
      })}
    </div>
  );
}
