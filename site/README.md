# Gavel marketing site

Vite + React + TypeScript + Tailwind v4 marketing site for [gavel](https://github.com/flanksource/gavel), deployed to `gavel.flanksource.com` on Vercel.

Independent of the Preact UIs in `pr/ui/` and `testrunner/ui/` — this package is not embedded in the Go binary.

## Develop

```bash
cd site
npm install
npm run dev          # http://localhost:5173
npm run build        # type-check + production build into dist/
npm run preview      # serve the built bundle
npm run test         # vitest unit tests
npm run test:e2e     # playwright end-to-end
```

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
