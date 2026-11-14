import { describe, expect, it } from "vitest";
import { threadItems, type NativeComment } from "./file-comments";

const proposal = { ref: "refs/gavel/proposals/c-1/m-2", base: "abc123", commit: "def456", files: [{ path: "docs/a.mdx", additions: 3, deletions: 1 }], state: "proposed" as const };

const row: NativeComment = {
  id: "c-1",
  bodyMarkdown: "Tighten this paragraph",
  version: 4,
  status: "open",
  createdAt: "2026-11-01T10:00:00Z",
  author: { id: "u-1", label: "Ada", kind: "user" },
  target: { selector: { kind: "document", version: 1, view: "preview" }, revision: "r1" },
  messages: [
    { id: "m-1", bodyMarkdown: "On it", createdAt: "2026-11-01T10:01:00Z", author: { id: "claude", label: "Claude", kind: "agent" } },
    { id: "m-2", bodyMarkdown: "Here is a patch", createdAt: "2026-11-01T10:02:00Z", author: { id: "claude", label: "Claude", kind: "agent" }, content: { proposal } },
    {
      id: "m-3",
      bodyMarkdown: "Revised",
      createdAt: "2026-11-01T10:03:00Z",
      author: { id: "claude", label: "Claude", kind: "agent" },
      content: { proposal: { ...proposal, state: "rejected", decidedBy: { kind: "user", id: "u-1", label: "Ada" }, decidedAt: "2026-11-01T10:04:00Z", feedback: "Too broad" } },
    },
  ],
};

describe("threadItems", () => {
  it("carries a reply's proposed patch and converts who decided it into a comment author", () => {
    const [, plain, open, decided] = threadItems(row);
    expect(plain).not.toHaveProperty("proposal");
    expect(open).toMatchObject({ id: "m-2", parentId: "c-1", proposal });
    expect(decided?.proposal).toEqual({ ...proposal, state: "rejected", decidedBy: { id: "u-1", name: "Ada", kind: "user" }, decidedAt: "2026-11-01T10:04:00Z", feedback: "Too broad" });
  });
});
