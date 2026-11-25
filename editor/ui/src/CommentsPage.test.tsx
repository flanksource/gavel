// @vitest-environment jsdom
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, describe, expect, it } from "vitest";
import { threadCallbacks } from "./comment-actions";
import { CommentsView, type OpenThread } from "./CommentsPage";
import type { CommentProject, ListedComment } from "./comments-page";
import type { ProjectListing } from "./project-comments";
import { NO_COMMENTS_FILTERS, type CommentsFilters } from "./routes";

(globalThis as { IS_REACT_ACT_ENVIRONMENT?: boolean }).IS_REACT_ACT_ENVIRONMENT = true;

const demo: CommentProject = {
  scope: "gavel://project/demo",
  project: "demo",
  dir: "/work/demo",
  files: [{ id: "file-readme", uri: "gavel://project/demo/files/README.md" }],
};

function thread(id: string, body: string, uri: string, at: string, extra: Partial<ListedComment> = {}): ListedComment {
  return {
    id,
    bodyMarkdown: body,
    version: 1,
    status: "open",
    phase: "active",
    unread: false,
    createdAt: at,
    updatedAt: at,
    author: { id: "user:alice", kind: "user", label: "Alice" },
    target: { uri, scopeUri: demo.scope, kind: "document", revision: "r1", selector: { kind: "document", version: 1, view: "preview" } },
    ...extra,
  };
}

const readmeBody = "Tighten the README intro";
const guideBody = "Guide link is broken";
const listings: ProjectListing[] = [
  {
    project: demo,
    rows: [
      thread("t-readme", readmeBody, "gavel://project/demo/files/README.md", "2026-11-05T10:00:00Z", { unread: true }),
      thread("t-guide", guideBody, "gavel://project/demo/files/docs/guide.md", "2026-11-01T10:00:00Z", { status: "resolved", phase: "awaiting_outcome" }),
    ],
  },
];

let root: Root | undefined;
let host: HTMLElement;
afterEach(() => {
  act(() => root?.unmount());
  host?.remove();
  root = undefined;
});

function render(props: { listings?: ProjectListing[]; error?: string; filters?: CommentsFilters; onOpen?: (thread: OpenThread) => void }) {
  host = document.body.appendChild(document.createElement("div"));
  root = createRoot(host);
  act(() =>
    root!.render(
      <CommentsView
        listings={props.listings}
        error={props.error}
        filters={props.filters ?? NO_COMMENTS_FILTERS}
        onFiltersChange={() => {}}
        callbacks={threadCallbacks([], async () => {})}
        onOpen={props.onOpen ?? (() => {})}
      />,
    ),
  );
  return host;
}

describe("CommentsView", () => {
  it("renders threads grouped by file, newest activity first, each group naming its project and count", () => {
    const text = render({ listings }).textContent ?? "";
    expect(text.indexOf(readmeBody)).toBeGreaterThan(-1);
    expect(text.indexOf(readmeBody)).toBeLessThan(text.indexOf(guideBody));
    expect(text.match(/demo · 1 thread/g)).toHaveLength(2);
  });

  it("opens an edited file's thread from its place, and only names a file that is not edited", () => {
    const opened: OpenThread[] = [];
    const page = render({ listings, onOpen: (thread) => opened.push(thread) });
    const link = page.querySelector<HTMLButtonElement>('button[title="Open README.md"]');
    const notOpen = [...page.querySelectorAll("span.font-mono")].find((span) => span.textContent === "docs/guide.md");
    act(() => link!.click());
    expect(opened).toEqual([{ kind: "file", key: "gavel://project/demo/files/README.md", label: "README.md", fileId: "file-readme", view: "preview" }]);
    expect(notOpen?.getAttribute("title")).toBe("Not open in this session: run gavel edit on this file to open it");
    expect(page.querySelector('button[title="Open docs/guide.md"]')).toBeNull();
  });

  it("shows only the threads the route's filters match", () => {
    const text = render({ listings, filters: { ...NO_COMMENTS_FILTERS, status: "resolved" } }).textContent ?? "";
    expect(text).toContain(guideBody);
    expect(text).not.toContain(readmeBody);
  });

  it.each([
    ["no thread matching the filters", { listings, filters: { ...NO_COMMENTS_FILTERS, q: "no such text" } }, "No threads match these filters"],
    ["loading", {}, "Loading comments…"],
    ["failed", { error: "boom" }, "Comments could not be loaded: boom"],
    ["empty", { listings: [{ project: demo, rows: [] }] }, "No comment threads yet"],
    ["outside every project", { listings: [] }, "No project holds the edited files"],
  ])("keeps the filter bar and tells a page %s apart", (_, props, message) => {
    const page = render(props);
    expect(page.querySelector('[aria-label="Search comments"]')).not.toBeNull();
    expect(page.textContent).toContain(message);
  });
});
