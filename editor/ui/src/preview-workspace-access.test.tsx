// @vitest-environment jsdom
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, expect, it, vi } from "vitest";
import type { WorkspaceAccessClient } from "@flanksource/clicky-ui/components";
import { ApiError } from "./api";
import { PreviewWorkspaceAccess, deniedWorkspaceFolder, forgetPreviewFolder, rememberedPreviewFolder } from "./preview-workspace-access";

(globalThis as { IS_REACT_ACT_ENVIRONMENT?: boolean }).IS_REACT_ACT_ENVIRONMENT = true;

const folder = "/project";
const grant = { id: "grant", path: folder, source: "saved" as const, available: true };
let root: Root | undefined;
let container: HTMLDivElement;

afterEach(() => {
  act(() => root?.unmount());
  container?.remove();
  vi.restoreAllMocks();
  window.localStorage.clear();
});

async function mount(client: WorkspaceAccessClient) {
  const approved = vi.fn();
  container = document.createElement("div");
  document.body.append(container);
  root = createRoot(container);
  await act(async () => root!.render(<PreviewWorkspaceAccess fileId="report" folder={folder} client={client} onApproved={approved} onManage={() => {}} />));
  return approved;
}

function clientFixture() {
  return { preview: vi.fn(async () => ({ path: folder })), grant: vi.fn(async () => grant), revoke: vi.fn(), list: vi.fn(), subscribe: vi.fn() } satisfies WorkspaceAccessClient;
}

function approve() {
  const button = Array.from(container.querySelectorAll("button")).find((item) => item.textContent === "Approve access");
  expect(button).toBeDefined();
  return act(async () => button!.click());
}

it("records successful access per file and forgets it when the folder is denied again", async () => {
  const client = clientFixture();
  const approved = await mount(client);
  expect(container.textContent).toContain(folder);
  expect(client.grant).not.toHaveBeenCalled();
  await approve();
  expect(client.grant).toHaveBeenCalledExactlyOnceWith(folder, folder);
  expect(approved).toHaveBeenCalledOnce();
  expect([rememberedPreviewFolder("report"), rememberedPreviewFolder("another")]).toEqual([folder, undefined]);
  const denied = new ApiError(403, JSON.stringify({ code: "workspace_access_denied", root: folder }));
  expect(deniedWorkspaceFolder(denied)).toBe(folder);
  expect(deniedWorkspaceFolder(new ApiError(403, "Forbidden"))).toBeUndefined();
  forgetPreviewFolder("report", folder);
  expect(rememberedPreviewFolder("report")).toBeUndefined();
});

it("blocks approval when canonical review no longer matches the denied folder", async () => {
  const client = clientFixture();
  client.preview.mockResolvedValue({ path: "/changed" });
  await mount(client);
  expect(container.textContent).toContain("Folder changed");
  expect(Array.from(container.querySelectorAll("button")).some((button) => button.textContent === "Approve access" && !button.disabled)).toBe(false);
  expect(client.grant).not.toHaveBeenCalled();
});

it("shows a failed grant and allows retry even when browser storage is unavailable", async () => {
  const client = clientFixture();
  client.grant.mockRejectedValueOnce(new Error("Could not save access"));
  vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => { throw new Error("Storage unavailable"); });
  const approved = await mount(client);
  await approve();
  expect(container.textContent).toContain("Could not save access");
  expect(approved).not.toHaveBeenCalled();
  expect(rememberedPreviewFolder("report")).toBeUndefined();
  await approve();
  expect(approved).toHaveBeenCalledOnce();
});
