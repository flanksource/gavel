import type { PRItem, PRSyncStatus } from '../types';
import { PRRow } from './PRRow';
import { groupByRepo, prKey } from '../utils';
import { Avatar } from './Avatar';

interface Props {
  prs: PRItem[];
  selected: PRItem | null;
  onSelect: (pr: PRItem) => void;
  unread?: Record<string, boolean>;
  syncStatus?: Record<string, PRSyncStatus>;
}

export function PRList({ prs, selected, onSelect, unread, syncStatus }: Props) {
  if (prs.length === 0) {
    return (
      <div class="p-6 text-center text-gray-400">
        <iconify-icon icon="codicon:git-pull-request" class="text-3xl mb-2" />
        <p>No pull requests found</p>
      </div>
    );
  }

  const groups = groupByRepo(prs);

  return (
    <div class="divide-y divide-gray-100">
      {groups.map(group => (
        <div key={group.repo}>
          <div class="px-3 py-2 bg-gray-50 sticky top-0 border-b border-gray-200 flex items-center gap-2 z-10">
            <Avatar
              src={group.repoAvatarUrl}
              alt={group.repo}
              size={22}
              rounded="md"
              href={`https://github.com/${group.repo}`}
              title={group.repo}
            />
            <div class="flex items-baseline gap-1 min-w-0 flex-1">
              {group.repoOwner && (
                <span class="text-xs text-gray-400 font-normal truncate">{group.repoOwner}/</span>
              )}
              <span class="text-sm font-semibold text-gray-800 truncate">{group.repoShort}</span>
            </div>
            <span class="text-xs text-gray-400 font-normal shrink-0">{group.items.length}</span>
          </div>
          {group.items.map(pr => (
            <PRRow
              key={prKey(pr)}
              pr={pr}
              selected={selected?.repo === pr.repo && selected?.number === pr.number}
              unread={!!unread?.[prKey(pr)]}
              syncStatus={syncStatus?.[prKey(pr)]}
              onClick={() => onSelect(pr)}
            />
          ))}
        </div>
      ))}
    </div>
  );
}
