import { Button } from '@flanksource/clicky-ui/components';
import { GavelIcon } from './GavelIcon';
import type { Tab } from '../routes';

// The search bar is a permanent fixture of the AppShell top bar (every tab), so
// its query survives tab switches and acts as one global filter. Each tab reads
// the shared query to narrow its own list (PRs, todos, test runs, activity), so
// the placeholder/label reflect whatever the active tab searches.
const PLACEHOLDERS: Record<Tab, string> = {
  prs: 'Search pull requests, branches, #id…',
  todos: 'Search todos, #id, labels…',
  tests: 'Search test runs…',
  activity: 'Search requests, URLs…',
};

const LABELS: Record<Tab, string> = {
  prs: 'Search pull requests',
  todos: 'Search todos',
  tests: 'Search test runs',
  activity: 'Search activity',
};

export function SearchBar({ tab, query, onChange }: {
  tab: Tab;
  query: string;
  onChange: (query: string) => void;
}) {
  return (
    <div className="flex w-full items-center gap-2 rounded-md border border-border bg-muted px-3 py-1.5">
      <GavelIcon name="codicon:search" className="text-muted-foreground text-sm shrink-0" />
      <input
        value={query}
        onChange={(e) => onChange((e.target as HTMLInputElement).value)}
        placeholder={PLACEHOLDERS[tab]}
        aria-label={LABELS[tab]}
        className="flex-1 min-w-0 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
      />
      {query && (
        <Button
          variant="ghost"
          type="button"
          onClick={() => onChange('')}
          className="text-muted-foreground hover:text-foreground shrink-0 h-auto p-0"
          aria-label="Clear search"
        >
          <GavelIcon name="codicon:close" className="text-xs" />
        </Button>
      )}
    </div>
  );
}
