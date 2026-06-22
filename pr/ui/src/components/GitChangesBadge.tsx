import { GavelIcon } from './GavelIcon';

// GitChangesBadge surfaces a workspace's uncommitted change count (staged,
// unstaged, untracked). Rendered only when there is at least one change.
export function GitChangesBadge({ count }: { count?: number }) {
  if (!count || count <= 0) return null;
  return (
    <span
      className="inline-flex items-center gap-0.5 rounded bg-amber-500/15 px-1 text-[10px] font-medium tabular-nums text-amber-600 dark:text-amber-400"
      title={`${count} uncommitted change${count === 1 ? '' : 's'}`}
    >
      <GavelIcon name="codicon:diff" className="text-[10px]" />
      {count}
    </span>
  );
}
