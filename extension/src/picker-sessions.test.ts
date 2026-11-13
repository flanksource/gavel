import { expect, it } from "vitest";
import { pickerSessions } from "./picker-sessions";

const live = { id: "s-1", source: "claude", title: "site work", model: "claude-example", cwd: "/work/site", status: "running", inCmux: true };
const detached = { id: "s-2", source: "codex", title: "fixes", inCmux: false };

it("lists gavel's live sessions with what the picker shows, flagging one that is not in cmux", () => {
  expect(pickerSessions({ sessions: [live, detached], selectedId: undefined })).toEqual([
    { id: "s-1", source: "claude", title: "site work", model: "claude-example", cwd: "/work/site", status: "running" },
    { id: "s-2", source: "codex", title: "fixes", status: "not in cmux" },
  ]);
});

it("lists ended managed sessions after the live ones, marked as resuming", () => {
  const ended = { id: "s-3", source: "claude", title: "old run", inCmux: false, status: "ended", managed: true, resumable: true };
  expect(pickerSessions({ sessions: [ended, live, detached], selectedId: "s-3" })).toEqual([
    { id: "s-1", source: "claude", title: "site work", model: "claude-example", cwd: "/work/site", status: "running" },
    { id: "s-2", source: "codex", title: "fixes", status: "not in cmux" },
    { id: "s-3", source: "claude", title: "old run · resumes", status: "ended" },
  ]);
});

it("keeps a remembered session that is no longer live so the choice stays visible and can be changed", () => {
  expect(pickerSessions({ sessions: [live], selectedId: "s-9" }).at(-1)).toEqual({ id: "s-9", source: "agent", title: "s-9 (not live)", status: "not live" });
  expect(pickerSessions({ sessions: [live], selectedId: "s-1" })).toHaveLength(1);
});

it("has nothing to list, and nothing to flag, while the sessions are still loading", () => {
  expect(pickerSessions({ sessions: undefined, selectedId: "s-9" })).toEqual([]);
});
