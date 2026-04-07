import type { PRItem } from '../types';
import { stateIcon, stateColor, reviewColor, checkSummaryText, timeAgo } from '../utils';

interface Props {
  pr: PRItem;
  selected: boolean;
  onClick: () => void;
}

export function PRRow({ pr, selected, onClick }: Props) {
  const hasConflict = !pr.isDraft && pr.mergeable === 'CONFLICTING';

  return (
    <div
      class={`px-3 py-2 cursor-pointer border-l-2 transition-colors ${
        selected
          ? 'bg-blue-50 border-blue-500'
          : 'border-transparent hover:bg-gray-50'
      }`}
      onClick={onClick}
    >
      <div class="flex items-center gap-2">
        <span class={`text-sm ${stateColor(pr.state, pr.isDraft)}`} title={pr.isDraft ? 'Draft' : pr.state}>
          {stateIcon(pr.state, pr.isDraft)}
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
