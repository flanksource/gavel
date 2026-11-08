import { render, screen } from "@testing-library/react";
import { expect, it } from "vitest";
import FixtureExample from "./FixtureExample";

it("renders Markdown headings and tables in the preview without an embedded schema viewer", async () => {
  const markdown = "---\nexec: printf\n---\n\n## Greeting\n\n| Name | Expected |\n| --- | --- |\n| Greeting | ready |\n";
  render(<FixtureExample>
    <pre><code>{"kind: Fixture\nsteps:\n  - exec: printf\n    args: [ready]"}</code></pre>
    <pre><code><span>{markdown.slice(0, 20)}</span>{markdown.slice(20)}</code></pre>
  </FixtureExample>);

  expect(await screen.findByRole("heading", { name: "Greeting" })).toBeInTheDocument();
  expect(screen.getByRole("table")).toHaveTextContent("Expected");
  expect(screen.queryByRole("tree")).not.toBeInTheDocument();
});
