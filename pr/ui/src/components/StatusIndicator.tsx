import { useEffect, useState } from 'preact/hooks';
import type { HealthStatus, Severity } from '../types';

// Colored dot in the PR UI header that polls /api/status. Overall severity
// drives the dot color; hovering reveals per-component status.
// Polls every 10s — matching ActivityView's cache refresh cadence. A
// fetch failure is treated as "down" (the daemon is up but can't answer —
// same urgency to the user as a down component).

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

export function StatusIndicator() {
  const [health, setHealth] = useState<HealthStatus | null>(null);
  const [fetchErr, setFetchErr] = useState<string | null>(null);

  useEffect(() => {
    const load = () => {
      fetch('/api/status')
        .then(r => r.ok ? r.json() : Promise.reject(new Error(`HTTP ${r.status}`)))
        .then((h: HealthStatus) => { setHealth(h); setFetchErr(null); })
        .catch(err => setFetchErr(err?.message || 'fetch failed'));
    };
    load();
    const timer = setInterval(load, POLL_INTERVAL_MS);
    return () => clearInterval(timer);
  }, []);

  // While waiting for the first response, show a muted dot rather than
  // flashing a red one.
  if (!health && !fetchErr) {
    return <div class="w-2.5 h-2.5 rounded-full bg-gray-300" title="Loading status..." />;
  }

  // Fetch failure — treat as down; the daemon responded with an error or
  // the endpoint is unreachable.
  if (fetchErr) {
    return (
      <div class="relative group">
        <div class="w-2.5 h-2.5 rounded-full bg-red-500" />
        <Tooltip title="Status unavailable" body={fetchErr} />
      </div>
    );
  }

  // Non-null here because of the early returns above.
  const h = health!;
  return (
    <div class="relative group">
      <div class={`w-2.5 h-2.5 rounded-full ${DOT_CLASS[h.overall]}`} />
      <Tooltip
        title={DOT_LABEL[h.overall]}
        body={
          <div class="flex flex-col gap-1">
            <ComponentLine label="Database" c={h.database} />
            <ComponentLine label="GitHub" c={h.github} />
          </div>
        }
      />
    </div>
  );
}

// Tooltip uses Tailwind group-hover to show/hide. Positioned top-right of
// the header so it opens downward and leftward against the viewport edge.
function Tooltip({ title, body }: { title: string; body: any }) {
  return (
    <div class="hidden group-hover:block absolute right-0 top-5 z-50 w-64 bg-white border border-gray-200 rounded shadow-lg p-3 text-xs">
      <div class="font-semibold text-gray-800 mb-1">{title}</div>
      <div class="text-gray-600">{body}</div>
    </div>
  );
}

function ComponentLine({ label, c }: { label: string; c: { severity: Severity; message: string } }) {
  return (
    <div class="flex items-start gap-2">
      <div class={`w-1.5 h-1.5 rounded-full mt-1 shrink-0 ${DOT_CLASS[c.severity]}`} />
      <div>
        <span class="font-medium text-gray-700">{label}:</span>{' '}
        <span class="text-gray-600">{c.message}</span>
      </div>
    </div>
  );
}
