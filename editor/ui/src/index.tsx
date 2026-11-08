/// <reference types="vite/client" />
import "./index.css";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "@flanksource/clicky-ui/hooks";
import { MonacoProvider } from "@flanksource/clicky-ui/monaco";
import EditorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";
import JSONWorker from "monaco-editor/esm/vs/language/json/json.worker?worker";
import { App } from "./App";
const root = document.getElementById("root");
if (!root) throw new Error("Editor root missing");
createRoot(root).render(
  <QueryClientProvider client={new QueryClient()}>
    <ThemeProvider>
      <MonacoProvider
        getWorker={(label) =>
          label === "json" ? new JSONWorker() : new EditorWorker()
        }
      >
        <App />
      </MonacoProvider>
    </ThemeProvider>
  </QueryClientProvider>,
);
