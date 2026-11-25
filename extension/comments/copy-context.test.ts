import { afterEach, beforeEach, describe, expect, it } from "vitest";

import { copyContextMarkdown } from "./copy-context";

const COMPONENT = "BadgeTrigger";
const STACK = "in BadgeTrigger (at src/pages/oipa/_transaction-math/math-badges.tsx)";

function stubReactGrab() {
  Object.assign(window, {
    __REACT_GRAB__: {
      getDisplayName: () => COMPONENT,
      getSource: () => Promise.resolve({ filePath: "math-badges.tsx", lineNumber: 45 }),
      getStackContext: () => Promise.resolve(STACK),
    },
  });
}

function renderRows(): HTMLButtonElement[] {
  document.body.innerHTML = `
    <main>
      <ul>
        <li><button data-badge="expr" aria-label="EXPR of Gross">EXPR</button></li>
        <li><button data-badge="expr" aria-label="EXPR of Net">EXPR</button></li>
      </ul>
    </main>`;
  return [...document.querySelectorAll("button")];
}

describe("copyContextMarkdown", () => {
  beforeEach(stubReactGrab);
  afterEach(() => {
    Reflect.deleteProperty(window, "__REACT_GRAB__");
    document.body.innerHTML = "";
  });

  it("carries the page URL, component, a resolvable selector, the stack and the HTML", async () => {
    const [, net] = renderRows();

    const markdown = await copyContextMarkdown([net!]);

    const selector = /\*\*Selector:\*\* `([^`]+)`/.exec(markdown)?.[1];
    expect(document.querySelector(selector!)).toBe(net);
    expect(markdown).toContain(`**Component:** \`<${COMPONENT}>\``);
    expect(markdown).toContain(`**Page:** ${window.location.href}`);
    expect(markdown).toContain(STACK);
    expect(markdown).toContain(net!.outerHTML);
  });

  it("separates each grabbed element's context", async () => {
    const rows = renderRows();

    const markdown = await copyContextMarkdown(rows);

    expect(markdown.split("\n\n---\n\n")).toHaveLength(rows.length);
  });
});
