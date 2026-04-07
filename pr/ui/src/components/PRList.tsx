import type { PRItem } from '../types';
import { PRRow } from './PRRow';
import { groupByRepo } from '../utils';

interface Props {
  prs: PRItem[];
  selected: PRItem | null;
  onSelect: (pr: PRItem) => void;
}

export function PRList({ prs, selected, onSelect }: Props) {
  if (prs.length === 0) {
    return (
      <div class="p-6 text-center text-gray-400">
        <iconify-icon icon="codicon:git-pull-request" class="text-3xl mb-2" />
        <p>No pull requests found</p>
      </div>
    );
  }

  const groups = groupByRepo(prs);
  const multiRepo = groups.length > 1;

  return (
    <div class="divide-y divide-gray-100">
      {groups.map(group => (
        <div key={group.repo}>
          {multiRepo && (
            <div class="px-3 py-1.5 bg-gray-50 text-xs font-semibold text-cyan-700 sticky top-0 border-b border-gray-100">
              <iconify-icon icon="codicon:repo" class="mr-1" />
              {group.repoShort}
              <span class="text-gray-400 font-normal ml-1">({group.items.length})</span>
            </div>
          )}
          {group.items.map(pr => (
            <PRRow
              key={`${pr.repo}#${pr.number}`}
              pr={pr}
              selected={selected?.repo === pr.repo && selected?.number === pr.number}
              onClick={() => onSelect(pr)}
            />
          ))}
        </div>
      ))}
    </div>
  );
}
