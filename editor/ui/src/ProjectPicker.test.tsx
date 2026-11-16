// @vitest-environment jsdom
import { act, type ReactNode } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { EditProject, OpenedFiles } from "./api";

(globalThis as { IS_REACT_ACT_ENVIRONMENT?: boolean }).IS_REACT_ACT_ENVIRONMENT = true;

const request = vi.fn();
vi.mock("./api", async (actual) => ({ ...(await actual<typeof import("./api")>()), request: (...args: unknown[]) => request(...args) }));
// The menu combobox and the modal render through floating overlays; a select and a
// plain dialog keep the spec on what the picker does with a choice.
vi.mock("@flanksource/clicky-ui/components", async (actual) => ({
  ...(await actual<object>()),
  Combobox: ({ options, value, onChange, ariaLabel }: { options: { value: string; label: string }[]; value: string; onChange: (value: string) => void; ariaLabel: string }) => (
    <select aria-label={ariaLabel} value={value} onChange={(event) => onChange(event.target.value)}>
      <option value="">none</option>
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  ),
  Modal: ({ title, children, footer }: { title: ReactNode; children: ReactNode; footer: ReactNode }) => (
    <div role="dialog" aria-label={String(title)}>
      {children}
      {footer}
    </div>
  ),
}));

const { ProjectPicker, currentProject } = await import("./ProjectPicker");

const gavel: EditProject = { name: "gavel", dir: "/work/gavel" };
const docs: EditProject = { name: "docs", dir: "/work/docs" };
const switched: OpenedFiles = { roots: [docs.dir], files: [{ id: "f1", path: "/work/docs/README.md", kind: "document" }] };

let root: Root | undefined;
let host: HTMLElement;
beforeEach(() => {
  request.mockImplementation(async (path: string) => (path === "/api/edit/projects" ? { projects: [docs, gavel], roots: [gavel.dir] } : switched));
});
afterEach(() => {
  act(() => root?.unmount());
  host?.remove();
  request.mockReset();
});

async function renderPicker(unsaved: number) {
  const onSwitched = vi.fn();
  host = document.createElement("div");
  document.body.append(host);
  root = createRoot(host);
  await act(async () => {
    root!.render(<ProjectPicker roots={[gavel.dir]} unsaved={() => unsaved} onSwitched={onSwitched} onAccessDenied={() => {}} onError={(message) => { throw new Error(message); }} />);
  });
  return onSwitched;
}

async function choose(dir: string) {
  const select = host.querySelector<HTMLSelectElement>('select[aria-label="Project"]')!;
  await act(async () => {
    select.value = dir;
    select.dispatchEvent(new Event("change", { bubbles: true }));
  });
}

const button = (label: string) => [...host.querySelectorAll("button")].find((candidate) => candidate.textContent === label)!;

describe("currentProject", () => {
  it("is the project whose directory is the only open folder", () => {
    expect(currentProject([docs, gavel], [gavel.dir])).toEqual(gavel);
  });
  it("is none while several folders are open", () => {
    expect(currentProject([docs, gavel], [gavel.dir, docs.dir])).toBeUndefined();
  });
});

describe("ProjectPicker", () => {
  it("switches to the chosen project at once when every draft is saved", async () => {
    const onSwitched = await renderPicker(0);
    expect(host.querySelector<HTMLSelectElement>("select")!.value).toBe(gavel.dir);

    await choose(docs.dir);

    expect(request).toHaveBeenCalledWith("/api/edit/project", { dir: docs.dir });
    expect(onSwitched).toHaveBeenCalledWith(switched);
  });

  it("asks before closing unsaved drafts, and keeps them when the reader keeps editing", async () => {
    const onSwitched = await renderPicker(2);

    await choose(docs.dir);
    expect(host.querySelector('[role="dialog"]')?.textContent).toContain("2 files have changes that are not saved yet");
    expect(request).not.toHaveBeenCalledWith("/api/edit/project", expect.anything());

    await act(async () => button("Keep editing").click());
    expect(onSwitched).not.toHaveBeenCalled();
  });

  it("switches after the reader confirms discarding the drafts", async () => {
    const onSwitched = await renderPicker(1);

    await choose(docs.dir);
    await act(async () => button("Discard and switch").click());

    expect(request).toHaveBeenCalledWith("/api/edit/project", { dir: docs.dir });
    expect(onSwitched).toHaveBeenCalledWith(switched);
  });
});
