import { fireEvent, render, screen } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { TodoProjectPicker } from "./TodoProjectPicker";

it("searches cached projects by name and submits their workspace directory", () => {
  const onChange = vi.fn();
  render(<TodoProjectPicker projects={[{ name: "Acme app", dir: "/work/acme" }, { name: "Docs", dir: "/work/docs" }]} dir="" disabled={false} onChange={onChange} />);
  const picker = screen.getByRole("combobox", { name: "Project" });
  fireEvent.change(picker, { target: { value: "Acme" } });
  fireEvent.mouseDown(screen.getByRole("option", { name: /Acme app/ }));
  expect(onChange).toHaveBeenCalledWith("/work/acme");
});
