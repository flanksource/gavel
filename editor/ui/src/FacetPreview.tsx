import { useEffect, useRef, useState } from "react";
import {
  createBrowserPreview,
  type BrowserExports,
  type RenderOptions,
} from "@flanksource/facet/browser-preview";
import { Button, DropdownMenu, type WorkspaceAccessClient } from "@flanksource/clicky-ui/components";
import {
  formatServerTimingDuration,
  parseServerTiming,
  StatusDot,
  type BadgeStatus,
  type ServerTimingMetric,
} from "@flanksource/clicky-ui/data";
import {
  UiCheck,
  UiCursorClick,
  UiDownload,
  UiExport,
  UiRefresh,
} from "@flanksource/clicky-ui/icons";
import {
  isElementAnchor,
  elementAnchor,
  parseElementAnchor,
  type FileReviewPreviewContext,
} from "@flanksource/clicky-ui/file-findings";
import { ApiError, request } from "./api";
import { previewDataOf } from "./data-preview";
import { PreviewTiming, trackPreviewBuild, type RuntimeBuild } from "./preview-timing";
import { deniedWorkspaceFolder, forgetPreviewFolder, PreviewWorkspaceAccess, readPreviewPreference, writePreviewPreference, type ManagePreviewAccess } from "./preview-workspace-access";

type Context = {
  page: string;
  paths: string[];
  dependencies: Record<string, string>;
  runtimeOrigin: string;
  runtime?: string;
  runtimeBuild?: RuntimeBuild;
};
type Catalog =
  | { pinnedRuntime: string }
  | { installed: string[] }
  | {
      versions: Record<string, { version: string }>;
      cached: string[];
      local: string[];
    };

/** A JSON or YAML document rendered through the preview-root-relative component it declares. */
export type DataPreview = { component: string; path: string };

/**
 * The Facet preview of file `id`: the page itself, or for a data document
 * (`data`) its declared component, given the parsed document as its data.
 */
export function FacetPreview({
  id,
  review,
  data,
  accessClient,
  onManageAccess,
}: {
  id: string;
  review: FileReviewPreviewContext;
  data?: DataPreview | undefined;
  accessClient: WorkspaceAccessClient;
  onManageAccess: ManagePreviewAccess;
}) {
  const source = review.source;
  const latest = useRef(review);
  latest.current = review;
  const bridge =
    useRef<Parameters<NonNullable<RenderOptions["onBridge"]>>[0]>(null);
  const agentReady = useRef(false);
  const container = useRef<HTMLDivElement>(null);
  const preview = useRef<ReturnType<typeof createBrowserPreview> | null>(null);
  const serial = useRef(Promise.resolve());
  const revision = useRef(0);
  const [status, setStatus] = useState("connecting");
  const [error, setError] = useState("");
  const [blockedFolder, setBlockedFolder] = useState<string>();
  const [context, setContext] = useState<Context | null>(null);
  const [serverTiming, setServerTiming] = useState<ServerTimingMetric[]>([]);
  const [buildTiming, setBuildTiming] = useState<ServerTimingMetric[]>([]);
  const [retry, setRetry] = useState(0);
  const [exports, setExports] = useState<BrowserExports | null>(null);
  const [catalogRevision, setCatalogRevision] = useState(0);
  const [catalog, setCatalog] = useState<Catalog | null>(null);
  const [selection, setSelection] = useState(
    () =>
      new URLSearchParams(location.search).get("facet") ??
      readPreviewPreference(`gavel-edit:facet:${id}`) ??
      "",
  );

  function sendPins() {
    if (!agentReady.current) return;
    bridge.current?.send({
      type: "pins",
      focus: latest.current.focusedAnchor,
      anchors: latest.current.comments
        .filter((comment) => comment.anchor && isElementAnchor(comment.anchor))
        .map((comment) => ({
          anchor: comment.anchor!,
          count: 1,
          locators: parseElementAnchor(comment.anchor!),
        })),
    });
  }
  useEffect(sendPins, [review.comments, review.focusedAnchor]);

  useEffect(() => {
    const instance = createBrowserPreview();
    preview.current = instance;
    return () => {
      revision.current++;
      instance.dispose();
      preview.current = null;
    };
  }, []);

  useEffect(() => {
    let cancelled = false;
    void request<Catalog>(
      `/api/preview/runtimes${catalogRevision ? "?refresh=1" : ""}`,
    )
      .then((catalog) => {
        if (!cancelled) setCatalog(catalog);
      })
      .catch((cause) => {
        if (!cancelled) setError(String(cause));
      });
    return () => {
      cancelled = true;
    };
  }, [catalogRevision]);

  useEffect(() => {
    let cancelled = false;
    setContext(null);
    setBlockedFolder(undefined);
    setStatus("connecting");
    setServerTiming([]);
    preview.current?.invalidate();
    const parameter = selection.includes("@") ? "runtime" : "facet";
    void fetch(
      `/api/file/${id}/preview${selection ? `?${parameter}=${encodeURIComponent(selection)}` : ""}`,
    )
      .then(async (response) => {
        if (!cancelled) setServerTiming(parseServerTiming(response.headers.get("Server-Timing")));
        if (!response.ok) throw new ApiError(response.status, await response.text());
        return response.json() as Promise<Context>;
      })
      .then((next) => {
        if (!cancelled) {
          setContext(next);
          setError("");
        }
      })
      .catch((cause) => {
        if (!cancelled) reportPreviewError(cause);
      });
    return () => {
      cancelled = true;
    };
  }, [id, selection, retry]);

  const [componentCode, setComponentCode] = useState<string | null>(null);
  useEffect(() => {
    setComponentCode(null);
    if (!context || !data) return;
    let cancelled = false;
    void fetch(`${context.runtimeOrigin}/sources/${id}/${data.component}`)
      .then(async (response) => {
        const text = await response.text();
        if (!response.ok) throw new ApiError(response.status, text);
        if (!cancelled) setComponentCode(text);
      })
      .catch((cause) => {
        if (!cancelled) reportPreviewError(cause);
      });
    return () => {
      cancelled = true;
    };
  }, [context, id, data?.component]);

  useEffect(() => {
    const current = ++revision.current;
    setBuildTiming([]);
    preview.current?.invalidate();
    if (!context) return;
    let page = context.page;
    let code = source;
    let pageData: RenderOptions["data"] = {};
    if (data) {
      if (componentCode === null) return;
      page = data.component;
      code = componentCode;
      try {
        pageData = previewDataOf(source, data.path);
      } catch (cause) {
        setError(`${data.path}: ${String(cause)}`);
        return;
      }
    }
    const timer = setTimeout(() => {
      serial.current = serial.current
        .then(async () => {
          const instance = preview.current;
          if (!instance || !container.current || current !== revision.current)
            return;
          setStatus("rendering");
          const build = trackPreviewBuild(performance.timeOrigin + performance.now(), setBuildTiming);
          await instance.render({
            container: container.current,
            runtimeOrigin: context.runtimeOrigin,
            code,
            ext: page.split(".").pop()!,
            page,
            sourceProvider: {
              url: `${context.runtimeOrigin}/sources/${id}`,
              paths: context.paths,
            },
            header: "",
            footer: "",
            css: "",
            data: pageData,
            dependencies: context.dependencies,
            format: "pdf",
            onStatus: (state) => {
              if (current !== revision.current) return;
              setStatus(state);
              if (state === "ready") build.ready(performance.timeOrigin + performance.now());
            },
            onError: (cause) => {
              if (current === revision.current) setError(cause.message);
            },
            onExports: setExports,
            onBridge: (next) => {
              bridge.current = next;
              agentReady.current = false;
            },
            onMessage: (message) => {
              if (message.type === "activity" && current === revision.current) build.activity(message);
              if (message.type === "status") {
                agentReady.current = message.state === "ready";
                if (agentReady.current) sendPins();
              }
              if (message.type === "error") agentReady.current = false;
              if (message.type === "focus")
                latest.current.onFocus(message.anchor);
              if (message.type === "selection") {
                const frame = container.current
                  ?.querySelector("iframe")
                  ?.getBoundingClientRect();
                const origin = container.current
                  ?.closest("[data-file-review-preview]")
                  ?.getBoundingClientRect();
                if (!frame || !origin)
                  throw new Error(
                    "Preview comment selection has no mounted frame",
                  );
                const rect = message.selection.rect;
                latest.current.onSelectElement({
                  anchor: elementAnchor(message.selection.element.locators),
                  label: message.selection.label,
                  position: {
                    top: frame.top - origin.top + rect.top,
                    bottom: frame.top - origin.top + rect.top + rect.height,
                    left: frame.left - origin.left + rect.left + rect.width / 2,
                  },
                });
              }
            },
          });
          if (current !== revision.current) instance.invalidate();
          else setError("");
        })
        .catch((cause) => {
          if (current === revision.current) setError(String(cause));
        });
    }, 300);
    return () => {
      clearTimeout(timer);
      revision.current++;
    };
  }, [context, source, id, componentCode, data?.component, data?.path]);

  function reportPreviewError(cause: unknown) {
    const folder = deniedWorkspaceFolder(cause);
    if (!folder) {
      setError(String(cause));
      return;
    }
    forgetPreviewFolder(id, folder);
    setContext(null);
    setBlockedFolder(folder);
    setError("");
  }

  const retryPreview = () => setRetry((value) => value + 1);

  function chooseRuntime(next: string) {
    writePreviewPreference(`gavel-edit:facet:${id}`, next);
    const url = new URL(location.href);
    if (next) url.searchParams.set("facet", next);
    else url.searchParams.delete("facet");
    history.replaceState(null, "", url);
    setSelection(next);
  }
  const runtimeItem = (label: string, value: string, active: boolean) => ({
    label,
    icon: UiCheck,
    iconClassName: active ? undefined : "invisible",
    group: "Facet runtime",
    onSelect: () => chooseRuntime(value),
  });
  const installed = catalog && "installed" in catalog ? catalog.installed : [];
  const pinnedRuntime =
    catalog && "pinnedRuntime" in catalog ? catalog.pinnedRuntime : undefined;
  const namedRuntime =
    context?.runtime ?? (selection.includes("@") ? selection : installed[0]);
  const runtimeLabel =
    pinnedRuntime ?? namedRuntime ?? `Facet ${selection || "default"}`;
  const runtimeItems = installed.map((selector, index) =>
    runtimeItem(
      index === 0 ? `${selector} (default)` : selector,
      index === 0 ? "" : selector,
      selector === namedRuntime,
    ),
  );
  if (catalog && "versions" in catalog) {
    runtimeItems.push(runtimeItem("Project default", "", selection === ""));
    for (const version of Object.keys(catalog.versions).reverse()) {
      const source = catalog.local.includes(version)
        ? "local"
        : catalog.cached.includes(version) ? "cached" : "npm";
      runtimeItems.push(runtimeItem(`${version} (${source})`, version, version === selection));
    }
  }
  const clickyUI = context?.dependencies["@flanksource/clicky-ui"];
  const buildTotal = buildTiming.find((metric) => metric.name === "total");
  const statusLabel = blockedFolder ? "Access required" : status;

  return (
    <div className="relative flex h-full min-h-0 flex-col">
      <DropdownMenu
        align="right"
        className="absolute right-3 top-3 z-10"
        menuLabel="Preview"
        trigger={
          <Button
            variant="outline"
            size="sm"
            className="rounded-full bg-background shadow-md"
            aria-label={`Preview ${statusLabel}`}
            title={`Preview ${statusLabel} · ${runtimeLabel}`}
          >
            <StatusDot status={previewTone(status, error || blockedFolder || "")} label={statusLabel} />
            {buildTotal && <span className="tabular-nums text-muted-foreground">{formatServerTimingDuration(buildTotal.duration)}</span>}
          </Button>
        }
        header={
          <div className="px-3 py-1 text-xs text-muted-foreground">
            {runtimeLabel}
            {clickyUI && ` · Clicky UI ${clickyUI}`}
            <PreviewTiming origin={context?.runtimeOrigin} runtimeBuild={context?.runtimeBuild} server={serverTiming} build={buildTiming} />
          </div>
        }
        items={[
          ...runtimeItems,
          ...(catalog && "versions" in catalog
            ? [{
                label: "Refresh versions",
                icon: UiRefresh,
                group: "Facet runtime",
                onSelect: () => setCatalogRevision((value) => value + 1),
              }]
            : []),
          {
            label: "Comment on element",
            icon: UiCursorClick,
            group: "Preview",
            disabled: Boolean(blockedFolder) || status !== "ready",
            onSelect: () => bridge.current?.send({ type: "pick" }),
          },
          {
            label: "Download HTML",
            icon: UiDownload,
            group: "Export",
            disabled: Boolean(blockedFolder) || !exports,
            onSelect: () => {
              const link = document.createElement("a");
              link.href = exports!.htmlUrl;
              link.download = `${(data?.path ?? context?.page ?? "report").split("/").pop()!.replace(/\.[^.]+$/, "")}.html`;
              link.click();
            },
          },
          {
            label: "Save PDF…",
            icon: UiExport,
            group: "Export",
            title: "One continuous page. Choose Save as PDF in the print dialog.",
            disabled: Boolean(blockedFolder) || !exports?.print,
            onSelect: () => exports!.print!(),
          },
        ]}
      />
      {blockedFolder ? (
        <PreviewWorkspaceAccess key={`${id}:${blockedFolder}`} fileId={id} folder={blockedFolder} client={accessClient} onApproved={retryPreview} onManage={onManageAccess} />
      ) : error && (
        <div className="flex items-start gap-2 border-b border-border p-2">
          <pre
            role="alert"
            className="min-w-0 flex-1 whitespace-pre-wrap text-sm text-destructive"
          >
            {error}
          </pre>
          <Button
            variant="outline"
            size="sm"
            onClick={retryPreview}
          >
            Retry preview
          </Button>
        </div>
      )}
      <div ref={container} className={blockedFolder ? "hidden" : "flex min-h-0 flex-1 flex-col"} />
    </div>
  );
}

/** The status-dot tone of a preview runtime state. */
function previewTone(state: string, error: string): BadgeStatus {
  if (error || state === "error") return "error";
  if (state === "ready") return "success";
  return "info";
}
