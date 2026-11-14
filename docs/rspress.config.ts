import { readdirSync } from "node:fs";
import { resolve, basename } from "node:path";
import { defineConfig } from "@rspress/core";
import tailwindcss from "@tailwindcss/postcss";
import { buildSidebar } from "./config/content";
import { codeAnnotationTransformer } from "./config/code-annotations";

export default defineConfig({
  root: "content",
  outDir: "dist",
  title: "Gavel",
  siteOrigin: "https://gavel.flanksource.com",
  description: "One CLI for tests, linting, and AI code review.",
  icon: "/favicon.svg",
  logo: "/favicon.svg",
  logoText: "Gavel",
  globalStyles: resolve("src/styles/globals.css"),
  globalUIComponents: process.env.VERCEL === "1" ? [resolve("src/components/site/Analytics.tsx")] : [],
  themeConfig: {
    nav: [
      { text: "Docs", link: "/docs/getting-started", activeMatch: "^/docs" },
      { text: "Blog", link: "/blog/", activeMatch: "^/blog" },
    ],
    sidebar: { "/docs/": buildSidebar() },
    socialLinks: [{ icon: "github", mode: "link", content: "https://github.com/flanksource/gavel" }],
    footer: { message: "© Flanksource · Gavel is MIT licensed." },
    enableScrollToTop: true,
  },
  markdown: {
    showLineNumbers: true,
    link: { checkDeadLinks: true },
    shiki: { transformers: [codeAnnotationTransformer()] },
  },
  route: { cleanUrls: true, exclude: ["docs/services/**"] },
  plugins: [{
    name: "gavel-documentation-routes",
    addPages() {
      return readdirSync("content/docs/services").filter((file) => file.endsWith(".mdx"))
        .map((file) => ({ routePath: `/docs/${basename(file, ".mdx")}`,
          filepath: resolve("content/docs/services", file) }));
    },
  }],
  builderConfig: {
    resolve: { alias: { "@": resolve("src") } },
    server: { port: 5190 },
    tools: {
      postcss: { postcssOptions: { plugins: [tailwindcss()] } },
      rspack(config, { addRules }) {
        addRules({ resourceQuery: /url/, type: "asset/resource" });
        return config;
      },
    },
  },
});
