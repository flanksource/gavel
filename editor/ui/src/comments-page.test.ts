import { describe, expect, it } from "vitest";
import {
  activityOf,
  authorOptions,
  filterThreads,
  groupThreads,
  locationOf,
  threadEntries,
  type CommentProject,
  type ListedComment,
} from "./comments-page";
import { NO_COMMENTS_FILTERS } from "./routes";

const demo: CommentProject = {
  scope: "gavel://project/demo",
  project: "demo",
  dir: "/work/demo",
  files: [{ id: "file-readme", uri: "gavel://project/demo/files/README.md" }],
};
const docs: CommentProject = { scope: "gavel://project/docs", project: "docs", dir: "/work/docs", files: [] };

const alice = { id: "user:alice", kind: "user" as const, label: "Alice" };
const agent = { id: "agent:reviewer", kind: "agent" as const, label: "Reviewer" };

function thread(id: string, overrides: Partial<ListedComment> & { uri: string; at: string }): ListedComment {
  const { uri, at, ...rest } = overrides;
  return {
    id,
    bodyMarkdown: `Body of ${id}`,
    version: 1,
    status: "open",
    phase: "active",
    unread: false,
    createdAt: at,
    updatedAt: at,
    author: alice,
    target: { uri, scopeUri: uri.startsWith("gavel://project/docs") ? docs.scope : demo.scope, kind: "document", revision: "r1", selector: { kind: "document", version: 1, view: "preview" } },
    ...rest,
  };
}

const readmeOld = thread("readme-old", { uri: "gavel://project/demo/files/README.md", at: "2026-11-01T10:00:00Z" });
const readmeReplied = thread("readme-replied", {
  uri: "gavel://project/demo/files/README.md",
  at: "2026-11-02T10:00:00Z",
  messages: [{ id: "reply-1", bodyMarkdown: "Fixed the typo", createdAt: "2026-11-05T09:00:00Z", author: agent }],
});
const guide = thread("guide", {
  uri: "gavel://project/docs/files/guide/intro.md",
  at: "2026-11-04T10:00:00Z",
  author: agent,
  status: "resolved",
  phase: "awaiting_outcome",
  unread: true,
  target: {
    uri: "gavel://project/docs/files/guide/intro.md",
    scopeUri: docs.scope,
    kind: "code-range",
    revision: "r1",
    selector: { kind: "code-range", version: 1, view: "source", patch: { version: 1, startLine: 12, endLine: 14, before: [], lines: ["a", "b", "c"], after: [] } },
  },
});
const page = thread("page", { uri: "https://example.com/pricing", at: "2026-11-03T10:00:00Z" });

const entries = threadEntries([
  { project: demo, rows: [readmeOld, readmeReplied, page] },
  { project: docs, rows: [guide] },
]);

describe("locationOf", () => {
  it("names an edited project file by its project path and the edited file it is", () => {
    expect(locationOf(readmeOld, demo)).toEqual({ kind: "file", key: readmeOld.target.uri, label: "README.md", fileId: "file-readme", view: "preview" });
  });

  it("places a code-range thread on a file that is not edited at its first line, with no file to open", () => {
    expect(locationOf(guide, docs)).toEqual({ kind: "file", key: guide.target.uri, label: "guide/intro.md", line: 12, view: "source" });
  });

  it("names a project web-page thread by the page URL it was left on", () => {
    const pageUri = "gavel://project/demo/pages/http%3A%2F%2F127.0.0.1%3A4000%2Fpricing";
    expect(locationOf(thread("web", { uri: pageUri, at: "2026-11-01T00:00:00Z" }), demo)).toEqual({ kind: "page", key: pageUri, label: "http://127.0.0.1:4000/pricing" });
  });

  it("names any other target by its URI", () => {
    expect(locationOf(page, demo)).toEqual({ kind: "page", key: "https://example.com/pricing", label: "https://example.com/pricing" });
  });
});

describe("activityOf", () => {
  it("is the newest reply when one is later than the thread itself", () => {
    expect(activityOf(readmeReplied)).toBe("2026-11-05T09:00:00Z");
  });

  it("is the last update of a thread nobody replied to", () => {
    expect(activityOf({ ...readmeOld, updatedAt: "2026-11-06T00:00:00Z" })).toBe("2026-11-06T00:00:00Z");
  });
});

describe("groupThreads", () => {
  it("groups threads by target, groups and their threads newest activity first", () => {
    expect(groupThreads(entries).map((group) => [group.location.label, group.threads.map((entry) => entry.row.id)])).toEqual([
      ["README.md", ["readme-replied", "readme-old"]],
      ["guide/intro.md", ["guide"]],
      ["https://example.com/pricing", ["page"]],
    ]);
  });

  it("keeps the same path in two projects as two groups", () => {
    const twin = thread("twin", { uri: "gavel://project/docs/files/README.md", at: "2026-11-01T00:00:00Z" });
    const groups = groupThreads(threadEntries([{ project: demo, rows: [readmeOld] }, { project: docs, rows: [twin] }]));
    expect(groups.map((group) => `${group.project.project}:${group.location.label}`)).toEqual(["demo:README.md", "docs:README.md"]);
  });
});

describe("filterThreads", () => {
  const ids = (filters: Partial<typeof NO_COMMENTS_FILTERS>) =>
    filterThreads(entries, { ...NO_COMMENTS_FILTERS, ...filters }).map((entry) => entry.row.id).sort();

  it.each([
    ["no filter", {}, ["guide", "page", "readme-old", "readme-replied"]],
    ["a status", { status: "resolved" }, ["guide"]],
    ["a phase", { phase: "active" }, ["page", "readme-old", "readme-replied"]],
    ["an author", { author: agent.id }, ["guide"]],
    ["a project", { project: "demo" }, ["page", "readme-old", "readme-replied"]],
    ["unread only", { unread: true }, ["guide"]],
    ["text in a reply, case-insensitively", { q: "TYPO" }, ["readme-replied"]],
    ["text in the target path", { q: "intro.md" }, ["guide"]],
    ["filters that exclude each other", { status: "resolved", project: "demo" }, []],
  ])("narrows to %s", (_, filters, expected) => {
    expect(ids(filters)).toEqual(expected);
  });
});

describe("authorOptions", () => {
  it("lists each thread author once, by label", () => {
    expect(authorOptions(entries)).toEqual([
      { value: "user:alice", label: "Alice" },
      { value: "agent:reviewer", label: "Reviewer" },
    ]);
  });
});
