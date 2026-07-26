import { useNow } from '../useNow';

// Elapsed renders how long a command has been running, ticking on the shared
// clock while it is live and freezing at its total once it ends.
export function Elapsed({ startedAt, endedAt, running }: { startedAt?: string; endedAt?: string; running: boolean }) {
  if (!startedAt) return null;
  if (running) return <RunningElapsed startedAt={startedAt} />;
  return <span className="text-muted-foreground">{formatElapsed(startedAt, endedAt)}</span>;
}

function RunningElapsed({ startedAt }: { startedAt: string }) {
  const now = useNow();
  return <span className="text-muted-foreground">{formatElapsed(startedAt, undefined, now)}</span>;
}

function formatElapsed(startedAt: string, endedAt?: string, now = Date.now()) {
  const end = endedAt ? new Date(endedAt).getTime() : now;
  const seconds = Math.max(0, Math.floor((end - new Date(startedAt).getTime()) / 1000));
  if (seconds < 60) return `${seconds}s`;
  return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
}
