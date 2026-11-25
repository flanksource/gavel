import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { PanelComposer } from "./PanelComposer";
import { GAVEL_COMMENT_CONFIG } from "./useGavelComments";

afterEach(() => { cleanup(); localStorage.clear(); });

describe("PanelComposer", () => {
  it("keeps the whole-page editor open in the footer instead of collapsing it to a pill", () => {
    render(<PanelComposer config={GAVEL_COMMENT_CONFIG} onPost={vi.fn()} onPick={vi.fn()} />);
    expect([screen.queryByTestId("comment-compose-input") !== null, screen.queryByRole("button", { name: "Comment on this page…" })]).toEqual([true, null]);
  });

  it("starts an element comment with or without a screenshot", () => {
    const onPick = vi.fn();
    render(<PanelComposer config={GAVEL_COMMENT_CONFIG} onPost={vi.fn()} onPick={onPick} />);
    fireEvent.click(screen.getByRole("button", { name: /Element/ }));
    fireEvent.click(screen.getByRole("button", { name: /Screenshot/ }));
    expect(onPick.mock.calls).toEqual([[{ screenshot: false }], [{ screenshot: true }]]);
  });
});
