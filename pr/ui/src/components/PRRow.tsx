import type { PRItem } from '../types';
import { reviewColor, checkSummaryText, timeAgo } from '../utils';

interface Props {
  pr: PRItem;
  selected: boolean;
  onClick: () => void;
}

function prStatusIcon(pr: PRItem): { icon: string; color: string; title: string } {
  if (pr.isDraft) return { icon: '○', color: 'text-gray-400', title: 'Draft' };
  if (pr.state === 'MERGED') return { icon: '●', color: 'text-purple-600', title: 'Merged' };
  if (pr.state === 'CLOSED') return { icon: '●', color: 'text-red-600', title: 'Closed' };
  if (pr.checkStatus) {
    if (pr.checkStatus.failed > 0) return { icon: '✗', color: 'text-red-600', title: `${pr.checkStatus.failed} checks failed` };
    if (pr.checkStatus.running > 0) return { icon: '●', color: 'text-yellow-600', title: `${pr.checkStatus.running} checks running` };
    if (pr.checkStatus.passed > 0) return { icon: '✓', color: 'text-green-600', title: 'All checks passed' };
  }
  return { icon: '●', color: 'text-green-600', title: 'Open' };
}

function borderColor(pr: PRItem, selected: boolean): string {
  if (selected) return 'border-blue-500';
  if (pr.isDraft || pr.state === 'MERGED' || pr.state === 'CLOSED') return 'border-transparent';
  if (pr.checkStatus?.failed) return 'border-red-400';
  if (pr.checkStatus?.running) return 'border-yellow-400';
  return 'border-transparent';
}

export function PRRow({ pr, selected, onClick }: Props) {
  const hasConflict = !pr.isDraft && pr.mergeable === 'CONFLICTING';
  const status = prStatusIcon(pr);

  return (
    <div
      class={`px-3 py-2 cursor-pointer border-l-2 transition-colors ${borderColor(pr, selected)} ${
        selected ? 'bg-blue-50' : 'hover:bg-gray-50'
      }`}
      onClick={onClick}
    >
      <div class="flex items-center gap-2">
        <span class={`text-sm ${status.color}`} title={status.title}>
          {status.icon}
        </span>
        <span class="text-xs text-gray-400">#{pr.number}</span>
        <span class="text-sm font-medium text-gray-800 truncate flex-1">{pr.title}</span>
        {hasConflict && (
          <span class="text-xs text-red-500" title="Merge conflicts">
            <iconify-icon icon="codicon:git-merge" class="text-red-500" />
          </span>
        )}
        {pr.isCurrent && (
          <span class="text-xs bg-yellow-100 text-yellow-700 px-1.5 py-0.5 rounded" title="Current branch">
            current
          </span>
        )}
        {pr.isDraft && (
          <span class="text-xs bg-gray-100 text-gray-500 px-1.5 py-0.5 rounded">draft</span>
        )}
      </div>

      <div class="flex items-center gap-2 mt-1 text-xs text-gray-500">
        <span class="text-cyan-600">{pr.source}</span>
        <span>→</span>
        <span class="text-cyan-600">{pr.target}</span>

        {pr.isCurrent && (pr.ahead ?? 0) + (pr.behind ?? 0) > 0 && (
          <span class="text-yellow-600">↑{pr.ahead ?? 0}↓{pr.behind ?? 0}</span>
        )}

        {pr.reviewDecision && (
          <span class={`${reviewColor(pr.reviewDecision)} font-medium`}>
            {pr.reviewDecision.replace(/_/g, ' ')}
          </span>
        )}

        {hasConflict && <span class="text-red-500 font-medium">CONFLICTS</span>}

        {pr.checkStatus && <span>{checkSummaryText(pr.checkStatus)}</span>}

        <span class="ml-auto text-gray-400" title={pr.updatedAt}>{timeAgo(pr.updatedAt)}</span>
      </div>
    </div>
  );
}
