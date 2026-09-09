import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ProjectsPlaceholder } from "./ProjectsPlaceholder";

const EMPTY = "No projects configured";

// The regression this guards: the menubar rendered EMPTY while /api/projects was
// still in flight, so a slow (or failed) response read as "you have nothing".
describe("ProjectsPlaceholder", () => {
  it("says it is loading, not empty, before /api/projects resolves", () => {
    render(<ProjectsPlaceholder loaded={false} emptyText={EMPTY} className="c" />);

    expect(screen.getByText(/loading projects/i).getAttribute("aria-busy")).toBe("true");
    expect(screen.queryByText(EMPTY)).toBeNull();
  });

  it("shows the failure as an alert instead of the empty copy", () => {
    render(<ProjectsPlaceholder loaded error="Load projects failed (HTTP 500)" emptyText={EMPTY} className="c" />);

    expect(screen.getByRole("alert").textContent).toBe("Load projects failed (HTTP 500)");
    expect(screen.queryByText(EMPTY)).toBeNull();
  });

  // A failure that arrives before the first successful load must still read as a
  // failure: `loaded` flips on any resolved attempt, including a rejected one.
  it("prefers the failure over the loading state when both apply", () => {
    render(<ProjectsPlaceholder loaded={false} error="connection refused" emptyText={EMPTY} className="c" />);

    expect(screen.getByRole("alert").textContent).toBe("connection refused");
    expect(screen.queryByText(/loading projects/i)).toBeNull();
  });

  it("shows the empty copy only once loaded with no error", () => {
    render(<ProjectsPlaceholder loaded emptyText={EMPTY} className="c" />);

    expect(screen.getByText(EMPTY)).not.toBeNull();
    expect(screen.queryByRole("alert")).toBeNull();
    expect(screen.queryByText(/loading projects/i)).toBeNull();
  });
});
