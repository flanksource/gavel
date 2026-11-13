import { usePages } from "@rspress/core/runtime";

export default function DocChildren({ parent }: { parent?: string }) {
  const { pages } = usePages();
  const children = pages.filter((page) => page.routePath.startsWith("/docs/")
    && page.routePath !== "/docs/" && page.frontmatter.parent === parent)
    .sort((a, b) => Number(a.frontmatter.order ?? 999) - Number(b.frontmatter.order ?? 999)
      || a.title.localeCompare(b.title));
  return <ul>
    {children.map((page) => <li key={page.routePath}>
      <a href={page.routePath}>{page.title}</a>
      {page.frontmatter.description && <span className="ml-2 text-sm text-muted-foreground">{page.frontmatter.description}</span>}
    </li>)}
  </ul>;
}
