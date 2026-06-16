import { useState } from 'react';
import type { PRItem, RateLimit } from '../types';
import { ProgressBar } from './ProgressBar';
import { computeCounts } from '../utils';
import { RadialGauge } from '@flanksource/clicky-ui/data';

interface Props {
  prs: PRItem[];
  fetchedAt: string;
  error?: string;
  nextFetchIn: number;
  paused: boolean;
  rateLimit?: RateLimit;
  onRefresh: () => void;
  onPause: () => void;
  networkBusy?: boolean;
}

export function Summary({ prs, fetchedAt, error, nextFetchIn, onRefresh, paused, rateLimit, onPause, networkBusy }: Props) {
  const counts = computeCounts(prs);

  return (
    <div className="flex flex-col items-end gap-1">
      {/* Controls only — fixed widths so the bar never reflows as values change.
          Per-state counts live on the sidebar org/repo titles. */}
      <div className="flex gap-3 items-center text-sm">
        <RateLimitGauge rateLimit={rateLimit} />
        <SyncStatus fetchedAt={fetchedAt} nextFetchIn={nextFetchIn} paused={paused} error={error} networkBusy={networkBusy} />
        <button
          onClick={onPause}
          className={`w-5 flex justify-center transition-colors ${paused ? 'text-yellow-500 hover:text-green-600' : 'text-muted-foreground hover:text-yellow-500'}`}
          title={paused ? 'Resume polling' : 'Pause polling'}
        >
          <iconify-icon icon={paused ? 'codicon:debug-start' : 'codicon:debug-pause'} />
        </button>
        <button
          onClick={onRefresh}
          className={`w-5 flex justify-center transition-colors ${networkBusy ? 'text-blue-500' : 'text-muted-foreground hover:text-blue-600'}`}
          title={networkBusy ? 'Loading...' : 'Refresh now'}
        >
          <iconify-icon icon={networkBusy ? 'svg-spinners:ring-resize' : 'codicon:refresh'} />
        </button>
      </div>
      {prs.length > 0 && (
        <div className="w-64">
          <ProgressBar
            segments={[
              { count: counts.passing, color: 'bg-green-500', label: 'passing' },
              { count: counts.running, color: 'bg-yellow-400', label: 'running' },
              { count: counts.failing, color: 'bg-red-500', label: 'failing' },
              { count: counts.noChecks, color: 'bg-gray-300', label: 'no checks' },
            ]}
            total={prs.length}
          />
        </div>
      )}
    </div>
  );
}

function RateLimitGauge({ rateLimit }: { rateLimit?: RateLimit }) {
  const github = <iconify-icon icon="mdi:github" className="text-[11px] text-muted-foreground" />;
  if (!rateLimit) {
    return (
      <RadialGauge value={0} max={1} tone="neutral" size={26} thickness={3} center={github} label="Rate Limit" title="GitHub API rate limit (pending)" />
    );
  }
  const pct = rateLimit.limit > 0 ? rateLimit.remaining / rateLimit.limit : 0;
  const tone = pct > 0.5 ? 'success' : pct > 0.15 ? 'warning' : 'danger';
  const resetAt = new Date(rateLimit.reset * 1000).toLocaleTimeString();
  return (
    <RadialGauge
      value={rateLimit.remaining}
      max={rateLimit.limit}
      tone={tone}
      size={26}
      thickness={3}
      center={github}
      label="Rate Limit"
      title={`GitHub API (${rateLimit.resource}): ${rateLimit.remaining}/${rateLimit.limit} remaining, ${rateLimit.used} used, resets ${resetAt}`}
    />
  );
}

interface SyncProps {
  fetchedAt: string;
  nextFetchIn: number;
  paused: boolean;
  error?: string;
  networkBusy?: boolean;
}

function SyncStatus({ fetchedAt, nextFetchIn, paused, error, networkBusy }: SyncProps) {
  const [hover, setHover] = useState(false);
  const ago = fetchedAt ? timeAgoShort(fetchedAt) : 'never';
  const countdown = fetchedAt
    ? Math.max(0, nextFetchIn - Math.floor((Date.now() - new Date(fetchedAt).getTime()) / 1000))
    : nextFetchIn;

  const cfg = error
    ? { icon: 'codicon:warning', color: 'text-red-500', label: 'Sync error' }
    : networkBusy
      ? { icon: 'svg-spinners:ring-resize', color: 'text-blue-500', label: 'Syncing…' }
      : { icon: 'codicon:sync', color: 'text-green-500', label: 'Synced' };

  return (
    <span
      className="relative inline-flex w-5 justify-center"
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
    >
      <iconify-icon icon={cfg.icon} className={cfg.color} title={cfg.label} />
      {hover && (
        <div className="absolute top-full right-0 mt-1.5 z-30 bg-card text-card-foreground border border-border rounded-md shadow-lg px-2.5 py-1.5 whitespace-nowrap text-xs space-y-0.5">
          <div className={`font-medium ${cfg.color}`}>{cfg.label}</div>
          <div className="text-muted-foreground">Last synced: {ago}</div>
          <div className="text-muted-foreground">
            {paused ? 'Polling paused' : countdown > 0 ? `Next refresh in ${countdown}s` : 'Refreshing…'}
          </div>
          {error && <div className="text-red-500 max-w-64 truncate">{error}</div>}
        </div>
      )}
    </span>
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
