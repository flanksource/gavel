# Gavel marketing site

Vite + React + TypeScript + Tailwind v4 marketing site for [gavel](https://github.com/flanksource/gavel), deployed to `gavel.flanksource.com` on Vercel.

Independent of the Preact UIs in `pr/ui/` and `testrunner/ui/` — this package is not embedded in the Go binary.

## Develop

```bash
cd site
npm install
npm run dev          # http://localhost:5190
npm run build        # type-check + production build into dist/
npm run preview      # serve the built bundle
npm run test         # vitest unit tests
npm run test:e2e     # playwright end-to-end
```

## Documentation references

Documentation lives in `content/docs/*.mdx`; frontmatter `parent` creates the sidebar hierarchy and `order` sorts siblings. Parents must name an existing doc slug; missing parents, duplicates, and cycles fail validation. Parent pages list their children from the same metadata. Dedicated pages cover every framework, linter, and registered fixture kind. The configuration reference imports the canonical `../gavel.schema.json`, and fixture references use `src/data/fixtures.schema.json`; service-entry definitions come from the native `../fixtures/kinds/services/schema.json`. Vite emits schema downloads from the same imports. `SchemaReference` from `@flanksource/clicky-ui/data` renders expandable objects, array items, maps, variants, constraints, and referenced types.

When fixture kinds or command options change, regenerate the fixture bundle from the checked-out source. From the repository root:

```bash
go run ./cmd/gavel fixtures --schema > site/src/data/fixtures.schema.json
```

Regenerate `gavel.schema.json` using the repository's existing `go generate .` workflow when its configuration schema changes. Do not hand-edit either generated JSON artifact.

The downloadable Markdown, YAML, and JSON smoke fixtures under `public/examples/` are runnable without a database, AI runtime, or network service. From the repository root:

```bash
gavel fixtures site/public/examples/smoke.fixture.md site/public/examples/smoke.fixture.yaml site/public/examples/smoke.fixture.json
```

From `site/`, run `npm run test` and `npm run build` to validate the documentation loader, schema rendering, TypeScript, and MDX compilation. Use `agent-browser` against the local development or preview server to check navigation, schema links/downloads, and table layout.

Documentation screenshots and screencasts live in `public/media/`. Capture browser UI with `agent-browser`, use a replayed run for stable examples, and inspect the output for machine-specific paths or credentials before linking it from MDX.

## Environment variables

Create `site/.env.local` (git-ignored) for local development, and set the same keys in Vercel project settings for deploys:

| Variable             | Required | Purpose                                              |
|----------------------|----------|------------------------------------------------------|
| `RESEND_API_KEY`     | yes      | Resend API key — waitlist confirmation email + audience subscribe |
| `RESEND_AUDIENCE_ID` | yes      | Resend audience (mailing list) id for waitlist       |
| `KV_REST_API_URL`    | yes      | Vercel KV URL — waitlist source-of-truth             |
| `KV_REST_API_TOKEN`  | yes      | Vercel KV read/write token                           |

Client code never reads these — they are accessed only from `api/waitlist.ts`.

## Deploy

1. Create a Vercel project; set **Root Directory** to `site`.
2. Add the env vars above under Project Settings → Environment Variables.
3. Point `gavel.flanksource.com` at the Vercel project via DNS CNAME.
4. Push to the default branch — Vercel builds and deploys automatically.

## Layout

```
src/
  routes/              # top-level pages
  components/site/     # marketing-specific sections (Hero, FeatureGrid, ...)
  components/ui/       # shadcn primitives (generated)
  magic/               # MagicUI components (copy-pasted, not an npm dep)
  lib/                 # useTheme, mdx loader, utils
  styles/globals.css   # Tailwind v4 + CSS vars
content/
  blog/*.mdx           # blog posts
  docs/*.mdx           # documentation
api/
  waitlist.ts          # Vercel serverless function
```
