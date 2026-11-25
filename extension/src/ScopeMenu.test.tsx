import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { ScopeMenu } from "./ScopeMenu";

afterEach(cleanup);

const HOST = "app.example.test:3000";

function renderMenu(overrides: { defaultLabel?: string } = {}) {
  const onReset = vi.fn();
  const onPromote = vi.fn();
  render(<ScopeMenu host={HOST} valueLabel="docs" onReset={onReset} onPromote={onPromote} {...overrides} />);
  return { onReset, onPromote };
}

it("is a compact This site button that keeps its menu closed until opened", () => {
  renderMenu();
  expect(screen.getByRole("button", { name: "This site" })).toBeTruthy();
  expect(screen.queryByRole("menu")).toBeNull();
});

it("says the value applies only to the host and offers the two moves", () => {
  renderMenu();
  fireEvent.click(screen.getByRole("button", { name: "This site" }));
  expect(screen.getByText(`docs applies only to ${HOST}.`)).toBeTruthy();
  expect(screen.getAllByRole("menuitem").map(item => item.textContent)).toEqual(["Use the default here", "Make this the default for all sites"]);
});

it("shows the default's label beside Use the default here when it is given", () => {
  renderMenu({ defaultLabel: "acme" });
  fireEvent.click(screen.getByRole("button", { name: "This site" }));
  expect(screen.getByRole("menuitem", { name: "Use the default here acme" })).toBeTruthy();
});

it("calls onReset for Use the default here and onPromote for Make this the default, each once", () => {
  const { onReset, onPromote } = renderMenu();
  fireEvent.click(screen.getByRole("button", { name: "This site" }));
  fireEvent.click(screen.getByRole("menuitem", { name: "Use the default here" }));
  expect(onReset).toHaveBeenCalledTimes(1);
  expect(onPromote).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "This site" }));
  fireEvent.click(screen.getByRole("menuitem", { name: "Make this the default for all sites" }));
  expect(onPromote).toHaveBeenCalledTimes(1);
  expect(onReset).toHaveBeenCalledTimes(1);
});
