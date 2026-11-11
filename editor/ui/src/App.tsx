import { Fragment, useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  AppShell,
  Button,
  CommandPalette,
  DropdownMenu,
  SegmentedControl,
  type DropdownMenuItem,
  type JsonSchemaObject,
} from "@flanksource/clicky-ui/components";
import { UiChatDots, UiDotsVertical, UiFolder, UiListChecks } from "@flanksource/clicky-ui/icons";
import gavelLockup from "../../../assets/brand/gavel-logo-mono.svg";
import {
  DocumentForm,
  DocumentRichEditor,
  documentParts,
} from "@flanksource/clicky-ui/monaco";
import {
  FileFindingsView,
  emptyProblemFilters,
  fileRefKey,
  groupFindingsByFile,
  isImagePath,
  lintFindings,
  relativeUrlPath,
  testFindings,
  type FileFinding,
  type FileRef,
  type FileReviewTab,
  type FileReviewView,
  type FileTreeSelection,
  type ProblemFilters,
} from "@flanksource/clicky-ui/file-findings";
import {
  TaskManager,
  TaskManagerButton,
  violationIgnoreActions,
  type LintAction,
  type LinterResult,
  type Test,
} from "@flanksource/clicky-ui/data";
import { useEventSourceFactory } from "@flanksource/clicky-ui/hooks";
import { createAutosave, type Autosave } from "./autosave";
import { request, type FileEntry, type Document } from "./api";
import { commonRoot, editorFindingRef, nodePath, relativeToFile } from "./paths";
import { taskRunFromHref } from "./routes";
import { commentCountsAt, useCommentSummary } from "./comment-summary";
import { RecentComments } from "./RecentComments";
import { useFileComments } from "./file-comments";
import { CreateProjectDialog } from "./CreateProjectDialog";
import { CurrentUser } from "./CurrentUser";
import { NewAgentSessionDialog } from "./NewAgentSessionDialog";
import { PageAgentControl } from "./PageAgentControl";

// The problems shown for the edited files: recorded lint violations and failed tests
// (GET /api/lint, streamed from /api/lint/stream), and gavel runs in progress.
type ActiveRun = { pid: number; command: string; groups: { name: string; summary: { total: number; passed: number; failed: number; pending?: number; running?: number } }[] };
type Run = {
  running: boolean;
  results: LinterResult[] | null;
  tests: Test[] | null;
  error?: string;
  stale?: boolean;
  staleFiles: string[] | null;
  recorded?: string;
  revisions: Record<string, string> | null;
  active: ActiveRun[] | null;
};

function activeRunLabel(run: ActiveRun): string {
  const progress = run.groups.map(({ name, summary }) => {
    const done = summary.total - (summary.pending ?? 0) - (summary.running ?? 0);
    return `${name} ${done}/${summary.total}`;
  });
  return [`${run.command} running`, ...progress].join(" · ");
}

function recordedLabel(recorded: string): string {
  const minutes = Math.round((Date.now() - new Date(recorded).getTime()) / 60000);
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes}m ago`;
  if (minutes < 60 * 24) return `${Math.round(minutes / 60)}h ago`;
  return new Date(recorded).toLocaleString();
}
// The open node is named by its path in the tree (relative to the edited files' common root).
// `tasks` (present, "" for the list or a run id) shows the full task manager instead.
type Route = { selection: FileTreeSelection | null; view: FileReviewView; line?: number; tasks?: string };

function readRoute(): Route {
  const params = new URLSearchParams(location.search);
  const line = Number(params.get("line"));
  const file = params.get("file");
  const folder = params.get("folder");
  return {
    selection: file ? { kind: "file", ref: { file } } : folder !== null ? { kind: "folder", ref: { file: folder } } : null,
    view: params.get("view") ?? "preview",
    ...(line > 0 ? { line } : {}),
    ...(params.has("tasks") ? { tasks: params.get("tasks")! } : {}),
  };
}

type CommentFilter = "unread" | "all";
// The tree's comment filter lives in the URL beside the route but outlives navigation.
const readCommentFilter = (): CommentFilter => (new URLSearchParams(location.search).get("comments") === "unread" ? "unread" : "all");

function writeCommentFilter(filter: CommentFilter) {
  const url = new URL(location.href);
  if (filter === "unread") url.searchParams.set("comments", "unread");
  else url.searchParams.delete("comments");
  history.replaceState(null, "", url);
}

/** The page URL of a route, keeping the comment filter that outlives navigation. */
function routeHref(route: Route): string {
  const url = new URL(location.href);
  const comments = url.searchParams.get("comments");
  url.search = "";
  if (comments) url.searchParams.set("comments", comments);
  if (route.tasks !== undefined) url.searchParams.set("tasks", route.tasks);
  if (route.selection) url.searchParams.set(route.selection.kind, route.selection.ref.file);
  url.searchParams.set("view", route.view);
  if (route.line) url.searchParams.set("line", String(route.line));
  return `${url.pathname}${url.search}`;
}

const writeRoute = (route: Route) => history.pushState(null, "", routeHref(route));

export function App() {
  const [files, setFiles] = useState<FileEntry[]>([]);
  const [route, setRoute] = useState(readRoute);
  const [filters, setFilters] = useState<ProblemFilters>(emptyProblemFilters);
  const [comparison, setComparison] = useState<string | null>(null);
  const [palette, setPalette] = useState(false);
  const [error, setError] = useState("");
  const [run, setRun] = useState<Run | null>(null);
  const [schema, setSchema] = useState<JsonSchemaObject | null>(null);
  const [schemaError, setSchemaError] = useState("");
  // Remounts the view so a file re-reads its text after the draft was replaced from disk.
  const [generation, setGeneration] = useState(0);
  const queues = useRef(new Map<string, Autosave>());
  const [, render] = useState(0);

  const root = useMemo(() => commonRoot(files.map((file) => file.path)), [files]);
  const refs = useMemo(
    () => files.map((file): FileRef => ({ file: file.path.slice(root.length) })),
    [files, root],
  );
  const byRef = useMemo(
    () => new Map(files.map((file, index) => [fileRefKey(refs[index]!), file])),
    [files, refs],
  );
  const entryOf = useCallback(
    (ref: FileRef) => {
      const entry = byRef.get(fileRefKey(ref));
      if (!entry) throw new Error(`${ref.file} is not one of the edited files`);
      return entry;
    },
    [byRef],
  );
  const relativePath = (entry: FileEntry) => entry.path.slice(root.length);
  const openFile = (entry: FileEntry, view: FileReviewView) =>
    navigate({ selection: { kind: "file", ref: { file: relativePath(entry) } }, view });
  const active = route.selection?.kind === "file" ? byRef.get(fileRefKey(route.selection.ref)) : undefined;
  const activeRef = active ? { file: relativePath(active) } : null;
  const draft = active ? queues.current.get(active.id)?.state() : undefined;
  const { comments, missingProject, createProject, newSession, pageAgent } = useFileComments({
    fileId: active?.kind === "document" ? active.id : null,
    queue: active ? queues.current.get(active.id) : undefined,
    view: route.view,
    onError: setError,
  });
  // A file outside every project prompts once per file; "Not now" leaves it uncommentable.
  const [skippedProject, setSkippedProject] = useState<string>();

  const navigate = (next: Route) => {
    setRoute(next);
    setComparison(null);
    writeRoute(next);
  };

  const { summary, refresh: refreshSummary } = useCommentSummary(setError);
  const [commentFilter, setCommentFilter] = useState<CommentFilter>(readCommentFilter);
  const commentCounts = useMemo(
    () => new Map(files.flatMap((file) => (summary.files[file.id] ? [[relativePath(file), summary.files[file.id]!] as const] : []))),
    [files, root, summary],
  );
  const unreadComments = Object.values(summary.files).reduce((sum, counts) => sum + counts.unread, 0);
  // The open file's threads are in front of the reader, so they are read, including ones they just wrote.
  const activeUnread = active ? (summary.files[active.id]?.unread ?? 0) : 0;
  useEffect(() => {
    if (!active || activeUnread === 0) return;
    void request(`/api/file/${active.id}/comments/read`, {})
      .then(refreshSummary)
      .catch((cause) => setError(String(cause)));
  }, [active?.id, activeUnread]);

  useEffect(() => {
    void request<FileEntry[]>("/api/files")
      .then(setFiles)
      .catch((cause) => setError(String(cause)));
    const back = () => setRoute(readRoute());
    const unload = (event: BeforeUnloadEvent) => {
      if ([...queues.current.values()].some((queue) => queue.state().status !== "saved")) {
        event.preventDefault();
        event.returnValue = "";
      }
    };
    window.addEventListener("popstate", back);
    window.addEventListener("beforeunload", unload);
    return () => {
      window.removeEventListener("popstate", back);
      window.removeEventListener("beforeunload", unload);
      for (const queue of queues.current.values()) queue.dispose();
    };
  }, []);

  const queueFor = (doc: Document) => {
    const queue = createAutosave({
      source: doc.source,
      revision: doc.revision,
      changed: () => render((value) => value + 1),
      write: async (source, revision) =>
        (await request<Document>(`/api/file/${doc.id}`, { source, revision }, "PUT")).revision,
    });
    queues.current.get(doc.id)?.dispose();
    queues.current.set(doc.id, queue);
    render((value) => value + 1);
    return queue;
  };

  const loadFile = useCallback(
    async (ref: FileRef) => {
      const entry = entryOf(ref);
      const queue = queues.current.get(entry.id);
      if (queue) return queue.state().source;
      const doc = await request<Document>(`/api/file/${entry.id}`);
      return queueFor(doc).state().source;
    },
    [entryOf],
  );

  const fileUrl = useCallback(
    (ref: FileRef) => {
      const entry = byRef.get(fileRefKey(ref));
      if (entry) return `/api/file/${entry.id}/image`;
      if (!active || !activeRef)
        throw new Error(`${ref.file} is not an edited file and no document is open`);
      return `/api/file/${active.id}/image?path=${encodeURIComponent(relativeToFile(activeRef.file, ref.file))}`;
    },
    [byRef, active?.id, activeRef?.file],
  );

  useEffect(() => {
    setSchema(null);
    setSchemaError("");
  }, [active?.id]);
  useEffect(() => {
    if (!active || !draft) return;
    let cancelled = false;
    const timer = setTimeout(() => {
      try {
        const parts = documentParts(draft.source, active.path);
        void request<JsonSchemaObject | null>(`/api/schema/${active.id}`, { source: parts.data })
          .then((next) => {
            if (!cancelled) {
              setSchema(next);
              setSchemaError("");
            }
          })
          .catch((cause) => !cancelled && setSchemaError(String(cause)));
      } catch (cause) {
        setSchemaError(String(cause));
      }
    }, 600);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [active?.id, draft?.source, active?.path]);

  const openStream = useEventSourceFactory();
  useEffect(() => {
    const stream = openStream("/api/lint/stream");
    stream.onmessage = (event) => setRun(JSON.parse(event.data) as Run);
    const failed = (event: Event) => {
      const data = (event as MessageEvent).data;
      if (typeof data === "string") setError(`problems stream: ${(JSON.parse(data) as { error: string }).error}`);
    };
    stream.addEventListener("error", failed);
    return () => {
      stream.removeEventListener("error", failed);
      stream.close();
    };
  }, [openStream]);

  const lint = async (targets: FileEntry[]) => {
    try {
      setError("");
      const documents = targets.filter((file) => file.kind === "document");
      await Promise.all(documents.map((file) => queues.current.get(file.id)?.flush()));
      setRun(await request<Run>("/api/lint", { files: documents.map((file) => file.id) }));
    } catch (cause) {
      setError(String(cause));
    }
  };

  const recover = async (entry: FileEntry, replace: boolean) => {
    try {
      setComparison(null);
      const disk = await request<Document>(`/api/file/${entry.id}`);
      const queue = queues.current.get(entry.id);
      const doc =
        replace && queue
          ? await request<Document>(`/api/file/${entry.id}`, { source: queue.state().source, revision: disk.revision }, "PUT")
          : disk;
      queueFor(doc);
      setGeneration((value) => value + 1);
    } catch (cause) {
      setError(String(cause));
    }
  };

  // Each shown finding is rebased onto the edited files' root; `origins` keeps the
  // linter and the work-dir-relative finding an ignore rule is written against.
  const { findings, origins } = useMemo(() => {
    const origins = new Map<FileFinding, { linter: string; finding: FileFinding }>();
    const findings = (run?.results ?? []).flatMap((result) =>
      lintFindings([result]).map((finding) => {
        const shown = { ...finding, ...editorFindingRef(finding, root), workDir: undefined };
        origins.set(shown, { linter: result.linter, finding });
        return shown;
      }),
    );
    const tests = testFindings(run?.tests ?? []).map((finding) => ({ ...finding, ...editorFindingRef(finding, root), workDir: undefined }));
    return { findings: [...findings, ...tests], origins };
  }, [run?.results, run?.tests, root]);
  const ignore = async (action: LintAction) => {
    try {
      setError("");
      await request("/api/lint/ignore", action.req);
      setRun(await request<Run>("/api/lint"));
    } catch (cause) {
      setError(String(cause));
    }
  };
  const findingActions = (shown: FileFinding): DropdownMenuItem[] => {
    if (shown.kind === "test") return [];
    const origin = origins.get(shown);
    if (!origin) throw new Error(`no lint result recorded for the finding at ${shown.file}:${shown.line}`);
    const { linter, finding } = origin;
    const violation = { ...(finding.rule ? { rule: { method: finding.rule } } : {}), ...(finding.line ? { line: finding.line } : {}) };
    return violationIgnoreActions({ violation, linter, file: finding.file, workDir: finding.workDir ?? "" }).map((action) => ({
      label: action.label,
      ...(action.title ? { title: action.title } : {}),
      disabled: Boolean(run?.running) || (Boolean(action.disabledWithoutWorkDir) && !action.req.work_dir),
      onSelect: () => void ignore(action),
    }));
  };
  const tree = useMemo(() => groupFindingsByFile(findings, refs), [findings, refs]);
  // "Unread" narrows the tree to files with unread comments; the open file always stays listed.
  const shownPaths = useMemo(
    () =>
      commentFilter === "all"
        ? null
        : new Set([...commentCounts].filter(([, counts]) => counts.unread > 0).map(([path]) => path).concat(activeRef ? [activeRef.file] : [])),
    [commentFilter, commentCounts, activeRef?.file],
  );
  const shownRefs = useMemo(() => (shownPaths ? refs.filter((ref) => shownPaths.has(ref.file)) : refs), [refs, shownPaths]);
  const shownFindings = useMemo(
    () => (shownPaths ? findings.filter((finding) => shownPaths.has(finding.file)) : findings),
    [findings, shownPaths],
  );
  const commentBadge = (kind: "file" | "folder", path: string) => {
    const counts = commentCountsAt(commentCounts, kind, path);
    if (counts.total === 0) return null;
    const unread = counts.unread > 0;
    const title = unread ? `${counts.unread} unread of ${counts.total} comments` : `${counts.total} comments`;
    return (
      <span
        title={title}
        aria-label={title}
        data-comment-count={unread ? "unread" : "read"}
        className={`inline-flex shrink-0 items-center gap-0.5 rounded-full px-1.5 text-[10px] leading-4 ${unread ? "bg-primary/10 font-semibold text-primary" : "text-muted-foreground"}`}
      >
        <UiChatDots />
        {unread ? counts.unread : counts.total}
      </span>
    );
  };
  const crumbs = route.selection ? nodePath(tree, route.selection) : [];
  const lintItem = (label: string, targets: FileEntry[]): DropdownMenuItem => ({
    label,
    disabled: run?.running || !targets.some((file) => file.kind === "document"),
    onSelect: () => void lint(targets),
  });
  const stale =
    run?.stale ||
    (run?.revisions &&
      Object.entries(run.revisions).some(([id, revision]) => {
        const state = queues.current.get(id)?.state();
        return state && (state.revision !== revision || state.status !== "saved");
      }));

  const fileToolbar = (ref: FileRef) => {
    const entry = entryOf(ref);
    const state = queues.current.get(entry.id)?.state();
    return (
      <div className="flex flex-wrap items-center gap-2 text-sm" role="status">
        {pageAgent && entry.id === active?.id && (
          <PageAgentControl
            fileId={entry.id}
            agent={pageAgent.agent}
            choice={pageAgent.choice}
            sessions={pageAgent.sessions}
            onChoose={pageAgent.choose}
            onSaved={pageAgent.refresh}
          />
        )}
        {state && <span>{state.status === "error" ? state.error : state.status}</span>}
        {state?.status === "error" && (
          <>
            <Button size="sm" variant="outline" onClick={() => void queues.current.get(entry.id)?.retry().catch((cause) => setError(String(cause)))}>
              Retry
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => {
                void request<Document>(`/api/file/${entry.id}`)
                  .then((disk) => {
                    setComparison(disk.source);
                    openFile(entry, "compare");
                  })
                  .catch((cause) => setError(String(cause)));
              }}
            >
              Compare disk
            </Button>
            <Button size="sm" variant="outline" onClick={() => void recover(entry, false)}>
              Reload disk (discard draft)
            </Button>
            <Button size="sm" variant="outline" onClick={() => void recover(entry, true)}>
              Replace disk with draft
            </Button>
          </>
        )}
      </div>
    );
  };

  const fileTabs = (ref: FileRef): FileReviewTab[] => {
    const entry = entryOf(ref);
    if (entry.kind !== "document") return [];
    const markdown = /\.mdx?$/i.test(ref.file);
    const tabs: FileReviewTab[] = [];
    if (markdown)
      tabs.push({
        id: "rich",
        label: "Rich text",
        render: (ctx) => (
          <DocumentRichEditor
            path={ref.file}
            value={ctx.source}
            onChange={(source) => ctx.onSourceChange?.(source)}
            frontmatterSchema={entry.id === active?.id ? schema : null}
            resolveImageUrl={(src) => {
              const path = relativeUrlPath(src);
              return path === null ? src : `/api/file/${entry.id}/image?path=${encodeURIComponent(path)}`;
            }}
          />
        ),
      });
    if (schema && entry.id === active?.id)
      tabs.push({
        id: "form",
        label: markdown ? "Frontmatter" : "Form",
        render: (ctx) => (
          <DocumentForm path={ref.file} value={ctx.source} schema={schema} onChange={(source) => ctx.onSourceChange?.(source)} />
        ),
      });
    if (comparison !== null && entry.id === active?.id)
      tabs.push({
        id: "compare",
        label: "Disk vs draft",
        render: (ctx) => (
          <div className="grid h-full grid-cols-2 gap-2 overflow-auto p-2 text-sm">
            <div>
              <p className="font-semibold">Disk</p>
              <pre className="whitespace-pre-wrap">{comparison}</pre>
            </div>
            <div>
              <p className="font-semibold">Draft</p>
              <pre className="whitespace-pre-wrap">{ctx.source}</pre>
            </div>
          </div>
        ),
      });
    return tabs;
  };

  return (
    <AppShell
      // The mono lockup as a mask takes the rail's foreground colour, which an <img> of it cannot.
      brand={
        <span
          role="img"
          aria-label="gavel"
          className="block aspect-[140/32] h-7 bg-current"
          style={{ mask: `url("${gavelLockup}") no-repeat left center / contain`, WebkitMask: `url("${gavelLockup}") no-repeat left center / contain` }}
        />
      }
      // The rail (logo, Files, Tasks) collapses; the breadcrumb stays in the top bar beside it.
      navSections={[
        {
          items: [
            {
              key: "files",
              label: "Files",
              icon: UiFolder,
              to: routeHref({ selection: route.selection, view: route.view, ...(route.line ? { line: route.line } : {}) }),
              active: route.tasks === undefined,
              ...(unreadComments ? { badge: unreadComments } : {}),
            },
            {
              key: "tasks",
              label: "Tasks",
              icon: UiListChecks,
              to: routeHref({ selection: null, view: route.view, tasks: "" }),
              active: route.tasks !== undefined,
            },
          ],
        },
      ]}
      collapsible
      collapsedStorageKey="gavel-edit:rail-collapsed"
      // Lint runs as a clicky task group; the indicator follows it (and any other task) live,
      // and its links open the full task manager below.
      actions={
        <div className="flex items-center gap-1">
        <RecentComments
          recent={summary.recent}
          unread={unreadComments}
          pathOf={(fileId) => {
            const entry = files.find((file) => file.id === fileId);
            if (!entry) throw new Error(`comment on ${fileId}, which is not one of the edited files`);
            return relativePath(entry);
          }}
          onOpen={(comment) => {
            const entry = files.find((file) => file.id === comment.fileId);
            if (!entry) return setError(`comment on ${comment.fileId}, which is not one of the edited files`);
            openFile(entry, route.view);
          }}
        />
        <TaskManagerButton
          basePath="/api/v1"
          onNavigate={(href) => {
            try {
              navigate({ selection: null, view: route.view, tasks: taskRunFromHref(href) });
            } catch (cause) {
              setError(String(cause));
            }
          }}
        />
        <CurrentUser onError={setError} />
        </div>
      }
      nav={
        route.tasks !== undefined ? (
          <nav aria-label="Breadcrumb" className="flex min-w-0 items-center gap-1 font-mono text-sm">
            <button type="button" className="text-muted-foreground hover:text-foreground" onClick={() => navigate({ selection: null, view: route.view })}>
              Files
            </button>
            <span className="text-muted-foreground/60">/</span>
            <span aria-current="page" className="font-medium text-foreground">
              Tasks
            </span>
          </nav>
        ) : crumbs.length > 0 && (
          <nav aria-label="Breadcrumb" className="flex min-w-0 items-center gap-1 font-mono text-sm">
            {crumbs.map((node, index) => (
              <Fragment key={node.key}>
                {index > 0 && <span className="text-muted-foreground/60">/</span>}
                {index < crumbs.length - 1 ? (
                  <button
                    type="button"
                    className="truncate text-muted-foreground hover:text-foreground"
                    onClick={() => navigate({ selection: { kind: node.kind, ref: node.ref }, view: route.view })}
                  >
                    {node.name}
                  </button>
                ) : (
                  <span aria-current="page" className="truncate font-medium text-foreground" title={node.ref.file}>
                    {node.name}
                  </span>
                )}
              </Fragment>
            ))}
          </nav>
        )
      }
      search={
        <Button variant="outline" onClick={() => setPalette(true)}>
          Search files · ⌘K
        </Button>
      }
      contentWidth="full"
      contentClassName="flex min-h-0 flex-col overflow-hidden"
    >
      {(error || schemaError) && (
        <p role="alert" className="px-3 py-1 text-sm text-destructive">
          {error || schemaError}
        </p>
      )}
      {route.tasks !== undefined ? (
        <div className="min-h-0 flex-1 overflow-auto p-4">
          <TaskManager
            basePath="/api/v1"
            {...(route.tasks ? { selectedId: route.tasks } : {})}
            onSelectRun={(id) => navigate({ selection: null, view: route.view, tasks: id ?? "" })}
          />
        </div>
      ) : (
      <FileFindingsView
        key={generation}
        className="min-h-0 flex-1"
        files={shownRefs}
        findings={shownFindings}
        treeNode={{
          badges: ({ node, default: counts }) => (
            <>
              {commentBadge(node.kind, node.ref.file)}
              {counts}
            </>
          ),
        }}
        // A file opens once it is listed: the view loads it as soon as it is selected.
        selected={route.selection?.kind === "file" && !active ? null : route.selection}
        onSelect={(selection) => navigate({ selection, view: route.view })}
        line={route.line}
        onLineChange={(line) => navigate({ ...route, ...(line ? { line } : { line: undefined }) })}
        view={route.view}
        onViewChange={(view) => navigate({ selection: route.selection, view })}
        showPath={false}
        problemFilters={filters}
        onProblemFiltersChange={setFilters}
        loadFile={loadFile}
        fileUrl={fileUrl}
        onSourceChange={(ref, source) => queues.current.get(entryOf(ref).id)?.edit(source)}
        fileToolbar={fileToolbar}
        fileTabs={fileTabs}
        getFindingActions={findingActions}
        getFileActions={(ref) => {
          const entry = byRef.get(fileRefKey(ref));
          return entry?.kind === "document" ? [lintItem("Lint file", [entry])] : [];
        }}
        getFolderActions={(folder) => [
          lintItem("Lint folder", files.filter((file) => !folder.file || relativePath(file).startsWith(`${folder.file}/`))),
        ]}
        comments={(ref) => (active && fileRefKey(ref) === fileRefKey(activeRef!) ? comments : undefined)}
        selectToComment
        treeActions={
          <div className="flex items-center gap-2 text-xs">
            <span className="flex items-center gap-1 text-muted-foreground">
              Comments:
              <SegmentedControl<CommentFilter>
                size="sm"
                value={commentFilter}
                options={[
                  { id: "unread", label: unreadComments ? `Unread ${unreadComments}` : "Unread" },
                  { id: "all", label: "All" },
                ]}
                onChange={(next) => {
                  setCommentFilter(next);
                  writeCommentFilter(next);
                }}
              />
            </span>
            {run && (
              <span role="status" className="text-muted-foreground">
                {[
                  run.running ? "Linting…" : run.error ? run.error : run.recorded ? `From gavel run ${recordedLabel(run.recorded)}` : null,
                  ...(run.active ?? []).map(activeRunLabel),
                  !run.running && stale ? "Results are stale" : null,
                ]
                  .filter(Boolean)
                  .join(" · ")}
              </span>
            )}
            <DropdownMenu
              align="right"
              icon={UiDotsVertical}
              hideChevron
              variant="ghost"
              size="sm"
              title="All files actions"
              menuLabel="All files actions"
              items={[lintItem("Lint all files", files)]}
            />
          </div>
        }
      />
      )}
      <CommandPalette
        open={palette}
        onOpenChange={setPalette}
        ariaLabel="Search files"
        placeholder="Search filenames or paths…"
        groups={[
          {
            id: "files",
            heading: "Files",
            items: files.map((file) => ({
              id: file.id,
              label: file.path.split("/").pop()!,
              description: file.path,
              onSelect: () => openFile(file, isImagePath(file.path) ? "preview" : route.view),
            })),
          },
        ]}
      />
      {missingProject && active && skippedProject !== active.id && (
        <CreateProjectDialog missing={missingProject} onCreate={createProject} onClose={() => setSkippedProject(active.id)} />
      )}
      {newSession && (
        <NewAgentSessionDialog
          thread={newSession.thread}
          fileId={newSession.fileId}
          onPreview={newSession.preview}
          onStart={newSession.start}
          onClose={newSession.cancel}
        />
      )}
    </AppShell>
  );
}
