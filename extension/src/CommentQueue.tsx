import { useCallback, useEffect, useRef, useState } from "react";
import { Button } from "@flanksource/clicky-ui";
import type { CommentReceipt, CommentSubmission } from "./comment-sync";
import { requestBackground } from "./messages";

const POLL_INTERVAL_MS = 1500;
const EMPTY: CommentReceipt[] = [];

export function useCommentQueue(pageUrl: string | undefined) {
  const [snapshot, setSnapshot] = useState<{ url: string; rows: CommentReceipt[] }>();
  const [failure, setFailure] = useState<{ url: string; message: string }>();
  const current = useRef(pageUrl);
  current.current = pageUrl;
  const generation = useRef(0);
  const refresh = useCallback(async () => {
    if (!pageUrl || pageUrl !== current.current) return;
    const request = ++generation.current;
    try {
      const rows = await requestBackground<CommentReceipt[]>({ type: "comment-queue", pageUrl });
      if (current.current !== pageUrl || generation.current !== request) return;
      setSnapshot({ url: pageUrl, rows });
      setFailure(undefined);
    } catch (cause) {
      if (current.current === pageUrl && generation.current === request) setFailure({ url: pageUrl, message: String(cause) });
    }
  }, [pageUrl]);
  const rows = snapshot?.url === pageUrl ? snapshot?.rows ?? EMPTY : EMPTY;
  const pending = rows.some(row => row.state === "pending");
  useEffect(() => { void refresh(); }, [refresh]);
  useEffect(() => {
    if (!pending) return;
    const timer = window.setInterval(() => void refresh(), POLL_INTERVAL_MS);
    return () => window.clearInterval(timer);
  }, [pending, refresh]);
  const capture = async (draft: CommentSubmission) => {
    const id = await requestBackground<string>({ type: "comment-capture", draft });
    void refresh();
    return id;
  };
  const retry = async (id: string) => {
    if (!pageUrl) throw new Error("The grabber has not reported the page yet");
    await requestBackground({ type: "comment-retry", id, pageUrl });
    await refresh();
  };
  return { rows, error: failure?.url === pageUrl ? failure?.message : undefined, capture, retry };
}

export function CommentQueue({ rows, retry }: Pick<ReturnType<typeof useCommentQueue>, "rows" | "retry">) {
  const [retrying, setRetrying] = useState<string>();
  const [error, setError] = useState<string>();
  const pending = rows.filter(row => row.state !== "synced");
  async function retryComment(id: string) {
    setRetrying(id);
    setError(undefined);
    try { await retry(id); }
    catch (cause) { setError(String(cause)); }
    finally { setRetrying(undefined); }
  }
  if (!pending.length) return null;
  return <section aria-label="Queued comments" className="shrink-0 space-y-1 border-b border-border px-3 py-2 text-xs">
    <p role="status">{pending.length} {pending.length === 1 ? "comment" : "comments"} awaiting sync</p>
    {error && <p role="alert" className="text-destructive">{error}</p>}
    {pending.filter(row => row.error).map(row => <details key={row.id}>
      <summary className="truncate">{row.state === "pending" ? "Waiting for Gavel" : "Comment could not sync"} · {row.body}</summary>
      <p className="break-words text-destructive">{row.error}</p>
      {row.state === "error" && <Button size="xs" variant="outline" loading={retrying === row.id} disabled={!!retrying} onClick={() => void retryComment(row.id)}>Retry comment</Button>}
    </details>)}
  </section>;
}
