import type { PRItem, CheckSummary, Project, ProcStatus, ProcProcess } from './types';

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

// Synthetic author-filter keys. @me collapses the viewer's own PRs and @bots
// collapses every bot account into a single chip so they're controlled as one,
// rather than listing each bot login individually.
export const ME_KEY = '@me';
export const AUTOMATED_AUTHOR_KEY = '@bots';

// isBotAuthor mirrors cmd/gavel.isBot — GitHub bot accounts end in "[bot]" and
// some legacy integrations just suffix "bot".
export function isBotAuthor(author: string): boolean {
  return author.endsWith('[bot]') || author.endsWith('bot');
}

// isBotPR reports whether a PR is bot-authored, trusting the server's App flag
// (covers App bots like renovate whose login lacks a bot suffix) and falling
// back to the login heuristic for user-account bots.
export function isBotPR(pr: PRItem): boolean {
  return !!pr.authorIsApp || isBotAuthor(pr.author);
}

// authorCategories returns the author-filter keys a PR belongs to: bots map to
// @bots, the viewer's own PRs to @me, everyone else to their login.
export function authorCategories(pr: PRItem, viewer: string): string[] {
  if (isBotPR(pr)) return [AUTOMATED_AUTHOR_KEY];
  if (viewer && pr.author === viewer) return [ME_KEY];
  return pr.author ? [pr.author] : [];
}

// collectAuthors builds the author-filter option keys: @me and @bots come first,
// then the remaining human authors sorted. @bots stays available when the server
// reports known bots (botsAvailable) even though they're excluded from the fetch.
export function collectAuthors(prs: PRItem[], viewer: string, botsAvailable: boolean): string[] {
  const humans = new Set<string>();
  let hasMe = false;
  let hasBots = botsAvailable;
  for (const pr of prs) {
    if (!pr.author) continue;
    if (isBotPR(pr)) { hasBots = true; continue; }
    if (viewer && pr.author === viewer) { hasMe = true; continue; }
    humans.add(pr.author);
  }
  const out: string[] = [];
  if (hasMe) out.push(ME_KEY);
  if (hasBots) out.push(AUTOMATED_AUTHOR_KEY);
  out.push(...[...humans].sort());
  return out;
}

// FacetModes maps an option key to whether it is included or excluded. Keys
// absent from the map are neutral. Mirrors clicky-ui's FilterBar `kind:"multi"`
// value shape so the tri-state chips bind directly to it.
export type FacetModes = Record<string, 'include' | 'exclude'>;

// passFacet applies tri-state include/exclude semantics to the category set a PR
// belongs to for one facet: any matching exclude rejects; if any include is set,
// at least one must match.
function passFacet(modes: FacetModes, cats: string[]): boolean {
  let hasInclude = false;
  let included = false;
  for (const key in modes) {
    if (modes[key] === 'exclude') {
      if (cats.includes(key)) return false;
    } else {
      hasInclude = true;
      if (cats.includes(key)) included = true;
    }
  }
  return !hasInclude || included;
}

function stateCategories(pr: PRItem): string[] {
  const c: string[] = [];
  if (pr.state === 'OPEN' && !pr.isDraft) c.push('open');
  if (pr.state === 'MERGED') c.push('merged');
  if (pr.state === 'CLOSED') c.push('closed');
  if (pr.isDraft) c.push('draft');
  return c;
}

function checkCategories(pr: PRItem): string[] {
  const cs = pr.checkStatus;
  if (!cs) return [];
  const c: string[] = [];
  if (cs.failed > 0) c.push('failing');
  if (cs.failed === 0 && cs.running === 0) c.push('passing');
  if (cs.running > 0) c.push('running');
  return c;
}

export function filterPRs(
  prs: PRItem[],
  stateFilter: FacetModes,
  checksFilter: FacetModes,
  repoFilter: FacetModes,
  authorFilter: FacetModes,
  viewer: string,
): PRItem[] {
  return prs.filter(pr =>
    passFacet(stateFilter, stateCategories(pr)) &&
    passFacet(checksFilter, checkCategories(pr)) &&
    passFacet(repoFilter, [pr.repo]) &&
    passFacet(authorFilter, authorCategories(pr, viewer)),
  );
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

// FlatProc is one supervised process tagged with its owning project, the row
// shape the process-manager table renders.
export interface FlatProc {
  project: Project;
  proc: ProcProcess;
}

// emptyProcStatus is the placeholder status for a configured project that has no
// live entry in the proc-status map yet (or no Procfile at all). It lets the
// workspace surfaces list every project from config without a type guard.
export const emptyProcStatus: ProcStatus = { hasProcfile: false, running: false };

// flattenProcesses gathers every supervised process across projects that have a
// Procfile, tagged with its project. It iterates projects (not the procStatus
// map, which is keyed by both project name and repo) so each process appears
// once.
export function flattenProcesses(projects: Project[], procStatus: Record<string, ProcStatus>): FlatProc[] {
  const out: FlatProc[] = [];
  for (const project of projects) {
    const st = procStatus[project.name];
    if (!st?.hasProcfile || !st.processes) continue;
    for (const proc of st.processes) {
      out.push({ project, proc });
    }
  }
  return out;
}

// statusDotClass maps a single process status to its dot color, shared by the
// table, the inline control, and the manager badge.
export function statusDotClass(status: string): string {
  switch (status) {
    case 'running': return 'bg-green-500';
    case 'starting':
    case 'restarting': return 'bg-yellow-400';
    case 'crashed': return 'bg-red-500';
    default: return 'bg-gray-300';
  }
}

// aggregateDotClass summarises a group of processes into one dot color: any
// crash is red, an all-running group is green, any (re)starting or partly
// running group is yellow, otherwise gray.
export function aggregateDotClass(procs: ProcProcess[]): string {
  if (procs.some(p => p.status === 'crashed')) return 'bg-red-500';
  const running = procs.filter(p => p.status === 'running').length;
  if (procs.length > 0 && running === procs.length) return 'bg-green-500';
  if (procs.some(p => p.status === 'starting' || p.status === 'restarting') || running > 0) return 'bg-yellow-400';
  return 'bg-gray-300';
}

// aggregateResources sums a group's live CPU and RSS into one workspace-level
// reading. Each process's cpuPercent/memoryRss is already its own process-group
// sum (the supervisor aggregates the tree), so summing across processes gives
// the workspace total. Missing samples (stopped processes) count as zero.
export function aggregateResources(procs: ProcProcess[]): { cpuPercent: number; memoryRss: number } {
  let cpuPercent = 0;
  let memoryRss = 0;
  for (const p of procs) {
    cpuPercent += p.cpuPercent ?? 0;
    memoryRss += p.memoryRss ?? 0;
  }
  return { cpuPercent, memoryRss };
}

// statusLabel is the human status, annotating a crash with its exit code so a
// crashed process reads "crashed (exit 3)" rather than looking like a stop.
export function statusLabel(proc: ProcProcess): string {
  if (proc.status === 'crashed' && proc.exitCode !== undefined) {
    return `${proc.status} (exit ${proc.exitCode})`;
  }
  return proc.status;
}

// crashedSummary describes the crashed processes in a group for a tooltip, e.g.
// "1 crashed (exit 3)" or "2 crashed (exit 1, exit 137)". Empty when none.
export function crashedSummary(procs: ProcProcess[]): string {
  const crashed = procs.filter(p => p.status === 'crashed');
  if (crashed.length === 0) return '';
  const codes = crashed
    .filter(p => p.exitCode !== undefined)
    .map(p => `exit ${p.exitCode}`);
  return `${crashed.length} crashed${codes.length ? ` (${codes.join(', ')})` : ''}`;
}

// humanizeBytes renders a byte count compactly (e.g. "1.5 GB"). Non-positive
// input renders as an em dash.
export function humanizeBytes(n?: number): string {
  if (!n || n <= 0) return '—';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let v = n;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return `${i > 0 && v < 10 ? v.toFixed(1) : Math.round(v)} ${units[i]}`;
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
