import { describe, expect, it } from "vitest";
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { findBlogPost, findDoc, listBlogPosts, listDocs } from "./mdx";
import { buildDocTree } from "./docs";
import fixtureSchema from "@/data/fixtures.schema.json";

describe("mdx loader", () => {
  it("lists blog posts sorted newest-first", () => {
    const posts = listBlogPosts();
    expect(posts.length).toBeGreaterThan(0);
    for (let i = 1; i < posts.length; i += 1) {
      const prev = posts[i - 1]!.frontmatter.date ?? "";
      const cur = posts[i]!.frontmatter.date ?? "";
      expect(prev >= cur).toBe(true);
    }
  });

  it("finds the launch blog post by slug", () => {
    const post = findBlogPost("why-we-built-gavel");
    expect(post).toBeDefined();
    expect(post?.frontmatter.title).toMatch(/gavel/i);
  });

  it("lists docs ordered by frontmatter order then title", () => {
    const docs = listDocs();
    expect(docs.length).toBeGreaterThan(0);
    expect(findDoc("getting-started")?.frontmatter.order).toBe(1);
  });

  it("returns undefined for missing slugs", () => {
    expect(findBlogPost("does-not-exist")).toBeUndefined();
    expect(findDoc("does-not-exist")).toBeUndefined();
  });

  it("resolves every internal documentation link and media asset", () => {
    for (const doc of listDocs()) {
      const path = resolve("content/docs", `${doc.slug}.mdx`);
      const content = readFileSync(path, "utf8");
      for (const match of content.matchAll(/\/docs\/([a-z0-9-]+)/g)) {
        expect(findDoc(match[1]!), `${path} links to ${match[1]}`).toBeDefined();
      }
      for (const match of content.matchAll(/\/media\/([a-z0-9-]+\.(?:png|webm))/g)) {
        expect(existsSync(resolve("public/media", match[1]!)), `${path} links to ${match[1]}`).toBe(true);
      }
    }
  });

  it("keeps Getting started as the documentation landing page", () => {
    expect(buildDocTree(listDocs())[0]?.doc.slug).toBe("getting-started");
  });

  it("provides dedicated framework and linter pages under their guides", () => {
    for (const framework of ["go", "ginkgo", "jest", "vitest", "playwright", "fixture"]) {
      expect(findDoc(`framework-${framework}`)?.frontmatter.parent).toBe("test-execution");
    }
    for (const linter of ["golangci-lint", "ruff", "eslint", "oxlint", "react-doctor", "pyright", "tsc", "markdownlint", "vale", "jscpd", "betterleaks"]) {
      expect(findDoc(`linter-${linter}`)?.frontmatter.parent).toBe("lint-execution");
    }
  });

  it("provides a dedicated page for every registered fixture kind", () => {
    for (const kind of fixtureSchema.document.properties.steps.items.oneOf) {
      const slug = kind.properties.kind.const.replaceAll("_", "-");
      expect(findDoc(`fixture-kind-${slug}`)?.frontmatter.parent).toBe("fixture-step-reference");
    }
  });
});
