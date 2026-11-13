import { listDocs } from "../lib/mdx";

export function usePages() {
  return { pages: listDocs().map((doc) => ({
    routePath: `/docs/${doc.slug}`,
    title: doc.frontmatter.title,
    frontmatter: doc.frontmatter,
  })) };
}
