import type { PRItem, PRSyncStatus, GavelResultsSummary } from '../types';
import { reviewColor, checkSummaryText, timeAgo } from '../utils';
import { SyncIndicator } from './SyncIndicator';
import { Avatar } from './Avatar';

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
  if (pr.isDraft) return { icon: 'octicon:git-pull-request-draft-16', color: 'text-gray-400', title: 'Draft' };
  if (pr.checkStatus) {
    if (pr.checkStatus.failed > 0) return { icon: 'octicon:x-circle-fill-16', color: 'text-red-600', title: `${pr.checkStatus.failed} checks failed` };
    if (pr.checkStatus.running > 0) return { icon: 'octicon:dot-fill-16', color: 'text-yellow-600', title: `${pr.checkStatus.running} checks running` };
    if (pr.checkStatus.passed > 0) return { icon: 'octicon:check-circle-fill-16', color: 'text-green-600', title: 'All checks passed' };
  }
  return { icon: 'octicon:git-pull-request-16', color: 'text-green-600', title: 'Open' };
}

function borderColor(pr: PRItem, selected: boolean, gavel?: GavelResultsSummary): string {
  if (selected) return 'border-blue-500';
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
        <iconify-icon icon="codicon:warning" />
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
      color: 'text-gray-500',
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
          <iconify-icon icon={it.icon} className="text-[12px]" />
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
    <div
      className={`px-3 py-2 cursor-pointer border-l-2 transition-colors ${borderColor(pr, selected, gavelResults)} ${
        selected ? 'bg-blue-50' : unread ? 'hover:bg-gray-50' : 'hover:bg-gray-50'
      }`}
      onClick={onClick}
    >
      <div className="flex items-center gap-2">
        <span
          className={`inline-block w-1.5 h-1.5 rounded-full shrink-0 ${unread ? 'bg-blue-600' : 'bg-transparent'}`}
          title={unread ? 'Unread — updated since last view' : ''}
          aria-label={unread ? 'unread' : ''}
        />
        <iconify-icon icon={status.icon} className={`text-sm ${status.color} shrink-0`} title={status.title} />
        <span className="text-xs text-gray-400">#{pr.number}</span>
        <span className={`text-sm truncate flex-1 ${unread ? 'font-semibold text-gray-900' : 'font-medium text-gray-800'}`}>{pr.title}</span>
        {hasConflict && (
          <span className="text-xs text-red-500" title="Merge conflicts">
            <iconify-icon icon="octicon:git-merge-16" className="text-red-500" />
          </span>
        )}
        {pr.isCurrent && (
          <span className="text-xs bg-yellow-100 text-yellow-700 px-1.5 py-0.5 rounded" title="Current branch">
            current
          </span>
        )}
        {pr.isDraft && (
          <span className="text-xs bg-gray-100 text-gray-500 px-1.5 py-0.5 rounded">draft</span>
        )}
      </div>

      <div className="flex items-center gap-2 mt-1 text-xs text-gray-500">
        <span className="text-cyan-600">{pr.source}</span>
        <span>→</span>
        <span className="text-cyan-600">{pr.target}</span>

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

        <span className="ml-auto inline-flex items-center gap-1.5 text-gray-400">
          {syncStatus && <SyncIndicator status={syncStatus} />}
          {pr.author && (
            <Avatar
              src={pr.authorAvatarUrl}
              alt={pr.author}
              size={16}
              href={`https://github.com/${pr.author}`}
              title={`@${pr.author}`}
            />
          )}
          <span title={pr.updatedAt}>{timeAgo(pr.updatedAt)}</span>
        </span>
      </div>
    </div>
  );
}
