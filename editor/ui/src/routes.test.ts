import { describe, expect, it } from "vitest";
import { commentsRouteFrom, NO_COMMENTS_FILTERS, taskRunFromHref, writeCommentsRoute, type CommentsFilters } from "./routes";

describe("Comments page route", () => {
  const encode = (filters: CommentsFilters) => {
    const params = new URLSearchParams();
    writeCommentsRoute(params, filters);
    return params.toString();
  };
  const everyFilter: CommentsFilters = { status: "open", phase: "active", author: "agent:claude", project: "demo app", unread: true, q: "fix README" };

  it("is no Comments page without the comments parameter", () => {
    expect(commentsRouteFrom(new URLSearchParams("file=a.md&status=open"))).toBeUndefined();
  });

  it("decodes a bare ?comments as the unfiltered page", () => {
    expect(commentsRouteFrom(new URLSearchParams("comments"))).toEqual(NO_COMMENTS_FILTERS);
  });

  it("encodes the unfiltered page as only the comments parameter", () => {
    expect(encode(NO_COMMENTS_FILTERS)).toBe("comments=");
  });

  it("encodes every filter and decodes it back unchanged", () => {
    const query = encode(everyFilter);
    expect(query).toBe("comments=&status=open&phase=active&author=agent%3Aclaude&project=demo+app&q=fix+README&unread=1");
    expect(commentsRouteFrom(new URLSearchParams(query))).toEqual(everyFilter);
  });

  it("reads unread only from unread=1", () => {
    expect(commentsRouteFrom(new URLSearchParams("comments&unread=true"))?.unread).toBe(false);
  });
});

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
