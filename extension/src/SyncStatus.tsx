import type { SyncStatus as Status } from "./sync-status";

const TONES = {
  connected: { dot: "bg-green-500", text: "text-muted-foreground" },
  connecting: { dot: "bg-muted-foreground", text: "text-muted-foreground" },
  syncing: { dot: "bg-primary", text: "text-muted-foreground" },
  "sync-error": { dot: "bg-destructive", text: "text-destructive" },
  "connection-error": { dot: "bg-destructive", text: "text-destructive" },
} as const;

function labelOf(status: Status): string {
  switch (status.kind) {
    case "connected": return "Connected";
    case "connecting": return "Connecting…";
    case "syncing": return `Syncing ${status.current} of ${status.total}`;
    case "sync-error": return "Sync error";
    case "connection-error": return "Connection error";
  }
}

function titleOf(status: Status): string | undefined {
  if (status.kind === "connection-error") return status.detail;
  if (status.kind !== "sync-error") return undefined;
  const failed = status.failed ? `${status.failed} failed` : undefined;
  return [failed, status.detail].filter(Boolean).join(": ") || undefined;
}

/** The popup footer's single status label. With `onOpen`, a sync error becomes a button that opens the saved-captures queue. */
export function SyncStatus({ status, onOpen }: { status: Status; onOpen?: (() => void) | undefined }) {
  const tone = TONES[status.kind];
  const label = labelOf(status);
  const title = titleOf(status);
  const interactive = status.kind === "sync-error" && onOpen;
  return <div role="status" {...(!interactive && title ? { title } : {})} className={`flex items-center gap-1.5 text-xs ${tone.text}`}>
    <span aria-hidden="true" className={`size-1.5 shrink-0 rounded-full ${tone.dot}`} />
    {interactive
      ? <button type="button" onClick={onOpen} {...(title ? { title } : {})} className="rounded-sm underline-offset-2 hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring">{label}</button>
      : label}
  </div>;
}
