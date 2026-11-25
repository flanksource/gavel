import { createRoot } from "react-dom/client";
import { ErrorWrapper } from "@flanksource/clicky-ui";
import "@flanksource/clicky-ui/styles.css";
import "./extension.css";
import { connectHostBridge } from "../host/useHostBridge";
import { GavelPanel } from "./GavelPanel";
import { requestBackground } from "./messages";
import { assertActivated, panelTarget, type Activation } from "./settings";
import { assertServerAccess } from "./server-access";

// Framed by the grabber host on the page. Like the relay, it only serves the
// page that frames it, and only for the settings the user opened the grabber
// with: the background worker holds that record, the page cannot forge it.
async function start() {
  const target = panelTarget({ search: location.search, ancestorOrigin: location.ancestorOrigins[0] });
  const record = await requestBackground<Activation | undefined>({ type: "activation" });
  assertActivated({ record, request: { server: target.server, project: target.project } });
  await assertServerAccess(target.server);
  connectHostBridge({ parentOrigin: target.parentOrigin, session: target.bridgeSession });
  const root = document.getElementById("app");
  if (!root) throw new Error("#app root not found");
  createRoot(root).render(
    <ErrorWrapper>
      <GavelPanel target={target} />
    </ErrorWrapper>,
  );
}

start().catch((cause: unknown) => {
  const error = document.createElement("p");
  error.setAttribute("role", "alert");
  error.className = "m-3 text-sm text-destructive";
  error.textContent = `Gavel grabber panel: ${cause instanceof Error ? cause.message : String(cause)}`;
  document.getElementById("app")?.replaceChildren(error);
  throw cause;
});
