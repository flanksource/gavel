# Gavel documentation and website

This directory contains the Rspress application for the landing page, blog, and documentation. See [CONTRIBUTING.md](../CONTRIBUTING.md#documentation-application) for development, schema generation, environment configuration, and deployment. Node.js 22.12 or newer is required.

```bash
cd docs
npm ci
npm test
npm run dev
```

Public guides live in `content/docs/`; reusable examples and migration bundles are served from `content/public/examples/services/`. `rspress.config.ts` configures the site, search, and navigation. Sidebar groups come from frontmatter `parent` and `order`; service guides retain their filename-based `/docs/service-*` URLs. Rspress builds static HTML into `dist/`, including a search index and Markdown downloads for AI tools.

Run `make test`, `make lint`, and `make build` from this directory to verify the application. Use `npm run preview` to serve the production build at `http://localhost:5190`.

Fixture example tests require the `gavel` CLI from this checkout on `PATH`. They parse both example tabs with `gavel fixtures outline` and check that they contain the same number of runnable checks; they do not execute the examples.

Use `src/components/site/FixtureExample.tsx` around YAML and Markdown code blocks, in that order. It adds YAML-first source tabs and a rendered Markdown preview. Place `FixtureDocumentReference` under a standalone `## Schema` heading once per page, or use the page's existing schema reference section. Use `src/components/site/Prerequisites.tsx` around Markdown paragraphs or lists to share the prerequisites callout used by fixture and service guides.

## Deployment handoff

Use `docs` as the Vercel project's **Root Directory**. `vercel.json` builds with Rspress and serves clean static URLs from `dist/`; `/api/waitlist` remains a Vercel function. Keep the existing environment variables and domain. No deployment settings are changed by local development.
