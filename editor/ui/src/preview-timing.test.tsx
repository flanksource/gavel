import { expect, it, vi } from "vitest";
import { renderToStaticMarkup } from "react-dom/server";
import type { ActivityMessage } from "@flanksource/page-runtime/host";
import { PreviewTiming, trackPreviewBuild } from "./preview-timing";

it("reports the current build and excludes work from earlier renders of the same frame", () => {
  const onTiming = vi.fn();
  const build = trackPreviewBuild(1100, onTiming);
  const activity: ActivityMessage = {
    type: "activity", load: "page", navigationStart: 1000, records: [],
    timeline: [{ phase: "ready", atMs: 89 }],
    tailwind: [
      { kind: "candidates", startMs: 30, durationMs: 50, candidates: 10, cssBytes: 32, cacheHit: false },
      { kind: "candidates", startMs: 120, durationMs: 7, candidates: 10, cssBytes: 32, cacheHit: false },
    ],
  };
  build.activity(activity);
  expect(onTiming).not.toHaveBeenCalled();
  build.ready(1130);
  build.ready(1200);
  build.activity({ ...activity, tailwind: [
    { kind: "libraries", startMs: 105, durationMs: 5, candidates: 0, cssBytes: 0, cacheHit: false },
    { kind: "candidates", startMs: 150, durationMs: 10, candidates: 10, cssBytes: 32, cacheHit: false },
  ] });
  expect(onTiming).toHaveBeenCalledTimes(2);
  expect(onTiming).toHaveBeenLastCalledWith([
    { name: "total", duration: 30, counters: {} },
    { name: "tailwind", duration: 12, counters: { builds: 2 } },
  ]);
});

it("shows the selected runtime build provenance alongside live preview timings", () => {
  const builtAt = "2026-11-17T10:00:00Z";
  const commit = "0123456789abcdef0123456789abcdef01234567";
  const view = renderToStaticMarkup(<PreviewTiming origin="http://127.0.0.1:9100" runtimeBuild={{ builtAt, git: { commit, dirty: true } }} server={[{ name: "total", duration: 14, counters: {} }]} build={[{ name: "total", duration: 320, counters: {} }, { name: "compile", duration: 92, counters: {} }]} />);
  for (const text of ["127.0.0.1:9100", "Runtime built", builtAt, commit, "0123456789ab", "dirty", "Server timing", "14 ms", "Build timing", "320 ms", "Compile", "92 ms"]) expect(view).toContain(text);
});
