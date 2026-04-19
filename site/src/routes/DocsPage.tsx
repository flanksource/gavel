import { Link, useParams } from "react-router-dom";
import { findDoc, listDocs } from "@/lib/mdx";
import MdxLayout from "@/components/site/MdxLayout";
import { cn } from "@/lib/utils";

export default function DocsPage() {
  const { slug } = useParams<{ slug: string }>();
  const docs = listDocs();
  const active = slug ? findDoc(slug) : docs[0];

  return (
    <div className="mx-auto flex max-w-6xl gap-10 px-6 py-12">
      <aside className="hidden w-56 shrink-0 lg:block">
        <nav className="sticky top-20 space-y-1 text-sm">
          <p className="mb-3 text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Docs
          </p>
          {docs.map((d) => (
            <Link
              key={d.slug}
              to={`/docs/${d.slug}`}
              className={cn(
                "block rounded px-2 py-1.5 transition-colors",
                active?.slug === d.slug
                  ? "bg-muted font-medium text-foreground"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              {d.frontmatter.title}
            </Link>
          ))}
        </nav>
      </aside>

      <div className="flex-1">
        {active ? (
          <MdxLayout>
            <h1>{active.frontmatter.title}</h1>
            {active.frontmatter.description && (
              <p className="-mt-4 text-lg text-muted-foreground">{active.frontmatter.description}</p>
            )}
            <active.Component />
          </MdxLayout>
        ) : (
          <div className="py-24 text-center">
            <h1 className="text-3xl font-bold">Nothing here yet</h1>
            <p className="mt-4 text-muted-foreground">Docs are coming soon.</p>
          </div>
        )}
      </div>
    </div>
  );
}
