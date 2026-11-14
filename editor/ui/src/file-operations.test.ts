import { describe, expect, it } from "vitest";
import { joinPath, retargetName, templateGroups, type Template } from "./file-operations";

const template = (id: string, group: string, name: string, suffix: string): Template => ({
  id,
  group,
  label: id,
  description: "",
  name,
  suffix,
  source: "",
});
const markdown = template("markdown", "Markdown", "untitled.md", ".md");
const fixture = template("fixture", "Markdown", "untitled.fixture.md", ".fixture.md");
const page = template("page", "Facet", "page.tsx", ".tsx");

describe("retargetName", () => {
  it("replaces the previous suggestion with the new template's", () => {
    expect(retargetName("untitled.md", markdown, page)).toBe("page.tsx");
  });
  it("keeps a typed stem under the new template's suffix", () => {
    expect(retargetName("checkout.md", markdown, fixture)).toBe("checkout.fixture.md");
    expect(retargetName("checkout.fixture.md", fixture, page)).toBe("checkout.tsx");
  });
  it("suggests the template's name when nothing was chosen before or the name lost its suffix", () => {
    expect(retargetName("", undefined, markdown)).toBe("untitled.md");
    expect(retargetName("notes.txt", markdown, page)).toBe("page.tsx");
  });
});

describe("joinPath", () => {
  it("places a name under a folder, or at the root for an empty folder", () => {
    expect(joinPath("docs/guides/", "intro.md")).toBe("docs/guides/intro.md");
    expect(joinPath("", "intro.md")).toBe("intro.md");
  });
});

describe("templateGroups", () => {
  it("groups templates in catalog order", () => {
    expect(templateGroups([markdown, page, fixture])).toEqual([
      ["Markdown", [markdown, fixture]],
      ["Facet", [page]],
    ]);
  });
});
