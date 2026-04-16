import { useState } from 'preact/hooks';
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
      return { icon: '○', color: 'text-gray-300', title: 'Queued for sync' };
    case 'syncing':
      return { icon: '', color: 'text-blue-400', title: `Syncing: ${phaseLabel(status.phase)}...` };
    case 'up-to-date':
      return { icon: '✓', color: 'text-green-400', title: status.lastSynced ? `Synced ${timeAgo(status.lastSynced)}` : 'Synced' };
    case 'out-of-date':
      return { icon: '○', color: 'text-yellow-400', title: 'Updated since last sync' };
    case 'error':
      return { icon: '!', color: 'text-red-400', title: status.error || 'Sync error' };
    default:
      return { icon: '○', color: 'text-gray-300', title: '' };
  }
}

export function SyncIndicator({ status }: Props) {
  const [hover, setHover] = useState(false);
  const cfg = stateConfig(status);

  return (
    <span
      class="relative inline-flex items-center"
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
    >
      {status.state === 'syncing' ? (
        <iconify-icon icon="svg-spinners:ring-resize" class="text-blue-400" style="font-size: 8px" />
      ) : (
        <span class={`text-[8px] leading-none ${cfg.color}`} title={cfg.title}>{cfg.icon}</span>
      )}
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
    <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-1.5 z-20 bg-white border border-gray-200 rounded-md shadow-lg px-2.5 py-1.5 whitespace-nowrap text-[11px]">
      <div class={`font-medium ${stateColors[status.state] || 'text-gray-600'}`}>
        {stateLabels[status.state] || status.state}
      </div>
      {status.state === 'syncing' && status.phase && (
        <div class="text-gray-500 mt-0.5">Fetching {phaseLabel(status.phase)}...</div>
      )}
      {status.lastSynced && (
        <div class="text-gray-400 mt-0.5">Last synced: {timeAgo(status.lastSynced)}</div>
      )}
      {status.error && (
        <div class="text-red-500 mt-0.5 max-w-48 truncate">{status.error}</div>
      )}
      <div class="absolute top-full left-1/2 -translate-x-1/2 -mt-px">
        <div class="w-1.5 h-1.5 bg-white border-b border-r border-gray-200 rotate-45 -translate-y-1" />
      </div>
    </div>
  );
}
