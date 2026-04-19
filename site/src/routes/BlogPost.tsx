import { Link, useParams } from "react-router-dom";
import { findBlogPost } from "@/lib/mdx";
import MdxLayout from "@/components/site/MdxLayout";

export default function BlogPost() {
  const { slug } = useParams<{ slug: string }>();
  const post = slug ? findBlogPost(slug) : undefined;

  if (!post) {
    return (
      <div className="mx-auto max-w-2xl px-6 py-24 text-center">
        <h1 className="text-3xl font-bold">Post not found</h1>
        <Link to="/blog" className="mt-6 inline-block text-brand-sky hover:underline">
          ← All posts
        </Link>
      </div>
    );
  }

  const { Component, frontmatter } = post;

  return (
    <MdxLayout>
      <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
        {frontmatter.date}
      </p>
      <h1>{frontmatter.title}</h1>
      {frontmatter.description && (
        <p className="-mt-4 text-lg text-muted-foreground">{frontmatter.description}</p>
      )}
      <Component />
    </MdxLayout>
  );
}
