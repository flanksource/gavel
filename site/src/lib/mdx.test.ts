import { describe, expect, it } from "vitest";
import { findBlogPost, findDoc, listBlogPosts, listDocs } from "./mdx";

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
});
