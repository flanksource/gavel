import { Icon } from '@flanksource/clicky-ui/data';
import { UiCircleOutline, UiLoader } from '@flanksource/clicky-ui/icons';
import type { Test } from '../types';
import { liveTaskGroups } from '../utils';

function statusOf(test: Test): string {
  return String((test.context as { status?: string } | undefined)?.status ?? '');
}

/**
 * The live clicky tasks behind a lint run, such as the per-review prompt tasks
 * under the `lint prompt` group, with each task's own status text (thinking,
 * tool: read_file). Renders nothing while no task is running or pending.
 */
export function LintTasks({ tests }: { tests: Test[] }) {
  const groups = liveTaskGroups(tests);
  if (groups.length === 0) return null;
  return (
    <section aria-label="Tasks" className="flex flex-col gap-1 border-b border-border px-density-3 py-density-2 text-xs">
      <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">Tasks</h3>
      {groups.map(group => {
        const rows = group.children?.length ? group.children : [group];
        return (
          <div key={group.name} className="flex flex-col gap-0.5">
            <span className="text-muted-foreground">{group.name}</span>
            <ul className="flex flex-col gap-0.5">
              {rows.map(row => {
                const running = statusOf(row) === 'running';
                return (
                  <li key={row.task_id ?? row.name} className="flex items-center gap-2">
                    <Icon icon={running ? UiLoader : UiCircleOutline} className={running ? 'animate-spin text-blue-500' : 'text-muted-foreground'} />
                    <span className="truncate font-medium">{row.name}</span>
                    <span className="truncate text-muted-foreground">{row.message || statusOf(row)}</span>
                  </li>
                );
              })}
            </ul>
          </div>
        );
      })}
    </section>
  );
}
