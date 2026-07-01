import { ListMenu, ListMenuHeader, ListMenuSection } from '@flanksource/clicky-ui/components';
import type { PRItem, PRSyncStatus, GavelResultsSummary, Project, ProcStatus } from '../types';
import { PRRow } from './PRRow';
import { ProcControl } from './ProcControl';
import { groupByOrg, prKey, computeCounts } from '../utils';
import { Avatar } from './Avatar';
import { GavelIcon } from './GavelIcon';
import { RepoIcon } from './RepoIcon';
import { TodoBadge } from './TodoBadge';
import { GitChangesBadge } from './GitChangesBadge';

interface Props {
  prs: PRItem[];
  selected: PRItem | null;
  onSelect: (pr: PRItem) => void;
  unread?: Record<string, boolean>;
  syncStatus?: Record<string, PRSyncStatus>;
  gavelResults?: Record<string, GavelResultsSummary>;
  projectsByRepo?: Record<string, Project>;
  procStatus?: Record<string, ProcStatus>;
  onProcChanged?: () => void;
  onProcEdit?: (project: Project) => void;
}

// GroupCounts shows per-group open (green) and failing (red) totals on the
// sidebar org/repo titles, with the overall count kept muted for context.
function GroupCounts({ items }: { items: PRItem[] }) {
  const c = computeCounts(items);
  return (
    <span className="flex items-center gap-1.5 shrink-0 text-xs font-normal tabular-nums">
      {c.failing > 0 && (
        <span className="text-red-600 inline-flex items-center gap-0.5" title={`${c.failing} failing`}>
          <GavelIcon name="codicon:error" />{c.failing}
        </span>
      )}
      {c.open > 0 && (
        <span className="text-green-600 inline-flex items-center gap-0.5" title={`${c.open} open`}>
          <GavelIcon name="codicon:git-pull-request" />{c.open}
        </span>
      )}
      <span className="text-muted-foreground" title={`${items.length} total`}>{items.length}</span>
    </span>
  );
}

export function PRList({ prs, selected, onSelect, unread, syncStatus, gavelResults, projectsByRepo, procStatus, onProcChanged, onProcEdit }: Props) {
  if (prs.length === 0) {
    return (
      <div className="p-6 text-center text-muted-foreground">
        <GavelIcon name="codicon:git-pull-request" className="text-3xl mb-2" />
        <p>No pull requests found</p>
      </div>
    );
  }

  const orgs = groupByOrg(prs);

  return (
    <ListMenu>
      {orgs.map(og => (
        <ListMenuSection key={og.org || '_'}>
          {og.org && (
            <ListMenuHeader className="z-20">
              <Avatar
                src={og.orgAvatarUrl}
                alt={og.org}
                size={20}
                rounded="md"
                href={`https://github.com/${og.org}`}
                title={og.org}
                colorKey={og.org}
              />
              <span className="text-sm font-semibold text-foreground truncate min-w-0 flex-1">{og.org}</span>
              <GroupCounts items={og.repos.flatMap(r => r.items)} />
            </ListMenuHeader>
          )}
          {og.repos.map(group => (
            <div key={group.repo}>
              <ListMenuHeader className="top-9 z-10 pl-6">
                <a
                  href={`https://github.com/${group.repo}`}
                  target="_blank"
                  rel="noopener"
                  className="inline-flex shrink-0"
                  onClick={(e) => e.stopPropagation()}
                >
                  <RepoIcon repo={group.repo} homepageUrl={group.repoHomepageUrl} size={20} />
                </a>
                <a
                  href={`https://github.com/${group.repo}`}
                  target="_blank"
                  rel="noopener"
                  onClick={(e) => e.stopPropagation()}
                  className="text-sm font-medium text-foreground truncate min-w-0 flex-1 hover:underline"
                  title={group.repo}
                >
                  {group.repoShort}
                </a>
                {onProcChanged && (
                  <ProcControl
                    repo={group.repo}
                    project={projectsByRepo?.[group.repo]}
                    status={procStatus?.[group.repo]}
                    onChanged={onProcChanged}
                    onEdit={onProcEdit}
                  />
                )}
                <TodoBadge counts={projectsByRepo?.[group.repo]?.todoCounts} />
                <GitChangesBadge count={procStatus?.[group.repo]?.gitChanges} />
                <GroupCounts items={group.items} />
              </ListMenuHeader>
              {group.items.map(pr => (
                <PRRow
                  key={prKey(pr)}
                  pr={pr}
                  selected={selected?.repo === pr.repo && selected?.number === pr.number}
                  unread={!!unread?.[prKey(pr)]}
                  syncStatus={syncStatus?.[prKey(pr)]}
                  gavelResults={gavelResults?.[prKey(pr)]}
                  onClick={() => onSelect(pr)}
                />
              ))}
            </div>
          ))}
        </ListMenuSection>
      ))}
    </ListMenu>
  );
}
