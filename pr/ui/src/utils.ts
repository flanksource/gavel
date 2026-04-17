import type { PRItem, CheckSummary } from './types';

// prKey mirrors the Go-side poller.prKey() so the unread map keys match.
export function prKey(pr: { repo: string; number: number }): string {
  return `${pr.repo}#${pr.number}`;
}

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
  repoOwner: string;
  repoShort: string;
  repoAvatarUrl?: string;
  repoHomepageUrl?: string;
  items: PRItem[];
}

export interface OrgGroup {
  org: string;
  orgAvatarUrl?: string;
  repos: RepoGroup[];
  itemCount: number;
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
  return order.map(repo => {
    const items = groups.get(repo)!;
    const slash = repo.indexOf('/');
    return {
      repo,
      repoOwner: slash >= 0 ? repo.slice(0, slash) : '',
      repoShort: slash >= 0 ? repo.slice(slash + 1) : repo,
      repoAvatarUrl: items.find(p => p.repoAvatarUrl)?.repoAvatarUrl,
      repoHomepageUrl: items.find(p => p.repoHomepageUrl)?.repoHomepageUrl,
      items,
    };
  });
}

export function groupByOrg(prs: PRItem[]): OrgGroup[] {
  const repoGroups = groupByRepo(prs);
  const order: string[] = [];
  const byOrg = new Map<string, RepoGroup[]>();
  for (const rg of repoGroups) {
    const org = rg.repoOwner || '';
    if (!byOrg.has(org)) {
      order.push(org);
      byOrg.set(org, []);
    }
    byOrg.get(org)!.push(rg);
  }
  return order.map(org => {
    const repos = byOrg.get(org)!;
    return {
      org,
      orgAvatarUrl: repos.find(r => r.repoAvatarUrl)?.repoAvatarUrl,
      repos,
      itemCount: repos.reduce((n, r) => n + r.items.length, 0),
    };
  });
}

// AVATAR_PALETTE: 16 Tailwind bg/text pairs chosen for WCAG AA contrast
// at small avatar sizes (20-28px). Add/remove entries carefully — the
// index is derived by hash and must stay stable.
export const AVATAR_PALETTE: readonly string[] = [
  'bg-rose-100 text-rose-700',
  'bg-pink-100 text-pink-700',
  'bg-fuchsia-100 text-fuchsia-700',
  'bg-purple-100 text-purple-700',
  'bg-violet-100 text-violet-700',
  'bg-indigo-100 text-indigo-700',
  'bg-blue-100 text-blue-700',
  'bg-sky-100 text-sky-700',
  'bg-cyan-100 text-cyan-700',
  'bg-teal-100 text-teal-700',
  'bg-emerald-100 text-emerald-700',
  'bg-green-100 text-green-700',
  'bg-lime-100 text-lime-800',
  'bg-amber-100 text-amber-800',
  'bg-orange-100 text-orange-700',
  'bg-red-100 text-red-700',
];

// fnv1a32 is a deterministic 32-bit hash. Used to pick a palette color for
// a repo name so the same name renders in the same color every session.
export function fnv1a32(s: string): number {
  let h = 0x811c9dc5;
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i);
    h = Math.imul(h, 0x01000193);
  }
  return h >>> 0;
}

export function paletteClass(key: string): string {
  return AVATAR_PALETTE[fnv1a32(key) % AVATAR_PALETTE.length];
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
