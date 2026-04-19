import { Link } from "react-router-dom";
import { listBlogPosts } from "@/lib/mdx";

export default function BlogIndex() {
  const posts = listBlogPosts();

  return (
    <div className="mx-auto max-w-3xl px-6 py-16">
      <header className="mb-10">
        <p className="text-sm font-medium uppercase tracking-wider text-brand-sky">Blog</p>
        <h1 className="mt-2 text-4xl font-bold tracking-tight">Notes from the gavel</h1>
      </header>

      {posts.length === 0 ? (
        <p className="text-muted-foreground">No posts yet — check back soon.</p>
      ) : (
        <ul className="divide-y divide-border">
          {posts.map((p) => (
            <li key={p.slug} className="py-6">
              <Link to={`/blog/${p.slug}`} className="group block">
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
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
