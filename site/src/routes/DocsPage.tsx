import { Link, useParams } from "react-router-dom";
import { findDoc, listDocs } from "@/lib/mdx";
import MdxLayout from "@/components/site/MdxLayout";
import { buildDocTree } from "@/lib/docs";
import DocsNavigation from "@/components/site/DocsNavigation";

export default function DocsPage() {
  const { slug } = useParams<{ slug: string }>();
  const docs = listDocs();
  const tree = buildDocTree(docs);
  const active = slug ? findDoc(slug) : tree[0]?.doc;
  const breadcrumbs = [];
  let parent = active?.frontmatter.parent;
  while (parent) {
    const ancestor = findDoc(parent)!;
    breadcrumbs.unshift(ancestor);
    parent = ancestor.frontmatter.parent;
  }

  return (
    <div className="mx-auto flex max-w-6xl gap-10 px-6 py-12">
      <aside className="hidden w-56 shrink-0 lg:block">
        <nav aria-label="Documentation" className="sticky top-20 max-h-[calc(100dvh-6rem)] space-y-1 overflow-y-auto text-sm">
          <p className="mb-3 text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Docs
          </p>
          <DocsNavigation key={active?.slug} nodes={tree} active={active?.slug} />
        </nav>
      </aside>

      <div className="min-w-0 flex-1">
        <details className="mb-6 rounded border border-border p-3 lg:hidden">
          <summary className="cursor-pointer font-medium">Documentation</summary>
          <nav aria-label="Mobile documentation" className="mt-3 max-h-80 overflow-y-auto">
            <DocsNavigation key={active?.slug} nodes={tree} active={active?.slug} />
          </nav>
        </details>
        {active ? (
          <MdxLayout>
            {breadcrumbs.length > 0 && <nav aria-label="Breadcrumb" className="mb-6 flex flex-wrap gap-2 text-sm">
              {breadcrumbs.map((ancestor) => <span key={ancestor.slug}>
                <Link to={`/docs/${ancestor.slug}`}>{ancestor.frontmatter.title}</Link><span className="ml-2 text-muted-foreground">/</span>
              </span>)}
              <span aria-current="page" className="text-muted-foreground">{active.frontmatter.title}</span>
            </nav>}
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
