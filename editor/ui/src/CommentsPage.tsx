import { useMemo, useRef } from "react";
import {
  buildThreadListHandlers,
  CommentProvider,
  CommentThreadList,
  DEFAULT_COMMENT_STATUSES,
  type CommentConfig,
} from "@flanksource/clicky-ui/comments";
import { DetailEmptyState, FilterBar, Section, type FilterBarFilter } from "@flanksource/clicky-ui/components";
import { Badge, formatRelativeTime } from "@flanksource/clicky-ui/data";
import { UiChatDots } from "@flanksource/clicky-ui/icons";
import { threadCallbacks, type ThreadCallbacks } from "./comment-actions";
import {
  authorOptions,
  filterThreads,
  groupThreads,
  PHASE_OPTIONS,
  STATUS_OPTIONS,
  threadEntries,
  type CommentProject,
  type ThreadEntry,
  type ThreadGroup,
  type ThreadLocation,
} from "./comments-page";
import { threadItems } from "./file-comments";
import { useProjectThreads, type ProjectListing } from "./project-comments";
import type { CommentsFilters } from "./routes";

const CONFIG: CommentConfig = { statuses: DEFAULT_COMMENT_STATUSES };
const NOT_OPEN = "Not open in this session: run gavel edit on this file to open it";

/** An edited file a thread is on, at the thread's line. */
export type OpenThread = Extract<ThreadLocation, { kind: "file" }> & { fileId: string };

type ViewProps = {
  listings: ProjectListing[] | undefined;
  error: string | undefined;
  filters: CommentsFilters;
  onFiltersChange: (filters: CommentsFilters) => void;
  callbacks: ThreadCallbacks;
  onOpen: (thread: OpenThread) => void;
};

function filterControls(filters: CommentsFilters, set: (patch: Partial<CommentsFilters>) => void, entries: readonly ThreadEntry[], projects: readonly CommentProject[]): FilterBarFilter[] {
  const choice = (key: "status" | "phase" | "author" | "project", label: string, options: { value: string; label: string }[]): FilterBarFilter => ({
    key,
    kind: "enum",
    label,
    value: filters[key],
    options,
    onChange: (value) => set({ [key]: value }),
  });
  return [
    choice("status", "Status", STATUS_OPTIONS),
    choice("phase", "Phase", PHASE_OPTIONS),
    choice("author", "Author", authorOptions(entries)),
    choice("project", "Project", projects.map(({ project }) => ({ value: project, label: project }))),
    { key: "unread", kind: "boolean", label: "Unread only", value: filters.unread, onChange: (unread) => set({ unread }) },
  ];
}

/** The thread's place: a link to the edited file at its line, or its path with why it cannot open. */
function ThreadPlace({ entry, onOpen }: { entry: ThreadEntry; onOpen: (thread: OpenThread) => void }) {
  const { location, row } = entry;
  const where = location.kind === "file" && location.line ? `${location.label}:${location.line}` : location.label;
  return (
    <div className="mb-1 flex min-w-0 items-center gap-1.5 text-xs text-muted-foreground">
      {row.unread && (
        <Badge tone="info" variant="soft" size="xs" clickToCopy={false}>
          Unread
        </Badge>
      )}
      {location.kind === "file" && location.fileId ? (
        <button type="button" className="truncate font-mono text-primary hover:underline" title={`Open ${where}`} onClick={() => onOpen({ ...location, fileId: location.fileId! })}>
          {where}
        </button>
      ) : (
        location.kind === "file" && (
          <span className="truncate font-mono" title={NOT_OPEN}>
            {where}
          </span>
        )
      )}
      <span className="ml-auto shrink-0">{formatRelativeTime(new Date(entry.activity))}</span>
    </div>
  );
}

function GroupSection({ group, callbacks, onOpen }: { group: ThreadGroup; callbacks: ThreadCallbacks; onOpen: (thread: OpenThread) => void }) {
  const items = group.threads.flatMap(({ row }) => threadItems(row));
  const byId = new Map(group.threads.map((entry) => [entry.row.id, entry]));
  const title = group.location.kind === "file" && !group.location.fileId ? NOT_OPEN : group.location.label;
  return (
    <Section
      defaultOpen
      icon={UiChatDots}
      title={<span title={title} className="font-mono">{group.location.label}</span>}
      summary={`${group.project.project} · ${group.threads.length} ${group.threads.length === 1 ? "thread" : "threads"}`}
    >
      <CommentThreadList
        comments={items}
        config={CONFIG}
        {...buildThreadListHandlers(items, CONFIG, callbacks)}
        renderRootMeta={(comment) => {
          const entry = byId.get(comment.id);
          return entry ? <ThreadPlace entry={entry} onOpen={onOpen} /> : null;
        }}
      />
    </Section>
  );
}

function ThreadGroups({ groups, callbacks, onOpen }: { groups: ThreadGroup[]; callbacks: ThreadCallbacks; onOpen: (thread: OpenThread) => void }) {
  const items = useMemo(() => groups.flatMap((group) => group.threads.flatMap(({ row }) => threadItems(row))), [groups]);
  return (
    <CommentProvider comments={items} config={CONFIG} {...callbacks}>
      <div className="flex flex-col gap-3">
        {groups.map((group) => (
          <GroupSection key={group.key} group={group} callbacks={callbacks} onOpen={onOpen} />
        ))}
      </div>
    </CommentProvider>
  );
}

/**
 * Every comment thread in the edited files' projects, grouped by target and
 * newest activity first, behind route-level filters that stay visible with no rows.
 */
export function CommentsView({ listings, error, filters, onFiltersChange, callbacks, onOpen }: ViewProps) {
  const entries = useMemo(() => threadEntries(listings ?? []), [listings]);
  const groups = useMemo(() => groupThreads(filterThreads(entries, filters)), [entries, filters]);
  const projects = (listings ?? []).map(({ project }) => project);
  // FilterBar's "Clear all" clears each filter in turn within one event, so patches merge onto the latest filters.
  const latest = useRef(filters);
  latest.current = filters;
  const set = (patch: Partial<CommentsFilters>) => {
    latest.current = { ...latest.current, ...patch };
    onFiltersChange(latest.current);
  };
  const body = () => {
    if (!listings && error) return <p role="alert" className="text-sm text-destructive">Comments could not be loaded: {error}</p>;
    if (!listings) return <p role="status" className="text-sm text-muted-foreground">Loading comments…</p>;
    if (projects.length === 0)
      return <DetailEmptyState icon={UiChatDots} label="No project holds the edited files" description="Create a gavel project for a file to comment on it." />;
    if (entries.length === 0)
      return <DetailEmptyState icon={UiChatDots} label="No comment threads yet" description={`Nothing has been commented on in ${projects.map(({ project }) => project).join(", ")}.`} />;
    if (groups.length === 0) return <DetailEmptyState icon={UiChatDots} label="No threads match these filters" />;
    return <ThreadGroups groups={groups} callbacks={callbacks} onOpen={onOpen} />;
  };
  return (
    <div className="flex min-h-0 flex-1 flex-col" data-testid="comments-page">
      <FilterBar
        className="shrink-0 border-b border-border px-3"
        search={{ value: filters.q, onChange: (q) => set({ q }), placeholder: "Comment text or path…", ariaLabel: "Search comments" }}
        filters={filterControls(filters, set, entries, projects)}
      />
      {listings && error && (
        <p role="alert" className="px-3 py-1 text-sm text-destructive">
          Refreshing comments failed: {error}
        </p>
      )}
      <div className="min-h-0 flex-1 overflow-auto p-4">{body()}</div>
    </div>
  );
}

/** The Comments page: polls every project's threads and acts on them through the comments API. */
export function CommentsPage({
  projects,
  projectsError,
  ...props
}: Omit<ViewProps, "listings" | "error" | "callbacks"> & { projects: CommentProject[] | undefined; projectsError: string | undefined }) {
  const { listings, error, refresh } = useProjectThreads(projects);
  const rows = useMemo(() => (listings ?? []).flatMap((listing) => listing.rows), [listings]);
  return <CommentsView listings={listings} error={projectsError ?? error} callbacks={threadCallbacks(rows, refresh)} {...props} />;
}
