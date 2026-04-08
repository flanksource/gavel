import type { PRItem, CheckSummary } from './types';

export function stateIcon(state: string, isDraft: boolean): string {
  if (isDraft) return '○';
  switch (state) {
    case 'OPEN': return '●';
    case 'MERGED': return '●';
    case 'CLOSED': return '●';
    default: return '?';
  }
}

export function stateColor(state: string, isDraft: boolean): string {
  if (isDraft) return 'text-gray-400';
  switch (state) {
    case 'OPEN': return 'text-green-600';
    case 'MERGED': return 'text-purple-600';
    case 'CLOSED': return 'text-red-600';
    default: return 'text-gray-400';
  }
}

export function reviewColor(decision: string): string {
  switch (decision) {
    case 'APPROVED': return 'text-green-600';
    case 'CHANGES_REQUESTED': return 'text-red-600';
    case 'REVIEW_REQUIRED': return 'text-yellow-600';
    default: return 'text-gray-500';
  }
}

export function checkSummaryText(cs: CheckSummary): string {
  const parts: string[] = [];
  if (cs.passed > 0) parts.push(`✓${cs.passed}`);
  if (cs.failed > 0) parts.push(`✗${cs.failed}`);
  if (cs.running > 0) parts.push(`●${cs.running}`);
  if (cs.pending > 0) parts.push(`○${cs.pending}`);
  return parts.join(' ');
}

export function timeAgo(iso: string): string {
  const d = new Date(iso);
  const s = Math.floor((Date.now() - d.getTime()) / 1000);
  if (s < 60) return 'just now';
  if (s < 3600) return `${Math.floor(s / 60)}m ago`;
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`;
  return `${Math.floor(s / 86400)}d ago`;
}

export interface RepoGroup {
  repo: string;
  repoShort: string;
  items: PRItem[];
}

export function groupByRepo(prs: PRItem[]): RepoGroup[] {
  const sorted = [...prs].sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime());
  const order: string[] = [];
  const groups = new Map<string, PRItem[]>();
  for (const pr of sorted) {
    if (!groups.has(pr.repo)) {
      order.push(pr.repo);
      groups.set(pr.repo, []);
    }
    groups.get(pr.repo)!.push(pr);
  }
  return order.map(repo => ({
    repo,
    repoShort: repo.includes('/') ? repo.split('/')[1] : repo,
    items: groups.get(repo)!,
  }));
}

export interface PRCounts {
  open: number;
  merged: number;
  closed: number;
  draft: number;
  failing: number;
  passing: number;
  running: number;
  noChecks: number;
}

export function computeCounts(prs: PRItem[]): PRCounts {
  const c: PRCounts = { open: 0, merged: 0, closed: 0, draft: 0, failing: 0, passing: 0, running: 0, noChecks: 0 };
  for (const pr of prs) {
    if (pr.isDraft) c.draft++;
    switch (pr.state) {
      case 'OPEN': c.open++; break;
      case 'MERGED': c.merged++; break;
      case 'CLOSED': c.closed++; break;
    }
    if (pr.checkStatus) {
      if (pr.checkStatus.failed > 0) c.failing++;
      else if (pr.checkStatus.running > 0) c.running++;
      else c.passing++;
    } else {
      c.noChecks++;
    }
  }
  return c;
}

export function collectRepos(prs: PRItem[]): string[] {
  const seen = new Set<string>();
  for (const pr of prs) {
    seen.add(pr.repo);
  }
  return [...seen].sort();
}

export function collectAuthors(prs: PRItem[]): string[] {
  const seen = new Set<string>();
  for (const pr of prs) {
    if (pr.author) seen.add(pr.author);
  }
  return [...seen].sort();
}

export function filterPRs(
  prs: PRItem[],
  stateFilter: Set<string>,
  checksFilter: Set<string>,
  repoFilter: Set<string>,
  authorFilter: Set<string>,
): PRItem[] {
  return prs.filter(pr => {
    if (stateFilter.size > 0) {
      const matches =
        (stateFilter.has('open') && pr.state === 'OPEN' && !pr.isDraft) ||
        (stateFilter.has('merged') && pr.state === 'MERGED') ||
        (stateFilter.has('closed') && pr.state === 'CLOSED') ||
        (stateFilter.has('draft') && pr.isDraft);
      if (!matches) return false;
    }
    if (checksFilter.size > 0) {
      const cs = pr.checkStatus;
      const matches =
        (checksFilter.has('failing') && cs && cs.failed > 0) ||
        (checksFilter.has('passing') && cs && cs.failed === 0 && cs.running === 0) ||
        (checksFilter.has('running') && cs && cs.running > 0);
      if (!matches) return false;
    }
    if (repoFilter.size > 0 && !repoFilter.has(pr.repo)) return false;
    if (authorFilter.size > 0 && !authorFilter.has(pr.author)) return false;
    return true;
  });
}

export function statusIcon(status: string, conclusion: string): string {
  if (status === 'completed' || status === 'COMPLETED') {
    switch (conclusion.toUpperCase()) {
      case 'SUCCESS': case 'NEUTRAL': case 'SKIPPED': return '✓';
      case 'FAILURE': case 'TIMED_OUT': case 'STARTUP_FAILURE': return '✗';
      case 'CANCELLED': return '⊘';
      default: return '?';
    }
  }
  if (status === 'in_progress' || status === 'IN_PROGRESS') return '●';
  return '○';
}

export function statusColor(status: string, conclusion: string): string {
  if (status === 'completed' || status === 'COMPLETED') {
    switch (conclusion.toUpperCase()) {
      case 'SUCCESS': case 'NEUTRAL': case 'SKIPPED': return 'text-green-600';
      case 'FAILURE': case 'TIMED_OUT': case 'STARTUP_FAILURE': return 'text-red-600';
      case 'CANCELLED': return 'text-gray-500';
      default: return 'text-gray-500';
    }
  }
  if (status === 'in_progress' || status === 'IN_PROGRESS') return 'text-yellow-600';
  return 'text-gray-400';
}

export function severityIcon(severity?: string): string {
  switch (severity) {
    case 'critical': return '🔴';
    case 'major': return '🟠';
    case 'minor': return '🟡';
    case 'nitpick': return '🧹';
    default: return '💬';
  }
}
