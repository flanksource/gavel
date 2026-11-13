import { describe, expect, it } from "vitest";
import { captureFromParams, capturePR } from "./model";

describe("todo capture parameters", () => {
  it("preserves the hosted form's prefilled fields and Markdown whitespace", () => {
    const body = "  indented code\n\nparagraph\n";
    const params = new URLSearchParams({ title: "Card issue", body, dir: "/work/acme", priority: "high", status: "pending", labels: "ui,bug", pr: "https://github.com/acme/repo/pull/123", sourceUrl: "https://acme.example/card" });
    params.append("label", "ui");
    expect(captureFromParams(params)).toEqual({ mode: "new", title: "Card issue", body, dir: "/work/acme", priority: "high", status: "pending", labels: ["ui", "bug"], pr: "https://github.com/acme/repo/pull/123", sourceUrl: "https://acme.example/card", ref: "", attachments: [] });
  });

  it("recognizes existing issue aliases and autosave's pending default", () => {
    expect(captureFromParams(new URLSearchParams({ issue: "t-card", description: "comment", workspace: "/work/acme", autoSave: "1" }))).toMatchObject({ mode: "existing", ref: "t-card", body: "comment", dir: "/work/acme", status: "pending" });
    expect(captureFromParams(new URLSearchParams())).toMatchObject({ mode: "new", priority: "medium", status: "draft" });
  });

  it.each([{ priority: "urgent" }, { status: "in_progress" }])("rejects invalid prefilled choices: %j", params => {
    expect(() => captureFromParams(new URLSearchParams(params))).toThrow(/Invalid/);
  });

  it.each(["https://github.com/acme/repo/pull/123#discussion", "https://dev.azure.com/acme/project/_git/repo/pullrequest/123?view=files"])("preserves PR links from the original form: %s", url => {
    expect(capturePR(url)).toEqual({ number: 123, url });
  });
});
