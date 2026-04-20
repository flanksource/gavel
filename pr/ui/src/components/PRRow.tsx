import type { PRItem, PRSyncStatus, GavelResultsSummary } from '../types';
import { reviewColor, checkSummaryText, timeAgo } from '../utils';
import { SyncIndicator } from './SyncIndicator';

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
      <span class="inline-flex items-center text-yellow-600" title={`gavel: ${g.error}`}>
        <iconify-icon icon="codicon:warning" />
      </span>
    );
  }
  const items: { icon: string; color: string; count: number; title: string }[] = [];
  if (g.testsTotal > 0) {
    const failed = g.testsFailed;
    items.push({
      icon: failed > 0 ? 'codicon:error' : 'codicon:pass',
      color: failed > 0 ? 'text-red-600' : 'text-green-600',
      count: failed > 0 ? failed : g.testsPassed,
      title: failed > 0
        ? `${failed} test${failed !== 1 ? 's' : ''} failed`
        : `${g.testsPassed} test${g.testsPassed !== 1 ? 's' : ''} passed`,
    });
  }
  if (g.lintLinters > 0) {
    items.push({
      icon: g.lintViolations > 0 ? 'codicon:warning' : 'codicon:check',
      color: g.lintViolations > 0 ? 'text-yellow-600' : 'text-green-600',
      count: g.lintViolations,
      title: g.lintViolations > 0
        ? `${g.lintViolations} lint violation${g.lintViolations !== 1 ? 's' : ''}`
        : `${g.lintLinters} linter${g.lintLinters !== 1 ? 's' : ''} clean`,
    });
  }
  if (g.hasBench) {
    const regs = g.benchRegressions ?? 0;
    items.push({
      icon: regs > 0 ? 'codicon:arrow-down' : 'codicon:graph',
      color: regs > 0 ? 'text-red-600' : 'text-blue-600',
      count: regs,
      title: regs > 0 ? `${regs} bench regression${regs !== 1 ? 's' : ''}` : 'no bench regressions',
    });
  }
  if (items.length === 0) return null;
  return (
    <span class="inline-flex items-center gap-1.5" aria-label="gavel results">
      {items.map((it, i) => (
        <span key={i} class={`inline-flex items-center gap-0.5 ${it.color} tabular-nums`} title={it.title}>
          <iconify-icon icon={it.icon} />
          {it.count > 0 && <span class="text-[11px] font-medium">{it.count}</span>}
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
      class={`px-3 py-2 cursor-pointer border-l-2 transition-colors ${borderColor(pr, selected, gavelResults)} ${
        selected ? 'bg-blue-50' : unread ? 'hover:bg-gray-50' : 'hover:bg-gray-50'
      }`}
      onClick={onClick}
    >
      <div class="flex items-center gap-2">
        <span
          class={`inline-block w-1.5 h-1.5 rounded-full shrink-0 ${unread ? 'bg-blue-600' : 'bg-transparent'}`}
          title={unread ? 'Unread — updated since last view' : ''}
          aria-label={unread ? 'unread' : ''}
        />
        <iconify-icon icon={status.icon} class={`text-sm ${status.color} shrink-0`} title={status.title} />
        <span class="text-xs text-gray-400">#{pr.number}</span>
        <span class={`text-sm truncate flex-1 ${unread ? 'font-semibold text-gray-900' : 'font-medium text-gray-800'}`}>{pr.title}</span>
        {hasConflict && (
          <span class="text-xs text-red-500" title="Merge conflicts">
            <iconify-icon icon="octicon:git-merge-16" class="text-red-500" />
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

        {gavelResults && <GavelBadges g={gavelResults} />}

        <span class="ml-auto inline-flex items-center gap-1.5 text-gray-400">
          {syncStatus && <SyncIndicator status={syncStatus} />}
          <span title={pr.updatedAt}>{timeAgo(pr.updatedAt)}</span>
        </span>
      </div>
    </div>
  );
}
