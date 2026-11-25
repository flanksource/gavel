import { timeAgo } from '../utils';
import { useNow } from '../useNow';

// RelativeTime renders a 'Xm ago' label that refreshes itself once per second by
// subscribing to the shared useNow() clock. Because it owns the subscription,
// only this leaf re-renders on the tick — its parents (PR rows, the process
// table, the detail panel) are not reconciled every second.
export function RelativeTime({ iso, title, className }: {
  iso: string;
  title?: string;
  className?: string;
}) {
  useNow();
  return <span className={className} title={title}>{timeAgo(iso)}</span>;
}
