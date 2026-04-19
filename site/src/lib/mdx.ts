import type { ComponentType } from "react";

export interface MdxFrontmatter {
  title: string;
  description?: string;
  date?: string;
  tags?: string[];
  order?: number;
}

export interface MdxEntry {
  slug: string;
  frontmatter: MdxFrontmatter;
  Component: ComponentType<Record<string, unknown>>;
}

interface MdxModule {
  default: ComponentType<Record<string, unknown>>;
  frontmatter?: Partial<MdxFrontmatter>;
}

function slugFromPath(path: string): string {
  return path.split("/").pop()!.replace(/\.mdx$/, "");
}

function toEntries(modules: Record<string, MdxModule>): MdxEntry[] {
  return Object.entries(modules).map(([path, mod]) => {
    const slug = slugFromPath(path);
    const fm = mod.frontmatter ?? {};
    if (!fm.title) {
      throw new Error(`MDX file ${path} is missing \`title\` in frontmatter`);
    }
    return {
      slug,
      frontmatter: fm as MdxFrontmatter,
      Component: mod.default,
    };
  });
}

const blogModules = import.meta.glob<MdxModule>("/content/blog/*.mdx", { eager: true });
const docsModules = import.meta.glob<MdxModule>("/content/docs/*.mdx", { eager: true });

export function listBlogPosts(): MdxEntry[] {
  return toEntries(blogModules).sort((a, b) => {
    const ad = a.frontmatter.date ?? "";
    const bd = b.frontmatter.date ?? "";
    return bd.localeCompare(ad);
  });
}

export function listDocs(): MdxEntry[] {
  return toEntries(docsModules).sort((a, b) => {
    const ao = a.frontmatter.order ?? 999;
    const bo = b.frontmatter.order ?? 999;
    if (ao !== bo) return ao - bo;
    return a.frontmatter.title.localeCompare(b.frontmatter.title);
  });
}

export function findBlogPost(slug: string): MdxEntry | undefined {
  return listBlogPosts().find((e) => e.slug === slug);
}

export function findDoc(slug: string): MdxEntry | undefined {
  return listDocs().find((e) => e.slug === slug);
}
