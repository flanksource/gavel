import { createRoot } from "react-dom/client";
import { DensityProvider, ErrorWrapper, ThemeProvider } from "@flanksource/clicky-ui";
import "@flanksource/clicky-ui/styles.css";
import "./extension.css";
import { Popup } from "./PopupView";

const root = document.getElementById("app");
if (!root) throw new Error("#app root not found");
createRoot(root).render(
  <ErrorWrapper>
    <ThemeProvider>
      <DensityProvider>
        <Popup />
      </DensityProvider>
    </ThemeProvider>
  </ErrorWrapper>,
);
