import { describe, expect, it } from "vitest";
import { taskRunFromHref } from "./routes";

describe("taskRunFromHref", () => {
  it.each([
    ["the tasks page", "/tasks", ""],
    ["one run", "/tasks/run-123", "run-123"],
    ["an encoded run id", "/tasks/lint%2Fgroup", "lint/group"],
  ])("maps %s to the run it selects", (_, href, run) => {
    expect(taskRunFromHref(href)).toBe(run);
  });

  it("rejects a link the editor has no page for", () => {
    expect(() => taskRunFromHref("/todos/1")).toThrow(/no editor page for task link "\/todos\/1"/);
  });
});
