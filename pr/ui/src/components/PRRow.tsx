import { ListMenuItem } from '@flanksource/clicky-ui/components';
import type { PRItem, PRSyncStatus, GavelResultsSummary } from '../types';
import { reviewColor, checkSummaryText } from '../utils';
import { SyncIndicator } from './SyncIndicator';
import { Avatar } from './Avatar';
import { GavelIcon } from './GavelIcon';
import { RelativeTime } from './RelativeTime';

interface Props {
  pr: PRItem;
  selected: boolean;
  unread?: boolean;
  syncStatus?: PRSyncStatus;
  gavelResults?: GavelResultsSummary;
  onClick: () => void;
}

function prStatusIcon(pr: PRItem, gavel?: GavelResultsSummary): { icon: string; color: string; title: string } {
  // Gavel failures override draft/open status so a draft PR with a failing
  // test suite still shows up as a red PR row.
  if (gavel) {
    if (gavel.testsFailed > 0) {
      return { icon: 'octicon:x-circle-fill-16', color: 'text-red-600', title: `${gavel.testsFailed} gavel test${gavel.testsFailed !== 1 ? 's' : ''} failed` };
    }
    if (gavel.lintViolations > 0) {
      return { icon: 'octicon:alert-fill-16', color: 'text-yellow-600', title: `${gavel.lintViolations} lint violation${gavel.lintViolations !== 1 ? 's' : ''}` };
    }
    if ((gavel.benchRegressions ?? 0) > 0) {
      return { icon: 'octicon:graph-16', color: 'text-red-600', title: `${gavel.benchRegressions} bench regression${gavel.benchRegressions !== 1 ? 's' : ''}` };
    }
  }
  if (pr.state === 'MERGED') return { icon: 'octicon:git-merge-16', color: 'text-purple-600', title: 'Merged' };
  if (pr.state === 'CLOSED') return { icon: 'octicon:git-pull-request-closed-16', color: 'text-red-600', title: 'Closed' };
  if (pr.isDraft) return { icon: 'octicon:git-pull-request-draft-16', color: 'text-muted-foreground', title: 'Draft' };
  if (pr.checkStatus) {
    if (pr.checkStatus.failed > 0) return { icon: 'octicon:x-circle-fill-16', color: 'text-red-600', title: `${pr.checkStatus.failed} checks failed` };
    if (pr.checkStatus.running > 0) return { icon: 'octicon:dot-fill-16', color: 'text-yellow-600', title: `${pr.checkStatus.running} checks running` };
    if (pr.checkStatus.passed > 0) return { icon: 'octicon:check-circle-fill-16', color: 'text-green-600', title: 'All checks passed' };
  }
  return { icon: 'octicon:git-pull-request-16', color: 'text-green-600', title: 'Open' };
}

function borderColor(pr: PRItem, gavel?: GavelResultsSummary): string {
  if (pr.isDraft || pr.state === 'MERGED' || pr.state === 'CLOSED') return 'border-transparent';
  if (pr.checkStatus?.failed) return 'border-red-400';
  if (gavel && gavel.testsFailed > 0) return 'border-red-400';
  if (pr.checkStatus?.running) return 'border-yellow-400';
  if (gavel && (gavel.lintViolations > 0 || (gavel.benchRegressions ?? 0) > 0)) return 'border-yellow-400';
  return 'border-transparent';
}

function GavelBadges({ g }: { g: GavelResultsSummary }) {
  if (g.error) {
    return (
      <span className="inline-flex items-center text-yellow-600" title={`gavel: ${g.error}`}>
        <GavelIcon name="codicon:warning" />
      </span>
    );
  }
  const items: { icon: string; color: string; count: number; title: string }[] = [];
  if (g.testsPassed > 0) {
    items.push({
      icon: 'codicon:pass',
      color: 'text-green-600',
      count: g.testsPassed,
      title: `${g.testsPassed} test${g.testsPassed !== 1 ? 's' : ''} passed`,
    });
  }
  if (g.testsFailed > 0) {
    items.push({
      icon: 'codicon:error',
      color: 'text-red-600',
      count: g.testsFailed,
      title: `${g.testsFailed} test${g.testsFailed !== 1 ? 's' : ''} failed`,
    });
  }
  if (g.testsSkipped > 0) {
    items.push({
      icon: 'codicon:debug-step-over',
      color: 'text-muted-foreground',
      count: g.testsSkipped,
      title: `${g.testsSkipped} skipped`,
    });
  }
  if (g.lintViolations > 0) {
    items.push({
      icon: 'codicon:warning',
      color: 'text-yellow-600',
      count: g.lintViolations,
      title: `${g.lintViolations} lint violation${g.lintViolations !== 1 ? 's' : ''}`,
    });
  }
  if ((g.benchRegressions ?? 0) > 0) {
    items.push({
      icon: 'codicon:arrow-down',
      color: 'text-red-600',
      count: g.benchRegressions ?? 0,
      title: `${g.benchRegressions} bench regression${g.benchRegressions !== 1 ? 's' : ''}`,
    });
  }
  if (items.length === 0) return null;
  return (
    <span className="inline-flex items-center gap-1" aria-label="gavel results">
      {items.map((it, i) => (
        <span key={i} className={`inline-flex items-center ${it.color} tabular-nums leading-none`} title={it.title}>
          <GavelIcon name={it.icon} className="text-[12px]" />
          <span className="text-[11px] font-medium">{it.count}</span>
        </span>
      ))}
    </span>
  );
}

export function PRRow({ pr, selected, unread, syncStatus, gavelResults, onClick }: Props) {
  const hasConflict = !pr.isDraft && pr.mergeable === 'CONFLICTING';
  const status = prStatusIcon(pr, gavelResults);

  return (
    <ListMenuItem
      active={selected}
      accentClassName={borderColor(pr, gavelResults)}
      className="px-3 py-2"
      onClick={onClick}
    >
      <div className="flex items-center gap-2">
        <span
          className={`inline-block w-1.5 h-1.5 rounded-full shrink-0 ${unread ? 'bg-primary' : 'bg-transparent'}`}
          title={unread ? 'Unread — updated since last view' : ''}
          aria-label={unread ? 'unread' : ''}
        />
        <GavelIcon name={status.icon} className={`text-sm ${status.color} shrink-0`} title={status.title} />
        <a
          href={pr.url}
          target="_blank"
          rel="noopener"
          onClick={(e) => e.stopPropagation()}
          className="text-xs text-muted-foreground hover:text-foreground hover:underline shrink-0"
          title={`Open #${pr.number} on GitHub`}
        >
          #{pr.number}
        </a>
        <span className="text-sm truncate min-w-0 flex-1 font-medium text-foreground">{pr.title}</span>
        {hasConflict && (
          <span className="text-xs text-red-500" title="Merge conflicts">
            <GavelIcon name="octicon:git-merge-16" className="text-red-500" />
          </span>
        )}
        {pr.isCurrent && (
          <span className="text-xs bg-yellow-100 text-yellow-700 px-1.5 py-0.5 rounded" title="Current branch">
            current
          </span>
        )}
        {pr.isDraft && (
          <span className="text-xs bg-muted text-muted-foreground px-1.5 py-0.5 rounded">draft</span>
        )}
      </div>

      <div className="flex items-center gap-2 mt-1 text-xs text-muted-foreground">
        <span className="flex min-w-0 items-center gap-1 overflow-hidden">
          <span className="inline-flex min-w-0 items-center gap-0.5 underline decoration-dotted underline-offset-2">
            <GavelIcon name="codicon:git-branch" className="text-muted-foreground/70 shrink-0" />
            <span className="truncate">{pr.source}</span>
          </span>
          <span className="shrink-0">→</span>
          <span className="inline-flex min-w-0 items-center gap-0.5 underline decoration-dotted underline-offset-2">
            <GavelIcon name="codicon:git-branch" className="text-muted-foreground/70 shrink-0" />
            <span className="truncate">{pr.target}</span>
          </span>
        </span>

        {pr.isCurrent && (pr.ahead ?? 0) + (pr.behind ?? 0) > 0 && (
          <span className="text-yellow-600">↑{pr.ahead ?? 0}↓{pr.behind ?? 0}</span>
        )}

        {pr.reviewDecision && (
          <span className={`${reviewColor(pr.reviewDecision)} font-medium`}>
            {pr.reviewDecision.replace(/_/g, ' ')}
          </span>
        )}

        {hasConflict && <span className="text-red-500 font-medium">CONFLICTS</span>}

        {pr.checkStatus && <span>{checkSummaryText(pr.checkStatus)}</span>}

        {gavelResults && <GavelBadges g={gavelResults} />}

        <span className="ml-auto inline-flex items-center gap-1.5 text-muted-foreground">
          {syncStatus && <SyncIndicator status={syncStatus} />}
          {pr.author && (
            <Avatar
              src={pr.authorAvatarUrl}
              alt={pr.author}
              size={20}
              href={`https://github.com/${pr.author}`}
              title={`@${pr.author}`}
            />
          )}
          <RelativeTime iso={pr.updatedAt} title={pr.updatedAt} />
        </span>
      </div>
    </ListMenuItem>
  );
}
