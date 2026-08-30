// react-grab is a dev-only DOM/React instrumentation overlay. Gate it on DEV so
// it loads under the `vite` dev server but is statically dead (and tree-shaken
// out) of every `vite build` bundle — the embedded production bundle served to
// the menubar webview and `gavel pr list --ui`. The previous `!import.meta.env.CI`
// gate was always true in builds (Vite never exposes CI), so the overlay shipped
// to production and ran under the app's constant re-render loop, growing memory
// without bound (the webview ballooned to ~26GB; the /processes page OOM-crashed).
if (import.meta.env.DEV) {
  import("react-grab");
  // Load gavel's React Grab plugin (the same artifact gavel serves for injection
  // into other apps) so the "Add to gavel todo" action is available in dev. It
  // polls for window.__REACT_GRAB__, which the import above sets.
  document.body.appendChild(
    Object.assign(document.createElement("script"), { src: "/react-grab-plugin.js" }),
  );
}

import "./index.css";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "@flanksource/clicky-ui/hooks";
import { ToastProvider } from "@flanksource/clicky-ui/components";
import { App } from "./App";
import { registerIconifyFallback } from "./icons/iconifyFallback";

// Resolves user-authored Iconify names on tag definitions. Every glyph the app
// itself draws is still imported offline; without this, clicky-ui's Icon renders
// a dashed placeholder for any runtime name. Registered before the first render
// because it sets module-global state.
registerIconifyFallback();

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
      refetchOnWindowFocus: false,
    },
  },
});

createRoot(document.getElementById("root")!).render(
  <QueryClientProvider client={queryClient}>
    <ThemeProvider defaultTheme="system">
      {/* Bulk actions report their outcome as a toast: the results of a batch
          have nowhere to live inside DataTable's flex toolbar, and a partial
          failure needs more room than a toolbar row. */}
      <ToastProvider>
        <App />
      </ToastProvider>
    </ThemeProvider>
  </QueryClientProvider>,
);
