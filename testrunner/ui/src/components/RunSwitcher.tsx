import { Select } from '@flanksource/clicky-ui/components';
import { UiHistory } from '@flanksource/clicky-ui/icons';
import type { RunHistory } from '../hooks/use-run-history';
import { runResultLabel, runStartLabel, runViewHref } from '../runHistory';

const LATEST = '';

/** Switches the page to another recorded run, or back to the latest results. */
export function RunSwitcher({ history, navigate }: { history: RunHistory; navigate: (href: string) => void }) {
  const current = history.current ?? LATEST;
  const options = [
    { value: LATEST, label: 'Latest results' },
    ...history.runs.map(run => ({
      value: run.runId,
      label: `${runStartLabel(run)} · ${run.kind} · ${runResultLabel(run)}${history.latest.includes(run.runId) ? ' (latest)' : ''}`,
    })),
  ];
  // A run opened from a page not loaded here still names itself.
  if (current !== LATEST && !history.runs.some(run => run.runId === current)) options.splice(1, 0, { value: current, label: current });

  return (
    <label className="inline-flex min-w-0 items-center gap-1 text-sm" data-run-switcher="">
      <UiHistory className="shrink-0 text-muted-foreground" aria-hidden />
      <span className="sr-only">Run</span>
      <Select
        className="h-7 max-w-[22rem] truncate py-0 text-xs"
        value={current}
        options={options}
        onChange={event => {
          const value = event.target.value;
          if (value === current) return;
          const run = history.runs.find(r => r.runId === value);
          navigate(value === LATEST ? history.latestHref : run ? runViewHref(run) : history.latestHref);
        }}
      />
    </label>
  );
}
