import type { RateLimit } from '../types';
import { GavelIcon } from './GavelIcon';

interface Props {
  fetchedAt: string;
  nextFetchIn: number;
  paused: boolean;
  rateLimit?: RateLimit;
  error?: string;
  onRefresh: () => void;
  onPause: () => void;
  networkBusy?: boolean;
}

// Compact polling status for the app bar: last-fetch age + countdown, a
// pause/resume + refresh control, and the GitHub rate-limit gauge with a
// health dot. Counts live in the toolbar; this is just liveness.
export function PollStatus({ fetchedAt, nextFetchIn, paused, rateLimit, error, onRefresh, onPause, networkBusy }: Props) {
  const ago = fetchedAt ? timeAgoShort(fetchedAt) : 'never';
  const countdown = fetchedAt
    ? Math.max(0, nextFetchIn - Math.floor((Date.now() - new Date(fetchedAt).getTime()) / 1000))
    : nextFetchIn;
  const rlLow = rateLimit && rateLimit.remaining < 100;

  return (
    <div className="flex items-center gap-2 text-xs text-muted-foreground">
      <span className="inline-flex items-center gap-1 font-mono" title={`Refreshes every ${nextFetchIn}s`}>
        <GavelIcon name="codicon:clock" />
        {ago}
        {paused
          ? <span className="text-yellow-600 dark:text-yellow-400">(paused)</span>
          : countdown > 0 && <span className="opacity-60">({countdown}s)</span>}
      </span>
      <button
        onClick={onPause}
        className={`transition-colors ${paused ? 'text-yellow-600 hover:text-foreground' : 'hover:text-foreground'}`}
        title={paused ? 'Resume polling' : 'Pause polling'}
      >
        <GavelIcon name={paused ? 'codicon:debug-start' : 'codicon:debug-pause'} />
      </button>
      <button
        onClick={onRefresh}
        className={`transition-colors ${networkBusy ? 'text-primary' : 'hover:text-primary'}`}
        title={networkBusy ? 'Loading...' : 'Refresh now'}
      >
        <GavelIcon name={networkBusy ? 'svg-spinners:ring-resize' : 'codicon:refresh'} />
      </button>
      {rateLimit && (
        <span
          className={`inline-flex items-center gap-1 font-mono ${rlLow ? 'text-red-600 dark:text-red-400' : ''}`}
          title={`GitHub API: ${rateLimit.used}/${rateLimit.limit} used (${rateLimit.resource}), resets ${new Date(rateLimit.reset * 1000).toLocaleTimeString()}`}
        >
          <span className={`w-1.5 h-1.5 rounded-full ${rlLow ? 'bg-red-500' : 'bg-green-500'}`} />
          {rateLimit.remaining}<span className="opacity-50">/{rateLimit.limit}</span>
        </span>
      )}
      {error && (
        <span className="text-red-600 dark:text-red-400 inline-flex items-center gap-1" title={error}>
          <GavelIcon name="codicon:warning" />
        </span>
      )}
    </div>
  );
}

function timeAgoShort(iso: string): string {
  const d = new Date(iso);
  const s = Math.floor((Date.now() - d.getTime()) / 1000);
  if (s < 5) return 'just now';
  if (s < 60) return `${s}s ago`;
  if (s < 3600) return `${Math.floor(s / 60)}m ago`;
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`;
  return `${Math.floor(s / 86400)}d ago`;
}
