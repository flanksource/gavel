import { useEffect, useId, useRef, useState } from "react";
import { Button, type WorkspaceAccessClient } from "@flanksource/clicky-ui/components";
import { ApiError } from "./api";

export function readPreviewPreference(key: string): string | undefined {
  try {
    return window.localStorage.getItem(key) ?? undefined;
  } catch {
    return undefined;
  }
}

export function writePreviewPreference(key: string, value: string | undefined): void {
  try {
    if (value === undefined) window.localStorage.removeItem(key);
    else window.localStorage.setItem(key, value);
  } catch {
    // The workspace grant is stored by Gavel even when browser storage is unavailable.
  }
}

const approvalKey = (fileId: string) => `gavel-edit:preview-approved-folder:${fileId}`;

export function rememberedPreviewFolder(fileId: string): string | undefined {
  return readPreviewPreference(approvalKey(fileId));
}

export function forgetPreviewFolder(fileId: string, folder: string): void {
  if (rememberedPreviewFolder(fileId) === folder) writePreviewPreference(approvalKey(fileId), undefined);
}

export function deniedWorkspaceFolder(cause: unknown): string | undefined {
  if (!(cause instanceof ApiError) || cause.status !== 403) return undefined;
  const denial = cause.json<{ code?: string; root?: unknown }>();
  if (denial?.code === "workspace_access_denied" && typeof denial.root === "string" && denial.root.trim()) return denial.root;
  return undefined;
}

export type ManagePreviewAccess = (folder: string, onGranted: (path: string) => void) => void;

export function PreviewWorkspaceAccess({ fileId, folder, client, onApproved, onManage }: {
  fileId: string;
  folder: string;
  client: WorkspaceAccessClient;
  onApproved: () => void;
  onManage: ManagePreviewAccess;
}) {
  const title = useId();
  const [reviewed, setReviewed] = useState(false);
  const [reviewAttempt, setReviewAttempt] = useState(0);
  const [granting, setGranting] = useState(false);
  const [approved, setApproved] = useState(false);
  const [error, setError] = useState("");
  const grantInFlight = useRef(false);
  const mounted = useRef(false);

  useEffect(() => {
    mounted.current = true;
    return () => { mounted.current = false; };
  }, []);

  useEffect(() => {
    let cancelled = false;
    setReviewed(false);
    setError("");
    void client.preview(folder).then(({ path }) => {
      if (cancelled) return;
      if (path === folder) setReviewed(true);
      else setError("Folder changed since the preview requested access. Review the folder in Workspace access.");
    }).catch((cause) => { if (!cancelled) setError(String(cause)); });
    return () => { cancelled = true; };
  }, [client, folder, reviewAttempt]);

  function finishApproval(path: string) {
    writePreviewPreference(approvalKey(fileId), path);
    if (mounted.current) {
      setApproved(true);
      onApproved();
    }
  }

  async function approve() {
    if (!reviewed || approved || grantInFlight.current) return;
    grantInFlight.current = true;
    setGranting(true);
    setError("");
    try {
      finishApproval((await client.grant(folder, folder)).path);
    } catch (cause) {
      if (mounted.current) setError(String(cause));
    } finally {
      grantInFlight.current = false;
      if (mounted.current) setGranting(false);
    }
  }

  return (
    <div className="flex min-h-0 flex-1 items-center justify-center overflow-auto px-density-4 pb-density-4 pt-14">
      <section aria-labelledby={title} className="w-full max-w-lg space-y-density-3 rounded-lg border border-border bg-card p-density-4 text-card-foreground">
        <h2 id={title} className="text-base font-semibold">Folder access required</h2>
        <p className="text-sm text-muted-foreground">This preview needs to read project files from this folder.</p>
        <code className="block break-all rounded-md bg-muted p-density-2 text-sm">{folder}</code>
        <p className="text-sm text-muted-foreground">Approval allows reading, editing, and running local tools in this folder and its subfolders across Gavel.</p>
        {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
        <div className="flex flex-wrap gap-density-2">
          {error && !reviewed ? (
            <Button onClick={() => setReviewAttempt((attempt) => attempt + 1)}>Retry folder review</Button>
          ) : (
            <Button disabled={!reviewed || granting || approved} onClick={() => void approve()}>{approved ? "Retrying preview…" : granting ? "Approving…" : "Approve access"}</Button>
          )}
          <Button variant="outline" disabled={granting || approved} onClick={() => onManage(folder, finishApproval)}>Manage workspace access</Button>
        </div>
      </section>
    </div>
  );
}
