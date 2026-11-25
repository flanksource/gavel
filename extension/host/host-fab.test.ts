import { expect, it, vi } from "vitest";
import { mountFabMovement } from "./host-fab";

it("clamps restored positions and persists keyboard movement", () => {
  const button = document.createElement("button");
  const changed = vi.fn();
  const movement = mountFabMovement(button, changed);
  movement.layout({ x: -100, y: -100 });
  expect(movement.position()).toEqual({ x: 8, y: 8 });
  button.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowRight" }));
  expect(changed).toHaveBeenLastCalledWith({ x: 24, y: 8 });
  movement.layout({ x: 100000, y: 100000 });
  expect(movement.position()).toEqual({ x: window.innerWidth - 56, y: window.innerHeight - 56 });
  movement.close();
});

it("does not activate the FAB after dragging", () => {
  const button = document.createElement("button");
  button.setPointerCapture = vi.fn();
  const changed = vi.fn(), clicked = vi.fn();
  const movement = mountFabMovement(button, changed);
  button.addEventListener("click", clicked);
  button.dispatchEvent(new MouseEvent("pointerdown", { clientX: 100, clientY: 100 }));
  button.dispatchEvent(new MouseEvent("pointermove", { clientX: 60, clientY: 60 }));
  button.dispatchEvent(new MouseEvent("pointerup"));
  button.click();
  expect(changed).toHaveBeenCalledOnce();
  expect(clicked).not.toHaveBeenCalled();
  button.click();
  expect(clicked).toHaveBeenCalledOnce();
  movement.close();
});
