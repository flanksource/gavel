import { describe, expect, it } from "vitest";
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { buildDocTree } from "./docs";
import { findDoc, listDocs } from "./mdx";

describe("consolidated documentation", () => {
  it("publishes worked service guides beneath Services", () => {
    for (const service of ["dbtest", "deps", "exec", "git"]) {
      expect(findDoc(`service-${service}`)?.frontmatter.parent).toBe("fixture-kind-services");
    }
    expect(() => buildDocTree(listDocs())).not.toThrow();
  });

  it("publishes the existing operational guides", () => {
    for (const slug of ["todo-run-preflight", "pr-todos", "todos-triage-new", "azure-devops-pr-support", "postgres-performance", "github-action", "commits", "pull-requests", "git-analysis", "projects", "infrastructure", "utilities"]) {
      expect(findDoc(slug), slug).toBeDefined();
    }
  });

  it("ships every service fixture and its migration bundle as downloadable files", () => {
    for (const name of ["dbtest", "dbtest-migrations", "dbtest-multiple", "dbtest-tests", "deps-postgres", "deps-valkey", "exec", "git"]) {
      expect(existsSync(resolve(`content/public/examples/services/${name}.fixture.yaml`)), name).toBe(true);
    }
    expect(readFileSync(resolve("content/public/examples/services/schema/instance.sql"), "utf8")).toContain("runs: always");
  });

  it("renders stable anchors for service guide links", () => {
    const doc = findDoc("service-dbtest");
    expect(doc).toBeDefined();
    expect(renderToStaticMarkup(createElement(doc!.Component))).toContain('id="automatic-connections"');
  });

  it("provides CLI references, contributor and agent entry points without the legacy site", () => {
    for (const path of ["package.json", "../MANUAL.md", "../SCHEMA.md", "../CONTRIBUTING.md", "../AGENTS.md", "../gavel.schema.json", "content/public/diagrams/SessionDataModel.html", "diagrams/SessionDataModel.tsx"]) {
      expect(existsSync(resolve(path)), path).toBe(true);
    }
    for (const path of ["../site", "services/README.md"]) {
      expect(existsSync(resolve(path)), path).toBe(false);
    }
  });
});
