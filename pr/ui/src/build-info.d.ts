// Build-time constants injected by Vite `define` (see vite.config.ts) and the
// backend build metadata injected by the Go server as window.__GAVEL__ (see
// pr/ui/handler.go). Declared so tsc --noEmit resolves them.

declare const __GAVEL_UI_VERSION__: string;
declare const __GAVEL_UI_COMMIT__: string;
declare const __GAVEL_UI_DATE__: string;

interface Window {
  __GAVEL__?: {
    version?: string;
    commit?: string;
    date?: string;
  };
}
