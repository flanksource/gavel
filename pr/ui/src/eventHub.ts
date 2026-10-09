import { createEventHub } from '@flanksource/clicky-ui/hooks';

// Chrome allows only six HTTP/1.1 connections per host, shared across every
// tab, so every stream this tab opens is multiplexed onto one connection to
// the server's /api/events hub (clicky's sse.Hub, mounted in pr/ui/handler.go).
// openEventStream(url) behaves like a native EventSource for url to its callers.

// The build id this page was served with, read once at module init from the
// `<meta name="gavel-ui-build">` tag the Go server stamps into every served
// SPA HTML page (pr/ui/ui_build.go). Absent in tests, Storybook, and a bare
// `vite dev` server with no Go backend in front of it — those never reload.
const pageBuildId = typeof document !== 'undefined'
  ? document.querySelector<HTMLMetaElement>('meta[name="gavel-ui-build"]')?.content || null
  : null;

export const openEventStream = createEventHub({ buildId: pageBuildId });
