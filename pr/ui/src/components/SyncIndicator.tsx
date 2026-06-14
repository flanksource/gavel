import { useState } from 'react';
import type { PRSyncStatus } from '../types';
import { timeAgo } from '../utils';

interface Props {
  status: PRSyncStatus;
}

function phaseLabel(phase?: string): string {
  switch (phase) {
    case 'metadata': return 'PR metadata';
    case 'runs': return 'workflow runs';
    case 'gavel': return 'gavel results';
    default: return 'details';
  }
}

function stateConfig(status: PRSyncStatus): { icon: string; color: string; title: string } {
  switch (status.state) {
    case 'queued':
      return { icon: 'codicon:sync-ignored', color: 'text-gray-400', title: 'Queued for sync' };
    case 'syncing':
      return { icon: 'svg-spinners:ring-resize', color: 'text-blue-500', title: `Syncing: ${phaseLabel(status.phase)}...` };
    case 'up-to-date':
      return { icon: 'codicon:sync', color: 'text-green-500', title: status.lastSynced ? `Synced ${timeAgo(status.lastSynced)}` : 'Synced' };
    case 'out-of-date':
      return { icon: 'codicon:refresh', color: 'text-yellow-500', title: 'Updated since last sync' };
    case 'error':
      return { icon: 'codicon:warning', color: 'text-red-500', title: status.error || 'Sync error' };
    default:
      return { icon: 'codicon:sync-ignored', color: 'text-gray-400', title: '' };
  }
}

export function SyncIndicator({ status }: Props) {
  const [hover, setHover] = useState(false);
  const cfg = stateConfig(status);

  return (
    <span
      className="relative inline-flex items-center"
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
    >
      <iconify-icon icon={cfg.icon} className={`${cfg.color} text-[11px]`} title={cfg.title} />
      {hover && <HoverCard status={status} />}
    </span>
  );
}

function HoverCard({ status }: { status: PRSyncStatus }) {
  const stateLabels: Record<string, string> = {
    'queued': 'Queued',
    'syncing': 'Syncing',
    'up-to-date': 'Up to date',
    'out-of-date': 'Out of date',
    'error': 'Error',
  };

  const stateColors: Record<string, string> = {
    'queued': 'text-gray-500',
    'syncing': 'text-blue-600',
    'up-to-date': 'text-green-600',
    'out-of-date': 'text-yellow-600',
    'error': 'text-red-600',
  };

  return (
    <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1.5 z-20 bg-white border border-gray-200 rounded-md shadow-lg px-2.5 py-1.5 whitespace-nowrap text-[11px]">
      <div className={`font-medium ${stateColors[status.state] || 'text-gray-600'}`}>
        {stateLabels[status.state] || status.state}
      </div>
      {status.state === 'syncing' && status.phase && (
        <div className="text-gray-500 mt-0.5">Fetching {phaseLabel(status.phase)}...</div>
      )}
      {status.lastSynced && (
        <div className="text-gray-400 mt-0.5">Last synced: {timeAgo(status.lastSynced)}</div>
      )}
      {status.error && (
        <div className="text-red-500 mt-0.5 max-w-48 truncate">{status.error}</div>
      )}
      <div className="absolute top-full left-1/2 -translate-x-1/2 -mt-px">
        <div className="w-1.5 h-1.5 bg-white border-b border-r border-gray-200 rotate-45 -translate-y-1" />
      </div>
    </div>
  );
}
