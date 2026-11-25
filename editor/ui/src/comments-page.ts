import type { NativeComment } from "./file-comments";
import type { SelectorView } from "./comment-anchors";
import type { CommentsFilters } from "./routes";

/** GET /api/comments/projects: a project holding edited files, and its edited files by target URI. */
export type CommentProject = { scope: string; project: string; dir: string; files: { id: string; uri: string }[] };

/** A row of GET /api/v1/comments: the thread, its workflow phase and whether the reader has seen its latest version. */
export type ListedComment = Omit<NativeComment, "target"> & {
  phase: string;
  updatedAt: string;
  unread: boolean;
  target: NativeComment["target"] & { uri: string; scopeUri: string; kind: string; label?: string };
};

/**
 * Where a thread lives: a project file (its path relative to the project, and
 * the edited file it is when one is open in this session) or another target,
 * such as a web page, named by its URI.
 */
export type ThreadLocation =
  | { kind: "file"; key: string; label: string; fileId?: string; line?: number; view?: SelectorView }
  | { kind: "page"; key: string; label: string };

export type ThreadEntry = { row: ListedComment; project: CommentProject; location: ThreadLocation; activity: string };

export type ThreadGroup = { key: string; project: CommentProject; location: ThreadLocation; activity: string; threads: ThreadEntry[] };

export function locationOf(row: ListedComment, project: CommentProject): ThreadLocation {
  const uri = row.target.uri;
  const files = `${project.scope}/files/`;
  const pages = `${project.scope}/pages/`;
  if (uri.startsWith(pages)) return { kind: "page", key: uri, label: decodeURIComponent(uri.slice(pages.length)) };
  if (!uri.startsWith(files)) return { kind: "page", key: uri, label: uri };
  const fileId = project.files.find((file) => file.uri === uri)?.id;
  const selector = row.target.selector;
  const line = selector.kind === "code-range" ? selector.patch?.startLine : undefined;
  return {
    kind: "file",
    key: uri,
    label: uri.slice(files.length),
    ...(fileId ? { fileId } : {}),
    ...(line ? { line } : {}),
    ...(selector.view ? { view: selector.view } : {}),
  };
}

/** A thread's latest activity: its newest reply, its last update, or its creation. */
export function activityOf(row: ListedComment): string {
  const times = [row.createdAt, row.updatedAt, ...(row.messages ?? []).map((message) => message.createdAt)];
  return times.reduce((latest, at) => (Date.parse(at) > Date.parse(latest) ? at : latest));
}

export function threadEntries(listings: readonly { project: CommentProject; rows: readonly ListedComment[] }[]): ThreadEntry[] {
  return listings.flatMap(({ project, rows }) =>
    rows.map((row) => ({ row, project, location: locationOf(row, project), activity: activityOf(row) })),
  );
}

const authorLabel = (author: { id: string; label?: string }) => author.label ?? author.id;

function matchesSearch(entry: ThreadEntry, query: string): boolean {
  const needle = query.trim().toLowerCase();
  if (!needle) return true;
  const { row, location } = entry;
  return [row.bodyMarkdown, location.label, ...(row.messages ?? []).map((message) => message.bodyMarkdown)].some((text) =>
    text.toLowerCase().includes(needle),
  );
}

export function filterThreads(entries: readonly ThreadEntry[], filters: CommentsFilters): ThreadEntry[] {
  return entries.filter(
    (entry) =>
      (!filters.status || entry.row.status === filters.status) &&
      (!filters.phase || entry.row.phase === filters.phase) &&
      (!filters.author || entry.row.author.id === filters.author) &&
      (!filters.project || entry.project.project === filters.project) &&
      (!filters.unread || entry.row.unread) &&
      matchesSearch(entry, filters.q),
  );
}

const newestFirst = (a: { activity: string }, b: { activity: string }) => Date.parse(b.activity) - Date.parse(a.activity);

/** Threads grouped by target, each group's threads and the groups themselves newest activity first. */
export function groupThreads(entries: readonly ThreadEntry[]): ThreadGroup[] {
  const groups = new Map<string, ThreadGroup>();
  for (const entry of [...entries].sort(newestFirst)) {
    const key = `${entry.project.scope}\n${entry.location.key}`;
    const group = groups.get(key);
    if (group) group.threads.push(entry);
    else groups.set(key, { key, project: entry.project, location: entry.location, activity: entry.activity, threads: [entry] });
  }
  return [...groups.values()];
}

/** The authors present in `entries`, as filter options ordered by label. */
export function authorOptions(entries: readonly ThreadEntry[]): { value: string; label: string }[] {
  const byId = new Map(entries.map(({ row }) => [row.author.id, authorLabel(row.author)]));
  return [...byId].map(([value, label]) => ({ value, label })).sort((a, b) => a.label.localeCompare(b.label));
}

export const STATUS_OPTIONS = [
  { value: "open", label: "Open" },
  { value: "in_progress", label: "In progress" },
  { value: "resolved", label: "Resolved" },
  { value: "closed", label: "Closed" },
];

export const PHASE_OPTIONS = [
  { value: "active", label: "Active" },
  { value: "awaiting_outcome", label: "Awaiting outcome" },
  { value: "final", label: "Final" },
];
