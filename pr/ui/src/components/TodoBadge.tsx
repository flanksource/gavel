import type { TodoCounts } from '../types';
import { GavelIcon } from './GavelIcon';

export function TodoBadge({ counts }: { counts?: TodoCounts }) {
  if (!counts || counts.open <= 0) return null;
  return (
    <span
      className="inline-flex items-center gap-1 rounded bg-blue-500/10 px-1 text-[10px] font-medium tabular-nums text-blue-600 dark:text-blue-400"
      title={`${counts.open} open todo${counts.open === 1 ? '' : 's'}${counts.inProgress > 0 ? `, ${counts.inProgress} in progress` : ''}`}
    >
      <GavelIcon name="codicon:check" className="text-[10px]" />
      {counts.open}
      {counts.inProgress > 0 && (
        <span className="inline-flex items-center gap-0.5 text-blue-700 dark:text-blue-300">
          <GavelIcon name="codicon:debug-start" className="text-[10px]" />
          {counts.inProgress}
        </span>
      )}
    </span>
  );
}
