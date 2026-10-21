import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { expect, it } from "vitest";
import { buildDocTree } from "@/lib/docs";
import type { MdxEntry } from "@/lib/mdx";
import DocsNavigation from "./DocsNavigation";

it("opens the active page's ancestors and closes unrelated branches", () => {
  const docs: MdxEntry[] = [
    { slug: "fixtures", frontmatter: { title: "Fixtures" }, Component: () => null },
    { slug: "steps", frontmatter: { title: "Steps", parent: "fixtures" }, Component: () => null },
    { slug: "exec", frontmatter: { title: "Exec", parent: "steps" }, Component: () => null },
    { slug: "lint", frontmatter: { title: "Lint" }, Component: () => null },
    { slug: "eslint", frontmatter: { title: "ESLint", parent: "lint" }, Component: () => null },
  ];
  render(<MemoryRouter><DocsNavigation nodes={buildDocTree(docs)} active="exec" /></MemoryRouter>);
  expect(screen.getByRole("link", { name: "Fixtures" }).closest("details")).toHaveAttribute("open");
  expect(screen.getByRole("link", { name: "Steps" }).closest("details")).toHaveAttribute("open");
  expect(screen.getByRole("link", { name: "Exec" })).toHaveAttribute("aria-current", "page");
  expect(screen.getByRole("link", { name: "Lint" }).closest("details")).not.toHaveAttribute("open");
});
