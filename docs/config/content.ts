import { readdirSync, readFileSync } from "node:fs";
import { basename, join } from "node:path";
import matter from "gray-matter";
import type { SidebarGroup, SidebarItem } from "@rspress/core";
import { buildDocTree, type DocNode } from "../src/lib/docs";
import type { MdxFrontmatter } from "../src/lib/mdx";

export function readDocs(directory = "content/docs"): DocNode["doc"][] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) return readDocs(path);
    if (!entry.name.endsWith(".mdx") || entry.name === "index.mdx") return [];
    const { data } = matter(readFileSync(path, "utf8"));
    if (typeof data.title !== "string" || !data.title.trim()) {
      throw new Error(`MDX file ${path} is missing a title`);
    }
    return [{ slug: basename(entry.name, ".mdx"), frontmatter: data as MdxFrontmatter }];
  }).sort((a, b) => (a.frontmatter.order ?? 999) - (b.frontmatter.order ?? 999)
    || a.frontmatter.title.localeCompare(b.frontmatter.title));
}

export function buildSidebar(docs = readDocs()): (SidebarGroup | SidebarItem)[] {
  function item({ doc, children }: DocNode): SidebarGroup | SidebarItem {
    const link = { text: doc.frontmatter.title, link: `/docs/${doc.slug}` };
    return children.length
      ? { ...link, items: children.map(item), collapsible: true, collapsed: true }
      : link;
  }
  return buildDocTree(docs).map(item);
}
