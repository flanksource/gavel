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
import { EventSourceProvider, ThemeProvider } from "@flanksource/clicky-ui/hooks";
import { ToastProvider } from "@flanksource/clicky-ui/components";
import { ChatWindowManagerProvider } from "@flanksource/clicky-ui/ai";
import { ClerkProvider } from "@clerk/react";
import { App } from "./App";
import { type AuthConfig, loadAuthConfig } from "./authConfig";
import { AuthGate } from "./components/AuthGate";
import { setAuthMode } from "./query";
import { ChatLayer, chatLayerEnabled } from "./components/ChatLayer";
import { openEventStream } from "./eventHub";
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

const root = createRoot(document.getElementById("root")!);

// render mounts the app for the server's sign-in mode. In clerk mode the gate wraps
// App from outside, so pages App returns early for are gated too.
function render(auth: AuthConfig) {
  setAuthMode(auth.mode);
  const hosted = auth.mode === "clerk";
  const app = (
    <QueryClientProvider client={queryClient}>
      {/* Routes every clicky-ui SSE-opening hook (useTaskRun, usePrompts,
          useLogTail, useDebugStream) through the same connection-multiplexing
          hub the app's own streams use, so they share the one real
          /api/events connection instead of opening native EventSources that
          would compete for Chrome's six-per-host HTTP/1.1 connection cap. */}
      <EventSourceProvider value={openEventStream}>
        <ThemeProvider defaultTheme="system">
          {/* Bulk actions report their outcome as a toast: the results of a batch
              have nowhere to live inside DataTable's flex toolbar, and a partial
              failure needs more room than a toolbar row. */}
          <ToastProvider>
            <ChatWindowManagerProvider storageId="gavel-chat">
              {hosted ? (
                <AuthGate>
                  <App />
                </AuthGate>
              ) : (
                <>
                  <App />
                  {chatLayerEnabled(window.location.pathname) && <ChatLayer />}
                </>
              )}
            </ChatWindowManagerProvider>
          </ToastProvider>
        </ThemeProvider>
      </EventSourceProvider>
    </QueryClientProvider>
  );
  root.render(
    hosted ? (
      <ClerkProvider publishableKey={auth.publishableKey ?? ""} afterSignOutUrl="/">
        {app}
      </ClerkProvider>
    ) : (
      app
    ),
  );
}

loadAuthConfig().then(render, (error: unknown) => {
  root.render(
    <div className="flex min-h-screen items-center justify-center p-4 text-sm text-muted-foreground">
      gavel can't start: {error instanceof Error ? error.message : String(error)}
    </div>,
  );
});
