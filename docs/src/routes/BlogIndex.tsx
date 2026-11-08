import { usePages } from "@rspress/core/runtime";

export default function BlogIndex() {
  const { pages } = usePages();
  const posts = pages.filter((page) => page.routePath.startsWith("/blog/") && page.routePath !== "/blog/")
    .sort((a, b) => String(b.frontmatter.date).localeCompare(String(a.frontmatter.date)))
    .map((page) => ({ slug: page.routePath, frontmatter: {
      title: page.title, description: page.description,
      date: page.frontmatter.date ? String(page.frontmatter.date).slice(0, 10) : undefined,
    } }));

  return (
    <>
      {posts.length === 0 ? (
        <p className="text-muted-foreground">No posts yet — check back soon.</p>
      ) : (
        <ul className="divide-y divide-border">
          {posts.map((p) => (
            <li key={p.slug} className="py-6">
              <a href={p.slug} className="group block">
                <h2 className="text-xl font-semibold group-hover:text-brand-sky">
                  {p.frontmatter.title}
                </h2>
                {p.frontmatter.date && (
                  <p className="mt-1 text-xs uppercase tracking-wider text-muted-foreground">
                    {p.frontmatter.date}
                  </p>
                )}
                {p.frontmatter.description && (
                  <p className="mt-2 text-sm text-muted-foreground">{p.frontmatter.description}</p>
                )}
              </a>
            </li>
          ))}
        </ul>
      )}
    </>
  );
}
