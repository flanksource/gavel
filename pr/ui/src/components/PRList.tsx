import { useState } from 'react';
import { IconButton, ListMenu, ListMenuHeader, ListMenuSection } from '@flanksource/clicky-ui/components';
import { useProcStatus } from '../procStatusQuery';
import type { PRItem, PRSyncStatus, GavelResultsSummary, Project, TodoCounts } from '../types';
import { PRRow } from './PRRow';
import { ProcControl } from './ProcControl';
import { groupByOrg, prKey, computeCounts, type RepoGroup } from '../utils';
import { Avatar } from './Avatar';
import { UiChevronDown, UiChevronRight, UiCircleXFilled, UiError, UiGitMerge, UiGitPr } from '@flanksource/clicky-ui/icons';
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
  onProcChanged?: () => void;
  // The signed-in GitHub login, so rows leave out the viewer's own avatar.
  viewer?: string;
  // Phone layout: headers keep one action count, rows keep one status icon.
  mobile?: boolean;
}

const plural = (n: number, word: string) => `${n} ${word}${n === 1 ? '' : 's'}`;

function conflictCount(items: PRItem[]): number {
  return items.filter(pr => !pr.isDraft && pr.mergeable === 'CONFLICTING').length;
}

// groupSummary is a group header's full breakdown, its title tooltip.
function groupSummary(items: PRItem[], todos?: TodoCounts, gitChanges?: number): string {
  const c = computeCounts(items);
  const conflicts = conflictCount(items);
  return [
    plural(items.length, 'pull request'),
    c.open > 0 && `${c.open} open`,
    c.failing > 0 && `${c.failing} failing`,
    conflicts > 0 && `${conflicts} conflicting`,
    todos && todos.open > 0 && `${plural(todos.open, 'open todo')}${todos.inProgress > 0 ? `, ${todos.inProgress} in progress` : ''}`,
    gitChanges && gitChanges > 0 && plural(gitChanges, 'uncommitted change'),
  ].filter(Boolean).join(' · ');
}

// ActionCount is a phone header's single counter: the failing-check count, or
// when nothing fails, the merge-conflict count.
function ActionCount({ items }: { items: PRItem[] }) {
  const failing = computeCounts(items).failing;
  const conflicts = conflictCount(items);
  if (failing === 0 && conflicts === 0) return null;
  const [Icon, count, label] = failing > 0
    ? [UiCircleXFilled, failing, `${failing} failing`]
    : [UiGitMerge, conflicts, `${conflicts} conflicting`];
  return (
    <span role="img" aria-label={label} className="inline-flex shrink-0 items-center gap-0.5 text-xs font-normal tabular-nums text-red-600">
      <Icon />{count}
    </span>
  );
}

// GroupCounts shows per-group open (green) and failing (red) totals on the
// sidebar org/repo titles, with the overall count kept muted for context.
function GroupCounts({ items }: { items: PRItem[] }) {
  const c = computeCounts(items);
  return (
    <span className="flex items-center gap-1.5 shrink-0 text-xs font-normal tabular-nums">
      {c.failing > 0 && (
        <span className="text-red-600 inline-flex items-center gap-0.5" title={`${c.failing} failing`}>
          <UiError />{c.failing}
        </span>
      )}
      {c.open > 0 && (
        <span className="text-green-600 inline-flex items-center gap-0.5" title={`${c.open} open`}>
          <UiGitPr />{c.open}
        </span>
      )}
      <span className="text-muted-foreground" title={`${items.length} total`}>{items.length}</span>
    </span>
  );
}

export function PRList({ prs, selected, onSelect, unread, syncStatus, gavelResults, projectsByRepo, onProcChanged, viewer, mobile }: Props) {
  if (prs.length === 0) {
    return (
      <div className="p-6 text-center text-muted-foreground">
        <UiGitPr className="text-3xl mb-2" />
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
            <ListMenuHeader className="z-20" title={groupSummary(og.repos.flatMap(r => r.items))}>
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
              {mobile
                ? <ActionCount items={og.repos.flatMap(r => r.items)} />
                : <GroupCounts items={og.repos.flatMap(r => r.items)} />}
            </ListMenuHeader>
          )}
          {og.repos.map(group => (
            <RepoSection
              key={group.repo}
              group={group}
              selected={selected}
              onSelect={onSelect}
              unread={unread}
              syncStatus={syncStatus}
              gavelResults={gavelResults}
              projectsByRepo={projectsByRepo}
              onProcChanged={onProcChanged}
              viewer={viewer}
              mobile={mobile}
            />
          ))}
        </ListMenuSection>
      ))}
    </ListMenu>
  );
}

type RepoSectionProps = Omit<Props, 'prs'> & { group: RepoGroup };

// RepoSection is one repository's header and pull requests; the arrow beside the
// repository hides its pull requests, keeping the header and counts.
function RepoSection({ group, selected, onSelect, unread, syncStatus, gavelResults, projectsByRepo, onProcChanged, viewer, mobile }: RepoSectionProps) {
  const procStatus = useProcStatus();
  const [open, setOpen] = useState(true);
  const todoCounts = projectsByRepo?.[group.repo]?.todoCounts;
  const gitChanges = procStatus?.[group.repo]?.gitChanges;
  return (
    <div>
      <ListMenuHeader className="top-9 z-10 pl-2" title={groupSummary(group.items, todoCounts, gitChanges)}>
        <IconButton
          icon={open ? UiChevronDown : UiChevronRight}
          label={`${open ? 'Collapse' : 'Expand'} ${group.repo}`}
          aria-expanded={open}
          onClick={() => setOpen(o => !o)}
          className="shrink-0 text-xs text-muted-foreground hover:text-foreground"
        />
        <a
          href={`https://github.com/${group.repo}`}
          target="_blank"
          rel="noopener"
          className="inline-flex shrink-0"
          aria-label={`Open ${group.repo} on GitHub`}
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
          />
        )}
        {mobile ? (
          <ActionCount items={group.items} />
        ) : (
          <>
            <TodoBadge counts={todoCounts} />
            <GitChangesBadge count={gitChanges} />
            <GroupCounts items={group.items} />
          </>
        )}
      </ListMenuHeader>
      {open && group.items.map(pr => (
        <PRRow
          key={prKey(pr)}
          pr={pr}
          viewer={viewer}
          mobile={mobile}
          selected={selected?.repo === pr.repo && selected?.number === pr.number}
          unread={!!unread?.[prKey(pr)]}
          syncStatus={syncStatus?.[prKey(pr)]}
          gavelResults={gavelResults?.[prKey(pr)]}
          onClick={() => onSelect(pr)}
        />
      ))}
    </div>
  );
}
