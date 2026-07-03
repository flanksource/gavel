import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import mdx from "@mdx-js/rollup";
import tailwindcss from "@tailwindcss/vite";
import remarkGfm from "remark-gfm";
import remarkFrontmatter from "remark-frontmatter";
import remarkMdxFrontmatter from "remark-mdx-frontmatter";
import rehypePrettyCode from "rehype-pretty-code";
import path from "node:path";

const prettyCodeOptions = {
  theme: { light: "github-light", dark: "github-dark" },
  keepBackground: false,
  defaultLang: {
    block: "plaintext",
    inline: "plaintext",
  },
  onVisitLine(element: any) {
    if (element.children.length === 0) {
      element.children = [{ type: "text", value: " " }];
    }
  },
  onVisitHighlightedLine(element: any) {
    element.properties.className = [
      ...(element.properties.className || []),
      "highlighted-line",
    ];
  },
  onVisitHighlightedChars(element: any) {
    element.properties.className = [
      ...(element.properties.className || []),
      "highlighted-chars",
    ];
  },
  onVisitTitle(element: any) {
    element.properties.className = [
      ...(element.properties.className || []),
      "code-title",
    ];
  },
  onVisitCaption(element: any) {
    element.properties.className = [
      ...(element.properties.className || []),
      "code-caption",
    ];
  },
};

export default defineConfig({
  plugins: [
    {
      enforce: "pre",
      ...mdx({
        remarkPlugins: [
          remarkGfm,
          remarkFrontmatter,
          [remarkMdxFrontmatter, { name: "frontmatter" }],
        ],
        rehypePlugins: [
          [rehypePrettyCode, prettyCodeOptions],
        ],
        providerImportSource: "@mdx-js/react",
      }),
    },
    react(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
});
