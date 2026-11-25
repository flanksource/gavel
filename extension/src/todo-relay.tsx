import { createRoot } from "react-dom/client";
import { DensityProvider, ErrorWrapper, ThemeProvider } from "@flanksource/clicky-ui";
import "@flanksource/clicky-ui/styles.css";
import "./extension.css";
import { requestBackground } from "./messages";
import { serverOrigin, type Activation } from "./settings";
import { readGlobalDefaults } from "./storage";
import { todoRelayTarget } from "./todo-relay-target";
import { TodoForm } from "./todos/TodoForm";

async function start() {
  const params = new URLSearchParams(location.search);
  let server: string;
  let project: string | undefined;
  let formParams = params;
  let attachmentSource: { parentOrigin: string } | undefined;
  if (window.parent !== window) {
    const activation = await requestBackground<Activation | undefined>({ type: "activation" });
    if (!activation) throw new Error("The grabber was not activated on this tab");
    const target = todoRelayTarget({ target: params.get("target"), parentOrigin: params.get("parentOrigin"), ancestorOrigin: location.ancestorOrigins[0], server: activation.server });
    server = activation.server;
    project = activation.project;
    formParams = target.url.searchParams;
    attachmentSource = { parentOrigin: target.parentOrigin };
  } else {
    const defaults = await readGlobalDefaults();
    server = serverOrigin(params.get("server") ?? defaults.server);
    project = params.get("project") ?? defaults.project;
  }
  const parentOrigin = attachmentSource?.parentOrigin;
  const root = document.getElementById("app");
  if (!root) throw new Error("#app root not found");
  createRoot(root).render(<ErrorWrapper><ThemeProvider><DensityProvider><TodoForm
    server={server} params={formParams} {...(project ? { project } : {})} {...(attachmentSource ? { attachmentSource } : {})}
    onDone={type => {
      if (parentOrigin) window.parent.postMessage({ source: "gavel-react-grab", type }, parentOrigin === "null" ? "*" : parentOrigin);
      else window.close();
    }}
  /></DensityProvider></ThemeProvider></ErrorWrapper>);
}

start().catch((cause: unknown) => {
  const error = document.createElement("p");
  error.setAttribute("role", "alert");
  error.className = "m-3 text-sm text-destructive";
  error.textContent = `Gavel todo capture: ${String(cause)}`;
  document.getElementById("app")?.replaceChildren(error);
});
