import { useEffect, useRef, useState } from "react";
import {
  createBrowserPreview,
  type RenderOptions,
} from "@flanksource/facet/browser-preview";
import {
  isElementAnchor,
  elementAnchor,
  parseElementAnchor,
  type FileReviewPreviewContext,
} from "@flanksource/clicky-ui/file-findings";
import { request } from "./api";

type Context = {
  page: string;
  paths: string[];
  dependencies: Record<string, string>;
  runtimeOrigin: string;
};
type Catalog = {
  versions: Record<string, { version: string }>;
  cached: string[];
  local: string[];
};

export function FacetPreview({
  id,
  review,
}: {
  id: string;
  review: FileReviewPreviewContext;
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
  const [context, setContext] = useState<Context | null>(null);
  const [retry, setRetry] = useState(0);
  const [catalogRevision, setCatalogRevision] = useState(0);
  const [versions, setVersions] = useState<
    { version: string; source: string }[]
  >([]);
  const [version, setVersion] = useState(
    () =>
      new URLSearchParams(location.search).get("facet") ??
      localStorage.getItem(`gavel-edit:facet:${id}`) ??
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
        if (!cancelled)
          setVersions(
            Object.keys(catalog.versions)
              .reverse()
              .map((version) => ({
                version,
                source: catalog.local.includes(version)
                  ? "local"
                  : catalog.cached.includes(version)
                    ? "cached"
                    : "npm",
              })),
          );
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
    preview.current?.invalidate();
    void request<Context>(
      `/api/file/${id}/preview${version ? `?facet=${encodeURIComponent(version)}` : ""}`,
    )
      .then((next) => {
        if (!cancelled) {
          setContext(next);
          setError("");
        }
      })
      .catch((cause) => {
        if (!cancelled) setError(String(cause));
      });
    return () => {
      cancelled = true;
    };
  }, [id, version, retry]);

  useEffect(() => {
    const current = ++revision.current;
    preview.current?.invalidate();
    if (!context) return;
    const timer = setTimeout(() => {
      serial.current = serial.current
        .then(async () => {
          const instance = preview.current;
          if (!instance || !container.current || current !== revision.current)
            return;
          setStatus("rendering");
          await instance.render({
            container: container.current,
            runtimeOrigin: context.runtimeOrigin,
            code: source,
            ext: context.page.split(".").pop()!,
            page: context.page,
            sourceProvider: {
              url: `${context.runtimeOrigin}/sources/${id}`,
              paths: context.paths,
            },
            header: "",
            footer: "",
            css: "",
            data: {},
            dependencies: context.dependencies,
            format: "pdf",
            onStatus: (state) => {
              if (current === revision.current) setStatus(state);
            },
            onError: (cause) => {
              if (current === revision.current) setError(cause.message);
            },
            onBridge: (next) => {
              bridge.current = next;
              agentReady.current = false;
            },
            onMessage: (message) => {
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
  }, [context, source, id]);

  return (
    <div className="flex h-full min-h-0 flex-col gap-2">
      <div className="flex flex-wrap items-center gap-2 px-2 text-sm">
        <label htmlFor={`facet-runtime-${id}`}>Facet runtime</label>
        <select
          id={`facet-runtime-${id}`}
          value={version}
          onChange={(event) => {
            const next = event.target.value;
            localStorage.setItem(`gavel-edit:facet:${id}`, next);
            const url = new URL(location.href);
            if (next) url.searchParams.set("facet", next);
            else url.searchParams.delete("facet");
            history.replaceState(null, "", url);
            setVersion(next);
          }}
        >
          <option value="">Project default</option>
          {versions.map((item) => (
            <option key={item.version} value={item.version}>
              {item.version} ({item.source})
            </option>
          ))}
        </select>
        <button
          type="button"
          onClick={() => setCatalogRevision((value) => value + 1)}
        >
          Refresh versions
        </button>
        <span role="status">
          {status}
          {context
            ? ` · Clicky UI ${context.dependencies["@flanksource/clicky-ui"]}`
            : ""}
        </span>
        <button
          type="button"
          disabled={status !== "ready"}
          onClick={() => bridge.current?.send({ type: "pick" })}
        >
          Comment on element
        </button>
      </div>
      {error && (
        <div>
          <pre
            role="alert"
            className="whitespace-pre-wrap p-2 text-sm text-destructive"
          >
            {error}
          </pre>
          <button type="button" onClick={() => setRetry((value) => value + 1)}>
            Retry preview
          </button>
        </div>
      )}
      <div ref={container} className="flex min-h-0 flex-1 flex-col" />
    </div>
  );
}
