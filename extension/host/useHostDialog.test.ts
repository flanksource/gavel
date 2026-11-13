// @vitest-environment jsdom
import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { useHostDialog } from "./useHostDialog";
import { sendToHost } from "./useHostBridge";

vi.mock("./useHostBridge", () => ({ sendToHost: vi.fn() }));

afterEach(() => {
  cleanup();
  document.body.replaceChildren();
  vi.clearAllMocks();
});

it.each(["menu", "listbox", "dialog"])("keeps the host size when a non-modal %s opens", async role => {
  renderHook(() => useHostDialog());
  await act(async () => {
    const popup = document.createElement("div");
    popup.setAttribute("role", role);
    popup.setAttribute("aria-modal", "false");
    document.body.append(popup);
  });
  expect(sendToHost).not.toHaveBeenCalled();
});

it("expands for a modal and restores the host when it closes", async () => {
  renderHook(() => useHostDialog());
  const modal = document.createElement("div");
  modal.setAttribute("role", "dialog");
  modal.setAttribute("aria-modal", "true");
  await act(async () => { document.body.append(modal); });
  await act(async () => { modal.remove(); });
  expect(vi.mocked(sendToHost).mock.calls).toEqual([
    [{ type: "dialog", open: true }],
    [{ type: "dialog", open: false }],
  ]);
});

it("restores an expanded host when its panel becomes hidden", () => {
  const modal = document.createElement("div");
  modal.setAttribute("aria-modal", "true");
  document.body.append(modal);
  const { rerender } = renderHook(({ visible }) => useHostDialog(visible), { initialProps: { visible: true } });
  rerender({ visible: false });
  expect(vi.mocked(sendToHost).mock.calls).toEqual([
    [{ type: "dialog", open: true }],
    [{ type: "dialog", open: false }],
  ]);
});
