import type { ReactNode } from 'react';
import { useCallback, useEffect, useState } from 'react';
import { DropdownMenu } from '@flanksource/clicky-ui/components';
import { Version } from '@flanksource/clicky-ui/data';
import type { HealthStatus, RateLimit, Severity } from '../types';

// Colored dot in the PR UI header. The dropdown owns the operational status,
// PR poller controls, GitHub rate-limit details, and build versions so the app
// bar does not carry multiple competing status controls.

const POLL_INTERVAL_MS = 10_000;

const DOT_CLASS: Record<Severity, string> = {
  ok: 'bg-green-500',
  degraded: 'bg-yellow-500',
  down: 'bg-red-500',
};

const DOT_LABEL: Record<Severity, string> = {
  ok: 'All systems operational',
  degraded: 'Partial degradation',
  down: 'Service issue',
};

interface Props {
  fetchedAt: string;
  nextFetchIn: number;
  paused: boolean;
  rateLimit?: RateLimit;
  error?: string;
  onRefresh: () => void;
  onPause: () => void;
  networkBusy?: boolean;
}

// Backend build metadata injected by the Go server (see pr/ui/handler.go),
// frontend metadata baked in by Vite (see pr/ui/vite.config.ts).
const backend = typeof window !== 'undefined' ? window.__GAVEL__ : undefined;
const frontend = { version: __GAVEL_UI_VERSION__, commit: __GAVEL_UI_COMMIT__ };

export function StatusIndicator({ fetchedAt, nextFetchIn, paused, rateLimit, error, onRefresh, onPause, networkBusy }: Props) {
  const [health, setHealth] = useState<HealthStatus | null>(null);
  const [fetchErr, setFetchErr] = useState<string | null>(null);
  const [healthLoading, setHealthLoading] = useState(false);

  const loadHealth = useCallback(() => {
    setHealthLoading(true);
    fetch('/api/status')
      .then(r => r.ok ? r.json() : Promise.reject(new Error(`HTTP ${r.status}`)))
      .then((h: HealthStatus) => { setHealth(h); setFetchErr(null); })
      .catch(err => setFetchErr(err?.message || 'fetch failed'))
      .finally(() => setHealthLoading(false));
  }, []);

  useEffect(() => {
    loadHealth();
    const timer = setInterval(loadHealth, POLL_INTERVAL_MS);
    return () => clearInterval(timer);
  }, [loadHealth]);

  const dotClass = fetchErr
    ? 'bg-red-500'
    : health
      ? DOT_CLASS[health.overall]
      : 'bg-muted-foreground/40';
  const title = fetchErr ? 'Status unavailable' : health ? DOT_LABEL[health.overall] : 'Loading status...';

  return (
    <DropdownMenu
      align="right"
      menuLabel="Status"
      menuClassName="w-[360px] max-w-[calc(100vw-24px)]"
      trigger={
        <button
          type="button"
          title={title}
          aria-label="Status, sync, and version details"
          className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
        >
          <span className={`h-2.5 w-2.5 rounded-full ${dotClass}`} />
        </button>
      }
    >
      {() => (
        <div className="p-3 text-xs text-foreground">
          <Section title="Health">
            {fetchErr ? (
              <div className="text-red-500">{fetchErr}</div>
            ) : health ? (
              <div className="space-y-1.5">
                <ComponentLine label="Overall" c={{ severity: health.overall, message: DOT_LABEL[health.overall] }} />
                <ComponentLine label="Database" c={health.database} />
                <ComponentLine label="GitHub" c={health.github} />
                <div className="text-[11px] text-muted-foreground">
                  Checked {timeAgoShort(health.checkedAt)}
                </div>
              </div>
            ) : (
              <div className="text-muted-foreground">Loading status...</div>
            )}
          </Section>

          <Section title="Rate limit">
            <RateLimitStatus rateLimit={rateLimit} />
          </Section>

          <Section title="Sync">
            <SyncStatus
              fetchedAt={fetchedAt}
              nextFetchIn={nextFetchIn}
              paused={paused}
              error={error}
              networkBusy={networkBusy}
              healthLoading={healthLoading}
              onRefresh={onRefresh}
              onPause={onPause}
              onStatusRefresh={loadHealth}
            />
          </Section>

          <Section title="Versions" last>
            <div className="space-y-1.5">
              <VersionRow label="clicky-ui">
                <Version date={false} />
              </VersionRow>
              <VersionRow label="gavel backend">
                {backend?.version || 'unknown'}
                {backend?.commit && backend.commit !== 'unknown' && (
                  <span className="text-muted-foreground"> · {backend.commit}</span>
                )}
              </VersionRow>
              <VersionRow label="gavel frontend">
                {frontend.version}
                {frontend.commit !== 'unknown' && (
                  <span className="text-muted-foreground"> · {frontend.commit}</span>
                )}
              </VersionRow>
            </div>
          </Section>
        </div>
      )}
    </DropdownMenu>
  );
}

function Section({ title, children, last }: { title: string; children: ReactNode; last?: boolean }) {
  return (
    <div className={`${last ? '' : 'border-b border-border pb-3 mb-3'}`}>
      <div className="mb-2 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
        {title}
      </div>
      {children}
    </div>
  );
}

function ComponentLine({ label, c }: { label: string; c: { severity: Severity; message: string } }) {
  return (
    <div className="flex items-start gap-2">
      <div className={`w-1.5 h-1.5 rounded-full mt-1 shrink-0 ${DOT_CLASS[c.severity]}`} />
      <div>
        <span className="font-medium text-foreground">{label}:</span>{' '}
        <span className="text-muted-foreground">{c.message}</span>
      </div>
    </div>
  );
}

function RateLimitStatus({ rateLimit }: { rateLimit?: RateLimit }) {
  if (!rateLimit) {
    return (
      <div className="space-y-1.5">
        <div className="h-2 w-full overflow-hidden rounded-full bg-muted" />
        <div className="flex items-center justify-between text-muted-foreground">
          <span>GitHub API pending</span>
          <span className="font-mono">--/--</span>
        </div>
      </div>
    );
  }

  const pct = rateLimit.limit > 0 ? Math.max(0, Math.min(100, (rateLimit.remaining / rateLimit.limit) * 100)) : 0;
  const tone = pct > 50 ? 'bg-green-500' : pct > 15 ? 'bg-yellow-500' : 'bg-red-500';
  const resetAt = new Date(rateLimit.reset * 1000).toLocaleTimeString();

  return (
    <div className="space-y-1.5">
      <div
        className="h-2 w-full overflow-hidden rounded-full bg-muted"
        title={`GitHub API (${rateLimit.resource}): ${rateLimit.remaining}/${rateLimit.limit} remaining`}
      >
        <div className={`h-full ${tone} transition-all duration-300`} style={{ width: `${pct}%` }} />
      </div>
      <div className="flex items-center justify-between gap-3">
        <span className="text-muted-foreground">{rateLimit.resource || 'core'} API</span>
        <span className="font-mono">{rateLimit.remaining}<span className="text-muted-foreground">/{rateLimit.limit}</span></span>
      </div>
      <div className="flex items-center justify-between gap-3 text-[11px] text-muted-foreground">
        <span>{rateLimit.used} used</span>
        <span>Resets {resetAt}</span>
      </div>
    </div>
  );
}

interface SyncProps {
  fetchedAt: string;
  nextFetchIn: number;
  paused: boolean;
  error?: string;
  networkBusy?: boolean;
  healthLoading: boolean;
  onRefresh: () => void;
  onPause: () => void;
  onStatusRefresh: () => void;
}

function SyncStatus({ fetchedAt, nextFetchIn, paused, error, networkBusy, healthLoading, onRefresh, onPause, onStatusRefresh }: SyncProps) {
  const ago = fetchedAt ? timeAgoShort(fetchedAt) : 'never';
  const countdown = fetchedAt
    ? Math.max(0, nextFetchIn - Math.floor((Date.now() - new Date(fetchedAt).getTime()) / 1000))
    : nextFetchIn;

  const cfg = error
    ? { icon: 'codicon:warning', color: 'text-red-500', label: 'Sync error' }
    : networkBusy
      ? { icon: 'svg-spinners:ring-resize', color: 'text-blue-500', label: 'Syncing' }
      : paused
        ? { icon: 'codicon:debug-pause', color: 'text-yellow-500', label: 'Paused' }
        : { icon: 'codicon:sync', color: 'text-green-500', label: 'Synced' };

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <iconify-icon icon={cfg.icon} className={`${cfg.color} text-sm`} />
        <div className="min-w-0 flex-1">
          <div className={`font-medium ${cfg.color}`}>{cfg.label}</div>
          <div className="text-[11px] text-muted-foreground">
            Last synced: {ago}
          </div>
        </div>
        <div className="text-right text-[11px] text-muted-foreground">
          {paused ? 'Polling paused' : countdown > 0 ? `${countdown}s` : 'Refreshing...'}
        </div>
      </div>
      {error && <div className="truncate text-red-500" title={error}>{error}</div>}
      <div className="flex items-center justify-end gap-1.5">
        <ActionButton icon={networkBusy ? 'svg-spinners:ring-resize' : 'codicon:sync'} label="Resync" onClick={onRefresh} />
        <ActionButton
          icon={paused ? 'codicon:debug-start' : 'codicon:debug-pause'}
          label={paused ? 'Resume' : 'Pause'}
          onClick={onPause}
        />
        <ActionButton
          icon={healthLoading ? 'svg-spinners:ring-resize' : 'codicon:refresh'}
          label="Refresh"
          onClick={onStatusRefresh}
        />
      </div>
    </div>
  );
}

function ActionButton({ icon, label, onClick }: { icon: string; label: string; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      title={label}
      className="inline-flex h-7 items-center gap-1 rounded border border-border px-2 text-[11px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
    >
      <iconify-icon icon={icon} className="text-xs" />
      {label}
    </button>
  );
}

function VersionRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex items-baseline justify-between gap-4">
      <span className="whitespace-nowrap text-muted-foreground">{label}</span>
      <span className="min-w-0 text-right font-mono text-foreground">{children}</span>
    </div>
  );
}

function timeAgoShort(iso: string): string {
  const d = new Date(iso);
  const s = Math.floor((Date.now() - d.getTime()) / 1000);
  if (!Number.isFinite(s)) return 'unknown';
  if (s < 5) return 'just now';
  if (s < 60) return `${s}s ago`;
  if (s < 3600) return `${Math.floor(s / 60)}m ago`;
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`;
  return `${Math.floor(s / 86400)}d ago`;
}
