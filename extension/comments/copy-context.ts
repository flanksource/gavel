import { cssPath } from "./dom-anchor";
import { captureElementContext } from "./element-context";

// elementContextMarkdown is what react-grab's Copy hands a coding agent for one
// element: the component, the page it is on, a selector for this instance
// (one component often renders many identical elements), the component stack
// and the raw HTML. It matches gavel's react-grab plugin body.
async function elementContextMarkdown(element: Element): Promise<string> {
  const { componentName, source, html } = await captureElementContext(element);
  return [
    ...(componentName ? [`**Component:** \`<${componentName}>\``] : []),
    `**Page:** ${window.location.href}`,
    `**Selector:** \`${cssPath(element, document.body).replace(/^:scope/, "body")}\``,
    "",
    "```",
    source,
    "```",
    "",
    "**HTML:**",
    "```html",
    html,
    "```",
  ].join("\n");
}

export async function copyContextMarkdown(elements: Element[]): Promise<string> {
  const parts = await Promise.all(elements.map(elementContextMarkdown));
  return parts.join("\n\n---\n\n");
}
