import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { DraftDeliveryPreview } from "./DeliveryPreview";
import type { GavelClient } from "./gavel-client";

afterEach(cleanup);

it("previews the draft mention and restores the page target when the mention is removed", async () => {
  const page = { url: "https://example.test/docs", title: "Docs" };
  const session = { id: "old-123456", source: "claude", title: "Page session", inCmux: false };
  const alternate = { id: "new-abcdef", source: "codex", title: "Draft session", inCmux: false };
  const previewDelivery = vi.fn().mockResolvedValue({
    route: { source: "mention:session:new-abcdef", target: { kind: "session", sessionId: alternate.id } },
    session: alternate,
  });
  const props = {
    client: { previewDelivery } as unknown as GavelClient,
    page,
    sessionId: session.id,
    revision: undefined,
    defaultState: { preview: { session } },
  };
  const { rerender } = render(<DraftDeliveryPreview {...props} body="Check the header" />);
  expect(screen.getByText("Page session")).toBeTruthy();

  rerender(<DraftDeliveryPreview {...props} body="@draft-session check the header" />);
  expect(await screen.findByText("Draft session")).toBeTruthy();
  expect(previewDelivery).toHaveBeenCalledWith({
    page, body: "@draft-session check the header", delivery: { defaultSessionId: session.id },
  });

  rerender(<DraftDeliveryPreview {...props} body="Check the header" />);
  expect(screen.getByText("Page session")).toBeTruthy();
});

it("previews the explicit Amp destination ahead of a remembered session and a mention", async () => {
  const page = { url: "https://example.test/docs", title: "Docs" };
  const session = { id: "T-review", source: "amp", title: "Amp review", inCmux: false };
  const previewDelivery = vi.fn().mockResolvedValue({ session });
  render(<DraftDeliveryPreview client={{ previewDelivery } as unknown as GavelClient} page={page} sessionId="remembered"
    revision={undefined} defaultState={{}} body="@claude check this" explicitSessionId={session.id} />);
  expect(await screen.findByText(session.title)).toBeTruthy();
  expect(previewDelivery).toHaveBeenCalledWith({ page, body: "@claude check this", deliver: { kind: "session", sessionId: session.id } });
});
