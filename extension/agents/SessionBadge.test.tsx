import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import { SessionBadge } from "./SessionBadge";

afterEach(cleanup);

const TITLE = "Fix the alignment of the heading and the body text in the checkout page";

describe("SessionBadge", () => {
  it("shows the session title and exposes it in full as the title attribute", () => {
    render(<SessionBadge title={TITLE} />);
    expect(screen.getByText(TITLE).getAttribute("title")).toBe(TITLE);
  });

  it("is plain text, not a copy button", () => {
    render(<SessionBadge title={TITLE} />);
    expect(screen.queryByRole("button")).toBeNull();
  });
});
