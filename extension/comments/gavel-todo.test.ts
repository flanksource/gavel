import { afterEach, describe, expect, it, vi } from "vitest";
import {
  HTML_LIMIT,
  STACK_TIMEOUT_MS,
  buildBody,
  buildTodoUrl,
  cssPath,
  fileLine,
  grabHtml,
  normalizeOpenFilePath,
  openFileHook,
  resolveWithTimeout,
  screenshotMarkdown,
  todoBodiesForElements,
  todoBodyForContext,
  todoBodyForElement,
  todoBodyWithScreenshot,
  todoTitle,
  type TodoGrab,
} from "./gavel-todo";

afterEach(() => {
  document.body.replaceChildren();
  vi.useRealTimers();
  vi.restoreAllMocks();
});

const pageUrl = "https://app.example/dashboard?tab=1";

function grab(overrides: Partial<TodoGrab> = {}): TodoGrab {
  return {
    getDisplayName: () => "SaveButton",
    getSource: () => Promise.resolve({ filePath: "src/Save.tsx", lineNumber: 12, columnNumber: null, componentName: "SaveButton" }),
    getStackContext: () => Promise.resolve("in SaveButton (at src/Save.tsx:12)"),
    ...overrides,
  };
}

function mount(html: string): Element {
  document.body.innerHTML = html;
  return document.body.firstElementChild!;
}

describe("grabHtml", () => {
  it("returns the element's outerHTML when it fits", () => {
    expect(grabHtml(mount('<button id="a">Save</button>'))).toBe('<button id="a">Save</button>');
  });

  it("clips markup past 2KB and marks the truncation", () => {
    const el = mount(`<div>${"x".repeat(HTML_LIMIT * 2)}</div>`);
    const html = grabHtml(el);
    expect(html.startsWith(el.outerHTML.slice(0, HTML_LIMIT))).toBe(true);
    expect(html.endsWith("\n<!-- …truncated to 2KB -->")).toBe(true);
    expect(html.length).toBe(HTML_LIMIT + "\n<!-- …truncated to 2KB -->".length);
  });

  it("is empty without an element", () => {
    expect(grabHtml(null)).toBe("");
  });
});

describe("cssPath", () => {
  it("stops at the nearest unique id", () => {
    const el = mount('<main><section id="panel"><p>a</p><p class="t">b</p></section></main>').querySelector(".t")!;
    expect(cssPath(el)).toBe("#panel > p:nth-child(2)");
  });

  it("walks nth-child steps up to body when no unique id exists", () => {
    const el = mount("<main><p>a</p><p>b</p></main>").querySelectorAll("p")[1]!;
    expect(cssPath(el)).toBe("body > main:nth-child(1) > p:nth-child(2)");
  });

  it("ignores an id that is duplicated or not a valid identifier", () => {
    document.body.innerHTML = '<div id="dup"></div><div id="dup"><span id="1x"></span></div>';
    const span = document.querySelector("span")!;
    expect(cssPath(span)).toBe("body > div:nth-child(2) > span:nth-child(1)");
  });
});

describe("todoTitle", () => {
  it("falls back from component to tag to a generic name", () => {
    const el = mount("<i></i>");
    expect([todoTitle(el, "Card", "div"), todoTitle(el, undefined, "div"), todoTitle(el, undefined, undefined)]).toEqual(["Card", "div", "UI element"]);
  });

  it("appends a trimmed aria-label to tell sibling instances apart", () => {
    expect(todoTitle(mount('<button aria-label="  Delete row 3 ">x</button>'), "RowAction", "button")).toBe("RowAction: Delete row 3");
  });
});

describe("buildBody", () => {
  it("renders component, page, selector, source stack and html as Markdown", () => {
    expect(buildBody({ componentName: "Card", pageUrl, selector: "#a > p:nth-child(1)", source: "stack line", html: "<p>x</p>" })).toBe(
      [
        "**Component:** `<Card>`",
        `**Page:** ${pageUrl}`,
        "**Selector:** `#a > p:nth-child(1)`",
        "",
        "```",
        "stack line",
        "```",
        "",
        "**HTML:**",
        "```html",
        "<p>x</p>",
        "```",
      ].join("\n"),
    );
  });

  it("omits the component, selector, source and html sections that are empty", () => {
    expect(buildBody({ componentName: "", pageUrl, selector: "", source: "", html: "" })).toBe(`**Page:** ${pageUrl}`);
  });
});

describe("fileLine", () => {
  it.each([
    [undefined, undefined, ""],
    ["src/a.tsx", 4, "src/a.tsx:4"],
    ["src/a.tsx", undefined, "src/a.tsx:"],
  ])("formats %s line %s as %j", (file, line, expected) => expect(fileLine(file, line)).toBe(expected));
});

describe("resolveWithTimeout", () => {
  it("resolves the value, or the fallback when the value is falsy", async () => {
    await expect(resolveWithTimeout(Promise.resolve("stack"), "fallback", 50)).resolves.toBe("stack");
    await expect(resolveWithTimeout(Promise.resolve(""), "fallback", 50)).resolves.toBe("fallback");
  });

  it("resolves the fallback on rejection", async () => {
    await expect(resolveWithTimeout(Promise.reject(new Error("boom")), "fallback", 50)).resolves.toBe("fallback");
  });

  it("resolves the fallback when the promise outlives the timeout", async () => {
    vi.useFakeTimers();
    const pending = resolveWithTimeout(new Promise<string>(() => {}), "fallback", STACK_TIMEOUT_MS);
    await vi.advanceTimersByTimeAsync(STACK_TIMEOUT_MS);
    await expect(pending).resolves.toBe("fallback");
  });
});

describe("todoBodyForElement", () => {
  const html = "<button>Save</button>";

  it("uses the component stack, the display name and the snapshotted selector", async () => {
    const el = mount(html);
    await expect(todoBodyForElement({ grab: grab(), pageUrl, element: el })).resolves.toBe(
      buildBody({ componentName: "SaveButton", pageUrl, selector: "body > button:nth-child(1)", source: "in SaveButton (at src/Save.tsx:12)", html }),
    );
  });

  it("falls back to the action's file:line when the stack is empty", async () => {
    const el = mount(html);
    const body = await todoBodyForElement({ grab: grab({ getStackContext: () => Promise.resolve("") }), pageUrl, element: el, componentName: "Given", filePath: "src/Given.tsx", lineNumber: 3 });
    expect(body).toBe(buildBody({ componentName: "Given", pageUrl, selector: "body > button:nth-child(1)", source: "src/Given.tsx:3", html }));
  });

  it("falls back to getSource's file:line when the stack never resolves within 1500ms", async () => {
    vi.useFakeTimers();
    const el = mount(html);
    const pending = todoBodyForElement({ grab: grab({ getStackContext: () => new Promise<string>(() => {}) }), pageUrl, element: el });
    await vi.advanceTimersByTimeAsync(STACK_TIMEOUT_MS);
    await expect(pending).resolves.toBe(buildBody({ componentName: "SaveButton", pageUrl, selector: "body > button:nth-child(1)", source: "src/Save.tsx:12", html }));
  });

  it("leaves the selector empty for a disconnected element and honors a supplied html snapshot", async () => {
    const el = document.createElement("button");
    const body = await todoBodyForElement({ grab: grab(), pageUrl, element: el, html: "<b>snapshot</b>" });
    expect(body).toBe(buildBody({ componentName: "SaveButton", pageUrl, selector: "", source: "in SaveButton (at src/Save.tsx:12)", html: "<b>snapshot</b>" }));
  });

  it("warns and keeps going when getSource fails", async () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
    const el = mount(html);
    const body = await todoBodyForElement({ grab: grab({ getSource: () => Promise.reject(new Error("no source")), getStackContext: () => Promise.resolve("") }), pageUrl, element: el });
    expect(body).toBe(buildBody({ componentName: "SaveButton", pageUrl, selector: "body > button:nth-child(1)", source: "", html }));
    expect(warn).toHaveBeenCalledTimes(1);
  });
});

describe("todoBodiesForElements and todoBodyForContext", () => {
  it("joins several elements' bodies with a horizontal rule and skips empty entries", async () => {
    document.body.innerHTML = "<p>a</p><p>b</p>";
    const [first, second] = [...document.querySelectorAll("p")];
    const bodies = await todoBodiesForElements({ grab: grab(), pageUrl, elements: [first!, null as unknown as Element, second!] });
    expect(bodies.split("\n\n---\n\n")).toHaveLength(2);
    expect(await todoBodiesForElements({ grab: grab(), pageUrl, elements: [] })).toBe("");
  });

  it("builds one body for a single-element context and a joined body for several", async () => {
    document.body.innerHTML = "<p>a</p><p>b</p>";
    const [first, second] = [...document.querySelectorAll("p")] as unknown as [Element, Element];
    const single = await todoBodyForContext({ grab: grab(), pageUrl, context: { element: first, elements: [first], componentName: "Given" } });
    expect(single).toContain("**Component:** `<Given>`");
    const several = await todoBodyForContext({ grab: grab(), pageUrl, context: { element: first, elements: [first, second] } });
    expect(several.split("\n\n---\n\n")).toHaveLength(2);
  });
});

describe("buildTodoUrl", () => {
  it("opens the dashboard's embed form with the title, body and source URL prefilled", () => {
    const url = new URL(buildTodoUrl({ server: "http://localhost:9092", title: "Card: a&b", body: "line1\nline2", sourceUrl: pageUrl }));
    expect([url.origin, url.pathname, Object.fromEntries(url.searchParams)]).toEqual([
      "http://localhost:9092",
      "/todos/new",
      { embed: "1", title: "Card: a&b", body: "line1\nline2", sourceUrl: pageUrl },
    ]);
  });

  it("leaves out empty fields", () => {
    const url = new URL(buildTodoUrl({ server: "http://localhost:9092", title: "", body: "", sourceUrl: "" }));
    expect([...url.searchParams.keys()]).toEqual(["embed"]);
  });
});

describe("screenshot body", () => {
  const attachment = { url: "http://localhost:9092/files/a.png", filename: "shot.png" };

  it("links the uploaded image, naming it screenshot.png when the server gave no filename", () => {
    expect(screenshotMarkdown(attachment)).toBe("![shot.png](http://localhost:9092/files/a.png)");
    expect(screenshotMarkdown({ url: attachment.url })).toBe("![screenshot.png](http://localhost:9092/files/a.png)");
  });

  it("appends the image under an Attachments heading, or is just the image for an empty body", () => {
    expect(todoBodyWithScreenshot("  body text \n", attachment)).toBe("body text\n\n## Attachments\n\n![shot.png](http://localhost:9092/files/a.png)");
    expect(todoBodyWithScreenshot("", attachment)).toBe("![shot.png](http://localhost:9092/files/a.png)");
  });
});

describe("open file hook", () => {
  it.each([
    ["/@fs/Users/me/app/src/a.tsx", "/Users/me/app/src/a.tsx"],
    ["/src/a.tsx", "/src/a.tsx"],
    ["", ""],
  ])("normalizes %j to %j", (input, expected) => expect(normalizeOpenFilePath(input)).toBe(expected));

  it("leaves ordinary paths to react-grab", () => {
    const fetcher = vi.fn();
    expect(openFileHook({ fetcher, openWindow: vi.fn() })("/src/a.tsx", 3)).toBe(false);
    expect(fetcher).not.toHaveBeenCalled();
  });

  it("asks the page's dev server to open a /@fs/ path and claims it handled", () => {
    const fetcher = vi.fn(() => Promise.resolve({ ok: true }));
    const openWindow = vi.fn();
    expect(openFileHook({ fetcher, openWindow })("/@fs/Users/me/a.tsx", 7)).toBe(true);
    expect(fetcher).toHaveBeenCalledWith("/__open-in-editor?file=%2FUsers%2Fme%2Fa.tsx&column=1&line=7");
    expect(openWindow).not.toHaveBeenCalled();
  });

  it.each([
    ["rejects", () => Promise.reject(new Error("offline"))],
    ["answers not ok", () => Promise.resolve({ ok: false })],
  ])("falls back to react-grab's open-file page when the dev server %s", async (_name, answer) => {
    const openWindow = vi.fn();
    const done = new Promise<void>((resolve) => openWindow.mockImplementation(() => resolve()));
    expect(openFileHook({ fetcher: answer, openWindow })("/@fs/Users/me/a.tsx", 7)).toBe(true);
    await done;
    expect(openWindow).toHaveBeenCalledWith("https://react-grab.com/open-file?url=%2FUsers%2Fme%2Fa.tsx&line=7", "_blank", "noopener,noreferrer");
  });

  it("omits the line parameters when there is no line", () => {
    const fetcher = vi.fn(() => Promise.resolve({ ok: true }));
    openFileHook({ fetcher, openWindow: vi.fn() })("/@fs/a.tsx");
    expect(fetcher).toHaveBeenCalledWith("/__open-in-editor?file=%2Fa.tsx&column=1");
  });
});
