import { useCallback, useEffect, useState } from "react";
import { request } from "./api";

export type FileCommentCounts = { total: number; unread: number };
export type RecentComment = {
  fileId: string;
  commentId: string;
  body: string;
  author: { id: string; kind: string; label?: string };
  at: string;
  status: string;
  unread: boolean;
};
/** GET /api/comments: per-file counts keyed by file id, and the newest threads first. */
export type CommentSummary = { files: Record<string, FileCommentCounts>; recent: RecentComment[] };

const EMPTY: CommentSummary = { files: {}, recent: [] };

/** A tree row's comment counts: a file's own, or the sum over the files beneath a folder ("" is the root). */
export function commentCountsAt(
  byPath: ReadonlyMap<string, FileCommentCounts>,
  kind: "file" | "folder",
  path: string,
): FileCommentCounts {
  if (kind === "file") return byPath.get(path) ?? { total: 0, unread: 0 };
  const sum = { total: 0, unread: 0 };
  for (const [file, counts] of byPath) {
    if (path && !file.startsWith(`${path}/`)) continue;
    sum.total += counts.total;
    sum.unread += counts.unread;
  }
  return sum;
}

/** Polls the comment summary while the page is visible; `refresh` re-reads it now. */
export function useCommentSummary(onError: (message: string) => void) {
  const [summary, setSummary] = useState<CommentSummary>(EMPTY);
  const refresh = useCallback(async () => {
    try {
      setSummary(await request<CommentSummary>("/api/comments"));
    } catch (cause) {
      onError(String(cause));
    }
  }, []);
  useEffect(() => {
    void refresh();
    const timer = setInterval(() => !document.hidden && void refresh(), 5000);
    return () => clearInterval(timer);
  }, [refresh]);
  return { summary, refresh };
}
