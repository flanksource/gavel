// @vitest-environment jsdom
import { act, type ReactNode } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { EditListing } from "./api";

(globalThis as { IS_REACT_ACT_ENVIRONMENT?: boolean }).IS_REACT_ACT_ENVIRONMENT = true;

const request = vi.fn();
vi.mock("./api", async (actual) => ({ ...(await actual<typeof import("./api")>()), request: (...args: unknown[]) => request(...args) }));
// Tag comboboxes become comma-separated inputs; the modal a plain dialog.
vi.mock("@flanksource/clicky-ui/components", async (actual) => ({
  ...(await actual<object>()),
  Combobox: ({ value, onChange, ariaLabel }: { value: string[]; onChange: (value: string[]) => void; ariaLabel: string }) => (
    <input aria-label={ariaLabel} defaultValue={value.join(",")} onBlur={(event) => onChange(event.target.value.split(",").filter(Boolean))} />
  ),
  Modal: ({ children, footer }: { children: ReactNode; footer: ReactNode }) => (
    <div role="dialog">
      {children}
      {footer}
    </div>
  ),
}));

const { EditOptionsDialog, listingSource } = await import("./EditOptionsDialog");

const listing: EditListing = {
  extensions: [".md"],
  includeIgnored: [],
  configured: { extensions: [".md"] },
  flags: {},
  dir: "/work/docs",
  defaults: [".md", ".yaml", ".png"],
};

let root: Root | undefined;
let host: HTMLElement;
beforeEach(() => {
  request.mockImplementation(async (_path: string, body?: unknown) => (body === undefined ? listing : { listing, roots: [listing.dir], files: [] }));
});
afterEach(() => {
  act(() => root?.unmount());
  host?.remove();
  request.mockReset();
});

async function renderDialog() {
  const onApplied = vi.fn();
  host = document.createElement("div");
  document.body.append(host);
  root = createRoot(host);
  await act(async () => {
    root!.render(<EditOptionsDialog onApplied={onApplied} onClose={() => {}} />);
  });
  return onApplied;
}

async function type(label: string, value: string) {
  const input = host.querySelector<HTMLInputElement>(`input[aria-label="${label}"]`)!;
  await act(async () => {
    input.value = value;
    input.dispatchEvent(new FocusEvent("focusout", { bubbles: true }));
  });
}

const click = async (label: string) => {
  const button = [...host.querySelectorAll("button")].find((candidate) => candidate.textContent === label)!;
  await act(async () => button.click());
};

describe("listingSource", () => {
  it("names the flag when the value in effect is the flag's", () => {
    expect(listingSource({ ...listing, extensions: [".go"], flags: { extensions: [".go"] } }, "extensions")).toBe("--ext flag");
  });
  it("names the folder's .gavel.yaml when the value in effect is the configured one", () => {
    expect(listingSource(listing, "extensions")).toBe("/work/docs/.gavel.yaml");
  });
  it("is the default when nothing sets the key", () => {
    expect(listingSource(listing, "includeIgnored")).toBe("default");
  });
  it("says a value applied in the editor is not saved, even over a flag", () => {
    expect(listingSource({ ...listing, extensions: [".md", ".txt"], flags: { extensions: [".go"] } }, "extensions")).toBe("applied in this editor, not saved");
  });
});

describe("EditOptionsDialog", () => {
  it("applies the edited lists without saving", async () => {
    const onApplied = await renderDialog();
    await type("Extensions", ".md,.go");
    await type("Show despite .gitignore", ".tmp");

    await click("Apply");

    expect(request).toHaveBeenLastCalledWith("/api/edit/options", { extensions: [".md", ".go"], includeIgnored: [".tmp"], save: false }, "PUT");
    expect(onApplied).toHaveBeenCalledWith({ listing, roots: [listing.dir], files: [] });
  });

  it("saves the lists to .gavel.yaml", async () => {
    await renderDialog();

    await click("Save to .gavel.yaml");

    expect(request).toHaveBeenLastCalledWith("/api/edit/options", { extensions: [".md"], includeIgnored: [], save: true }, "PUT");
  });

  it("shows the server's rejection and stays open", async () => {
    const onApplied = await renderDialog();
    request.mockRejectedValueOnce(new Error('400: edit.extensions: "go" must look like .md'));

    await click("Apply");

    expect(host.querySelector('[role="alert"]')?.textContent).toContain('"go" must look like .md');
    expect(onApplied).not.toHaveBeenCalled();
  });
});
