import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { findDoc } from "./mdx";
import Callout from "../components/site/Callout";
import FixtureKindReference from "../components/site/FixtureKindReference";
import bundle from "../data/fixtures.schema.json";

describe("schema reference pages", () => {
  it.each(bundle.document.properties.steps.items.oneOf)("separates global and kind fields for $title", (branch) => {
    const kind = branch.properties.kind.const;
    const page = document.createElement("main");
    page.innerHTML = renderToStaticMarkup(createElement(FixtureKindReference, { kind }));
    expect(page.querySelectorAll('[role="tree"]')).toHaveLength(1);
    const sections = [...page.querySelectorAll('[role="tree"] > [role="treeitem"]')];
    expect(sections.map((section) => section.querySelector(".font-mono")?.textContent)).toEqual(["Global fields", `${kind} fields`]);
    const names = sections.map((section) => [...section.querySelectorAll(':scope > [role="group"] > [role="treeitem"]')].map((row) => row.querySelector(".font-mono")?.textContent));
    expect(names[0]?.slice(0, 11)).toEqual(["id", "name", "cel", "timeout", "wait", "retry", "traces", "export", "override", "background", "continueOnError"]);
    expect(names[1]?.[0]).toBe("kind");
    const format = [...sections[0]!.querySelectorAll(':scope > [role="group"] > [role="treeitem"]')].find(row => row.querySelector(".font-mono")?.textContent === "format");
    const shortcuts = [...format?.querySelectorAll(':scope > [role="group"] > [role="treeitem"] > div > span > .font-mono') ?? []].map(label => label.textContent);
    expect([...names.flat(), ...shortcuts].sort()).toEqual(Object.keys(branch.properties).sort());
    if (kind === "test" || kind === "lint") {
      expect(names[0]).toEqual(expect.arrayContaining(["format", "no-color", "log-level", "loglevel", "no-progress", "max-concurrent", "cwd", "db-url", "properties"]));
      expect(names[1]).not.toEqual(expect.arrayContaining(["json", "format", "no-progress"]));
      expect(shortcuts).toContain("json");
    }
  });

  it.each([
    ["configuration-reference", "gavel-config"],
    ["getting-started", "fixture-document-schema"],
    ["fixtures-overview", "fixture-document-schema"],
    ["fixture-formats", "fixture-document-schema"],
    ["fixture-runner-steps", "fixture-document-schema"],
    ["fixture-ai-review", "fixture-document-schema"],
    ["fixture-kind-services", "fixture-kind-services"],
    ["fixture-reference", "fixture-frontmatter"],
    ["fixture-step-reference", "fixture-document"],
    ["framework-go", "framework-go"],
    ["linter-eslint", "linter-eslint"],
  ])("renders %s using the installed shared component", (slug, id) => {
    const doc = findDoc(slug);
    expect(doc).toBeDefined();
    const html = renderToStaticMarkup(
      createElement(doc!.Component, { components: { Callout } }),
    );
    expect(html).toContain(`id="${id}"`);
    expect(html).toContain('role="tree"');
    expect(html).toContain("Expand all");
    if (id === "fixture-document-schema") {
      const page = document.createElement("main");
      page.innerHTML = html;
      const schema = page.querySelector(`#${id}`)!;
      expect(schema.previousElementSibling).toMatchObject({ tagName: "H2", textContent: "Schema" });
      expect(page.querySelectorAll(`#${id}`)).toHaveLength(1);
    }
  });
});
