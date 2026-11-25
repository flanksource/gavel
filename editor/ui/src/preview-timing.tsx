import {
  formatServerTimingDuration,
  serverTimingMetricLabel,
  type ServerTimingMetric,
} from "@flanksource/clicky-ui/data";
import {
  activityTimings,
  appendActivity,
  emptyActivity,
  type ActivityMessage,
} from "@flanksource/page-runtime/host";

export function trackPreviewBuild(startedAt: number, onTiming: (metrics: ServerTimingMetric[]) => void) {
  let activity = emptyActivity();
  let readyAt: number | undefined;

  function publish() {
    if (readyAt === undefined) return;
    const elapsed = readyAt - startedAt;
    const load = activity.current;
    onTiming(load
      ? activityTimings({ ...load, timeline: [{ phase: "ready", atMs: elapsed }] })
      : [{ name: "total", duration: elapsed, counters: {} }]);
  }

  return {
    ready(at: number) {
      if (readyAt !== undefined) return;
      readyAt = at;
      publish();
    },
    activity(message: ActivityMessage) {
      // Edits reuse the frame and its clock. Only count work from this render.
      const duringBuild = (at: number) => at >= startedAt && (readyAt === undefined || at <= readyAt);
      const records = message.records.filter((record) => duringBuild(record.startedAt));
      const tailwind = message.tailwind.filter((build) => duringBuild(message.navigationStart + build.startMs));
      if (!records.length && !tailwind.length) return;
      activity = appendActivity({ state: activity, message: { ...message, records, tailwind } });
      publish();
    },
  };
}

function TimingGroup({ label, metrics }: { label: string; metrics: ServerTimingMetric[] }) {
  const total = metrics.find((metric) => metric.name === "total");
  if (!total) return null;
  const rows = [{ label, duration: total.duration }, ...metrics
    .filter((metric) => metric.name !== "total")
    .map((metric) => ({ label: serverTimingMetricLabel(metric), duration: metric.duration }))];
  return (
    <dl aria-label={label} className="grid grid-cols-[1fr_auto] gap-x-4 gap-y-0.5">
      {rows.map((row, index) => (
        <div key={row.label} className="contents">
          <dt className={index === 0 ? "font-medium text-foreground" : "pl-2"}>{row.label}</dt>
          <dd className="text-right tabular-nums">{formatServerTimingDuration(row.duration)}</dd>
        </div>
      ))}
    </dl>
  );
}

export type RuntimeBuild = {
  builtAt?: string;
  git?: { commit?: string; dirty: boolean };
};

export function PreviewTiming({ origin, runtimeBuild, server, build }: {
  origin?: string | undefined;
  runtimeBuild?: RuntimeBuild | undefined;
  server: ServerTimingMetric[];
  build: ServerTimingMetric[];
}) {
  if (!origin && !runtimeBuild && !server.length && !build.length) return null;
  const { builtAt, git } = runtimeBuild ?? {};
  return (
    <div className="mt-2 space-y-2 border-t border-border pt-2">
      {origin && <div>Runtime server <span className="ml-1 break-all font-mono text-foreground">{origin}</span></div>}
      {runtimeBuild && (
        <dl aria-label="Runtime build" className="grid grid-cols-[auto_1fr] gap-x-2 gap-y-0.5">
          <dt>Runtime built</dt>
          <dd className="text-foreground">
            {builtAt ? <time dateTime={builtAt} title={builtAt}>{new Date(builtAt).toLocaleString()}</time> : "Unknown"}
          </dd>
          <dt>Git</dt>
          <dd className="text-foreground">
            {git ? <><code title={git.commit}>{git.commit?.slice(0, 12) ?? "No commits"}</code> ({git.dirty ? "dirty" : "clean"})</> : "Unknown"}
          </dd>
        </dl>
      )}
      <TimingGroup label="Server timing" metrics={server} />
      <TimingGroup label="Build timing" metrics={build} />
    </div>
  );
}
