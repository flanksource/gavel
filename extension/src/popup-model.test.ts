import { expect, it } from "vitest";
import { draftOf, hasUnsavedChanges, settingsOf } from "./popup-model";

const server = "http://localhost:9092";

it("round-trips optional fields and per-site enablement", () => {
  const draft = draftOf({ server });
  expect(draft).toEqual({ server, project: "", runtime: undefined, grouping: "page", session: "", enabled: false });
  expect(settingsOf(draft)).toEqual({ server, grouping: "page" });
  expect(settingsOf({ ...draft, project: "acme", session: "session-1", enabled: true })).toEqual({ server, project: "acme", grouping: "page", session: "session-1", enabled: true });
});

it("tracks runtime, grouping and server edits while ignoring the enabled switch", () => {
  const saved = draftOf({ server, runtime: { spec: { model: "example-model", effort: "high" }, presets: ["fast"] } });
  expect(hasUnsavedChanges({ saved, draft: { ...saved, enabled: true } })).toBe(false);
  expect(hasUnsavedChanges({ saved, draft: { ...saved, runtime: { presets: ["fast"], spec: { effort: "high", model: "example-model" } } } })).toBe(false);
  for (const patch of [{ runtime: undefined }, { grouping: "shared" as const }, { server: "https://other.example" }]) {
    expect(hasUnsavedChanges({ saved, draft: { ...saved, ...patch } })).toBe(true);
  }
});
