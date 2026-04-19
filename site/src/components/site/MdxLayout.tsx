import type { ReactNode } from "react";
import { MDXProvider } from "@mdx-js/react";
import Callout from "./Callout";

const components = {
  Callout,
};

export default function MdxLayout({ children }: { children: ReactNode }) {
  return (
    <MDXProvider components={components}>
      <article
        className="mx-auto max-w-3xl px-6 py-16
          prose-headings:font-semibold prose-headings:tracking-tight
          [&_h1]:text-4xl [&_h1]:mt-2 [&_h1]:mb-6
          [&_h2]:text-2xl [&_h2]:mt-12 [&_h2]:mb-4
          [&_h3]:text-xl [&_h3]:mt-8 [&_h3]:mb-3
          [&_p]:leading-7 [&_p]:text-foreground/90 [&_p]:my-4
          [&_a]:text-brand-sky [&_a]:underline-offset-4 hover:[&_a]:underline
          [&_ul]:my-4 [&_ul]:list-disc [&_ul]:pl-6 [&_ol]:my-4 [&_ol]:list-decimal [&_ol]:pl-6
          [&_li]:my-1
          [&_code:not(pre_code)]:rounded [&_code:not(pre_code)]:bg-muted [&_code:not(pre_code)]:px-1.5 [&_code:not(pre_code)]:py-0.5 [&_code:not(pre_code)]:text-[0.9em] [&_code:not(pre_code)]:font-mono
          [&_pre]:my-6 [&_pre]:overflow-x-auto [&_pre]:rounded-lg [&_pre]:border [&_pre]:border-border [&_pre]:bg-muted/40 [&_pre]:p-4 [&_pre]:text-sm
          [&_blockquote]:my-6 [&_blockquote]:border-l-2 [&_blockquote]:border-brand-sky [&_blockquote]:pl-4 [&_blockquote]:text-muted-foreground"
      >
        {children}
      </article>
    </MDXProvider>
  );
}
