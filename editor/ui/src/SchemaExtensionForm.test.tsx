// @vitest-environment jsdom
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, describe, expect, it, vi } from "vitest";

const pageForm = vi.fn((_props: Record<string, unknown>) => <p>runtime form</p>);
vi.mock("@flanksource/page-runtime/host", () => ({ PageForm: (props: Record<string, unknown>) => pageForm(props) }));
const request = vi.fn();
vi.mock("./api", () => ({ request: (...args: unknown[]) => request(...args) }));
const { previewExtensionForm } = await import("./SchemaExtensionForm");

(globalThis as { IS_REACT_ACT_ENVIRONMENT?: boolean }).IS_REACT_ACT_ENVIRONMENT = true;

let root: Root | undefined;
let container: HTMLDivElement | undefined;
afterEach(() => {
  act(() => root?.unmount());
  container?.remove();
  pageForm.mockClear();
  request.mockReset();
});

async function mount(schema: Record<string, unknown>) {
  const Form = previewExtensionForm("file-7");
  container = document.createElement("div");
  document.body.append(container);
  root = createRoot(container);
  await act(async () => root!.render(<Form schema={schema} value={{ name: "initial" }} onChange={() => {}} size="sm" />));
  return container;
}

describe("previewExtensionForm", () => {
  it("renders a schema without extensions as a plain form, with no runtime", async () => {
    const view = await mount({ type: "object", properties: { name: { type: "string", title: "Name" } } });
    expect({ input: view.querySelector("input")?.value, preview: request.mock.calls, runtime: pageForm.mock.calls }).toEqual({ input: "initial", preview: [], runtime: [] });
  });

  it("fails the form with the runtime's error, naming the phase and file", async () => {
    request.mockResolvedValue({ runtimeOrigin: "http://127.0.0.1:9100", paths: [], dependencies: {} });
    await mount({ type: "object", "x-clicky-extensions": [{ ref: "x.tsx" }] });
    const onMessage = pageForm.mock.calls.at(-1)?.[0]?.onMessage as (message: unknown) => void;
    expect(() => act(() => onMessage({ type: "error", error: { phase: "compile", file: "x.tsx", line: 3, message: "Unexpected token" } }))).toThrow(
      "form extensions could not load: compile x.tsx:3: Unexpected token",
    );
  });

  it("renders a schema that declares extensions in the page runtime, serving the file's preview sources", async () => {
    request.mockResolvedValue({ runtimeOrigin: "http://127.0.0.1:9100", paths: ["schemas/extensions/pickers.tsx"], dependencies: { "@flanksource/clicky-ui": "0.3.57" } });
    const schema = { type: "object", "x-clicky-extensions": [{ ref: "schemas/extensions/pickers.tsx" }] };
    const view = await mount(schema);
    const props = pageForm.mock.calls.at(-1)?.[0];
    expect({
      text: view.textContent,
      preview: request.mock.calls,
      runtime: { schema: props?.schema, value: props?.value, formProps: props?.formProps, manifest: props?.manifest, sourceProvider: props?.sourceProvider, extensions: props?.extensions },
    }).toEqual({
      text: "runtime form",
      preview: [["/api/file/file-7/preview"]],
      runtime: {
        schema,
        value: { name: "initial" },
        formProps: { size: "sm" },
        manifest: { dependencies: { "@flanksource/clicky-ui": "0.3.57" } },
        sourceProvider: { url: "http://127.0.0.1:9100/sources/file-7", paths: ["schemas/extensions/pickers.tsx"] },
        extensions: [],
      },
    });
  });
});
