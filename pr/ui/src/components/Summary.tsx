import type { PRItem, RateLimit } from '../types';
import { ProgressBar } from './ProgressBar';
import { computeCounts } from '../utils';

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
  const ago = fetchedAt ? timeAgoShort(fetchedAt) : 'never';
  const countdown = fetchedAt
    ? Math.max(0, nextFetchIn - Math.floor((Date.now() - new Date(fetchedAt).getTime()) / 1000))
    : nextFetchIn;

  return (
    <div className="flex flex-col items-end gap-1">
      <div className="flex gap-3 text-sm text-muted-foreground items-center">
        <span className="font-medium text-foreground">{prs.length} PRs</span>
        {counts.open > 0 && <><Sep /><span className="text-green-600">{counts.open} open</span></>}
        {counts.failing > 0 && <><Sep /><span className="text-red-600">{counts.failing} failing</span></>}
        {counts.running > 0 && <><Sep /><span className="text-yellow-600">{counts.running} running</span></>}
        {counts.merged > 0 && <><Sep /><span className="text-purple-600">{counts.merged} merged</span></>}
        <Sep />
        <span className="text-muted-foreground" title={`Refreshes every ${nextFetchIn}s`}>
          <iconify-icon icon="codicon:clock" className="mr-0.5" />
          {ago}
          {paused
            ? <span className="text-yellow-500 ml-1">(paused)</span>
            : countdown > 0 && <span className="text-muted-foreground/50 ml-1">({countdown}s)</span>
          }
        </span>
        <button
          onClick={onPause}
          className={`transition-colors ${paused ? 'text-yellow-500 hover:text-green-600' : 'text-muted-foreground hover:text-yellow-500'}`}
          title={paused ? 'Resume polling' : 'Pause polling'}
        >
          <iconify-icon icon={paused ? 'codicon:debug-start' : 'codicon:debug-pause'} />
        </button>
        <button
          onClick={onRefresh}
          className={`transition-colors ${networkBusy ? 'text-blue-500' : 'text-muted-foreground hover:text-blue-600'}`}
          title={networkBusy ? 'Loading...' : 'Refresh now'}
        >
          <iconify-icon icon={networkBusy ? 'svg-spinners:ring-resize' : 'codicon:refresh'} />
        </button>
        {rateLimit && (
          <span
            className={`text-xs ${rateLimit.remaining < 100 ? 'text-red-500' : 'text-muted-foreground'}`}
            title={`API: ${rateLimit.used}/${rateLimit.limit} used (${rateLimit.resource}), resets ${new Date(rateLimit.reset * 1000).toLocaleTimeString()}`}
          >
            {rateLimit.remaining}/{rateLimit.limit}
          </span>
        )}
        {error && (
          <span className="text-red-500 text-xs" title={error}>
            <iconify-icon icon="codicon:warning" className="mr-0.5" />
            Error
          </span>
        )}
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

function Sep() {
  return <span className="text-muted-foreground/50">|</span>;
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
