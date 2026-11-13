import { resolve } from "node:path";
import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import mdx from "@mdx-js/rollup";
import remarkGfm from "remark-gfm";
import remarkFrontmatter from "remark-frontmatter";
import remarkMdxFrontmatter from "remark-mdx-frontmatter";
import rehypeSlug from "rehype-slug";

export default defineConfig({
    server: { fs: { allow: [resolve("..")] } },
    plugins: [
      { enforce: "pre", ...mdx({
        remarkPlugins: [remarkGfm, remarkFrontmatter, [remarkMdxFrontmatter, { name: "frontmatter" }]],
        rehypePlugins: [rehypeSlug],
        providerImportSource: "@mdx-js/react",
      }) },
      react(),
    ],
    resolve: { alias: {
      "@": resolve("src"),
      "@rspress/core/runtime": resolve("src/test/rspress-runtime.ts"),
      "@rspress/core/theme": resolve("src/test/rspress-theme.tsx"),
    } },
    test: {
      environment: "jsdom",
      globals: true,
      setupFiles: ["./src/test-setup.ts"],
    },
});
