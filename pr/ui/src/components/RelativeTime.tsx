import { timeAgo, timeAgoShort } from '../utils';
import { useNow } from '../useNow';

// RelativeTime renders a 'Xm ago' label that refreshes itself once per second by
// subscribing to the shared useNow() clock. Because it owns the subscription,
// only this leaf re-renders on the tick — its parents (PR rows, the process
// table, the detail panel) are not reconciled every second. `short` switches to
// second-level granularity for very fresh timestamps.
export function RelativeTime({ iso, short, title, className }: {
  iso: string;
  short?: boolean;
  title?: string;
  className?: string;
}) {
  useNow();
  return <span className={className} title={title}>{short ? timeAgoShort(iso) : timeAgo(iso)}</span>;
}
